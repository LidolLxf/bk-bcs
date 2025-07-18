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

// Package utils provides utility functions for the mesh manager store package
package utils

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
)

const (
	// DataTableNamePrefix is the prefix for data table names
	DataTableNamePrefix = "bcs_mesh_"
)

// EnsureTable ensures the data table exists, creates it if it doesn't exist
func EnsureTable(ctx context.Context, db drivers.DB, tableName string, indexes []drivers.Index) error {
	hasTable, err := db.HasTable(ctx, tableName)
	if err != nil {
		return err
	}
	if !hasTable {
		tErr := db.CreateTable(ctx, tableName)
		if tErr != nil {
			return tErr
		}
	}
	// only ensure index when index name is not empty
	for _, idx := range indexes {
		hasIndex, iErr := db.Table(tableName).HasIndex(ctx, idx.Name)
		if iErr != nil {
			return iErr
		}
		if !hasIndex {
			if iErr = db.Table(tableName).CreateIndex(ctx, idx); iErr != nil {
				return iErr
			}
		}
	}
	return nil
}

// ListOption defines options for listing data
// The listed data index ranges from page*size to page*size+(size-1)
type ListOption struct {
	// Sort map for sorting list results
	Sort map[string]int

	// Page number for the current list
	Page int64

	// Size of each page in the current list
	Size int64
}
