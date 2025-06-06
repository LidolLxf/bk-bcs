/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/aws/aws-sdk-go/service/autoscaling"
)

// AutoScalingClient aws auto scaling client
type AutoScalingClient struct {
	asClient *autoscaling.AutoScaling
}

// NewAutoScalingClient init autoscaling client
func NewAutoScalingClient(opt *cloudprovider.CommonOption) (*AutoScalingClient, error) {
	sess, err := NewSession(opt)
	if err != nil {
		return nil, err
	}

	return &AutoScalingClient{
		asClient: autoscaling.New(sess),
	}, nil
}

// DescribeAutoScalingGroups describes AutoScalingGroups
func (as *AutoScalingClient) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (
	[]*autoscaling.Group, error) {
	blog.Infof("DescribeAutoScalingGroups input: %s", utils.ToJSONString(input))
	output, err := as.asClient.DescribeAutoScalingGroups(input)
	if err != nil {
		blog.Errorf("DescribeAutoScalingGroups failed: %v", err)
		return nil, err
	}
	if output == nil || output.AutoScalingGroups == nil {
		blog.Errorf("DescribeAutoScalingGroups lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}

	return output.AutoScalingGroups, nil
}

// UpdateAutoScalingGroup update AutoScalingGroup
func (as *AutoScalingClient) UpdateAutoScalingGroup(input *autoscaling.UpdateAutoScalingGroupInput) error {
	blog.Infof("UpdateAutoScalingGroup input: %s", utils.ToJSONString(input))
	_, err := as.asClient.UpdateAutoScalingGroup(input)
	if err != nil {
		blog.Errorf("UpdateAutoScalingGroup failed: %v", err)
		return err
	}

	return nil
}

// SetDesiredCapacity describes AutoScalingGroups
func (as *AutoScalingClient) SetDesiredCapacity(asgName string, capacity int64) error {
	blog.Infof("SetDesiredCapacity set autoScalingGroup[%s] capacity to %d", asgName, capacity)
	_, err := as.asClient.SetDesiredCapacity(
		&autoscaling.SetDesiredCapacityInput{
			AutoScalingGroupName: &asgName,
			DesiredCapacity:      &capacity,
		},
	)
	if err != nil {
		blog.Errorf("SetDesiredCapacity failed: %v", err)
		return err
	}
	blog.Infof("SetDesiredCapacity for %s successful, capacity %d", asgName, capacity)

	return nil
}

// TerminateInstanceInAutoScalingGroup terminates instance in AutoScalingGroups
func (as *AutoScalingClient) TerminateInstanceInAutoScalingGroup(
	input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.Activity, error) {
	blog.Infof("TerminateInstanceInAutoScalingGroup input: %s", utils.ToJSONString(input))
	output, err := as.asClient.TerminateInstanceInAutoScalingGroup(input)
	if err != nil {
		blog.Errorf("TerminateInstanceInAutoScalingGroup failed: %v", err)
		return nil, err
	}
	blog.Infof("TerminateInstanceInAutoScalingGroup instance %s successful", input.InstanceId)

	return output.Activity, nil
}

// DetachInstances detach instances in AutoScalingGroups
func (as *AutoScalingClient) DetachInstances(input *autoscaling.DetachInstancesInput) ([]*autoscaling.Activity, error) {
	blog.Infof("DetachInstances input: %s", utils.ToJSONString(input))
	output, err := as.asClient.DetachInstances(input)
	if err != nil {
		blog.Errorf("DetachInstances failed: %v", err)
		return nil, err
	}
	blog.Infof("DetachInstances instances %v for group %s successful", input.InstanceIds, input.AutoScalingGroupName)

	return output.Activities, nil
}

// DescribeLifecycleHooks list all lifecycle hooks in AutoScalingGroups
func (as *AutoScalingClient) DescribeLifecycleHooks(asName *string) ([]*autoscaling.LifecycleHook, error) {
	blog.Infof("DescribeLifecycleHooks asName: %s", asName)

	output, err := as.asClient.DescribeLifecycleHooks(&autoscaling.DescribeLifecycleHooksInput{
		AutoScalingGroupName: asName,
	})
	if err != nil {
		blog.Errorf("DescribeLifecycleHooks failed: %v", err)
		return nil, err
	}
	blog.Infof("DescribeLifecycleHooks find hooks %s", output.GoString())

	return output.LifecycleHooks, nil
}

// DeleteLifecycleHooks delete lifecycle hooks in AutoScalingGroups
func (as *AutoScalingClient) DeleteLifecycleHooks(asName *string, hookName []string) error {
	for _, hook := range hookName {
		_, err := as.asClient.DeleteLifecycleHook(&autoscaling.DeleteLifecycleHookInput{
			AutoScalingGroupName: asName,
			LifecycleHookName:    &hook,
		})
		if err != nil {
			blog.Errorf("DeleteLifecycleHook failed: %s", err)
			return err
		}

		blog.Infof("DeleteLifecycleHooks delete hooks %s successful", hook)
	}

	return nil
}
