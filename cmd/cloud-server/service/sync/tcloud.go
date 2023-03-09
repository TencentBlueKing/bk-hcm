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
	protodisk "hcm/pkg/api/hc-service/disk"
	protoeip "hcm/pkg/api/hc-service/eip"
	protoimage "hcm/pkg/api/hc-service/image"
	protohcregion "hcm/pkg/api/hc-service/region"
	"hcm/pkg/api/hc-service/zone"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncTCloudAll sync tcloud all resource
func SyncTCloudAll(c *client.ClientSet, kit *kit.Kit, header http.Header, accountID string) error {

	regions, err := c.DataService().TCloud.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.TCloudRegionListReq{
			Filter: tools.EqualExpression("vendor", enumor.TCloud),
			Page:   core.DefaultBasePage,
		},
	)
	if err != nil {
		logs.Errorf("sync list tcloud region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, region := range regions.Details {
		err = SyncTCloudSG(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do tcloud sync sg failed, err: %v, regionID: %s, rid: %s",
				err, region.RegionID, kit.Rid)
		}

		err = SyncTCloudSGRule(c, kit, header, region.RegionID, accountID)
		if err != nil {
			logs.Errorf("sync do tcloud sync sg rule failed, err: %v, regionID: %s, rid: %s",
				err, region.RegionID, kit.Rid)
		}

		err = SyncTCloudDisk(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do tcloud sync disk failed, err: %v, regionID: %s, rid: %s",
				err, region.RegionID, kit.Rid)
		}

		err = SyncTCloudEip(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do tcloud sync eip failed, err: %v, regionID: %s, rid: %s",
				err, region.RegionID, kit.Rid)
		}
	}

	if err != nil {
		return err
	}

	return nil
}

// SyncTCloudSG ...
func SyncTCloudSG(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().TCloud.SecurityGroup.SyncSecurityGroup(
		kit.Ctx,
		header,
		&proto.SecurityGroupSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do tcloud sync sg failed, err: %v, regionID: %s, rid: %s",
			err, region, kit.Rid)
		return err
	}

	return nil
}

// SyncTCloudDisk ...
func SyncTCloudDisk(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().TCloud.Disk.SyncDisk(
		kit.Ctx,
		header,
		&protodisk.DiskSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do tcloud sync disk failed, err: %v, regionID: %s, rid: %s",
			err, region, kit.Rid)
		return err
	}

	return nil
}

// SyncTCloudEip ...
func SyncTCloudEip(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().TCloud.Eip.SyncEip(
		kit.Ctx,
		header,
		&protoeip.EipSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do tcloud sync eip failed, err: %v, regionID: %s, rid: %s",
			err, region, kit.Rid)
		return err
	}

	return nil
}

// SyncTCloudSGRule ...
func SyncTCloudSGRule(c *client.ClientSet, kit *kit.Kit, header http.Header,
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
							Value: enumor.TCloud,
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
			logs.Errorf("list tcloud security group failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		if len(results.Details) == 0 {
			break
		}

		for _, v := range results.Details {
			err = c.HCService().TCloud.SecurityGroup.SyncSecurityGroupRule(
				kit.Ctx,
				header,
				&proto.SecurityGroupSyncReq{
					AccountID: v.AccountID,
					Region:    v.Region,
				},
				v.ID,
			)
			if err != nil {
				logs.Errorf("sync do tcloud sync sg  rule failed, err: %v, regionID: %s, rid: %s",
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

// SyncTCloudPublicResource ...
func SyncTCloudPublicResource(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	err := SyncTCloudRegion(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do tcloud sync region failed, err: %v, rid: %s", err, kit.Rid)
	}

	err = SyncTCloudZone(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do tcloud sync zone failed, err: %v, rid: %s", err, kit.Rid)
	}

	err = SyncTCloudImage(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do tcloud sync image failed, err: %v, rid: %s", err, kit.Rid)
	}

	return nil
}

// SyncTCloudImage ...
func SyncTCloudImage(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	regions, err := c.DataService().TCloud.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.TCloudRegionListReq{
			Filter: tools.EqualExpression("vendor", enumor.TCloud),
			Page:   core.DefaultBasePage,
		},
	)
	if err != nil {
		logs.Errorf("sync list tcloud region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, region := range regions.Details {
		err = c.HCService().TCloud.Image.SyncImage(
			kit.Ctx,
			header,
			&protoimage.TCloudImageSyncReq{
				AccountID: accountID,
				Region:    region.RegionID,
			},
		)
		if err != nil {
			logs.Errorf("sync tcloud image failed, err: %v, rid: %s", err, kit.Rid)
			continue
		}
	}

	return err
}

// SyncTCloudZone sync tcloud zone
func SyncTCloudZone(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	regions, err := c.DataService().TCloud.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.TCloudRegionListReq{
			Filter: tools.EqualExpression("vendor", enumor.TCloud),
			Page:   core.DefaultBasePage,
		},
	)
	if err != nil {
		logs.Errorf("sync list tcloud region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, region := range regions.Details {
		err = c.HCService().TCloud.Zone.SyncZone(
			kit.Ctx,
			header,
			&zone.TCloudZoneSyncReq{
				AccountID: accountID,
				Region:    region.RegionID,
			},
		)
		if err != nil {
			logs.Errorf("sync do tcloud sync zone failed, err: %v, regionID: %s, rid: %s",
				err, region.RegionID, kit.Rid)
		}
	}
	return err
}

// SyncTCloudRegion sync tcloud region
func SyncTCloudRegion(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	err := c.HCService().TCloud.Region.SyncRegion(
		kit.Ctx,
		header,
		&protohcregion.TCloudRegionSyncReq{
			AccountID: accountID,
		},
	)
	if err != nil {
		logs.Errorf("sync do tcloud sync region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}
	return nil
}
