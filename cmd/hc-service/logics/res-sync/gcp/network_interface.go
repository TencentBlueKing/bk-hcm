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

package gcp

import (
	"fmt"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typesni "hcm/pkg/adaptor/types/network-interface"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	dataproto "hcm/pkg/api/data-service"
	datacloud "hcm/pkg/api/data-service/cloud"
	datani "hcm/pkg/api/data-service/cloud/network-interface"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/concurrence"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// SyncNIOption ...
type SyncNIOption struct {
	Region string `json:"region" validate:"required"`
	Zone   string `json:"zone" validate:"required"`
}

// Validate ...
func (opt SyncNIOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// NetworkInterface 网络接口的同步依赖主机同步，db的网络接口是通过主机关联关系查询出来的。
func (cli *client) NetworkInterface(kt *kit.Kit, params *SyncBaseParams, opt *SyncNIOption) (*SyncResult, error) {

	if err := validator.ValidateTool(params, opt); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := concurrence.BaseExec(constant.SyncConcurrencyDefaultMaxLimit, params.CloudIDs, func(param string) error {
		syncOpt := &syncNIOption{
			AccountID:  params.AccountID,
			Region:     opt.Region,
			Zone:       opt.Zone,
			CloudCvmID: param,
		}
		if _, err := cli.syncNetworkInterface(kt, syncOpt); err != nil {
			logs.ErrorDepthf(1, "[%s] account: %s cvm: %s sync network interface failed, err: %v, rid: %s", enumor.Gcp,
				params.AccountID, param, err, kt.Rid)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return new(SyncResult), nil
}

type syncNIOption struct {
	AccountID  string `json:"account_id" validate:"required"`
	Region     string `json:"region" validate:"required"`
	Zone       string `json:"zone" validate:"required"`
	CloudCvmID string `json:"cloud_cvm_id" validate:"required"`
}

// Validate ...
func (opt syncNIOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// syncNetworkInterface 同步网络接口
func (cli *client) syncNetworkInterface(kt *kit.Kit, opt *syncNIOption) (*SyncResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	networkInterfaceFromCloud, err := cli.listNetworkInterfaceFromCloud(kt, opt.Zone, opt.CloudCvmID)
	if err != nil {
		return nil, err
	}

	networkInterfaceFromDB, cvm, err := cli.listNetworkInterfaceFromDB(kt, opt)
	if err != nil {
		return nil, err
	}

	if len(networkInterfaceFromCloud) == 0 && len(networkInterfaceFromDB) == 0 {
		return new(SyncResult), nil
	}

	addSlice, updateMap, delCloudIDs := common.Diff[typesni.GcpNI, coreni.
		NetworkInterface[coreni.GcpNIExtension]](networkInterfaceFromCloud, networkInterfaceFromDB, isNIChange)

	if len(delCloudIDs) > 0 {
		if err = cli.deleteNetworkInterface(kt, delCloudIDs, opt); err != nil {
			return nil, err
		}
	}

	if len(addSlice) > 0 {
		if err = cli.createNetworkInterface(kt, opt.AccountID, cvm, addSlice); err != nil {
			return nil, err
		}
	}

	if len(updateMap) > 0 {
		if err = cli.updateNetworkInterface(kt, opt.AccountID, updateMap); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// deleteNetworkInterface 删除网络接口
func (cli *client) deleteNetworkInterface(kt *kit.Kit, delCloudIDs []string, opt *syncNIOption) error {

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(delCloudIDs) <= 0 {
		return fmt.Errorf("delete network interfaces, network interfaces id is required")
	}

	niLists, err := cli.listNetworkInterfaceFromCloud(kt, opt.Zone, opt.CloudCvmID)
	if err != nil {
		return err
	}

	delMap := converter.StringSliceToMap(delCloudIDs)
	for _, one := range niLists {
		if _, exist := delMap[*one.CloudID]; exist {
			logs.Errorf("[%s] validate networkInterface not exist failed, before delete, opt: %v, failed_count: %d, "+
				"rid: %s", enumor.Gcp, opt, len(delCloudIDs), kt.Rid)
			return fmt.Errorf("validate networkInterface not exist failed, before delete")
		}
	}

	deleteReq := &dataproto.BatchDeleteReq{
		Filter: tools.ContainersExpression("cloud_id", delCloudIDs),
	}
	if err = cli.dbCli.Global.NetworkInterface.BatchDelete(kt.Ctx, kt.Header(), deleteReq); err != nil {
		logs.Errorf("[%s] request dataservice to batch delete networkInterface failed, err: %v, rid: %s", enumor.Gcp,
			err, kt.Rid)
		return err
	}

	logs.V(3).Infof("[%s] sync network interface to delete ni success, accountID: %s, count: %d, rid: %s",
		enumor.Gcp, opt.AccountID, len(delCloudIDs), kt.Rid)

	return nil
}

// updateNetworkInterface 更新网络接口
func (cli *client) updateNetworkInterface(kt *kit.Kit, accountID string, updateMap map[string]typesni.GcpNI) error {

	if len(updateMap) <= 0 {
		return fmt.Errorf("update network interfaces, network interfaces is required")
	}

	if err := cli.completeNetworkInterfaceUpdateInfo(kt, updateMap); err != nil {
		return fmt.Errorf("complete network interface update info failed, err: %v", err)
	}

	lists := make([]datani.NetworkInterfaceUpdateReq[datani.GcpNICreateExt], 0)
	for id, item := range updateMap {
		tmpRes := datani.NetworkInterfaceUpdateReq[datani.GcpNICreateExt]{
			ID:            id,
			AccountID:     accountID,
			Name:          converter.PtrToVal(item.Name),
			Region:        converter.PtrToVal(item.Region),
			Zone:          converter.PtrToVal(item.Zone),
			CloudID:       converter.PtrToVal(item.CloudID),
			VpcID:         converter.PtrToVal(item.VpcID),
			CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
			SubnetID:      converter.PtrToVal(item.SubnetID),
			CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
			PrivateIPv4:   item.PrivateIPv4,
			PrivateIPv6:   item.PrivateIPv6,
			PublicIPv4:    item.PublicIPv4,
			PublicIPv6:    item.PublicIPv6,
			InstanceID:    converter.PtrToVal(item.InstanceID),
		}
		if item.Extension != nil {
			tmpRes.Extension = convertGcpNIExtension(item)
		}

		lists = append(lists, tmpRes)
	}

	updateReq := &datani.NetworkInterfaceBatchUpdateReq[datani.GcpNICreateExt]{
		NetworkInterfaces: lists,
	}
	if err := cli.dbCli.Gcp.NetworkInterface.BatchUpdate(kt.Ctx, kt.Header(), updateReq); err != nil {
		logs.Errorf("[%s] request dataservice gcp BatchUpdateNetworkInterface failed, err: %v, rid: %s",
			enumor.Gcp, err, kt.Rid)
		return err
	}

	logs.V(3).Infof("[%s] sync network interface to update ni success, accountID: %s, count: %d, rid: %s",
		enumor.Gcp, accountID, len(updateMap), kt.Rid)

	return nil
}

// completeNetworkInterfaceUpdateInfo 补全网络接口更新信息
func (cli *client) completeNetworkInterfaceUpdateInfo(kt *kit.Kit, niMap map[string]typesni.GcpNI) error {
	subnetCloudIDMap := make(map[string]struct{}, 0)
	for _, ni := range niMap {
		if ni.CloudSubnetID == nil {
			continue
		}

		subnetCloudIDMap[*ni.CloudSubnetID] = struct{}{}
	}

	subnetMap, err := cli.getSubnetMapByCloudID(kt, converter.MapKeyToStringSlice(subnetCloudIDMap))
	if err != nil {
		return err
	}

	for _, ni := range niMap {
		if ni.CloudSubnetID == nil {
			continue
		}

		cloudID := *ni.CloudSubnetID
		subnet, exist := subnetMap[cloudID]
		if !exist {
			return fmt.Errorf("subnet: %s not found", cloudID)
		}

		ni.CloudVpcID = converter.ValToPtr(subnet.CloudVpcID)
		ni.VpcID = converter.ValToPtr(subnet.VpcID)
		ni.SubnetID = converter.ValToPtr(subnet.ID)
	}

	return nil
}

// completeNetworkInterfaceCreateInfo 补全网络接口创建信息
func (cli *client) completeNetworkInterfaceCreateInfo(kt *kit.Kit, nis []typesni.GcpNI) error {
	subnetCloudIDMap := make(map[string]struct{}, 0)
	for _, ni := range nis {
		if ni.CloudSubnetID != nil {
			subnetCloudIDMap[*ni.CloudSubnetID] = struct{}{}
		}
	}

	subnetMap, err := cli.getSubnetMapByCloudID(kt, converter.MapKeyToStringSlice(subnetCloudIDMap))
	if err != nil {
		return err
	}

	for index := range nis {
		if nis[index].CloudSubnetID == nil {
			continue
		}

		cloudID := *nis[index].CloudSubnetID
		subnet, exist := subnetMap[cloudID]
		if !exist {
			return fmt.Errorf("subnet: %s not found", cloudID)
		}

		nis[index].CloudVpcID = converter.ValToPtr(subnet.CloudVpcID)
		nis[index].VpcID = converter.ValToPtr(subnet.VpcID)
		nis[index].SubnetID = converter.ValToPtr(subnet.ID)
	}

	return nil
}

// getSubnetMapByCloudID 根据云子网ID获取子网信息
func (cli *client) getSubnetMapByCloudID(kt *kit.Kit, cloudIDs []string) (map[string]cloud.BaseSubnet, error) {
	subnetMap := make(map[string]cloud.BaseSubnet, len(cloudIDs))
	split := slice.Split(cloudIDs, int(core.DefaultMaxPageLimit))
	for _, part := range split {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("cloud_id", part),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := cli.dbCli.Global.Subnet.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("[%s] list subnet failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			subnetMap[one.CloudID] = one
		}
	}
	return subnetMap, nil
}

// createNetworkInterface 创建网络接口
func (cli *client) createNetworkInterface(kt *kit.Kit, accountID string, cvm *corecvm.BaseCvm,
	addSlice []typesni.GcpNI) error {

	if len(addSlice) <= 0 {
		return fmt.Errorf("create network interfaces, network interfaces is required")
	}

	if err := cli.completeNetworkInterfaceCreateInfo(kt, addSlice); err != nil {
		return fmt.Errorf("complete network interface create info failed, err: %v", err)
	}

	lists := make([]datani.NetworkInterfaceReq[datani.GcpNICreateExt], 0)

	for _, item := range addSlice {
		tmpRes := datani.NetworkInterfaceReq[datani.GcpNICreateExt]{
			AccountID:     accountID,
			Vendor:        string(enumor.Gcp),
			Name:          converter.PtrToVal(item.Name),
			Region:        converter.PtrToVal(item.Region),
			Zone:          converter.PtrToVal(item.Zone),
			CloudID:       converter.PtrToVal(item.CloudID),
			VpcID:         converter.PtrToVal(item.VpcID),
			CloudVpcID:    converter.PtrToVal(item.CloudVpcID),
			SubnetID:      converter.PtrToVal(item.SubnetID),
			CloudSubnetID: converter.PtrToVal(item.CloudSubnetID),
			PrivateIPv4:   item.PrivateIPv4,
			PrivateIPv6:   item.PrivateIPv6,
			PublicIPv4:    item.PublicIPv4,
			PublicIPv6:    item.PublicIPv6,
			InstanceID:    converter.PtrToVal(item.InstanceID),
			BkBizID:       cvm.BkBizID,
		}
		if item.Extension != nil {
			if item.Extension != nil {
				tmpRes.Extension = convertGcpNIExtension(item)
			}
		}

		lists = append(lists, tmpRes)
	}

	createReq := &datani.NetworkInterfaceBatchCreateReq[datani.GcpNICreateExt]{
		NetworkInterfaces: lists,
	}
	createResult, err := cli.dbCli.Gcp.NetworkInterface.BatchCreate(kt.Ctx, kt.Header(), createReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create gcp networkInterface failed, err: %v, rid: %s",
			enumor.Gcp, err, kt.Rid)
		return err
	}

	relLists := make([]datacloud.NetworkInterfaceCvmRelCreateReq, 0)
	for _, id := range createResult.IDs {
		relLists = append(relLists, datacloud.NetworkInterfaceCvmRelCreateReq{
			NetworkInterfaceID: id,
			CvmID:              cvm.ID,
		})
	}

	createRelReq := &datacloud.NetworkInterfaceCvmRelBatchCreateReq{
		Rels: relLists,
	}
	err = cli.dbCli.Global.NetworkInterfaceCvmRel.BatchCreateNetworkCvmRels(kt.Ctx, kt.Header(), createRelReq)
	if err != nil {
		logs.Errorf("[%s] request dataservice to create cvm_ni_rel failed, err: %v, rid: %s",
			enumor.Gcp, err, kt.Rid)
		return err
	}

	logs.V(3).Infof("[%s] sync network interface to create ni success, accountID: %s, count: %d, rid: %s",
		enumor.Gcp, accountID, len(addSlice), kt.Rid)

	return nil
}

func convertGcpNIExtension(item typesni.GcpNI) *datani.GcpNICreateExt {

	extension := &datani.GcpNICreateExt{
		CanIpForward:   item.Extension.CanIpForward,
		Status:         item.Extension.Status,
		StackType:      item.Extension.StackType,
		VpcSelfLink:    item.Extension.VpcSelfLink,
		SubnetSelfLink: item.Extension.SubnetSelfLink,
	}
	// 网卡私网IP信息列表
	var tmpAccConfigs []*datani.AccessConfig
	for _, accConfigItem := range item.Extension.AccessConfigs {
		tmpAccConfigs = append(tmpAccConfigs, &datani.AccessConfig{
			Name:        accConfigItem.Name,
			NatIP:       accConfigItem.NatIP,
			NetworkTier: accConfigItem.NetworkTier,
			Type:        accConfigItem.Type,
		})
	}
	extension.AccessConfigs = tmpAccConfigs
	return extension
}

// listNetworkInterfaceFromCloud 从云上获取网络接口列表
func (cli *client) listNetworkInterfaceFromCloud(kt *kit.Kit, zone, cloudCvmID string) ([]typesni.GcpNI, error) {

	listOpt := &typesni.GcpListByCvmIDOption{
		Zone:        zone,
		CloudCvmIDs: []string{cloudCvmID},
	}
	result, err := cli.cloudCli.ListNetworkInterfaceByCvmID(kt, listOpt)
	if err != nil {
		logs.Errorf("[%s] list networkInterface from cloud failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("cvm: %s not found", cloudCvmID)
	}

	return result[cloudCvmID], nil
}

// listNetworkInterfaceFromDB 从数据库获取网络接口列表
func (cli *client) listNetworkInterfaceFromDB(kt *kit.Kit, opt *syncNIOption) (
	[]coreni.NetworkInterface[coreni.GcpNIExtension], *corecvm.BaseCvm, error) {

	if err := opt.Validate(); err != nil {
		return nil, nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listCvmReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "cloud_id", Op: filter.Equal.Factory(), Value: opt.CloudCvmID},
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: opt.AccountID},
				&filter.AtomRule{Field: "zone", Op: filter.Equal.Factory(), Value: opt.Zone},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	cvmResult, err := cli.dbCli.Global.Cvm.ListCvm(kt, listCvmReq)
	if err != nil {
		logs.Errorf("[%s] list cvm failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, nil, err
	}

	if len(cvmResult.Details) == 0 {
		return nil, nil, fmt.Errorf("cvm: %s not found", opt.CloudCvmID)
	}

	listRelReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "cvm_id", Op: filter.Equal.Factory(), Value: cvmResult.Details[0].ID},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	relResult, err := cli.dbCli.Global.NetworkInterfaceCvmRel.ListNetworkCvmRels(kt, listRelReq)
	if err != nil {
		logs.Errorf("[%s] list ni_cvm_rel failed, err: %v, rid: %s", enumor.Gcp, err, kt.Rid)
		return nil, nil, err
	}

	if len(relResult.Details) == 0 {
		return make([]coreni.NetworkInterface[coreni.GcpNIExtension], 0), &cvmResult.Details[0], nil
	}

	niIDs := make([]string, 0, len(relResult.Details))
	for _, one := range relResult.Details {
		niIDs = append(niIDs, one.NetworkInterfaceID)
	}

	req := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: opt.AccountID},
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: niIDs},
				&filter.AtomRule{Field: "region", Op: filter.Equal.Factory(), Value: opt.Region},
				&filter.AtomRule{Field: "zone", Op: filter.Equal.Factory(), Value: opt.Zone},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.dbCli.Gcp.NetworkInterface.ListNetworkInterfaceExt(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("[%s] list networkInterface from db failed, err: %v, account: %s, req: %v, rid: %s", enumor.Gcp,
			err, opt.AccountID, req, kt.Rid)
		return nil, nil, err
	}

	return result.Details, &cvmResult.Details[0], nil
}

// isNIChange 判断网络接口是否有变更
func isNIChange(item typesni.GcpNI, dbInfo coreni.NetworkInterface[coreni.GcpNIExtension]) bool {

	if dbInfo.Name != converter.PtrToVal(item.Name) || dbInfo.Region != converter.PtrToVal(item.Region) ||
		dbInfo.Zone != converter.PtrToVal(item.Zone) || dbInfo.CloudID != converter.PtrToVal(item.CloudID) ||
		dbInfo.CloudSubnetID != converter.PtrToVal(item.CloudSubnetID) ||
		dbInfo.InstanceID != converter.PtrToVal(item.InstanceID) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PrivateIPv4, dbInfo.PrivateIPv4) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PrivateIPv6, dbInfo.PrivateIPv6) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PublicIPv4, dbInfo.PublicIPv4) {
		return true
	}
	if !assert.IsStringSliceEqual(item.PublicIPv6, dbInfo.PublicIPv6) {
		return true
	}
	if dbInfo.Extension == nil {
		return false
	}
	extRet := isNIExtChange(item, dbInfo)
	if extRet {
		return true
	}

	return false
}

// isNIExtChange 判断网络接口扩展信息是否有变更
func isNIExtChange(item typesni.GcpNI, dbInfo coreni.NetworkInterface[coreni.GcpNIExtension]) bool {
	if item.Extension.VpcSelfLink != dbInfo.Extension.VpcSelfLink {
		return true
	}
	if item.Extension.SubnetSelfLink != dbInfo.Extension.SubnetSelfLink {
		return true
	}
	if item.Extension.CanIpForward != dbInfo.Extension.CanIpForward {
		return true
	}
	if item.Extension.Status != dbInfo.Extension.Status {
		return true
	}
	if item.Extension.StackType != dbInfo.Extension.StackType {
		return true
	}

	for _, remote := range item.Extension.AccessConfigs {
		for _, db := range dbInfo.Extension.AccessConfigs {
			if remote.Name != db.Name || remote.NatIP != db.NatIP || remote.NetworkTier != db.NetworkTier ||
				remote.Type != db.Type {
				return true
			}
		}
	}
	return false
}
