/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/glogx"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog"
	nodeLabel "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	metricsinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/metrics"
)

const (
	// ReschedulerTaintKey is the name of the taint created by rescheduler.
	ReschedulerTaintKey = "CriticalAddonsOnly"

	gkeNodeTerminationHandlerTaint = "cloud.google.com/impending-node-termination"

	// filterNodeResourceAnnoKey filters nodes when calculating buffer and total resource
	filterNodeResourceAnnoKey = "io.tencent.bcs.dev/filter-node-resource"

	// nodeInstanceTypeLabelKey is the instance type of node
	nodeInstanceTypeLabelKey = "node.kubernetes.io/instance-type"
	// nodeInstanceTypeEklet indicates the instance type of node is eklet
	nodeInstanceTypeEklet = "eklet"

	// How old the oldest unschedulable pod should be before starting scale up.
	unschedulablePodTimeBuffer = 2 * time.Second
	// How old the oldest unschedulable pod with GPU should be before starting scale up.
	// The idea is that nodes with GPU are very expensive and we're ready to sacrifice
	// a bit more latency to wait for more pods and make a more informed scale-up decision.
	unschedulablePodWithGpuTimeBuffer = 30 * time.Second
	// How long should Cluster Autoscaler wait for nodes to become ready after start.
	nodesNotReadyAfterStartTimeout = 10 * time.Minute

	valueTrue = "true"
)

var nodeConditionTaints = taintKeySet{
	schedulerapi.TaintNodeNotReady:           true,
	schedulerapi.TaintNodeUnreachable:        true,
	schedulerapi.TaintNodeUnschedulable:      true,
	schedulerapi.TaintNodeMemoryPressure:     true,
	schedulerapi.TaintNodeDiskPressure:       true,
	schedulerapi.TaintNodeNetworkUnavailable: true,
	schedulerapi.TaintNodePIDPressure:        true,
	schedulerapi.TaintExternalCloudProvider:  true,
	schedulerapi.TaintNodeShutdown:           true,
	gkeNodeTerminationHandlerTaint:           true,
}

var nodeInfoCacheExpiredTime = 10 * time.Minute

type cacheItem struct {
	*schedulernodeinfo.NodeInfo
	added time.Time
}

func isCacheItemExpired(added time.Time) bool {
	return time.Since(added) > nodeInfoCacheExpiredTime
}

// Following data structure is used to avoid running predicates #pending_pods * #nodes
// times (which turned out to be very expensive if there are thousands of pending pods).
// This optimization is based on the assumption that if there are that many pods they're
// likely created by controllers (deployment, replication controller, ...).
// So instead of running all predicates for every pod we first check whether we've
// already seen identical pod (in this step we're not binpacking, just checking if
// the pod would fit anywhere right now) and if so we use the result we already
// calculated.
// To decide if two pods are similar enough we check if they have identical label
// and spec and are owned by the same controller. The problem is the whole
// podSchedulableInfo struct is not hashable and keeping a list and running deep
// equality checks would likely also be expensive. So instead we use controller
// UID as a key in initial lookup and only run full comparison on a set of
// podSchedulableInfos created for pods owned by this controller.
type podSchedulableInfo struct {
	spec            apiv1.PodSpec
	labels          map[string]string
	schedulingError *simulator.PredicateError
}

type podSchedulableMap map[string][]podSchedulableInfo

type taintKeySet map[string]bool

func (psi *podSchedulableInfo) match(pod *apiv1.Pod) bool {
	return reflect.DeepEqual(pod.Labels, psi.labels) && apiequality.Semantic.DeepEqual(pod.Spec, psi.spec)
}

func (podMap podSchedulableMap) get(pod *apiv1.Pod) (*simulator.PredicateError, bool) {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return nil, false
	}
	uid := string(ref.UID)
	if infos, found := podMap[uid]; found {
		for _, info := range infos {
			if info.match(pod) {
				return info.schedulingError, true
			}
		}
	}
	return nil, false
}

func (podMap podSchedulableMap) set(pod *apiv1.Pod, err *simulator.PredicateError) {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return
	}
	uid := string(ref.UID)
	podMap[uid] = append(podMap[uid], podSchedulableInfo{
		spec:            pod.Spec,
		labels:          pod.Labels,
		schedulingError: err,
	})
}

// filterOutExpendableAndSplit filters out expendable pods and splits into:
//   - waiting for lower priority pods preemption
//   - other pods.
func filterOutExpendableAndSplit(unschedulableCandidates []*apiv1.Pod,
	expendablePodsPriorityCutoff int,
) ([]*apiv1.Pod, []*apiv1.Pod) {
	var unschedulableNonExpendable []*apiv1.Pod
	var waitingForLowerPriorityPreemption []*apiv1.Pod
	for _, pod := range unschedulableCandidates {
		if pod.Spec.Priority != nil && int(*pod.Spec.Priority) < expendablePodsPriorityCutoff {
			klog.V(4).Infof("Pod %s has priority below %d (%d) and will scheduled when enough resources is free."+
				" Ignoring in scale up.", pod.Name, expendablePodsPriorityCutoff, *pod.Spec.Priority)
		} else if nominatedNodeName := pod.Status.NominatedNodeName; nominatedNodeName != "" {
			waitingForLowerPriorityPreemption = append(waitingForLowerPriorityPreemption, pod)
			klog.V(4).Infof("Pod %s will be scheduled after low priority pods are preempted on %s. Ignoring in scale up.",
				pod.Name, nominatedNodeName)
		} else {
			unschedulableNonExpendable = append(unschedulableNonExpendable, pod)
		}
	}
	return unschedulableNonExpendable, waitingForLowerPriorityPreemption
}

// filterOutExpendablePods filters out expendable pods.
func filterOutExpendablePods(pods []*apiv1.Pod, expendablePodsPriorityCutoff int) []*apiv1.Pod {
	var result []*apiv1.Pod
	for _, pod := range pods {
		if pod.Spec.Priority == nil || int(*pod.Spec.Priority) >= expendablePodsPriorityCutoff {
			result = append(result, pod)
		}
	}
	return result
}

type equivalenceGroupID int

type podsSchedulableOnNodeChecker struct {
	context               *context.AutoscalingContext
	pods                  []*apiv1.Pod
	podsEquivalenceGroups map[types.UID]equivalenceGroupID
}

func newPodsSchedulableOnNodeChecker(context *context.AutoscalingContext,
	pods []*apiv1.Pod,
) *podsSchedulableOnNodeChecker {
	checker := podsSchedulableOnNodeChecker{
		context:               context,
		pods:                  pods,
		podsEquivalenceGroups: make(map[types.UID]equivalenceGroupID),
	}

	// compute the podsEquivalenceGroups
	var nextGroupID equivalenceGroupID
	type equivalanceGroup struct {
		id           equivalenceGroupID
		representant *apiv1.Pod
	}

	equivalenceGroupsByController := make(map[types.UID][]equivalanceGroup)

	for _, pod := range pods {
		controllerRef := drain.ControllerRef(pod)
		if controllerRef == nil {
			checker.podsEquivalenceGroups[pod.UID] = nextGroupID
			nextGroupID++
			continue
		}

		matchingFound := false
		for _, g := range equivalenceGroupsByController[controllerRef.UID] {
			if reflect.DeepEqual(pod.Labels, g.representant.Labels) && apiequality.Semantic.DeepEqual(
				pod.Spec, g.representant.Spec) {
				matchingFound = true
				checker.podsEquivalenceGroups[pod.UID] = g.id
				break
			}
		}

		if !matchingFound {
			newGroup := equivalanceGroup{
				id:           nextGroupID,
				representant: pod,
			}
			equivalenceGroupsByController[controllerRef.UID] = append(equivalenceGroupsByController[controllerRef.UID], newGroup)
			checker.podsEquivalenceGroups[pod.UID] = newGroup.id
			nextGroupID++
		}
	}

	return &checker
}

// checkPodsSchedulableOnNode checks if pods can be scheduled on the given node.
func (c *podsSchedulableOnNodeChecker) checkPodsSchedulableOnNode(nodeGroupID string,
	nodeInfo *schedulernodeinfo.NodeInfo,
) map[*apiv1.Pod]*simulator.PredicateError {
	loggingQuota := glogx.PodsLoggingQuota()
	schedulingErrors := make(map[equivalenceGroupID]*simulator.PredicateError)
	for _, pod := range c.pods {
		equivalenceGroup := c.podsEquivalenceGroups[pod.UID]
		err, found := schedulingErrors[equivalenceGroup]
		if found && err != nil {
			glogx.V(2).UpTo(loggingQuota).Infof("Pod %s can't be scheduled on %s. Used cached predicate check results",
				pod.Name, nodeGroupID)
		}
		// Not found in cache, have to run the predicates.
		if !found {
			err = c.context.PredicateChecker.CheckPredicates(pod, nil, nodeInfo)
			schedulingErrors[equivalenceGroup] = err
			if err != nil {
				// Always log for the first pod in a controller.
				klog.V(2).Infof("Pod %s can't be scheduled on %s, predicate failed: %v", pod.Name, nodeGroupID, err.VerboseError())
			}
		}
	}
	glogx.V(2).Over(loggingQuota).Infof("%v other pods can't be scheduled on %s.", -loggingQuota.Left(), nodeGroupID)

	schedulingErrorsByPod := make(map[*apiv1.Pod]*simulator.PredicateError)
	for _, pod := range c.pods {
		schedulingErrorsByPod[pod] = schedulingErrors[c.podsEquivalenceGroups[pod.UID]]
	}
	return schedulingErrorsByPod
}

// getNodeInfosForGroups finds NodeInfos for all node groups used to manage the given nodes.
// It also returns a node group to sample node mapping.
// DOTO(mwielgus): This returns map keyed by url, while most code (including scheduler) uses node.Name for a key.
// DOTO(mwielgus): Review error policy - sometimes we may continue with partial errors.
// nolint
func getNodeInfosForGroups(nodes []*apiv1.Node, nodeInfoCache map[string]cacheItem,
	cloudProvider cloudprovider.CloudProvider, listers kube_util.ListerRegistry,
	daemonsets []*appsv1.DaemonSet, predicateChecker *simulator.PredicateChecker,
	ignoredTaints taintKeySet,
) (map[string]*schedulernodeinfo.NodeInfo, errors.AutoscalerError) {
	result := make(map[string]*schedulernodeinfo.NodeInfo)
	seenGroups := make(map[string]bool)

	podsForNodes, err := getPodsForNodes(listers)
	if err != nil {
		return map[string]*schedulernodeinfo.NodeInfo{}, err
	}

	// processNode returns information whether the nodeTemplate was generated and if there was an error.
	processNode := func(node *apiv1.Node) (bool, string, errors.AutoscalerError) {
		nodeGroup, getErr := cloudProvider.NodeGroupForNode(node)
		if getErr != nil {
			return false, "", errors.ToAutoscalerError(errors.CloudProviderError, getErr)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			return false, "", nil
		}
		id := nodeGroup.Id()
		if _, found := result[id]; !found {
			// Build nodeInfo.
			nodeInfo, buildErr := simulator.BuildNodeInfoForNode(node, podsForNodes)
			if buildErr != nil {
				return false, "", buildErr
			}
			sanitizedNodeInfo, sanErr := sanitizeNodeInfo(nodeInfo, id, ignoredTaints)
			if sanErr != nil {
				return false, "", sanErr
			}
			result[id] = sanitizedNodeInfo
			return true, id, nil
		}
		// if founded in result, but not found in cache, should add to node info cache
		if nodeInfoCache != nil {
			if _, found := nodeInfoCache[id]; !found {
				return true, id, nil
			}
		}
		return false, "", nil
	}

	seenGroups, result, err = generateNodeInfos(cloudProvider, nodeInfoCache, daemonsets,
		predicateChecker, ignoredTaints, seenGroups, result)
	if err != nil {
		return map[string]*schedulernodeinfo.NodeInfo{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}

	// if cannot get node template from provider, try to generate with real-world example
	for _, node := range nodes {
		// Broken nodes might have some stuff missing. Skipping.
		if !kube_util.IsNodeReadyAndSchedulable(node) {
			continue
		}
		added, id, typedErr := processNode(node)
		if typedErr != nil {
			return map[string]*schedulernodeinfo.NodeInfo{}, typedErr
		}
		// DOTO: support node pool label update
		if added && nodeInfoCache != nil {
			if nodeInfoCopy, err := deepCopyNodeInfo(result[id]); err == nil {
				nodeInfoCache[id] = cacheItem{NodeInfo: nodeInfoCopy, added: time.Now()}
			}
		}
	}

	// Remove invalid node groups from cache
	for id := range nodeInfoCache {
		if _, ok := seenGroups[id]; !ok {
			delete(nodeInfoCache, id)
		}
	}

	// Last resort - unready/unschedulable nodes.
	for _, node := range nodes {
		// Allowing broken nodes
		if !kube_util.IsNodeReadyAndSchedulable(node) {
			added, _, typedErr := processNode(node)
			if typedErr != nil {
				return map[string]*schedulernodeinfo.NodeInfo{}, typedErr
			}
			nodeGroup, err := cloudProvider.NodeGroupForNode(node)
			if err != nil {
				return map[string]*schedulernodeinfo.NodeInfo{}, errors.ToAutoscalerError(
					errors.CloudProviderError, err)
			}
			if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
				klog.Infof("No node group for node %s, skipping", node.Name)
				continue
			}
			if added {
				klog.Warningf("Built template for %s based on unready/unschedulable node %s", nodeGroup.Id(), node.Name)
			}
		}
	}

	return result, nil
}

func generateNodeInfos(cloudProvider cloudprovider.CloudProvider,
	nodeInfoCache map[string]cacheItem,
	daemonsets []*appsv1.DaemonSet, predicateChecker *simulator.PredicateChecker,
	ignoredTaints taintKeySet, seenGroups map[string]bool,
	result map[string]*schedulernodeinfo.NodeInfo) (map[string]bool,
	map[string]*schedulernodeinfo.NodeInfo, errors.AutoscalerError,
) {
	for _, nodeGroup := range cloudProvider.NodeGroups() {
		id := nodeGroup.Id()
		seenGroups[id] = true
		if _, found := result[id]; found {
			continue
		}

		// No good template, check cache of previously running nodes.
		if nodeInfoCache != nil {
			if item, found := nodeInfoCache[id]; found {
				if isCacheItemExpired(item.added) {
					delete(nodeInfoCache, id)
				} else if nodeInfoCopy, err := deepCopyNodeInfo(item.NodeInfo); err == nil {
					result[id] = nodeInfoCopy
					continue
				}
			}
		}

		// No good template, trying to generate one.
		nodeInfo, err := getNodeInfoFromTemplate(nodeGroup, daemonsets, predicateChecker, ignoredTaints)
		if err != nil {
			if err == cloudprovider.ErrNotImplemented {
				continue
			} else {
				klog.Errorf("Unable to build proper template node for %s: %v", id, err)
				return seenGroups, result, errors.ToAutoscalerError(errors.CloudProviderError, err)
			}
		}
		result[id] = nodeInfo
	}
	return seenGroups, result, nil
}

// getNodeInfos finds NodeInfos for all nodes
func getNodeInfos(listers kube_util.ListerRegistry) (map[string]*schedulernodeinfo.NodeInfo, errors.AutoscalerError) {
	// results := make(map[string]*schedulernodeinfo.NodeInfo)

	podsForNodes, typeErr := getPodsForNodes(listers)
	if typeErr != nil {
		return map[string]*schedulernodeinfo.NodeInfo{}, typeErr
	}
	klog.V(4).Infof("Node num: %v", len(podsForNodes))
	nodes, err := listers.ReadyNodeLister().List()
	if err != nil {
		return map[string]*schedulernodeinfo.NodeInfo{}, errors.NewAutoscalerError(
			errors.InternalError, err.Error())
	}
	results := buildNodeInfos(nodes, podsForNodes)
	return results, nil
}

func getPodsForNodes(listers kube_util.ListerRegistry) (map[string][]*apiv1.Pod, errors.AutoscalerError) {
	pods, err := listers.ScheduledPodLister().List()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	podsForNodes := map[string][]*apiv1.Pod{}
	for _, p := range pods {
		podsForNodes[p.Spec.NodeName] = append(podsForNodes[p.Spec.NodeName], p)
	}
	return podsForNodes, nil
}

// getNodeInfoFromTemplate returns NodeInfo object built base on TemplateNodeInfo
// returned by NodeGroup.TemplateNodeInfo().
func getNodeInfoFromTemplate(nodeGroup cloudprovider.NodeGroup, daemonsets []*appsv1.DaemonSet,
	predicateChecker *simulator.PredicateChecker, ignoredTaints taintKeySet) (*schedulernodeinfo.NodeInfo,
	errors.AutoscalerError,
) {
	id := nodeGroup.Id()
	baseNodeInfo, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}

	pods := daemonset.GetDaemonSetPodsForNode(baseNodeInfo, daemonsets, predicateChecker)
	pods = append(pods, baseNodeInfo.Pods()...)
	fullNodeInfo := schedulernodeinfo.NewNodeInfo(pods...)
	err = fullNodeInfo.SetNode(baseNodeInfo.Node())
	if err != nil {
		klog.Errorf("SetNode failed. Error: %v", err)
	}
	sanitizedNodeInfo, typedErr := sanitizeNodeInfo(fullNodeInfo, id, ignoredTaints)
	if typedErr != nil {
		return nil, typedErr
	}
	return sanitizedNodeInfo, nil
}

// filterOutNodesFromNotAutoscaledGroups return subset of input nodes for which cloud provider does not
// return autoscaled node group.
// NOCC:tosa/fn_length(设计如此)
func filterOutNodesFromNotAutoscaledGroups(nodes []*apiv1.Node, cloudProvider cloudprovider.CloudProvider) (
	[]*apiv1.Node, errors.AutoscalerError,
) {
	result := make([]*apiv1.Node, 0)

	for _, node := range nodes {
		nodeGroup, err := cloudProvider.NodeGroupForNode(node)
		if err != nil {
			return []*apiv1.Node{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			result = append(result, node)
		}
	}
	return result, nil
}

func deepCopyNodeInfo(nodeInfo *schedulernodeinfo.NodeInfo) (*schedulernodeinfo.NodeInfo, errors.AutoscalerError) {
	newPods := make([]*apiv1.Pod, 0)
	for _, pod := range nodeInfo.Pods() {
		newPods = append(newPods, pod.DeepCopy())
	}

	// Build a new node info.
	newNodeInfo := schedulernodeinfo.NewNodeInfo(newPods...)
	if err := newNodeInfo.SetNode(nodeInfo.Node().DeepCopy()); err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return newNodeInfo, nil
}

func sanitizeNodeInfo(nodeInfo *schedulernodeinfo.NodeInfo, nodeGroupName string,
	ignoredTaints taintKeySet,
) (*schedulernodeinfo.NodeInfo, errors.AutoscalerError) {
	// Sanitize node name.
	sanitizedNode, err := sanitizeTemplateNode(nodeInfo.Node(), nodeGroupName, ignoredTaints)
	if err != nil {
		return nil, err
	}

	// Update nodename in pods.
	sanitizedPods := make([]*apiv1.Pod, 0)
	for _, pod := range nodeInfo.Pods() {
		sanitizedPod := pod.DeepCopy()
		sanitizedPod.Spec.NodeName = sanitizedNode.Name
		sanitizedPods = append(sanitizedPods, sanitizedPod)
	}

	// Build a new node info.
	sanitizedNodeInfo := schedulernodeinfo.NewNodeInfo(sanitizedPods...)
	if err := sanitizedNodeInfo.SetNode(sanitizedNode); err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return sanitizedNodeInfo, nil
}

// nolint result 1 (error) is always nil
func sanitizeTemplateNode(node *apiv1.Node, nodeGroup string,
	ignoredTaints taintKeySet,
) (*apiv1.Node, errors.AutoscalerError) {
	newNode := node.DeepCopy()
	nodeName := fmt.Sprintf("template-node-for-%s-%d", nodeGroup, rand.Int63()) // nolint
	newNode.Labels = make(map[string]string, len(node.Labels))
	for k, v := range node.Labels {
		// if !validLabel(k) {
		// 	continue
		// }
		if k != apiv1.LabelHostname {
			newNode.Labels[k] = v
		} else {
			newNode.Labels[k] = nodeName
		}
	}
	newNode.Name = nodeName
	newTaints := make([]apiv1.Taint, 0)
	for _, taint := range node.Spec.Taints {
		// Rescheduler can put this taint on a node while evicting non-critical pods.
		// New nodes will not have this taint and so we should strip it when creating
		// template node.
		switch taint.Key {
		case ReschedulerTaintKey:
			klog.V(4).Infof("Removing rescheduler taint when creating template from node %s", node.Name)
			continue
		case deletetaint.ToBeDeletedTaint:
			klog.V(4).Infof("Removing autoscaler taint when creating template from node %s", node.Name)
			continue
		case deletetaint.DeletionCandidateTaint:
			klog.V(4).Infof("Removing autoscaler soft taint when creating template from node %s", node.Name)
			continue
		}

		// ignore conditional taints as they represent a transient node state.
		if exists := nodeConditionTaints[taint.Key]; exists {
			klog.V(4).Infof("Removing node condition taint %s, when creating template from node %s", taint.Key, node.Name)
			continue
		}

		if exists := ignoredTaints[taint.Key]; exists {
			klog.V(4).Infof("Removing ignored taint %s, when creating template from node %s", taint.Key, node.Name)
			continue
		}

		newTaints = append(newTaints, taint)
	}
	newNode.Spec.Taints = newTaints
	return newNode, nil
}

// removeOldUnregisteredNodes removes unregistered nodes if needed. Returns true if anything
// was removed and error if such occurred.
func removeOldUnregisteredNodes(unregisteredNodes []clusterstate.UnregisteredNode, context *context.AutoscalingContext,
	currentTime time.Time, logRecorder *utils.LogEventRecorder,
) (bool, error) {
	removedAny := false
	for _, unregisteredNode := range unregisteredNodes {
		if unregisteredNode.UnregisteredSince.Add(context.MaxNodeProvisionTime).Before(currentTime) {
			klog.V(0).Infof("Removing unregistered node %v", unregisteredNode.Node.Name)
			nodeGroup, err := context.CloudProvider.NodeGroupForNode(unregisteredNode.Node)
			if err != nil {
				klog.Warningf("Failed to get node group for %s: %v", unregisteredNode.Node.Name, err)
				return removedAny, err
			}
			if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
				klog.Warningf("No node group for node %s, skipping", unregisteredNode.Node.Name)
				continue
			}
			size, err := nodeGroup.TargetSize()
			if err != nil {
				klog.Warningf("Failed to get node group size; unregisteredNode=%v; nodeGroup=%v; err=%v",
					unregisteredNode.Node.Name, nodeGroup.Id(), err)
				continue
			}
			if nodeGroup.MinSize() >= size {
				klog.Warningf("Failed to remove node %s: node group min size reached, skipping unregistered node removal",
					unregisteredNode.Node.Name)
				continue
			}
			err = nodeGroup.DeleteNodes([]*apiv1.Node{unregisteredNode.Node})
			if err != nil {
				klog.Warningf("Failed to remove node %s: %v", unregisteredNode.Node.Name, err)
				logRecorder.Eventf(apiv1.EventTypeWarning, "DeleteUnregisteredFailed",
					"Failed to remove node %s: %v", unregisteredNode.Node.Name, err)
				return removedAny, err
			}
			logRecorder.Eventf(apiv1.EventTypeNormal, "DeleteUnregistered",
				"Removed unregistered node %v", unregisteredNode.Node.Name)
			metrics.RegisterOldUnregisteredNodesRemoved(1)
			removedAny = true
		}
	}
	return removedAny, nil
}

// fixNodeGroupSize sets the target size of node groups to the current number of nodes in them
// if the difference was constant for a prolonged time. Returns true if managed
// to fix something.
func fixNodeGroupSize(context *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry,
	currentTime time.Time,
) (bool, error) {
	fixed := false
	for _, nodeGroup := range context.CloudProvider.NodeGroups() {
		incorrectSize := clusterStateRegistry.GetIncorrectNodeGroupSize(nodeGroup.Id())
		if incorrectSize == nil {
			continue
		}
		// set the target size of node groups to the current number of nodes
		// may descrease or increase the target size
		if incorrectSize.FirstObserved.Add(context.MaxNodeProvisionTime).Before(currentTime) {
			delta := incorrectSize.CurrentSize - incorrectSize.ExpectedSize
			klog.V(0).Infof("Fix size of %s, expected=%d current=%d delta=%d", nodeGroup.Id(),
				incorrectSize.ExpectedSize, incorrectSize.CurrentSize, delta)
			if err := nodeGroup.DecreaseTargetSize(delta); err != nil {
				return fixed, fmt.Errorf("failed to fix %s: %v", nodeGroup.Id(), err)
			}
			fixed = true
		}
	}
	return fixed, nil
}

func hasHardInterPodAffinity(affinity *apiv1.Affinity) bool {
	if affinity == nil {
		return false
	}
	if affinity.PodAffinity != nil {
		if len(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
			return true
		}
	}
	if affinity.PodAntiAffinity != nil {
		if len(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
			return true
		}
	}
	return false
}

func anyPodHasHardInterPodAffinity(pods []*apiv1.Pod) bool {
	for _, pod := range pods {
		if hasHardInterPodAffinity(pod.Spec.Affinity) {
			return true
		}
	}
	return false
}

// ConfigurePredicateCheckerForLoop can be run to update predicateChecker configuration
// based on current state of the cluster.
func ConfigurePredicateCheckerForLoop(unschedulablePods []*apiv1.Pod, schedulablePods []*apiv1.Pod,
	predicateChecker *simulator.PredicateChecker,
) {
	podsWithAffinityFound := anyPodHasHardInterPodAffinity(unschedulablePods)
	if !podsWithAffinityFound {
		podsWithAffinityFound = anyPodHasHardInterPodAffinity(schedulablePods)
	}
	predicateChecker.SetAffinityPredicateEnabled(podsWithAffinityFound)
	if !podsWithAffinityFound {
		klog.V(1).Info("No pod using affinity / antiaffinity found in cluster, disabling affinity predicate for this loop")
	}
}

func getNodeCoresAndMemory(node *apiv1.Node) (int64, int64) {
	// filter eklet node
	if node.Labels[nodeInstanceTypeLabelKey] == nodeInstanceTypeEklet {
		return 0, 0
	}
	if node.Annotations[filterNodeResourceAnnoKey] == valueTrue {
		return 0, 0
	}
	cores := getNodeResource(node, apiv1.ResourceCPU)
	memory := getNodeResource(node, apiv1.ResourceMemory)
	return cores, memory
}

func getNodeResource(node *apiv1.Node, resource apiv1.ResourceName) int64 {
	nodeCapacity, found := node.Status.Capacity[resource]
	if !found {
		return 0
	}

	nodeCapacityValue := nodeCapacity.Value()
	if nodeCapacityValue < 0 {
		nodeCapacityValue = 0
	}

	return nodeCapacityValue
}

// UpdateClusterStateMetrics updates metrics related to cluster state
func UpdateClusterStateMetrics(csr *clusterstate.ClusterStateRegistry) {
	if csr == nil || reflect.ValueOf(csr).IsNil() {
		return
	}
	metrics.UpdateClusterSafeToAutoscale(csr.IsClusterHealthy())
	readiness := csr.GetClusterReadiness()
	// fix(bcs): 删除中节点也应统计到指标中, 为减少改动, 添加到 Unregistered 中
	metrics.UpdateNodesCount(readiness.Ready, readiness.Unready+readiness.LongNotStarted, readiness.NotStarted,
		readiness.LongUnregistered, readiness.Unregistered+readiness.Deleted)
}

func getOldestCreateTime(pods []*apiv1.Pod) time.Time {
	oldest := time.Now()
	for _, pod := range pods {
		if oldest.After(pod.CreationTimestamp.Time) {
			oldest = pod.CreationTimestamp.Time
		}
	}
	return oldest
}

func getOldestCreateTimeWithGpu(pods []*apiv1.Pod) (bool, time.Time) {
	oldest := time.Now()
	gpuFound := false
	for _, pod := range pods {
		if gpu.PodRequestsGpu(pod) {
			gpuFound = true
			if oldest.After(pod.CreationTimestamp.Time) {
				oldest = pod.CreationTimestamp.Time
			}
		}
	}
	return gpuFound, oldest
}

// updateEmptyClusterStateMetrics updates metrics related to empty cluster's state.
// DOTO(aleksandra-malinowska): use long unregistered value from ClusterStateRegistry.
func updateEmptyClusterStateMetrics() {
	metrics.UpdateClusterSafeToAutoscale(false)
	metrics.UpdateNodesCount(0, 0, 0, 0, 0)
}

func allPodsAreNew(pods []*apiv1.Pod, currentTime time.Time) bool {
	if getOldestCreateTime(pods).Add(unschedulablePodTimeBuffer).After(currentTime) {
		return true
	}
	found, oldest := getOldestCreateTimeWithGpu(pods)
	return found && oldest.Add(unschedulablePodWithGpuTimeBuffer).After(currentTime)
}

func buildNodeForNodeTemplate(nodeTemplate *schedulernodeinfo.NodeInfo, index int) *apiv1.Node {
	node := nodeTemplate.Node().DeepCopy()
	node.Name = fmt.Sprintf("%s-%d", node.Name, index)
	node.UID = uuid.NewUUID()
	return node
}

func getUpcomingNodeInfos(registry *clusterstate.ClusterStateRegistry,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
) []*apiv1.Node {
	upcomingNodes := make([]*apiv1.Node, 0)
	for nodeGroup, numberOfNodes := range registry.GetUpcomingNodes() {
		nodeTemplate, found := nodeInfos[nodeGroup]
		if !found {
			klog.Warningf("Couldn't find template for node group %s", nodeGroup)
			continue
		}
		for i := 0; i < numberOfNodes; i++ {
			// Ensure new nodes having different names because nodeName would used as a map key.
			upcomingNodes = append(upcomingNodes, buildNodeForNodeTemplate(nodeTemplate, i))
		}
	}
	return upcomingNodes
}

func checkResourceNotEnough(nodes map[string]*schedulernodeinfo.NodeInfo,
	podsToReschedule []*apiv1.Pod, cpuRatio, memRatio, ratio float64,
) bool {
	var leftResourcesList, sumResourcesList schedulernodeinfo.Resource
	for _, nodeInfo := range nodes {
		node := nodeInfo.Node()
		if node == nil {
			continue
		}
		if node.Spec.Unschedulable {
			continue
		}
		if node.Labels["node.kubernetes.io/instance-type"] == "eklet" {
			continue
		}
		if node.Annotations[filterNodeResourceAnnoKey] == valueTrue {
			continue
		}
		if node.Labels[nodeLabel.LabelNodeRoleMaster] == valueTrue {
			continue
		}
		klog.V(6).Infof("resource: %+v", node.Status.Allocatable)
		allocatable := nodeInfo.AllocatableResource()
		sumResourcesList.Add(allocatable.ResourceList())
		leftResourcesList.Add(singleNodeResource(nodeInfo).ResourceList())
	}

	if len(podsToReschedule) > 0 {
		leftResourcesList = substractRescheduledPodResources(leftResourcesList,
			podsToReschedule)
	}

	leftResources := leftResourcesList.ResourceList()
	sumResources := sumResourcesList.ResourceList()
	for name, sum := range sumResources {
		left, ok := leftResources[name]
		if !ok {
			continue
		}
		if sum.IsZero() {
			continue
		}
		r := float64(left.MilliValue()) / float64(sum.MilliValue())
		metricsinternal.UpdateResourceUsedRatio(name.String(), 1-r)
		switch name {
		case apiv1.ResourceCPU:
			klog.V(4).Infof("%v ratio %v, desired CPU ratio %v", name, r, cpuRatio)
			if r < cpuRatio {
				return true
			}
		case apiv1.ResourceMemory:
			klog.V(4).Infof("%v ratio %v, desired Memory ratio %v", name, r, memRatio)
			if r < memRatio {
				return true
			}
		default:
			klog.V(4).Infof("%v ratio %v, desired ratio %v", name, r, ratio)
			if r < ratio {
				return true
			}
		}
	}
	return false
}

func substractRescheduledPodResources(leftResourcesList schedulernodeinfo.Resource,
	podsToReschedule []*apiv1.Pod,
) schedulernodeinfo.Resource {
	var podResources schedulernodeinfo.Resource
	for _, pod := range podsToReschedule {
		for _, container := range pod.Spec.Containers {
			podResources.Add(container.Resources.Requests)
		}
	}

	leftResourcesList.AllowedPodNumber -= len(podsToReschedule)
	leftResourcesList.MilliCPU -= podResources.MilliCPU
	leftResourcesList.Memory -= podResources.Memory
	leftResourcesList.EphemeralStorage -= podResources.EphemeralStorage

	// calculate extend resources
	for k, v := range podResources.ScalarResources {
		_, ok := leftResourcesList.ScalarResources[k]
		if ok {
			leftResourcesList.ScalarResources[k] -= v
		}
	}

	return leftResourcesList
}

func singleNodeResource(info *schedulernodeinfo.NodeInfo) *schedulernodeinfo.Resource {
	leftResource := schedulernodeinfo.Resource{
		ScalarResources: make(map[apiv1.ResourceName]int64),
	}

	allocatable := info.AllocatableResource()
	requested := info.RequestedResource()

	podCount := requested.AllowedPodNumber
	if podCount == 0 {
		podCount = len(info.Pods())
	}

	leftResource.AllowedPodNumber = allocatable.AllowedPodNumber - podCount
	leftResource.MilliCPU = allocatable.MilliCPU - requested.MilliCPU
	leftResource.Memory = allocatable.Memory - requested.Memory
	leftResource.EphemeralStorage = allocatable.EphemeralStorage - requested.EphemeralStorage

	// calculate extend resources
	for k, allocatableEx := range allocatable.ScalarResources {
		requestEx, ok := requested.ScalarResources[k]
		if !ok {
			leftResource.ScalarResources[k] = allocatableEx
		} else {
			leftResource.ScalarResources[k] = allocatableEx - requestEx
		}
	}
	klog.Infof("Node %v left resource %+v", info.Node().Name, leftResource)
	return &leftResource
}

func getPodsInfos(pods []*apiv1.Pod) map[string][]*apiv1.Pod {
	podsForNodes := map[string][]*apiv1.Pod{}
	for _, p := range pods {
		podsForNodes[p.Spec.NodeName] = append(podsForNodes[p.Spec.NodeName], p)
	}
	return podsForNodes
}

func buildNodeInfos(nodes []*apiv1.Node, podsForNodes map[string][]*apiv1.Pod) map[string]*schedulernodeinfo.NodeInfo {
	results := make(map[string]*schedulernodeinfo.NodeInfo)
	for _, node := range nodes {
		if node.Spec.Unschedulable {
			continue
		}
		nodeCopy := node.DeepCopy()
		requiredPods := podsForNodes[nodeCopy.Name]
		result := schedulernodeinfo.NewNodeInfo(requiredPods...)
		if err := result.SetNode(nodeCopy); err != nil {
			return nil
		}
		results[node.Name] = result
	}
	return results
}
