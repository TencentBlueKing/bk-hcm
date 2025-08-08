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

// Package cvm ...
package cvm

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	lgccvm "hcm/cmd/cloud-server/logics/cvm"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	protoaudit "hcm/pkg/api/data-service/audit"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// BatchResetAsyncCvm batch reset async cvm.
func (svc *cvmSvc) BatchResetAsyncCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.batchResetAsyncCvm(cts, constant.UnassignedBiz, handler.ResOperateAuth)
}

// BatchResetAsyncBizCvm batch reset async biz cvm.
func (svc *cvmSvc) BatchResetAsyncBizCvm(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, err
	}
	return svc.batchResetAsyncCvm(cts, bkBizID, handler.BizOperateAuth)
}

func (svc *cvmSvc) batchResetAsyncCvm(cts *rest.Contexts, bkBizID int64, validHandler handler.ValidWithAuthHandler) (
	any, error) {

	req := new(cscvm.BatchResetCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cvmIDs := make([]string, 0, len(req.Hosts))
	cvmIDMap := make(map[string]cscvm.BatchCvmHostItem)
	for _, host := range req.Hosts {
		cvmIDs = append(cvmIDs, host.ID)
		cvmIDMap[host.ID] = host
	}

	basicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.CvmCloudResType,
		IDs:          cvmIDs,
		Fields:       append(types.CommonBasicInfoFields, "region", "recycle_status"),
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, basicInfoReq)
	if err != nil {
		logs.Errorf("reset cvm list basic info failed, err: %v, cvmIDs: %v, rid: %s", err, cvmIDs, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.Cvm,
		Action: meta.ResetSystem, BasicInfos: basicInfoMap})
	if err != nil {
		logs.Errorf("reset cvm auth failed, err: %v, cvmIDs: %v, rid: %s", err, cvmIDs, cts.Kit.Rid)
		return nil, err
	}

	// 创建审计记录
	for _, ids := range slice.Split(cvmIDs, constant.BatchOperationMaxLimit) {
		if err = svc.audit.ResBaseOperationAudit(
			cts.Kit, enumor.CvmAuditResType, protoaudit.ResetSystem, ids); err != nil {
			logs.Errorf("create reset cvm operation audit failed, err: %v, cvmIDs: %v, rid: %s",
				err, ids, cts.Kit.Rid)
			return nil, err
		}
	}

	taskManagementID, err := svc.createCvmResetTaskManage(cts.Kit, bkBizID, cvmIDs, cvmIDMap, req.Pwd)
	if err != nil {
		logs.Errorf("create cvm reset task manage failed, err: %v, cvmIDs: %v, rid: %s", err, cvmIDs, cts.Kit.Rid)
		return nil, err
	}

	return cscvm.BatchCvmOperateResp{
		TaskManagementID: taskManagementID,
	}, nil
}

func (svc *cvmSvc) createCvmResetTaskManage(kt *kit.Kit, bkBizID int64, cvmIDs []string,
	cvmIDMap map[string]cscvm.BatchCvmHostItem, pwd string) (string, error) {

	cvmList, err := svc.batchListCvmByIDs(kt, cvmIDs)
	if err != nil {
		logs.Errorf("batch list cvm by ids failed, err: %v, cvmIDs: %v, rid: %s", err, cvmIDs, kt.Rid)
		return "", err
	}

	taskManageReq := &lgccvm.TaskManageBaseReq{
		Vendors:       make([]enumor.Vendor, 0),
		AccountIDs:    make([]string, 0),
		BkBizID:       bkBizID,
		Source:        enumor.TaskManagementSourceAPI,
		Resource:      enumor.TaskManagementResCVM,
		TaskOperation: enumor.TaskCvmResetSystem,
		TaskType:      enumor.ResetCvmTaskType,
		Details:       make([]*lgccvm.CvmResetTaskDetailReq, 0),
	}

	uniqueID, err := calCvmResetUniqueID(kt, bkBizID, cvmIDs)
	if err != nil {
		logs.Errorf("cal cvm reset unique key failed, err: %v, cvmIDs: %v, rid: %s", err, cvmIDs, kt.Rid)
		return "", err
	}
	taskManageReq.UniqueID = uniqueID

	vendors := make([]enumor.Vendor, 0)
	accountIDs := make([]string, 0)
	for _, host := range cvmList {
		taskDetail := &lgccvm.CvmResetTaskDetailReq{
			CvmBatchOperateHostInfo: cscvm.CvmBatchOperateHostInfo{
				ID:                   host.ID,
				Vendor:               host.Vendor,
				AccountID:            host.AccountID,
				CloudID:              host.CloudID,
				PrivateIPv4Addresses: host.PrivateIPv4Addresses,
				PrivateIPv6Addresses: host.PrivateIPv6Addresses,
				PublicIPv4Addresses:  host.PublicIPv4Addresses,
				PublicIPv6Addresses:  host.PublicIPv6Addresses,
				CloudVpcIDs:          host.CloudVpcIDs,
				CloudSubnetIDs:       host.CloudSubnetIDs,
				DeviceType:           cvmIDMap[host.ID].DeviceType,
				Region:               host.Region,
				Zone:                 host.Zone,
			},
			ImageNameOld: cvmIDMap[host.ID].ImageNameOld,
			CloudImageID: cvmIDMap[host.ID].CloudImageID,
			ImageName:    cvmIDMap[host.ID].ImageName,
			Pwd:          pwd,
		}
		vendors = append(vendors, host.Vendor)
		accountIDs = append(accountIDs, host.AccountID)
		taskManageReq.Details = append(taskManageReq.Details, taskDetail)
	}
	taskManageReq.Vendors = slice.Unique(vendors)
	taskManageReq.AccountIDs = slice.Unique(accountIDs)

	taskManageID, err := svc.cvmLgc.CvmResetSystem(kt, taskManageReq)
	if err != nil {
		logs.Errorf("reset cvm system failed, err: %v, cvmIDs: %v, taskManageReq: %+v, rid: %s", err, cvmIDs,
			cvt.PtrToVal(taskManageReq), kt.Rid)
		return "", err
	}
	return taskManageID, nil
}

func calCvmResetUniqueID(kt *kit.Kit, bkBizID int64, ids []string) (string, error) {
	sort.Strings(ids)
	idsStr := strings.Join(ids, "")

	// 计算 MD5 哈希值
	hash := md5.New()
	hash.Write([]byte(idsStr))
	hashBytes := hash.Sum(nil)

	// 将哈希值转换为十六进制字符串
	hashString := fmt.Sprintf("%d-%s", bkBizID, hex.EncodeToString(hashBytes))
	logs.Infof("cal cvm reset unique key, hashString: %s, bkBizID: %d, ids: %v, idsStr: %s, rid: %s",
		hashString, bkBizID, ids, idsStr, kt.Rid)
	return hashString, nil
}
