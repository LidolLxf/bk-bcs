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

package actions

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	helmrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
)

// ReleaseUpgradeAction release upgrade action
type ReleaseUpgradeAction struct {
	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	projectCode    string
	projectID      string
	clusterID      string
	name           string
	namespace      string
	repoName       string
	chartName      string
	version        string
	values         []string
	args           []string
	createBy       string
	updateBy       string
	AuthUser       string
	IsShardCluster bool

	contents []byte
	result   *release.HelmUpgradeResult
}

// ReleaseUpgradeActionOption options
type ReleaseUpgradeActionOption struct {
	Model          store.HelmManagerModel
	Platform       repo.Platform
	ReleaseHandler release.Handler

	ProjectCode    string
	ProjectID      string
	ClusterID      string
	Name           string
	Namespace      string
	RepoName       string
	ChartName      string
	Version        string
	Values         []string
	Args           []string
	CreateBy       string
	UpdateBy       string
	AuthUser       string
	IsShardCluster bool
}

// NewReleaseUpgradeAction new release upgrade action
func NewReleaseUpgradeAction(o *ReleaseUpgradeActionOption) *ReleaseUpgradeAction {
	return &ReleaseUpgradeAction{
		model:          o.Model,
		platform:       o.Platform,
		releaseHandler: o.ReleaseHandler,
		projectCode:    o.ProjectCode,
		projectID:      o.ProjectID,
		clusterID:      o.ClusterID,
		name:           o.Name,
		namespace:      o.Namespace,
		repoName:       o.RepoName,
		chartName:      o.ChartName,
		version:        o.Version,
		values:         o.Values,
		args:           o.Args,
		createBy:       o.CreateBy,
		updateBy:       o.UpdateBy,
		AuthUser:       o.AuthUser,
		IsShardCluster: o.IsShardCluster,
	}
}

var _ operation.Operation = &ReleaseUpgradeAction{}

// Action xxx
func (r *ReleaseUpgradeAction) Action() string {
	return "Upgrade"
}

// Name xxx
func (r *ReleaseUpgradeAction) Name() string {
	return fmt.Sprintf("upgrade-%s", r.name)
}

// Prepare xxx
func (r *ReleaseUpgradeAction) Prepare(ctx context.Context) error {
	repository, err := r.model.GetProjectRepository(ctx, r.projectCode, r.repoName)
	if err != nil {
		return fmt.Errorf("get %s/%s repo info in cluster %s error, %s",
			r.namespace, r.name, r.clusterID, err.Error())
	}

	// 下载到具体的chart version信息
	contents, err := r.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		Chart(r.chartName).
		Download(ctx, r.version)
	if err != nil {
		return fmt.Errorf("download chart %s/%s in cluster %s error, %s",
			r.namespace, r.name, r.clusterID, err.Error())
	}

	r.contents = contents
	return nil
}

// Validate xxx
func (r *ReleaseUpgradeAction) Validate(ctx context.Context) error {
	blog.V(5).Infof("start to validate release %s/%s upgrade", r.namespace, r.name)
	// 非真实用户无法在权限中心鉴权，跳过检测
	if len(r.AuthUser) == 0 {
		return nil
	}
	// 如果是共享集群，且集群不属于该项目，说明是用户使用共享集群，需要单独鉴权
	cls, err := clustermanager.GetCluster(ctx, r.clusterID)
	if err != nil {
		return err
	}
	if !r.IsShardCluster || cls.ProjectID == r.projectID {
		return nil
	}

	// get manifest from helm dry run
	result, err := release.UpgradeRelease(r.releaseHandler, r.projectID, r.projectCode, r.clusterID, r.name,
		r.namespace, r.chartName, r.version, r.createBy, r.updateBy, r.args, nil, r.contents, r.values, true)
	if err != nil {
		return err
	}
	if result == nil || result.Release == nil || (result.Release.Manifest == "" && result.Release.Hooks == nil) {
		blog.Infof("release %s/%s is nil in cluster %s", r.namespace, r.name, r.clusterID)
		return nil
	}
	manifest, err := release.GetManifestSimpleHeadFromRelease(result.Release, r.namespace)
	if err != nil {
		return err
	}
	blog.V(5).Infof("release %s/%s has %d manifest", r.namespace, r.name, len(manifest))

	// get server resources
	client, err := component.GetK8SClientByClusterID(r.clusterID)
	if err != nil {
		return err
	}
	resources, err := client.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		return err
	}
	blog.V(5).Infof("cluster %s has %d api-resources", r.clusterID, len(resources))

	permInfo := basePermInfo{
		username:       r.AuthUser,
		projectCode:    r.projectCode,
		projectID:      r.projectID,
		clusterID:      r.clusterID,
		isShardCluster: r.IsShardCluster,
	}
	// check access
	return checkReleaseAccess(manifest, resources, permInfo)
}

// Execute xxx
func (r *ReleaseUpgradeAction) Execute(ctx context.Context) error {
	vls := make([]*release.File, 0, len(r.values))
	for index, v := range r.values {
		vls = append(vls, &release.File{
			Name:    "values-" + strconv.Itoa(index) + ".yaml",
			Content: []byte(v),
		})
	}
	result, err := r.releaseHandler.Cluster(r.clusterID).Upgrade(
		ctx, release.HelmUpgradeConfig{
			ProjectCode: r.projectCode,
			Name:        r.name,
			Namespace:   r.namespace,
			Chart: &release.File{
				Name:    r.chartName + "-" + r.version + ".tgz",
				Content: r.contents,
			},
			Args:   r.args,
			Values: vls,
			PatchTemplateValues: map[string]string{
				common.PTKProjectID: r.projectID,
				common.PTKClusterID: r.clusterID,
				common.PTKNamespace: r.namespace,
				common.PTKCreator:   stringx.ReplaceIllegalChars(r.createBy),
				common.PTKUpdator:   stringx.ReplaceIllegalChars(r.updateBy),
				common.PTKVersion:   r.version,
				common.PTKName:      r.name,
			},
		})
	r.result = result
	if err != nil {
		return fmt.Errorf("upgrade %s/%s in cluster %s error, %s",
			r.namespace, r.name, r.clusterID, err.Error())
	}
	return nil
}

// Done xxx
func (r *ReleaseUpgradeAction) Done(err error) {
	status := helmrelease.StatusDeployed
	message := ""
	if err != nil {
		status = common.ReleaseStatusUpgradeFailed
		message = err.Error()
	}
	rl := entity.M{
		entity.FieldKeyChartName:    r.chartName,
		entity.FieldKeyChartVersion: r.version,
		entity.FieldKeyValues:       r.values,
		entity.FieldKeyArgs:         r.args,
		entity.FieldKeyUpdateBy:     r.updateBy,
		entity.FieldKeyStatus:       status.String(),
		entity.FieldKeyMessage:      message,
	}
	if r.result != nil {
		rl.Update(entity.FieldKeyRevision, r.result.Revision)
	}
	_ = r.model.UpdateRelease(context.Background(), r.clusterID, r.namespace, r.name, rl)
}
