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

// Package service 提供 Istio 的服务分析器
package service

import (
	"fmt"

	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/analysis/analyzers/util"
	configKube "istio.io/istio/pkg/config/kube"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// PortNameAnalyzer checks the port name of the service
type PortNameAnalyzer struct{}

var _ analysis.Analyzer = &PortNameAnalyzer{}

// Metadata implements Analyzer
func (s *PortNameAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "service.PortNameAnalyzer",
		Description: "Checks the port names associated with each service",
		Inputs: []config.GroupVersionKind{
			gvk.Service,
		},
	}
}

// Analyze implements Analyzer
func (s *PortNameAnalyzer) Analyze(c analysis.Context) {
	c.ForEach(gvk.Service, func(r *resource.Instance) bool {
		svcNs := r.Metadata.FullName.Namespace

		// Skip system namespaces entirely
		if util.IsSystemNamespace(svcNs) {
			return true
		}

		// Skip port name check for istio control plane
		if util.IsIstioControlPlane(r) {
			return true
		}

		s.analyzeService(r, c)
		return true
	})
}

func (s *PortNameAnalyzer) analyzeService(r *resource.Instance, c analysis.Context) {
	svc := r.Message.(*v1.ServiceSpec)
	for i, port := range svc.Ports {
		instance := configKube.ConvertProtocol(port.Port, port.Name, port.Protocol, port.AppProtocol)
		if instance.IsUnsupported() || port.Name == "tcp" && svc.Type == "ExternalName" {

			m := msg.NewPortNameIsNotUnderNamingConvention(
				r, port.Name, int(port.Port), port.TargetPort.String())

			if svc.Type == "ExternalName" {
				m = msg.NewExternalNameServiceTypeInvalidPortName(r)
			}

			if line, ok := util.ErrorLine(r, fmt.Sprintf(util.PortInPorts, i)); ok {
				m.Line = line
			}
			c.Report(gvk.Service, m)
		}
	}
}
