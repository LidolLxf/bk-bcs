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

package business

import (
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// 磁盘相关接口

// ListAvailableDiskTypes 列出可用的磁盘类型
func ListAvailableDiskTypes(availableZone map[string]struct{}, data []*cbs.DiskConfig) []*proto.DiskConfigSet {
	dataMap := make(map[string]map[string]*cbs.DiskConfig)
	systemMap := make(map[string]map[string]*cbs.DiskConfig)
	zoneMap := make(map[string]struct{})
	for _, disk := range data {
		if disk.Available == nil || !*disk.Available {
			continue
		}

		if _, ok := availableZone[*disk.Zone]; !ok {
			continue
		}

		zoneMap[*disk.Zone] = struct{}{}
		if *disk.DiskUsage == "SYSTEM_DISK" {
			if v, ok := systemMap[*disk.Zone]; !ok {
				systemMap[*disk.Zone] = map[string]*cbs.DiskConfig{
					*disk.DiskType: disk,
				}
			} else {
				v[*disk.DiskType] = disk
			}
		} else {
			if v, ok := dataMap[*disk.Zone]; !ok {
				dataMap[*disk.Zone] = map[string]*cbs.DiskConfig{
					*disk.DiskType: disk,
				}
			} else {
				v[*disk.DiskType] = disk
			}
		}
	}

	dataDiskTypeMap := make(map[string]int)
	systemDiskTypeMap := make(map[string]int)
	for zone := range zoneMap {
		for diskType := range dataMap[zone] {
			dataDiskTypeMap[diskType]++
		}
		for diskType := range systemMap[zone] {
			systemDiskTypeMap[diskType]++
		}
	}

	count := len(zoneMap)
	result := make([]*proto.DiskConfigSet, 0)

	result = append(result, listDiskType(dataDiskTypeMap, dataMap, count)...)
	result = append(result, listDiskType(systemDiskTypeMap, systemMap, count)...)

	return result
}

func listDiskType(diskTypeMap map[string]int, diskMap map[string]map[string]*cbs.DiskConfig,
	count int) []*proto.DiskConfigSet {
	result := make([]*proto.DiskConfigSet, 0)

	for k, v := range diskTypeMap {
		if count == v {
			for _, m := range diskMap {
				if disk, ok := m[k]; ok {
					var stepSize int32
					if disk.StepSize != nil {
						stepSize = int32(*disk.StepSize)
					}

					result = append(result, &proto.DiskConfigSet{
						DiskType:     *disk.DiskType,
						DiskTypeName: common.DiskType[k],
						DiskUsage:    *disk.DiskUsage,
						MinDiskSize:  int32(*disk.MinDiskSize),
						MaxDiskSize:  int32(*disk.MaxDiskSize),
						StepSize:     stepSize,
					})

					break
				}
			}
		}
	}

	return result
}
