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

package azure

import (
	"fmt"

	protocloud "hcm/pkg/api/data-service/cloud"
	protodisk "hcm/pkg/api/data-service/cloud/disk"
	"hcm/pkg/criteria/enumor"
)

// Deliver 执行资源交付
func (a *ApplicationOfCreateAzureCvm) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	// 创建主机
	result, err := a.Client.HCService().Azure.Cvm.BatchCreateCvm(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		a.toHcProtoAzureBatchCreateReq(),
	)
	if err != nil || result == nil {
		return enumor.DeliverError, map[string]interface{}{"error": err}, err
	}

	deliverDetail := map[string]interface{}{"result": result}
	// 全部失败
	if len(result.FailedCloudIDs) == int(a.req.RequiredCount) {
		err = fmt.Errorf("all cvm create failed, message: %s", result.FailedMessage)
		deliverDetail["error"] = err
		return enumor.DeliverError, deliverDetail, err
	}

	// 将创建成功的主机进行业务分配
	cvmIDs, err := a.assignToBiz(result.SuccessCloudIDs)
	deliverDetail["error"] = err
	deliverDetail["cvm_ids"] = cvmIDs
	if err != nil {

		return enumor.DeliverError, deliverDetail, err
	}

	status := enumor.Completed
	// 部分成功
	if len(result.SuccessCloudIDs) != int(a.req.RequiredCount) {
		status = enumor.DeliverPartial
	}

	return status, deliverDetail, nil
}

func (a *ApplicationOfCreateAzureCvm) assignToBiz(cloudCvmIDs []string) ([]string, error) {
	req := a.req
	// 云ID查询主机
	cvmInfo, err := a.ListCvm(a.vendor, req.AccountID, cloudCvmIDs)
	if err != nil {
		return []string{}, err
	}
	cvmIDs := make([]string, 0, len(cvmInfo))
	for _, cvm := range cvmInfo {
		cvmIDs = append(cvmIDs, cvm.ID)
	}

	// 主机分配给业务
	err = a.Client.DataService().Global.Cvm.BatchUpdateCvmCommonInfo(
		a.Cts.Kit.Ctx,
		a.Cts.Kit.Header(),
		&protocloud.CvmCommonInfoBatchUpdateReq{IDs: cvmIDs, BkBizID: req.BkBizID},
	)
	if err != nil {
		return cvmIDs, err
	}

	// 主机关联资源分配给业务，目前只有硬盘是一同创建出来的
	diskIDs, err := a.ListDiskIDByCvm(cvmIDs)
	if err != nil {
		return cvmIDs, err
	}
	if len(diskIDs) > 0 {
		// 硬盘分配给业务
		_, err = a.Client.DataService().Global.BatchUpdateDisk(
			a.Cts.Kit.Ctx,
			a.Cts.Kit.Header(),
			&protodisk.DiskBatchUpdateReq{
				IDs:     diskIDs,
				BkBizID: uint64(req.BkBizID),
			},
		)
		if err != nil {
			return cvmIDs, err
		}
	}
	return cvmIDs, nil
}
