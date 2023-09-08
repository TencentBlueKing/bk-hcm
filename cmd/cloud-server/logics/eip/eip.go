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
	protoaudit "hcm/pkg/api/data-service/audit"
	hcproto "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Interface define eip interface.
type Interface interface {
	AssociateEip(kt *kit.Kit, vendor enumor.Vendor, eipID, cvmID, nicID, accountID string) error
	DisassociateEip(kt *kit.Kit, vendor enumor.Vendor, eipID, cvmID, nicID, accountID string) error
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
