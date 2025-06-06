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

package telemetry

import (
	telemetryapi "istio.io/api/telemetry/v1alpha1"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/analysis"
	"istio.io/istio/pkg/config/resource"
	"istio.io/istio/pkg/config/schema/gvk"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/plugin/istiocheck/msg"
)

// ProviderAnalyzer 提供者分析器
type ProviderAnalyzer struct{}

var _ analysis.Analyzer = &ProviderAnalyzer{}

// Metadata implements Analyzer
func (a *ProviderAnalyzer) Metadata() analysis.Metadata {
	return analysis.Metadata{
		Name:        "telemetry.ProviderAnalyzer",
		Description: "Validates that providers in telemetry resource is valid",
		Inputs: []config.GroupVersionKind{
			gvk.Telemetry,
			gvk.MeshConfig,
		},
	}
}

// Analyze implements Analyzer
func (a *ProviderAnalyzer) Analyze(c analysis.Context) {
	meshConfig := fetchMeshConfig(c)
	if meshConfig.DefaultProviders == nil ||
		len(meshConfig.DefaultProviders.AccessLogging) == 0 {
		c.ForEach(gvk.Telemetry, func(r *resource.Instance) bool {
			telemetry := r.Message.(*telemetryapi.Telemetry)

			for _, l := range telemetry.AccessLogging {
				if l.Disabled != nil && l.Disabled.Value {
					continue
				}
				if len(l.Providers) == 0 {
					c.Report(gvk.Telemetry,
						msg.NewInvalidTelemetryProvider(r, string(r.Metadata.FullName.Name), string(r.Metadata.FullName.Namespace)))
				}
			}

			return true
		})
	}
}
