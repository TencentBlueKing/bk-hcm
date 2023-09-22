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
	"hcm/pkg/api/core"
	networkinterface "hcm/pkg/api/core/cloud/network-interface"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	dataeip "hcm/pkg/api/data-service/cloud/eip"
	hcproto "hcm/pkg/api/hc-service/eip"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// Interface define eip interface.
type Interface interface {
	AssociateEip(kt *kit.Kit, vendor enumor.Vendor, eipID, cvmID, nicID, accountID string) error
	DisassociateEip(kt *kit.Kit, vendor enumor.Vendor, eipID, cvmID, nicID, accountID string) error

	GetEipByCvm(kt *kit.Kit, cvmIds []string) (map[string]string, map[string]*dataeip.EipResult, error)
	BatchDisassociateEip(kt *kit.Kit, eipCvmMap map[string]string, eipMap map[string]*dataeip.EipResult) (
		*core.BatchOperateAllResult, error)
	BatchReAssociateEip(kt *kit.Kit, eipCvmMap map[string]string, eipDetailMap map[string]*dataeip.EipResult) (
		*core.BatchOperateAllResult, error)
	BatchDisassociateWithRollback(kt *kit.Kit, cvmIds []string) (batchResult, BatchRollBackFunc, error)
	DeleteEip(kt *kit.Kit, vendor enumor.Vendor, eipId string) error
}

// BatchRollBackFunc 批量操作回滚操作
type BatchRollBackFunc func(kt *kit.Kit, rollbackIds []string) (*core.BatchOperateAllResult, error)

type batchResult struct {
	SucceedResCvm map[string]string
	FailedCvm     map[string]error
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

// GetEipByCvm 查找主机和eip的关联，返回eip到cvmID的对应关系和eip为key的map[eipID]EipResult。
// 和 BatchDisassociateEip 一起组装出批量回收cvm下eip的接口
func (e *eip) GetEipByCvm(kt *kit.Kit, cvmIds []string) (map[string]string, map[string]*dataeip.EipResult, error) {
	// list eip and cvm relation
	relReq := &cloud.EipCvmRelListReq{
		Filter: tools.ContainersExpression("cvm_id", cvmIds),
		Page:   core.NewDefaultBasePage(),
	}
	cvmEips := make(map[string][]string)
	relRes, err := e.client.DataService().Global.ListEipCvmRel(kt.Ctx, kt.Header(), relReq)
	if err != nil {
		logs.Errorf("fail to ListEipCvmRel, cvm_id: %v,err: %v, rid: %s", cvmIds, err, kt.Rid)
		return nil, nil, err
	}

	if len(relRes.Details) == 0 {
		return make(map[string]string), make(map[string]*dataeip.EipResult), nil
	}

	eipCvmMap := make(map[string]string)
	eipIDs := make([]string, 0, len(relRes.Details))
	for _, detail := range relRes.Details {
		eipCvmMap[detail.EipID] = detail.CvmID
		eipList := cvmEips[detail.CvmID]
		cvmEips[detail.CvmID] = append(eipList, detail.EipID)
		eipIDs = append(eipIDs, detail.EipID)
	}

	// list eip
	eipReq := &dataeip.EipListReq{
		Filter: tools.ContainersExpression("id", eipIDs),
		Page:   core.NewDefaultBasePage(),
	}
	eipRes, err := e.client.DataService().Global.ListEip(kt.Ctx, kt.Header(), eipReq)
	if err != nil {
		logs.Errorf("fail to ListEip, eip_id: %v,err: %v, rid: %s", eipIDs, err, kt.Rid)
		return nil, nil, err
	}

	// list network interface attached with eip
	eipMap := make(map[string]*dataeip.EipResult)
	publicIPs := make([]string, 0, len(eipRes.Details))
	for _, detail := range eipRes.Details {
		eipMap[detail.ID] = detail
		publicIPs = append(publicIPs, detail.PublicIp)
	}

	nicMap, err := e.listNicByCvmAndEip(kt, cvmIds, publicIPs)
	if err != nil {
		logs.Errorf("fail to listNicByCvmAndEip, cvmIds: %v, publicIPs: %s, err: %v, rid: %s",
			cvmIds, publicIPs, eipIDs, err, kt.Rid)
		return nil, nil, err
	}

	for eipID, cvmID := range eipCvmMap {
		ip := eipMap[eipID].PublicIp
		for _, nic := range nicMap[cvmID] {
			if slice.IsItemInSlice(nic.PublicIPv4, ip) || slice.IsItemInSlice(nic.PublicIPv6, ip) {
				eipMap[eipID].InstanceID = converter.ValToPtr(nic.ID)
			}
		}
	}

	return eipCvmMap, eipMap, nil
}

// TODO save eip and nic relation in db
func (e *eip) listNicByCvmAndEip(kt *kit.Kit, ids []string, publicIPs []string) (
	map[string][]networkinterface.BaseNetworkInterface, error) {

	nicRelRes, err := e.client.DataService().Global.NetworkInterfaceCvmRel.List(kt.Ctx, kt.Header(),
		&core.ListReq{Filter: tools.ContainersExpression("cvm_id", ids), Page: core.NewDefaultBasePage()})
	if err != nil {
		return nil, err
	}

	if len(nicRelRes.Details) == 0 {
		return make(map[string][]networkinterface.BaseNetworkInterface), nil
	}

	nicIDs := make([]string, len(nicRelRes.Details))
	nicCvmMap := make(map[string]string)
	for idx, rel := range nicRelRes.Details {
		nicIDs[idx] = rel.NetworkInterfaceID
		nicCvmMap[rel.NetworkInterfaceID] = rel.CvmID
	}

	nicRes, err := e.client.DataService().Global.NetworkInterface.List(kt.Ctx, kt.Header(),
		&core.ListReq{Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{tools.ContainersExpression("id", nicIDs),
				&filter.Expression{Op: filter.Or, Rules: []filter.RuleFactory{
					filter.AtomRule{Field: "public_ipv4", Op: filter.JSONOverlaps.Factory(), Value: publicIPs},
					filter.AtomRule{Field: "public_ipv6", Op: filter.JSONOverlaps.Factory(), Value: publicIPs},
				}},
			},
		}, Page: core.NewDefaultBasePage()})
	if err != nil {
		return nil, err
	}

	nicMap := make(map[string][]networkinterface.BaseNetworkInterface)
	for _, nic := range nicRes.Details {
		nicMap[nicCvmMap[nic.ID]] = append(nicMap[nicCvmMap[nic.ID]], nic)
	}
	return nicMap, nil
}

// BatchDisassociateEip 解绑cvm上的eip，全部成功时err是nil，否则需要检查操作结果
func (e *eip) BatchDisassociateEip(kt *kit.Kit, eipCvmMap map[string]string, eipMap map[string]*dataeip.EipResult) (
	*core.BatchOperateAllResult, error) {
	results := &core.BatchOperateAllResult{}
	for eipId, cvmID := range eipCvmMap {
		eipInstance, exists := eipMap[eipId]
		if !exists {
			return nil, errf.Newf(errf.InvalidParameter, "eipInstance %s not exists", eipId)
		}
		vendor := enumor.Vendor(eipInstance.Vendor)

		// TODO get nic eipId by eipInstance InstanceType
		var nicID string
		switch vendor {
		case enumor.Azure, enumor.Gcp, enumor.HuaWei:
			nicID = converter.PtrToVal(eipInstance.InstanceID)
		}

		err := e.DisassociateEip(kt, vendor, eipId, cvmID, nicID, eipInstance.AccountID)
		if err != nil {
			logs.Errorf("disassociate eipInstance %s failed, err: %v, cvm: %s, nic: %s, rid: %s",
				eipId, err, cvmID, nicID, kt.Rid)
			results.Failed = append(results.Failed, core.FailedInfo{ID: eipId, Error: err})
		} else {
			results.Succeeded = append(results.Succeeded, eipId)
		}
	}
	if len(results.Failed) != 0 {
		return results, errf.New(errf.PartialFailed, "")
	}

	return nil, nil
}

// BatchReAssociateEip 重新绑定Eip到cvm，作为上面 BatchDisassociateEip 的回滚操作
func (e *eip) BatchReAssociateEip(kt *kit.Kit, eipCvmMap map[string]string,
	eipDetailMap map[string]*dataeip.EipResult) (*core.BatchOperateAllResult, error) {

	results := &core.BatchOperateAllResult{}
	for eipId, cvmID := range eipCvmMap {
		eip := eipDetailMap[eipId]
		vendor := enumor.Vendor(eip.Vendor)

		var nicID string
		switch vendor {
		case enumor.Azure, enumor.Gcp, enumor.HuaWei:
			nicID = converter.PtrToVal(eip.InstanceID)
		}

		err := e.AssociateEip(kt, vendor, eipId, cvmID, nicID, eip.AccountID)
		if err != nil {
			logs.Errorf("asst eip %s failed, err: %v, cvm: %s, nic: %s, rid: %s", eipId, err, cvmID, nicID, kt.Rid)
			results.Failed = append(results.Failed, core.FailedInfo{ID: eipId, Error: err})
		}
	}
	if len(results.Failed) != 0 {
		return results, errf.New(errf.PartialFailed, "")
	}
	return nil, nil
}

func allSuccessRollback(kt *kit.Kit, rollbackIds []string) (*core.BatchOperateAllResult, error) {
	return &core.BatchOperateAllResult{Succeeded: rollbackIds}, nil
}

// BatchDisassociateWithRollback 批量解绑eip，
// 1. 并提供回滚操作，用户自行决定是否回滚。
// 2. 不同cvm下的eip失败不互相影响；同一个cvm下的eip只会失败一次，失败后不在处理通cvm下的eip
func (e *eip) BatchDisassociateWithRollback(kt *kit.Kit,
	detachEipCvmIDs []string) (batchResult, BatchRollBackFunc, error) {

	operationResult := batchResult{map[string]string{}, make(map[string]error)}
	// 	1. 获取eip 信息
	eipCvmMap, eipMap, err := e.GetEipByCvm(kt, detachEipCvmIDs)
	if err != nil {
		for _, cvmID := range detachEipCvmIDs {
			operationResult.FailedCvm[cvmID] = err
		}
		// 还没操作所以直接返回全部回滚成功的回滚函数
		return operationResult, allSuccessRollback, err
	}

	rollback := func(kt *kit.Kit, rbCvmIds []string) (*core.BatchOperateAllResult, error) {

		if len(detachEipCvmIDs) == 0 {
			return nil, nil
		}
		logs.V(3).Infof("rollback for BatchDisassociateEip, rollback cvm ids: %v, rid:%s", rbCvmIds, kt.Rid)
		rbCvmMap := converter.StringSliceToMap(rbCvmIds)
		rbEipCvmMap := map[string]string{}
		for eipId, cvmId := range eipCvmMap {
			if _, ok := rbCvmMap[cvmId]; ok {
				rbEipCvmMap[eipId] = cvmId
			}
		}
		reAssociateResult, err := e.BatchReAssociateEip(kt, rbEipCvmMap, eipMap)
		if err != nil {
			logs.Errorf("rollback failed, err: %v, rollback CvmIds: %v, rid: %s", err, rbCvmIds, kt.Rid)
			return reAssociateResult, err
		}
		return reAssociateResult, nil
	}

	// 	2.尝试解绑
	for eipId, cvmID := range eipCvmMap {
		if operationResult.FailedCvm[cvmID] != nil {
			// we only fail once
			continue
		}
		eipInst := eipMap[eipId]

		vendor := enumor.Vendor(eipInst.Vendor)

		var nicID string
		switch vendor {
		case enumor.Azure, enumor.Gcp, enumor.HuaWei:
			nicID = converter.PtrToVal(eipInst.InstanceID)
		}

		err := e.DisassociateEip(kt, vendor, eipId, cvmID, nicID, eipInst.AccountID)
		if err != nil {
			logs.Errorf("disassociate eipInstance %s failed, err: %v, cvm: %s, nic: %s, rid: %s",
				eipId, err, cvmID, nicID, kt.Rid)
			operationResult.FailedCvm[cvmID] = err
		} else {
			operationResult.SucceedResCvm[eipId] = cvmID
		}
	}
	return operationResult, rollback, err

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
