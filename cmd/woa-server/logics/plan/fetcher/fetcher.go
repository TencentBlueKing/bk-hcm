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

// Package fetcher ...
package fetcher

import (
	"hcm/cmd/woa-server/logics/biz"
	demandtime "hcm/cmd/woa-server/logics/plan/demand-time"
	"hcm/cmd/woa-server/types/device"
	ptypes "hcm/cmd/woa-server/types/plan"
	dt "hcm/pkg/api/core/cloud/device-type"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	dmtypes "hcm/pkg/dal/dao/types/meta"
	rpdaotypes "hcm/pkg/dal/dao/types/resource-plan"
	tablegconf "hcm/pkg/dal/table/global-config"
	rpd "hcm/pkg/dal/table/resource-plan/res-plan-demand"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	"hcm/pkg/thirdparty/cvmapi"
)

// Fetcher fetch resource plan all resource, like demand / ticket
type Fetcher interface {
	// GetResPlanDemandDetail get resource plan demand detail
	GetResPlanDemandDetail(kt *kit.Kit, demandID string, bkBizIDs []int64) (*ptypes.GetPlanDemandDetailResp, error)
	// ListResPlanDemandByAggregateKey list res plan demand by aggregate key
	ListResPlanDemandByAggregateKey(kt *kit.Kit, demandKey ptypes.ResPlanDemandAggregateKey) (
		[]rpd.ResPlanDemandTable, error)

	// ListAllResPlanTicket list all res plan ticket
	ListAllResPlanTicket(kt *kit.Kit, listFilter *filter.Expression) ([]rpdaotypes.RPTicketWithStatus, error)
	// GetTicketInfo get ticket info
	GetTicketInfo(kt *kit.Kit, ticketID string) (*ptypes.TicketInfo, error)
	// GetResPlanTicketStatusInfo get res plan ticket status info
	GetResPlanTicketStatusInfo(kt *kit.Kit, ticketID string) (*ptypes.GetRPTicketStatusInfo, error)
	// GetResPlanTicketAudit get res plan ticket audit
	GetResPlanTicketAudit(kt *kit.Kit, ticketID string, bkBizID int64) (*ptypes.GetResPlanTicketAuditResp, error)

	// ListResPlanSubTicket list resource plan sub_ticket.
	ListResPlanSubTicket(kt *kit.Kit, req *ptypes.ListResPlanSubTicketReq) (*ptypes.ListResPlanSubTicketResp, error)
	// GetResPlanSubTicketDetail get resource plan sub_ticket detail.
	GetResPlanSubTicketDetail(kt *kit.Kit, subTicketID string) (*ptypes.GetSubTicketDetailResp, string, error)
	// GetResPlanSubTicketAudit get res plan sub ticket audit
	GetResPlanSubTicketAudit(kt *kit.Kit, bizID int64, subTicketID string) (*ptypes.GetSubTicketAuditResp, string,
		error)
	// GetAdminAuditors get admin auditors
	GetAdminAuditors() []string
	// GetSubTicketInfo get sub ticket info
	GetSubTicketInfo(kt *kit.Kit, subTicketID string) (*ptypes.SubTicketInfo, error)

	// GetCrpCurrentApprove get crp current approve
	GetCrpCurrentApprove(kt *kit.Kit, bkBizID int64, orderID string) ([]*ptypes.CrpAuditStep, error)
	// GetCrpApproveLogs get crp approve logs
	GetCrpApproveLogs(kt *kit.Kit, orderID string) ([]*ptypes.CrpAuditLog, error)
	// QueryCRPTransferPoolDemands query crp transfer pool demands
	// Notice: Need to focus on the Year of the data
	QueryCRPTransferPoolDemands(kt *kit.Kit, obsProjects []enumor.ObsProject, technicalClasses []string) (
		[]*cvmapi.CvmCbsPlanQueryItem, error)

	// GetConfigsFromData get configs from global_config table
	GetConfigsFromData(kt *kit.Kit, configType string) ([]tablegconf.GlobalConfigTable, error)
	// GetPlanTransferQuotaConfigs get plan transfer quota configs
	GetPlanTransferQuotaConfigs(kt *kit.Kit) (ptypes.TransferQuotaConfig, error)
	// ListRemainTransferQuota list remain transfer quota
	ListRemainTransferQuota(kt *kit.Kit, req *ptypes.ListResPlanTransferQuotaSummaryReq) (
		*ptypes.ResPlanTransferQuotaSummaryResp, error)

	// GetMetaMaps get res plan meta maps, like zoneMap, regionAreaMap and deviceTypeMap.
	GetMetaMaps(kt *kit.Kit) (map[string]string, map[string]dmtypes.RegionArea, map[string]dt.DistinctDeviceType,
		error)
	// GetMetaNameMapsFromIDMap get meta name maps from id map
	GetMetaNameMapsFromIDMap(zoneMap map[string]string, regionAreaMap map[string]dmtypes.RegionArea) (
		map[string]string, map[string]dmtypes.RegionArea)

	// GetOrderList 根据销毁单据查询预测返还信息
	GetOrderList(kt *kit.Kit, orderID string) ([]*cvmapi.QueryOrderInfo, error)
}

// ResPlanFetcher ...
type ResPlanFetcher struct {
	resPlanCfg   cc.ResPlan
	dao          dao.Set
	demandTime   demandtime.DemandTime
	bizLogics    biz.Logics
	client       *client.ClientSet
	crpAuditNode cc.StateNode
	crpCli       cvmapi.CVMClientInterface
	itsmCli      itsm.Client

	deviceTypesMap *device.DeviceTypesMap
}

// New create Fetcher
func New(dao dao.Set, demandTime demandtime.DemandTime, client *client.ClientSet, crpCli cvmapi.CVMClientInterface,
	itsmCli itsm.Client, bizLogics biz.Logics, deviceTypesMap *device.DeviceTypesMap) Fetcher {

	var itsmFlowCfg cc.ItsmFlow
	for _, itsmFlow := range cc.WoaServer().ItsmFlows {
		if itsmFlow.ServiceName == enumor.TicketSvcNameResPlan {
			itsmFlowCfg = itsmFlow
			break
		}
	}

	var crpAuditNode cc.StateNode
	for _, node := range itsmFlowCfg.StateNodes {
		if node.NodeName == enumor.TicketNodeNameCrpAudit {
			crpAuditNode = node
		}
	}

	return &ResPlanFetcher{
		resPlanCfg:     cc.WoaServer().ResPlan,
		dao:            dao,
		demandTime:     demandTime,
		bizLogics:      bizLogics,
		client:         client,
		crpAuditNode:   crpAuditNode,
		crpCli:         crpCli,
		itsmCli:        itsmCli,
		deviceTypesMap: deviceTypesMap,
	}
}
