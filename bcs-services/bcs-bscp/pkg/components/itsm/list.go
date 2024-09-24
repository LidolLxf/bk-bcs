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

// Package itsm xxx
package itsm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

var (
	getTicketInfoPath = "/itsm/ticket_approval_result/"
	listTicketsPath   = "/itsm/get_tickets/"
	limit             = 100
)

// ListTicketsResp itsm list tickets resp
type ListTicketsResp struct {
	CommonResp
	Data ListTicketsData `json:"data"`
}

// ListTicketsData list tickets data
type ListTicketsData struct {
	Page      int           `json:"page"`
	TotalPage int           `json:"total_page"`
	Count     int           `json:"count"`
	Next      string        `json:"next"`
	Previous  string        `json:"previous"`
	Items     []TicketsItem `json:"items"`
}

// TicketsItem ITSM list tickets item
type TicketsItem struct {
	ID            int    `json:"id"`
	SN            string `json:"sn"`
	Title         string `json:"title"`
	CatalogID     int    `json:"catalog_id"`
	ServiceID     int    `json:"service_id"`
	ServiceType   string `json:"service_type"`
	FlowID        int    `json:"flow_id"`
	CurrentStatus string `json:"current_status"`
	CommentID     string `json:"comment_id"`
	IsCommented   bool   `json:"is_commented"`
	UpdatedBy     string `json:"updated_by"`
	UpdateAt      string `json:"update_at"`
	EndAt         string `json:"ent_at"`
	Creator       string `json:"creator"`
	CreateAt      string `json:"creat_at"`
	BkBizID       int    `json:"bk_biz_id"`
	TicketURL     string `json:"ticket_url"`
}

// ListTickets list itsm tickets by sn list
func ListTickets(snList []string) ([]TicketsItem, error) {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	tickets := []TicketsItem{}
	var page = 1
	for {
		reqURL := fmt.Sprintf("%s%s?page=%d&page_size=%d", host, listTicketsPath, page, limit)

		reqData := map[string]interface{}{
			"sns": snList,
		}

		body, err := ItsmRequest(context.Background(), http.MethodPost, reqURL, reqData)
		if err != nil {
			logs.Errorf("request list itsm tickets %v failed, %s", snList, err.Error())
			return nil, fmt.Errorf("request list itsm tickets %v failed, %s", snList, err.Error())
		}
		// 解析返回的body
		resp := &ListTicketsResp{}
		if err := json.Unmarshal(body, resp); err != nil {
			logs.Errorf("parse itsm body error, body: %v", body)
			return nil, err
		}
		if resp.Code != 0 {
			logs.Errorf("list itsm tickets %v failed, msg: %s", snList, resp.Message)
			return nil, errors.New(resp.Message)
		}
		tickets = append(tickets, resp.Data.Items...)
		if page >= resp.Data.TotalPage {
			break
		}
		page++
	}
	return tickets, nil
}

// GetTicketInfoData get ticket data
type GetTicketInfoData struct {
	CommonResp
	Data GetTicketInfoDetail `json:"data"`
}

// GetTicketInfoDetail ITSM get ticket detail item
type GetTicketInfoDetail struct {
	CurrentStatus string         `json:"current_status"`
	CurrentSteps  []CurrentSteps `json:"current_steps"`
}

// CurrentSteps 单据当前步骤
type CurrentSteps struct {
	ActionType     string `json:"action_type"`
	Name           string `json:"name"`
	ProcessorsType string `json:"processors_type"`
	Processors     string `json:"processors"`
	StateId        int    `json:"state_id"`
	Status         string `json:"status"`
}

// GetTicketInfo get itsm ticket info by sn
func GetTicketInfo(sn string) (GetTicketInfoData, error) {
	itsmConf := cc.DataService().ITSM
	// 默认使用网关访问，如果为外部版，则使用ESB访问
	host := itsmConf.GatewayHost
	if itsmConf.External {
		host = itsmConf.Host
	}
	reqURL := fmt.Sprintf("%s%s?sn=%s&sn=%s", host, getTicketInfoPath, sn, "REQ20240924000009")

	// 解析返回的body
	resp := GetTicketInfoData{}
	body, err := ItsmRequest(context.Background(), http.MethodGet, reqURL, nil)
	if err != nil {
		logs.Errorf("request get itsm ticket %v failed, %s", sn, err.Error())
		return resp, fmt.Errorf("request get itsm ticket %v failed, %s", sn, err.Error())
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		logs.Errorf("parse itsm body error, body: %v", body)
		return resp, err
	}
	if resp.Code != 0 {
		logs.Errorf("get itsm ticket %v failed, msg: %s", sn, resp.Message)
		return resp, errors.New(resp.Message)
	}

	return resp, nil
}
