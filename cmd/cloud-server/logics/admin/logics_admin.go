/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

package logicsadmin

import (
	"encoding/json"
	"fmt"

	apisysteminit "hcm/pkg/api/cloud-server/system-init"
	"hcm/pkg/api/core"
	gccore "hcm/pkg/api/core/global-config"
	proto "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	datagconf "hcm/pkg/api/data-service/global_config"
	"hcm/pkg/api/data-service/tenant"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/bkuser"
	"hcm/pkg/thirdparty/api-gateway/itsm"
	cvt "hcm/pkg/tools/converter"
)

// Interface admin logic interface
type Interface interface {
	InitVendorOtherAccount(kt *kit.Kit) (*apisysteminit.OtherAccountInitResult, error)
	GetTenantFromBkUser(kt *kit.Kit) (*bkuser.Tenant, error)
	UpsertLocalTenant(kt *kit.Kit, targetTenant *bkuser.Tenant) (message string, err error)
	InitItsmProcess(kt *kit.Kit, systemID string) error
}

type admin struct {
	c       *client.ClientSet
	bkUser  bkuser.Client
	itsmCli itsm.Client
}

// NewAdminLogic new admin logic
func NewAdminLogic(c *client.ClientSet, userClient bkuser.Client, itsmCli itsm.Client) Interface {
	return &admin{c: c, bkUser: userClient, itsmCli: itsmCli}
}

// InternalOtherVendorAccountName 内置账号名称
const InternalOtherVendorAccountName = "内置账号"

// InitVendorOtherAccount 查找是否存在vendor为other的账号，若有则返回，没有则创建
func (a *admin) InitVendorOtherAccount(kt *kit.Kit) (*apisysteminit.OtherAccountInitResult, error) {

	listReq := &core.ListReq{
		Filter: tools.EqualExpression("vendor", enumor.Other),
		Page:   core.NewDefaultBasePage(),
	}
	accResp, err := a.c.DataService().Global.Account.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("fail to list other vendor account, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(accResp.Details) > 0 {
		return &apisysteminit.OtherAccountInitResult{ExistsAccountID: accResp.Details[0].ID}, nil
	}

	// 创建other vendor用户
	createReq := &protocloud.AccountCreateReq[protocloud.OtherAccountExtensionCreateReq]{
		Name:     InternalOtherVendorAccountName,
		Managers: []string{"admin"},
		Type:     enumor.ResourceAccount,
		Site:     enumor.InternationalSite,
		Memo:     cvt.ValToPtr(InternalOtherVendorAccountName),
		Extension: &protocloud.OtherAccountExtensionCreateReq{
			CloudID:     string(enumor.Other),
			CloudSecKey: "",
		},
		UsageBizIDs: []int64{constant.AttachedAllBiz},
	}
	createResp, err := a.c.DataService().Other.Account.Create(kt, createReq)
	if err != nil {
		logs.Errorf("fail to create other vendor account, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &apisysteminit.OtherAccountInitResult{CreatedAccountID: createResp.ID}, nil
}

// UpsertLocalTenant 插入或更新租户信息
func (a *admin) UpsertLocalTenant(kt *kit.Kit, targetTenant *bkuser.Tenant) (message string, err error) {
	listReq := &core.ListReq{
		Filter: tools.EqualExpression("tenant_id", kt.TenantID),
		Page:   core.NewDefaultBasePage(),
	}
	localTenantResp, err := a.c.DataService().Global.Tenant.List(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list local tenant, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if len(localTenantResp.Details) > 0 {
		// 2.1 存在则更新
		localTenant := localTenantResp.Details[0]
		status := targetTenant.GetStatus()
		// 	更新租户
		updateReq := &tenant.UpdateTenantReq{Items: []tenant.UpdateTenantField{{
			ID:     localTenant.ID,
			Status: status,
		}}}
		err := a.c.DataService().Global.Tenant.Update(kt, updateReq)
		if err != nil {
			return "", err
		}
		logs.Infof("tenant updated: %s, local id: %s, rid: %s", targetTenant.String(), localTenant.ID, kt.Rid)
		return fmt.Sprintf("tenant update success, %s", localTenant.ID), nil
	}

	// 2.2 不存在则创建
	createReq := &tenant.CreateTenantReq{
		Items: []tenant.CreateTenantField{{
			TenantID: kt.TenantID,
			Status:   targetTenant.GetStatus(),
		}},
	}
	created, err := a.c.DataService().Global.Tenant.Create(kt, createReq)
	if err != nil {
		return "", err
	}
	if len(created.IDs) < 1 {
		return "", fmt.Errorf("tenant created but no any id has been returned")
	}
	createdID := created.IDs[0]
	logs.Infof("tenant created: %s, local id: %s, rid: %s", targetTenant.String(), createdID, kt.Rid)
	return fmt.Sprintf("tenant create success, %s", createdID), nil
}

// GetTenantFromBkUser 尝试从bk user获取租户信息
func (a *admin) GetTenantFromBkUser(kt *kit.Kit) (*bkuser.Tenant, error) {
	if !cc.TenantEnable() {
		return nil, fmt.Errorf("tenant is not enabled")
	}

	// 1. 查找是否是合法租户
	tenantResult, err := a.bkUser.ListTenant(kt)
	if err != nil {
		logs.Errorf("fail to list tenant by bk user, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tenantList := tenantResult.Data
	var targetTenant *bkuser.Tenant
	for _, t := range tenantList {
		if t.Id == kt.TenantID {
			targetTenant = cvt.ValToPtr(t)
			break
		}
	}
	if targetTenant == nil {
		logs.Infof("tenant not found by tenant id: %s, tenant list: %s, rid: %s",
			kt.TenantID, tenantList, kt.Rid)
		return nil, fmt.Errorf("invalid tenant: %s", kt.TenantID)
	}
	return targetTenant, nil
}

// InitItsmProcess 初始化itsm流程
func (a *admin) InitItsmProcess(kt *kit.Kit, systemID string) error {
	if err := a.migrateItsmTemplates(kt, systemID); err != nil {
		logs.Errorf("migrate itsm template failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// 单租户流程不会超过500个，不关心分页
	req := &proto.ApprovalProcessListReq{
		Filter: tools.AllExpression(),
		Page:   core.NewDefaultBasePage(),
	}
	existProcess, err := a.c.DataService().Global.ApprovalProcess.ListApprovalProcesses(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("fail to list approval process, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	processMap := make(map[string]struct{}, len(existProcess.Details))
	for _, process := range existProcess.Details {
		processMap[process.WorkflowKey] = struct{}{}
	}

	createItems := make([]proto.ApprovalProcessCreateReq, 0, len(enumor.ApplicationWorkflow))
	for applicationName := range enumor.ApplicationWorkflow {
		workflowKey := applicationName.WorkflowKey(kt.TenantID)
		if _, exists := processMap[workflowKey]; exists {
			continue
		}
		createItems = append(createItems, proto.ApprovalProcessCreateReq{
			ApplicationType: applicationName,
			WorkflowKey:     workflowKey,
			// TODO 目前无法自动获取租户下的管理员bk_username
			// Managers: "",
		})
	}
	if len(createItems) == 0 {
		return nil
	}
	batchCreateReq := &proto.ApprovalProcessBatchCreateReq{
		Items: createItems,
	}
	_, err = a.c.DataService().Global.ApprovalProcess.BatchCreateApprovalProcesses(kt.Ctx, kt.Header(), batchCreateReq)
	if err != nil {
		logs.Errorf("fail to batch create approval process, err: %v, tenant_id: %s, req: %+v, rid: %s", err,
			kt.TenantID, batchCreateReq.Items, kt.Rid)
		return err
	}

	return nil
}

// migrateItsmTemplates 按进度批量注册ITSM流程模板
func (a *admin) migrateItsmTemplates(kt *kit.Kit, systemID string) error {
	configKey := fmt.Sprintf("%s_%s", enumor.GlobalConfigKeyItsmMigrateVersionPrefix, kt.TenantID)

	lastApplied, configID, err := a.getItsmMigrateProgress(kt, configKey)
	if err != nil {
		logs.Errorf("get itsm migrate progress failed, config_key: %s, err: %v, rid: %s", configKey, err, kt.Rid)
		return err
	}

	startIdx := 0
	if lastApplied != "" {
		for i, tmpl := range itsm.MigrateTemplates {
			if tmpl.Name == lastApplied {
				startIdx = i + 1
				break
			}
		}
	}

	if startIdx >= len(itsm.MigrateTemplates) {
		logs.Infof("all itsm templates already migrated, tenant: %s, rid: %s", kt.TenantID, kt.Rid)
		return nil
	}

	for i := startIdx; i < len(itsm.MigrateTemplates); i++ {
		tmpl := itsm.MigrateTemplates[i]
		logs.Infof("migrating itsm template %s (%d/%d), rid: %s", tmpl.Name, i+1, len(itsm.MigrateTemplates), kt.Rid)

		if err = a.itsmCli.SystemMigrate(kt, systemID, tmpl.Content); err != nil {
			logs.Errorf("migrate itsm template %s failed, err: %v, rid: %s", tmpl.Name, err, kt.Rid)
			return err
		}

		configID, err = a.saveItsmMigrateProgress(kt, configKey, configID, tmpl.Name)
		if err != nil {
			logs.Errorf("save itsm migrate progress failed, template: %s, err: %v, rid: %s", tmpl.Name, err, kt.Rid)
			return err
		}
	}

	return nil
}

// getItsmMigrateProgress 从 global_config 读取当前租户的 ITSM 流程注册进度
func (a *admin) getItsmMigrateProgress(kt *kit.Kit, configKey string) (string, string, error) {
	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", string(enumor.GlobalConfigTypeITSM)),
			tools.RuleEqual("config_key", configKey),
		),
		Page: &core.BasePage{Limit: 1},
	}
	result, err := a.c.DataService().Global.GlobalConfig.List(kt, req)
	if err != nil {
		logs.Errorf("get itsm migrate progress failed, config_key: %s, err: %v, rid: %s", configKey, err, kt.Rid)
		return "", "", fmt.Errorf("list itsm migrate progress failed, err: %v", err)
	}
	if len(result.Details) == 0 {
		return "", "", nil
	}

	var templateName string
	if err := json.Unmarshal([]byte(result.Details[0].ConfigValue), &templateName); err != nil {
		logs.Errorf("unmarshal itsm migrate progress failed, config_key: %s, err: %v, rid: %s", configKey, err, kt.Rid)
		return "", "", fmt.Errorf("unmarshal itsm migrate progress failed, err: %v", err)
	}
	return templateName, result.Details[0].ID, nil
}

// saveItsmMigrateProgress 创建或更新 ITSM 迁移进度记录，返回最新的 configID
func (a *admin) saveItsmMigrateProgress(kt *kit.Kit, configKey, configID, templateName string) (string, error) {
	if configID == "" {
		createReq := &datagconf.BatchCreateReq{
			Configs: []gccore.GlobalConfig{{
				ConfigKey:   configKey,
				ConfigValue: templateName,
				ConfigType:  string(enumor.GlobalConfigTypeITSM),
			}},
		}
		created, err := a.c.DataService().Global.GlobalConfig.BatchCreate(kt, createReq)
		if err != nil {
			logs.Errorf("create itsm migrate progress failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}
		if len(created.IDs) == 0 {
			return "", fmt.Errorf("create itsm migrate progress but no id returned")
		}
		return created.IDs[0], nil
	}

	updateReq := &datagconf.BatchUpdateReq{
		Configs: []gccore.GlobalConfig{{
			ID:          configID,
			ConfigValue: templateName,
		}},
	}
	if err := a.c.DataService().Global.GlobalConfig.BatchUpdate(kt, updateReq); err != nil {
		logs.Errorf("update itsm migrate progress failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return configID, nil
}
