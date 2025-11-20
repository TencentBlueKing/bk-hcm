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

package enumor

// GlobalConfigType global config type
type GlobalConfigType string

const (
	// GlobalConfigResDissolve resource dissolve global config
	GlobalConfigResDissolve GlobalConfigType = "res_dissolve"
	// GlobalConfigTypeRegionDefaultVpc 地域默认vpc
	GlobalConfigTypeRegionDefaultVpc GlobalConfigType = "region_default_vpc"
	// GlobalConfigTypeRegionDefaultSecGroup 地域默认安全组
	GlobalConfigTypeRegionDefaultSecGroup GlobalConfigType = "region_default_security_group"
	// GlobalConfigTypeCvmResetBizIDList 主机重装的业务白名单
	GlobalConfigTypeCvmResetBizIDList GlobalConfigType = "cvm_reset_biz_id_list"
	// GlobalConfigTypeCLBBandwidthPackageRecommend CLB带宽推荐
	GlobalConfigTypeCLBBandwidthPackageRecommend GlobalConfigType = "clb_bandwidth_package_recommend"
	// GlobalConfigTypeBs2ToBkBizIDMap 二级业务配置映射
	// config_value 格式为 JSON 数组，包含完整的业务信息
	GlobalConfigTypeBs2ToBkBizIDMap GlobalConfigType = "bs2_to_bk_biz_id_map"
)

// GlobalConfigResDissolveKey resource dissolve global config key
type GlobalConfigResDissolveKey string

const (
	// GlobalConfigDissolveHostApplyTime resource dissolve host apply time
	GlobalConfigDissolveHostApplyTime GlobalConfigResDissolveKey = "dissolve_host_apply_time"
	// GlobalConfigDissolveApprovalLimit resource dissolve approval limit
	GlobalConfigDissolveApprovalLimit GlobalConfigResDissolveKey = "dissolve_host_approval_limit"
)

// GlobalConfigKeyClbBandPkgRecommend resource global config key for clb bandwidth package recommend
type GlobalConfigKeyClbBandPkgRecommend string

const (
	// GlobalConfigKeyCLBBandwidthPackageRecommend CLB带宽推荐
	GlobalConfigKeyCLBBandwidthPackageRecommend GlobalConfigKeyClbBandPkgRecommend = "clb_bandwidth_package_recommend"
)

// GlobalConfigKeyBs2Biz resource global config key for bs2 to bk biz
type GlobalConfigKeyBs2Biz string

const (
	// GlobalConfigKeyBs2BizMapping 二级业务配置的固定 config_key
	GlobalConfigKeyBs2BizMapping GlobalConfigKeyBs2Biz = "bs2_biz_mapping"
)
