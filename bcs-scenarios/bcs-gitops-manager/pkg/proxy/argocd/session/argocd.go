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

// Package session xxx
package session

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

// ArgoSession purpose: simple revese proxy for argocd according kubernetes service.
// gitops proxy implements http.Handler interface.
type ArgoSession struct {
	option       *options.Options
	store        store.Store
	reverseProxy *httputil.ReverseProxy
}

// NewArgoSession create the session of argoCD
func NewArgoSession() *ArgoSession {
	s := &ArgoSession{
		option: options.GlobalOptions(),
		store:  store.GlobalStore(),
	}
	s.initReverseProxy()
	return s
}

var (
	argoAllowedTypes = map[string]struct{}{
		"application/json":           {},
		"application/grpc-web+proto": {},
	}
)

func (s *ArgoSession) initReverseProxy() {
	s.reverseProxy = &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			// setting login session token for pass through, for http 1.x
			token := s.store.GetToken(request.Context())
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			headerContentType := request.Header.Get("Content-Type")
			if _, ok := argoAllowedTypes[headerContentType]; !ok {
				if strings.Contains(request.RequestURI, "Service") {
					request.Header.Set("Content-Type", "application/grpc-web+proto")
				} else {
					request.Header.Set("Content-Type", "application/json")
				}
			}
			// for http 2
			request.Header.Set("Token", token)
		},
		ErrorHandler: func(res http.ResponseWriter, request *http.Request, e error) {
			requestID := request.Context().Value(traceconst.RequestIDHeaderKey).(string)
			// backend real path with encoded format
			realPath := strings.TrimPrefix(request.URL.RequestURI(), common.GitOpsProxyURL)
			fullPath := fmt.Sprintf("https://%s%s", s.option.GitOps.Service, realPath)
			if !utils.IsContextCanceled(e) {
				metric.ManagerArgoProxyFailed.WithLabelValues().Inc()
				blog.Errorf("RequestID[%s] GitOps proxy %s failure, %s. header: %+v",
					requestID, fullPath, e.Error(), request.Header)
			}
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("gitops proxy session failure, requestID=" + requestID)) // nolint
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
		},
		ModifyResponse: func(r *http.Response) error {
			requestID := r.Request.Context().Value(traceconst.RequestIDHeaderKey).(string)
			// backend real path with encoded format
			realPath := strings.TrimPrefix(r.Request.URL.RequestURI(), common.GitOpsProxyURL)
			fullPath := fmt.Sprintf("https://%s%s", s.option.GitOps.Service, realPath)
			blog.Infof("RequestID[%s] gitops proxy %s response header details: %+v, status %s, code: %d",
				requestID, fullPath, r.Header, r.Status, r.StatusCode)
			return nil
		},
	}
}

// ServeHTTP http.Handler implementation
func (s *ArgoSession) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestID := req.Context().Value(traceconst.RequestIDHeaderKey).(string)
	// backend real path with encoded format
	realPath := strings.TrimPrefix(req.URL.RequestURI(), common.GitOpsProxyURL)
	// !force https link
	fullPath := fmt.Sprintf("https://%s%s", s.option.GitOps.Service, realPath)
	newURL, err := url.Parse(fullPath)
	if err != nil {
		blog.Errorf("RequestID[%s] GitOps session build new fullpath %s failed, %s",
			requestID, fullPath, err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("URL conversion failure in manager")) // nolint
		return
	}
	req.URL = newURL
	req.Header.Set(traceconst.RequestIDHeaderKey, requestID)
	// all ready to serve
	blog.Infof("RequestID[%s] gitops serve %s %s", requestID, req.Method, fullPath)
	s.reverseProxy.ServeHTTP(rw, req)
}
