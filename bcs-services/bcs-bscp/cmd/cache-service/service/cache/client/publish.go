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

package client

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/keys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/go-redis/redis/v8"
)

// GetPublishTime get publish time
func (c *client) GetPublishTime(kt *kit.Kit, bizID uint32, publishTime int64) ([]redis.Z, error) {
	return c.bds.ZRangeWithScores(kt.Ctx, keys.Key.PublishTime(bizID), publishTime, publishTime+1)
}

// SetPublishTime set publish time
func (c *client) SetPublishTime(kt *kit.Kit, bizID uint32, publishTime int64, strategyId uint32) (int64, error) {

	scoresResults, err := c.bds.ZRangeWithScores(kt.Ctx, keys.Key.PublishTime(bizID), publishTime, publishTime+1)
	if err != nil {
		return 0, err
	}

	// 同一时间上线应该不超99999个
	account := 0.00001
	max := float64(publishTime)
	for _, v := range scoresResults {
		if v.Score > max {
			max = v.Score
		}
	}
	newScore := max + account
	return c.bds.ZAdd(kt.Ctx, keys.Key.PublishTime(bizID), newScore, strategyId)
}
