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

package tcloud

import (
	"fmt"
	"strconv"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	typeargstpl "hcm/pkg/adaptor/types/argument-template"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// ListVpcTaskResult 查询VPC异步任务执行结果
// reference: https://cloud.tencent.com/document/api/215/59037
func (t *TCloudImpl) ListVpcTaskResult(kt *kit.Kit, opt *typeargstpl.TCloudVpcTaskResultOption) (
	*vpc.DescribeVpcTaskResultResponseParams, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list vpc task result option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeVpcTaskResultRequest()
	req.TaskId = common.StringPtr(opt.TaskID)

	resp, err := client.DescribeVpcTaskResultWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud vpc task result failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 异步任务执行结果。结果：SUCCESS、FAILED、RUNNING
	return resp.Response, nil
}

// ListArgsTplAddress list address.
// reference: https://cloud.tencent.com/document/api/215/16717
func (t *TCloudImpl) ListArgsTplAddress(kt *kit.Kit, opt *typeargstpl.TCloudListOption) (
	[]typeargstpl.TCloudArgsTplAddress, uint64, error) {

	if opt == nil {
		return nil, 0, errf.New(errf.InvalidParameter, "list address option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, 0, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeAddressTemplatesRequest()

	if len(opt.Filters) > 0 {
		req.Filters = opt.Filters
	}

	if opt.Page != nil {
		req.Offset = common.StringPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = common.StringPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	resp, err := client.DescribeAddressTemplatesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud argument template address failed, err: %v, rid: %s", err, kt.Rid)
		return nil, 0, err
	}

	list := make([]typeargstpl.TCloudArgsTplAddress, 0, len(resp.Response.AddressTemplateSet))
	for _, one := range resp.Response.AddressTemplateSet {
		list = append(list, typeargstpl.TCloudArgsTplAddress{AddressTemplate: one})
	}

	return list, converter.PtrToVal(resp.Response.TotalCount), nil
}

// CreateArgsTplAddress 创建IP地址模板
// reference: https://cloud.tencent.com/document/api/215/16708
func (t *TCloudImpl) CreateArgsTplAddress(kt *kit.Kit, opt *typeargstpl.TCloudCreateAddressOption) (
	*vpc.AddressTemplate, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create address option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewCreateAddressTemplateRequest()
	req.AddressTemplateName = common.StringPtr(opt.TemplateName)
	req.AddressesExtra = opt.AddressesExtra

	resp, err := client.CreateAddressTemplateWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud argument template address instance failed, opt: %+v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return nil, err
	}

	return resp.Response.AddressTemplate, nil
}

// DeleteArgsTplAddress 删除IP地址模板
// reference: https://cloud.tencent.com/document/api/213/15723
func (t *TCloudImpl) DeleteArgsTplAddress(kt *kit.Kit, opt *typeargstpl.TCloudDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete address option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteAddressTemplateRequest()
	req.AddressTemplateId = common.StringPtr(opt.CloudID)

	resp, err := client.DeleteAddressTemplateWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete argument template address instance failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
		poller.BaseDoneResult]{Handler: &vpcTaskResultPollingHandler{}}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.RequestId},
		types.NewBatchDeleteArgsTplPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud argument template delete address poller failed, resp: %+v, err: %v, rid: %s",
			converter.PtrToVal(resp.Response), err, kt.Rid)
		return err
	}

	if len(result.FailedCloudIDs) > 0 {
		return errf.Newf(errf.PartialFailed, "delete argument template address failed, msg: %s", result.FailedMessage)
	}

	return nil
}

// UpdateArgsTplAddress 修改IP地址模板
// reference: https://cloud.tencent.com/document/api/215/16720
func (t *TCloudImpl) UpdateArgsTplAddress(kt *kit.Kit, opt *typeargstpl.TCloudUpdateAddressOption) (
	*poller.BaseDoneResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "update address option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewModifyAddressTemplateAttributeRequest()
	req.AddressTemplateId = common.StringPtr(opt.TemplateID)

	if len(opt.TemplateName) > 0 {
		req.AddressTemplateName = common.StringPtr(opt.TemplateName)
	}

	if len(opt.AddressesExtra) > 0 {
		req.AddressesExtra = opt.AddressesExtra
	}

	resp, err := client.ModifyAddressTemplateAttributeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud argument template address failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	respPoller := poller.Poller[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
		poller.BaseDoneResult]{Handler: &vpcTaskResultPollingHandler{}}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.RequestId},
		types.NewBatchUpdateArgsTplPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud argument template update address poller failed, resp: %+v, err: %v, rid: %s",
			converter.PtrToVal(resp.Response), err, kt.Rid)
		return nil, err
	}

	return result, nil
}

// ListArgsTplAddressGroup list address group.
// reference: https://cloud.tencent.com/document/api/215/16716
func (t *TCloudImpl) ListArgsTplAddressGroup(kt *kit.Kit, opt *typeargstpl.TCloudListOption) (
	[]typeargstpl.TCloudArgsTplAddressGroup, uint64, error) {

	if opt == nil {
		return nil, 0, errf.New(errf.InvalidParameter, "list address group option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, 0, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeAddressTemplateGroupsRequest()

	if len(opt.Filters) > 0 {
		req.Filters = opt.Filters
	}

	if opt.Page != nil {
		req.Offset = common.StringPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = common.StringPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	resp, err := client.DescribeAddressTemplateGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud argument template address group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, 0, err
	}

	list := make([]typeargstpl.TCloudArgsTplAddressGroup, 0, len(resp.Response.AddressTemplateGroupSet))
	for _, one := range resp.Response.AddressTemplateGroupSet {
		list = append(list, typeargstpl.TCloudArgsTplAddressGroup{AddressTemplateGroup: one})
	}

	return list, converter.PtrToVal(resp.Response.TotalCount), nil
}

// CreateArgsTplAddressGroup 创建IP地址模板集合
// reference: https://cloud.tencent.com/document/api/215/16709
func (t *TCloudImpl) CreateArgsTplAddressGroup(kt *kit.Kit, opt *typeargstpl.TCloudCreateAddressGroupOption) (
	*vpc.AddressTemplateGroup, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create address group option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewCreateAddressTemplateGroupRequest()
	req.AddressTemplateGroupName = common.StringPtr(opt.TemplateGroupName)
	req.AddressTemplateIds = common.StringPtrs(opt.TemplateIDs)

	resp, err := client.CreateAddressTemplateGroupWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud argument template address group instance failed, opt: %+v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return nil, err
	}

	return resp.Response.AddressTemplateGroup, nil
}

// DeleteArgsTplAddressGroup 删除IP地址模板集合
// reference: https://cloud.tencent.com/document/api/213/15723
func (t *TCloudImpl) DeleteArgsTplAddressGroup(kt *kit.Kit, opt *typeargstpl.TCloudDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete address group option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteAddressTemplateGroupRequest()
	req.AddressTemplateGroupId = common.StringPtr(opt.CloudID)

	resp, err := client.DeleteAddressTemplateGroupWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete argument template address group instance failed, opt: %+v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
		poller.BaseDoneResult]{Handler: &vpcTaskResultPollingHandler{}}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.RequestId},
		types.NewBatchDeleteArgsTplPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud argument template delete address group poller failed, resp: %+v, "+
			"err: %v, rid: %s", converter.PtrToVal(resp.Response), err, kt.Rid)
		return err
	}

	if len(result.FailedCloudIDs) > 0 {
		return errf.Newf(errf.PartialFailed, "delete argument template address group failed, msg: %s",
			result.FailedMessage)
	}

	return nil
}

// UpdateArgsTplAddressGroup 修改IP地址模板集合
// reference: https://cloud.tencent.com/document/api/215/16721
func (t *TCloudImpl) UpdateArgsTplAddressGroup(kt *kit.Kit, opt *typeargstpl.TCloudUpdateAddressGroupOption) (
	*poller.BaseDoneResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "update address group option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewModifyAddressTemplateGroupAttributeRequest()
	req.AddressTemplateGroupId = common.StringPtr(opt.TemplateGroupID)

	if len(opt.TemplateGroupName) > 0 {
		req.AddressTemplateGroupName = common.StringPtr(opt.TemplateGroupName)
	}

	if len(opt.TemplateIDs) > 0 {
		req.AddressTemplateIds = common.StringPtrs(opt.TemplateIDs)
	}

	resp, err := client.ModifyAddressTemplateGroupAttributeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud argument template address group failed, opt: %+v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return nil, err
	}

	respPoller := poller.Poller[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
		poller.BaseDoneResult]{Handler: &vpcTaskResultPollingHandler{}}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.RequestId},
		types.NewBatchUpdateArgsTplPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud argument template update address group poller failed, resp: %+v, "+
			"err: %v, rid: %s", converter.PtrToVal(resp.Response), err, kt.Rid)
		return nil, err
	}

	return result, nil
}

// ListArgsTplService 查询协议端口模板
// reference: https://cloud.tencent.com/document/api/215/16719
func (t *TCloudImpl) ListArgsTplService(kt *kit.Kit, opt *typeargstpl.TCloudListOption) (
	[]typeargstpl.TCloudArgsTplService, uint64, error) {

	if opt == nil {
		return nil, 0, errf.New(errf.InvalidParameter, "list service option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, 0, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeServiceTemplatesRequest()

	if len(opt.Filters) > 0 {
		req.Filters = opt.Filters
	}

	if opt.Page != nil {
		req.Offset = common.StringPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = common.StringPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	resp, err := client.DescribeServiceTemplatesWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud argument template service port failed, err: %v, rid: %s", err, kt.Rid)
		return nil, 0, err
	}

	list := make([]typeargstpl.TCloudArgsTplService, 0, len(resp.Response.ServiceTemplateSet))
	for _, one := range resp.Response.ServiceTemplateSet {
		list = append(list, typeargstpl.TCloudArgsTplService{ServiceTemplate: one})
	}

	return list, converter.PtrToVal(resp.Response.TotalCount), nil
}

// CreateArgsTplService 创建协议端口模板
// reference: https://cloud.tencent.com/document/api/215/16710
func (t *TCloudImpl) CreateArgsTplService(kt *kit.Kit, opt *typeargstpl.TCloudCreateServiceOption) (
	*vpc.ServiceTemplate, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create service option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewCreateServiceTemplateRequest()
	req.ServiceTemplateName = common.StringPtr(opt.TemplateName)
	req.ServicesExtra = opt.ServicesExtra

	resp, err := client.CreateServiceTemplateWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud argument template service port instance failed, opt: %+v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return nil, err
	}

	return resp.Response.ServiceTemplate, nil
}

// DeleteArgsTplService 删除协议端口模板
// reference: https://cloud.tencent.com/document/api/215/16714
func (t *TCloudImpl) DeleteArgsTplService(kt *kit.Kit, opt *typeargstpl.TCloudDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete service option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteServiceTemplateRequest()
	req.ServiceTemplateId = common.StringPtr(opt.CloudID)

	resp, err := client.DeleteServiceTemplateWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete argument template service port instance failed, opt: %+v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
		poller.BaseDoneResult]{Handler: &vpcTaskResultPollingHandler{}}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.RequestId},
		types.NewBatchDeleteArgsTplPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud argument template delete service poller failed, resp: %+v, err: %v, rid: %s",
			converter.PtrToVal(resp.Response), err, kt.Rid)
		return err
	}

	if len(result.FailedCloudIDs) > 0 {
		return errf.Newf(errf.PartialFailed, "delete argument template service failed, msg: %s", result.FailedMessage)
	}

	return nil
}

// UpdateArgsTplService 修改协议端口模板
// reference: https://cloud.tencent.com/document/api/215/16722
func (t *TCloudImpl) UpdateArgsTplService(kt *kit.Kit, opt *typeargstpl.TCloudUpdateServiceOption) (
	*poller.BaseDoneResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "update service option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewModifyServiceTemplateAttributeRequest()
	req.ServiceTemplateId = common.StringPtr(opt.TemplateID)

	if len(opt.TemplateName) > 0 {
		req.ServiceTemplateName = common.StringPtr(opt.TemplateName)
	}

	if len(opt.ServicesExtra) > 0 {
		req.ServicesExtra = opt.ServicesExtra
	}

	resp, err := client.ModifyServiceTemplateAttributeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud argument template service failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	respPoller := poller.Poller[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
		poller.BaseDoneResult]{Handler: &vpcTaskResultPollingHandler{}}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.RequestId},
		types.NewBatchUpdateArgsTplPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud argument template update service poller failed, resp: %+v, err: %v, rid: %s",
			converter.PtrToVal(resp.Response), err, kt.Rid)
		return nil, err
	}

	return result, nil
}

// ListArgsTplServiceGroup 查询协议端口模板集合
// reference: https://cloud.tencent.com/document/api/215/16718
func (t *TCloudImpl) ListArgsTplServiceGroup(kt *kit.Kit, opt *typeargstpl.TCloudListOption) (
	[]typeargstpl.TCloudArgsTplServiceGroup, uint64, error) {

	if opt == nil {
		return nil, 0, errf.New(errf.InvalidParameter, "list service group option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, 0, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, 0, fmt.Errorf("new tcloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDescribeServiceTemplateGroupsRequest()

	if len(opt.Filters) > 0 {
		req.Filters = opt.Filters
	}

	if opt.Page != nil {
		req.Offset = common.StringPtr(strconv.FormatUint(opt.Page.Offset, 10))
		req.Limit = common.StringPtr(strconv.FormatUint(opt.Page.Limit, 10))
	}

	resp, err := client.DescribeServiceTemplateGroupsWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("list tcloud argument template service port group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, 0, err
	}

	list := make([]typeargstpl.TCloudArgsTplServiceGroup, 0, len(resp.Response.ServiceTemplateGroupSet))
	for _, one := range resp.Response.ServiceTemplateGroupSet {
		list = append(list, typeargstpl.TCloudArgsTplServiceGroup{ServiceTemplateGroup: one})
	}

	return list, converter.PtrToVal(resp.Response.TotalCount), nil
}

// CreateArgsTplServiceGroup 创建协议端口模板集合
// reference: https://cloud.tencent.com/document/api/215/16711
func (t *TCloudImpl) CreateArgsTplServiceGroup(kt *kit.Kit, opt *typeargstpl.TCloudCreateServiceGroupOption) (
	*vpc.ServiceTemplateGroup, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create service group option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewCreateServiceTemplateGroupRequest()
	req.ServiceTemplateGroupName = common.StringPtr(opt.TemplateGroupName)
	req.ServiceTemplateIds = common.StringPtrs(opt.TemplateIDs)

	resp, err := client.CreateServiceTemplateGroupWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud argument template service port group instance failed, opt: %+v, "+
			"err: %v, rid: %s", opt, err, kt.Rid)
		return nil, err
	}

	return resp.Response.ServiceTemplateGroup, nil
}

// DeleteArgsTplServiceGroup 删除协议端口模板集合
// reference: https://cloud.tencent.com/document/api/215/16715
func (t *TCloudImpl) DeleteArgsTplServiceGroup(kt *kit.Kit, opt *typeargstpl.TCloudDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete service group option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewDeleteServiceTemplateGroupRequest()
	req.ServiceTemplateGroupId = common.StringPtr(opt.CloudID)

	resp, err := client.DeleteServiceTemplateGroupWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("delete argument template service group instance failed, opt: %+v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
		poller.BaseDoneResult]{Handler: &vpcTaskResultPollingHandler{}}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.RequestId},
		types.NewBatchDeleteArgsTplPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud argument template delete service group poller failed, resp: %+v, "+
			"err: %v, rid: %s", converter.PtrToVal(resp.Response), err, kt.Rid)
		return err
	}

	if len(result.FailedCloudIDs) > 0 {
		return errf.Newf(errf.PartialFailed, "delete argument template service group failed, msg: %s",
			result.FailedMessage)
	}

	return nil
}

// UpdateArgsTplServiceGroup 修改协议端口模板集合
// reference: https://cloud.tencent.com/document/api/215/16723
func (t *TCloudImpl) UpdateArgsTplServiceGroup(kt *kit.Kit, opt *typeargstpl.TCloudUpdateServiceGroupOption) (
	*poller.BaseDoneResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "update service group option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.VpcClient(constant.TCloudDefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud vpc client failed, err: %v", err)
	}

	req := vpc.NewModifyServiceTemplateGroupAttributeRequest()
	req.ServiceTemplateGroupId = common.StringPtr(opt.TemplateGroupID)

	if len(opt.TemplateGroupName) > 0 {
		req.ServiceTemplateGroupName = common.StringPtr(opt.TemplateGroupName)
	}

	if len(opt.TemplateIDs) > 0 {
		req.ServiceTemplateIds = common.StringPtrs(opt.TemplateIDs)
	}

	resp, err := client.ModifyServiceTemplateGroupAttributeWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("run tencent cloud argument template service group failed, opt: %+v, err: %v, rid: %s",
			opt, err, kt.Rid)
		return nil, err
	}

	respPoller := poller.Poller[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
		poller.BaseDoneResult]{Handler: &vpcTaskResultPollingHandler{}}
	result, err := respPoller.PollUntilDone(t, kt, []*string{resp.Response.RequestId},
		types.NewBatchUpdateArgsTplPollerOption())
	if err != nil {
		logs.Errorf("run tencent cloud argument template update service group poller failed, resp: %+v, "+
			"err: %v, rid: %s", converter.PtrToVal(resp.Response), err, kt.Rid)
		return nil, err
	}

	return result, nil
}

type vpcTaskResultPollingHandler struct {
	region string
}

// Done 异步任务执行结果。结果：SUCCESS、FAILED、RUNNING
func (h *vpcTaskResultPollingHandler) Done(vpcTaskResults []*vpc.DescribeVpcTaskResultResponseParams) (
	bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
		UnknownCloudIDs: make([]string, 0),
	}
	flag := true
	for _, instance := range vpcTaskResults {
		if instance == nil {
			return flag, result
		}

		switch converter.PtrToVal(instance.Status) {
		case "RUNNING":
			result.UnknownCloudIDs = append(result.UnknownCloudIDs, *instance.RequestId)
			flag = false
		case "FAILED":
			result.FailedCloudIDs = append(result.FailedCloudIDs, *instance.RequestId)
			result.FailedMessage = converter.PtrToVal(instance.Output)
		case "SUCCESS":
			result.SuccessCloudIDs = append(result.SuccessCloudIDs, *instance.RequestId)
		default:
			break
		}
	}

	return flag, result
}

func (h *vpcTaskResultPollingHandler) Poll(client *TCloudImpl, kt *kit.Kit, cloudIDs []*string) (
	[]*vpc.DescribeVpcTaskResultResponseParams, error) {

	cloudIDSplit := slice.Split(cloudIDs, core.TCloudQueryLimit)
	results := make([]*vpc.DescribeVpcTaskResultResponseParams, 0, len(cloudIDs))
	for _, partIDs := range cloudIDSplit {
		for _, tmpCloud := range partIDs {
			opt := &typeargstpl.TCloudVpcTaskResultOption{
				TaskID: converter.PtrToVal(tmpCloud),
			}
			resp, err := client.ListVpcTaskResult(kt, opt)
			if err != nil {
				return nil, err
			}

			results = append(results, resp)
		}
	}
	if len(results) != len(cloudIDs) {
		return nil, fmt.Errorf("query tasks count: %d not equal return count: %d", len(cloudIDs), len(results))
	}

	return results, nil
}

var _ poller.PollingHandler[*TCloudImpl, []*vpc.DescribeVpcTaskResultResponseParams,
	poller.BaseDoneResult] = new(vpcTaskResultPollingHandler)
