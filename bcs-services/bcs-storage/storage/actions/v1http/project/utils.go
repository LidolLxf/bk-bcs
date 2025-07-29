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

package project

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go-micro.dev/v4/broker"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

const (
	projectIDTag    = "projectID"
	projectCodeTag  = "projectCode"
	projectNameTag  = "name"
	resourceTypeTag = "resourceType"

	tablePro      = "Project"
	dataTag       = "data"
	updateTimeTag = "updateTime"
	createTimeTag = "createTime"
)

var proFeatTags = []string{projectIDTag, projectNameTag, projectCodeTag, resourceTypeTag}

// Use Mongodb for storage.
const dbConfig = "mongodb/dynamic"

func isExistResourceQueue(features map[string]string) bool {
	if len(features) == 0 {
		return false
	}

	resourceType, ok := features[resourceTypeTag]
	if !ok {
		return false
	}

	if _, ok := apiserver.GetAPIResource().GetMsgQueue().ResourceToQueue[resourceType]; !ok {
		return false
	}

	return true
}

func publishProjectInfoToQueue(data operator.M, featTags []string, event msgqueue.EventKind) error {
	var (
		err     error
		message = &broker.Message{
			Header: map[string]string{},
		}
	)

	startTime := time.Now()
	for _, feat := range featTags {
		if v, ok := data[feat].(string); ok {
			message.Header[feat] = v
		}
	}
	message.Header[string(msgqueue.EventType)] = string(event)
	message.Header[resourceTypeTag] = tablePro

	exist := isExistResourceQueue(message.Header)
	if !exist {
		return nil
	}

	// NOCC:revive/early-return(设计如此:)
	// nolint
	if v, ok := data[dataTag]; ok {
		codec.EncJson(v, &message.Body)
	} else {
		blog.Infof("object[%v] not exist data", data[dataTag])
		return nil
	}

	err = apiserver.GetAPIResource().GetMsgQueue().MsgQueue.Publish(message)
	if err != nil {
		return err
	}

	if queueName, ok := message.Header[resourceTypeTag]; ok {
		metrics.ReportQueuePushMetrics(queueName, err, startTime)
	}

	return nil
}
