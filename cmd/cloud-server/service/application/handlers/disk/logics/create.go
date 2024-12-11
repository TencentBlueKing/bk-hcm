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

package logics

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud/disk"
	hcproto "hcm/pkg/api/hc-service/disk"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// CheckResultAndAssign ...
func CheckResultAndAssign(kt *kit.Kit, cli *dataservice.Client, result *hcproto.BatchCreateResult,
	diskCount uint32, bkBizID int64, audit audit.Interface, region string, vendor enumor.Vendor) (
	enumor.ApplicationStatus, map[string]interface{}, error) {

	deliverDetail := map[string]interface{}{"result": result}

	// 全部失败
	if len(result.SuccessCloudIDs) == 0 {
		err := fmt.Errorf("all disk create failed, message: %s", result.FailedMessage)
		deliverDetail["error"] = err.Error()
		return enumor.DeliverError, deliverDetail, err
	}

	// 如果部分成功，需要日志打印
	if len(result.SuccessCloudIDs) != int(diskCount) {
		logs.Warnf("request hc service to batch create cvm partial failed, result: %v, rid: %s", result, kt.Rid)
	}

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("cloud_id", result.SuccessCloudIDs),
			tools.RuleEqual("region", region),
			tools.RuleEqual("vendor", vendor),
		),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id"},
	}
	listResult, err := cli.Global.ListDisk(kt, listReq)
	if err != nil {
		deliverDetail["error"] = err.Error()
		return enumor.DeliverError, deliverDetail, err
	}

	ids := make([]string, 0, len(listResult.Details))
	for _, one := range listResult.Details {
		ids = append(ids, one.ID)
	}

	_, err = cli.Global.BatchUpdateDisk(kt, &dataproto.DiskBatchUpdateReq{IDs: ids, BkBizID: uint64(bkBizID)})
	if err != nil {
		deliverDetail["error"] = err.Error()
		return enumor.DeliverError, deliverDetail, err
	}

	// create deliver audit
	err = audit.ResDeliverAudit(kt, enumor.DiskAuditResType, ids, bkBizID)
	if err != nil {
		deliverDetail["error"] = err.Error()
		return enumor.DeliverError, deliverDetail, err
	}

	deliverDetail["disk_ids"] = ids
	status := enumor.Completed
	// 部分成功
	if len(result.SuccessCloudIDs) != int(diskCount) {
		status = enumor.DeliverPartial
	}

	return status, deliverDetail, nil
}
