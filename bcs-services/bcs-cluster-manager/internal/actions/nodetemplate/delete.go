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

package nodetemplate

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// DeleteAction action for delete nodeTemplate
type DeleteAction struct {
	ctx context.Context

	model    store.ClusterManagerModel
	template *cmproto.NodeTemplate
	req      *cmproto.DeleteNodeTemplateRequest
	resp     *cmproto.DeleteNodeTemplateResponse
	project  *bcsproject.Project
}

// NewDeleteAction create delete action for nodeTemplate
func NewDeleteAction(model store.ClusterManagerModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (da *DeleteAction) validate() error {
	err := da.req.Validate()
	if err != nil {
		return err
	}
	template, err := da.model.GetNodeTemplate(da.ctx, da.req.ProjectID, da.req.NodeTemplateID)
	if err != nil {
		return err
	}

	da.template = template

	proInfo, errLocal := project.GetProjectManagerClient().GetProjectInfo(da.ctx, da.req.ProjectID, true)
	if errLocal == nil {
		da.project = proInfo
	}

	return nil
}

// Handle handle delete node template
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteNodeTemplateRequest, resp *cmproto.DeleteNodeTemplateResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete nodeTemplate failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := da.model.DeleteNodeTemplate(da.ctx, da.req.ProjectID, da.req.NodeTemplateID); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.Project.String(),
		ResourceID:   req.ProjectID,
		TaskID:       "",
		Message:      fmt.Sprintf("项目[%s]删除节点模板信息[%s]", da.project.GetName(), req.NodeTemplateID),
		OpUser:       auth.GetUserFromCtx(ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
		ProjectID:    req.ProjectID,
		ResourceName: da.project.GetName(),
	})
	if err != nil {
		blog.Errorf("DeleteNodeTemplate[%s] CreateOperationLog failed: %v", req.ProjectID, err)
	}

	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
