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

// Package itsm 在 ITSM 注册服务，包括：创建命名空间、更新命名空间、删除命名空间, 允许重复执行
package itsm

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/itsm"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

var (
	daoSet dao.Set

	// WorkflowTemplates itsm templates
	//go:embed templates
	WorkflowTemplates embed.FS
)

// InitServices 初始化BSCP相关流程服务
func InitServices() error {

	// initial DAO set
	set, err := dao.NewDaoSet(cc.DataService().Sharding, cc.DataService().Credential)
	if err != nil {
		return fmt.Errorf("initial dao set failed, err: %v", err)
	}

	daoSet = set

	if err := InitApproveITSMServices(); err != nil {
		fmt.Printf("init approve itsm services failed, err: %s\n", err.Error())
		return err
	}

	return nil
}

// InitApproveITSMServices 初始化上线审批相关流程服务
func InitApproveITSMServices() error {
	// 2. create itsm catalog
	catalogID, err := createITSMCatalog()
	if err != nil {
		return err
	}
	// 3. import approve services
	if err := importCountSignApproveService(catalogID); err != nil {
		return err
	}
	if err := importOrSignApproveService(catalogID); err != nil {
		return err
	}
	return nil
}

func createITSMCatalog() (uint32, error) {
	catalogs, err := itsm.ListCatalogs()
	if err != nil {
		return 0, err
	}

	var rootID uint32
	var parentID uint32
	for _, rootCatalog := range catalogs {
		if rootCatalog.Key == "root" {
			rootID = rootCatalog.ID
			for _, parentCatalog := range rootCatalog.Children {
				if parentCatalog.Name == "服务配置中心" {
					parentID = parentCatalog.ID
					for _, catalog := range parentCatalog.Children {
						if catalog.Name == "上线审批" {
							return catalog.ID, nil
						}
					}
				}
			}
		}
	}
	if rootID == 0 {
		return 0, fmt.Errorf("root catalog not found")
	}
	if parentID == 0 {
		parentID, err = itsm.CreateCatalog(itsm.CreateCatalogReq{
			ProjectKey: "0",
			ParentID:   rootID,
			Name:       "服务配置中心",
			Desc:       "服务配置中心相关流程",
		})
		if err != nil {
			return 0, err
		}
	}
	// create namespace catalog
	catalogID, err := itsm.CreateCatalog(itsm.CreateCatalogReq{
		ProjectKey: "0",
		ParentID:   parentID,
		Name:       "上线审批",
		Desc:       "服务配置上线操作",
	})
	if err != nil {
		return 0, err
	}
	return catalogID, nil
}

func importCountSignApproveService(catalogID uint32) error {
	// check whether the service has been imported before
	// if not, import it, else update it.

	kt := kit.New()
	var imported bool
	itsmConfig, err := daoSet.ItsmConfig().GetConfig(kt, constant.CreateCountSignApproveItsmServiceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		imported = true
	} else {
		imported = false
	}

	// 自定义模板分隔符为 [[ ]]，例如 [[ .Name ]]，避免和 ITSM 模板变量格式冲突
	tmpl, err := template.New("create_shared_count_sign_approve.json.tpl").Delims("[[", "]]").
		ParseFS(WorkflowTemplates, "templates/create_shared_count_sign_approve.json.tpl")
	if err != nil {
		return err
	}
	stringBuffer := &strings.Builder{}
	if err = tmpl.Execute(stringBuffer, map[string]string{
		"BCSGateway": "",
		"BCSToken":   "",
		"Approvers":  "",
	}); err != nil {
		return err
	}
	mp := map[string]interface{}{}
	if err = json.Unmarshal([]byte(stringBuffer.String()), &mp); err != nil {
		return err
	}
	importReq := itsm.ImportServiceReq{
		Key:             "request",
		Name:            "创建上线会签审批",
		Desc:            "创建上线会签审批",
		CatelogID:       catalogID,
		Owners:          "admin",
		CanTicketAgency: false,
		IsValid:         true,
		DisplayType:     "OPEN",
		DisplayRole:     "",
		Source:          "custom",
		ProjectKey:      "0",
		Workflow:        mp,
	}
	var serviceID int
	if !imported {
		serviceID, err = itsm.ImportService(importReq)
		if err != nil {
			return err
		}
		workflowId, err := itsm.GetWorkflowByService(serviceID)
		if err != nil {
			return err
		}

		stateApproveId, err := itsm.GetStateApproveByWorkfolw(workflowId)
		if err != nil {
			return err
		}
		return daoSet.ItsmConfig().SetConfig(kt, &table.ItsmConfig{
			Key:            constant.CreateCountSignApproveItsmServiceID,
			Value:          serviceID,
			WorkflowId:     workflowId,
			StateApproveId: stateApproveId,
		})
	}

	return itsm.UpdateService(itsm.UpdateServiceReq{
		ID:               itsmConfig.Value,
		ImportServiceReq: importReq,
	})
}

func importOrSignApproveService(catalogID uint32) error {
	// check whether the service has been imported before
	// if not, import it, else update it.

	kt := kit.New()
	var imported bool
	itsmConfig, err := daoSet.ItsmConfig().GetConfig(kt, constant.CreateOrSignApproveItsmServiceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		imported = true
	} else {
		imported = false
	}
	// 自定义模板分隔符为 [[ ]]，例如 [[ .Name ]]，避免和 ITSM 模板变量格式冲突
	tmpl, err := template.New("create_shared_or_sign_approve.json.tpl").Delims("[[", "]]").
		ParseFS(WorkflowTemplates, "templates/create_shared_or_sign_approve.json.tpl")
	if err != nil {
		return err
	}
	stringBuffer := &strings.Builder{}
	if err = tmpl.Execute(stringBuffer, map[string]string{
		"BCSGateway": "",
		"BCSToken":   "",
		"Approvers":  "",
	}); err != nil {
		return err
	}
	mp := map[string]interface{}{}
	if err = json.Unmarshal([]byte(stringBuffer.String()), &mp); err != nil {
		return err
	}
	importReq := itsm.ImportServiceReq{
		Key:             "request",
		Name:            "创建上线或签审批",
		Desc:            "创建上线或签审批",
		CatelogID:       catalogID,
		Owners:          "admin",
		CanTicketAgency: false,
		IsValid:         true,
		DisplayType:     "OPEN",
		DisplayRole:     "",
		Source:          "custom",
		ProjectKey:      "0",
		Workflow:        mp,
	}
	var serviceID int
	if !imported {
		serviceID, err = itsm.ImportService(importReq)
		if err != nil {
			return err
		}
		workflowId, err := itsm.GetWorkflowByService(serviceID)
		if err != nil {
			return err
		}

		stateApproveId, err := itsm.GetStateApproveByWorkfolw(workflowId)
		if err != nil {
			return err
		}
		return daoSet.ItsmConfig().SetConfig(kt, &table.ItsmConfig{
			Key:            constant.CreateOrSignApproveItsmServiceID,
			Value:          serviceID,
			WorkflowId:     workflowId,
			StateApproveId: stateApproveId,
		})
	}

	return itsm.UpdateService(itsm.UpdateServiceReq{
		ID:               itsmConfig.Value,
		ImportServiceReq: importReq,
	})
}
