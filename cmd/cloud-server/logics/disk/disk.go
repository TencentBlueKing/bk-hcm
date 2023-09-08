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

// Package disk ...
package disk

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// Interface define disk interface.
type Interface interface {
	DetachDisk(kt *kit.Kit, vendor enumor.Vendor, cvmID, diskID, accountID string) error
	DeleteDisk(kt *kit.Kit, vendor enumor.Vendor, diskID, accountID string) error
	DeleteRecycledDisk(kt *kit.Kit, infoMap map[string]types.CloudResourceBasicInfo) (*core.BatchOperateResult, error)
}

type disk struct {
	client *client.ClientSet
	audit  audit.Interface
}

// NewDisk new disk.
func NewDisk(client *client.ClientSet, audit audit.Interface) Interface {
	return &disk{
		client: client,
		audit:  audit,
	}
}

// DetachDisk detach disk from cvm.
// TODO remove account id parameter, this should be acquired in hc-service.
func (d *disk) DetachDisk(kt *kit.Kit, vendor enumor.Vendor, cvmID, diskID, accountID string) error {
	// create audit
	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.DiskAuditResType,
		ResID:             diskID,
		Action:            protoaudit.Disassociate,
		AssociatedResType: enumor.CvmAuditResType,
		AssociatedResID:   cvmID,
	}

	err := d.audit.ResOperationAudit(kt, operationInfo)
	if err != nil {
		logs.Errorf("create detach disk audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	detachReq := &hcproto.DiskDetachReq{
		AccountID: accountID,
		CvmID:     cvmID,
		DiskID:    diskID,
	}

	switch vendor {
	case enumor.TCloud:
		return d.client.HCService().TCloud.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	case enumor.Aws:
		return d.client.HCService().Aws.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	case enumor.HuaWei:
		return d.client.HCService().HuaWei.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	case enumor.Gcp:
		return d.client.HCService().Gcp.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	case enumor.Azure:
		return d.client.HCService().Azure.Disk.DetachDisk(kt.Ctx, kt.Header(), detachReq)
	default:
		return errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

// DeleteDisk delete disk.
// TODO remove account id parameter, this should be acquired in hc-service.
func (d *disk) DeleteDisk(kt *kit.Kit, vendor enumor.Vendor, diskID, accountID string) error {
	// create delete audit.
	err := d.audit.ResDeleteAudit(kt, enumor.DiskAuditResType, []string{diskID})
	if err != nil {
		logs.Errorf("create delete audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	deleteReq := &hcproto.DiskDeleteReq{DiskID: diskID, AccountID: accountID}

	switch vendor {
	case enumor.TCloud:
		return d.client.HCService().TCloud.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Aws:
		return d.client.HCService().Aws.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	case enumor.HuaWei:
		return d.client.HCService().HuaWei.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Gcp:
		return d.client.HCService().Gcp.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Azure:
		return d.client.HCService().Azure.Disk.DeleteDisk(kt.Ctx, kt.Header(), deleteReq)
	default:
		return errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

// DeleteRecycledDisk batch delete recycled disk.
func (d *disk) DeleteRecycledDisk(kt *kit.Kit, basicInfoMap map[string]types.CloudResourceBasicInfo) (
	*core.BatchOperateResult, error) {

	if len(basicInfoMap) == 0 {
		return nil, nil
	}

	if len(basicInfoMap) > constant.BatchOperationMaxLimit {
		return nil, errf.Newf(errf.InvalidParameter, "disk length should <= %d", constant.BatchOperationMaxLimit)
	}

	ids := make([]string, 0, len(basicInfoMap))
	for id := range basicInfoMap {
		ids = append(ids, id)
	}

	// check if disks are all detached
	relReq := &cloud.DiskCvmRelListReq{
		Filter: tools.ContainersExpression("disk_id", ids),
		Page:   &core.BasePage{Count: true},
	}
	relRes, err := d.client.DataService().Global.ListDiskCvmRel(kt.Ctx, kt.Header(), relReq)
	if err != nil {
		return nil, err
	}

	if converter.PtrToVal(relRes.Count) > 0 {
		logs.Errorf("some recycled disks(ids: %+v) are attached, cannot be deleted, rid: %s", ids, kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "recycled disk is attached, cannot be deleted")
	}

	res := new(core.BatchOperateResult)

	// delete disk
	for _, id := range ids {
		info := basicInfoMap[id]
		err = d.DeleteDisk(kt, info.Vendor, id, info.AccountID)
		if err != nil {
			res.Failed = &core.FailedInfo{ID: id, Error: err}
			return res, err
		}
		res.Succeeded = append(res.Succeeded, id)
	}

	return nil, nil
}
