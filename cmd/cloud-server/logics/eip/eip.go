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

// Package eip ...
package eip

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/pkg/api/cloud-server/recycle"
	"hcm/pkg/api/core"
	recyclerecord "hcm/pkg/api/core/recycle-record"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	dataeip "hcm/pkg/api/data-service/cloud/eip"
	hcproto "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// Interface define eip interface.
type Interface interface {
	AssociateEip(kt *kit.Kit, vendor enumor.Vendor, eipID, cvmID, nicID, accountID string) error
	DisassociateEip(kt *kit.Kit, vendor enumor.Vendor, eipID, cvmID, nicID, accountID string) error

	DeleteEip(kt *kit.Kit, vendor enumor.Vendor, eipId string) error
	BatchGetEipInfo(kt *kit.Kit, cvmStatus map[string]*recycle.CvmDetail) error
	BatchUnbind(kt *kit.Kit, cvmStatus map[string]*recycle.CvmDetail) (failed []string, err error)
	BatchRebind(kt *kit.Kit, cvmRecycleMap map[string]*recycle.CvmDetail) error
}

type eip struct {
	client *client.ClientSet
	audit  audit.Interface
}

// NewEip new eip.
func NewEip(client *client.ClientSet, audit audit.Interface) Interface {
	return &eip{
		client: client,
		audit:  audit,
	}
}

// AssociateEip associate eip from cvm.
// TODO remove account id parameter, this should be acquired in hc-service.
// TODO confirm if bind nic or cvm scenario needs to be separated, and how to deal with association with both nic & cvm.
func (e *eip) AssociateEip(kt *kit.Kit, vendor enumor.Vendor, eipID, cvmID, nicID, accountID string) error {
	switch vendor {
	case enumor.TCloud:
		if nicID != "" {
			err := e.associateEipAudit(kt, enumor.NetworkInterfaceAuditResType, eipID, nicID)
			if err != nil {
				return err
			}
		} else {
			err := e.associateEipAudit(kt, enumor.CvmAuditResType, eipID, cvmID)
			if err != nil {
				return err
			}
		}

		return e.client.HCService().TCloud.Eip.AssociateEip(kt.Ctx, kt.Header(), &hcproto.TCloudEipAssociateReq{
			AccountID: accountID,
			CvmID:     cvmID,
			EipID:     eipID,
		})
	case enumor.Aws:
		err := e.associateEipAudit(kt, enumor.CvmAuditResType, eipID, cvmID)
		if err != nil {
			return err
		}

		return e.client.HCService().Aws.Eip.AssociateEip(kt.Ctx, kt.Header(), &hcproto.AwsEipAssociateReq{
			AccountID: accountID,
			CvmID:     cvmID,
			EipID:     eipID,
		})
	case enumor.HuaWei:
		err := e.associateEipAudit(kt, enumor.NetworkInterfaceAuditResType, eipID, nicID)
		if err != nil {
			return err
		}

		return e.client.HCService().HuaWei.Eip.AssociateEip(kt.Ctx, kt.Header(), &hcproto.HuaWeiEipAssociateReq{
			AccountID:          accountID,
			CvmID:              cvmID,
			EipID:              eipID,
			NetworkInterfaceID: nicID,
		})
	case enumor.Gcp:
		err := e.associateEipAudit(kt, enumor.NetworkInterfaceAuditResType, eipID, nicID)
		if err != nil {
			return err
		}

		return e.client.HCService().Gcp.Eip.AssociateEip(kt.Ctx, kt.Header(), &hcproto.GcpEipAssociateReq{
			AccountID:          accountID,
			CvmID:              cvmID,
			EipID:              eipID,
			NetworkInterfaceID: nicID,
		})
	case enumor.Azure:
		err := e.associateEipAudit(kt, enumor.NetworkInterfaceAuditResType, eipID, nicID)
		if err != nil {
			return err
		}

		return e.client.HCService().Azure.Eip.AssociateEip(kt.Ctx, kt.Header(), &hcproto.AzureEipAssociateReq{
			AccountID:          accountID,
			CvmID:              cvmID,
			EipID:              eipID,
			NetworkInterfaceID: nicID,
		})
	default:
		return errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (e *eip) associateEipAudit(kt *kit.Kit, resType enumor.AuditResourceType, eipID, resID string) error {
	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.EipAuditResType,
		ResID:             eipID,
		Action:            protoaudit.Associate,
		AssociatedResType: resType,
		AssociatedResID:   resID,
	}

	err := e.audit.ResOperationAudit(kt, operationInfo)
	if err != nil {
		logs.Errorf("create associate eip audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DisassociateEip disassociate eip from cvm.
// TODO remove account id parameter, this should be acquired in hc-service.
// TODO confirm if bind nic or cvm scenario needs to be separated, and how to deal with association with both nic & cvm.
func (e *eip) DisassociateEip(kt *kit.Kit, vendor enumor.Vendor, eipID, cvmID, nicID, accountID string) error {
	// TODO 增加审计

	switch vendor {
	case enumor.TCloud:
		err := e.disassociateEipAudit(kt, enumor.CvmAuditResType, eipID, cvmID)
		if err != nil {
			return err
		}

		return e.client.HCService().TCloud.Eip.DisassociateEip(kt.Ctx, kt.Header(), &hcproto.TCloudEipDisassociateReq{
			AccountID: accountID,
			CvmID:     cvmID,
			EipID:     eipID,
		})
	case enumor.Aws:
		err := e.disassociateEipAudit(kt, enumor.CvmAuditResType, eipID, cvmID)
		if err != nil {
			return err
		}

		return e.client.HCService().Aws.Eip.DisassociateEip(kt.Ctx, kt.Header(), &hcproto.AwsEipDisassociateReq{
			AccountID: accountID,
			CvmID:     cvmID,
			EipID:     eipID,
		})
	case enumor.HuaWei:
		err := e.disassociateEipAudit(kt, enumor.NetworkInterfaceAuditResType, eipID, nicID)
		if err != nil {
			return err
		}

		return e.client.HCService().HuaWei.Eip.DisassociateEip(kt.Ctx, kt.Header(), &hcproto.HuaWeiEipDisassociateReq{
			AccountID: accountID,
			CvmID:     cvmID,
			EipID:     eipID,
		})
	case enumor.Gcp:
		err := e.disassociateEipAudit(kt, enumor.NetworkInterfaceAuditResType, eipID, nicID)
		if err != nil {
			return err
		}

		return e.client.HCService().Gcp.Eip.DisassociateEip(kt.Ctx, kt.Header(), &hcproto.GcpEipDisassociateReq{
			AccountID:          accountID,
			CvmID:              cvmID,
			EipID:              eipID,
			NetworkInterfaceID: nicID,
		})
	case enumor.Azure:
		err := e.disassociateEipAudit(kt, enumor.NetworkInterfaceAuditResType, eipID, nicID)
		if err != nil {
			return err
		}

		return e.client.HCService().Azure.Eip.DisassociateEip(kt.Ctx, kt.Header(), &hcproto.AzureEipDisassociateReq{
			AccountID:          accountID,
			CvmID:              cvmID,
			EipID:              eipID,
			NetworkInterfaceID: nicID,
		})
	default:
		return errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
}

func (e *eip) disassociateEipAudit(kt *kit.Kit, resType enumor.AuditResourceType, eipID, resID string) error {
	operationInfo := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.EipAuditResType,
		ResID:             eipID,
		Action:            protoaudit.Disassociate,
		AssociatedResType: resType,
		AssociatedResID:   resID,
	}

	err := e.audit.ResOperationAudit(kt, operationInfo)
	if err != nil {
		logs.Errorf("create associate eip audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// BatchUnbind 批量解绑eip
func (e *eip) BatchUnbind(kt *kit.Kit, cvmRecycleMap map[string]*recycle.CvmDetail) (failed []string,
	err error) {

	kt = kt.NewSubKit()

	for _, detail := range cvmRecycleMap {
		if detail.FailedAt != "" {
			continue
		}
		for _, eip := range detail.EipList {
			err := e.DisassociateEip(kt, detail.Vendor, detail.CvmID,
				eip.NicID, eip.EipID, detail.AccountID)
			if err != nil {
				eip.Err = err
				detail.FailedAt = enumor.EipCloudResType
				failed = append(failed, detail.CvmID)
				logs.Errorf("failed to unbind eip ，err: %v cvmID: %s, eipID: %s, rid:%s",
					err, detail.CvmID, eip.EipID, kt.Rid)
				break
			}
		}
	}
	return failed, nil
}

// BatchGetEipInfo 批量获取Cvm的Eip挂载信息，填充到传入的recycle.CvmRecycleDetail中
func (e *eip) BatchGetEipInfo(kt *kit.Kit, cvmDetail map[string]*recycle.CvmDetail) (err error) {
	if len(cvmDetail) == 0 {
		return nil
	}
	if len(cvmDetail) > constant.BatchOperationMaxLimit {
		return errf.Newf(errf.InvalidParameter, "cvmIDs should <= %d", constant.BatchOperationMaxLimit)
	}
	cvmIDs := converter.MapKeyToStringSlice(cvmDetail)
	relReq := &cloud.EipCvmRelListReq{
		Filter: tools.ContainersExpression("cvm_id", cvmIDs),
		Page:   core.NewDefaultBasePage(),
	}
	cvmEipRel, err := e.client.DataService().Global.ListEipCvmRel(kt.Ctx, kt.Header(), relReq)
	if err != nil {
		logs.Errorf("fail to ListEipCvmRel, cvm_id: %v,err: %v, rid: %s", cvmIDs, err, kt.Rid)
		return err
	}

	// list eip for nic
	eipIDs := slice.Map(cvmEipRel.Details, func(v *cloud.EipCvmRelResult) string { return v.EipID })
	if len(eipIDs) == 0 {
		return nil
	}
	eipReq := &dataeip.EipListReq{
		Filter: tools.ContainersExpression("id", eipIDs),
		Page:   core.NewDefaultBasePage(),
	}
	eipRes, err := e.client.DataService().Global.ListEip(kt.Ctx, kt.Header(), eipReq)
	if err != nil {
		logs.Errorf("fail to ListEip, err: %v, eipIDs: %v, rid: %s", err, eipIDs, kt.Rid)
		return err
	}
	eipMap := map[string]*dataeip.EipResult{}
	for _, eipDetail := range eipRes.Details {
		eipMap[eipDetail.ID] = eipDetail
	}

	// fill eip cvm relation into cvm detail map
	for _, ceRel := range cvmEipRel.Details {
		cvmRecycleDetail := cvmDetail[ceRel.CvmID]
		if cvmRecycleDetail == nil {
			logs.Errorf("ListEipCvmRel return unknown cvm ID or cvm detail map corrupted, cvmId: %v, rid: %s",
				ceRel.CvmID, kt.Rid)
		}
		var nic string
		switch cvmRecycleDetail.Vendor {
		case enumor.Azure, enumor.Gcp, enumor.HuaWei:
			// 	Azure、Gcp、华为云需要查询网卡id
			nic = converter.PtrToVal(eipMap[ceRel.EipID].InstanceID)
		default:
			nic = ""
		}
		cvmRecycleDetail.EipList = append(
			cvmRecycleDetail.EipList, recyclerecord.EipBindInfo{EipID: ceRel.EipID, NicID: nic},
		)

	}
	return nil
}

// BatchRebind 批量重新绑定eip
func (e *eip) BatchRebind(kt *kit.Kit, cvmRecycleMap map[string]*recycle.CvmDetail) error {
	for _, detail := range cvmRecycleMap {
		for _, ip := range detail.EipList {
			if ip.Err != nil {
				break
			}
			rebindErr := e.AssociateEip(kt, detail.Vendor, ip.EipID, detail.CvmID, ip.NicID, detail.AccountID)
			if rebindErr != nil {
				logs.Errorf("failed to rebind eip(%s) to cvm(%s), err: %v, rid: %s",
					ip.EipID, detail.CvmID, rebindErr, kt.Rid)
				return rebindErr
			}
		}

	}
	return nil
}

// DeleteEip 删除指定eip
func (e *eip) DeleteEip(kt *kit.Kit, vendor enumor.Vendor, eipId string) (err error) {
	deleteReq := &hcproto.EipDeleteReq{EipID: eipId}
	switch vendor {
	case enumor.TCloud:
		err = e.client.HCService().TCloud.Eip.DeleteEip(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Aws:
		err = e.client.HCService().Aws.Eip.DeleteEip(kt.Ctx, kt.Header(), deleteReq)
	case enumor.HuaWei:
		err = e.client.HCService().HuaWei.Eip.DeleteEip(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Gcp:
		err = e.client.HCService().Gcp.Eip.DeleteEip(kt.Ctx, kt.Header(), deleteReq)
	case enumor.Azure:
		err = e.client.HCService().Azure.Eip.DeleteEip(kt.Ctx, kt.Header(), deleteReq)
	default:
		err = errf.NewFromErr(errf.InvalidParameter, fmt.Errorf("no support vendor: %s", vendor))
	}
	return err
}
