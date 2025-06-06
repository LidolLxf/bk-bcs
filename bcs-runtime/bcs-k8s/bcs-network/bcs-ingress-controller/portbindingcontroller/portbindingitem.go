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

package portbindingcontroller

import (
	"context"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sapitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

type portBindingItemHandler struct {
	ctx       context.Context
	k8sClient client.Client
}

func newPortBindingItemHandler(ctx context.Context, k8sClient client.Client) *portBindingItemHandler {
	return &portBindingItemHandler{
		ctx:       ctx,
		k8sClient: k8sClient,
	}
}

func (pbih *portBindingItemHandler) ensureItem(
	portBinding *networkextensionv1.PortBinding, tmpTargetGroup *networkextensionv1.ListenerTargetGroup,
	item *networkextensionv1.PortBindingItem, itemStatus *networkextensionv1.
		PortBindingStatusItem) *networkextensionv1.PortBindingStatusItem {
	// when status is empty, just return initializing status
	if itemStatus == nil {
		return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusInitializing)
	}
	// update listener
	portPool := &networkextensionv1.PortPool{}
	if err := pbih.k8sClient.Get(pbih.ctx, k8sapitypes.NamespacedName{
		Name:      item.PoolName,
		Namespace: item.PoolNamespace,
	}, portPool); err != nil {
		blog.Warnf("failed to get port pool %s/%s failed, err %s", item.PoolName, item.PoolNamespace, err.Error())
		return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusInitializing)
	}

	countReady := 0
	for _, lbObj := range item.PoolItemLoadBalancers {
		listenerName := common.GetListenerNameWithProtocol(
			lbObj.LoadbalancerID, item.Protocol, item.StartPort, item.EndPort)
		rawListener := &networkextensionv1.Listener{}
		if err := pbih.k8sClient.Get(context.Background(), k8sapitypes.NamespacedName{
			Name:      listenerName,
			Namespace: item.PoolNamespace,
		}, rawListener); err != nil {
			blog.Warnf("failed to get listener %s/%s, err %s", listenerName, item.PoolNamespace, err.Error())
			return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusInitializing)
		}

		// do not update informer cache directly
		listener := rawListener.DeepCopy()
		// listener has targetGroup
		if listener.Spec.TargetGroup != nil && len(listener.Spec.TargetGroup.Backends) != 0 {
			// listener has targetGroup and targetGroup(include pod ip) has no changed
			if reflect.DeepEqual(listener.Spec.TargetGroup, tmpTargetGroup) {
				// listener has not synced
				if listener.Status.Status != networkextensionv1.ListenerStatusSynced {
					blog.V(4).Infof("listener %s/%s changes not synced", listenerName, item.PoolNamespace)
					return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusNotReady)
				}
				
				// if targetGroup is same and status is synced, listener is ready
				countReady++
				continue
			}
			// listener has targetGroup but targetGroup(include pod ip) has changed
		}

		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			li := &networkextensionv1.Listener{}
			if err := pbih.k8sClient.Get(context.Background(), k8sapitypes.NamespacedName{
				Namespace: item.PoolNamespace,
				Name:      listenerName,
			}, li); err != nil {
				return err
			}

			cpListener := li.DeepCopy()
			// listener has no targetGroup or ip has changed
			cpListener.Spec.ListenerAttribute = portPool.Spec.ListenerAttribute
			if item.ListenerAttribute != nil {
				cpListener.Spec.ListenerAttribute = item.ListenerAttribute
			}
			cpListener.Status.Status = networkextensionv1.ListenerStatusNotSynced
			cpListener.Spec.TargetGroup = tmpTargetGroup
			if cpListener.Labels == nil {
				cpListener.Labels = make(map[string]string)
			}
			cpListener.Labels[networkextensionv1.LabelKeyForSourceNamespace] = portBinding.GetNamespace()
			cpListener.Labels[networkextensionv1.LabelKeyForSourceName] = portBinding.GetName()

			if err := pbih.k8sClient.Update(context.Background(), cpListener, &client.UpdateOptions{}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			blog.Warnf("failed to update listener %s/%s, err %s", listenerName, item.PoolNamespace, err.Error())
			return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusInitializing)
		}
		blog.V(3).Infof("update listener %s/%s successfully", listenerName, item.PoolNamespace)
	}
	if countReady == len(item.PoolItemLoadBalancers) && countReady != 0 {
		return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusReady)
	}
	return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusNotReady)
}

func (pbih *portBindingItemHandler) generateStatus(
	item *networkextensionv1.PortBindingItem, itemStatus *networkextensionv1.PortBindingStatusItem,
	status string) *networkextensionv1.PortBindingStatusItem {
	uptimeCheckStatus := &networkextensionv1.UptimeCheckTaskStatus{}
	if itemStatus != nil {
		uptimeCheckStatus = itemStatus.UptimeCheckStatus
	}
	return &networkextensionv1.PortBindingStatusItem{
		PoolName:          item.PoolName,
		PoolNamespace:     item.PoolNamespace,
		PoolItemName:      item.PoolItemName,
		StartPort:         item.StartPort,
		EndPort:           item.EndPort,
		Status:            status,
		UptimeCheckStatus: uptimeCheckStatus,
	}
}

func (pbih *portBindingItemHandler) deleteItem(
	item *networkextensionv1.PortBindingItem, itemStatus *networkextensionv1.PortBindingStatusItem) *networkextensionv1.PortBindingStatusItem {
	for _, lbObj := range item.PoolItemLoadBalancers {
		listenerName := common.GetListenerNameWithProtocol(
			lbObj.LoadbalancerID, item.Protocol, item.StartPort, item.EndPort)
		rawListener := &networkextensionv1.Listener{}
		if err := pbih.k8sClient.Get(context.Background(), k8sapitypes.NamespacedName{
			Name:      listenerName,
			Namespace: item.PoolNamespace,
		}, rawListener); err != nil {
			if k8serrors.IsNotFound(err) {
				blog.Warnf("listener %s/%s not found, no need to clean", listenerName, item.PoolNamespace)
				continue
			}
			blog.Warnf("get listener %s/%s failed, err %s", listenerName, item.PoolNamespace, err.Error())
			return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusDeleting)
		}
		// do not update informer cache directly
		cpListener := rawListener.DeepCopy()
		if cpListener.Spec.TargetGroup == nil || len(cpListener.Spec.TargetGroup.Backends) == 0 {
			if cpListener.Status.Status == networkextensionv1.ListenerStatusSynced {
				blog.Infof("listener %s/%s backend cleaned and synced", listenerName, item.PoolNamespace)
				continue
			}
			blog.Warnf("listener %s/%s backend cleaned, but not synced", listenerName, item.PoolNamespace)
			return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusDeleting)
		}
		cpListener.Spec.TargetGroup = nil
		cpListener.Status.Status = networkextensionv1.ListenerStatusNotSynced
		if cpListener.Labels != nil {
			delete(cpListener.Labels, networkextensionv1.LabelKeyForSourceNamespace)
			delete(cpListener.Labels, networkextensionv1.LabelKeyForSourceName)
		}
		if err := pbih.k8sClient.Update(context.Background(), cpListener, &client.UpdateOptions{}); err != nil {
			blog.Warnf("failed to update listener %s/%s, err %s", listenerName, item.PoolNamespace, err.Error())
			return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusDeleting)
		}
		blog.Infof("update listener %s/%s successfully", listenerName, item.PoolNamespace)
	}
	return pbih.generateStatus(item, itemStatus, constant.PortBindingItemStatusCleaned)
}
