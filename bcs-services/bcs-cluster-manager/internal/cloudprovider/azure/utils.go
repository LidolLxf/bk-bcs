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

package azure

import (
	"fmt"
	"strconv"
	"strings"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

var (
	cloudName = "azure"
)

// task template name
const (
	// createClusterTaskTemplate bk-sops add task template
	createClusterTaskTemplate = "aks-create cluster: %s"
	// deleteClusterTaskTemplate bk-sops add task template
	deleteClusterTaskTemplate = "aks-delete cluster: %s"
	// createNodeGroupTaskTemplate bk-sops add task template
	createNodeGroupTaskTemplate = "aks-create node group: %s/%s"
	// switchNodeGroupAutoScalingTaskTemplate bk-sops add task template
	switchNodeGroupAutoScalingTaskTemplate = "aks-switch node group auto scaling: %s/%s"
	// deleteNodeGroupTaskTemplate bk-sops add task template
	deleteNodeGroupTaskTemplate = "aks-delete node group: %s/%s"
	// updateNodeGroupTaskTemplate bk-sops add task template
	updateNodeGroupTaskTemplate = "aks-update node group: %s/%s"
	// updateNodeGroupDesiredNode bk-sops add task template
	updateNodeGroupDesiredNodeTemplate = "aks-update node group desired node: %s/%s"
	// updateAutoScalingOptionTemplate bk-sops add task template
	updateAutoScalingOptionTemplate = "aks-update auto scaling option: %s"
	// cleanNodeGroupNodesTaskTemplate bk-sops add task template
	cleanNodeGroupNodesTaskTemplate = "aks-remove node group nodes: %s/%s"
	// switchAutoScalingOptionStatusTemplate bk-sops add task template
	switchAutoScalingOptionStatusTemplate = "aks-switch auto scaling option status: %s"
)

// tasks
var (
	// create cluster task
	createAKSClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateCloudClusterTask", cloudName),
		StepName:   "创建集群",
	}
	checkAKSClusterStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCloudClusterStatusTask", cloudName),
		StepName:   "检测集群状态",
	}
	checkAKSNodeGroupsStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-AKS-CheckCloudNodeGroupStatusTask", cloudName),
		StepName:   "检测集群节点池状态",
	}
	updateAKSNodeGroupsToDBStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateCleanNodeGroupNodesDBInfoTask", cloudName),
		StepName:   "更新节点池信息",
	}
	checkCreateClusterNodeStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCreateClusterNodeStatusTask", cloudName),
		StepName:   "检测集群节点状态",
	}
	registerAKSClusterKubeConfigStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-AKS-RegisterClusterKubeConfigTask", cloudName),
		StepName:   "注册集群连接信息",
	}
	updateAKSNodesToDBStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateAddNodeDBInfoTask", cloudName),
		StepName:   "更新任务状态",
	}

	// import cluster task
	importClusterNodesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-ImportClusterNodesTask", cloudName),
		StepName:   "导入集群节点",
	}
	registerClusterKubeConfigStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-RegisterClusterKubeConfigTask", cloudName),
		StepName:   "注册集群kubeConfig认证",
	}

	// delete cluster task
	deleteAKSClusterStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-DeleteCloudClusterTask", cloudName),
		StepName:   "删除集群",
	}
	cleanClusterDBInfoStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CleanClusterDBInfoTask", cloudName),
		StepName:   "清理集群数据",
	}

	// create nodeGroup task
	createCloudNodeGroupStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CreateCloudNodeGroupTask", cloudName),
		StepName:   "创建云节点组",
	}
	checkCloudNodeGroupStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckCloudNodeGroupStatusTask", cloudName),
		StepName:   "检测云节点组状态",
	}

	// clean node in nodeGroup task
	cleanNodeGroupNodesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CleanNodeGroupNodesTask", cloudName),
		StepName:   "下架节点组节点",
	}

	// delete nodeGroup task
	deleteNodeGroupStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-DeleteNodeGroupTask", cloudName),
		StepName:   "删除云节点组",
	}

	// update desired nodes task
	applyInstanceMachinesStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-%s", cloudName, cloudprovider.ApplyInstanceMachinesTask),
		StepName:   "申请节点任务",
	}
	checkClusterNodesStatusStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-CheckClusterNodesStatusTask", cloudName),
		StepName:   "检测节点状态",
	}

	// update nodegroup task
	updateAKSNodeGroupStep = cloudprovider.StepInfo{
		StepMethod: fmt.Sprintf("%s-UpdateCloudNodeGroupTask", cloudName),
		StepName:   "更新节点池",
	}
)

// CreateClusterTaskOption 创建集群构建step子任务
type CreateClusterTaskOption struct {
	Cluster      *proto.Cluster
	NodeGroupIDs []string
}

// BuildCreateClusterStep 创建集群任务
func (cn *CreateClusterTaskOption) BuildCreateClusterStep(task *proto.Task) {
	createStep := cloudprovider.InitTaskStep(createAKSClusterStep)
	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	createStep.Params[cloudprovider.NodeGroupIDKey.String()] = strings.Join(cn.NodeGroupIDs, ",")

	task.Steps[createAKSClusterStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, createAKSClusterStep.StepMethod)
}

// BuildCheckClusterStatusStep 检测集群状态任务
func (cn *CreateClusterTaskOption) BuildCheckClusterStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkAKSClusterStatusStep)
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[checkAKSClusterStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkAKSClusterStatusStep.StepMethod)
}

// BuildCheckNodeGroupsStatusStep 检测集群节点池状态任务
func (cn *CreateClusterTaskOption) BuildCheckNodeGroupsStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkAKSNodeGroupsStatusStep)
	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	checkStep.Params[cloudprovider.NodeGroupIDKey.String()] = strings.Join(cn.NodeGroupIDs, ",")

	task.Steps[checkAKSNodeGroupsStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkAKSNodeGroupsStatusStep.StepMethod)
}

// BuildUpdateNodeGroupsToDBStep 更新集群节点池信息任务
func (cn *CreateClusterTaskOption) BuildUpdateNodeGroupsToDBStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateAKSNodeGroupsToDBStep)
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[updateAKSNodeGroupsToDBStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateAKSNodeGroupsToDBStep.StepMethod)
}

// BuildCheckClusterNodesStatusStep 检测创建集群节点状态任务
func (cn *CreateClusterTaskOption) BuildCheckClusterNodesStatusStep(task *proto.Task) {
	createStep := cloudprovider.InitTaskStep(checkCreateClusterNodeStatusStep)
	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[checkCreateClusterNodeStatusStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, checkCreateClusterNodeStatusStep.StepMethod)
}

// BuildUpdateNodesToDBStep 更新集群节点信息任务
func (cn *CreateClusterTaskOption) BuildUpdateNodesToDBStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(updateAKSNodesToDBStep)
	updateStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[updateAKSNodesToDBStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, updateAKSNodesToDBStep.StepMethod)
}

// BuildRegisterClsKubeConfigStep 托管集群注册连接信息
func (cn *CreateClusterTaskOption) BuildRegisterClsKubeConfigStep(task *proto.Task) {
	registerStep := cloudprovider.InitTaskStep(registerAKSClusterKubeConfigStep)
	registerStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	registerStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider
	registerStep.Params[cloudprovider.IsExtranetKey.String()] = icommon.True

	task.Steps[registerAKSClusterKubeConfigStep.StepMethod] = registerStep
	task.StepSequence = append(task.StepSequence, registerAKSClusterKubeConfigStep.StepMethod)
}

// BuildImportClusterNodesStep 纳管集群节点
func (cn *CreateClusterTaskOption) BuildImportClusterNodesStep(task *proto.Task) {
	importNodesStep := cloudprovider.InitTaskStep(importClusterNodesStep)
	importNodesStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Cluster.ClusterID
	importNodesStep.Params[cloudprovider.CloudIDKey.String()] = cn.Cluster.Provider

	task.Steps[importClusterNodesStep.StepMethod] = importNodesStep
	task.StepSequence = append(task.StepSequence, importClusterNodesStep.StepMethod)
}

// ImportClusterTaskOption 纳管集群
type ImportClusterTaskOption struct {
	Cluster *proto.Cluster
}

// BuildRegisterKubeConfigStep 注册集群kubeConfig
func (ic *ImportClusterTaskOption) BuildRegisterKubeConfigStep(task *proto.Task) {
	registerKubeConfigStep := cloudprovider.InitTaskStep(registerClusterKubeConfigStep)
	registerKubeConfigStep.Params[cloudprovider.ClusterIDKey.String()] = ic.Cluster.ClusterID
	registerKubeConfigStep.Params[cloudprovider.CloudIDKey.String()] = ic.Cluster.Provider

	task.Steps[registerClusterKubeConfigStep.StepMethod] = registerKubeConfigStep
	task.StepSequence = append(task.StepSequence, registerClusterKubeConfigStep.StepMethod)
}

// BuildImportClusterNodesStep 纳管集群节点
func (ic *ImportClusterTaskOption) BuildImportClusterNodesStep(task *proto.Task) {
	importNodesStep := cloudprovider.InitTaskStep(importClusterNodesStep)
	importNodesStep.Params[cloudprovider.ClusterIDKey.String()] = ic.Cluster.ClusterID
	importNodesStep.Params[cloudprovider.CloudIDKey.String()] = ic.Cluster.Provider

	task.Steps[importClusterNodesStep.StepMethod] = importNodesStep
	task.StepSequence = append(task.StepSequence, importClusterNodesStep.StepMethod)
}

// DeleteClusterTaskOption 删除集群
type DeleteClusterTaskOption struct {
	Cluster    *proto.Cluster
	DeleteMode string
}

// BuildDeleteAKSClusterStep 删除集群
func (dc *DeleteClusterTaskOption) BuildDeleteAKSClusterStep(task *proto.Task) {
	deleteStep := cloudprovider.InitTaskStep(deleteAKSClusterStep)

	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = dc.Cluster.ClusterID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = dc.Cluster.Provider
	deleteStep.Params[cloudprovider.DeleteModeKey.String()] = dc.DeleteMode

	task.Steps[deleteAKSClusterStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteAKSClusterStep.StepMethod)
}

// BuildCleanClusterDBInfoStep 清理集群数据
func (dc *DeleteClusterTaskOption) BuildCleanClusterDBInfoStep(task *proto.Task) {
	updateStep := cloudprovider.InitTaskStep(cleanClusterDBInfoStep)

	updateStep.Params[cloudprovider.ClusterIDKey.String()] = dc.Cluster.ClusterID
	updateStep.Params[cloudprovider.CloudIDKey.String()] = dc.Cluster.Provider

	task.Steps[cleanClusterDBInfoStep.StepMethod] = updateStep
	task.StepSequence = append(task.StepSequence, cleanClusterDBInfoStep.StepMethod)
}

// CreateNodeGroupTaskOption 创建节点组
type CreateNodeGroupTaskOption struct {
	Group *proto.NodeGroup
}

// BuildCreateCloudNodeGroupStep 通过云接口创建节点组
func (cn *CreateNodeGroupTaskOption) BuildCreateCloudNodeGroupStep(task *proto.Task) {
	createStep := cloudprovider.InitTaskStep(createCloudNodeGroupStep)

	createStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	createStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	createStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	task.Steps[createCloudNodeGroupStep.StepMethod] = createStep
	task.StepSequence = append(task.StepSequence, createCloudNodeGroupStep.StepMethod)
}

// BuildCheckCloudNodeGroupStatusStep 检测节点组状态
func (cn *CreateNodeGroupTaskOption) BuildCheckCloudNodeGroupStatusStep(task *proto.Task) {
	checkStep := cloudprovider.InitTaskStep(checkCloudNodeGroupStatusStep)

	checkStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	checkStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	checkStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider

	task.Steps[checkCloudNodeGroupStatusStep.StepMethod] = checkStep
	task.StepSequence = append(task.StepSequence, checkCloudNodeGroupStatusStep.StepMethod)
}

// CleanNodeInGroupTaskOption 节点组缩容节点
type CleanNodeInGroupTaskOption struct {
	Group    *proto.NodeGroup
	NodeIPs  []string
	NodeIds  []string
	Operator string
}

// BuildCleanNodeGroupNodesStep 清理节点池节点
func (cn *CleanNodeInGroupTaskOption) BuildCleanNodeGroupNodesStep(task *proto.Task) {
	cleanStep := cloudprovider.InitTaskStep(cleanNodeGroupNodesStep)

	cleanStep.Params[cloudprovider.ClusterIDKey.String()] = cn.Group.ClusterID
	cleanStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.Group.NodeGroupID
	cleanStep.Params[cloudprovider.CloudIDKey.String()] = cn.Group.Provider
	cleanStep.Params[cloudprovider.NodeIPsKey.String()] = strings.Join(cn.NodeIPs, ",")
	cleanStep.Params[cloudprovider.NodeIDsKey.String()] = strings.Join(cn.NodeIds, ",")

	task.Steps[cleanNodeGroupNodesStep.StepMethod] = cleanStep
	task.StepSequence = append(task.StepSequence, cleanNodeGroupNodesStep.StepMethod)
}

// DeleteNodeGroupTaskOption 删除节点组
type DeleteNodeGroupTaskOption struct {
	Group *proto.NodeGroup
}

// BuildDeleteNodeGroupStep 删除云节点组
func (dn *DeleteNodeGroupTaskOption) BuildDeleteNodeGroupStep(task *proto.Task) {
	deleteStep := cloudprovider.InitTaskStep(deleteNodeGroupStep)

	deleteStep.Params[cloudprovider.ClusterIDKey.String()] = dn.Group.ClusterID
	deleteStep.Params[cloudprovider.NodeGroupIDKey.String()] = dn.Group.NodeGroupID
	deleteStep.Params[cloudprovider.CloudIDKey.String()] = dn.Group.Provider

	task.Steps[deleteNodeGroupStep.StepMethod] = deleteStep
	task.StepSequence = append(task.StepSequence, deleteNodeGroupStep.StepMethod)
}

// UpdateDesiredNodesTaskOption 扩容节点组节点
type UpdateDesiredNodesTaskOption struct {
	Group    *proto.NodeGroup
	Desired  uint32
	Operator string
}

// BuildApplyInstanceMachinesStep 申请节点实例
func (ud *UpdateDesiredNodesTaskOption) BuildApplyInstanceMachinesStep(task *proto.Task) {
	applyInstanceStep := cloudprovider.InitTaskStep(applyInstanceMachinesStep)

	applyInstanceStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Group.ClusterID
	applyInstanceStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.Group.NodeGroupID
	applyInstanceStep.Params[cloudprovider.CloudIDKey.String()] = ud.Group.Provider
	applyInstanceStep.Params[cloudprovider.ScalingNodesNumKey.String()] = strconv.Itoa(int(ud.Desired))
	applyInstanceStep.Params[cloudprovider.OperatorKey.String()] = ud.Operator

	task.Steps[applyInstanceMachinesStep.StepMethod] = applyInstanceStep
	task.StepSequence = append(task.StepSequence, applyInstanceMachinesStep.StepMethod)
}

// BuildCheckClusterNodeStatusStep 检测节点实例状态
func (ud *UpdateDesiredNodesTaskOption) BuildCheckClusterNodeStatusStep(task *proto.Task) {
	checkClusterNodeStatusStep := cloudprovider.InitTaskStep(checkClusterNodesStatusStep)

	checkClusterNodeStatusStep.Params[cloudprovider.ClusterIDKey.String()] = ud.Group.ClusterID
	checkClusterNodeStatusStep.Params[cloudprovider.NodeGroupIDKey.String()] = ud.Group.NodeGroupID
	checkClusterNodeStatusStep.Params[cloudprovider.CloudIDKey.String()] = ud.Group.Provider

	task.Steps[checkClusterNodesStatusStep.StepMethod] = checkClusterNodeStatusStep
	task.StepSequence = append(task.StepSequence, checkClusterNodesStatusStep.StepMethod)
}

// UpdateNodeGroupTaskOption 创建集群构建step子任务
type UpdateNodeGroupTaskOption struct {
	NodeGroup *proto.NodeGroup
}

// BuildUpdateNodeGroupStep 更新节点池
func (cn *UpdateNodeGroupTaskOption) BuildUpdateNodeGroupStep(task *proto.Task) {
	updateNodeGroupStep := cloudprovider.InitTaskStep(updateAKSNodeGroupStep)
	updateNodeGroupStep.Params[cloudprovider.ClusterIDKey.String()] = cn.NodeGroup.ClusterID
	updateNodeGroupStep.Params[cloudprovider.NodeGroupIDKey.String()] = cn.NodeGroup.NodeGroupID
	updateNodeGroupStep.Params[cloudprovider.CloudIDKey.String()] = cn.NodeGroup.Provider

	task.Steps[updateAKSNodeGroupStep.StepMethod] = updateNodeGroupStep
	task.StepSequence = append(task.StepSequence, updateAKSNodeGroupStep.StepMethod)
}
