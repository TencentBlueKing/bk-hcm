/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

//go:generate  mockgen -destination ../mock/tcloud/tcloud_mock.go  -package=mocktcloud -typed -source=interface.go

package tcloud

import (
	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/account"
	typeargstpl "hcm/pkg/adaptor/types/argument-template"
	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/adaptor/types/cert"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/adaptor/types/disk"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/adaptor/types/image"
	"hcm/pkg/adaptor/types/instance-type"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/adaptor/types/region"
	"hcm/pkg/adaptor/types/route-table"
	"hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/adaptor/types/security-group-rule"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/adaptor/types/zone"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/kit"

	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// TCloud adaptor interface for tencent cloud
type TCloud interface {
	ListImage(kt *kit.Kit,
		opt *image.TCloudImageListOption) (*image.TCloudImageListResult, error)
	CreateSubnet(kt *kit.Kit, opt *adtysubnet.TCloudSubnetCreateOption) (*adtysubnet.TCloudSubnet,
		error)
	CreateSubnets(kt *kit.Kit, opt *adtysubnet.TCloudSubnetsCreateOption) ([]adtysubnet.TCloudSubnet,
		error)
	UpdateSubnet(_ *kit.Kit, _ *adtysubnet.TCloudSubnetUpdateOption) error
	DeleteSubnet(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error
	ListSubnet(kt *kit.Kit, opt *core.TCloudListOption) (*adtysubnet.TCloudSubnetListResult, error)
	CountSubnet(kt *kit.Kit, region string) (int32, error)
	ListZone(kt *kit.Kit, opt *zone.TCloudZoneListOption) ([]zone.TCloudZone, error)
	CreateSecurityGroup(kt *kit.Kit, opt *securitygroup.TCloudCreateOption) (*v20170312.SecurityGroup,
		error)
	DeleteSecurityGroup(kt *kit.Kit, opt *securitygroup.TCloudDeleteOption) error
	UpdateSecurityGroup(kt *kit.Kit, opt *securitygroup.TCloudUpdateOption) error
	ListSecurityGroupNew(kt *kit.Kit, opt *securitygroup.TCloudListOption) ([]securitygroup.TCloudSG,
		error)
	CountSecurityGroup(kt *kit.Kit, region string) (int32, error)
	SecurityGroupCvmAssociate(kt *kit.Kit, opt *securitygroup.TCloudAssociateCvmOption) error
	SecurityGroupCvmDisassociate(kt *kit.Kit, opt *securitygroup.TCloudAssociateCvmOption) error
	SecurityGroupCvmBatchAssociate(kt *kit.Kit, opt *securitygroup.TCloudBatchAssociateCvmOption) error
	SecurityGroupCvmBatchDisassociate(kt *kit.Kit, opt *securitygroup.TCloudBatchAssociateCvmOption) error
	ListAccount(kt *kit.Kit) ([]account.TCloudAccount, error)
	CountAccount(kt *kit.Kit) (int32, error)
	GetAccountZoneQuota(kt *kit.Kit, opt *account.GetTCloudAccountZoneQuotaOption) (
		*account.TCloudAccountQuota, error)
	GetAccountInfoBySecret(kt *kit.Kit) (*cloud.TCloudInfoBySecret, error)
	CreateDisk(kt *kit.Kit, opt *disk.TCloudDiskCreateOption) (*poller.BaseDoneResult, error)
	InquiryPriceDisk(kt *kit.Kit, opt *disk.TCloudDiskCreateOption) (
		*cvm.InquiryPriceResult, error)
	ListDisk(kt *kit.Kit, opt *core.TCloudListOption) ([]disk.TCloudDisk, error)
	CountDisk(kt *kit.Kit, region string) (int32, error)
	DeleteDisk(kt *kit.Kit, opt *disk.TCloudDiskDeleteOption) error
	AttachDisk(kt *kit.Kit, opt *disk.TCloudDiskAttachOption) error
	DetachDisk(kt *kit.Kit, opt *disk.TCloudDiskDetachOption) error
	ListEip(kt *kit.Kit, opt *eip.TCloudEipListOption) (*eip.TCloudEipListResult, error)
	CountEip(kt *kit.Kit, region string) (int32, error)
	DeleteEip(kt *kit.Kit, opt *eip.TCloudEipDeleteOption) error
	AssociateEip(kt *kit.Kit, opt *eip.TCloudEipAssociateOption) error
	DisassociateEip(kt *kit.Kit, opt *eip.TCloudEipDisassociateOption) error
	DetermineIPv6Type(kt *kit.Kit, region string, ipv6Addresses []*string) ([]*string,
		[]*string, error,
	)
	CreateEip(kt *kit.Kit, opt *eip.TCloudEipCreateOption) (*poller.BaseDoneResult, error)
	ListRegion(kt *kit.Kit) (*region.TCloudRegionListResult, error)
	GetBillList(kt *kit.Kit, opt *typesBill.TCloudBillListOption) (*billing.DescribeBillDetailResponseParams, error)
	ListInstanceType(kt *kit.Kit, opt *instancetype.TCloudInstanceTypeListOption) (
		[]instancetype.TCloudInstanceType, error,
	)
	UpdateRouteTable(_ *kit.Kit, _ *routetable.TCloudRouteTableUpdateOption) error
	DeleteRouteTable(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error
	ListRouteTable(kt *kit.Kit, opt *core.TCloudListOption) (*routetable.TCloudRouteTableListResult,
		error)
	CountRouteTable(kt *kit.Kit, region string) (int32, error)
	CreateSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.TCloudCreateOption) error
	DeleteSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.TCloudDeleteOption) error
	UpdateSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.TCloudUpdateOption) error
	ListSecurityGroupRule(kt *kit.Kit, opt *securitygrouprule.TCloudListOption) (
		*v20170312.SecurityGroupPolicySet, error)
	CreateVpc(kt *kit.Kit, opt *types.TCloudVpcCreateOption) (*types.TCloudVpc, error)
	UpdateVpc(_ *kit.Kit, _ *types.TCloudVpcUpdateOption) error
	DeleteVpc(kt *kit.Kit, opt *core.BaseRegionalDeleteOption) error
	ListVpc(kt *kit.Kit, opt *core.TCloudListOption) (*types.TCloudVpcListResult, error)
	CountVpc(kt *kit.Kit, region string) (int32, error)
	ListCvm(kt *kit.Kit, opt *cvm.TCloudListOption) ([]cvm.TCloudCvm, error)
	ListCvmWithCount(kt *kit.Kit, opt *cvm.ListCvmWithCountOption) (*cvm.CvmWithCountResp, error)
	CountCvm(kt *kit.Kit, region string) (int32, error)
	DeleteCvm(kt *kit.Kit, opt *cvm.TCloudDeleteOption) error
	StartCvm(kt *kit.Kit, opt *cvm.TCloudStartOption) error
	StopCvm(kt *kit.Kit, opt *cvm.TCloudStopOption) error
	RebootCvm(kt *kit.Kit, opt *cvm.TCloudRebootOption) error
	ResetCvmPwd(kt *kit.Kit, opt *cvm.TCloudResetPwdOption) error
	CreateCvm(kt *kit.Kit, opt *cvm.TCloudCreateOption) (*poller.BaseDoneResult, error)
	InquiryPriceCvm(kt *kit.Kit, opt *cvm.TCloudCreateOption) (
		*cvm.InquiryPriceResult, error)
	ListPoliciesGrantingServiceAccess(kt *kit.Kit, opt *account.TCloudListPolicyOption) (
		[]*v20190116.ListGrantServiceAccessNode, error)
	ListArgsTplAddress(kt *kit.Kit, opt *typeargstpl.TCloudListOption) (
		[]typeargstpl.TCloudArgsTplAddress, uint64, error)
	CreateArgsTplAddress(kt *kit.Kit, opt *typeargstpl.TCloudCreateAddressOption) (*v20170312.AddressTemplate, error)
	DeleteArgsTplAddress(kt *kit.Kit, opt *typeargstpl.TCloudDeleteOption) error
	UpdateArgsTplAddress(kt *kit.Kit, opt *typeargstpl.TCloudUpdateAddressOption) (*poller.BaseDoneResult, error)
	ListArgsTplAddressGroup(kt *kit.Kit, opt *typeargstpl.TCloudListOption) (
		[]typeargstpl.TCloudArgsTplAddressGroup, uint64, error)
	CreateArgsTplAddressGroup(kt *kit.Kit, opt *typeargstpl.TCloudCreateAddressGroupOption) (
		*v20170312.AddressTemplateGroup, error)
	DeleteArgsTplAddressGroup(kt *kit.Kit, opt *typeargstpl.TCloudDeleteOption) error
	UpdateArgsTplAddressGroup(kt *kit.Kit, opt *typeargstpl.TCloudUpdateAddressGroupOption) (
		*poller.BaseDoneResult, error)
	ListArgsTplService(kt *kit.Kit, opt *typeargstpl.TCloudListOption) (
		[]typeargstpl.TCloudArgsTplService, uint64, error)
	CreateArgsTplService(kt *kit.Kit, opt *typeargstpl.TCloudCreateServiceOption) (*v20170312.ServiceTemplate, error)
	DeleteArgsTplService(kt *kit.Kit, opt *typeargstpl.TCloudDeleteOption) error
	UpdateArgsTplService(kt *kit.Kit, opt *typeargstpl.TCloudUpdateServiceOption) (*poller.BaseDoneResult, error)
	ListArgsTplServiceGroup(kt *kit.Kit, opt *typeargstpl.TCloudListOption) (
		[]typeargstpl.TCloudArgsTplServiceGroup, uint64, error)
	CreateArgsTplServiceGroup(kt *kit.Kit, opt *typeargstpl.TCloudCreateServiceGroupOption) (
		*v20170312.ServiceTemplateGroup, error)
	DeleteArgsTplServiceGroup(kt *kit.Kit, opt *typeargstpl.TCloudDeleteOption) error
	UpdateArgsTplServiceGroup(kt *kit.Kit, opt *typeargstpl.TCloudUpdateServiceGroupOption) (
		*poller.BaseDoneResult, error)
	CreateLoadBalancer(kt *kit.Kit, opt *typelb.TCloudCreateClbOption) (*poller.BaseDoneResult, error)
	ListLoadBalancer(kt *kit.Kit, opt *typelb.TCloudListOption) ([]typelb.TCloudClb, error)
	DescribeResources(kt *kit.Kit, opt *typelb.TCloudDescribeResourcesOption) (
		*tclb.DescribeResourcesResponseParams, error)
	DescribeNetworkAccountType(kt *kit.Kit) (*v20170312.DescribeNetworkAccountTypeResponseParams, error)
	CreateCert(kt *kit.Kit, opt *cert.TCloudCreateOption) (*poller.BaseDoneResult, error)
	DeleteCert(kt *kit.Kit, opt *cert.TCloudDeleteOption) error
	ListCert(kt *kit.Kit, opt *cert.TCloudListOption) ([]cert.TCloudCert, error)
	SetLoadBalancerSecurityGroups(kt *kit.Kit, opt *typelb.TCloudSetClbSecurityGroupOption) (
		*tclb.SetLoadBalancerSecurityGroupsResponseParams, error)
	DeleteLoadBalancer(kt *kit.Kit, opt *typelb.TCloudDeleteOption) error
	UpdateLoadBalancer(kt *kit.Kit, opt *typelb.TCloudUpdateOption) (*string, error)
	CreateListener(kt *kit.Kit, opt *typelb.TCloudCreateListenerOption) (*poller.BaseDoneResult, error)
	UpdateListener(kt *kit.Kit, opt *typelb.TCloudUpdateListenerOption) error
	DeleteListener(kt *kit.Kit, opt *typelb.TCloudDeleteListenerOption) error
	CreateRule(kt *kit.Kit, opt *typelb.TCloudCreateRuleOption) (*poller.BaseDoneResult, error)
	UpdateRule(kt *kit.Kit, opt *typelb.TCloudUpdateRuleOption) error
	UpdateDomainAttr(kt *kit.Kit, opt *typelb.TCloudUpdateDomainAttrOption) error
	DeleteRule(kt *kit.Kit, opt *typelb.TCloudDeleteRuleOption) error

	ListListener(kt *kit.Kit, opt *typelb.TCloudListListenersOption) ([]typelb.TCloudListener, error)
	RegisterTargets(kt *kit.Kit, opt *typelb.TCloudRegisterTargetsOption) ([]string, error)
	DeRegisterTargets(kt *kit.Kit, opt *typelb.TCloudRegisterTargetsOption) ([]string, error)
	ModifyTargetPort(kt *kit.Kit, opt *typelb.TCloudTargetPortUpdateOption) error
	ModifyTargetWeight(kt *kit.Kit, opt *typelb.TCloudTargetWeightUpdateOption) error

	ListTargets(kt *kit.Kit, opt *typelb.TCloudListTargetsOption) ([]typelb.TCloudListenerTarget, error)
	ListTargetHealth(kt *kit.Kit, opt *typelb.TCloudListTargetHealthOption) ([]typelb.TCloudTargetHealth, error)

	InquiryPriceLoadBalancer(kt *kit.Kit, opt *typelb.TCloudCreateClbOption) (*typelb.TCloudLBPrice, error)
	ListLoadBalancerQuota(kt *kit.Kit, opt *typelb.ListTCloudLoadBalancerQuotaOption) (
		[]typelb.TCloudLoadBalancerQuota, error)
}
