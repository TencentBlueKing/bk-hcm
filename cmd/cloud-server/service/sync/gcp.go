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
	protoregion "hcm/pkg/api/data-service/cloud/region"
	protozone "hcm/pkg/api/data-service/cloud/zone"
	proto "hcm/pkg/api/hc-service"
	protocvm "hcm/pkg/api/hc-service/cvm"
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

// SyncGcpAll sync gcp all resource
func SyncGcpAll(c *client.ClientSet, kit *kit.Kit, header http.Header, accountID string) error {

	regions, err := c.DataService().Gcp.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.GcpRegionListReq{
			Filter: tools.EqualExpression("vendor", enumor.Gcp),
			Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		},
	)
	if err != nil {
		logs.Errorf("sync list gcp region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, region := range regions.Details {
		err = SyncGcpSG(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do gcp sync sg failed, err: %v, accountID: %s, regionID: %s, rid: %s",
				err, accountID, region.RegionID, kit.Rid)
		}

		err = SyncGcpEip(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do gcp sync eip failed, err: %v, accountID: %s, regionID: %s, rid: %s",
				err, accountID, region.RegionID, kit.Rid)
		}

		err = SyncGcpVpc(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do gcp sync vpc failed, err: %v, accountID: %s, regionID: %s, rid: %s",
				err, accountID, region.RegionID, kit.Rid)
		}

		err = SyncGcpSubnet(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do gcp sync subnet failed, err: %v, accountID: %s, regionID: %s, rid: %s",
				err, accountID, region.RegionID, kit.Rid)
		}

		err = SyncGcpCvm(c, kit, header, accountID, region.RegionID)
		if err != nil {
			logs.Errorf("sync do gcp sync cvm failed, err: %v, accountID: %s, regionID: %s, rid: %s",
				err, accountID, region.RegionID, kit.Rid)
		}
	}

	err = SyncGcpDisk(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do gcp sync disk failed, err: %v, rid: %s", err, kit.Rid)
	}

	return err
}

// SyncGcpDisk sync gcp disk
func SyncGcpDisk(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	zones, err := c.DataService().Global.Zone.ListZone(
		kit.Ctx,
		header,
		&protozone.ZoneListReq{
			Filter: tools.EqualExpression("vendor", enumor.Gcp),
			Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		},
	)
	if err != nil {
		logs.Errorf("sync list gcp zone failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, zone := range zones.Details {
		// disk
		err = c.HCService().Gcp.Disk.SyncDisk(
			kit.Ctx,
			header,
			&protodisk.DiskSyncReq{
				AccountID: accountID,
				Zone:      zone.Name,
			},
		)
		if err != nil {
			logs.Errorf("sync do gcp sync disk failed, err: %v, zone: %s, rid: %s",
				err, zone.Name, kit.Rid)
		}
	}

	return nil
}

// SyncGcpCvm ...
func SyncGcpCvm(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	zones, err := c.DataService().Global.Zone.ListZone(
		kit.Ctx,
		kit.Header(),
		&protozone.ZoneListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "vendor",
						Op:    filter.Equal.Factory(),
						Value: enumor.Gcp,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: region,
					},
				},
			},
			Page: &core.BasePage{
				Start: uint32(0),
				Limit: core.DefaultMaxPageLimit,
			},
		},
	)

	if err != nil {
		logs.Errorf("sync list gcp zone failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, zone := range zones.Details {
		err := c.HCService().Gcp.Cvm.SyncCvm(
			kit.Ctx,
			header,
			&protocvm.CvmSyncReq{
				AccountID: accountID,
				Region:    region,
				Zone:      zone.Name,
			},
		)
		if err != nil {
			logs.Errorf("sync do gcp sync cvm failed, err: %v, regionID: %s, rid: %s",
				err, region, kit.Rid)
		}
	}

	return nil
}

// SyncGcpSubnet ...
func SyncGcpSubnet(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().Gcp.Subnet.SyncSubnet(
		kit.Ctx,
		header,
		&proto.ResourceSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do gcp sync subnet failed, err: %v, accountID: %s, regionID: %s, rid: %s",
			err, accountID, region, kit.Rid)
		return err
	}

	return nil
}

// SyncGcpVpc ...
func SyncGcpVpc(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().Gcp.Vpc.SyncVpc(
		kit.Ctx,
		header,
		&proto.ResourceSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do gcp sync vpc failed, err: %v, accountID: %s, regionID: %s, rid: %s",
			err, accountID, region, kit.Rid)
		return err
	}

	return nil
}

// SyncGcpSG ...
func SyncGcpSG(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().Gcp.Firewall.SyncFirewall(
		kit.Ctx,
		header,
		&proto.SecurityGroupSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do gcp sync sg failed, err: %v, regionID: %s, rid: %s",
			err, region, kit.Rid)
		return err
	}

	return nil
}

// SyncGcpEip ...
func SyncGcpEip(c *client.ClientSet, kit *kit.Kit, header http.Header,
	accountID string, region string) error {

	err := c.HCService().Gcp.Eip.SyncEip(
		kit.Ctx,
		header,
		&protoeip.EipSyncReq{
			AccountID: accountID,
			Region:    region,
		},
	)
	if err != nil {
		logs.Errorf("sync do gcp sync eip failed, err: %v, regionID: %s, rid: %s",
			err, region, kit.Rid)
		return err
	}

	return nil
}

// SyncGcpPublicResource ...
func SyncGcpPublicResource(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	err := SyncGcpRegion(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do gcp sync region failed, err: %v, rid: %s", err, kit.Rid)
	}
	err = SyncGcpZone(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do gcp sync zone failed, err: %v, rid: %s", err, kit.Rid)
	}
	err = SyncGcpImage(kit, c, header, accountID)
	if err != nil {
		logs.Errorf("sync do gcp sync image failed, err: %v, rid: %s", err, kit.Rid)
	}
	return err
}

// SyncGcpImage ...
func SyncGcpImage(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	regions, err := c.DataService().Gcp.Region.ListRegion(
		kit.Ctx,
		header,
		&protoregion.GcpRegionListReq{
			Filter: tools.EqualExpression("vendor", enumor.Gcp),
			Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
		},
	)
	if err != nil {
		logs.Errorf("sync list gcp region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	for _, region := range regions.Details {
		err = c.HCService().Gcp.Image.SyncImage(
			kit.Ctx,
			header,
			&protoimage.GcpImageSyncReq{
				AccountID: accountID,
				Region:    region.RegionID,
			},
		)
		if err != nil {
			logs.Errorf("sync gcp image failed, err: %v, rid: %s", err, kit.Rid)
			continue
		}
	}
	return err
}

// SyncGcpZone sync gcp zone
func SyncGcpZone(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	err := c.HCService().Gcp.Zone.SyncZone(
		kit.Ctx,
		header,
		&zone.GcpZoneSyncReq{
			AccountID: accountID,
		},
	)
	if err != nil {
		logs.Errorf("sync do gcp sync zone failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}
	return nil
}

// SyncGcpRegion sync gcp region
func SyncGcpRegion(kit *kit.Kit, c *client.ClientSet, header http.Header, accountID string) error {
	err := c.HCService().Gcp.Region.SyncRegion(
		kit.Ctx,
		header,
		&protohcregion.GcpRegionSyncReq{
			AccountID: accountID,
		},
	)
	if err != nil {
		logs.Errorf("sync do gcp sync region failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}
	return nil
}
