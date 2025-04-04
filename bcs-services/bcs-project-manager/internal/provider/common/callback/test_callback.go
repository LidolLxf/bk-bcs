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

// Package callback xxx
package callback

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

const (
	// CallBackExampleName example
	CallBackExampleName = "example"
)

// NewTestCallback 创建回调实例
func NewTestCallback() task.CallbackInterface {
	return &callBack{}
}

type callBack struct{}

// GetName 回调方法名称
func (c *callBack) GetName() string {
	return CallBackExampleName
}

// Callback 回调方法,根据任务成功状态更新实体对象状态
func (c *callBack) Callback(isSuccess bool, task *types.Task) {
	if isSuccess {
		blog.Infof("task[%s] execute %s", task.GetTaskID(), "success")
		return
	}

	blog.Infof("task[%s] execute %s", task.GetTaskID(), "failure")
}
