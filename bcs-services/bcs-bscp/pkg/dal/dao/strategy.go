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

package dao

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Strategy supplies all the Strategy related operations.
type Strategy interface {
	// Get last strategy.
	GetLast(kit *kit.Kit, bizID, appID, releasedID uint32) (*table.Strategy, error)
}

var _ Strategy = new(strategyDao)

type strategyDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Get strategy kv.
func (dao *strategyDao) GetLast(kit *kit.Kit, bizID, appID, releasedID uint32) (*table.Strategy, error) {
	m := dao.genQ.Strategy
	return m.WithContext(kit.Ctx).Where(
		m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ReleaseID.Eq(releasedID)).Last()
}
