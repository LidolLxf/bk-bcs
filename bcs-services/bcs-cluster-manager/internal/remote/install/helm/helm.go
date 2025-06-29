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

// Package helm for helm
package helm

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"github.com/avast/retry-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	// Helm helmInstall
	Helm install.InstallerType = "helm"
)

// HelmInstaller is the helm installer
type HelmInstaller struct { // nolint
	projectID        string
	clusterID        string
	releaseNamespace string
	releaseName      string

	chartName    string
	isPublicRepo bool
	repo         string
	debug        bool

	client helmmanager.HelmManagerClient
	close  func()
}

// HelmOptions xxx
type HelmOptions struct { // nolint
	ProjectID   string
	ClusterID   string
	Namespace   string
	ReleaseName string
	ChartName   string
	IsPublic    bool
}

// NewHelmInstaller creates a new helm installer
func NewHelmInstaller(opts HelmOptions, client *HelmClient,
	debug bool) (*HelmInstaller, error) {
	hi := &HelmInstaller{
		projectID:        opts.ProjectID,
		clusterID:        opts.ClusterID,
		releaseNamespace: opts.Namespace,
		releaseName:      opts.ReleaseName,
		chartName:        opts.ChartName,
		isPublicRepo:     opts.IsPublic,
		debug:            debug,
	}

	cli, conClose, err := client.GetHelmManagerClient()
	if err != nil {
		blog.Errorf("NewHelmInstaller GetHelmManagerClient failed: %v", err)
		return nil, err
	}
	hi.client = cli
	hi.close = conClose

	return hi, nil
}

var _ install.Installer = &HelmInstaller{}

// IsInstalled returns whether the app is installed
func (h *HelmInstaller) IsInstalled(ctx context.Context, clusterID string) (bool, error) {
	if h.debug {
		return true, nil
	}

	start := time.Now()
	resp, err := h.client.GetReleaseDetailV1(ctx, &helmmanager.GetReleaseDetailV1Req{
		ProjectCode: &h.projectID,
		ClusterID:   &clusterID,
		Namespace:   &h.releaseNamespace,
		Name:        &h.releaseName,
	})
	if err != nil {
		metrics.ReportLibRequestMetric("helm", "GetReleaseDetailV1", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("[HelmInstaller] GetReleaseDetail failed, err: %s", err.Error())
		return false, err
	}
	metrics.ReportLibRequestMetric("helm", "GetReleaseDetailV1", "grpc", metrics.LibCallStatusOK, start)
	if resp == nil {
		blog.Errorf("[HelmInstaller] GetReleaseDetail failed, resp is empty")
		return false, fmt.Errorf("GetReleaseDetail failed, resp is empty")
	}
	// not found release
	if resp.Code != nil && *resp.Code != 0 {
		blog.Errorf("[HelmInstaller] GetReleaseDetail failed, code: %d, message: %s", resp.Code, resp.Message)
		return false, nil
	}

	blog.Infof("[HelmInstaller] [%s:%s] GetReleaseDetail success[%s:%s] status: %s",
		resp.Data.Chart, resp.Data.ChartVersion, resp.Data.Namespace, resp.Data.Name, resp.Data.Status)

	return true, nil
}

func (h *HelmInstaller) getChartLatestVersion(ctx context.Context, project string, repo, chart string) (string, error) {
	start := time.Now()
	resp, err := h.client.GetChartDetailV1(ctx, &helmmanager.GetChartDetailV1Req{
		ProjectCode: &project,
		RepoName:    &repo,
		Name:        &chart,
	})
	if err != nil {
		metrics.ReportLibRequestMetric("helm", "GetChartDetailV1", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("[HelmInstaller] getChartLatestVersion failed: %v", err)
		return "", err
	}
	metrics.ReportLibRequestMetric("helm", "GetChartDetailV1", "grpc", metrics.LibCallStatusOK, start)

	if (resp.Code != nil && *resp.Code != 0) || (resp.Result != nil && !*resp.Result) {
		blog.Errorf("[HelmInstaller] getChartLatestVersion[%s] failed: %v", *resp.RequestID, *resp.Message)
		return "", err
	}

	return *resp.Data.LatestVersion, nil
}

func (h *HelmInstaller) setRepo() {
	// default use public-repo
	if h.isPublicRepo || h.repo == "" {
		h.repo = types.PubicRepo
	}
}

// Install installs the app
func (h *HelmInstaller) Install(ctx context.Context, clusterID, values string) error {
	if h.debug {
		return nil
	}

	h.setRepo()
	// get chart latest version
	version, err := h.getChartLatestVersion(ctx, h.projectID, h.repo, h.chartName)
	if err != nil {
		blog.Errorf("[HelmInstaller] getChartLatestVersion failed: %v", err)
		return err
	}

	// create app
	req := &helmmanager.InstallReleaseV1Req{
		ProjectCode: &h.projectID,
		ClusterID:   &clusterID,
		Namespace:   &h.releaseNamespace,
		Name:        &h.releaseName,
		Repository:  &h.repo,
		Chart:       &h.chartName,
		Version:     &version,
		Values:      []string{values},
		Args:        install.InstallDefaultArgsFlag,
	}

	resp := &helmmanager.InstallReleaseV1Resp{}
	err = retry.Do(func() error {
		start := time.Now()
		resp, err = h.client.InstallReleaseV1(ctx, req)
		if err != nil {
			metrics.ReportLibRequestMetric("helm", "InstallReleaseV1", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("[HelmInstaller] InstallRelease failed, err: %s", err.Error())
			return err
		}
		metrics.ReportLibRequestMetric("helm", "InstallReleaseV1", "grpc", metrics.LibCallStatusOK, start)

		if resp == nil {
			blog.Errorf("[HelmInstaller] InstallRelease failed, resp is empty")
			return fmt.Errorf("InstallRelease failed, resp is empty")
		}

		if (resp.Code != nil && *resp.Code != 0) || (resp.Result != nil && !*resp.Result) {
			blog.Errorf("[HelmInstaller] InstallRelease failed, code: %d, message: %s", *resp.Code, *resp.Message)
			return fmt.Errorf("InstallRelease failed, code: %d, message: %s", *resp.Code, *resp.Message)
		}

		return nil
	}, retry.Attempts(types.RetryCount), retry.Delay(types.DefaultTimeOut), retry.DelayType(retry.FixedDelay))
	if err != nil {
		return fmt.Errorf("call api HelmInstaller InstallRelease failed: %v, resp: %s", err, utils.ToJSONString(resp))
	}

	return nil
}

// Upgrade upgrades the app
func (h *HelmInstaller) Upgrade(ctx context.Context, clusterID, values string) error {
	if h.debug {
		return nil
	}

	// upgrade need app status deployed
	ok, err := h.CheckAppStatus(ctx, clusterID, time.Minute*10, true)
	if err != nil {
		blog.Errorf("[HelmInstaller] Upgrade CheckAppStatus failed: %v", err)
		return err
	}
	if !ok {
		return fmt.Errorf("[HelmInstaller] Upgrade release %s status abnormal", h.releaseName)
	}

	h.setRepo()
	// get chart latest version
	/*
		version, err := h.getChartLatestVersion(h.projectID, h.repo, h.chartName)
		if err != nil {
			blog.Errorf("[HelmInstaller] getChartLatestVersion failed: %v", err)
			return err
		}
	*/

	// update app: default not update chart version
	req := &helmmanager.UpgradeReleaseV1Req{
		ProjectCode: &h.projectID,
		ClusterID:   &clusterID,
		Namespace:   &h.releaseNamespace,
		Name:        &h.releaseName,
		Repository:  &h.repo,
		Chart:       &h.chartName,
		//Version:     version,
		Values: []string{values},
		Args:   install.UpgradeDefaultArgsFlag,
	}

	start := time.Now()
	resp, err := h.client.UpgradeReleaseV1(ctx, req)
	if err != nil {
		metrics.ReportLibRequestMetric("helm", "UpgradeReleaseV1", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("[HelmInstaller] UpgradeRelease failed, err: %s", err.Error())
		return err
	}
	metrics.ReportLibRequestMetric("helm", "UpgradeReleaseV1", "grpc", metrics.LibCallStatusOK, start)
	if resp == nil {
		blog.Errorf("[HelmInstaller] UpgradeRelease failed, resp is empty")
		return fmt.Errorf("UpgradeRelease failed, resp is empty")
	}
	if resp.Code != nil && *resp.Code != 0 {
		blog.Errorf("[HelmInstaller] UpgradeRelease failed, code: %d, message: %s", resp.Code, resp.Message)
		return fmt.Errorf("UpgradeRelease failed, code: %d, message: %s, requestID: %s", *resp.Code, *resp.Message,
			*resp.RequestID)
	}

	return nil
}

// Uninstall uninstalls the app
func (h *HelmInstaller) Uninstall(ctx context.Context, clusterID string) error {
	if h.debug {
		return nil
	}

	// get project cluster release
	ok, err := h.IsInstalled(ctx, clusterID)
	if err != nil {
		blog.Errorf("[HelmInstaller] check app installed failed, err: %s", err.Error())
		return err
	}
	if !ok {
		blog.Infof("app %s not installed", h.releaseName)
		return nil
	}

	start := time.Now()
	// delete app
	resp, err := h.client.UninstallReleaseV1(ctx, &helmmanager.UninstallReleaseV1Req{
		ProjectCode: &h.projectID,
		Name:        &h.releaseName,
		Namespace:   &h.releaseNamespace,
		ClusterID:   &clusterID,
	})
	if err != nil {
		metrics.ReportLibRequestMetric("helm", "UninstallReleaseV1", "grpc", metrics.LibCallStatusErr, start)
		blog.Errorf("[HelmInstaller] delete app failed, err: %s", err.Error())
		return err
	}
	metrics.ReportLibRequestMetric("helm", "UninstallReleaseV1", "grpc", metrics.LibCallStatusOK, start)
	if resp.Code != nil && *resp.Code != 0 {
		blog.Errorf("[HelmInstaller] UninstallRelease failed, code: %d, message: %s", *resp.Code, *resp.Message)
		return fmt.Errorf("UninstallRelease failed, code: %d, message: %s, requestID: %s", *resp.Code, *resp.Message,
			*resp.RequestID)
	}

	blog.Infof("[HelmInstaller] delete app successful[%s:%s:%v]", clusterID, h.releaseNamespace, h.releaseName)
	return nil
}

// CheckAppStatus check app install status
func (h *HelmInstaller) CheckAppStatus(
	ctx context.Context, clusterID string, timeout time.Duration, pre bool) (bool, error) {
	if h.debug {
		return true, nil
	}

	// get project cluster appID
	ok, err := h.IsInstalled(ctx, clusterID)
	if err != nil {
		blog.Errorf("[HelmInstaller] check app installed failed, err: %s", err.Error())
		return false, err
	}
	if !ok {
		blog.Errorf("app %s not installed", h.releaseName)
		return false, fmt.Errorf("app %s not installed", h.releaseName)
	}

	// 等待应用正常
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err = loop.LoopDoFunc(ctx, func() error {
		start := time.Now()
		// get app
		resp, err := h.client.GetReleaseDetailV1(ctx, &helmmanager.GetReleaseDetailV1Req{ // nolint
			ProjectCode: &h.projectID,
			ClusterID:   &clusterID,
			Namespace:   &h.releaseNamespace,
			Name:        &h.releaseName,
		})
		if err != nil {
			metrics.ReportLibRequestMetric("helm", "GetReleaseDetailV1", "grpc", metrics.LibCallStatusErr, start)
			blog.Errorf("[HelmInstaller] GetReleaseDetail failed, err: %s", err.Error())
			return err
		}
		metrics.ReportLibRequestMetric("helm", "GetReleaseDetailV1", "grpc", metrics.LibCallStatusOK, start)
		if resp == nil {
			return fmt.Errorf("[HelmInstaller] GetReleaseDetail failed, resp is empty")
		}
		if resp.Code != nil && *resp.Code != 0 {
			return fmt.Errorf("[HelmInstaller] GetReleaseDetail failed, code: %d, message: %s, requestID: %s",
				*resp.Code, *resp.Message, *resp.RequestID)
		}

		if resp.Data == nil {
			return fmt.Errorf("[HelmInstaller] GetReleaseDetail failed, resp is empty")
		}
		if resp.Data.Status == nil {
			return fmt.Errorf("[HelmInstaller] GetReleaseDetail failed, status is empty")
		}

		blog.Infof("[HelmInstaller] GetReleaseDetail status: %s", *resp.Data.Status)

		// 前置检查
		if pre {
			switch *resp.Data.Status {
			case types.DeployedInstall, types.DeployedRollback, types.DeployedUpgrade, types.FailedInstall,
				types.FailedRollback, types.FailedUpgrade, types.FailedState, types.FailedUninstall:
				return loop.EndLoop
			default:
			}

			blog.Warnf("[HelmInstaller] GetReleaseDetail[%v] is on transitioning, waiting, %s", pre,
				utils.ToJSONString(resp.Data))
			return nil
		}

		// 后置检查

		// 成功状态 / 失败状态 则终止
		switch *resp.Data.Status {
		case types.DeployedInstall, types.DeployedRollback, types.DeployedUpgrade:
			return loop.EndLoop
		case types.FailedInstall, types.FailedRollback, types.FailedUpgrade, types.FailedState:
			return fmt.Errorf("[HelmInstaller] CheckAppStatus[%s] failed: %s", *resp.RequestID, *resp.Data.Status)
		default:
		}

		blog.Warnf("[HelmInstaller] GetReleaseDetail[%v] is on transitioning, waiting, %s", pre,
			utils.ToJSONString(resp.Data))
		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("[HelmInstaller] GetReleaseDetail installed failed, err: %s", err.Error())
		return false, err
	}

	blog.Infof("[HelmInstaller] app install successful[%s:%s:%v]", clusterID, h.releaseNamespace, h.releaseName)
	return true, nil
}

// Close clean operation
func (h *HelmInstaller) Close() {
	if h.close != nil {
		h.close()
	}
}
