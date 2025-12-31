/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/json"
)

// VpcIf provides management interface for operations of vpc config
type VpcIf interface {
	// GetVpc get vpc type config list
	GetVpc(kt *kit.Kit, regions []string) (*types.GetVpcResult, error)
	// GetVpcList get vpc id list
	GetVpcList(kt *kit.Kit, regions []string) (*types.GetVpcListRst, error)
	// GetRegionDftVpc gets the default vpc of a region.
	GetRegionDftVpc(kt *kit.Kit, region string) (string, error)
	// IsRegionDftVpc check if given vpc is the default vpc of a region.
	IsRegionDftVpc(kt *kit.Kit, vpc string) (bool, error)
	// UpsertRegionDftVpc upsert the default vpc of region.
	UpsertRegionDftVpc(kt *kit.Kit, input []types.RegionDftVpc) error
}

// NewVpcOp creates a vpc interface
func NewVpcOp(client *client.ClientSet, thirdCli *thirdparty.Client) VpcIf {
	return &vpc{
		cvm:    thirdCli.OldCVM,
		client: client,
	}
}

type vpc struct {
	cvm    cvmapi.CVMClientInterface
	client *client.ClientSet
}

// GetVpc get vpc type config list
func (v *vpc) GetVpc(kt *kit.Kit, regions []string) (*types.GetVpcResult, error) {
	// 查询账号信息
	accountID, err := getTCloudZiyanAccount(kt, v.client)
	if err != nil {
		logs.Errorf("get vpc %s account failed, err: %v, rid: %s", enumor.TCloudZiyan, err, kt.Rid)
		return nil, err
	}

	// 构建查询条件
	filterRules := []*filter.AtomRule{
		tools.RuleEqual("vendor", enumor.TCloudZiyan),
		tools.RuleEqual("account_id", accountID),
		tools.RuleIn("region", regions),
		tools.RuleJSONEqual("extension.enable_cvm", "true"),
	}

	vpcList := make([]cloud.Vpc[cloud.TCloudVpcExtension], 0)
	// 从MySQL查询VPC列表
	listReq := &core.ListReq{
		Page:   core.NewDefaultBasePage(),
		Filter: tools.ExpressionAnd(filterRules...),
	}
	for {
		vpcListResult, err := v.client.DataService().TCloudZiyan.Vpc.ListVpcExt(kt.Ctx, kt.Header(), listReq)
		if err != nil {
			logs.Errorf("failed to list vpc, err: %v, vendor: %s, accountID: %s, rid: %s",
				err, enumor.TCloudZiyan, accountID, kt.Rid)
			return nil, err
		}

		vpcList = append(vpcList, vpcListResult.Details...)
		if len(vpcListResult.Details) < int(listReq.Page.Limit) {
			break
		}

		listReq.Page.Start += uint32(listReq.Page.Limit)
	}

	vpcResult := make([]*types.Vpc, 0, len(vpcList))
	for _, vpcDetail := range vpcList {
		vpcResult = append(vpcResult, &types.Vpc{
			BkInstId: vpcDetail.ID,
			Region:   vpcDetail.Region,
			VpcId:    vpcDetail.CloudID,
			VpcName:  vpcDetail.Name,
		})
	}

	rst := &types.GetVpcResult{
		Count: int64(len(vpcResult)),
		Info:  vpcResult,
	}
	return rst, nil
}

// GetVpcList get vpc id list
func (v *vpc) GetVpcList(kt *kit.Kit, regions []string) (*types.GetVpcListRst, error) {
	vpcList, err := v.GetVpc(kt, regions)
	if err != nil {
		return nil, err
	}

	rst := &types.GetVpcListRst{}
	for _, item := range vpcList.Info {
		rst.Info = append(rst.Info, item.VpcId)
	}

	return rst, nil
}

// 请勿继续添加内容，应该通过/config/region/default_vpc/upsert接口添加到db
var regionToVpc = map[string]string{
	"ap-guangzhou":       "vpc-03nkx9tv",
	"ap-tianjin":         "vpc-1yoew5gc",
	"ap-shanghai":        "vpc-2x7lhtse",
	"eu-frankfurt":       "vpc-38klpz7z",
	"ap-singapore":       "vpc-706wf55j",
	"ap-tokyo":           "vpc-8iple1iq",
	"ap-seoul":           "vpc-99wg8fre",
	"ap-hongkong":        "vpc-b5okec48",
	"na-toronto":         "vpc-drefwt2v",
	"ap-xian-ec":         "vpc-efw4kf6r",
	"ap-nanjing":         "vpc-fb7sybzv",
	"ap-chongqing":       "vpc-gelpqsur",
	"ap-shenzhen":        "vpc-kwgem8tj",
	"na-siliconvalley":   "vpc-n040n5bl",
	"ap-hangzhou-ec":     "vpc-puhasca0",
	"ap-fuzhou-ec":       "vpc-hdxonj2q",
	"ap-wuhan-ec":        "vpc-867lsj6w",
	"ap-beijing":         "vpc-bhb0y6g8",
	"ap-jinan-ec":        "vpc-kgepmcdd",
	"ap-chengdu":         "vpc-r1wicnlq",
	"ap-zhengzhou-ec":    "vpc-54mjeaf8",
	"ap-shenyang-ec":     "vpc-rea7a2kc",
	"ap-changsha-ec":     "vpc-erdqk82h",
	"ap-hefei-ec":        "vpc-e0a5jxa7",
	"ap-shijiazhuang-ec": "vpc-6b3vbija",
}

// GetRegionDftVpc gets the default vpc of a region.
func (v *vpc) GetRegionDftVpc(kt *kit.Kit, region string) (string, error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeRegionDefaultVpc),
			tools.RuleEqual("config_key", region),
		),
		Page: core.NewDefaultBasePage(),
	}

	list, err := v.client.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default vpc, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return "", err
	}
	if len(list.Details) == 0 {
		// 兜底兼容逻辑，防止部署时还没添加默认值
		vpcVal, ok := regionToVpc[region]
		if !ok {
			return "", fmt.Errorf("found no default vpc with region: %s", region)
		}
		return vpcVal, nil
	}
	result := new(types.DftVpc)
	if err = json.UnmarshalFromString(string(list.Details[0].ConfigValue), &result); err != nil {
		logs.Errorf("failed to unmarshal vpc, err: %v, region: %s, rid: %s", err, region, kt.Rid)
		return "", err
	}

	return result.VpcID, nil
}

// IsRegionDftVpc check if given vpc is the default vpc of a region.
func (v *vpc) IsRegionDftVpc(kt *kit.Kit, vpc string) (bool, error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeRegionDefaultVpc),
			tools.RuleJSONEqual("config_value.vpc_id", vpc),
		),
		Page: &core.BasePage{Count: true},
	}

	list, err := v.client.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default vpc, err: %v, vpc: %s, rid: %s", err, vpc, kt.Rid)
		return false, err
	}
	if list.Count > 0 {
		return true, nil
	}

	// 兜底兼容逻辑，防止部署时还没添加默认值
	for _, val := range regionToVpc {
		if vpc == val {
			return true, nil
		}
	}

	return false, nil
}

// UpsertRegionDftVpc upsert the default vpc of region.
func (v *vpc) UpsertRegionDftVpc(kt *kit.Kit, input []types.RegionDftVpc) error {
	if len(input) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("input length must be less than %d", constant.BatchOperationMaxLimit)
	}

	regions := make([]string, 0, len(input))
	regionVpcMap := make(map[string]types.DftVpc, len(input))
	for _, regionDftVpc := range input {
		if err := regionDftVpc.Validate(); err != nil {
			return err
		}

		if _, ok := regionVpcMap[regionDftVpc.Region]; ok {
			return fmt.Errorf("found duplicate region: %s", regionDftVpc.Region)
		}

		regions = append(regions, regionDftVpc.Region)
		regionVpcMap[regionDftVpc.Region] = regionDftVpc.DftVpc
	}

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", enumor.GlobalConfigTypeRegionDefaultVpc),
			tools.RuleJsonIn("config_key", regions),
		),
		Page: core.NewDefaultBasePage(),
	}

	list, err := v.client.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default vpc, err: %v, region: %v, rid: %s", err, regions, kt.Rid)
		return err
	}
	existRegionDftVpc := make(map[string]cgconf.GlobalConfig, len(list.Details))
	for _, detail := range list.Details {
		existRegionDftVpc[detail.ConfigKey] = cgconf.GlobalConfig{
			ID:          detail.ID,
			ConfigKey:   detail.ConfigKey,
			ConfigValue: detail.ConfigValue,
			ConfigType:  detail.ConfigType,
			Memo:        detail.Memo,
		}
	}

	update := make([]cgconf.GlobalConfig, 0)
	create := make([]cgconf.GlobalConfigT[any], 0)
	for regionKey, vpcVal := range regionVpcMap {
		if detail, ok := existRegionDftVpc[regionKey]; ok {
			detail.ConfigValue = vpcVal
			update = append(update, detail)
			continue
		}
		create = append(create, cgconf.GlobalConfigT[any]{
			ConfigKey:   regionKey,
			ConfigValue: vpcVal,
			ConfigType:  string(enumor.GlobalConfigTypeRegionDefaultVpc),
		})
	}

	if len(update) != 0 {
		updateReq := &datagconf.BatchUpdateReq{Configs: update}
		if err = v.client.DataService().Global.GlobalConfig.BatchUpdate(kt, updateReq); err != nil {
			logs.Errorf("failed to update region default vpc, err: %v, data: %v, rid: %s", err, update, kt.Rid)
			return err
		}
	}

	if len(create) != 0 {
		createReq := &datagconf.BatchCreateReqT[any]{Configs: create}
		if _, err = v.client.DataService().Global.GlobalConfig.BatchCreate(kt, createReq); err != nil {
			logs.Errorf("failed to create region default vpc, err: %v, data: %v, rid: %s", err, create, kt.Rid)
			return err
		}
	}

	return nil
}
