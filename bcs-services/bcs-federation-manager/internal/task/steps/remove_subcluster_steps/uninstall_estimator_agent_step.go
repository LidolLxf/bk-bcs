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

// Package steps include all steps for federation manager
package steps

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

var (
	// UninstallEstimatorAgentStepName step name for create cluster
	UninstallEstimatorAgentStepName = fedsteps.StepNames{
		Alias: "uninstall estimator agent",
		Name:  "UNINSTALL_ESTIMATOR_AGENT",
	}
)

// NewUninstallEstimatorAgentStep sum step
func NewUninstallEstimatorAgentStep() *UninstallEstimatorAgentStep {
	return &UninstallEstimatorAgentStep{}
}

// UninstallEstimatorAgentStep sum step
type UninstallEstimatorAgentStep struct{}

// Alias step name
func (s UninstallEstimatorAgentStep) Alias() string {
	return UninstallEstimatorAgentStepName.Alias
}

// GetName step name
func (s UninstallEstimatorAgentStep) GetName() string {
	return UninstallEstimatorAgentStepName.Name
}

// DoWork for worker exec task
func (s UninstallEstimatorAgentStep) DoWork(t *types.Task) error {
	step, exist := t.GetStep(s.GetName())
	if !exist {
		return fmt.Errorf("task[%s] not exist step[%s]", t.TaskID, s.GetName())
	}

	// get common params
	// host cluster may not in same project with federation cluster
	hostProjectId, ok := t.GetCommonParams(fedsteps.HostProjectIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostProjectIdKey)
	}

	hostClusterId, ok := t.GetCommonParams(fedsteps.HostClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.HostClusterIdKey)
	}

	subClusterId, ok := t.GetCommonParams(fedsteps.SubClusterIdKey)
	if !ok {
		return fedsteps.ParamsNotFoundError(t.TaskID, fedsteps.SubClusterIdKey)
	}

	err := helm.GetHelmClient().UninstallEstimatorAgent(&helm.BcsEstimatorAgentOptions{
		ReleaseBaseOptions: helm.ReleaseBaseOptions{
			ProjectID: hostProjectId,
			ClusterID: hostClusterId,
		},
		SubClusterId: subClusterId,
	})
	if err != nil {
		return err
	}

	blog.Infof("taskId: %s, taskType: %s, taskName: %s result: %v\n", t.GetTaskID(), t.GetTaskType(), step.GetName(), fedsteps.Success)
	return nil
}

// BuildStep build step
func (s UninstallEstimatorAgentStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	// stepName/s.GetName() 用于标识这个step
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
