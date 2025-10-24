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

// Package component xxxx
package component

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	clientOnce   sync.Once
	globalClient *http.Client
	dialer       = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	// defaultTransport default transport
	defaultTransport http.RoundTripper = &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// NOCC:gas/tls(设计如此)
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
	}
)

// GetHttpClient : 新建Client
func GetHttpClient() *http.Client {
	if globalClient == nil {
		clientOnce.Do(func() {
			globalClient = &http.Client{
				Transport: defaultTransport,
			}
		})
	}
	return globalClient
}

// HttpRequest http 请求
func HttpRequest(ctx context.Context, url, method string, header http.Header, data io.Reader) ([]byte, error) {
	var req *http.Request
	var err error
	if data != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, data)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	resp, err := GetHttpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("http request failed, code: %d, status: %s,message: %s", resp.StatusCode,
			resp.Status, string(body))
	}
	return body, nil
}
