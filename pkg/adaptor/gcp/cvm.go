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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"google.golang.org/api/compute/v1"
)

// ListCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
func (g *Gcp) ListCvm(kt *kit.Kit, opt *typecvm.GcpListOption) ([]typecvm.GcpCvm, string, error) {
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

	if len(opt.Names) > 0 {
		request.Filter(generateResourceFilter("name", opt.Names))
	}

	if opt.Page != nil {
		request.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	}

	resp, err := request.Do()
	if err != nil {
		logs.Errorf("list instance failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return nil, "", err
	}

	cvms := make([]typecvm.GcpCvm, 0, len(resp.Items))
	for _, one := range resp.Items {
		cvms = append(cvms, typecvm.GcpCvm{one})
	}

	return cvms, resp.NextPageToken, nil
}

// CountCvmAndNI reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
func (g *Gcp) CountCvmAndNI(kt *kit.Kit) (int32, int32, error) {

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return 0, 0, err
	}

	request := client.Instances.AggregatedList(g.CloudProjectID()).Context(kt.Ctx)

	var cvmCount int32
	var niCount int32
	for {
		resp, err := request.Do()
		if err != nil {
			logs.Errorf("list instance failed, err: %v, rid: %s", err, kt.Rid)
			return 0, 0, err
		}

		for _, one := range resp.Items {
			cvmCount += int32(len(one.Instances))

			for _, instance := range one.Instances {
				niCount += int32(len(instance.NetworkInterfaces))
			}
		}

		if resp.NextPageToken == "" {
			break
		}
	}

	return cvmCount, niCount, nil
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

	handler := &stopCvmPollingHandler{
		opt.Zone,
	}
	respPoller := poller.Poller[*Gcp, []*compute.Instance, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(g, kt, []*string{to.Ptr(opt.Name)},
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
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

	handler := &startCvmPollingHandler{
		opt.Zone,
	}
	respPoller := poller.Poller[*Gcp, []*compute.Instance, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(g, kt, []*string{to.Ptr(opt.Name)},
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
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

	handler := &resetCvmPollingHandler{
		opt.Zone,
	}
	respPoller := poller.Poller[*Gcp, []*compute.Instance, poller.BaseDoneResult]{Handler: handler}
	_, err = respPoller.PollUntilDone(g, kt, []*string{to.Ptr(opt.Name)},
		types.NewBatchOperateCvmPollerOpt())
	if err != nil {
		return err
	}

	return nil
}

// CreateCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/bulkInsert
func (g *Gcp) CreateCvm(kt *kit.Kit, opt *typecvm.GcpCreateOption) (*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "reset option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	script, err := opt.ImageProjectType.StartupScript(opt.Password)
	if err != nil {
		return nil, err
	}

	req := buildCreateCvmReq(opt, script)

	resp := new(compute.Operation)
	if len(opt.RequestID) != 0 {
		resp, err = client.Instances.BulkInsert(g.CloudProjectID(), opt.Zone, req).
			RequestId(opt.RequestID).Context(kt.Ctx).Do()
	} else {
		resp, err = client.Instances.BulkInsert(g.CloudProjectID(), opt.Zone, req).Context(kt.Ctx).Do()
	}
	if err != nil {
		logs.Errorf("create instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	handler := &createCvmPollingHandler{
		opt.Zone,
	}
	respPoller := poller.Poller[*Gcp, []*compute.Operation, poller.BaseDoneResult]{Handler: handler}
	result, err := respPoller.PollUntilDone(g, kt, []*string{to.Ptr(resp.OperationGroupId)},
		types.NewBatchCreateCvmPollerOption())
	if err != nil {
		return nil, err
	}

	g.deleteCvmMetadataStartScript(kt, client, opt.Zone, result.SuccessCloudIDs)

	return result, nil
}

func buildCreateCvmReq(opt *typecvm.GcpCreateOption, script string) *compute.BulkInsertInstanceResource {
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
						DiskType:    string(opt.SystemDisk.DiskType),
						SourceImage: opt.CloudImageSelfLink,
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
		},
		MinCount:    opt.RequiredCount,
		NamePattern: opt.NamePrefix + "-####",
	}

	if opt.PublicIPAssigned {
		req.InstanceProperties.NetworkInterfaces = []*compute.NetworkInterface{
			{
				Network:    opt.CloudVpcSelfLink,
				Subnetwork: opt.CloudSubnetSelfLink,
				AccessConfigs: []*compute.AccessConfig{
					{
						Name: "External NAT",
						Type: "ONE_TO_ONE_NAT",
					},
				},
			},
		}
	} else {
		req.InstanceProperties.NetworkInterfaces = []*compute.NetworkInterface{
			{
				Network:    opt.CloudVpcSelfLink,
				Subnetwork: opt.CloudSubnetSelfLink,
			},
		}
	}

	if len(opt.DataDisk) != 0 {
		for index, disk := range opt.DataDisk {
			req.InstanceProperties.Disks = append(req.InstanceProperties.Disks, &compute.AttachedDisk{
				Index: int64(index) + 1,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskSizeGb: disk.SizeGb,
					DiskType:   string(disk.DiskType),
				},
				Mode:       string(disk.Mode),
				AutoDelete: disk.AutoDelete,
			})
		}
	}
	return req
}

func (g *Gcp) deleteCvmMetadataStartScript(kt *kit.Kit, client *compute.Service, zone string, ids []string) {

	resp, err := client.Instances.List(g.CloudProjectID(), zone).Context(kt.Ctx).
		Filter(generateResourceIDsFilter(ids)).Do()
	if err != nil {
		logs.Errorf("%s: delete cvm metadata start script to list cvm failed, err: %v, ids: %v, rid: %s",
			constant.DeleteCvmStartScriptFailed, err, ids, kt.Rid)
		return
	}

	for _, one := range resp.Items {
		one.Metadata = &compute.Metadata{
			Items: []*compute.MetadataItems{},
		}

		_, err := client.Instances.Update(g.CloudProjectID(), zone, one.Name, one).Do()
		if err != nil {
			logs.Errorf("%s: delete cvm metadata start script to update cvm failed, err: %v, name: %s, rid: %s",
				constant.DeleteCvmStartScriptFailed, err, one.Name, kt.Rid)
			continue
		}
	}

	return
}

type startCvmPollingHandler struct {
	zone string
}

// Done ...
func (h *startCvmPollingHandler) Done(instances []*compute.Instance) (bool, *poller.BaseDoneResult) {
	return done(instances, []string{"RUNNING"})
}

// Poll ...
func (h *startCvmPollingHandler) Poll(client *Gcp, kt *kit.Kit, names []*string) ([]*compute.Instance, error) {
	return poll(client, kt, h.zone, names)
}

type stopCvmPollingHandler struct {
	zone string
}

// Done ...
func (h *stopCvmPollingHandler) Done(instances []*compute.Instance) (bool, *poller.BaseDoneResult) {
	return done(instances, []string{"STOPPED", "TERMINATED"})
}

// Poll ...
func (h *stopCvmPollingHandler) Poll(client *Gcp, kt *kit.Kit, names []*string) ([]*compute.Instance, error) {
	return poll(client, kt, h.zone, names)
}

type resetCvmPollingHandler struct {
	zone string
}

// Done ...
func (h *resetCvmPollingHandler) Done(instances []*compute.Instance) (bool, *poller.BaseDoneResult) {
	return done(instances, []string{"RUNNING"})
}

// Poll ...
func (h *resetCvmPollingHandler) Poll(client *Gcp, kt *kit.Kit, names []*string) ([]*compute.Instance, error) {
	return poll(client, kt, h.zone, names)
}

func done(instances []*compute.Instance, succeed []string) (bool, *poller.BaseDoneResult) {
	result := new(poller.BaseDoneResult)

	succeeMap := converter.StringSliceToMapBool(succeed)

	flag := true
	for _, instance := range instances {
		// not done
		if !succeeMap[instance.Status] {
			flag = false
			continue
		}

		result.SuccessCloudIDs = append(result.SuccessCloudIDs, strconv.FormatUint(instance.Id, 10))
	}

	return flag, result
}

func poll(client *Gcp, kt *kit.Kit, zone string, names []*string) ([]*compute.Instance, error) {
	cli, err := client.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	request := cli.Instances.List(client.CloudProjectID(), zone).Context(kt.Ctx)
	listNames := converter.PtrToSlice(names)
	request.Filter(generateResourceFilter("name", listNames))

	resp, err := request.Do()
	if err != nil {
		logs.Errorf("list instance failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.Items, nil
}

type createCvmPollingHandler struct {
	zone string
}

// Done ...
func (h *createCvmPollingHandler) Done(items []*compute.Operation) (bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
	}
	flag := true
	for _, item := range items {
		if item.OperationType == "insert" && item.Status == "DONE" {
			result.SuccessCloudIDs = append(result.SuccessCloudIDs, strconv.FormatUint(item.TargetId, 10))
			continue
		}

		if item.OperationType == "insert" && item.Status == "PENDING" {
			flag = false
			result.UnknownCloudIDs = append(result.UnknownCloudIDs, strconv.FormatUint(item.TargetId, 10))
			continue
		}

		if item.OperationType == "insert" && item.Status == "RUNNING" {
			flag = false
			result.UnknownCloudIDs = append(result.UnknownCloudIDs, strconv.FormatUint(item.TargetId, 10))
			continue
		}
	}

	return flag, result
}

func (h *createCvmPollingHandler) Poll(client *Gcp, kt *kit.Kit, operGroupIDs []*string) ([]*compute.Operation, error) {

	if len(operGroupIDs) == 0 {
		return nil, errors.New("operation group id is required")
	}

	computeClient, err := client.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	operResp, err := computeClient.ZoneOperations.List(client.CloudProjectID(), h.zone).Context(kt.Ctx).
		Filter(fmt.Sprintf(`operationGroupId="%s"`, *operGroupIDs[0])).Do()
	if err != nil {
		logs.Errorf("list zone operations failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(operResp.Items) <= 1 {
		return nil, errors.New("operation has not been created yet, need to wait")
	}

	return operResp.Items, nil
}

var _ poller.PollingHandler[*Gcp, []*compute.Operation, poller.BaseDoneResult] = new(createCvmPollingHandler)
