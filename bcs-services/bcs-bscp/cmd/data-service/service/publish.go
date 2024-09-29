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

package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/itsm"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	pbgroup "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/group"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Publish exec publish strategy.
// nolint: funlen
func (s *Service) Publish(ctx context.Context, req *pbds.PublishReq) (*pbds.PublishResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	groupIDs := make([]uint32, 0)
	tx := s.dao.GenQuery().Begin()

	release, err := s.dao.Release().Get(grpcKit, req.BizId, req.AppId, req.ReleaseId)
	if err != nil {
		return nil, err
	}
	if release.Spec.Deprecated {
		return nil, fmt.Errorf("release %s is deprecated, can not be published", release.Spec.Name)
	}

	if !req.All {
		if req.GrayPublishMode == "" {
			// !NOTE: Compatible with previous pipelined plugins version
			req.GrayPublishMode = table.PublishByGroups.String()
		}
		publishMode := table.GrayPublishMode(req.GrayPublishMode)
		if e := publishMode.Validate(); e != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, e
		}
		// validate and query group ids.
		if publishMode == table.PublishByGroups {
			for _, groupID := range req.Groups {
				if groupID == 0 {
					groupIDs = append(groupIDs, groupID)
					continue
				}
				group, e := s.dao.Group().Get(grpcKit, groupID, req.BizId)
				if e != nil {
					if rErr := tx.Rollback(); rErr != nil {
						logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
					}
					return nil, fmt.Errorf("group %d not exist", groupID)
				}
				groupIDs = append(groupIDs, group.ID)
			}
		}
		if publishMode == table.PublishByLabels {
			groupID, gErr := s.getOrCreateGroupByLabels(grpcKit, tx, req.BizId, req.AppId, req.GroupName, req.Labels)
			if gErr != nil {
				logs.Errorf("create group by labels failed, err: %v, rid: %s", gErr, grpcKit.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
				}
				return nil, gErr
			}
			groupIDs = append(groupIDs, groupID)
		}
	}

	opt := &types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: req.ReleaseId,
		All:       req.All,
		Default:   req.Default,
		Memo:      req.Memo,
		Groups:    groupIDs,
		Revision: &table.CreatedRevision{
			Creator: grpcKit.User,
		},
	}

	pshID, err := s.dao.Publish().PublishWithTx(grpcKit, tx, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	app, err := s.dao.App().GetByID(grpcKit, req.AppId)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	var havePull bool
	if !app.Spec.LastConsumedTime.IsZero() {
		havePull = true
	}

	haveCredentials, err := s.checkAppHaveCredentials(grpcKit, req.BizId, req.AppId)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.PublishResp{
		PublishedStrategyHistoryId: pshID,
		HaveCredentials:            haveCredentials,
		HavePull:                   havePull,
	}
	return resp, nil
}

// SubmitPublishApprove submit publish strategy.
// nolint: funlen
func (s *Service) SubmitPublishApprove(
	ctx context.Context, req *pbds.SubmitPublishApproveReq) (*pbds.PublishResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	groupIDs := make([]uint32, 0)
	tx := s.dao.GenQuery().Begin()

	app, err := s.dao.App().Get(grpcKit, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}

	release, err := s.dao.Release().Get(grpcKit, req.BizId, req.AppId, req.ReleaseId)
	if err != nil {
		return nil, err
	}
	if release.Spec.Deprecated {
		return nil, fmt.Errorf("release %s is deprecated, can not be submited", release.Spec.Name)
	}

	// group name
	groupName := []string{}
	if !req.All {
		if req.GrayPublishMode == "" {
			// !NOTE: Compatible with previous pipelined plugins version
			req.GrayPublishMode = table.PublishByGroups.String()
		}
		publishMode := table.GrayPublishMode(req.GrayPublishMode)
		if e := publishMode.Validate(); e != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, e
		}
		// validate and query group ids.
		if publishMode == table.PublishByGroups {
			for _, groupID := range req.Groups {
				if groupID == 0 {
					groupIDs = append(groupIDs, groupID)
					continue
				}
				group, e := s.dao.Group().Get(grpcKit, groupID, req.BizId)
				if e != nil {
					if rErr := tx.Rollback(); rErr != nil {
						logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
					}
					return nil, fmt.Errorf("group %d not exist", groupID)
				}
				groupIDs = append(groupIDs, group.ID)
				groupName = append(groupName, group.Spec.Name)
			}
		}
		if publishMode == table.PublishByLabels {
			groupID, gErr := s.getOrCreateGroupByLabels(grpcKit, tx, req.BizId, req.AppId, req.GroupName, req.Labels)
			if gErr != nil {
				logs.Errorf("create group by labels failed, err: %v, rid: %s", gErr, grpcKit.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
				}
				return nil, gErr
			}
			groupIDs = append(groupIDs, groupID)
			groupName = append(groupName, req.GroupName)
		}
	}

	opt := s.parsePublishOption(req, app)
	opt.Groups = groupIDs
	opt.Revision = &table.CreatedRevision{
		Creator: grpcKit.User,
	}

	pshID, err := s.dao.Publish().SubmitWithTx(grpcKit, tx, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	haveCredentials, err := s.checkAppHaveCredentials(grpcKit, req.BizId, req.AppId)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	if req.All {
		groupName = []string{"all"}
	}

	resInstance := fmt.Sprintf("releases_name: %s\ngroup: %s", release.Spec.Name, strings.Join(groupName, ","))

	// audit this to create strategy details
	ad := s.dao.AuditDao().DecoratorV3(grpcKit, opt.BizID, &table.AuditField{
		OperateWay:       grpcKit.OperateWay,
		Action:           enumor.PublishVersionConfig,
		ResourceInstance: resInstance,
		Status:           enumor.AuditStatus(opt.PublishStatus),
		AppId:            app.AppID(),
		StrategyId:       pshID,
		IsCompare:        req.IsCompare,
	}).PrepareCreateByInstance(pshID, req)
	if err = ad.Do(tx.Query); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	// 定时上线
	err = s.setPublishTime(grpcKit, pshID, req)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	// itsm流程创建ticket
	if app.Spec.IsApprove {
		scope := strings.Join(groupName, ",")
		ticketData, errCreate := s.submitCreateApproveTicket(grpcKit, app, release.Spec.Name, scope, grpcKit.User)
		if errCreate != nil {
			logs.Errorf("submit create approve ticket, err: %v, rid: %s", errCreate, grpcKit.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
		}

		err = s.dao.Strategy().UpdateByID(grpcKit, tx, pshID, map[string]interface{}{
			"itsm_ticket_type":     constant.ItsmTicketTypeCreate,
			"itsm_ticket_url":      ticketData.TicketURL,
			"itsm_ticket_sn":       ticketData.SN,
			"itsm_ticket_status":   constant.ItsmTicketStatusCreated,
			"itsm_ticket_state_id": ticketData.StateID,
		})

		if err != nil {
			logs.Errorf("update strategy by id err: %v, rid: %s", err, grpcKit.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.PublishResp{
		PublishedStrategyHistoryId: pshID,
		HaveCredentials:            haveCredentials,
	}
	return resp, nil
}

// Approve publish approve.
func (s *Service) Approve(ctx context.Context, req *pbds.ApproveReq) (*pbds.ApproveResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	tx := s.dao.GenQuery().Begin()
	defer func() {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
	}()

	release, err := s.dao.Release().Get(grpcKit, req.BizId, req.AppId, req.ReleaseId)
	if err != nil {
		return nil, err
	}
	if release.Spec.Deprecated {
		return nil, fmt.Errorf("release %s is deprecated, can not be revoke", release.Spec.Name)
	}

	strategy, err := s.dao.Strategy().GetLast(grpcKit, req.BizId, req.AppId, req.ReleaseId)
	if err != nil {
		return nil, err
	}

	// // 获取ticket相关内容，如果状态不正常则更新表
	// tikectInfo, err := s.checkAndUpdateTicketInfo(tx, strategy.Spec.ItsmTicketSn, req.PublishStatus)
	// if err != nil {
	// 	return nil, err
	// }

	var updateContent map[string]interface{}
	itsmUpdata := make(map[string]interface{})
	switch req.PublishStatus {
	case string(table.RevokedPublish):
		updateContent, err = s.revokeApprove(grpcKit, req, strategy)
		if err != nil {
			return nil, err
		}
		itsmUpdata = map[string]interface{}{
			"sn":             strategy.Spec.ItsmTicketSn,
			"operator":       grpcKit.User,
			"action_type":    "WITHDRAW",
			"action_message": fmt.Sprintf("BSCP 代理用户 %s 撤回: %s", grpcKit.User, req.Reason),
		}
	case string(table.RejectedApproval):
		updateContent, err = s.rejectApprove(grpcKit, req, strategy)
		if err != nil {
			return nil, err
		}
		itsmUpdata = map[string]interface{}{
			"sn":       strategy.Spec.ItsmTicketSn,
			"state id": strategy.Spec.ItsmTicketStateID,
			"approver": grpcKit.User,
			"action":   "false",
			"remark":   req.Reason,
		}
	case string(table.PendPublish):
		updateContent, err = s.passApprove(grpcKit, tx, req, strategy)
		if err != nil {
			return nil, err
		}
		itsmUpdata = map[string]interface{}{
			"sn":       strategy.Spec.ItsmTicketSn,
			"state id": strategy.Spec.ItsmTicketStateID,
			"approver": grpcKit.User,
			"action":   "true",
		}
	case string(table.AlreadyPublish):
		updateContent, err = s.publishApprove(grpcKit, tx, req, strategy)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid publish_status: %s", req.PublishStatus)
	}

	updateContent["reviser"] = grpcKit.User
	err = s.dao.Strategy().UpdateByID(grpcKit, tx, strategy.ID, updateContent)
	if err != nil {
		return nil, err
	}

	// update audit details
	err = s.dao.AuditDao().UpdateByStrategyID(grpcKit, tx, strategy.ID, map[string]interface{}{
		"status": updateContent["publish_status"],
	})
	if err != nil {
		return nil, err
	}

	if req.PublishStatus == string(table.RevokedPublish) {
		err = itsm.WithdrawTicket(grpcKit.Ctx, itsmUpdata)
		if err != nil {
			return nil, err
		}
	}

	// if req.PublishStatus == string(table.RejectedApproval) || req.PublishStatus == string(table.PendPublish) {
	// 	_, err = itsm.UpdateTicketByApporver(grpcKit.Ctx, itsmUpdata)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	haveCredentials, err := s.checkAppHaveCredentials(grpcKit, req.BizId, req.AppId)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	return &pbds.ApproveResp{
		HaveCredentials: haveCredentials,
	}, nil
}

// 获取ticket相关内容，如果状态不正常则更新表,后面再处理
// func (s *Service) checkAndUpdateTicketInfo(tx *gen.QueryTx, sn, publishStatus string) (types.TicketInfo, error) {
// 	var ticketInfo types.TicketInfo
// 	// 查询单据是否正常
// 	ticketsItem, err := itsm.ListTickets([]string{sn})
// 	if err != nil {
// 		return ticketInfo, err
// 	}

// 	updateContent := make(map[string]interface{})
// 	var errReturn error
// 	for _, ticket := range ticketsItem {
// 		updateContent["reviser"] = ticket.UpdatedBy
// 		// 单据已经被撤销
// 		if ticket.CurrentStatus == constant.TicketRevokedStatu {
// 			updateContent["publish_status"] = table.RevokedPublish
// 			errReturn = errors.New("this approve has been revoked by itsm")
// 		}

// 		// 单据已经被驳回或者关闭
// 		if ticket.CurrentStatus == constant.TicketTerminatedStatu {
// 			updateContent["publish_status"] = table.RejectedApproval
// 			errReturn = errors.New("this approve has been revoked by itsm")
// 		}

// 		// 挂起不是正常单据直接报错
// 		if ticket.CurrentStatus == constant.TicketSuspendedStatu {
// 			return ticketInfo, errors.New("this approve has been suspended by itsm")
// 		}
// 		// if ticket.CurrentStatus == "RUNNING" ||
// 		// 	ticket.CurrentStatus == "TERMINATED" ||
// 		// 	ticket.CurrentStatus == "REVOKED" {
// 		// 	return errors.New("this approve is finished, ")
// 		// }
// 		// // 其他情况
// 		ticketInfo.Status = ticket.CurrentStatus
// 		ticketInfo.Operater = ticket.UpdateAt
// 	}

// 	err = s.dao.Strategy().UpdateByID(grpcKit, tx, strategy.ID, updateContent)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// update audit details
// 	err = s.dao.AuditDao().UpdateByStrategyID(grpcKit, tx, strategy.ID, map[string]interface{}{
// 		"status": updateContent["publish_status"],
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// // 非running状态直接return报错
// 	// if tikectInfo.Status != constant.TicketRunningStatu {
// 	// 	return nil, fmt.Errorf("invalid approve data, ticket statu: %s", tikectInfo.Status)
// 	// }

// 	// 已处理【负责人审批】(拒绝)
// 	// 已处理【负责人审批】(通过)
// 	// 关闭了单据：测试.
// 	return ticketInfo, nil
// }

// GenerateReleaseAndPublish generate release and publish.
// nolint: funlen
func (s *Service) GenerateReleaseAndPublish(ctx context.Context, req *pbds.GenerateReleaseAndPublishReq) (
	*pbds.PublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	app, err := s.dao.App().GetByID(grpcKit, req.AppId)
	if err != nil {
		logs.Errorf("get app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if _, e := s.dao.Release().GetByName(grpcKit, req.BizId, req.AppId, req.ReleaseName); e == nil {
		return nil, fmt.Errorf("release name %s already exists", req.ReleaseName)
	}

	tx := s.dao.GenQuery().Begin()

	groupIDs, err := s.genReleaseAndPublishGroupID(grpcKit, tx, req)
	if err != nil {
		return nil, err
	}

	// create release.
	release := &table.Release{
		Spec: &table.ReleaseSpec{
			Name: req.ReleaseName,
			Memo: req.ReleaseMemo,
		},
		Attachment: &table.ReleaseAttachment{
			BizID: req.BizId,
			AppID: req.AppId,
		},
		Revision: &table.CreatedRevision{
			Creator: grpcKit.User,
		},
	}
	releaseID, err := s.dao.Release().CreateWithTx(grpcKit, tx, release)
	if err != nil {
		logs.Errorf("create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	// create released hook.
	if err = s.createReleasedHook(grpcKit, tx, req.BizId, req.AppId, releaseID); err != nil {
		logs.Errorf("create released hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	switch app.Spec.ConfigType {
	case table.File:

		// Note: need to change batch operator to query config item and it's commit.
		// query app's all config items.
		cfgItems, e := s.getAppConfigItems(grpcKit)
		if e != nil {
			logs.Errorf("query app config item list failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}

		// get app template revisions which are template config items
		tmplRevisions, e := s.getAppTmplRevisions(grpcKit)
		if e != nil {
			logs.Errorf("get app template revisions failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}

		// if no config item, return directly.
		if len(cfgItems) == 0 && len(tmplRevisions) == 0 {
			return nil, errors.New("app config items is empty")
		}

		// do template and non-template config item related operations for create release.
		if err = s.doConfigItemOperations(grpcKit, req.Variables, tx, release.ID, tmplRevisions, cfgItems); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			logs.Errorf("do template action for create release failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
	case table.KV:
		if err = s.doKvOperations(grpcKit, tx, req.AppId, req.BizId, release.ID); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			logs.Errorf("do kv action for create release failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
	}

	// publish with transaction.
	kt := kit.FromGrpcContext(ctx)

	opt := &types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: releaseID,
		All:       req.All,
		Memo:      req.ReleaseMemo,
		Groups:    groupIDs,
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
	}
	pshID, err := s.dao.Publish().PublishWithTx(kt, tx, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	// commit transaction.
	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &pbds.PublishResp{PublishedStrategyHistoryId: pshID}, nil
}

// revokeApprove revoke publish approve.
func (s *Service) revokeApprove(
	kit *kit.Kit, req *pbds.ApproveReq, strategy *table.Strategy) (map[string]interface{}, error) {

	app, err := s.dao.App().Get(kit, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}

	// 有服务权限的人才能撤销
	if app.Revision.Creator != kit.User {
		return nil, errors.New("no permission to revoke")
	}

	// 只有待上线以及待审批的类型才允许撤回
	if strategy.Spec.PublishStatus != table.PendPublish && strategy.Spec.PublishStatus != table.PendApproval {
		return nil, fmt.Errorf("revoked not allowed, current publish status is: %s", strategy.Spec.PublishStatus)
	}

	return map[string]interface{}{
		"publish_status": table.RevokedPublish,
		"reject_reason":  req.Reason,
	}, nil
}

// rejectApprove reject publish approve.
func (s *Service) rejectApprove(
	kit *kit.Kit, req *pbds.ApproveReq, strategy *table.Strategy) (map[string]interface{}, error) {

	if strategy.Spec.PublishStatus != table.PendApproval {
		return nil, fmt.Errorf("rejected not allowed, current publish status is: %s", strategy.Spec.PublishStatus)
	}

	if req.Reason == "" {
		return nil, errors.New("reason can not empty")
	}

	// 判断是否在审批人队列
	isApprover := false
	users := strings.Split(strategy.Spec.ApproverProgress, ",")
	for _, v := range users {
		if v == kit.User {
			isApprover = true
		}

	}

	// 需要审批但不是审批人的情况返回无权限审批
	if !isApprover {
		return nil, errors.New("no permission to approve")
	}

	return map[string]interface{}{
		"publish_status": table.RejectedApproval,
		"reject_reason":  req.Reason,
	}, nil
}

// passApprove pass publish approve.
func (s *Service) passApprove(
	kit *kit.Kit, tx *gen.QueryTx, req *pbds.ApproveReq, strategy *table.Strategy) (map[string]interface{}, error) {

	if strategy.Spec.PublishStatus != table.PendApproval {
		return nil, fmt.Errorf("pass not allowed, current publish status is: %s", strategy.Spec.PublishStatus)
	}

	// 存在app更改成不审批的情况，要根据审批人来确定是会签还是或签
	// 判断是否在审批人队列
	isApprover := false
	approverProgress := strategy.Spec.ApproverProgress
	progressUsers := strings.Split(approverProgress, ",")
	var newProgressUsers []string
	for _, v := range progressUsers {
		if v == kit.User {
			isApprover = true
			continue
		}
		newProgressUsers = append(newProgressUsers, v)
	}

	// 不是审批人的情况返回无权限审批
	if !isApprover {
		return nil, errors.New("no permission to approve")
	}

	publishStatus := table.PendApproval
	// 或签通过
	if strings.Contains(strategy.Spec.Approver, "|") || strategy.Spec.Approver == kit.User {
		publishStatus = table.PendPublish
	} else {
		// 最后一个的情况下，直接待上线
		if len(newProgressUsers) == 0 {
			publishStatus = table.PendPublish
		} else {
			approverProgress = strings.Join(newProgressUsers, ",")
		}
	}

	// 自动上线则直接上线
	if publishStatus == table.PendPublish && strategy.Spec.PublishType == table.Automatically {
		opt := types.PublishOption{
			BizID:     req.BizId,
			AppID:     req.AppId,
			ReleaseID: req.ReleaseId,
			All:       false,
		}

		if len(strategy.Spec.Scope.Groups) == 0 {
			opt.All = true
		}

		err := s.dao.Publish().UpsertPublishWithTx(kit, tx, &opt, strategy)

		if err != nil {
			return nil, err
		}
		publishStatus = table.AlreadyPublish
	}

	return map[string]interface{}{
		"publish_status":    publishStatus,
		"approver_progress": approverProgress,
	}, nil
}

// 根据审批类型获取itsms的多种id
func (s *Service) getItsmInfoId(kt *kit.Kit, signType table.ApproveType) (int, int, int, error) {
	itsmConf := cc.DataService().ITSM
	// 或签和会签是不同的模板
	var getConfigKey string
	var itsmConfServiceID int
	var serviceID int
	var stateId int
	var workflowId int
	switch signType {
	case table.OrSign:
		getConfigKey = constant.CreateOrSignApproveItsmServiceID
		itsmConfServiceID = itsmConf.CreateOrSignServiceID
	case table.CountSign:
		getConfigKey = constant.CreateCountSignApproveItsmServiceID
		itsmConfServiceID = itsmConf.CreateCountSignServiceID
	}
	if itsmConf.AutoRegister {
		itsmConfig, err := s.dao.ItsmConfig().GetConfig(kt, getConfigKey)
		if err != nil {
			return 0, 0, 0, err
		}
		serviceID = itsmConfig.Value
		stateId = itsmConfig.StateApproveId
	} else {
		serviceID = itsmConfServiceID
		var err error
		// 获取流程信息及workflow id
		workflowId, err = itsm.GetWorkflowByService(serviceID)
		if err != nil {
			return 0, 0, 0, err
		}

		stateApproveId, err := itsm.GetStateApproveByWorkfolw(workflowId)
		if err != nil {
			return 0, 0, 0, err
		}
		stateId = stateApproveId
	}
	return serviceID, workflowId, stateId, nil
}

// publishApprove publish approve.
func (s *Service) publishApprove(
	kit *kit.Kit, tx *gen.QueryTx, req *pbds.ApproveReq, strategy *table.Strategy) (map[string]interface{}, error) {

	if strategy.Spec.PublishStatus != table.PendPublish {
		return nil, fmt.Errorf("publish not allowed, current publish status is: %s", strategy.Spec.PublishStatus)
	}

	opt := types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: req.ReleaseId,
		All:       false,
	}

	if len(strategy.Spec.Scope.Groups) == 0 {
		opt.All = true
	}

	err := s.dao.Publish().UpsertPublishWithTx(kit, tx, &opt, strategy)

	if err != nil {
		return nil, err
	}
	publishStatus := table.AlreadyPublish

	return map[string]interface{}{
		"pub_state":      table.Publishing,
		"publish_status": publishStatus,
	}, nil
}

// parse publish option
func (s *Service) parsePublishOption(req *pbds.SubmitPublishApproveReq, app *table.App) *types.PublishOption {

	opt := &types.PublishOption{
		BizID:         req.BizId,
		AppID:         req.AppId,
		ReleaseID:     req.ReleaseId,
		All:           req.All,
		Default:       req.Default,
		Memo:          req.Memo,
		PublishType:   table.PublishType(req.PublishType),
		PublishTime:   req.PublishTime,
		PublishStatus: table.PendPublish,
		PubState:      string(table.Publishing),
	}

	// if approval required, current approver required, pub_state unpublished
	if app.Spec.IsApprove {
		opt.PublishStatus = table.PendApproval
		opt.Approver = app.Spec.Approver
		opt.ApproverProgress = app.Spec.Approver
		opt.PubState = string(table.Unpublished)
	}

	// 后续app改审批方式的时候可以判断是或签还是会签
	if app.Spec.ApproveType == table.OrSign {
		opt.Approver = app.Spec.Approver
		approver := strings.Split(app.Spec.Approver, ",")
		opt.Approver = strings.Join(approver, "|")
	}

	// publish immediately
	if req.PublishType == string(table.Immediately) {
		opt.PublishStatus = table.AlreadyPublish
	}

	return opt
}

// checkAppHaveCredentials check if there is available credential for app.
// 1. credential scope can match app name.
// 2. credential is enabled.
func (s *Service) checkAppHaveCredentials(grpcKit *kit.Kit, bizID, appID uint32) (bool, error) {
	app, err := s.dao.App().Get(grpcKit, bizID, appID)
	if err != nil {
		return false, err
	}
	matchedCredentials := make([]uint32, 0)
	scopes, err := s.dao.CredentialScope().ListAll(grpcKit, bizID)
	if err != nil {
		return false, err
	}
	if len(scopes) == 0 {
		return false, nil
	}
	for _, scope := range scopes {
		match, e := scope.Spec.CredentialScope.MatchApp(app.Spec.Name)
		if e != nil {
			return false, e
		}
		if match {
			matchedCredentials = append(matchedCredentials, scope.Attachment.CredentialId)
		}
	}
	credentials, e := s.dao.Credential().BatchListByIDs(grpcKit, bizID, matchedCredentials)
	if e != nil {
		return false, e
	}
	for _, credential := range credentials {
		if credential.Spec.Enable {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) genReleaseAndPublishGroupID(grpcKit *kit.Kit, tx *gen.QueryTx,
	req *pbds.GenerateReleaseAndPublishReq) ([]uint32, error) {

	groupIDs := make([]uint32, 0)

	if !req.All {
		if req.GrayPublishMode == "" {
			// !NOTE: Compatible with previous pipelined plugins version
			req.GrayPublishMode = table.PublishByGroups.String()
		}
		publishMode := table.GrayPublishMode(req.GrayPublishMode)
		if e := publishMode.Validate(); e != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, e
		}
		// validate and query group ids.
		if publishMode == table.PublishByGroups {
			for _, name := range req.Groups {
				group, e := s.dao.Group().GetByName(grpcKit, req.BizId, name)
				if e != nil {
					if rErr := tx.Rollback(); rErr != nil {
						logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
					}
					return nil, fmt.Errorf("group %s not exist", name)
				}
				groupIDs = append(groupIDs, group.ID)
			}
		}
		if publishMode == table.PublishByLabels {
			groupID, e := s.getOrCreateGroupByLabels(grpcKit, tx, req.BizId, req.AppId, req.GroupName, req.Labels)
			if e != nil {
				logs.Errorf("create group by labels failed, err: %v, rid: %s", e, grpcKit.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
				}
				return nil, e
			}
			groupIDs = append(groupIDs, groupID)
		}
	}

	return groupIDs, nil
}

func (s *Service) getOrCreateGroupByLabels(grpcKit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32, groupName string,
	labels []*structpb.Struct) (uint32, error) {
	elements := make([]selector.Element, 0)
	for _, label := range labels {
		element, err := pbgroup.UnmarshalElement(label)
		if err != nil {
			return 0, fmt.Errorf("unmarshal group label failed, err: %v", err)
		}
		elements = append(elements, *element)
	}
	sel := &selector.Selector{
		LabelsAnd: elements,
	}
	groups, err := s.dao.Group().ListAppValidGroups(grpcKit, bizID, appID)
	if err != nil {
		return 0, err
	}
	exists := make([]*table.Group, 0)
	for _, group := range groups {
		if group.Spec.Selector.Equal(sel) {
			exists = append(exists, group)
		}
	}
	// if same labels group exists, return it's id.
	if len(exists) > 0 {
		return exists[0].ID, nil
	}
	// else create new one.
	if groupName != "" {
		// if group name is not empty, use it as group name.
		_, err = s.dao.Group().GetByName(grpcKit, bizID, groupName)
		// if group name already exists, return error.
		if err == nil {
			return 0, fmt.Errorf("group %s already exists", groupName)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	} else {
		// generate group name by time.
		groupName = time.Now().Format("20060102150405.000")
		groupName = fmt.Sprintf("g_%s", strings.ReplaceAll(groupName, ".", ""))
	}
	group := table.Group{
		Spec: &table.GroupSpec{
			Name:     groupName,
			Public:   false,
			Mode:     table.GroupModeCustom,
			Selector: sel,
		},
		Attachment: &table.GroupAttachment{
			BizID: bizID,
		},
		Revision: &table.Revision{
			Creator: grpcKit.User,
			Reviser: grpcKit.User,
		},
	}
	groupID, err := s.dao.Group().CreateWithTx(grpcKit, tx, &group)
	if err != nil {
		return 0, err
	}
	if err := s.dao.GroupAppBind().BatchCreateWithTx(grpcKit, tx, []*table.GroupAppBind{
		{
			GroupID: groupID,
			AppID:   appID,
			BizID:   bizID,
		},
	}); err != nil {
		return 0, err
	}
	return groupID, nil
}

func (s *Service) createReleasedHook(grpcKit *kit.Kit, tx *gen.QueryTx, bizID, appID, releaseID uint32) error {
	pre, err := s.dao.ReleasedHook().Get(grpcKit, bizID, appID, 0, table.PreHook)
	if err == nil {
		pre.ID = 0
		pre.ReleaseID = releaseID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, pre); e != nil {
			logs.Errorf("create released pre-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			return e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released pre-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}
	post, err := s.dao.ReleasedHook().Get(grpcKit, bizID, appID, 0, table.PostHook)
	if err == nil {
		post.ID = 0
		post.ReleaseID = releaseID
		if _, e := s.dao.ReleasedHook().CreateWithTx(grpcKit, tx, post); e != nil {
			logs.Errorf("create released post-hook failed, err: %v, rid: %s", e, grpcKit.Rid)
			return e
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("query released post-hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return err
	}
	return nil
}

// submitCreateApproveTicket create new itsm create approve ticket
func (s *Service) submitCreateApproveTicket(
	kt *kit.Kit, app *table.App, releaseName, scope, username string) (*itsm.CreateTicketData, error) {
	var serviceID int
	itsmConf := cc.DataService().ITSM

	// 或签和会签是不同的模板
	var getConfigKey string
	var itsmConfServiceID int
	var stateId int
	switch app.Spec.ApproveType {
	case table.OrSign:
		getConfigKey = constant.CreateOrSignApproveItsmServiceID
		itsmConfServiceID = itsmConf.CreateOrSignServiceID
	case table.CountSign:
		getConfigKey = constant.CreateCountSignApproveItsmServiceID
		itsmConfServiceID = itsmConf.CreateCountSignServiceID
	}
	if itsmConf.AutoRegister {
		itsmConfig, err := s.dao.ItsmConfig().GetConfig(kt, getConfigKey)
		if err != nil {
			return nil, err
		}
		serviceID = itsmConfig.Value
		stateId = itsmConfig.StateApproveId
	} else {
		serviceID = itsmConfServiceID
		// 获取流程信息及workflow id
		workflowId, err := itsm.GetWorkflowByService(serviceID)
		if err != nil {
			return nil, err
		}

		stateApproveId, err := itsm.GetStateApproveByWorkfolw(workflowId)
		if err != nil {
			return nil, err
		}
		stateId = stateApproveId

	}

	// 获取所有的业务信息
	bizList, err := s.esb.Cmdb().ListAllBusiness(kt.Ctx)
	if err != nil {
		return nil, err
	}

	if len(bizList.Info) == 0 {
		return nil, fmt.Errorf("biz list is empty")
	}

	var bizName string
	for _, biz := range bizList.Info {
		if biz.BizID == int64(app.BizID) {
			bizName = biz.BizName
			break
		}
	}

	fields := []map[string]interface{}{
		{
			"key":   "title",
			"value": "服务配置中心(BSCP)版本上线审批",
		}, {
			"key":   "BIZ",
			"value": fmt.Sprintf(bizName+"(%d)", app.BizID),
		}, {
			"key":   "APP",
			"value": app.Spec.Name,
		}, {
			"key":   "VERSION_NAME",
			"value": releaseName,
		}, {
			"key":   "SCOPE",
			"value": scope,
		}, {
			"key":   "COMPARE",
			"value": "test",
		},
	}

	stateApproveId := strconv.Itoa(stateId)
	reqData := map[string]interface{}{
		"creator":    username,
		"service_id": serviceID,
		"fields":     fields,
		"meta": map[string]interface{}{
			"state_processors": map[string]interface{}{
				stateApproveId: app.Spec.Approver,
			}},
	}
	return itsm.CreateTicket(kt.Ctx, reqData)
}

// 定时上线
func (s *Service) setPublishTime(kt *kit.Kit, pshID uint32, req *pbds.SubmitPublishApproveReq) error {
	if req.PublishType == string(table.Periodically) {
		// 通过当前时区计算unix
		location := time.Now().Location()
		publishTime, err := time.ParseInLocation(time.DateTime, req.PublishTime, location)
		if err != nil {
			logs.Errorf("parse time failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		_, err = s.cs.SetPublishTime(kt.Ctx, &pbcs.SetPublishTimeReq{
			BizId:       req.BizId,
			StrategyId:  pshID,
			PublishTime: publishTime.UTC().Unix(),
			AppId:       req.AppId,
		})
		if err != nil {
			logs.Errorf("set publish time failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}
