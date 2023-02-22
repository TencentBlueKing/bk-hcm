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

package gcp

import (
	"strings"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"google.golang.org/api/compute/v1"
)

// ListCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
func (g *Gcp) ListCvm(kt *kit.Kit, opt *typecvm.GcpListOption) ([]*compute.Instance, string, error) {
	if opt == nil {
		return nil, "", errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, "", err
	}

	request := client.Instances.List(g.CloudProjectID(), opt.Zone).Context(kt.Ctx)

	if len(opt.CloudIDs) > 0 {
		request.Filter(generateResourceIDsFilter(opt.CloudIDs))
	}

	if opt.Page != nil {
		request.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := request.Do()
	if err != nil {
		logs.Errorf("list instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, "", err
	}

	return resp.Items, resp.NextPageToken, nil
}

// GetMachineType gcp设备类型为url，需要截取最后一个单词
// e.g: https://www.googleapis.com/compute/v1/projects/xxx/zones/us-central1-a/machineTypes/e2-medium
func GetMachineType(typ string) string {
	return typ[strings.LastIndex(typ, "/")+1:]
}

// GetGcpIPAddresses ...
func GetGcpIPAddresses(networkInterfaces []*compute.NetworkInterface) ([]string, []string, []string, []string) {
	privateIPv4Map := make(map[string]struct{}, 0)
	privateIPv6Map := make(map[string]struct{}, 0)
	publicIPv4Map := make(map[string]struct{}, 0)
	publicIPv6Map := make(map[string]struct{}, 0)

	for _, one := range networkInterfaces {
		if one.StackType == "IPV4_ONLY" {
			privateIPv4Map[one.NetworkIP] = struct{}{}

			for _, config := range one.AccessConfigs {
				publicIPv4Map[config.NatIP] = struct{}{}
			}
		}

		if one.StackType == "IPV6_ONLY" {
			privateIPv6Map[one.NetworkIP] = struct{}{}

			for _, config := range one.AccessConfigs {
				publicIPv6Map[config.NatIP] = struct{}{}
			}
		}
	}

	return converter.MapKeyToStringSlice(privateIPv4Map), converter.MapKeyToStringSlice(publicIPv4Map),
		converter.MapKeyToStringSlice(privateIPv6Map), converter.MapKeyToStringSlice(publicIPv6Map)
}

// DeleteCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/delete
func (g *Gcp) DeleteCvm(kt *kit.Kit, opt *typecvm.GcpDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.Delete(g.CloudProjectID(), opt.Zone, opt.Name).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("delete instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// StopCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/stop
func (g *Gcp) StopCvm(kt *kit.Kit, opt *typecvm.GcpStopOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "stop option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.Stop(g.CloudProjectID(), opt.Zone, opt.Name).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("stop instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// StartCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/start
func (g *Gcp) StartCvm(kt *kit.Kit, opt *typecvm.GcpStartOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "start option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.Start(g.CloudProjectID(), opt.Zone, opt.Name).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("start instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// ResetCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/reset
func (g *Gcp) ResetCvm(kt *kit.Kit, opt *typecvm.GcpResetOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "reset option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.Start(g.CloudProjectID(), opt.Zone, opt.Name).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("reset instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// CreateCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/reset
func (g *Gcp) CreateCvm(kt *kit.Kit, opt *typecvm.GcpCreateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "reset option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	script, err := opt.ImageProjectType.StartupScript(opt.Password)
	if err != nil {
		return err
	}

	req := &compute.BulkInsertInstanceResource{
		Count: opt.RequiredCount,
		InstanceProperties: &compute.InstanceProperties{
			Description: opt.Description,
			Disks: []*compute.AttachedDisk{
				{
					AutoDelete: true,
					Boot:       true,
					InitializeParams: &compute.AttachedDiskInitializeParams{
						DiskSizeGb:  opt.SystemDisk.SizeGb,
						DiskType:    opt.SystemDisk.DiskType,
						SourceImage: opt.CloudImageID,
					},
				},
			},
			MachineType: opt.InstanceType,
			Metadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					{
						Key:   "startup-script",
						Value: converter.ValToPtr(script),
					},
				},
			},
			NetworkInterfaces: []*compute.NetworkInterface{
				{
					Network:    opt.CloudVpcID,
					Subnetwork: opt.CloudSubnetID,
				},
			},
		},
		MinCount:    opt.RequiredCount,
		NamePattern: opt.Name,
	}

	if len(opt.DataDisk) != 0 {
		for index, disk := range opt.DataDisk {
			req.InstanceProperties.Disks = append(req.InstanceProperties.Disks, &compute.AttachedDisk{
				Index: int64(index) + 1,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskSizeGb: disk.SizeGb,
					DiskType:   disk.DiskType,
				},
			})
		}
	}

	_, err = client.Instances.BulkInsert(g.CloudProjectID(), opt.Zone, req).
		RequestId(opt.RequestID).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("reset instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	// TODO: 用什么标识这批资源？

	return nil
}
