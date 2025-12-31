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

package cvm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	model "hcm/cmd/woa-server/model/cvm"
	cfgtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/utils"
)

// CVM create cvm request param
type CVM struct {
	AppId             string            `json:"appId"`
	ApplyType         int64             `json:"applyType"`
	AppModuleId       int64             `json:"appModuleId"`
	Operator          string            `json:"operator"`
	ApplyNumber       uint              `json:"applyNumber"`
	NoteInfo          string            `json:"noteInfo"`
	VPCId             string            `json:"vpcId"`
	SubnetId          string            `json:"subnetId"`
	Area              string            `json:"area"`
	Zone              string            `json:"zone"`
	ImageId           string            `json:"image_id"`
	ImageName         string            `json:"image_name"`
	InstanceType      string            `json:"instanceType"`
	SecurityGroupId   string            `json:"securityGroupId"`
	SecurityGroupName string            `json:"securityGroupName"`
	SecurityGroupDesc string            `json:"securityGroupDesc"`
	ChargeType        cvmapi.ChargeType `json:"chargeType"`
	ChargeMonths      uint              `json:"chargeMonths"`
	InheritInstanceId string            `json:"inherit_instance_id"`
	BkProductID       int64             `json:"bk_product_id"`
	BkProductName     string            `json:"bk_product_name"`
	SystemDisk        enumor.DiskSpec   `json:"system_disk"`
	DataDisk          []enumor.DiskSpec `json:"data_disk"`
}

// ExecuteApplyOrder CVM生产-创建单据
func (l *logics) ExecuteApplyOrder(kt *kit.Kit, order *types.ApplyOrder) error {
	kt = &kit.Kit{Ctx: context.Background(), User: kt.User, Rid: kt.Rid, AppCode: kt.AppCode, TenantID: kt.TenantID,
		RequestSource: kt.RequestSource}

	// update generate record status to running
	if err := l.updateApplyOrder(order, types.ApplyStatusRunning, "handling", "", 0); err != nil {
		logs.Errorf("failed to create cvm when update generate record, order id: %d, err: %v, rid: %s",
			order.OrderId, err, kt.Rid)
		return err
	}

	// launch cvm request
	request, err := l.buildCvmReq(kt, order)
	if err != nil {
		logs.Errorf("scheduler:logics:execute:apply:order:failed, failed to launch cvm when build cvm request, "+
			"err: %v, order id: %d, rid: %s", err, order.OrderId, kt.Rid)

		// update generate record status to failed
		if subErr := l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); subErr != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %d, err: %v, rid: %s",
				order.OrderId, subErr, kt.Rid)
			return subErr
		}

		return err
	}

	// launch cvm request
	taskId, err := l.createCVM(request)
	if err != nil {
		logs.Errorf("scheduler:logics:execute:apply:order:failed, failed to create cvm when launch generate task, "+
			"order id: %d, err: %v, rid: %s", order.OrderId, err, kt.Rid)

		// update generate record status to failed
		if subErr := l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); subErr != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %d, err: %v, rid: %s",
				order.OrderId, subErr, kt.Rid)
			return subErr
		}

		return err
	}

	// update apply order status to running
	if err = l.updateApplyOrder(order, types.ApplyStatusRunning, "handling", taskId, 0); err != nil {
		logs.Errorf("failed to create cvm when update generate record, order id: %d, taskId: %s, err: %v, rid: %s",
			order.OrderId, taskId, err, kt.Rid)
		return err
	}

	// create cvm device info from task result
	order.TaskId = taskId
	if err = l.CreateCvmFromTaskResult(kt, order); err != nil {
		logs.Errorf("failed to create cvm from task result, order id: %d, err: %v, rid: %s", order.OrderId, err, kt.Rid)
		return err
	}

	return nil
}

// CreateCvmFromTaskResult create cvm from task result
func (l *logics) CreateCvmFromTaskResult(kt *kit.Kit, order *types.ApplyOrder) error {
	if order == nil {
		logs.Errorf("create cvm failed, order is nil, rid: %s", kt.Rid)
		return errors.New("create cvm failed, order is nil")
	}

	if order.TaskId == "" {
		logs.Errorf("create cvm failed, can not find task id, task: %v, rid: %s", cvt.PtrToVal(order), kt.Rid)
		err := fmt.Errorf("create cvm failed, can not find task id, order id: %d", order.OrderId)
		if subErr := l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); subErr != nil {
			logs.Errorf("update apply order failed, err: %v, order id: %d, rid: %s", subErr, order.OrderId, kt.Rid)
			return subErr
		}
		return err
	}

	// check cvm task result
	orderStr := strconv.FormatUint(order.OrderId, 10)
	if err := l.taskLogic.Scheduler().GetGenerator().CheckCVM(kt, order.TaskId, orderStr,
		order.Spec.ChargeType); err != nil {
		logs.Errorf("create cvm product failed, failed to check crp cvm status, orderID: %d, crpTaskID: %s, "+
			"err: %v, rid: %s", order.OrderId, order.TaskId, err, kt.Rid)

		if subErr := l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); subErr != nil {
			logs.Errorf("update apply order failed, err: %v, order id: %d, rid: %s", subErr, order.OrderId, kt.Rid)
			return subErr
		}
		return err
	}

	// get generated cvm instances
	hosts, err := l.listCVM(order.TaskId)
	if err != nil {
		logs.Errorf("failed to list created cvm, err: %v, order id: %d, task id: %s, rid: %s", err, order.OrderId,
			order.TaskId, kt.Rid)
		if subErr := l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); subErr != nil {
			logs.Errorf("update apply order failed, err: %v, order id: %d, rid: %s", subErr, order.OrderId, kt.Rid)
			return subErr
		}
		return err
	}

	// save generated cvm instances info
	if err = l.createDeviceInfo(order, hosts, order.TaskId); err != nil {
		logs.Errorf("create cvm device info failed, err: %v, order id: %d, taskId: %s, rid: %s", err, order.OrderId,
			order.TaskId, kt.Rid)
		if subErr := l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); subErr != nil {
			logs.Errorf("update apply order failed, err: %v, order id: %d, rid: %s", subErr, order.OrderId, kt.Rid)
			return subErr
		}
		return err
	}

	// update apply order status to success
	if err = l.updateApplyOrder(order, types.ApplyStatusSuccess, "success", "", uint(len(hosts))); err != nil {
		logs.Errorf("update apply order failed, err: %v, order id: %d, taskId: %s, rid: %s", err, order.OrderId,
			order.TaskId, kt.Rid)
		return err
	}

	if enumor.RequireType(order.RequireType) == enumor.RequireTypeRollServer {
		appliedTypes := []enumor.AppliedType{enumor.CvmProduceAppliedType}
		subOrderID := strconv.FormatUint(order.OrderId, 10)
		deviceTypeCountMap := map[string]int{order.Spec.DeviceType: len(hosts)}

		if err = l.rsLogic.UpdateSubOrderRollingDeliveredCore(kt, order.BkBizId, subOrderID, appliedTypes,
			deviceTypeCountMap); err != nil {
			logs.Errorf("update rolling delivered cpu field failed, err: %v, suborder_id: %s, bizID: %d, "+
				"deviceTypeCountMap: %v, rid: %s", err, subOrderID, order.BkBizId, deviceTypeCountMap, kt.Rid)

			if subErr := l.updateApplyOrder(order, types.ApplyStatusFailed, err.Error(), "", 0); subErr != nil {
				logs.Errorf("update apply order failed, err: %v, order id: %d, rid: %s", subErr, order.OrderId, kt.Rid)
				return subErr
			}
			return err
		}
	}

	return nil
}

// createCVM starts a cvm creating task(CVM生产-创建单据)
func (l *logics) createCVM(cvm *CVM) (string, error) {
	// construct cvm launch request
	createReq := &cvmapi.OrderCreateReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmLaunchMethod,
		},
		Params: &cvmapi.OrderCreateParams{
			Zone:          cvm.Zone,
			DeptName:      cvmapi.CvmLaunchDeptName,
			ProductName:   cvm.BkProductName,
			Business1Id:   cvmapi.CvmLaunchBiz1Id,
			Business1Name: cvmapi.CvmLaunchBiz1Name,
			Business2Id:   cvmapi.CvmLaunchBiz2Id,
			Business2Name: cvmapi.CvmLaunchBiz2Name,
			// Business3Id:   cvmapi.CvmLaunchBiz3Id,
			// Business3Name: cvmapi.CvmLaunchBiz3Name,
			Business3Id:   662584,
			Business3Name: "CC_SA云化池",
			ProjectId:     int(cvm.BkProductID),
			Image: &cvmapi.Image{
				ImageId:   cvm.ImageId,
				ImageName: cvm.ImageName,
			},
			InstanceType:   cvm.InstanceType,
			SystemDiskType: cvm.SystemDisk.DiskType,
			SystemDiskSize: cvm.SystemDisk.DiskSize,
			DataDisk:       make([]*cvmapi.DataDisk, 0),
			VpcId:          cvm.VPCId,
			SubnetId:       cvm.SubnetId,
			ApplyNum:       int(cvm.ApplyNumber),
			PassWord:       l.cliConf.CvmOpt.CvmLaunchPassword,
			Security: &cvmapi.Security{
				SecurityGroupId:   cvm.SecurityGroupId,
				SecurityGroupName: cvm.SecurityGroupName,
				SecurityGroupDesc: cvm.SecurityGroupDesc,
			},
			UseTime:           time.Now().Format(constant.DateTimeLayout),
			Memo:              cvm.NoteInfo,
			Operator:          cvm.Operator,
			BakOperator:       cvm.Operator,
			ChargeType:        cvm.ChargeType,        // 计费模式
			InheritInstanceId: cvm.InheritInstanceId, // 被继承云主机的实例id
		},
	}

	// 计费模式-包年包月时才需要设置计费时长
	if cvm.ChargeType == cvmapi.ChargeTypePrePaid {
		createReq.Params.ChargeMonths = cvm.ChargeMonths
	}

	// 数据盘可以指定多块硬盘
	for _, dd := range cvm.DataDisk {
		for i := uint(0); i < dd.DiskNum; i++ {
			createReq.Params.DataDisk = append(createReq.Params.DataDisk, &cvmapi.DataDisk{
				DataDiskType: dd.DiskType,
				DataDiskSize: dd.DiskSize,
			})
		}
	}

	// set obs project type
	requireType := enumor.RequireType(cvm.ApplyType)
	createReq.Params.ObsProject = string(requireType.ToObsProject())
	if requireType == enumor.RequireTypeGreenChannel {
		createReq.Params.ResourceType = cvmapi.ResourceTypeQuick
	}

	// call cvm api to launchCvm cvm order
	createResp, err := l.cvm.CreateCvmOrder(nil, nil, createReq)
	if err != nil {
		logs.Errorf("scheduler:logics:create:cvm:failed, err: %v, req: %+v", err, createReq)
		return "", err
	}

	// 记录cvm创建成功的日志，方便排查问题
	// 需要记录crp返回的所有结果包括traceid
	jsonCreateReq, err := json.Marshal(createReq)
	if err != nil {
		logs.Errorf("scheduler:logics:create:cvm:failed to marshal createReq, err: %v", err)
		return "", err
	}
	jsonCreateResp, err := json.Marshal(createResp)
	if err != nil {
		logs.Errorf("scheduler:logics:create:cvm:failed to marshal createResp, err: %v", err)
		return "", err
	}
	logs.Infof("scheduler:logics:create:cvm:end, req: %s, resp: %s", jsonCreateReq, jsonCreateResp)

	if createResp.Error.Code != 0 {
		return "", fmt.Errorf("cvm order create task failed, code: %d, msg: %s", createResp.Error.Code,
			createResp.Error.Message)
	}

	if createResp.Result.OrderId == "" {
		return "", fmt.Errorf("cvm order create task return empty order id")
	}

	return createResp.Result.OrderId, nil
}

// checkCVM checks cvm creating task result
func (l *logics) checkCVM(orderId string) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to query cvm order by id %s, err: %v", orderId, err)
		}

		if obj == nil {
			return false, fmt.Errorf("cvm order %s not found", orderId)
		}

		resp, ok := obj.(*cvmapi.OrderQueryResp)
		if !ok {
			return false, fmt.Errorf("object with order id %s is not a cvm order response: %+v", orderId, obj)
		}

		if resp.Error.Code != 0 {
			return false, fmt.Errorf("query cvm order failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
		}

		if resp.Result == nil {
			return false, fmt.Errorf("query cvm order failed, for result is null, resp: %+v", resp)
		}

		num := len(resp.Result.Data)
		if num != 1 {
			return false, fmt.Errorf("query cvm order return %d orders with order id: %s", num, orderId)
		}

		status := enumor.CrpOrderStatus(resp.Result.Data[0].Status)
		if status != enumor.CrpOrderStatusFinish &&
			status != enumor.CrpOrderStatusReject &&
			status != enumor.CrpOrderStatusFailed {
			return false, fmt.Errorf("cvm order %s handling", orderId)
		}

		if status != enumor.CrpOrderStatusFinish {
			return true, fmt.Errorf("order %s failed, status: %d", resp.Result.Data[0].OrderId,
				resp.Result.Data[0].Status)
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// construct order status request
		req := cvmapi.NewOrderQueryReq(&cvmapi.OrderQueryParam{OrderId: []string{orderId}})

		// call cvm api to query cvm order status
		return l.cvm.QueryCvmOrders(nil, nil, req)
	}

	// TODO: get retry strategy from config
	_, err := utils.Retry(doFunc, checkFunc, 86400, 60)
	return err
}

// listCVM lists created cvm by order id
func (l *logics) listCVM(orderId string) ([]*cvmapi.InstanceItem, error) {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, err
		}
		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// construct order status request
		req := &cvmapi.InstanceQueryReq{
			ReqMeta: cvmapi.ReqMeta{
				Id:      cvmapi.CvmId,
				JsonRpc: cvmapi.CvmJsonRpc,
				Method:  cvmapi.CvmInstanceStatusMethod,
			},
			Params: &cvmapi.InstanceQueryParam{
				OrderId: []string{orderId},
			},
		}
		return l.cvm.QueryCvmInstances(nil, nil, req)
	}

	// TODO: get retry strategy from config
	obj, err := utils.Retry(doFunc, checkFunc, 120, 5)

	if err != nil {
		return nil, err
	}
	resp, ok := obj.(*cvmapi.InstanceQueryResp)
	if !ok {
		return nil, fmt.Errorf("object with order id %s is not a cvm instance response: %+v", orderId, obj)
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		logs.Warnf("get zone capacity failed to marshal capacityResp, orderId: %s, err: %v", orderId, err)
	}
	logs.Infof("scheduler:logics:cvm:list:success, orderId: %s, get cvm instance resp: %s", orderId, respJSON)

	if resp.Error.Code != 0 {
		return nil, fmt.Errorf("list cvm instance failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		return nil, errors.New("list cvm instance failed, for result is null")
	}

	// CRP生产失败没有返回主机列表
	if len(resp.Result.Data) == 0 {
		return nil, fmt.Errorf("list cvm instance empty, can not find crp created cvm, orderID: %s, crpTraceID: %s",
			orderId, resp.TraceId)
	}

	return resp.Result.Data, nil
}

// createDeviceInfo update generate record
func (l *logics) createDeviceInfo(order *types.ApplyOrder, items []*cvmapi.InstanceItem, taskId string) error {
	// 1. save device info to db
	now := time.Now()
	for _, item := range items {
		device := &types.CvmInfo{
			OrderId:   order.OrderId,
			CvmTaskId: taskId,
			CvmInstId: item.InstanceId,
			Ip:        item.LanIp,
			AssetId:   item.AssetId,
			UpdateAt:  now,
		}
		if err := model.Operation().CvmInfo().CreateCvmInfo(context.Background(), device); err != nil {
			logs.Errorf("failed to save device info to db, order id: %d, err: %v", order.OrderId, err)
			return err
		}
	}

	return nil
}

// updateApplyOrder updates generate record
func (l *logics) updateApplyOrder(order *types.ApplyOrder, status types.ApplyStatus, msg, vmTaskId string,
	succNum uint) error {

	filter := &mapstr.MapStr{
		"order_id": order.OrderId,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    status,
		"update_at": now,
	}

	if len(msg) != 0 {
		doc["message"] = msg
	}

	if len(vmTaskId) != 0 {
		link := cvmapi.CvmOrderLinkPrefix + vmTaskId
		doc["task_id"] = vmTaskId
		doc["task_link"] = link
	}

	if status == types.ApplyStatusSuccess || status == types.ApplyStatusFailed {
		doc["success_num"] = succNum
		doc["pending_num"] = 0
		doc["failed_num"] = order.Total - succNum
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, &doc); err != nil {
		logs.Errorf("failed to update apply order, order id: %d, update: %+v, err: %v", order.OrderId, doc, err)
		return err
	}

	return nil
}

// buildCvmReq construct a cvm creating task request(CVM生产-创建单据)
func (l *logics) buildCvmReq(kt *kit.Kit, order *types.ApplyOrder) (*CVM, error) {
	kt = kt.NewSubKit()
	kt.Ctx = context.Background()
	// 记录cvm请求日志，方便排查问题
	logs.Infof("scheduler:logics:build:cvm:request:start, order: %+v, rid: %s", cvt.PtrToVal(order), kt.Rid)
	// TODO: get parameters from config
	// construct cvm launch req
	req := &CVM{
		AppId:             "931",
		ApplyType:         order.RequireType,
		AppModuleId:       51524,
		Operator:          order.User,
		ApplyNumber:       order.Total,
		NoteInfo:          order.Remark,
		Area:              order.Spec.Region,
		Zone:              order.Spec.Zone,
		InstanceType:      order.Spec.DeviceType,
		SystemDisk:        order.Spec.SystemDisk,
		DataDisk:          order.Spec.DataDisk,
		ChargeType:        order.Spec.ChargeType, // 计费模式
		InheritInstanceId: order.Spec.InheritInstanceId,
	}

	// 计费模式-包年包月时才需要设置计费时长
	if req.ChargeType == cvmapi.ChargeTypePrePaid {
		req.ChargeMonths = order.Spec.ChargeMonths
	}

	// vpc and subnet
	if order.Spec.Vpc != "" {
		req.VPCId = order.Spec.Vpc
	} else {
		vpc, err := l.confLogic.Vpc().GetRegionDftVpc(kt, order.Spec.Region)
		if err != nil {
			logs.Errorf("scheduler:logics:build:cvm:request:failed, build cvm req get cvm vpc failed, err: %v, "+
				"order: %+v, rid: %s", err, cvt.PtrToVal(order), kt.Rid)
			return nil, err
		}
		req.VPCId = vpc
	}
	if order.Spec.Subnet != "" {
		req.SubnetId = order.Spec.Subnet
	} else {
		subnet, leftIp, err := l.getCvmSubnet(kt, order.Spec.Region, order.Spec.Zone, req.VPCId)
		if err != nil {
			logs.Errorf("scheduler:logics:build:cvm:request:failed, build cvm req get subnet failed, err: %v, "+
				"req: %+v, order: %+v, rid: %s", err, req, cvt.PtrToVal(order), kt.Rid)
			return nil, err
		}
		req.SubnetId = subnet
		if leftIp < order.Total {
			// set apply number to min(replicas, leftIp)
			req.ApplyNumber = leftIp
		}
	}
	// image
	req.ImageId = order.Spec.ImageId
	// security group
	sg, err := l.confLogic.Sg().GetRegionDftSg(kt, order.Spec.Region)
	if err != nil {
		logs.Errorf("build cvm req get cvm default secGroup failed, err: %v, order: %+v, rid: %s", err,
			cvt.PtrToVal(order), kt.Rid)
		return nil, err
	}
	req.SecurityGroupId = sg.SgID
	req.SecurityGroupName = sg.SgName
	req.SecurityGroupDesc = sg.SgDesc

	productID, productName, err := l.getProductMsg(kt, order)
	if err != nil {
		logs.Errorf("get product message failed, err: %v, order: %+v, rid: %s", err, cvt.PtrToVal(order), kt.Rid)
		return nil, err
	}

	req.BkProductID = productID
	req.BkProductName = productName

	logs.Infof("scheduler:logics:build:cvm:request:end, order: %+v, req: %+v, rid: %s",
		cvt.PtrToVal(order), cvt.PtrToVal(req), kt.Rid)
	return req, nil
}

func (l *logics) getProductMsg(kt *kit.Kit, order *types.ApplyOrder) (int64, string, error) {
	if enumor.RequireType(order.RequireType).IsUseManageBizPlan() {
		return cvmapi.CvmLaunchProjectId, cvmapi.CvmLaunchProductName, nil
	}

	param := &cmdb.SearchBizCompanyCmdbInfoParams{BizIDs: []int64{order.BkBizId}}
	resp, err := l.cmdbCli.SearchBizCompanyCmdbInfo(kt, param)
	if err != nil {
		logs.Errorf("failed to search biz belonging, err: %v, param: %+v, rid: %s", err, *param, kt.Rid)
		return 0, "", err
	}
	if resp == nil || len(*resp) != 1 {
		logs.Errorf("search biz belonging, but resp is empty or len resp != 1, rid: %s", kt.Rid)
		return 0, "", errors.New("search biz belonging, but resp is empty or len resp != 1")
	}

	bizBelong := (*resp)[0]
	return bizBelong.BkProductID, bizBelong.BkProductName, nil
}

func (l *logics) getCvmSubnet(kt *kit.Kit, region, zone, vpc string) (string, uint, error) {
	if len(zone) <= 0 {
		return "", 0, fmt.Errorf("get cvm subnet failed, zone is empty")
	}

	subnetList, err := l.getSubnetList(kt, region, zone, vpc)
	if err != nil {
		return "", 0, err
	}

	subnetReq := &cfgtypes.GetAllSubnetReq{
		Region:     region,
		CloudVpcID: vpc,
	}
	// 园区-分区Campus
	if len(zone) > 0 && zone != cvmapi.CvmSeparateCampus {
		subnetReq.Zones = []string{zone}
	}
	cfgSubnets, err := l.confLogic.Subnet().GetAllSubnet(kt, subnetReq)
	if err != nil {
		logs.Errorf("failed to get config cvm subnet info, err: %v, rid: %s", err, kt.Rid)
		return "", 0, err
	}
	mapIdTosubnet := make(map[string]*cfgtypes.Subnet)
	for _, subnet := range cfgSubnets.Info {
		mapIdTosubnet[subnet.SubnetId] = subnet
	}

	subnetId := ""
	leftIp := uint(0)
	for _, subnet := range subnetList {
		// subnet is not effective
		if _, ok := mapIdTosubnet[subnet.Id]; !ok {
			continue
		}
		// use subnet name with prefix "cvm_use" only
		if !strings.HasPrefix(subnet.Name, "cvm_use") {
			continue
		}
		if subnet.LeftIpNum >= int(leftIp) {
			subnetId = subnet.Id
			leftIp = uint(subnet.LeftIpNum)
		}
	}

	if subnetId == "" {
		logs.Errorf("getCvmSubnet found no subnet with region: %s, zone: %s, vpc: %s, cfgSubnets: %+v, "+
			"subnetList: %+v, rid: %s", region, zone, vpc, cvt.PtrToSlice(cfgSubnets.Info),
			cvt.PtrToSlice(subnetList), kt.Rid)
		return "", 0, fmt.Errorf("found no subnet with region: %s, zone: %s, vpc: %s", region, zone, vpc)
	}

	if leftIp <= 0 {
		logs.Errorf("getCvmSubnet found no subnet with left ip > 0, region: %s, zone: %s, vpc: %s, cfgSubnets: %+v, "+
			"subnetList: %+v, rid: %s", region, zone, vpc, cvt.PtrToSlice(cfgSubnets.Info),
			cvt.PtrToSlice(subnetList), kt.Rid)
		return subnetId, leftIp, fmt.Errorf("found no subnet with left ip > 0, region: %s, zone: %s, vpc: %s", region,
			zone, vpc)
	}

	return subnetId, leftIp, nil
}

func (l *logics) getSubnetList(kt *kit.Kit, region string, zone string, vpc string) ([]*cvmapi.SubnetInfo, error) {
	zones := make([]string, 0)
	// 园区-分区Campus，获取该区域下的可用区列表
	if zone == cvmapi.CvmSeparateCampus {
		zoneResp, err := l.confLogic.Zone().GetZone(kt, &cfgtypes.GetZoneParam{
			Region: []string{region},
		})
		if err != nil {
			logs.Errorf("failed to get cvm subnet zone list, err: %v, region: %s, zone: %s, vpc: %s, rid: %s",
				err, region, zone, vpc, kt.Rid)
			return nil, err
		}

		for _, zoneItem := range zoneResp.Info {
			zones = append(zones, zoneItem.Zone)
		}
	} else {
		zones = append(zones, zone)
	}

	var subnetList = make([]*cvmapi.SubnetInfo, 0)
	// 遍历可用区，获取子网列表
	for _, zoneValue := range zones {
		subnetReq := cvmapi.SubnetRealParam{
			Region:      region,
			CloudCampus: zoneValue,
			VpcId:       vpc,
		}
		resp, err := l.cvm.QueryRealCvmSubnet(kt, subnetReq)
		if err != nil {
			logs.Errorf("failed to loop get cvm subnet info, err: %v, region: %s, zone: %s, vpc: %s, rid: %s",
				err, region, zoneValue, vpc, kt.Rid)
			return nil, err
		}
		subnetList = append(subnetList, resp.Result...)
	}
	return subnetList, nil
}
