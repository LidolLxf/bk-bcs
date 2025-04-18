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

// Package config xxx
package config

import (
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-common/common/encryptv2" // nolint
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	registry "github.com/Tencent/bk-bcs/bcs-common/pkg/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

var userManagerConfig *UserMgrConfig

// GloablIAMClient global iam client
var GloablIAMClient iam.PermClient

// GlobalCryptor global cryptor
var GlobalCryptor encryptv2.Cryptor

// SetGlobalConfig global config
func SetGlobalConfig(config *UserMgrConfig) {
	userManagerConfig = config
}

// GetGlobalConfig global config
func GetGlobalConfig() *UserMgrConfig {
	return userManagerConfig
}

// CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

// RedisConfig is configuration of Redis
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	MasterName   string
	RedisMode    string
	DialTimeout  int
	ReadTimeout  int
	WriteTimeout int
	PoolSize     int
	MinIdleConns int
	IdleTimeout  int
}

// Encrypt define encrypt config
type Encrypt struct {
	Enable    bool          `json:"enable" yaml:"enable"`
	Algorithm string        `json:"algorithm" yaml:"algorithm"`
	Secret    EncryptSecret `json:"secret" yaml:"secret"`
}

// EncryptSecret define encrypt secret
type EncryptSecret struct {
	Key    string `json:"key" yaml:"key"`
	Secret string `json:"secret" yaml:"secret"`
}

// UserMgrConfig is a configuration of bcs-user-manager
type UserMgrConfig struct {
	Address         string
	IPv6Address     string
	Port            uint
	InsecureAddress string
	InsecurePort    uint
	LocalIp         string
	Sock            string
	MetricPort      uint
	ServCert        *CertConfig
	ClientCert      *CertConfig
	// server http tls authentication
	TlsServerConfig *tls.Config
	// client http tls authentication
	TlsClientConfig *tls.Config

	VerifyClientTLS bool

	DSN             string
	SlowSQLLatency  uint
	RedisDSN        string
	RedisConfig     RedisConfig
	EnableTokenSync bool
	BootStrapUsers  []options.BootStrapUser
	TKE             options.TKEOptions
	PeerToken       string

	IAMConfig  options.IAMConfig
	EtcdConfig registry.CMDOptions

	PermissionSwitch bool
	CommunityEdition bool
	BcsAPI           *options.BcsAPI

	// Encrypt
	Encrypt options.Encrypt

	// 操作记录清理
	Activity options.Activity
}

var (
	// Tke option for sync tke cluster credentials
	Tke options.TKEOptions
	// CliTls for
	CliTls *tls.Config
)

// NewUserMgrConfig create a config object
func NewUserMgrConfig() *UserMgrConfig {
	return &UserMgrConfig{
		Address: "127.0.0.1",
		Port:    80,
		ServCert: &CertConfig{
			CertPasswd: static.ServerCertPwd,
			IsSSL:      false,
		},
		ClientCert: &CertConfig{
			CertPasswd: static.ClientCertPwd,
			IsSSL:      false,
		},
	}
}
