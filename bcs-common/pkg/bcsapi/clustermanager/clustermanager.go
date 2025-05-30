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

// Package clustermanager xxx
package clustermanager

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	discovery "github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/header"
	headerpkg "github.com/Tencent/bk-bcs/bcs-common/pkg/header"
)

var (
	clientConfig *bcsapi.ClientConfig
)

// SetClientConfig set cluster manager client config
// disc nil 表示使用k8s 内置的service 进行服务访问
func SetClientConfig(tlsConfig *tls.Config, disc *discovery.ModuleDiscovery) {
	clientConfig = &bcsapi.ClientConfig{
		TLSConfig: tlsConfig,
		Discovery: disc,
	}
}

// GetClient get cm client by discovery
func GetClient(innerClientName string) (ClusterManagerClient, func(), error) {
	if clientConfig == nil {
		return nil, nil, bcsapi.ErrNotInited
	}
	var addr string
	if discovery.UseServiceDiscovery() {
		addr = fmt.Sprintf("%s:%d", discovery.ClusterManagerServiceName, discovery.ServiceGrpcPort)
	} else {
		if clientConfig.Discovery == nil {
			return nil, nil, fmt.Errorf("cluster manager module not enable discovery")
		}

		nodeServer, err := clientConfig.Discovery.GetRandomServiceNode()
		if err != nil {
			return nil, nil, err
		}
		addr = nodeServer.Address
	}
	klog.Infof("get cluster manager client with address: %s", addr)
	conf := &bcsapi.Config{
		Hosts:           []string{addr},
		TLSConfig:       clientConfig.TLSConfig,
		InnerClientName: innerClientName,
	}
	cli, closeCon := NewClusterManager(conf)

	return cli, closeCon, nil
}

// NewClusterManager create ClusterManager SDK implementation
func NewClusterManager(config *bcsapi.Config) (ClusterManagerClient, func()) {
	// NOCC: gosec/crypto(没有特殊的安全需求)
	//nolint:gosec
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if len(config.Hosts) == 0 {
		// ! pay more attention for nil return
		return nil, nil
	}
	// create grpc connection
	header := map[string]string{
		"x-content-type":            "application/grpc+proto",
		"Content-Type":              "application/grpc",
		header.InnerClientHeaderKey: config.InnerClientName,
	}
	if len(config.AuthToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", config.AuthToken)
	}
	for k, v := range config.Header {
		header[k] = v
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	auth := &bcsapi.Authentication{InnerClientName: config.InnerClientName}
	// 添加 requestID interceptor
	opts = append(opts, grpc.WithUnaryInterceptor(headerpkg.LaneHeaderInterceptor()))
	if config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		auth.Insecure = true
	}
	opts = append(opts, grpc.WithPerRPCCredentials(auth))
	if config.AuthToken != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(bcsapi.NewTokenAuth(config.AuthToken)))
	}
	var conn *grpc.ClientConn
	var err error
	maxTries := 3
	for i := 0; i < maxTries; i++ {
		selected := r.Intn(1024) % len(config.Hosts)
		addr := config.Hosts[selected]
		conn, err = grpc.Dial(addr, opts...)
		if err != nil {
			klog.Errorf("Create cluster manager grpc client with %s error: %s", addr, err.Error())
			continue
		}
		if conn != nil {
			break
		}
	}
	if conn == nil {
		klog.Errorf("create no cluster manager client after all instance tries")
		return nil, nil
	}

	// init cluster manager client
	// nolint
	return NewClusterManagerClient(conn), func() { conn.Close() }
}
