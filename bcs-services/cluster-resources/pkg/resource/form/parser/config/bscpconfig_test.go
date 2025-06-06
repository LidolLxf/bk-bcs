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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightBscpConfigManifest = map[string]interface{}{
	"spec": map[string]interface{}{
		"configSyncer": []interface{}{
			map[string]interface{}{
				"data":         []interface{}{},
				"matchConfigs": []interface{}{"123", "123123"},
			}, map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"key":       "key",
						"refConfig": "refConfig1",
					},
					map[string]interface{}{
						"key":       "key2",
						"refConfig": "refConfig2",
					},
				},
				"matchConfigs": nil,
			}, map[string]interface{}{
				"data":         nil,
				"matchConfigs": []interface{}{"1231223", "3123"},
			}, map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"key":       "key1",
						"refConfig": "refConfig1",
					},
					map[string]interface{}{
						"key":       "key2",
						"refConfig": "refConfig2",
					},
				},
				"matchConfigs": nil,
			}, map[string]interface{}{
				"data": []interface{}{
					"key", "key2",
				},
				"matchConfigs": nil,
			},
		},
	},
}

var exceptedBscpConfigData = model.BscpConfigSpec{
	ConfigSyncer: []model.ConfigSyncer{
		{
			ConfigData:       nil,
			AssociationRules: "matchConfigs",
			SecretType:       "Opaque",
			ResourceType:     "secret",
			MatchConfigs: []model.MatchConfigs{
				{
					Value: "123",
				}, {
					Value: "123123",
				},
			},
		}, {
			AssociationRules: "data",
			SecretType:       "Opaque",
			ResourceType:     "secret",
			ConfigData: []model.ConfigSyncerData{
				{
					Key:       "key",
					RefConfig: "refConfig1",
				}, {
					Key:       "key2",
					RefConfig: "refConfig2",
				},
			},
			MatchConfigs: nil,
		}, {
			AssociationRules: "matchConfigs",
			SecretType:       "Opaque",
			ResourceType:     "secret",
			ConfigData:       nil,
			MatchConfigs: []model.MatchConfigs{
				{
					Value: "1231223",
				}, {
					Value: "3123",
				},
			},
		}, {
			AssociationRules: "data",
			SecretType:       "Opaque",
			ResourceType:     "secret",
			ConfigData: []model.ConfigSyncerData{
				{
					Key:       "key1",
					RefConfig: "refConfig1",
				}, {
					Key:       "key2",
					RefConfig: "refConfig2",
				},
			},
			MatchConfigs: nil,
		}, {
			AssociationRules: "data",
			SecretType:       "Opaque",
			ResourceType:     "secret",
			ConfigData:       nil,
			MatchConfigs:     nil,
		},
	},
}

func TestParseBscpConfigSpec(t *testing.T) {
	actualBscpConfigData := model.BscpConfigSpec{}
	ParseBscpConfigSpec(lightBscpConfigManifest, &actualBscpConfigData)
	assert.Equal(t, exceptedBscpConfigData, actualBscpConfigData)
}
