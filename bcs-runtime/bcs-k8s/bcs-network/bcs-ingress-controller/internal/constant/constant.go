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

// Package constant variables
package constant

var (
	// Layer4Protocol 四层协议
	Layer4Protocol = []string{ProtocolTCP, ProtocolUDP, ProtocolTCPSSL, ProtocolQUIC}

	// Layer7Protocol 七层协议
	Layer7Protocol = []string{ProtocolHTTP, ProtocolHTTPS}
)

const (
	// ProtocolTCP protocol of TCP
	ProtocolTCP = "TCP"
	// ProtocolUDP protocol of UDP
	ProtocolUDP = "UDP"
	// ProtocolHTTPS protocol of HTTPS
	ProtocolHTTPS = "HTTPS"
	// ProtocolHTTP protocol of HTTP
	ProtocolHTTP = "HTTP"
	// ProtocolTCPSSL protocol of TCP_SSL
	ProtocolTCPSSL = "TCP_SSL"
	// ProtocolQUIC protocol of QUIC
	ProtocolQUIC = "QUIC"
	// PortPoolPortProtocolTCPUDP protocol of port in pool is tcp&udp
	PortPoolPortProtocolTCPUDP = "TCP_UDP"

	// LoadBalancerTypeLoadBalancer default type load balancer
	LoadBalancerTypeLoadBalancer = "loadbalancer"
	// LoadBalancerTypeApplicationGateway type for azure application gateway
	LoadBalancerTypeApplicationGateway = "applicationgateway"

	// ProtocolLayerDefault protocol layer default
	ProtocolLayerDefault = "default"
	// ProtocolLayerTransport protocol layer transport
	ProtocolLayerTransport = "transport"
	// ProtocolLayerApplication protocol layer application
	ProtocolLayerApplication = "application"

	// FinalizerNameBcsIngressController finalizer name of bcs ingress controller
	FinalizerNameBcsIngressController = "ingresscontroller.bkbcs.tencent.com"
	// FinalizerNameUptimeCheck finalizer name of uptime check cleaning
	FinalizerNameUptimeCheck = "uptimecheck.bkbcs.tencent.com"
	// CloudTencent tencent cloud
	CloudTencent = "tencentcloud"
	// CloudAWS aws cloud
	CloudAWS = "aws"
	// CloudGCP gcp cloud
	CloudGCP = "gcp"
	// CloudAzure Azure cloud
	CloudAzure = "azure"

	// EnvNameIsTCPUDPPortReuse env name for option if the loadbalancer provider support tcp udp port reuse
	// if enabled, we will find protocol info in 4 layer listener name
	EnvNameIsTCPUDPPortReuse = "TCP_UDP_PORT_REUSE"
	// EnvNameIsBulkMode env name for option if use bulk interface for cloud lb
	EnvNameIsBulkMode = "IS_BULK_MODE"
	// EnvNamePodIPs env name for pod ips
	EnvNamePodIPs = "POD_IPS"
	// EnvNameImageTag env name for controller image tag
	EnvNameImageTag = "IMAGE_TAG"

	// EnvNameBkBCSClusterID env name for bcs_cluster_id
	EnvNameBkBCSClusterID = "BK_BCS_CLUSTER_ID"
	// EnvNameBkBizID env name for bk_biz_id
	EnvNameBkBizID = "BK_BIZ_ID"
	// EnvNameBkAppCode env name for bk_app_code
	EnvNameBkAppCode = "BK_APP_CODE"
	// EnvNameBkAppSecret env name for bk_app_secret
	EnvNameBkAppSecret = "BK_APP_SECRET"

	// EnvNamePortLeakThresholdSeconds env name for port leak threshold seconds
	EnvNamePortLeakThresholdSeconds = "PORT_LEAK_THRESHOLD_SECONDS"

	// EnvNameListenerNameValidateMode env name for enable listener name validate
	// option value:
	// empty: default value, means disable validate
	// NORMAL: enable validate, but not validate cluster ID
	// STRICT: enable validate, and validate cluster ID
	EnvNameListenerNameValidateMode = "LISTENER_NAME_VALIDATE_MODE"

	// ListenerNameValidateModeClose close listener name validate
	ListenerNameValidateModeClose = "CLOSE"
	// ListenerNameValidateModeNormal normal mode for listener name validate
	// exp: lb-xxx-protocol-port / lb-xxx-port
	ListenerNameValidateModeNormal = "NORMAL"
	// ListenerNameValidateModeStrict strict mode for listener name validate
	// exp: BCS-K8S-XXX-lb-xxx-protocol-port / BCS-K8S-XXX-lb-xxx-port
	ListenerNameValidateModeStrict = "STRICT"

	// DelimiterForLbID delimiter for lb id
	DelimiterForLbID = ":"

	// PortPoolStatusReady ready status for port pool
	PortPoolStatusReady = "Ready"
	// PortPoolStatusNotReady not ready status for port pool
	PortPoolStatusNotReady = "NotReady"

	// PortPoolItemStatusError error status of port pool item
	PortPoolItemStatusError = "Error"
	// PortPoolItemStatusInitialize initial status of port pool item
	PortPoolItemStatusInitialize = "Initialize"
	// PortPoolItemStatusReady ready status of port pool item
	PortPoolItemStatusReady = "Ready"
	// PortPoolItemStatusNotReady the status of port pool item is not ready
	PortPoolItemStatusNotReady = "NotReady"
	// PortPoolItemStatusDeleting deleting status of port pool item
	PortPoolItemStatusDeleting = "Deleting"

	// PortPoolItemMessageReady ready message for port pool item
	PortPoolItemMessageReady = "Ready"

	// PortBindingItemStatusInitializing the status of port binding item is initializing
	// means that binding info is not passed to listener yet
	PortBindingItemStatusInitializing = "Initializeing"
	// PortBindingItemStatusNotReady the status of port binding item is not ready
	PortBindingItemStatusNotReady = "NotReady"
	// PortBindingItemStatusReady the status of port binding item is ready
	PortBindingItemStatusReady = "Ready"
	// PortBindingItemStatusDeleting the port binding item is in deleting
	PortBindingItemStatusDeleting = "Deleting"
	// PortBindingItemStatusCleaned the listener of the port binding item is cleaned
	PortBindingItemStatusCleaned = "Cleaned"
	// PortBindingStatusNotReady the status of port binding is not ready
	PortBindingStatusNotReady = "NotReady"
	// PortBindingStatusReady the status of port binding is ready
	PortBindingStatusReady = "Ready"
	// PortBindingStatusCleaning the listener of the port binding is being cleaned
	PortBindingStatusCleaning = "Cleaning"
	// PortBindingStatusCleaned the listener of the port binding is all cleaned
	PortBindingStatusCleaned = "Cleaned"

	// AnnotationForPortBindingNotReadyTimestamp 记录PortBinding上一次被记为NotReady的时间
	AnnotationForPortBindingNotReadyTimestamp = "unready_timestamp.networkextension.bkbcs.tencent.com"

	// AnnotationForPodStatusReady pod status ready
	AnnotationForPodStatusReady = "Ready"
	// AnnotationForPodStatusNotReady pod status not ready
	AnnotationForPodStatusNotReady = "NotReady"

	// AnnotationForPortPool annotation for claims for port pool 声明是否需要注入端口，值为true/ false
	AnnotationForPortPool = "portpools.networkextension.bkbcs.tencent.com"
	// AnnotationForPortPoolPorts annotation for port pool ports 声明需要注入的端口池、协议、对应Pod端口等信息
	AnnotationForPortPoolPorts = "ports.portpools.networkextension.bkbcs.tencent.com"
	// AnnotationForPortPoolBindings annotation for port pool bindings 分配的端口信息，创建后通过webhook注入
	AnnotationForPortPoolBindings = "poolbindings.portpool.networkextension.bkbcs.tencent.com"
	// AnnotationForPortPoolBindingStatus annotation for port pool ports binding status 声明端口绑定是否成功，值为Ready/NotReady
	AnnotationForPortPoolBindingStatus = "status.portpools.networkextension.bkbcs.tencent.com"
	// AnnotationForPortPoolReadinessGate port pool readiness gate 声明是否需要为Pod写入端口绑定ReadinessGate
	AnnotationForPortPoolReadinessGate = "readinessgate.portpools.networkextension.bkbcs.tencent.com"

	// ConditionTypeBcsIngressPortBinding readiness gate condition type for port binding of bcs-ingress-controller
	ConditionTypeBcsIngressPortBinding = "networkextension.bkbcs.tencent.com/portbinding-ready"
	// ConditionReasonReadyBcsIngressPortBinding ready reason for port binding condition
	ConditionReasonReadyBcsIngressPortBinding = "Ready"
	// ConditionMessageReadyBcsIngressPortBinding ready message for port binding condition
	ConditionMessageReadyBcsIngressPortBinding = "ports ares binded for the pod"
	// ConditionReasonNotReadyBcsIngressPortBinding unready reason for port binding condition
	ConditionReasonNotReadyBcsIngressPortBinding = "NotReady"
	// ConditionMessageNotReadyBcsIngressPortBinding unready message for port binding condition
	ConditionMessageNotReadyBcsIngressPortBinding = "port are not bound to the pod"

	// PatchOperationAdd patch add operation
	PatchOperationAdd = "add"
	// PatchOperationReplace patch replace operation
	PatchOperationReplace = "replace"
	// PatchOperationRemove patch remove operation
	PatchOperationRemove = "remove"
	// PatchPathPodAnnotations annotations path for patch operation
	PatchPathPodAnnotations = "/metadata/annotations"
	// PatchPathContainerEnv container env path for patch operation
	PatchPathContainerEnv = "/spec/containers/%v/env"
	// PathPathInitContainerEnv init container env path for patch operation
	PathPathInitContainerEnv = "/spec/initContainers/%v/env"
	// PatchPathPodReadinessGate readiness gate path for patch operation
	PatchPathPodReadinessGate = "/spec/readinessGates"

	// EnvVIPsPrefixForPortPoolPort env prefix for port in port pool
	EnvVIPsPrefixForPortPoolPort = "BCS_PORTPOOL_PORT_VIPLIST_"

	// MaxPortQuantityForEachLoadbalancer max port quantity for each loadbalancer
	MaxPortQuantityForEachLoadbalancer = 4000

	// PortPoolItemProtocolDelimiter separate protocol in portpool item, like "TCP,UDP"
	PortPoolItemProtocolDelimiter = ","

	// LoadBalanceCheckFormatWithAp regular expression for check lb format "ap-xxxxx:lb-xxxxx"
	LoadBalanceCheckFormatWithApLbID = "^(ap|na|eu|sa)-[A-Za-z0-9-]+:lb-[A-Za-z0-9]+"
	// LoadBalanceCheckFormat regular expression for check lb format "lb-xxxxx"
	LoadBalanceCheckFormat = "^lb-[A-Za-z0-9]+"
	// LoadBalanceCheckFormatWithApLbName
	LoadBalanceCheckFormatWithApLbName = "^(ap|na|eu|sa)-[A-Za-z0-9-]+:[A-Za-z0-9]+"

	// LeaderLabel label
	LeaderLabel = "leader"
	// LeaderLabelValueTrue value
	LeaderLabelValueTrue = "true"
	// LeaderLabelValueFalse value
	LeaderLabelValueFalse = "false"

	// EnvIngressPodName env for ingress pod name
	EnvIngressPodName = "INGRESS_POD_NAME"
	// EnvIngressPodNamespace env for ingress pod namespace
	EnvIngressPodNamespace = "INGRESS_POD_NAMESPACE"

	// KindIngress kind of ingress
	KindIngress = "Ingress"
	// KindPortPool kind of port pool
	KindPortPool = "PortPool"
	// KindListener kind of listener
	KindListener = "Listener"
	// KindPortBinding kind of port binding
	KindPortBinding = "PortBinding"
	// KindCRD of CRD
	KindCRD = "CustomResourceDefinition"

	// EventIngressBindFailed Ingress bind failed event
	EventIngressBindFailed = "IngressBindFailed"
	// EventReasonAllocatePortFailed event reason when allocate port failed
	EventReasonAllocatePortFailed = "AllocatePortFailed"
)
