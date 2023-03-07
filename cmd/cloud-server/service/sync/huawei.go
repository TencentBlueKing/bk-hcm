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

package sync

import (
	"net/http"

	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	protoregion "hcm/pkg/api/data-service/cloud/region"
	proto "hcm/pkg/api/hc-service"
	protocvm "hcm/pkg/api/hc-service/cvm"
	protodisk "hcm/pkg/api/hc-service/disk"
	protoeip "hcm/pkg/api/hc-service/eip"
	protoimage "hcm/pkg/api/hc-service/image"
	protohcregion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/api/hc-service/zone"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncHuaWeiAll sync huawei all resource
func SyncHuaWeiAll(c *client.ClientSet, kit *kit.Kit, header http.Header, accountID string) error {

	regions, err := c.DataService().HuaWei.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.HuaWeiRegionListReq{
			Filter: tools.EqualExpression("type", constant.SyncTimingListHuaWeiRegion),
			Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		},
	)
	if err != nil {
		logs.Errorf("sync list huawei region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, region := range regions.Details {
		err = SyncHuaWeiSG(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do huawei sync sg failed, err: %v, regionID: %s,  rid: %s",
				err, region.RegionID, kit.Rid)
		}

		err = SyncHuaWeiSGRule(c, kit, header, region.RegionID, accountID)
		if err != nil {
			logs.Errorf("sync do huawei sync sg rule failed, err: %v, regionID: %s,  rid: %s",
				err, region.RegionID, kit.Rid)
		}

		err = SyncHuaWeiDisk(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do huawei sync disk failed, err: %v, regionID: %s,  rid: %s",
				err, region.RegionID, kit.Rid)
		}

		err = SyncHuaWeiEip(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do huawei sync eip failed, err: %v, regionID: %s,  rid: %s",
				err, region.RegionID, kit.Rid)
		}

		err = SyncHuaWeiVpc(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do huawei sync vpc failed, err: %v, regionID: %s,  rid: %s",
				err, region.RegionID, kit.Rid)
		}

		err = SyncHuaWeiSubnet(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do huawei sync subnet failed, err: %v, accountID: %s, regionID: %s, rid: %s",
				err, accountID, region.RegionID, kit.Rid)
		}

		err = SyncHuaWeiCvm(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do huawei sync cvm failed, err: %v, regionID: %s,  rid: %s",
				err, region.RegionID, kit.Rid)
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// SyncHuaWeiSubnet ...
func SyncHuaWeiSubnet(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().HuaWei.Subnet.SyncSubnet(
		kit.Ctx,
		header,
		&proto.HuaWeiResourceSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do huawei sync subnet failed, err: %v, accountID: %s, regionID: %s, rid: %s",
			err, accountID, region, kit.Rid)
		return err
	}

	return nil
}

// SyncHuaWeiVpc ...
func SyncHuaWeiVpc(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().HuaWei.Vpc.SyncVpc(
		kit.Ctx,
		header,
		&proto.HuaWeiResourceSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do huawei sync vpc failed, err: %v, accountID: %s, regionID: %s, rid: %s",
			err, accountID, region, kit.Rid)
		return err
	}

	return nil
}

// SyncHuaWeiSG ...
func SyncHuaWeiSG(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().HuaWei.SecurityGroup.SyncSecurityGroup(
		kit.Ctx,
		header,
		&proto.SecurityGroupSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do huawei sync sg failed, err: %v, regionID: %s, rid: %s",
			err, region, kit.Rid)
		return err
	}

	return nil
}

// SyncHuaWeiDisk ...
func SyncHuaWeiDisk(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().HuaWei.Disk.SyncDisk(
		kit.Ctx,
		header,
		&protodisk.DiskSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do huawei sync disk failed, err: %v, regionID: %s,  rid: %s",
			err, region, kit.Rid)
		return err
	}

	return nil
}

// SyncHuaWeiEip ...
func SyncHuaWeiEip(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().HuaWei.Eip.SyncEip(
		kit.Ctx,
		header,
		&protoeip.EipSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do huawei sync eip failed, err: %v, regionID: %s, rid: %s",
			err, region, kit.Rid)
		return err
	}

	return nil
}

// SyncHuaWeiSGRule ...
func SyncHuaWeiSGRule(c *client.ClientSet, kit *kit.Kit, header http.Header,
	region string, accountID string) error {

	start := 0
	for {
		results, err := c.DataService().Global.SecurityGroup.ListSecurityGroup(
			kit.Ctx,
			header,
			&dataproto.SecurityGroupListReq{
				Filter: &filter.Expression{
					Op: filter.And,
					Rules: []filter.RuleFactory{
						&filter.AtomRule{
							Field: "vendor",
							Op:    filter.Equal.Factory(),
							Value: enumor.HuaWei,
						},
						&filter.AtomRule{
							Field: "region",
							Op:    filter.Equal.Factory(),
							Value: region,
						},
						&filter.AtomRule{
							Field: "account_id",
							Op:    filter.Equal.Factory(),
							Value: accountID,
						},
					},
				},
				Page: &core.BasePage{
					Start: uint32(start),
					Limit: core.DefaultMaxPageLimit,
				},
			},
		)
		if err != nil {
			logs.Errorf("list huawei security group failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		if len(results.Details) == 0 {
			break
		}

		for _, v := range results.Details {
			err = c.HCService().HuaWei.SecurityGroup.SyncSecurityGroupRule(
				kit.Ctx,
				header,
				&proto.SecurityGroupSyncReq{
					AccountID: v.AccountID,
					Region:    v.Region,
				},
				v.ID,
			)
			if err != nil {
				logs.Errorf("sync do huawei sync sg  rule failed, err: %v, regionID: %s, rid: %s",
					err, v.Region, kit.Rid)
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < core.DefaultMaxPageLimit {
			break
		}
	}

	return nil
}

// SyncHuaWeiCvm ...
func SyncHuaWeiCvm(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().HuaWei.Cvm.SyncCvm(
		kit.Ctx,
		header,
		&protocvm.CvmSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)

	if err != nil {
		logs.Errorf("sync do huawei sync cvm failed, err: %v, regionID: %s, rid: %s",
			err, region, kit.Rid)
	}

	return err
}

// SyncHuaWeiPublicResource ...
func SyncHuaWeiPublicResource(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	err := SyncHuaWeiRegion(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do huawei sync region failed, err: %v, rid: %s", err, kit.Rid)
	}

	err = SyncHuaWeiZone(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do huawei sync zone failed, err: %v, rid: %s", err, kit.Rid)
	}

	err = SyncHuaWeiImage(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do huawei sync image failed, err: %v, rid: %s", err, kit.Rid)
	}

	return err
}

// SyncHuaWeiImage ...
func SyncHuaWeiImage(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	regions, err := c.DataService().HuaWei.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.HuaWeiRegionListReq{
			Filter: tools.EqualExpression("type", constant.SyncTimingListHuaWeiRegion),
			Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		},
	)
	if err != nil {
		logs.Errorf("sync list huawei region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, region := range regions.Details {
		err = c.HCService().HuaWei.Image.SyncImage(
			kit.Ctx,
			header,
			&protoimage.HuaWeiImageSyncReq{
				AccountID: accountID,
				Region:    region.RegionID,
			},
		)
		if err != nil {
			logs.Errorf("sync huawei image failed, err: %v, rid: %s", err, kit.Rid)
			continue
		}
	}

	return err
}

// SyncHuaWeiZone sync huawei zone
func SyncHuaWeiZone(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	regions, err := c.DataService().HuaWei.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.HuaWeiRegionListReq{
			Filter: tools.EqualExpression("type", constant.SyncTimingListHuaWeiRegion),
			Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		},
	)
	if err != nil {
		logs.Errorf("sync list huawei region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, region := range regions.Details {
		err = c.HCService().HuaWei.Zone.SyncZone(
			kit.Ctx,
			header,
			&zone.HuaWeiZoneSyncReq{
				AccountID: accountID,
				Region:    region.RegionID,
			},
		)
		if err != nil {
			logs.Errorf("sync do huawei sync zone failed, err: %v, regionID: %s, rid: %s",
				err, region.RegionID, kit.Rid)
		}
	}

	return err
}

// SyncHuaWeiRegion sync huawei region
func SyncHuaWeiRegion(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	err := c.HCService().HuaWei.Region.SyncRegion(
		kit.Ctx,
		header,
		&protohcregion.HuaWeiRegionSyncReq{
			AccountID: accountID,
		},
	)
	if err != nil {
		logs.Errorf("sync do huawei sync region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	return nil
}
