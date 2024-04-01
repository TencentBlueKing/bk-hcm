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

package common

import (
	"hcm/pkg/adaptor/types"
	"hcm/pkg/adaptor/types/account"
	typeargstpl "hcm/pkg/adaptor/types/argument-template"
	"hcm/pkg/adaptor/types/cert"
	typescvm "hcm/pkg/adaptor/types/cvm"
	typesdisk "hcm/pkg/adaptor/types/disk"
	typeseip "hcm/pkg/adaptor/types/eip"
	firewallrule "hcm/pkg/adaptor/types/firewall-rule"
	typesimage "hcm/pkg/adaptor/types/image"
	typeslb "hcm/pkg/adaptor/types/load-balancer"
	typesni "hcm/pkg/adaptor/types/network-interface"
	typesregion "hcm/pkg/adaptor/types/region"
	typesresourcegroup "hcm/pkg/adaptor/types/resource-group"
	typesroutetable "hcm/pkg/adaptor/types/route-table"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	typessecuritygrouprule "hcm/pkg/adaptor/types/security-group-rule"
	adtysubnet "hcm/pkg/adaptor/types/subnet"
	typeszone "hcm/pkg/adaptor/types/zone"
	cloudcore "hcm/pkg/api/core/cloud"
	coreargstpl "hcm/pkg/api/core/cloud/argument-template"
	corecert "hcm/pkg/api/core/cloud/cert"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	coredisk "hcm/pkg/api/core/cloud/disk"
	coreimage "hcm/pkg/api/core/cloud/image"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	corecloudni "hcm/pkg/api/core/cloud/network-interface"
	coreregion "hcm/pkg/api/core/cloud/region"
	coreresourcegroup "hcm/pkg/api/core/cloud/resource-group"
	cloudcoreroutetable "hcm/pkg/api/core/cloud/route-table"
	coresubaccount "hcm/pkg/api/core/cloud/sub-account"
	corezone "hcm/pkg/api/core/cloud/zone"
	corerecyclerecord "hcm/pkg/api/core/recycle-record"
	dataeip "hcm/pkg/api/data-service/cloud/eip"
)

// CloudResType 云资源类型
type CloudResType interface {
	GetCloudID() string

	typesregion.HuaWeiRegionModel |
		typesregion.AzureRegion |
		typesregion.TCloudRegion |
		typesregion.AwsRegion |
		typesregion.GcpRegion |

		typesresourcegroup.AzureResourceGroup |

		typeszone.TCloudZone |
		typeszone.HuaWeiZone |
		typeszone.GcpZone |
		typeszone.AwsZone |

		typesimage.TCloudImage |
		typesimage.HuaWeiImage |
		typesimage.AwsImage |
		typesimage.AzureImage |
		typesimage.GcpImage |

		typessecuritygrouprule.HuaWeiSGRule |
		typessecuritygrouprule.AwsSGRule |
		typessecuritygrouprule.AzureSGRule |

		types.TCloudVpc |
		types.AwsVpc |
		types.GcpVpc |
		types.HuaWeiVpc |
		types.AzureVpc |

		adtysubnet.TCloudSubnet |
		adtysubnet.AwsSubnet |
		adtysubnet.HuaWeiSubnet |
		adtysubnet.GcpSubnet |
		adtysubnet.AzureSubnet |

		typesdisk.TCloudDisk |
		typesdisk.HuaWeiDisk |
		typesdisk.AwsDisk |
		typesdisk.GcpDisk |
		typesdisk.AzureDisk |

		securitygroup.TCloudSG |
		securitygroup.HuaWeiSG |
		securitygroup.AwsSG |
		securitygroup.AzureSecurityGroup |

		firewallrule.GcpFirewall |

		typescvm.TCloudCvm |
		typescvm.HuaWeiCvm |
		typescvm.AwsCvm |
		typescvm.GcpCvm |
		typescvm.AzureCvm |

		*typeseip.TCloudEip |
		*typeseip.HuaWeiEip |
		*typeseip.GcpEip |
		*typeseip.AwsEip |
		*typeseip.AzureEip |

		typesroutetable.TCloudRouteTable |
		typesroutetable.HuaWeiRouteTable |
		typesroutetable.AwsRouteTable |
		typesroutetable.AzureRouteTable |

		typesni.HuaWeiNI |
		typesni.GcpNI |
		typesni.AzureNI |

		typesroutetable.GcpRoute |
		typesroutetable.TCloudRoute |
		typesroutetable.HuaWeiRoute |
		typesroutetable.AzureRoute |
		typesroutetable.AwsRoute |

		account.TCloudAccount |
		account.HuaWeiAccount |
		account.AwsAccount |
		account.AzureAccount |
		account.GcpAccount |

		corerecyclerecord.EipBindInfo |
		corerecyclerecord.DiskAttachInfo |

		typeargstpl.TCloudArgsTplAddress |
		typeargstpl.TCloudArgsTplAddressGroup |
		typeargstpl.TCloudArgsTplService |
		typeargstpl.TCloudArgsTplServiceGroup |

		cert.TCloudCert |
		typeslb.TCloudClb |
		typeslb.TCloudListener |
		typeslb.TCloudUrlRule |
		typeslb.Backend
}

// DBResType 本地资源类型
type DBResType interface {
	GetID() string
	GetCloudID() string

	coreregion.HuaWeiRegion |
		coreregion.AzureRegion |
		coreregion.TCloudRegion |
		coreregion.AwsRegion |
		coreregion.GcpRegion |

		coreresourcegroup.AzureRG |

		corezone.BaseZone |

		coreimage.Image[coreimage.TCloudExtension] |
		coreimage.Image[coreimage.HuaWeiExtension] |
		coreimage.Image[coreimage.AwsExtension] |
		coreimage.Image[coreimage.AzureExtension] |
		coreimage.Image[coreimage.GcpExtension] |

		cloudcore.HuaWeiSecurityGroupRule |
		cloudcore.AwsSecurityGroupRule |
		cloudcore.AzureSecurityGroupRule |

		cloudcore.Vpc[cloudcore.TCloudVpcExtension] |
		cloudcore.Vpc[cloudcore.AwsVpcExtension] |
		cloudcore.Vpc[cloudcore.GcpVpcExtension] |
		cloudcore.Vpc[cloudcore.HuaWeiVpcExtension] |
		cloudcore.Vpc[cloudcore.AzureVpcExtension] |

		cloudcore.Subnet[cloudcore.TCloudSubnetExtension] |
		cloudcore.Subnet[cloudcore.AwsSubnetExtension] |
		cloudcore.Subnet[cloudcore.HuaWeiSubnetExtension] |
		cloudcore.Subnet[cloudcore.GcpSubnetExtension] |
		cloudcore.Subnet[cloudcore.AzureSubnetExtension] |

		*coredisk.Disk[coredisk.TCloudExtension] |
		*coredisk.Disk[coredisk.HuaWeiExtension] |
		*coredisk.Disk[coredisk.AwsExtension] |
		*coredisk.Disk[coredisk.GcpExtension] |
		*coredisk.Disk[coredisk.AzureExtension] |

		cloudcore.SecurityGroup[cloudcore.TCloudSecurityGroupExtension] |
		cloudcore.SecurityGroup[cloudcore.HuaWeiSecurityGroupExtension] |
		cloudcore.SecurityGroup[cloudcore.AwsSecurityGroupExtension] |
		cloudcore.SecurityGroup[cloudcore.AzureSecurityGroupExtension] |

		cloudcore.GcpFirewallRule |

		corecvm.Cvm[corecvm.TCloudCvmExtension] |
		corecvm.Cvm[corecvm.HuaWeiCvmExtension] |
		corecvm.Cvm[corecvm.AwsCvmExtension] |
		corecvm.Cvm[corecvm.GcpCvmExtension] |
		corecvm.Cvm[corecvm.AzureCvmExtension] |

		*dataeip.EipExtResult[dataeip.TCloudEipExtensionResult] |
		*dataeip.EipExtResult[dataeip.HuaWeiEipExtensionResult] |
		*dataeip.EipExtResult[dataeip.GcpEipExtensionResult] |
		*dataeip.EipExtResult[dataeip.AwsEipExtensionResult] |
		*dataeip.EipExtResult[dataeip.AzureEipExtensionResult] |

		cloudcoreroutetable.TCloudRouteTable |
		cloudcoreroutetable.HuaWeiRouteTable |
		cloudcoreroutetable.AwsRouteTable |
		cloudcoreroutetable.AzureRouteTable |

		corecloudni.NetworkInterface[corecloudni.HuaWeiNIExtension] |
		corecloudni.NetworkInterface[corecloudni.GcpNIExtension] |
		corecloudni.NetworkInterface[corecloudni.AzureNIExtension] |

		cloudcoreroutetable.GcpRoute |
		cloudcoreroutetable.TCloudRoute |
		cloudcoreroutetable.HuaWeiRoute |
		cloudcoreroutetable.AzureRoute |
		cloudcoreroutetable.AwsRoute |

		coresubaccount.SubAccount[coresubaccount.TCloudExtension] |
		coresubaccount.SubAccount[coresubaccount.HuaWeiExtension] |
		coresubaccount.SubAccount[coresubaccount.AwsExtension] |
		coresubaccount.SubAccount[coresubaccount.AzureExtension] |
		coresubaccount.SubAccount[coresubaccount.GcpExtension] |

		corerecyclerecord.EipBindInfo |
		corerecyclerecord.DiskAttachInfo |

		*coreargstpl.ArgsTpl[coreargstpl.TCloudArgsTplExtension] |

		*corecert.Cert[corecert.TCloudCertExtension] |

		corelb.TCloudLoadBalancer |
		corelb.TCloudLbUrlRule |
		corelb.TCloudListener |
		corelb.BaseTarget
}

// Diff 对比云和db资源，划分出新增数据，更新数据，删除数据。
func Diff[CloudType CloudResType, DBType DBResType](dataFromCloud []CloudType, dataFromDB []DBType,
	isChange func(CloudType, DBType) bool) ([]CloudType, map[string]CloudType, []string) {

	dbMap := make(map[string]DBType, len(dataFromDB))
	for _, one := range dataFromDB {
		dbMap[one.GetCloudID()] = one
	}

	newAddData := make([]CloudType, 0)
	updateMap := make(map[string]CloudType, 0)
	for _, oneFromCloud := range dataFromCloud {
		oneFromDB, exist := dbMap[oneFromCloud.GetCloudID()]
		if !exist {
			newAddData = append(newAddData, oneFromCloud)
			continue
		}

		delete(dbMap, oneFromCloud.GetCloudID())
		if isChange(oneFromCloud, oneFromDB) {
			updateMap[oneFromDB.GetID()] = oneFromCloud
		}
	}

	delCloudIDs := make([]string, 0)
	for cloudID := range dbMap {
		delCloudIDs = append(delCloudIDs, cloudID)
	}

	return newAddData, updateMap, delCloudIDs
}
