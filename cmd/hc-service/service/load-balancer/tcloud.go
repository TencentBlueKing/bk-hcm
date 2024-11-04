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

package loadbalancer

import (
	"fmt"
	"net/http"

	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/adaptor/tcloud"
	adcore "hcm/pkg/adaptor/types/core"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	protolb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

func (svc *clbSvc) initTCloudClbService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("BatchCreateTCloudClb", http.MethodPost,
		"/vendors/tcloud/load_balancers/batch/create", svc.BatchCreateTCloudClb)
	h.Add("InquiryPriceTCloudLB", http.MethodPost,
		"/vendors/tcloud/load_balancers/prices/inquiry", svc.InquiryPriceTCloudLB)
	h.Add("ListTCloudClb", http.MethodPost, "/vendors/tcloud/load_balancers/list", svc.ListTCloudClb)
	h.Add("TCloudDescribeResources", http.MethodPost,
		"/vendors/tcloud/load_balancers/resources/describe", svc.TCloudDescribeResources)
	h.Add("TCloudUpdateCLB", http.MethodPatch, "/vendors/tcloud/load_balancers/{id}", svc.TCloudUpdateCLB)
	h.Add("BatchDeleteTCloudLoadBalancer", http.MethodDelete,
		"/vendors/tcloud/load_balancers/batch", svc.BatchDeleteTCloudLoadBalancer)
	h.Add("ListQuotaTCloudLB", http.MethodPost, "/vendors/tcloud/load_balancers/quota", svc.ListTCloudLBQuota)
	h.Add("TCloudCreateSnatIps", http.MethodPost,
		"/vendors/tcloud/load_balancers/snat_ips/create", svc.TCloudCreateSnatIps)
	h.Add("TCloudDeleteSnatIps", http.MethodDelete,
		"/vendors/tcloud/load_balancers/snat_ips", svc.TCloudDeleteSnatIps)

	h.Add("TCloudCreateUrlRule", http.MethodPost,
		"/vendors/tcloud/listeners/{lbl_id}/rules/batch/create", svc.TCloudCreateUrlRule)
	h.Add("TCloudUpdateUrlRule", http.MethodPatch,
		"/vendors/tcloud/listeners/{lbl_id}/rules/{rule_id}", svc.TCloudUpdateUrlRule)
	h.Add("TCloudBatchDeleteUrlRule", http.MethodDelete,
		"/vendors/tcloud/listeners/{lbl_id}/rules/batch", svc.TCloudBatchDeleteUrlRule)
	h.Add("TCloudBatchDeleteUrlRuleByDomain", http.MethodDelete,
		"/vendors/tcloud/listeners/{lbl_id}/rules/by/domain/batch", svc.TCloudBatchDeleteUrlRuleByDomain)

	// 监听器
	h.Add("CreateTCloudListener", http.MethodPost, "/vendors/tcloud/listeners/create", svc.CreateTCloudListener)
	h.Add("UpdateTCloudListener", http.MethodPatch, "/vendors/tcloud/listeners/{id}", svc.UpdateTCloudListener)
	h.Add("UpdateTCloudListenerHealthCheck", http.MethodPatch,
		"/vendors/tcloud/listeners/{lbl_id}/health_check", svc.UpdateTCloudListenerHealthCheck)
	h.Add("DeleteTCloudListener", http.MethodDelete, "/vendors/tcloud/listeners/batch", svc.DeleteTCloudListener)

	// 域名、规则
	h.Add("UpdateTCloudDomainAttr", http.MethodPatch,
		"/vendors/tcloud/listeners/{lbl_id}/domains", svc.UpdateTCloudDomainAttr)

	// 目标组
	h.Add("BatchCreateTCloudTargets", http.MethodPost,
		"/vendors/tcloud/target_groups/{target_group_id}/targets/create", svc.BatchCreateTCloudTargets)
	h.Add("BatchRemoveTCloudTargets", http.MethodDelete,
		"/vendors/tcloud/target_groups/{target_group_id}/targets/batch", svc.BatchRemoveTCloudTargets)
	h.Add("BatchModifyTCloudTargetsPort", http.MethodPatch,
		"/vendors/tcloud/target_groups/{target_group_id}/targets/port", svc.BatchModifyTCloudTargetsPort)
	h.Add("BatchModifyTCloudTargetsWeight", http.MethodPatch,
		"/vendors/tcloud/target_groups/{target_group_id}/targets/weight", svc.BatchModifyTCloudTargetsWeight)
	h.Add("ListTCloudTargetsHealth", http.MethodPost,
		"/vendors/tcloud/load_balancers/targets/health", svc.ListTCloudTargetsHealth)

	h.Add("RegisterTargetToListenerRule", http.MethodPost,
		"/vendors/tcloud/load_balancers/{lb_id}/targets/create", svc.RegisterTargetToListenerRule)

	h.Add("QueryListenerTargetsByCloudIDs", http.MethodPost,
		"/vendors/tcloud/targets/query_by_cloud_ids", svc.QueryListenerTargetsByCloudIDs)
	h.Load(cap.WebService)
}

// BatchCreateTCloudClb ...
func (svc *clbSvc) BatchCreateTCloudClb(cts *rest.Contexts) (interface{}, error) {
	req := new(protolb.TCloudLoadBalancerCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloudAdpt, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	createOpt := &typelb.TCloudCreateClbOption{
		Region:           req.Region,
		LoadBalancerType: req.LoadBalancerType,
		LoadBalancerName: req.Name,
		VpcID:            req.CloudVpcID,
		SubnetID:         req.CloudSubnetID,
		Vip:              req.Vip,
		VipIsp:           req.VipIsp,

		InternetChargeType:      req.InternetChargeType,
		InternetMaxBandwidthOut: req.InternetMaxBandwidthOut,

		BandwidthPackageID: req.BandwidthPackageID,
		SlaType:            req.SlaType,
		Number:             req.RequireCount,
		ClientToken:        cvt.StrNilPtr(cts.Kit.Rid),

		BandwidthpkgSubType: req.BandwidthpkgSubType,
	}
	if cvt.PtrToVal(req.CloudEipID) != "" {
		createOpt.EipAddressID = req.CloudEipID
	}
	if req.AddressIPVersion == "" {
		req.AddressIPVersion = typelb.IPV4IPVersion
	}
	// 负载均衡实例的网络类型-公网属性
	if req.LoadBalancerType == typelb.OpenLoadBalancerType {
		// IP版本-仅适用于公网负载均衡
		createOpt.AddressIPVersion = req.AddressIPVersion
		// 静态单线IP 线路类型-仅适用于公网负载均衡, 如果不指定本参数，则默认使用BGP
		createOpt.VipIsp = req.VipIsp

		// 设置跨可用区容灾时的可用区ID-仅适用于公网负载均衡
		if len(req.BackupZones) > 0 && len(req.Zones) > 0 {
			// 主备可用区，传递zones（单元素数组），以及backup_zones
			createOpt.MasterZoneID = cvt.ValToPtr(req.Zones[0])
			createOpt.SlaveZoneID = cvt.ValToPtr(req.BackupZones[0])
		} else if len(req.Zones) > 0 {
			// 单可用区
			createOpt.ZoneID = cvt.ValToPtr(req.Zones[0])
		}
	}

	result, err := tcloudAdpt.CreateLoadBalancer(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("create tcloud clb failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	respData := &protolb.BatchCreateResult{
		UnknownCloudIDs: result.UnknownCloudIDs,
		SuccessCloudIDs: result.SuccessCloudIDs,
		FailedCloudIDs:  result.FailedCloudIDs,
		FailedMessage:   result.FailedMessage,
	}

	if len(result.SuccessCloudIDs) == 0 {
		return respData, nil
	}

	// 数据库创建失败也继续同步
	_ = svc.createTCloudDBLoadBalancer(cts, req, result.SuccessCloudIDs)

	if err := svc.lbSync(cts.Kit, tcloudAdpt, req.AccountID, req.Region, result.SuccessCloudIDs); err != nil {
		return nil, err
	}
	return respData, nil
}

func (svc *clbSvc) createTCloudDBLoadBalancer(cts *rest.Contexts, req *protolb.TCloudLoadBalancerCreateReq,
	cloudIDs []string) (err error) {

	dataReq := &dataproto.TCloudCLBCreateReq{Lbs: make([]dataproto.TCloudCLBCreate, len(cloudIDs))}
	for i, cloudID := range cloudIDs {
		dataReq.Lbs[i].CloudID = cloudID
		dataReq.Lbs[i].Vendor = enumor.TCloud
		dataReq.Lbs[i].BkBizID = req.BkBizID

		dataReq.Lbs[i].Name = fmt.Sprintf("%s-%d", cvt.PtrToVal(req.Name), i)
		dataReq.Lbs[i].AccountID = req.AccountID
		dataReq.Lbs[i].Region = req.Region
		dataReq.Lbs[i].LoadBalancerType = string(req.LoadBalancerType)
		dataReq.Lbs[i].IPVersion = req.AddressIPVersion.Convert()
		dataReq.Lbs[i].Zones = req.Zones
		dataReq.Lbs[i].BackupZones = req.BackupZones

	}
	// 创建本地数据，保存业务信息
	_, err = svc.dataCli.TCloud.LoadBalancer.BatchCreateTCloudClb(cts.Kit, dataReq)
	if err != nil {
		logs.Errorf("fail to create db load balancer after cloud create, err: %v, rid: %s", err, cts.Kit.Rid)
		// 	失败也继续尝试同步
	}
	return err
}

// ListTCloudClb list tcloud clb
func (svc *clbSvc) ListTCloudClb(cts *rest.Contexts) (interface{}, error) {
	req := new(protolb.TCloudListOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tcloudAdpt, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	if req.Page.Limit > adcore.TCloudQueryLimit {
		req.Page.Limit = adcore.TCloudQueryLimit
	}
	opt := &typelb.TCloudListOption{
		Region:   req.Region,
		CloudIDs: req.CloudIDs,
		Page:     req.Page,
	}
	result, err := tcloudAdpt.ListLoadBalancer(cts.Kit, opt)
	if err != nil {
		logs.Errorf("[%s] list tcloud clb failed, req: %+v, err: %v, rid: %s",
			enumor.TCloud, req, err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// TCloudDescribeResources 查询clb地域下可用资源
func (svc *clbSvc) TCloudDescribeResources(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudDescribeResourcesOption)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	return client.DescribeResources(cts.Kit, req.TCloudDescribeResourcesOption)
}

// TCloudUpdateCLB 更新clb属性
func (svc *clbSvc) TCloudUpdateCLB(cts *rest.Contexts) (any, error) {
	lbID := cts.PathParameter("id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(protolb.TCloudLBUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 获取lb基本信息
	lb, err := svc.dataCli.TCloud.LoadBalancer.Get(cts.Kit, lbID)
	if err != nil {
		logs.Errorf("fail to get tcloud clb(%s), err: %v, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}

	// 调用云上更新接口
	client, err := svc.ad.TCloud(cts.Kit, lb.AccountID)
	if err != nil {
		return nil, err
	}

	adtOpt := &typelb.TCloudUpdateOption{
		Region:                   lb.Region,
		LoadBalancerId:           lb.CloudID,
		LoadBalancerName:         req.Name,
		InternetChargeType:       req.InternetChargeType,
		InternetMaxBandwidthOut:  req.InternetMaxBandwidthOut,
		BandwidthpkgSubType:      req.BandwidthpkgSubType,
		LoadBalancerPassToTarget: req.LoadBalancerPassToTarget,
		SnatPro:                  req.SnatPro,
		DeleteProtect:            req.DeleteProtect,
		ModifyClassicDomain:      req.ModifyClassicDomain,
	}

	if req.TargetRegion != nil || req.TargetCloudVpcID != nil {
		adtOpt.TargetRegionInfo.Region = req.TargetRegion
		adtOpt.TargetRegionInfo.VpcId = req.TargetCloudVpcID
	}

	_, err = client.UpdateLoadBalancer(cts.Kit, adtOpt)
	if err != nil {
		logs.Errorf("fail to call tcloud update load balancer(id:%s),err: %v, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}

	// 同步云上变更信息
	return nil, svc.lbSync(cts.Kit, client, lb.AccountID, lb.Region, []string{lb.CloudID})

}

// 同步云上资源
func (svc *clbSvc) lbSync(kt *kit.Kit, tcloud tcloud.TCloud, accountID string, region string, lbIDs []string) error {

	syncClient := synctcloud.NewClient(svc.dataCli, tcloud)
	params := &synctcloud.SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  lbIDs,
	}
	_, err := syncClient.LoadBalancer(kt, params, &synctcloud.SyncLBOption{})
	if err != nil {
		logs.Errorf("sync load  balancer failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (svc *clbSvc) getListenerWithLb(kt *kit.Kit, lblID string) (*corelb.BaseLoadBalancer,
	*corelb.BaseListener, error) {

	// 查询监听器数据
	lblResp, err := svc.dataCli.Global.LoadBalancer.ListListener(kt, &core.ListReq{
		Filter: tools.EqualExpression("id", lblID),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	})
	if err != nil {
		logs.Errorf("fail to list tcloud listener, err: %v, id: %s, rid: %s", err, lblID, kt.Rid)
		return nil, nil, err
	}
	if len(lblResp.Details) < 1 {
		return nil, nil, errf.Newf(errf.InvalidParameter, "lbl not found")
	}
	listener := lblResp.Details[0]

	// 查询负载均衡
	lbResp, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(kt, &core.ListReq{
		Filter: tools.EqualExpression("id", listener.LbID),
		Page:   core.NewDefaultBasePage(),
		Fields: nil,
	})
	if err != nil {
		logs.Errorf("fail to tcloud list load balancer, err: %v, id: %s, rid: %s", err, listener.LbID, kt.Rid)
		return nil, nil, err
	}
	if len(lbResp.Details) < 1 {
		return nil, nil, errf.Newf(errf.RecordNotFound, "lb not found")
	}
	lb := lbResp.Details[0]
	return &lb, &listener, nil
}

// CreateTCloudListener 创建监听器
func (svc *clbSvc) CreateTCloudListener(cts *rest.Contexts) (interface{}, error) {
	req := new(protolb.ListenerWithRuleCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 根据lbID，查询负载均衡信息
	lbReq := &core.ListReq{
		Filter: tools.EqualExpression("id", req.LbID),
		Page:   core.NewDefaultBasePage(),
	}
	lbList, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(cts.Kit, lbReq)
	if err != nil {
		logs.Errorf("list load balancer by id failed, id: %s, err: %v, rid: %s", req.LbID, err, cts.Kit.Rid)
		return nil, err
	}
	if len(lbList.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "load balancer: %s not found", req.LbID)
	}
	lbInfo := lbList.Details[0]

	// 查询目标组是否存在
	targetGroupList, err := svc.getTargetGroupByID(cts.Kit, req.TargetGroupID)
	if err != nil {
		logs.Errorf("list target group by id failed, tgID: %s, err: %v, rid: %s", req.TargetGroupID, err, cts.Kit.Rid)
		return nil, err
	}
	if len(targetGroupList) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "target group: %s not found", req.TargetGroupID)
	}
	targetGroupInfo := targetGroupList[0]

	// 检查目标组是否已经绑定了其他监听器
	relOpt := &core.ListReq{
		Filter: tools.EqualExpression("target_group_id", req.TargetGroupID),
		Page:   core.NewDefaultBasePage(),
	}
	relList, err := svc.dataCli.Global.LoadBalancer.ListTargetGroupListenerRel(cts.Kit, relOpt)
	if err != nil {
		logs.Errorf("list target listener rule rel failed, tgID: %s, err: %v, rid: %s",
			req.TargetGroupID, err, cts.Kit.Rid)
		return nil, err
	}
	if len(relList.Details) > 0 {
		return nil, errf.Newf(errf.InvalidParameter, "target_group_id: %s has bound listener", req.TargetGroupID)
	}

	// 创建云端监听器、规则
	cloudLblID, cloudRuleID, err := svc.createListenerWithRule(cts.Kit, req, lbInfo)
	if err != nil {
		return nil, err
	}

	// 插入新的监听器、规则信息到DB
	_, err = svc.insertListenerWithRule(cts.Kit, req, lbInfo, cloudLblID, cloudRuleID, targetGroupInfo)
	if err != nil {
		return nil, err
	}

	return &protolb.ListenerWithRuleCreateResult{CloudLblID: cloudLblID, CloudRuleID: cloudRuleID}, nil
}

func (svc *clbSvc) createListenerWithRule(kt *kit.Kit, req *protolb.ListenerWithRuleCreateReq,
	lbInfo corelb.BaseLoadBalancer) (string, string, error) {

	tcloudAdpt, err := svc.ad.TCloud(kt, lbInfo.AccountID)
	if err != nil {
		return "", "", err
	}

	lblOpt := &typelb.TCloudCreateListenerOption{
		Region:            lbInfo.Region,
		LoadBalancerId:    lbInfo.CloudID,
		ListenerName:      req.Name,
		Protocol:          req.Protocol,
		Port:              req.Port,
		SessionExpireTime: req.SessionExpire,
		Scheduler:         req.Scheduler,
		SniSwitch:         req.SniSwitch,
		SessionType:       cvt.ValToPtr(req.SessionType),
		Certificate:       req.Certificate,
	}
	// 7层监听器，不管SNI开启还是关闭，都需要传入证书参数
	// 7层监听器并且SNI开启时，创建监听器接口，不需要证书
	if req.Protocol == enumor.HttpsProtocol {
		if req.Certificate == nil {
			return "", "", errf.New(errf.InvalidParameter, "certificate is required when layer 7 listener")
		}
		if cvt.PtrToVal(req.Certificate.CaCloudID) == "" && len(req.Certificate.CertCloudIDs) == 0 {
			return "", "", errf.New(errf.InvalidParameter,
				"certificate.ca_cloud_id and certificate.cert_cloud_ids is required")
		}
		if req.SniSwitch == enumor.SniTypeOpen {
			lblOpt.Certificate = nil
		}
	}
	result, err := tcloudAdpt.CreateListener(kt, lblOpt)
	if err != nil {
		logs.Errorf("create tcloud listener api failed, err: %v, lblOpt: %+v, cert: %+v, rid: %s",
			err, lblOpt, cvt.PtrToVal(lblOpt.Certificate), kt.Rid)
		return "", "", err
	}
	cloudLblID := result.SuccessCloudIDs[0]

	// 只有7层规则才走云端创建规则接口
	var cloudRuleID string
	if req.Protocol.IsLayer7Protocol() {
		ruleOpt := &typelb.TCloudCreateRuleOption{
			Region:         lbInfo.Region,
			LoadBalancerId: lbInfo.CloudID,
			ListenerId:     cloudLblID,
			Rules:          []*typelb.RuleInfo{},
		}
		oneRule := &typelb.RuleInfo{
			Url:               cvt.ValToPtr(req.Url),
			SessionExpireTime: cvt.ValToPtr(req.SessionExpire),
			DefaultServer:     cvt.ValToPtr(true),
		}
		if len(req.Domain) > 0 {
			oneRule.Domain = cvt.ValToPtr(req.Domain)
		}
		if len(req.Scheduler) > 0 {
			oneRule.Scheduler = cvt.ValToPtr(req.Scheduler)
		}
		if req.Certificate != nil {
			oneRule.Certificate = req.Certificate
		}
		ruleOpt.Rules = append(ruleOpt.Rules, oneRule)
		ruleResult, err := tcloudAdpt.CreateRule(kt, ruleOpt)
		if err != nil {
			logs.Errorf("create tcloud listener rule api failed, err: %v, ruleOpt: %+v, cert: %+v, rid: %s",
				err, ruleOpt, cvt.PtrToVal(req.Certificate), kt.Rid)
			return "", "", err
		}
		cloudRuleID = ruleResult.SuccessCloudIDs[0]
	}

	return cloudLblID, cloudRuleID, nil
}

func (svc *clbSvc) insertListenerWithRule(kt *kit.Kit, req *protolb.ListenerWithRuleCreateReq,
	lbInfo corelb.BaseLoadBalancer, cloudLblID string, cloudRuleID string, targetGroupInfo corelb.BaseTargetGroup) (
	*core.BatchCreateResult, error) {

	var domain, url string
	var ruleType = enumor.Layer4RuleType
	if req.Protocol.IsLayer7Protocol() {
		ruleType = enumor.Layer7RuleType
		// 只有7层监听器才有域名、URL
		domain = req.Domain
		url = req.Url
	} else {
		// 4层监听器对应的云端规则ID就是云监听器ID
		cloudRuleID = cloudLblID
	}

	lblRuleReq := &dataproto.ListenerWithRuleBatchCreateReq{
		ListenerWithRules: []dataproto.ListenerWithRuleCreateReq{
			{
				CloudID:            cloudLblID,
				Name:               req.Name,
				Vendor:             enumor.TCloud,
				AccountID:          lbInfo.AccountID,
				BkBizID:            req.BkBizID,
				LbID:               req.LbID,
				CloudLbID:          lbInfo.CloudID,
				Protocol:           req.Protocol,
				Port:               req.Port,
				CloudRuleID:        cloudRuleID,
				Scheduler:          req.Scheduler,
				RuleType:           ruleType,
				SessionType:        req.SessionType,
				SessionExpire:      req.SessionExpire,
				TargetGroupID:      req.TargetGroupID,
				CloudTargetGroupID: targetGroupInfo.CloudID,
				Domain:             domain,
				Url:                url,
				SniSwitch:          req.SniSwitch,
				Certificate:        req.Certificate,
			},
		},
	}
	ids, err := svc.dataCli.TCloud.LoadBalancer.BatchCreateTCloudListenerWithRule(kt, lblRuleReq)
	if err != nil {
		logs.Errorf("create tcloud listener with rule failed, req: %+v, lblRuleReq: %+v, err: %v, rid: %s",
			req, lblRuleReq, err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

func (svc *clbSvc) getTargetGroupByID(kt *kit.Kit, targetGroupID string) ([]corelb.BaseTargetGroup, error) {
	tgReq := &core.ListReq{
		Filter: tools.EqualExpression("id", targetGroupID),
		Page:   core.NewDefaultBasePage(),
	}
	targetGroupInfo, err := svc.dataCli.Global.LoadBalancer.ListTargetGroup(kt, tgReq)
	if err != nil {
		logs.Errorf("list target group db failed, tgID: %s, err: %v, rid: %s", targetGroupID, err, kt.Rid)
		return nil, err
	}

	return targetGroupInfo.Details, nil
}

// UpdateTCloudListener 更新监听器信息
func (svc *clbSvc) UpdateTCloudListener(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(protolb.ListenerWithRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 获取监听器基本信息
	lblInfo, err := svc.dataCli.TCloud.LoadBalancer.GetListener(cts.Kit, lblID)
	if err != nil {
		logs.Errorf("fail to get tcloud listener(%s), err: %v, rid: %s", lblID, err, cts.Kit.Rid)
		return nil, err
	}

	// 只有HTTPS支持开启SNI开关
	if lblInfo.Protocol != enumor.HttpsProtocol && req.SniSwitch == enumor.SniTypeOpen {
		return nil, errf.Newf(errf.InvalidParameter, "only https listener support sni")
	}

	lbInfo, err := svc.dataCli.TCloud.LoadBalancer.Get(cts.Kit, lblInfo.LbID)
	if err != nil {
		logs.Errorf("fail to get tcloud load balancer(%s), err: %v, rid: %s", lblInfo.LbID, err, cts.Kit.Rid)
		return nil, err
	}

	// 调用云上更新接口
	client, err := svc.ad.TCloud(cts.Kit, lblInfo.AccountID)
	if err != nil {
		return nil, err
	}

	// 更新云端监听器信息
	lblOpt := &typelb.TCloudUpdateListenerOption{
		Region:         lbInfo.Region,
		LoadBalancerId: lblInfo.CloudLbID,
		ListenerId:     lblInfo.CloudID,
		ListenerName:   req.Name,
		SniSwitch:      req.SniSwitch,
	}
	if req.Extension != nil {
		lblOpt.Certificate = req.Extension.Certificate
	}
	err = client.UpdateListener(cts.Kit, lblOpt)
	if err != nil {
		logs.Errorf("fail to call tcloud update listener(id:%s), err: %v, rid: %s", lblID, err, cts.Kit.Rid)
		return nil, err
	}
	if err := svc.lblSync(cts.Kit, client, &lbInfo.BaseLoadBalancer, []string{lblInfo.CloudID}); err != nil {
		// 调用同步的方法内会打印错误，这里只标记调用方
		logs.Errorf("fail to sync listener for update listener(%s), rid: %s", lblInfo.ID, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// UpdateTCloudListenerHealthCheck 更新监听器信健康检查信息
func (svc *clbSvc) UpdateTCloudListenerHealthCheck(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(protolb.HealthCheckUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 获取监听器基本信息
	lblInfo, err := svc.dataCli.TCloud.LoadBalancer.GetListener(cts.Kit, lblID)
	if err != nil {
		logs.Errorf("fail to get tcloud listener(%s), err: %v, rid: %s", lblID, err, cts.Kit.Rid)
		return nil, err
	}

	// 改接口只支持修改四层监听器健康检查
	if lblInfo.Protocol.IsLayer7Protocol() {
		return nil, errf.Newf(errf.InvalidParameter, "only layer 4 listener support update health check")
	}

	lbInfo, err := svc.dataCli.TCloud.LoadBalancer.Get(cts.Kit, lblInfo.LbID)
	if err != nil {
		logs.Errorf("fail to get tcloud load balancer(%s), err: %v, rid: %s", lblInfo.LbID, err, cts.Kit.Rid)
		return nil, err
	}

	// 调用云上更新接口
	client, err := svc.ad.TCloud(cts.Kit, lblInfo.AccountID)
	if err != nil {
		return nil, err
	}

	// 更新云端监听器信息
	lblOpt := &typelb.TCloudUpdateListenerOption{
		Region:         lbInfo.Region,
		LoadBalancerId: lblInfo.CloudLbID,
		ListenerId:     lblInfo.CloudID,
		HealthCheck:    req.HealthCheck,
	}
	err = client.UpdateListener(cts.Kit, lblOpt)
	if err != nil {
		logs.Errorf("fail to call tcloud update listener(id:%s), err: %v, rid: %s", lblID, err, cts.Kit.Rid)
		return nil, err
	}
	if err := svc.lblSync(cts.Kit, client, &lbInfo.BaseLoadBalancer, []string{lblInfo.CloudID}); err != nil {
		// 调用同步的方法内会打印错误，这里只标记调用方
		logs.Errorf("fail to sync listener for update listener(%s), rid: %s", lblInfo.ID, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// DeleteTCloudListener 删除监听器信息
func (svc *clbSvc) DeleteTCloudListener(cts *rest.Contexts) (any, error) {
	req := new(core.BatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(req.IDs) > constant.BatchListenerMaxLimit {
		return nil, fmt.Errorf("batch delete listener count should <= %d", constant.BatchListenerMaxLimit)
	}

	lblIDs, lblCloudIDs, lbList, ruleMap, err := svc.getListenerWithRule(cts.Kit, req)
	if err != nil {
		return nil, err
	}

	lbInfo := lbList.Details[0]
	client, err := svc.ad.TCloud(cts.Kit, lbInfo.AccountID)
	if err != nil {
		return nil, err
	}

	// 批量删除云端监听器规则
	for tmpCloudLblID, tmpCloudRuleIDs := range ruleMap {
		ruleOpt := &typelb.TCloudDeleteRuleOption{
			Region:         lbInfo.Region,
			LoadBalancerId: lbInfo.CloudID,
			ListenerId:     tmpCloudLblID,
			CloudIDs:       tmpCloudRuleIDs,
		}
		err = client.DeleteRule(cts.Kit, ruleOpt)
		if err != nil {
			logs.Errorf("fail to call tcloud delete listener rule, lbID: %s, lblID: %s, ruleIDs: %+v, err: %v, rid: %s",
				lbInfo.CloudID, tmpCloudLblID, tmpCloudRuleIDs, err, cts.Kit.Rid)
			return nil, err
		}
	}

	// 批量删除云端监听器
	lblOpt := &typelb.TCloudDeleteListenerOption{
		Region:         lbInfo.Region,
		LoadBalancerId: lbInfo.CloudID,
		CloudIDs:       lblCloudIDs,
	}
	err = client.DeleteListener(cts.Kit, lblOpt)
	if err != nil {
		logs.Errorf("fail to call tcloud delete listener, lblCloudIDs: %v, err: %v, rid: %s",
			lblCloudIDs, err, cts.Kit.Rid)
		return nil, err
	}

	// 更新DB监听器信息
	delLblReq := &dataproto.LoadBalancerBatchDeleteReq{
		Filter: tools.ContainersExpression("id", lblIDs),
	}
	err = svc.dataCli.Global.LoadBalancer.DeleteListener(cts.Kit, delLblReq)
	if err != nil {
		logs.Errorf("delete tcloud listener db failed, lblIDs: %+v, err: %v, rid: %s", lblIDs, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *clbSvc) getListenerWithRule(kt *kit.Kit, req *core.BatchDeleteReq) ([]string, []string,
	*dataproto.LbListResult, map[string][]string, error) {

	// 获取监听器列表
	lblReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	lblList, err := svc.dataCli.Global.LoadBalancer.ListListener(kt, lblReq)
	if err != nil {
		logs.Errorf("fail to list tcloud listener, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, nil, nil, nil, err
	}
	if len(lblList.Details) == 0 {
		return nil, nil, nil, nil, errf.Newf(errf.RecordNotFound, "listeners: %v not found", req.IDs)
	}

	lblIDs := make([]string, 0)
	lbIDs := make([]string, 0)
	lblCloudIDs := make([]string, 0)
	for _, item := range lblList.Details {
		lblIDs = append(lblIDs, item.ID)
		lbIDs = append(lbIDs, item.LbID)
		lblCloudIDs = append(lblCloudIDs, item.CloudID)
	}

	// 根据lbID，查询负载均衡信息
	lbReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", lbIDs),
		Page:   core.NewDefaultBasePage(),
	}
	lbList, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(kt, lbReq)
	if err != nil {
		logs.Errorf("list load balancer by id failed, lbIDs: %v, err: %v, rid: %s", lbIDs, err, kt.Rid)
		return nil, nil, nil, nil, err
	}
	if len(lbList.Details) != 1 {
		return nil, nil, nil, nil, errf.Newf(errf.RecordNotFound, "load balancer: [%v] not found or "+
			"need belong to the same load balancer", lbIDs)
	}

	// 查询监听器规则列表
	ruleReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("lbl_id", lblIDs),
			tools.RuleEqual("rule_type", enumor.Layer7RuleType),
		),
		Page: core.NewDefaultBasePage(),
	}
	lblRuleList, err := svc.dataCli.TCloud.LoadBalancer.ListUrlRule(kt, ruleReq)
	if err != nil {
		logs.Errorf("fail to list tcloud listeners url rule, lblIDs: %v, err: %v, rid: %s", lblIDs, err, kt.Rid)
		return nil, nil, nil, nil, err
	}

	ruleMap := make(map[string][]string)
	for _, ruleItem := range lblRuleList.Details {
		if _, ok := ruleMap[ruleItem.CloudLBLID]; !ok {
			ruleMap[ruleItem.CloudLBLID] = make([]string, 0)
		}
		ruleMap[ruleItem.CloudLBLID] = append(ruleMap[ruleItem.CloudLBLID], ruleItem.CloudID)
	}

	return lblIDs, lblCloudIDs, lbList, ruleMap, nil
}

// UpdateTCloudDomainAttr 更新域名属性
func (svc *clbSvc) UpdateTCloudDomainAttr(cts *rest.Contexts) (any, error) {
	lblID := cts.PathParameter("lbl_id").String()
	if len(lblID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "lbl_id is required")
	}

	req := new(protolb.DomainAttrUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 获取监听器基本信息
	lblInfo, err := svc.dataCli.TCloud.LoadBalancer.GetListener(cts.Kit, lblID)
	if err != nil || lblInfo == nil {
		logs.Errorf("fail to get tcloud listener(%s), err: %v, rid: %s", lblID, err, cts.Kit.Rid)
		return nil, err
	}
	// 只有7层监听器才能更新域名
	if !lblInfo.Protocol.IsLayer7Protocol() {
		return nil, errf.Newf(errf.InvalidParameter, "only layer 7 listeners can be updated")
	}
	// 只有SNI开启的监听器，才能更新域名下的证书信息（非sni更新证书是在监听器上的，单个规则/域名没有单独的证书信息）
	if req.Certificate != nil && lblInfo.SniSwitch == enumor.SniTypeClose {
		return nil, errf.Newf(errf.InvalidParameter, "the certificate of the domain can not update when SNI closed")
	}

	// 调用云上更新接口
	return nil, svc.updateTCloudDomainAttr(cts.Kit, req, lblInfo)

}

func (svc *clbSvc) updateTCloudDomainAttr(kt *kit.Kit, req *protolb.DomainAttrUpdateReq,
	lblInfo *corelb.Listener[corelb.TCloudListenerExtension]) error {

	// 获取规则列表
	ruleOpt := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("lbl_id", lblInfo.ID), tools.RuleEqual("domain", req.Domain)),
		Page:   core.NewDefaultBasePage(),
	}
	ruleList, err := svc.dataCli.TCloud.LoadBalancer.ListUrlRule(kt, ruleOpt)
	if err != nil {
		logs.Errorf("fail to list tcloud rule, lblID: %s, err: %v, rid: %s", lblInfo.ID, err, kt.Rid)
		return err
	}

	if len(ruleList.Details) == 0 {
		return errf.Newf(errf.RecordNotFound, "domain: %s not found", req.Domain)
	}

	// 获取负载均衡信息
	lbInfo, err := svc.dataCli.TCloud.LoadBalancer.Get(kt, lblInfo.LbID)
	if err != nil {
		logs.Errorf("fail to get tcloud load balancer(%s), err: %v, rid: %s", lblInfo.LbID, err, kt.Rid)
		return err
	}
	if lbInfo == nil {
		return errf.Newf(errf.RecordNotFound, "load balancer: %s not found", lblInfo.LbID)
	}

	client, err := svc.ad.TCloud(kt, lbInfo.AccountID)
	if err != nil {
		return err
	}

	// 更新云端域名属性信息
	domainOpt := &typelb.TCloudUpdateDomainAttrOption{
		Region:         lbInfo.Region,
		LoadBalancerId: lbInfo.CloudID,
		ListenerId:     lblInfo.CloudID,
		Domain:         req.Domain,
	}
	if len(req.NewDomain) > 0 {
		domainOpt.NewDomain = req.NewDomain
	}
	if req.Certificate != nil {
		domainOpt.Certificate = req.Certificate
	}
	if req.DefaultServer != nil {
		domainOpt.DefaultServer = req.DefaultServer
	}
	// 只有HTTPS域名才能开启Http2、Quic
	if lblInfo.Protocol == enumor.HttpsProtocol {
		domainOpt.Http2 = req.Http2
		domainOpt.Quic = req.Quic
	}
	err = client.UpdateDomainAttr(kt, domainOpt)
	if err != nil {
		logs.Errorf("fail to call tcloud update domain attr, err: %v, lblID: %s, rid: %s", err, lblInfo.ID, kt.Rid)
		return err
	}
	if err := svc.lblSync(kt, client, &lbInfo.BaseLoadBalancer, []string{lblInfo.CloudID}); err != nil {
		// 调用同步的方法内会打印错误，这里只标记调用方
		logs.Errorf("fail to sync listener for update domain(%s), lblID: %s, rid: %s",
			domainOpt.Domain, lblInfo.ID, kt.Rid)
		return err
	}
	return nil
}

// BatchDeleteTCloudLoadBalancer ...
func (svc *clbSvc) BatchDeleteTCloudLoadBalancer(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudBatchDeleteLoadbalancerReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.LoadBalancer.ListLoadBalancer(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request data service list tcloud loadBalancer failed, err: %v, ids: %v, rid: %s",
			err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	delCloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		delCloudIDs = append(delCloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typelb.TCloudDeleteOption{
		Region:   req.Region,
		CloudIDs: delCloudIDs,
	}
	if err = client.DeleteLoadBalancer(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to delete tcloud loadBalancer failed, err: %v, opt: %v, rid: %s", err, opt,
			cts.Kit.Rid)
		return nil, err
	}

	delReq := &dataproto.LoadBalancerBatchDeleteReq{
		Filter: tools.ContainersExpression("id", req.IDs),
	}
	if err = svc.dataCli.Global.LoadBalancer.BatchDeleteLoadBalancer(cts.Kit, delReq); err != nil {
		logs.Errorf("request data service delete tcloud loadBalancer failed, err: %v, ids: %v, rid: %s", err, req.IDs,
			cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// InquiryPriceTCloudLB inquiry price tcloud clb.
func (svc *clbSvc) InquiryPriceTCloudLB(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudLoadBalancerCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(false); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	adaptor, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}
	if req.AddressIPVersion == "" {
		req.AddressIPVersion = typelb.IPV4IPVersion
	}
	createOpt := &typelb.TCloudCreateClbOption{
		Region:           req.Region,
		LoadBalancerType: req.LoadBalancerType,
		LoadBalancerName: req.Name,
		VpcID:            req.CloudVpcID,
		SubnetID:         req.CloudSubnetID,
		Vip:              req.Vip,
		VipIsp:           req.VipIsp,
		AddressIPVersion: req.AddressIPVersion,

		InternetChargeType:      req.InternetChargeType,
		InternetMaxBandwidthOut: req.InternetMaxBandwidthOut,
		BandwidthpkgSubType:     req.BandwidthpkgSubType,

		BandwidthPackageID: req.BandwidthPackageID,
		SlaType:            req.SlaType,
		Number:             req.RequireCount,
		ClientToken:        cvt.StrNilPtr(cts.Kit.Rid),
	}
	if cvt.PtrToVal(req.CloudEipID) != "" {
		createOpt.EipAddressID = req.CloudEipID
	}
	// 负载均衡实例的网络类型-公网属性
	if req.LoadBalancerType == typelb.OpenLoadBalancerType {
		// 静态单线IP 线路类型-仅适用于公网负载均衡, 如果不指定本参数，则默认使用BGP
		createOpt.VipIsp = req.VipIsp

		// 设置跨可用区容灾时的可用区ID-仅适用于公网负载均衡
		if len(req.BackupZones) > 0 && len(req.Zones) > 0 {
			// 主备可用区，传递zones（单元素数组），以及backup_zones
			createOpt.MasterZoneID = cvt.ValToPtr(req.Zones[0])
			createOpt.SlaveZoneID = cvt.ValToPtr(req.BackupZones[0])
		} else if len(req.Zones) > 0 {
			// 单可用区
			createOpt.ZoneID = cvt.ValToPtr(req.Zones[0])
		}
	}
	result, err := adaptor.InquiryPriceLoadBalancer(cts.Kit, createOpt)
	if err != nil {
		logs.Errorf("inquiry load balancer price failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}

// ListTCloudLBQuota list tcloud clb quota.
func (svc *clbSvc) ListTCloudLBQuota(cts *rest.Contexts) (any, error) {
	req := new(protolb.TCloudListLoadBalancerQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	adaptor, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	result, err := adaptor.ListLoadBalancerQuota(cts.Kit, &typelb.ListTCloudLoadBalancerQuotaOption{
		Region: req.Region,
	})
	if err != nil {
		logs.Errorf("list tcloud load balancer quota failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return result, nil
}
