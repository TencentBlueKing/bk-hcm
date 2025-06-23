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
	"fmt"

	logicaudit "hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/logics/disk"
	"hcm/cmd/cloud-server/logics/eip"
	logicsni "hcm/cmd/cloud-server/logics/network-interface"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// Assign 分配主机及关联资源到业务下
func Assign(kt *kit.Kit, cli *dataservice.Client, cvms []AssignedCvmInfo) error {
	if len(cvms) == 0 {
		return fmt.Errorf("cvms is required")
	}

	ids := make([]string, 0, len(cvms))
	bizIDCvmIDsMap := make(map[int64][]string)
	for _, cvmInfo := range cvms {
		ids = append(ids, cvmInfo.CvmID)
		bizIDCvmIDsMap[cvmInfo.BkBizID] = append(bizIDCvmIDsMap[cvmInfo.BkBizID], cvmInfo.CvmID)
	}
	// 校验主机信息
	if err := ValidateBeforeAssign(kt, cli, ids); err != nil {
		return err
	}

	// 获取主机关联资源
	cvmIDEipIDsMap, cvmIDDiskIDsMap, cvmIDNiIDsMap, err := GetCvmRelResIDs(kt, cli, ids)
	if err != nil {
		return err
	}

	for bizID, cvmIDs := range bizIDCvmIDsMap {
		eipIDs := make([]string, 0)
		diskIDs := make([]string, 0)
		niIDs := make([]string, 0)
		for _, cvmID := range cvmIDs {
			eipIDs = append(eipIDs, cvmIDEipIDsMap[cvmID]...)
			diskIDs = append(diskIDs, cvmIDDiskIDsMap[cvmID]...)
			niIDs = append(niIDs, cvmIDNiIDsMap[cvmID]...)
		}

		// 校验主机关联资源信息
		if err := ValidateCvmRelResBeforeAssign(kt, cli, bizID, eipIDs, diskIDs, niIDs); err != nil {
			return err
		}

		// create assign audit
		audit := logicaudit.NewAudit(cli)
		if err := audit.ResBizAssignAudit(kt, enumor.CvmAuditResType, cvmIDs, bizID); err != nil {
			logs.Errorf("create assign cvm audit failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		// 分配主机关联资源
		if err := AssignCvmRelRes(kt, cli, eipIDs, diskIDs, niIDs, bizID); err != nil {
			return err
		}
	}

	// 分配主机
	for _, batch := range slice.Split(cvms, constant.BatchOperationMaxLimit) {
		updateCvms := make([]dataproto.CvmCommonInfoBatchUpdateData, 0, len(batch))
		for _, cvmInfo := range batch {
			updateCvms = append(updateCvms, dataproto.CvmCommonInfoBatchUpdateData{
				ID:        cvmInfo.CvmID,
				BkBizID:   converter.ValToPtr(cvmInfo.BkBizID),
				BkCloudID: converter.ValToPtr(cvmInfo.BkCloudID),
			})
		}
		update := &dataproto.CvmCommonInfoBatchUpdateReq{Cvms: updateCvms}
		if err := cli.Global.Cvm.BatchUpdateCvmCommonInfo(kt, update); err != nil {
			logs.Errorf("batch update cvm common info failed, err: %v, req: %v, rid: %s", err, update, kt.Rid)
			return err
		}
	}

	return nil
}

// AssignCvmRelRes 分配主机关联资源
func AssignCvmRelRes(kt *kit.Kit, cli *dataservice.Client, eipIDs []string,
	diskIDs []string, niIDs []string, bizID int64) error {

	if len(eipIDs) != 0 {
		if err := eip.Assign(kt, cli, eipIDs, uint64(bizID), true); err != nil {
			return err
		}
	}

	if len(diskIDs) != 0 {
		if err := disk.Assign(kt, cli, diskIDs, uint64(bizID), true); err != nil {
			return err
		}
	}

	if len(niIDs) != 0 {
		if err := logicsni.Assign(kt, cli, niIDs, bizID, true); err != nil {
			return err
		}
	}

	return nil
}

// GetCvmRelResIDs 获取主机关联资源ID列表
func GetCvmRelResIDs(kt *kit.Kit, cli *dataservice.Client, ids []string) (cvmIDEipIDsMap map[string][]string,
	cvmIDDiskIDsMap map[string][]string, cvmIDNiIDsMap map[string][]string, err error) {

	cvmIDEipIDsMap = make(map[string][]string)
	for {
		listRelReq := &core.ListReq{
			Filter: tools.ContainersExpression("cvm_id", ids),
			Page:   core.NewDefaultBasePage(),
		}
		eipResp, err := cli.Global.ListEipCvmRel(kt, listRelReq)
		if err != nil {
			logs.Errorf("list eip cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, nil, err
		}
		for _, detail := range eipResp.Details {
			cvmIDEipIDsMap[detail.CvmID] = append(cvmIDEipIDsMap[detail.CvmID], detail.EipID)
		}

		if len(eipResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		listRelReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	cvmIDDiskIDsMap = make(map[string][]string)
	for {
		listRelReq := &core.ListReq{
			Filter: tools.ContainersExpression("cvm_id", ids),
			Page:   core.NewDefaultBasePage(),
		}
		diskResp, err := cli.Global.ListDiskCvmRel(kt, listRelReq)
		if err != nil {
			logs.Errorf("list disk cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, nil, err
		}
		for _, detail := range diskResp.Details {
			cvmIDDiskIDsMap[detail.CvmID] = append(cvmIDDiskIDsMap[detail.CvmID], detail.DiskID)
		}

		if len(diskResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		listRelReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	cvmIDNiIDsMap = make(map[string][]string)
	for {
		listRelReq := &core.ListReq{
			Filter: tools.ContainersExpression("cvm_id", ids),
			Page:   core.NewDefaultBasePage(),
		}
		niResp, err := cli.Global.NetworkInterfaceCvmRel.ListNetworkCvmRels(kt, listRelReq)
		if err != nil {
			logs.Errorf("list network_interface cvm rel failed, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, nil, err
		}
		for _, detail := range niResp.Details {
			cvmIDNiIDsMap[detail.CvmID] = append(cvmIDNiIDsMap[detail.CvmID], detail.NetworkInterfaceID)
		}

		if len(niResp.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		listRelReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return
}

// ValidateCvmRelResBeforeAssign 校验主机关联资源在分配前
func ValidateCvmRelResBeforeAssign(kt *kit.Kit, cli *dataservice.Client, targetBizId int64, eipIDs []string,
	diskIDs []string, niIDs []string) error {

	if len(eipIDs) != 0 {
		if err := eip.ValidateBeforeAssign(kt, cli, targetBizId, eipIDs, true); err != nil {
			return err
		}
	}

	if len(diskIDs) != 0 {
		if err := disk.ValidateBeforeAssign(kt, cli, targetBizId, diskIDs, true); err != nil {
			return err
		}
	}

	if len(niIDs) != 0 {
		if err := logicsni.ValidateBeforeAssign(kt, cli, targetBizId, niIDs, true); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBeforeAssign 分配主机前校验主机信息
func ValidateBeforeAssign(kt *kit.Kit, cli *dataservice.Client, ids []string) error {
	listReq := &core.ListReq{
		Fields: []string{"id", "bk_biz_id"},
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: ids},
			},
		},
		Page: core.NewDefaultBasePage(),
	}
	result, err := cli.Global.Cvm.ListCvm(kt, listReq)
	if err != nil {
		logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	accountIDMap := make(map[string]struct{}, 0)
	assignedIDs := make([]string, 0)
	for _, one := range result.Details {
		accountIDMap[one.AccountID] = struct{}{}

		if one.BkBizID != constant.UnassignedBiz {
			assignedIDs = append(assignedIDs, one.ID)
		}
	}

	if len(assignedIDs) != 0 {
		return fmt.Errorf("cvm(ids=%v) already assigned", assignedIDs)
	}

	return nil
}

// AssignPreview 分配主机预览
func AssignPreview(kt *kit.Kit, cmdbCli cmdb.Client, cli *client.ClientSet, ids []string) (
	map[string][]PreviewCvmMatchResult, error) {

	// 1.查询cvm信息(云实例id、云厂商、内网IP、mac地址、账号所属业务)
	cvmInfos, err := getAssignedCvmInfo(kt, cli, ids)
	if err != nil {
		logs.Errorf("get assigned cvm info failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	// 2.获取可能匹配的cc主机
	fields := []string{"bk_host_id", "bk_cloud_id", "bk_cloud_inst_id", "bk_cloud_vendor", "bk_host_innerip", "bk_mac"}
	ccHosts, ccBizHostIDsMap, err := GetAssignedHostInfoFromCC(kt, cmdbCli, cvmInfos, fields)
	if err != nil {
		logs.Errorf("get assign host from cc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 3.根据匹配规则，返回匹配cc主机结果
	return matchAssignedCvm(cvmInfos, ccHosts, ccBizHostIDsMap)
}

func getAssignedCvmInfo(kt *kit.Kit, cli *client.ClientSet, ids []string) ([]PreviewAssignedCvmInfo, error) {
	// 1.查询cvm信息
	accountIDs := make([]string, 0)
	cvmMap := make(map[string]corecvm.BaseCvm)
	vendorCvmMap := make(map[enumor.Vendor]map[string]corecvm.BaseCvm)
	for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		listCvmReq := &core.ListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		cvmList, err := cli.DataService().Global.Cvm.ListCvm(kt, listCvmReq)
		if err != nil {
			logs.Errorf("list cvm failed, err: %v, ids: %v, rid: %s", err, batch, kt.Rid)
			return nil, err
		}
		for _, detail := range cvmList.Details {
			accountIDs = append(accountIDs, detail.AccountID)
			cvmMap[detail.ID] = detail
			if _, ok := vendorCvmMap[detail.Vendor]; !ok {
				vendorCvmMap[detail.Vendor] = make(map[string]corecvm.BaseCvm)
			}
			vendorCvmMap[detail.Vendor][detail.ID] = detail
		}
	}

	// 2.查询cvm所属账号对应的业务id
	accountIDs = slice.Unique(accountIDs)
	accountBizIDMap := make(map[string][]int64)
	for _, batch := range slice.Split(accountIDs, int(core.DefaultMaxPageLimit)) {
		accountReq := &dataproto.AccountListReq{
			Filter: tools.ContainersExpression("id", batch),
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := cli.DataService().Global.Account.List(kt.Ctx, kt.Header(), accountReq)
		if err != nil {
			logs.Errorf("list account failed, err: %v, account ids: %v, rid: %s", err, batch, kt.Rid)
			return nil, err
		}
		if resp == nil || len(resp.Details) == 0 {
			return nil, fmt.Errorf("not found account by ids(%v)", batch)
		}
		for _, detail := range resp.Details {
			accountBizIDMap[detail.ID] = detail.UsageBizIDs
		}
	}

	// 3.查询cvm对应的mac地址
	cvmIPMacAddrMap, err := getAssignedCvmMacAddr(kt, cli, vendorCvmMap)
	if err != nil {
		logs.Errorf("get assigned cvm mac addr failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	// 4.组合结果数据
	infos := make([]PreviewAssignedCvmInfo, 0, len(ids))
	for _, id := range ids {
		cvmInfo, ok := cvmMap[id]
		if !ok {
			logs.Errorf("not found cvm info by id(%s), rid: %s", id, kt.Rid)
			return nil, fmt.Errorf("not found cvm info by id(%s)", id)
		}
		bizIDs, ok := accountBizIDMap[cvmInfo.AccountID]
		if !ok {
			logs.Errorf("not found biz ids by account id(%s), cvm id(%s), rid: %s", cvmInfo.AccountID, id, kt.Rid)
			return nil, fmt.Errorf("not found biz ids by account id(%s), cvm id(%s)", cvmInfo.AccountID, id)
		}
		var innerIPv4, macAddr string
		if len(cvmInfo.PrivateIPv4Addresses) != 0 {
			innerIPv4 = cvmInfo.PrivateIPv4Addresses[0]
			if ipMacAddrMap, ok := cvmIPMacAddrMap[id]; ok {
				macAddr = ipMacAddrMap[innerIPv4]
			}
		}
		info := PreviewAssignedCvmInfo{CvmID: id, AccountBizIDs: bizIDs, Vendor: cvmInfo.Vendor,
			CloudID: cvmInfo.CloudID, InnerIPv4: innerIPv4, MacAddr: macAddr}
		infos = append(infos, info)
	}
	return infos, nil
}

func getAssignedCvmMacAddr(kt *kit.Kit, cli *client.ClientSet,
	vendorCvmMap map[enumor.Vendor]map[string]corecvm.BaseCvm) (map[string]map[string]string, error) {

	cvmIPv4MacAddrMap := make(map[string]map[string]string)
	for vendor, cvmMap := range vendorCvmMap {
		var err error
		var subCvmIPv4MacAddrMap map[string]map[string]string

		switch vendor {
		case enumor.TCloud, enumor.Aws:
			subCvmIPv4MacAddrMap, err = getAssignedCvmMacAddrFromCloud(kt, cli, vendor, cvmMap)
			if err != nil {
				logs.Errorf("get assigned cvm mac addr from cloud failed, err: %v, cvmMap: %v, rid: %s", err, cvmMap,
					kt.Rid)
				return nil, err
			}

		case enumor.HuaWei, enumor.Azure:
			cvmIDs := converter.MapKeyToStringSlice(cvmMap)
			subCvmIPv4MacAddrMap, err = getAssignedCvmMacAddrFromDB(kt, cli, vendor, cvmIDs)
			if err != nil {
				logs.Errorf("get assigned cvm mac addr from db failed, err: %v, ids: %v, rid: %s", err, cvmIDs, kt.Rid)
				return nil, err
			}

		case enumor.Gcp:
		// todo 暂不能通过接口获取mac地址

		default:
			return nil, fmt.Errorf("no support vendor: %s", vendor)
		}

		cvmIPv4MacAddrMap = maps.MapAppend(cvmIPv4MacAddrMap, subCvmIPv4MacAddrMap)
	}

	return cvmIPv4MacAddrMap, nil
}

func getAssignedCvmMacAddrFromCloud(kt *kit.Kit, cli *client.ClientSet, vendor enumor.Vendor,
	cvmMap map[string]corecvm.BaseCvm) (map[string]map[string]string, error) {

	accountIDRegionCvmIDsMap := make(map[string]map[string][]string)
	for _, cvmInfo := range cvmMap {
		if _, ok := accountIDRegionCvmIDsMap[cvmInfo.AccountID]; !ok {
			accountIDRegionCvmIDsMap[cvmInfo.AccountID] = make(map[string][]string)
		}
		accountIDRegionCvmIDsMap[cvmInfo.AccountID][cvmInfo.Region] =
			append(accountIDRegionCvmIDsMap[cvmInfo.AccountID][cvmInfo.Region], cvmInfo.ID)
	}

	cvmIPv4MacAddrMap := make(map[string]map[string]string)
	for accountID, regionCvmIDsMap := range accountIDRegionCvmIDsMap {
		for region, cvmIDs := range regionCvmIDsMap {
			for _, batch := range slice.Split(cvmIDs, 50) {
				req := &protocvm.ListCvmNetworkInterfaceReq{
					AccountID: accountID,
					Region:    region,
					CvmIDs:    batch,
				}
				var err error
				result := new(map[string]*protocvm.ListCvmNetworkInterfaceRespItem)
				switch vendor {
				case enumor.TCloud:
					result, err = cli.HCService().TCloud.Cvm.ListCvmNetworkInterface(kt, req)
					if err != nil {
						logs.Errorf("list cvm network interface failed, err: %v, vendor: %s, req: %v, rid: %s", err,
							vendor, req, kt.Rid)
						return nil, err
					}
				case enumor.Aws:
					result, err = cli.HCService().Aws.Cvm.ListCvmNetworkInterface(kt, req)
					if err != nil {
						logs.Errorf("list cvm network interface failed, err: %v, vendor: %s, req: %v, rid: %s", err,
							vendor, req, kt.Rid)
						return nil, err
					}
				default:
					return nil, fmt.Errorf("no support vendor: %s", vendor)
				}

				for cvmID, item := range *result {
					if _, ok := cvmIPv4MacAddrMap[cvmID]; !ok {
						cvmIPv4MacAddrMap[cvmID] = make(map[string]string)
					}
					for macAddress, innerIPv4s := range item.MacAddressToPrivateIpAddresses {
						for _, innerIPv4 := range innerIPv4s {
							cvmIPv4MacAddrMap[cvmID][innerIPv4] = macAddress
						}
					}
				}
			}
		}
	}

	return cvmIPv4MacAddrMap, nil
}

func getAssignedCvmMacAddrFromDB(kt *kit.Kit, cli *client.ClientSet, vendor enumor.Vendor, ids []string) (
	map[string]map[string]string, error) {

	cvmIPv4MacAddrMap := make(map[string]map[string]string)
	for _, batch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		reqData := &dataproto.NetworkInterfaceCvmRelWithListReq{CvmIDs: batch}
		switch vendor {
		case enumor.HuaWei:
			resp, err := cli.DataService().HuaWei.ListNetworkCvmRelWithExt(kt.Ctx, kt.Header(), reqData)
			if err != nil {
				logs.Errorf("list network interface cvm rel failed, err: %v, ids: %v, rid: %s", err, batch, kt.Rid)
				return nil, err
			}
			for _, detail := range resp {
				if detail.Extension == nil {
					continue
				}
				if _, ok := cvmIPv4MacAddrMap[detail.CvmID]; !ok {
					cvmIPv4MacAddrMap[detail.CvmID] = make(map[string]string)
				}
				for _, innerIPv4 := range detail.BaseNetworkInterface.PrivateIPv4 {
					cvmIPv4MacAddrMap[detail.CvmID][innerIPv4] = *detail.Extension.MacAddr
				}
			}

		case enumor.Azure:
			resp, err := cli.DataService().Azure.ListNetworkCvmRelWithExt(kt.Ctx, kt.Header(), reqData)
			if err != nil {
				logs.Errorf("list network interface cvm rel failed, err: %v, ids: %v, rid: %s", err, batch, kt.Rid)
				return nil, err
			}
			for _, detail := range resp {
				if detail.Extension == nil {
					continue
				}
				if _, ok := cvmIPv4MacAddrMap[detail.CvmID]; !ok {
					cvmIPv4MacAddrMap[detail.CvmID] = make(map[string]string)
				}
				for _, innerIPv4 := range detail.BaseNetworkInterface.PrivateIPv4 {
					cvmIPv4MacAddrMap[detail.CvmID][innerIPv4] = *detail.Extension.MacAddress
				}
			}

		default:
			return nil, fmt.Errorf("no support vendor: %s", vendor)
		}
	}

	return cvmIPv4MacAddrMap, nil
}

// GetAssignedHostInfoFromCC get assigned host from cc
func GetAssignedHostInfoFromCC(kt *kit.Kit, cmdbCli cmdb.Client, cvmInfos []PreviewAssignedCvmInfo, fields []string) (
	map[int64]cmdb.Host, map[int64][]int64, error) {

	innerIPv4s := make([]string, 0)
	macAddrs := make([]string, 0)
	cloudIDs := make([]string, 0)
	for _, info := range cvmInfos {
		if info.InnerIPv4 != "" {
			innerIPv4s = append(innerIPv4s, info.InnerIPv4)
		}
		if info.MacAddr != "" {
			macAddrs = append(macAddrs, info.MacAddr)
		}
		if info.CloudID != "" {
			cloudIDs = append(cloudIDs, info.CloudID)
		}
	}
	rules := make([]cmdb.Rule, 0)
	if len(innerIPv4s) != 0 {
		rule := &cmdb.CombinedRule{
			Condition: "AND",
			Rules: []cmdb.Rule{
				&cmdb.AtomRule{Field: "bk_addressing", Operator: cmdb.OperatorEqual, Value: cmdb.StaticAddressing},
				&cmdb.AtomRule{Field: "bk_host_innerip", Operator: cmdb.OperatorIn, Value: innerIPv4s},
			},
		}
		rules = append(rules, rule)
	}
	if len(macAddrs) != 0 {
		rules = append(rules, &cmdb.AtomRule{Field: "bk_mac", Operator: cmdb.OperatorIn, Value: macAddrs})
	}
	if len(cloudIDs) != 0 {
		rule := &cmdb.AtomRule{Field: "bk_cloud_inst_id", Operator: cmdb.OperatorIn, Value: cloudIDs}
		rules = append(rules, rule)
	}
	fields = append(fields, "bk_cloud_id")
	listParams := &cmdb.ListHostWithoutBizParams{
		Fields:             fields,
		Page:               &cmdb.BasePage{Sort: "bk_host_id", Start: 0, Limit: int64(core.DefaultMaxPageLimit)},
		HostPropertyFilter: &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "OR", Rules: rules}},
	}
	hostIDs := make([]int64, 0)
	hostMap := make(map[int64]cmdb.Host, 0)
	for {
		hostRes, err := cmdbCli.ListHostWithoutBiz(kt, listParams)
		if err != nil {
			logs.Errorf("list host from cc failed, err: %v, rid: %s", err, kt.Rid)
			return nil, nil, err
		}
		for _, host := range hostRes.Info {
			if host.BkCloudID == 0 { // todo 暂不支持管控区域为0的机器
				continue
			}
			hostIDs = append(hostIDs, host.BkHostID)
			hostMap[host.BkHostID] = host
		}
		if len(hostRes.Info) < int(core.DefaultMaxPageLimit) {
			break
		}
		listParams.Page.Start += int64(core.DefaultMaxPageLimit)
	}

	bizHostIDsMap := make(map[int64][]int64, 0)
	for _, batch := range slice.Split(hostIDs, int(core.DefaultMaxPageLimit)) {
		param := cmdb.HostModuleRelationParams{HostID: batch}
		relationRes, err := cmdbCli.FindHostBizRelations(kt, &param)
		if err != nil {
			logs.Errorf("find cmdb topo relation failed, err: %v, param: %+v, rid: %s", err, param, kt.Rid)
			return nil, nil, err
		}
		for _, relation := range *relationRes {
			if _, ok := bizHostIDsMap[relation.BizID]; !ok {
				bizHostIDsMap[relation.BizID] = make([]int64, 0)
			}
			bizHostIDsMap[relation.BizID] = append(bizHostIDsMap[relation.BizID], relation.HostID)
		}
	}

	return hostMap, bizHostIDsMap, nil
}

func matchAssignedCvm(cvmInfos []PreviewAssignedCvmInfo, ccHosts map[int64]cmdb.Host,
	ccBizHostIDsMap map[int64][]int64) (
	map[string][]PreviewCvmMatchResult, error) {

	result := make(map[string][]PreviewCvmMatchResult, len(cvmInfos))

	for _, cvmInfo := range cvmInfos {
		for _, bizID := range cvmInfo.AccountBizIDs {
			hostIDs, ok := ccBizHostIDsMap[bizID]
			if !ok || len(hostIDs) == 0 {
				continue
			}

			for _, hostID := range hostIDs {
				ccHost, ok := ccHosts[hostID]
				if !ok {
					continue
				}

				ccHostVendor := cmdb.CmdbHcmVendorMap[ccHost.BkCloudVendor]
				if cvmInfo.Vendor == ccHostVendor && cvmInfo.CloudID == ccHost.BkCloudInstID {
					result[cvmInfo.CvmID] = append(result[cvmInfo.CvmID], PreviewCvmMatchResult{
						BkBizID:   bizID,
						BkCloudID: ccHost.BkCloudID,
					})
					continue
				}

				if cvmInfo.InnerIPv4 == ccHost.BkHostInnerIP && cvmInfo.MacAddr == ccHost.BkMac {
					result[cvmInfo.CvmID] = append(result[cvmInfo.CvmID], PreviewCvmMatchResult{
						BkBizID:   bizID,
						BkCloudID: ccHost.BkCloudID,
					})
				}
			}
		}
	}

	return result, nil
}
