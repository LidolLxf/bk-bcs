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

package qcloud

import (
	"fmt"
	"net"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cidrtree"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var vpcMgr sync.Once

func init() {
	vpcMgr.Do(func() {
		// init VPC manager
		cloudprovider.InitVPCManager(cloudName, &VPCManager{})
	})
}

// VPCManager is the manager for VPC
type VPCManager struct{}

// ListVpcs list vpcs
func (c *VPCManager) ListVpcs(vpcID string, opt *cloudprovider.ListNetworksOption) ([]*proto.CloudVpc, error) {
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	filter := make([]*api.Filter, 0)
	if vpcID != "" {
		filter = append(filter, &api.Filter{Name: "vpc-id", Values: []string{vpcID}})
	}

	vpcs, err := vpcCli.DescribeVpcs(nil, filter)
	if err != nil {
		return nil, err
	}
	result := make([]*proto.CloudVpc, 0)
	for _, v := range vpcs {
		cloudVpc := &proto.CloudVpc{
			Name:     utils.StringPtrToString(v.VpcName),
			VpcId:    utils.StringPtrToString(v.VpcId),
			Ipv4Cidr: utils.StringPtrToString(v.CidrBlock),
			Ipv6Cidr: utils.StringPtrToString(v.Ipv6CidrBlock),
			// 除主网段外, 可扩展的网段
			Cidrs: func() []*proto.AssistantCidr {
				cidrs := make([]*proto.AssistantCidr, 0)

				for _, c := range v.AssistantCidrSet {
					cidrs = append(cidrs, &proto.AssistantCidr{
						Cidr:     utils.StringPtrToString(c.CidrBlock),
						CidrType: int32(utils.Int64PtrToInt64(c.AssistantType)),
					})
				}

				return cidrs
			}(),
		}
		result = append(result, cloudVpc)

		// get free ipNet list
		freeIPNets, err := business.GetFreeIPNets(&opt.CommonOption, vpcID)
		if err != nil {
			blog.Errorf("vpc GetFreeIPNets failed: %v", err)
			continue
		}
		var ipCnt uint32
		for i := range freeIPNets {
			ipNum, err := cidrtree.GetIPNum(freeIPNets[i])
			if err != nil {
				blog.Errorf("vpc GetIPNum failed: %v", err)
				continue
			}
			ipCnt += ipNum
		}
		cloudVpc.AllocateIpNum = ipCnt
	}
	return result, nil
}

// ListSubnets list vpc subnets
func (c *VPCManager) ListSubnets(vpcID, zone string, opt *cloudprovider.ListNetworksOption) ([]*proto.Subnet, error) {
	blog.Infof("ListSubnets input: vpcID/%s", vpcID)
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	filter := make([]*api.Filter, 0)
	filter = append(filter, &api.Filter{Name: "vpc-id", Values: []string{vpcID}})
	if len(zone) > 0 {
		filter = append(filter, &api.Filter{Name: "zone", Values: []string{zone}})
	}

	subnets, err := vpcCli.DescribeSubnets(nil, filter)
	if err != nil {
		return nil, err
	}
	result := make([]*proto.Subnet, 0)
	for _, v := range subnets {
		result = append(result, &proto.Subnet{
			VpcID:                   *v.VpcId,
			SubnetID:                *v.SubnetId,
			SubnetName:              *v.SubnetName,
			CidrRange:               *v.CidrBlock,
			Ipv6CidrRange:           *v.Ipv6CidrBlock,
			Zone:                    *v.Zone,
			AvailableIPAddressCount: *v.AvailableIpAddressCount,
			TotalIpAddressCount:     *v.TotalIpAddressCount,
		})
	}
	return result, nil
}

// CreateSubnets create vpc subnets
func (c *VPCManager) CreateSubnets(opt *cloudprovider.NetworksSubnetOption) (*proto.Subnet, error) {
	blog.Infof("CreateSubnets input: vpcId/%s, subnetName/%s, cidrBlock/%s, zone/%s",
		opt.Subnets.VpcId, opt.Subnets.SubnetName, opt.Subnets.CidrBlock, opt.Subnets.Zone)
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	_, cidrBlock, err := net.ParseCIDR(opt.Subnets.CidrBlock)
	if err != nil {
		return nil, err
	}

	subnet, err := vpcCli.CreateSubnet(opt.Subnets.VpcId, opt.Subnets.SubnetName, opt.Subnets.Zone, cidrBlock)
	if err != nil {
		return nil, err
	}

	return &proto.Subnet{
		VpcID:                   *subnet.VpcId,
		SubnetID:                *subnet.SubnetId,
		SubnetName:              *subnet.SubnetName,
		CidrRange:               *subnet.CidrBlock,
		Ipv6CidrRange:           *subnet.Ipv6CidrBlock,
		Zone:                    *subnet.Zone,
		AvailableIPAddressCount: *subnet.AvailableIpAddressCount,
		TotalIpAddressCount:     *subnet.TotalIpAddressCount,
	}, nil
}

// UpdateSubnets update vpc subnets
func (c *VPCManager) UpdateSubnets(opt *cloudprovider.NetworksSubnetOption) error {
	blog.Infof("UpdateSubnets input: subnetId/%s, subnetName/%s, enableBroadcast/%s",
		opt.Subnets.SubnetId, opt.Subnets.SubnetName, opt.Subnets.EnableBroadcast)
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return err
	}

	return vpcCli.ModifySubnetAttribute(opt.Subnets.SubnetId, opt.Subnets.SubnetName, opt.Subnets.EnableBroadcast)
}

// DeleteSubnets delete vpc subnets
func (c *VPCManager) DeleteSubnets(opt *cloudprovider.NetworksSubnetOption) error {
	blog.Infof("DeleteSubnets input: subnetId/%s", opt.Subnets.SubnetId)
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return err
	}

	return vpcCli.DeleteSubnet(opt.Subnets.SubnetId)
}

// ListSecurityGroups list security groups
func (c *VPCManager) ListSecurityGroups(opt *cloudprovider.ListNetworksOption) ([]*proto.SecurityGroup, error) {
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return nil, err
	}

	sgs, err := vpcCli.DescribeSecurityGroups(nil, nil)
	if err != nil {
		blog.Errorf("ListSecurityGroups DescribeSecurityGroups failed: %v", err)
		return nil, err
	}

	result := make([]*proto.SecurityGroup, 0)
	for _, v := range sgs {
		result = append(result, &proto.SecurityGroup{
			SecurityGroupID:   v.ID,
			SecurityGroupName: v.Name,
			Description:       v.Desc,
		})
	}

	return result, nil
}

// GetCloudNetworkAccountType 查询用户网络类型
func (c *VPCManager) GetCloudNetworkAccountType(opt *cloudprovider.CommonOption) (*proto.CloudAccountType, error) {
	if opt.Region == "" {
		opt.Region = defaultRegion
	}

	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client failed: %v", err)
		return nil, err
	}

	accountType, err := vpcCli.DescribeNetworkAccountTypeRequest()
	if err != nil {
		blog.Errorf("DescribeNetworkAccountType failed: %v", err)
		return nil, err
	}

	return &proto.CloudAccountType{Type: accountType}, nil
}

// ListBandwidthPacks packs
func (c *VPCManager) ListBandwidthPacks(opt *cloudprovider.CommonOption) ([]*proto.BandwidthPackageInfo, error) {
	vpcCli, err := api.NewVPCClient(opt)
	if err != nil {
		blog.Errorf("create VPC client failed: %v", err)
		return nil, err
	}

	bwps, err := vpcCli.DescribeBandwidthPackages(nil, nil)
	if err != nil {
		blog.Errorf("ListBandwidthPacks describeBandwidthPackages failed: %v", err)
		return nil, err
	}

	result := make([]*proto.BandwidthPackageInfo, 0)
	for _, v := range bwps {
		result = append(result, &proto.BandwidthPackageInfo{
			Id:          *v.BandwidthPackageId,
			Name:        *v.BandwidthPackageName,
			NetworkType: *v.NetworkType,
			Status:      *v.Status,
			Bandwidth: func() int32 {
				if v != nil && v.Bandwidth != nil {
					return int32(*v.Bandwidth)
				}
				return 0
			}(),
		})
	}

	return result, nil
}

// CheckConflictInVpcCidr check cidr if conflict with vpc cidrs
func (c *VPCManager) CheckConflictInVpcCidr(vpcID string, cidr string,
	opt *cloudprovider.CheckConflictInVpcCidrOption) ([]string, error) {
	return business.CheckConflictFromVpc(&opt.CommonOption, vpcID, cidr)
}

// AllocateOverlayCidr allocate overlay cidr
func (c *VPCManager) AllocateOverlayCidr(vpcId string, cluster *proto.Cluster, cidrLens []uint32,
	reservedBlocks []*net.IPNet, opt *cloudprovider.CommonOption) ([]string, error) {
	return nil, nil
}

// AddClusterOverlayCidr add cidr to cluster
func (c *VPCManager) AddClusterOverlayCidr(clusterId string, cidrs []string, opt *cloudprovider.CommonOption) error {
	return nil
}

// GetVpcIpUsage get vpc ipTotal/ipSurplus
func (c *VPCManager) GetVpcIpUsage(
	vpcId string, ipType string, reservedBlocks []*net.IPNet, opt *cloudprovider.CommonOption) (uint32, uint32, error) {
	return 0, 0, nil
}

// GetClusterIpUsage get cluster ip usage
func (c *VPCManager) GetClusterIpUsage(clusterId string, ipType string, opt *cloudprovider.CommonOption) (
	uint32, uint32, error) {
	return 0, 0, nil
}

// ListVpcsByPage list vpcs by page
func (c *VPCManager) ListVpcsByPage(opt *cloudprovider.ListNetworksOption) (int64, []*proto.CloudVpcs, error) {
	if opt == nil {
		return 0, nil, fmt.Errorf("opt is nil")
	}
	vpcCli, err := api.NewVPCClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create VPC client when failed: %v", err)
		return 0, nil, err
	}

	filter := make([]*api.Filter, 0)
	if len(opt.VpcIds) > 0 {
		filter = append(filter, &api.Filter{Name: "vpc-id", Values: opt.VpcIds})
	}

	if len(opt.VpcName) > 0 {
		filter = append(filter, &api.Filter{Name: "vpc-name", Values: opt.VpcName})
	}

	vpcs, err := vpcCli.DescribeVpcsByPage(nil, filter, opt.Offset, opt.Limit)
	if err != nil {
		return 0, nil, err
	}
	result := make([]*proto.CloudVpcs, 0)
	for _, v := range vpcs.VpcSet {
		overlayNums, err :=
			getIpNumsAndCidr(&opt.CommonOption, opt.CloudId, utils.StringPtrToString(v.VpcId), *v.CidrBlock, 1)
		if err != nil {
			return 0, nil, err
		}
		underlayNums, err :=
			getIpNumsAndCidr(&opt.CommonOption, opt.CloudId, utils.StringPtrToString(v.VpcId), *v.CidrBlock, 0)
		if err != nil {
			return 0, nil, err
		}
		result = append(result, &proto.CloudVpcs{
			VpcName:                utils.StringPtrToString(v.VpcName),
			VpcID:                  utils.StringPtrToString(v.VpcId),
			Region:                 opt.Region,
			OverlayCidr:            overlayNums.CidrBlock,
			AvailableOverlayIpNum:  uint32(overlayNums.AvailableIpAddressCount),
			AvailableOverlayCidr:   overlayNums.AvailableCidrBlock,
			TotalOverlayIpNum:      uint32(overlayNums.TotalIpAddressCount),
			UnderlayCidr:           underlayNums.CidrBlock,
			AvailableUnderlayIpNum: uint32(underlayNums.AvailableIpAddressCount),
			AvailableUnderlayCidr:  underlayNums.AvailableCidrBlock,
			TotalUnderlayIpNum:     uint32(underlayNums.TotalIpAddressCount),
			CreateTime:             utils.StringPtrToString(v.CreatedTime),
		})
	}
	return utils.Uint64PtrToInt64(vpcs.TotalCount), result, nil
}

// 获取overlay/underlay ip可使用数量, 总量及cidr
func getIpNumsAndCidr(
	opt *cloudprovider.CommonOption, cloudId, vpcId, cidrBlock string, assistantType int) (*cidrtree.VpcInfo, error) {
	vpcInfo := &cidrtree.VpcInfo{}
	switch assistantType {
	case 0:
		freeIPNets, err := business.GetFreeIPNets(opt, vpcId)
		if err != nil {
			return nil, err
		}
		var cidrs []string
		for _, v := range freeIPNets {
			cidrs = append(cidrs, v.String())
		}
		subnets, err := business.GetAllocatedSubnetsInfoByVpc(opt, vpcId)
		if err != nil {
			return nil, err
		}
		// 已使用ip数量
		usedIp := uint32(subnets.TotalIpAddressCount) - uint32(subnets.AvailableIpAddressCount)
		_, allIpNets, err := net.ParseCIDR(cidrBlock)
		if err != nil {
			return nil, err
		}
		totalIpsNums, err := cidrtree.GetIPNetsNum([]*net.IPNet{allIpNets})
		if err != nil {
			return nil, err
		}
		vpcInfo.CidrBlock = []string{cidrBlock}
		vpcInfo.AvailableCidrBlock = cidrs
		vpcInfo.AvailableIpAddressCount = int64(totalIpsNums - usedIp)
		vpcInfo.TotalIpAddressCount = int64(totalIpsNums)
		return vpcInfo, nil
	case 1:
		// 获取可用网段、ip总数量、总网段
		freeIPNets, err := business.GetVpcGrFreeIPNetsAndNums(opt, cloudId, vpcId, nil)
		if err != nil {
			return nil, err
		}
		// 获取已分配的可使用ip数，总ip数
		overlayNums, err := business.GetVpcOverlayCIDRAndIpNum(opt, vpcId)
		if err != nil {
			return nil, err
		}
		// 已使用ip数量
		usedIp := overlayNums.TotalIpAddressCount - overlayNums.AvailableIpAddressCount
		vpcInfo.AvailableIpAddressCount = freeIPNets.TotalIpAddressCount - usedIp

		return vpcInfo, nil
	default:
		return nil, fmt.Errorf("assistantType[%d] not support", assistantType)
	}
}
