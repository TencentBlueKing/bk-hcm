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

package cvm

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	disk "hcm/cmd/hc-service/logics/sync/disk"
	"hcm/cmd/hc-service/logics/sync/eip"
	networkinterface "hcm/cmd/hc-service/logics/sync/network-interface"
	securitygroup "hcm/cmd/hc-service/logics/sync/security-group"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/api/data-service/cloud"
	dataproto "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service"
	protocvm "hcm/pkg/api/hc-service/cvm"
	protodisk "hcm/pkg/api/hc-service/disk"
	protoeip "hcm/pkg/api/hc-service/eip"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// GetHcCvmDatas get cvm datas from hc
func GetHcCvmDatas(kt *kit.Kit, req *protocvm.CvmSyncReq,
	dataCli *dataservice.Client) (map[string]corecvm.BaseCvm, error) {

	dsMap := make(map[string]corecvm.BaseCvm)

	start := 0
	for {
		dataReq := &dataproto.CvmListReq{
			Filter: &filter.Expression{
				Op: filter.And,
				Rules: []filter.RuleFactory{
					&filter.AtomRule{
						Field: "account_id",
						Op:    filter.Equal.Factory(),
						Value: req.AccountID,
					},
					&filter.AtomRule{
						Field: "region",
						Op:    filter.Equal.Factory(),
						Value: req.Region,
					},
				},
			},
			Page: &core.BasePage{
				Start: uint32(start),
				Limit: core.DefaultMaxPageLimit,
			},
		}

		if len(req.CloudIDs) > 0 {
			filter := filter.AtomRule{Field: "cloud_id", Op: filter.In.Factory(), Value: req.CloudIDs}
			dataReq.Filter.Rules = append(dataReq.Filter.Rules, filter)
		}

		results, err := dataCli.Global.Cvm.ListCvm(kt.Ctx, kt.Header(), dataReq)
		if err != nil {
			logs.Errorf("from data-service list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return dsMap, err
		}

		if len(results.Details) > 0 {
			for _, detail := range results.Details {
				dsMap[detail.CloudID] = detail
			}
		}

		start += len(results.Details)
		if uint(len(results.Details)) < dataReq.Page.Limit {
			break
		}
	}

	return dsMap, nil
}

func queryVpcID(kt *kit.Kit, dataCli *dataservice.Client, vpcCloudID string) (
	string, int64, error) {

	req := &core.ListReq{
		Filter: tools.EqualExpression("cloud_id", vpcCloudID),
		Page:   core.DefaultBasePage,
		Fields: []string{"id"},
	}
	vpcResult, err := dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("list vpc failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return "", 0, err
	}

	if len(vpcResult.Details) != 1 {
		return "", 0, errf.Newf(errf.RecordNotFound, "vpc: %s not found", vpcCloudID)
	}

	return vpcResult.Details[0].ID, vpcResult.Details[0].BkCloudID, nil
}

func querySubnetIDs(kt *kit.Kit, dataCli *dataservice.Client, subnetCloudIDs []string) (
	[]string, error) {

	req := &core.ListReq{
		Filter: tools.ContainersExpression("cloud_id", subnetCloudIDs),
		Page:   core.DefaultBasePage,
		Fields: []string{"id"},
	}
	subnetResult, err := dataCli.Global.Subnet.List(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("list subnet failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return nil, err
	}

	if len(subnetResult.Details) != 1 {
		return nil, errf.Newf(errf.RecordNotFound, "subnet: %v not found", subnetCloudIDs)
	}

	subnetIDs := make([]string, 0)
	for _, v := range subnetResult.Details {
		subnetIDs = append(subnetIDs, v.ID)
	}
	return subnetIDs, nil
}

func getSGHcIDs(kt *kit.Kit, req *protocvm.OperateSyncReq, dataCli *dataservice.Client,
	cloudSGMap map[string]*CVMOperateSync) error {

	cvmIDs := make([]string, 0)
	sgCloudIDs := make([]string, 0)
	for _, id := range cloudSGMap {
		sgCloudIDs = append(sgCloudIDs, id.RelID)
		cvmIDs = append(cvmIDs, id.InstanceID)
	}

	sgReq := &hcservice.SecurityGroupSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  sgCloudIDs,
	}

	sgDatas, err := securitygroup.GetDatasFromDSForSecurityGroupSync(kt, sgReq, dataCli)
	if err != nil {
		logs.Errorf("request get sg from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmReq := &protocvm.CvmSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cvmIDs,
	}

	cvmDatas, err := GetHcCvmDatas(kt, cvmReq, dataCli)
	if err != nil {
		logs.Errorf("request get cvm from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, cloudSG := range cloudSGMap {
		if _, ok := sgDatas[cloudSG.RelID]; ok {
			cloudSG.HCRelID = sgDatas[cloudSG.RelID].HcSecurityGroup.ID
		} else {
			return fmt.Errorf("security group: %s not found", cloudSG.RelID)
		}

		if _, ok := cvmDatas[cloudSG.InstanceID]; ok {
			cloudSG.HCInstanceID = cvmDatas[cloudSG.InstanceID].ID
		} else {
			return fmt.Errorf("cvm: %s not found", cloudSG.InstanceID)
		}
	}

	return nil
}

func getDiskHcIDs(kt *kit.Kit, req *protocvm.OperateSyncReq, dataCli *dataservice.Client,
	cloudDiskMap map[string]*CVMOperateSync) error {

	cvmIDs := make([]string, 0)
	diskCloudIDs := make([]string, 0)
	for _, id := range cloudDiskMap {
		diskCloudIDs = append(diskCloudIDs, id.RelID)
		cvmIDs = append(cvmIDs, id.InstanceID)
	}

	diskReq := &protodisk.DiskSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  diskCloudIDs,
	}

	diskDatas, err := disk.GetDatasFromDSForDiskSync(kt, diskReq, dataCli)
	if err != nil {
		logs.Errorf("request get disk from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmReq := &protocvm.CvmSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cvmIDs,
	}

	cvmDatas, err := GetHcCvmDatas(kt, cvmReq, dataCli)
	if err != nil {
		logs.Errorf("request get cvm from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, cloudDisk := range cloudDiskMap {
		if _, ok := diskDatas[cloudDisk.RelID]; ok {
			cloudDisk.HCRelID = diskDatas[cloudDisk.RelID].HcDisk.ID
		} else {
			return fmt.Errorf("disk: [%s] not found", cloudDisk.RelID)
		}

		if _, ok := cvmDatas[cloudDisk.InstanceID]; ok {
			cloudDisk.HCInstanceID = cvmDatas[cloudDisk.InstanceID].ID
		} else {
			return fmt.Errorf("cvm: %s not found", cloudDisk.InstanceID)
		}
	}

	return nil
}

func getEipHcIDs(kt *kit.Kit, req *protocvm.OperateSyncReq, dataCli *dataservice.Client,
	cloudEipMap map[string]*CVMOperateSync) error {

	cvmIDs := make([]string, 0)
	eipCloudIDs := make([]string, 0)
	for _, id := range cloudEipMap {
		eipCloudIDs = append(eipCloudIDs, id.RelID)
		cvmIDs = append(cvmIDs, id.InstanceID)
	}

	eipReq := &protoeip.EipSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  eipCloudIDs,
	}

	eipDatas, err := eip.GetHcEipDatas(kt, eipReq, dataCli)
	if err != nil {
		logs.Errorf("request get eip from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmReq := &protocvm.CvmSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cvmIDs,
	}

	cvmDatas, err := GetHcCvmDatas(kt, cvmReq, dataCli)
	if err != nil {
		logs.Errorf("request get cvm from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, cloudEip := range cloudEipMap {
		if _, ok := eipDatas[cloudEip.RelID]; ok {
			cloudEip.HCRelID = eipDatas[cloudEip.RelID].ID
		} else {
			return fmt.Errorf("eip: %s not found", cloudEip.RelID)
		}

		if _, ok := cvmDatas[cloudEip.InstanceID]; ok {
			cloudEip.HCInstanceID = cvmDatas[cloudEip.InstanceID].ID
		} else {
			return fmt.Errorf("cvm: %s not found", cloudEip.InstanceID)
		}
	}

	return nil
}

func getNetworkInterfaceHcIDs(kt *kit.Kit, req *protocvm.OperateSyncReq, dataCli *dataservice.Client,
	cloudNetworkInterfaceMap map[string]*CVMOperateSync) error {

	cvmIDs := make([]string, 0)
	networkInterfaceCloudIDs := make([]string, 0)
	for _, id := range cloudNetworkInterfaceMap {
		networkInterfaceCloudIDs = append(networkInterfaceCloudIDs, id.RelID)
		cvmIDs = append(cvmIDs, id.InstanceID)
	}

	networkInterfaceReq := &protocvm.OperateSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  networkInterfaceCloudIDs,
	}
	networkInterfaceDatas, err := networkinterface.GetHcNetworkInterfaceDatas(kt, networkInterfaceReq, dataCli)
	if err != nil {
		logs.Errorf("request get networkinterface from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	cvmReq := &protocvm.CvmSyncReq{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cvmIDs,
	}
	cvmDatas, err := GetHcCvmDatas(kt, cvmReq, dataCli)
	if err != nil {
		logs.Errorf("request get cvm from hc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, cloudNetworkInterface := range cloudNetworkInterfaceMap {
		if _, ok := networkInterfaceDatas[cloudNetworkInterface.RelID]; ok {
			cloudNetworkInterface.HCRelID = networkInterfaceDatas[cloudNetworkInterface.RelID].ID
		} else {
			return fmt.Errorf("network_interface: %s not found", cloudNetworkInterface.RelID)
		}

		if _, ok := cvmDatas[cloudNetworkInterface.InstanceID]; ok {
			cloudNetworkInterface.HCInstanceID = cvmDatas[cloudNetworkInterface.InstanceID].ID
		} else {
			return fmt.Errorf("cvm: %s not found", cloudNetworkInterface.InstanceID)
		}
	}

	return nil
}

func syncCvmDelete(kt *kit.Kit, deleteCloudIDs []string, dataCli *dataservice.Client) error {
	batchDeleteReq := &cloud.CvmBatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", deleteCloudIDs),
	}
	if err := dataCli.Global.Cvm.BatchDeleteCvm(kt.Ctx, kt.Header(), batchDeleteReq); err != nil {
		logs.Errorf("request dataservice delete tcloud cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func getCVMRelID(relID string, instanceID string) string {
	flag := relID + instanceID
	h := md5.New()
	h.Write([]byte(flag))
	return hex.EncodeToString(h.Sum(nil))
}

func changCloudMapToHcMap(cloudMap map[string]*CVMOperateSync) map[string]*CVMOperateSync {
	hcMap := make(map[string]*CVMOperateSync)
	for _, v := range cloudMap {
		id := getCVMRelID(v.HCRelID, v.HCInstanceID)
		hcMap[id] = v
	}
	return hcMap
}
