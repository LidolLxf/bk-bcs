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

package release

import (
	"context"
	"fmt"
	"sort"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	ptypes "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// GetMatchedRelease get the app instance's matched release id.
func (rs *ReleasedService) GetMatchedRelease(kt *kit.Kit, meta *types.AppInstanceMeta) (uint32, error) {

	ctx, cancel := context.WithTimeout(context.TODO(), rs.matchReleaseWaitTime)
	defer cancel()

	if err := rs.limiter.Wait(ctx); err != nil {
		return 0, err
	}

	am, err := rs.cache.App.GetMeta(kt, meta.BizID, meta.AppID)
	if err != nil {
		return 0, err
	}

	switch am.ConfigType {
	case table.File:
	case table.KV:
	default:
		return 0, errf.New(errf.InvalidParameter, "only supports File and KV configuration types.")
	}

	groups, err := rs.listReleasedGroups(kt, meta)
	if err != nil {
		return 0, err
	}

	matched, err := rs.matchReleasedGroupWithLabels(kt, groups, meta)
	if err != nil {
		return 0, err
	}

	return matched.ReleaseID, nil
}

// listReleasedGroups list released groups
func (rs *ReleasedService) listReleasedGroups(kt *kit.Kit, meta *types.AppInstanceMeta) (
	[]*ptypes.ReleasedGroupCache, error) {
	list, err := rs.cache.ReleasedGroup.Get(kt, meta.BizID, meta.AppID)
	if err != nil {
		return nil, fmt.Errorf("get current published strategy failed, err: %v", err)
	}

	return list, nil
}

type matchedMeta struct {
	StrategyID uint32
	ReleaseID  uint32
	GroupID    uint32
}

// matchOneStrategyWithLabels match at most only one strategy with app instance labels.
func (rs *ReleasedService) matchReleasedGroupWithLabels(
	_ *kit.Kit,
	groups []*ptypes.ReleasedGroupCache,
	meta *types.AppInstanceMeta) (*matchedMeta, error) {
	// 1. sort released groups by update time
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].UpdatedAt.After(groups[j].UpdatedAt)
	})
	// 2. match groups with labels
	matchedList := []*matchedMeta{}
	var def *matchedMeta
	for _, group := range groups {
		switch group.Mode {
		case table.GroupModeDebug:
			if group.UID == meta.Uid {
				matchedList = append(matchedList, &matchedMeta{
					ReleaseID:  group.ReleaseID,
					GroupID:    group.GroupID,
					StrategyID: group.StrategyID,
				})
			}
		case table.GroupModeCustom:
			if group.Selector == nil {
				return nil, errf.New(errf.InvalidParameter, "custom group must have selector")
			}
			matched, err := group.Selector.MatchLabels(meta.Labels)
			if err != nil {
				return nil, err
			}
			if matched {
				matchedList = append(matchedList, &matchedMeta{
					ReleaseID:  group.ReleaseID,
					GroupID:    group.GroupID,
					StrategyID: group.StrategyID,
				})
			}
		case table.GroupModeDefault:
			def = &matchedMeta{
				ReleaseID:  group.ReleaseID,
				GroupID:    group.GroupID,
				StrategyID: group.StrategyID,
			}
		}
	}

	if len(matchedList) == 0 {
		if def == nil {
			return nil, errf.ErrAppInstanceNotMatchedRelease
		}
		return def, nil
	}

	// released groups were sorted by strategy id, so the first one is the latest one.

	return matchedList[0], nil
}
