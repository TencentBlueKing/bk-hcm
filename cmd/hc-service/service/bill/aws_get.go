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

package bill

import (
	"time"

	"hcm/pkg/adaptor/aws"
	typesBill "hcm/pkg/adaptor/types/bill"
	adcore "hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/core/cloud"
	protobill "hcm/pkg/api/data-service/cloud/bill"
	hcbill "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// AwsGetBillList get aws bill list.
func (b bill) AwsGetBillList(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbill.AwsBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := b.ad.Aws(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("aws bill get cloud client failed, req: %+v, err: %+v", req, err)
		return nil, err
	}

	// 查询aws账单基础表
	billInfo, err := b.GetBillInfo(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("aws bill config get base info db failed, accountID: %s, err: %+v", req.AccountID, err)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "account_id: %s is not found", req.AccountID)
	}

	if billInfo.Status != constant.StatusSuccess {
		return nil, errf.Newf(errf.Aborted, "account_id: %s has not ready yet", req.AccountID)
	}

	opt := &typesBill.AwsBillListOption{
		AccountID: req.AccountID,
		BeginDate: req.BeginDate,
		EndDate:   req.EndDate,
	}
	if req.Page != nil {
		opt.Page = &typesBill.AwsBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		}
	}
	total, list, err := cli.GetBillList(cts.Kit, opt, billInfo)
	if err != nil {
		logs.Errorf("request adaptor list aws bill failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return &hcbill.AwsBillListResult{
		Count:   total,
		Details: list,
	}, nil
}

// GetBillInfo get bill info.
func (b bill) GetBillInfo(kt *kit.Kit, accountID string) (
	*cloud.AccountBillConfig[cloud.AwsBillConfigExtension], error) {

	// 查询aws账单基础表
	billList, err := b.cs.DataService().Aws.Bill.List(kt.Ctx, kt.Header(), &core.ListReq{
		Filter: tools.EqualExpression("account_id", accountID),
		Page:   &core.BasePage{Count: false, Start: 0, Limit: 1},
	})
	if err != nil {
		logs.Errorf("aws get base info from db failed, accountID: %s, err: %+v", accountID, err)
		return nil, err
	}
	if len(billList.Details) == 0 {
		return nil, nil
	}

	return &billList.Details[0], nil
}

// CheckBillInfo check bill info.
func (b bill) CheckBillInfo(kt *kit.Kit, req *hcbill.BillPipelineReq) (
	*cloud.AccountBillConfig[cloud.AwsBillConfigExtension], bool, error) {

	billInfo, err := b.GetBillInfo(kt, req.AccountID)
	if err != nil {
		logs.Errorf("aws bill pipeline get base info db failed, accountID: %s, err: %+v", req.AccountID, err)
		return nil, false, err
	}

	if billInfo == nil {
		billInfo = &cloud.AccountBillConfig[cloud.AwsBillConfigExtension]{
			BaseAccountBillConfig: cloud.BaseAccountBillConfig{
				Vendor:    enumor.Aws,
				AccountID: req.AccountID,
				// 状态(0:默认1:创建存储桶2:设置存储桶权限3:创建成本报告4:检查yml文件5:创建CloudFormation模版100:正常)
				Status: constant.StatusDefault,
			},
			Extension: &cloud.AwsBillConfigExtension{Region: aws.BucketRegion},
		}

		billReq := &protobill.AccountBillConfigBatchCreateReq[cloud.AwsBillConfigExtension]{
			Bills: []protobill.AccountBillConfigReq[cloud.AwsBillConfigExtension]{
				{
					Vendor:    billInfo.Vendor,
					AccountID: billInfo.AccountID,
					Status:    billInfo.Status,
				},
			},
		}
		billResp, err := b.cs.DataService().Aws.Bill.BatchCreate(kt.Ctx, kt.Header(), billReq)
		if err != nil {
			logs.Errorf("aws bill pipeline bucket db create of bill failed, req: %+v, err: %+v", req, err)
			return nil, false, err
		}
		billInfo.ID = billResp.IDs[0]
	} else {
		if billInfo.Status == constant.StatusSuccess {
			return billInfo, true, nil
		}

		resTime, err := time.Parse(constant.TimeStdFormat, billInfo.CreatedAt)
		if err != nil {
			return nil, false, err
		}

		nowTime := time.Now()
		hourDur := nowTime.Sub(resTime).Hours()
		if billInfo.Status != constant.StatusSuccess && hourDur > aws.BucketTimeOut {
			logs.Errorf("aws bill pipeline is timeout, accountID: %s, CreatedAt: %s, now: %v, hourDur: %f, "+
				"rid: %s", req.AccountID, billInfo.CreatedAt, nowTime.Local(), hourDur, kt.Rid)
			return billInfo, false, errf.New(errf.PartialFailed, "aws bill config pipeline has timeout")
		}
		if billInfo.Extension != nil && billInfo.Extension.Region == "" {
			billInfo.Extension.Region = aws.BucketRegion
		}
	}

	return billInfo, false, nil
}

// AwsGetRootAccountBillList get aws bill record list
func (b bill) AwsGetRootAccountBillList(cts *rest.Contexts) (any, error) {

	req := new(hcbill.AwsRootBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Page == nil {
		req.Page = &hcbill.AwsBillListPage{Offset: 0, Limit: adcore.AwsQueryLimit}
	}

	// 查询aws账单基础表
	billInfo, err := getRootAccountBillConfigInfo[billcore.AwsBillConfigExtension](
		cts.Kit, req.RootAccountID, b.cs.DataService())
	if err != nil {
		logs.Errorf("aws root account bill config get base info db failed, main_account_cloud_id: %s, err: %+v,rid: %s",
			req.MainAccountCloudID, err, cts.Kit.Rid)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "bill config for root_account_id: %s is not found",
			req.RootAccountID)
	}

	cli, err := b.ad.AwsRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("aws request adaptor client err, req: %+v, err: %+v,rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typesBill.AwsMainBillListOption{
		CloudAccountID: req.MainAccountCloudID,
		BeginDate:      req.BeginDate,
		EndDate:        req.EndDate,
		Page: &typesBill.AwsBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		},
	}
	count, resp, err := cli.GetMainAccountBillList(cts.Kit, opt, billInfo)
	if err != nil {
		logs.Errorf("fail to list main account bill for aws, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &hcbill.AwsBillListResult{
		Count:   count,
		Details: resp,
	}, nil
}

// AwsListRootOutsideMonthBill 查询存储在当前账单月份，但是使用日期不在当前账单月份的账单条目
func (b bill) AwsListRootOutsideMonthBill(cts *rest.Contexts) (any, error) {

	req := new(hcbill.AwsRootOutsideMonthBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if req.Page == nil {
		req.Page = &hcbill.AwsBillListPage{Offset: 0, Limit: adcore.AwsQueryLimit}
	}

	rootAccount, err := b.cs.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("fait to find root account for outside month bill, err: %+v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 查询aws账单基础表
	billInfo, err := getRootAccountBillConfigInfo[billcore.AwsBillConfigExtension](cts.Kit, req.RootAccountID,
		b.cs.DataService())
	if err != nil {
		logs.Errorf("failed to get aws root account bill config, root account: %s, err: %+v rid: %s",
			req.RootAccountID, err, cts.Kit.Rid)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "bill config for root_account_id: %s is not found",
			req.RootAccountID)
	}

	cli, err := b.ad.AwsRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("aws request adaptor client err, req: %+v, err: %+v,rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typesBill.AwsMainOutsideMonthBillLitOpt{
		PayerAccountID:  rootAccount.CloudID,
		UsageAccountIDs: req.MainAccountCloudIds,
		Year:            req.BillYear,
		Month:           req.BillMonth,
		Page:            &typesBill.AwsBillPage{Offset: req.Page.Offset, Limit: req.Page.Limit},
	}
	resp, err := cli.AwsListRootOutsideMonthBill(cts.Kit, opt, billInfo)
	if err != nil {
		logs.Errorf("fail to list main account outside month bill for aws, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return &hcbill.AwsBillListResult{
		Details: resp,
	}, nil
}

// AwsGetRootAccountSpTotalUsage ...
func (b bill) AwsGetRootAccountSpTotalUsage(cts *rest.Contexts) (any, error) {

	req := new(hcbill.AwsRootSpUsageTotalReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rootAccount, err := b.cs.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("fait to find root account, err: %+v,rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 查询aws账单基础表
	billInfo, err := getRootAccountBillConfigInfo[billcore.AwsBillConfigExtension](
		cts.Kit, req.RootAccountID, b.cs.DataService())
	if err != nil {
		logs.Errorf("aws get root account(id: %s) bill config for aws sp usage total failed, err: %+v, rid: %s",
			req.RootAccountID, err, cts.Kit.Rid)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "bill config for root account: %s is not found",
			req.RootAccountID)
	}

	cli, err := b.ad.AwsRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("aws request adaptor client err, req: %+v, err: %+v,rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}
	opt := &typesBill.AwsRootSpUsageOption{
		PayerCloudID:  rootAccount.CloudID,
		UsageCloudIDs: req.SpUsageAccountCloudIds,
		SpArnPrefix:   req.SpArnPrefix,
		Year:          req.Year,
		Month:         req.Month,
		StartDay:      req.StartDay,
		EndDay:        req.EndDay,
	}
	usage, err := cli.GetRootSpTotalUsage(cts.Kit, billInfo, opt)
	if err != nil {
		logs.Errorf("fail to get root account sp total usage, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	result := hcbill.AwsSpUsageTotalResult{
		UnblendedCost: usage.UnblendedCost,
		SPCost:        usage.SpCost,
		SPNetCost:     usage.SpNetCost,
		AccountCount:  usage.AccountCount,
	}
	return result, nil
}

// AwsListRootBillItems 查询当前账单月份指定字段的列表
func (b bill) AwsListRootBillItems(cts *rest.Contexts) (any, error) {
	req := new(hcbill.AwsRootBillItemsListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if req.Page == nil {
		req.Page = &hcbill.AwsBillListPage{Offset: 0, Limit: adcore.AwsQueryLimit}
	}

	rootAccount, err := b.cs.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("fail to find root account for deduct bill, err: %+v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	// 查询aws账单基础表
	billInfo, err := getRootAccountBillConfigInfo[billcore.AwsBillConfigExtension](cts.Kit, req.RootAccountID,
		b.cs.DataService())
	if err != nil {
		logs.Errorf("failed to get aws root account bill items config, root account: %s, err: %+v, rid: %s",
			req.RootAccountID, err, cts.Kit.Rid)
		return nil, err
	}
	if billInfo == nil {
		logs.Errorf("bill items config for root_account_id: %s is not found, rid: %s", req.RootAccountID, cts.Kit.Rid)
		return nil, errf.Newf(errf.RecordNotFound, "bill items config for root_account_id: %s is not found",
			req.RootAccountID)
	}

	cli, err := b.ad.AwsRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("aws request items adaptor client err, req: %+v, err: %+v,rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typesBill.AwsRootDeductBillListOpt{
		PayerAccountID: rootAccount.CloudID,
		Year:           req.Year,
		Month:          req.Month,
		BeginDate:      req.BeginDate,
		EndDate:        req.EndDate,
		FieldsMap:      req.FieldsMap,
		Page:           &typesBill.AwsBillPage{Offset: req.Page.Offset, Limit: req.Page.Limit},
	}
	resp, err := cli.AwsRootBillListByQueryFields(cts.Kit, opt, billInfo)
	if err != nil {
		logs.Errorf("fail to list root account bill items for aws, err: %v, opt: %+v, rid: %s",
			err, cvt.PtrToVal(opt), cts.Kit.Rid)
		return nil, err
	}

	return &hcbill.AwsBillListResult{
		Details: resp,
	}, nil
}
