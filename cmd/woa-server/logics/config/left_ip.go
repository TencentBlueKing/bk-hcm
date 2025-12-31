/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
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
	"time"

	"hcm/cmd/woa-server/dal/config/dao"
	"hcm/cmd/woa-server/dal/config/table"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
)

// LeftIPIf provides management interface for operations of left ip config
type LeftIPIf interface {
	// GetLeftIP get left ip config list
	GetLeftIP(kt *kit.Kit, input *types.GetLeftIPParam) (*types.GetLeftIPRst, error)
	// CreateLeftIP creates left ip config
	CreateLeftIP(kt *kit.Kit, input *table.ZoneLeftIP) (mapstr.MapStr, error)
	// UpdateLeftIP updates left ip config
	UpdateLeftIP(kt *kit.Kit, instId int64, input map[string]interface{}) error
	// UpdateLeftIPBatch updates left ip config in batch
	UpdateLeftIPBatch(kt *kit.Kit, cond, update map[string]interface{}) error
	// SyncLeftIP sync zone left ip from yunti
	SyncLeftIP(kt *kit.Kit, input *types.SyncLeftIPParam) error
}

// NewLeftIPOp creates a left ip interface
func NewLeftIPOp(vpc VpcIf, subnet SubnetIf, thirdCli *thirdparty.Client) LeftIPIf {
	return &leftIP{
		cvm:    thirdCli.OldCVM,
		subnet: subnet,
		vpc:    vpc,
	}
}

type leftIP struct {
	cvm    cvmapi.CVMClientInterface
	subnet SubnetIf
	vpc    VpcIf
}

// GetLeftIP get left ip config list
func (l *leftIP) GetLeftIP(kt *kit.Kit, input *types.GetLeftIPParam) (*types.GetLeftIPRst, error) {
	filter, err := input.GetFilter()
	if err != nil {
		logs.Errorf("failed to get filter, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst := &types.GetLeftIPRst{}
	if input.Page.EnableCount {
		cnt, err := dao.Set().ZoneLeftIP().CountZoneLeftIP(kt.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get config left ip count, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.ZoneLeftIP, 0)
		return rst, nil
	}

	insts, err := dao.Set().ZoneLeftIP().FindManyZoneLeftIP(kt.Ctx, input.Page, filter)
	if err != nil {
		logs.Errorf("failed to get config left ip, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// CreateLeftIP creates left ip config
func (l *leftIP) CreateLeftIP(kt *kit.Kit, input *table.ZoneLeftIP) (mapstr.MapStr, error) {
	id, err := dao.Set().ZoneLeftIP().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create config left ip, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	input.ID = id
	if err := dao.Set().ZoneLeftIP().CreateZoneLeftIP(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create config left ip, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": id,
	}

	return rst, nil
}

// UpdateLeftIP updates left ip config
func (l *leftIP) UpdateLeftIP(kt *kit.Kit, instId int64, input map[string]interface{}) error {
	filter := map[string]interface{}{
		"id": instId,
	}

	if err := dao.Set().ZoneLeftIP().UpdateZoneLeftIP(kt.Ctx, filter, input); err != nil {
		logs.Errorf("failed to update left ip, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// UpdateLeftIPBatch updates left ip config in batch
func (l *leftIP) UpdateLeftIPBatch(kt *kit.Kit, cond, update map[string]interface{}) error {
	if err := dao.Set().ZoneLeftIP().UpdateZoneLeftIP(kt.Ctx, cond, update); err != nil {
		logs.Errorf("failed to update left ip, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// SyncLeftIP sync zone left ip from yunti
func (l *leftIP) SyncLeftIP(kt *kit.Kit, input *types.SyncLeftIPParam) error {
	// 1. query subnet from db
	vpc, err := l.vpc.GetRegionDftVpc(kt, input.Region)
	if err != nil {
		logs.Errorf("failed to get default vpc for err: %v, region: %s, rid: %s", err, input.Region, kt.Rid)
		return err
	}

	subnetReq := &types.GetAllSubnetReq{
		Region:     input.Region,
		Zones:      []string{input.Zone},
		CloudVpcID: vpc,
		Name:       "cvm_use_",
	}
	cfgSubnetList, err := l.subnet.GetAllSubnet(kt, subnetReq)
	if err != nil {
		logs.Errorf("failed to find subnet with subnetReq: %+v, err: %v, rid: %s", subnetReq, err, kt.Rid)
		return err
	}

	// 2. query left ip info from cvm
	subnetToLeftIp := make(map[string]*cvmapi.SubnetInfo)
	cvmSubnetList, err := l.querySubnet(kt, input.Region, input.Zone, vpc)
	if err != nil {
		logs.Errorf("failed to get cvm subnet info, err: %v, rid: %s", err, kt.Rid)
		return nil
	}
	for _, subnetItem := range cvmSubnetList {
		subnetToLeftIp[subnetItem.Id] = subnetItem
	}

	leftIPNum := 0
	for _, subnetItem := range cfgSubnetList.Info {
		if subnetToLeftIp[subnetItem.SubnetId] != nil {
			leftIPNum = leftIPNum + subnetToLeftIp[subnetItem.SubnetId].LeftIpNum
		}
	}

	// 3. update left ip info in db
	filterLeftIP := map[string]interface{}{
		"region": input.Region,
		"zone":   input.Zone,
	}

	update := map[string]interface{}{
		"left_ip_num": leftIPNum,
		"update_at":   time.Now(),
	}

	if err = dao.Set().ZoneLeftIP().UpdateZoneLeftIP(kt.Ctx, filterLeftIP, update); err != nil {
		logs.Errorf("failed to update zone with left ip info in db, err: %v, %s", err, kt.Rid)
		return err
	}

	return nil
}

func (l *leftIP) querySubnet(kt *kit.Kit, region, zone, vpc string) ([]*cvmapi.SubnetInfo, error) {
	subnetReq := cvmapi.SubnetRealParam{
		Region:      region,
		CloudCampus: zone,
		VpcId:       vpc,
	}
	resp, err := l.cvm.QueryRealCvmSubnet(kt, subnetReq)
	if err != nil {
		logs.Errorf("failed to get cvm subnet info, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Result, nil
}
