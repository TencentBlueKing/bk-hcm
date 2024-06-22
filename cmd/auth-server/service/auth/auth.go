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

// Package auth ...
package auth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"hcm/cmd/auth-server/options"
	"hcm/cmd/auth-server/service/capability"
	authserver "hcm/pkg/api/auth-server"
	"hcm/pkg/api/core"
	dsproto "hcm/pkg/api/data-service"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/meta"
	"hcm/pkg/iam/sdk/auth"
	"hcm/pkg/iam/sys"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// Auth related operate.
type Auth struct {
	// auth related operate.
	auth auth.Authorizer
	// ds data service's auth related api.
	ds *dataservice.Client
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
	// disableWriteOpt defines which biz's write operation needs to be disabled
	disableWriteOpt *options.DisableWriteOption
	// esb client.
	esbCli esb.Client
}

// NewAuth new auth.
func NewAuth(auth auth.Authorizer, ds *dataservice.Client, disableAuth bool, esbCli esb.Client,
	disableWriteOpt *options.DisableWriteOption) (*Auth, error) {

	if auth == nil {
		return nil, errf.New(errf.InvalidParameter, "auth is nil")
	}

	if ds == nil {
		return nil, errf.New(errf.InvalidParameter, "data client is nil")
	}

	if disableWriteOpt == nil {
		return nil, errf.New(errf.InvalidParameter, "disable write operation is nil")
	}

	i := &Auth{
		auth:            auth,
		ds:              ds,
		disableAuth:     disableAuth,
		disableWriteOpt: disableWriteOpt,
		esbCli:          esbCli,
	}

	return i, nil
}

// InitAuthService initialize the iam authorize service
func (a *Auth) InitAuthService(c *capability.Capability) {
	h := rest.NewHandler()

	h.Add("AuthorizeBatch", "POST", "/auth/authorize/batch", a.AuthorizeBatch)
	h.Add("AuthorizeAnyBatch", "POST", "/auth/authorize/any/batch", a.AuthorizeAnyBatch)
	h.Add("GetPermissionToApply", "POST", "/auth/find/permission_to_apply", a.GetPermissionToApply)
	h.Add("GetApplyPermUrl", "POST", "/auth/find/apply_perm_url", a.GetApplyPermUrl)
	h.Add("ListAuthorizedInstances", "POST", "/auth/list/authorized_resource", a.ListAuthorizedInstances)
	h.Add("RegisterResCreatorAction", "POST", "/auth/register/resource_create_action", a.RegisterResourceCreatorAction)

	h.Load(c.WebService)
}

// AuthorizeBatch authorize resource batch.
func (a *Auth) AuthorizeBatch(cts *rest.Contexts) (interface{}, error) {
	req := new(authserver.AuthorizeBatchReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	return a.authorizeBatch(cts.Kit, req, true)
}

// AuthorizeAnyBatch batch authorize if resource has any permission.
func (a *Auth) AuthorizeAnyBatch(cts *rest.Contexts) (interface{}, error) {
	req := new(authserver.AuthorizeBatchReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	return a.authorizeBatch(cts.Kit, req, false)
}

// AuthorizeBatch authorize resource batch.
func (a *Auth) authorizeBatch(kt *kit.Kit, req *authserver.AuthorizeBatchReq, exact bool) ([]meta.Decision, error) {
	if len(req.Resources) == 0 {
		return make([]meta.Decision, 0), nil
	}

	// if write operations are disabled, returns corresponding error
	if err := a.isWriteOperationDisabled(kt, req.Resources); err != nil {
		return nil, err
	}

	// if auth is disabled, returns authorized for all request resources
	if a.disableAuth {
		decisions := make([]meta.Decision, len(req.Resources))
		for index := range req.Resources {
			decisions[index] = meta.Decision{Authorized: true}
		}
		return decisions, nil
	}

	// parse hcm resource to iam resource
	opts, decisions, err := parseAttributesToBatchOptions(kt, req.User, req.Resources...)
	if err != nil {
		return nil, err
	}

	// all resources are skipped
	if opts == nil {
		return decisions, nil
	}

	// do authentication
	var authDecisions []*client.Decision
	if exact {
		authDecisions, err = a.auth.AuthorizeBatch(kt.Ctx, opts)
		if err != nil {
			logs.Errorf("authorize batch failed, err: %v ,ops: %#v, req: %#v, rid: %s", err, opts, req, kt.Rid)
			return nil, err
		}
	} else {
		authDecisions, err = a.auth.AuthorizeAnyBatch(kt.Ctx, opts)
		if err != nil {
			logs.Errorf("authorize any batch failed, err: %v, ops: %#v, req: %#v, rid: %s", err, opts, req, kt.Rid)
			return nil, err
		}
	}

	index := 0
	decisionLen := len(decisions)
	for _, decision := range authDecisions {
		// skip resources' decisions are already set as authorized
		for index < decisionLen && decisions[index].Authorized {
			index++
		}

		if index >= decisionLen {
			break
		}

		decisions[index].Authorized = decision.Authorized
		index++
	}

	return decisions, nil
}

func (a *Auth) isWriteOperationDisabled(kt *kit.Kit, resources []meta.ResourceAttribute) error {
	if !a.disableWriteOpt.IsDisabled {
		return nil
	}

	for _, resource := range resources {
		action := resource.Basic.Action
		if action == meta.Find || action == meta.SkipAction {
			continue
		}

		if a.disableWriteOpt.IsAll {
			logs.Errorf("all %s operation is disabled, rid: %s", action, kt.Rid)
			return errf.New(errf.Aborted, "hcm server is publishing, wring operation is not allowed")
		}

		bizID := resource.BizID
		if _, exists := a.disableWriteOpt.BizIDMap.Load(bizID); exists {
			logs.Errorf("biz id %d %s operation is disabled, rid: %s", bizID, action, kt.Rid)
			return errf.New(errf.Aborted, "hcm server is publishing, wring operation is not allowed")
		}
	}

	return nil
}

// parseAttributesToBatchOptions parse auth attributes to authorize batch options
func parseAttributesToBatchOptions(kt *kit.Kit, user *meta.UserInfo, resources ...meta.ResourceAttribute) (
	*client.AuthBatchOptions, []meta.Decision, error) {

	authBatchArr := make([]client.AuthBatch, 0)
	decisions := make([]meta.Decision, len(resources))
	for index, resource := range resources {
		decisions[index] = meta.Decision{Authorized: false}

		// this resource should be skipped, do not need to verify in auth center.
		if resource.Basic.Action == meta.SkipAction {
			decisions[index].Authorized = true
			logs.V(5).Infof("skip authorization for resource: %+v, rid: %s", resource, kt.Rid)
			continue
		}

		action, iamResources, err := AdaptAuthOptions(&resource)
		if err != nil {
			logs.Errorf("adapt hcm resource to iam failed, err: %s, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		// this resource should be skipped, do not need to verify in auth center.
		if action == sys.Skip {
			decisions[index].Authorized = true
			logs.V(5).Infof("skip authorization for resource: %+v, rid: %s", resource, kt.Rid)
			continue
		}

		authBatchArr = append(authBatchArr, client.AuthBatch{
			Action:    client.Action{ID: string(action)},
			Resources: iamResources,
		})
	}

	// all resources are skipped
	if len(authBatchArr) == 0 {
		return nil, decisions, nil
	}

	ops := &client.AuthBatchOptions{
		System: sys.SystemIDHCM,
		Subject: client.Subject{
			Type: sys.UserSubjectType,
			ID:   user.UserName,
		},
		Batch: authBatchArr,
	}
	return ops, decisions, nil
}

// GetPermissionToApply get iam permission to apply when user has no permission to some resources.
func (a *Auth) GetPermissionToApply(cts *rest.Contexts) (interface{}, error) {
	req := new(authserver.GetPermissionToApplyReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	permission, err := a.getPermissionToApply(cts.Kit, req.Resources)
	if err != nil {
		return nil, err
	}

	return permission, nil
}

func (a *Auth) getPermissionToApply(kt *kit.Kit, resources []meta.ResourceAttribute) (*meta.IamPermission, error) {
	permission := new(meta.IamPermission)
	permission.SystemID = sys.SystemIDHCM
	permission.SystemName = sys.SystemNameHCM

	// parse hcm auth resource
	resTypeIDsMap, permissionMap, err := a.parseResources(kt, resources)
	if err != nil {
		logs.Errorf("get inst ID and name map failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get hcm resource name by id, then assign it to corresponding iam auth resource
	instIDNameMap, err := a.getInstIDNameMap(kt, resTypeIDsMap)
	if err != nil {
		return nil, err
	}

	for actionID, permissionTypeMap := range permissionMap {
		action := meta.IamAction{
			ID:                   string(actionID),
			Name:                 sys.ActionIDNameMap[actionID],
			RelatedResourceTypes: make([]meta.IamResourceType, 0),
		}

		for rscType := range permissionTypeMap {
			iamResourceType := permissionTypeMap[rscType]

			for idx, resources := range iamResourceType.Instances {
				for idx2, resource := range resources {
					iamResourceType.Instances[idx][idx2].Name = instIDNameMap[client.TypeID(resource.Type)][resource.ID]
				}
			}

			action.RelatedResourceTypes = append(action.RelatedResourceTypes, iamResourceType)
		}
		permission.Actions = append(permission.Actions, action)
	}

	return permission, nil
}

// parseResources parse hcm auth resource to iam permission resources in organized way
func (a *Auth) parseResources(kt *kit.Kit, resources []meta.ResourceAttribute) (map[client.TypeID][]string,
	map[client.ActionID]map[client.TypeID]meta.IamResourceType, error) {

	// resTypeIDsMap maps resource type to resource ids to get resource names.
	resTypeIDsMap := make(map[client.TypeID][]string)
	// permissionMap maps ActionID and TypeID to ResourceInstances
	permissionMap := make(map[client.ActionID]map[client.TypeID]meta.IamResourceType, 0)

	for _, r := range resources {
		// parse hcm auth resource to iam action id and iam resources
		actionID, resources, err := AdaptAuthOptions(&r)
		if err != nil {
			logs.Errorf("adaptor hcm resource to iam failed, err: %s, rid: %s", err, kt.Rid)
			return nil, nil, err
		}

		if _, ok := permissionMap[actionID]; !ok {
			permissionMap[actionID] = make(map[client.TypeID]meta.IamResourceType, 0)
		}

		// generate iam resource resources by its paths and itself
		for _, res := range resources {
			if len(res.ID) == 0 && res.Attribute == nil {
				resType, exists := permissionMap[actionID][res.Type]
				if !exists {
					resType = meta.IamResourceType{
						SystemID:   res.System,
						SystemName: sys.SystemIDNameMap[res.System],
						Type:       string(res.Type),
						TypeName:   sys.ResourceTypeIDMap[res.Type],
						Instances:  make([][]meta.IamResourceInstance, 0),
					}
				}
				permissionMap[actionID][res.Type] = resType
				continue
			}

			resTypeIDsMap[res.Type] = append(resTypeIDsMap[res.Type], res.ID)

			resource := make([]meta.IamResourceInstance, 0)
			if res.Attribute != nil {
				// parse hcm auth resource iam path attribute to iam ancestor resources
				iamPath, ok := res.Attribute[client.IamPathKey].([]string)
				if !ok {
					return nil, nil, fmt.Errorf("iam path(%v) is not string array", res.Attribute[client.IamPathKey])
				}

				ancestors, err := a.parseIamPathToAncestors(iamPath)
				if err != nil {
					return nil, nil, err
				}
				resource = append(resource, ancestors...)

				// record ancestor resource ids to get names from them afterwards
				for _, ancestor := range ancestors {
					ancestorType := client.TypeID(ancestor.Type)
					resTypeIDsMap[ancestorType] = append(resTypeIDsMap[ancestorType], ancestor.ID)
				}
			}

			// add iam resource of auth resource to the related iam resources after its ancestors
			resource = append(resource,
				meta.IamResourceInstance{Type: string(res.Type), TypeName: sys.ResourceTypeIDMap[res.Type], ID: res.ID})

			resType, exists := permissionMap[actionID][res.Type]
			if !exists {
				resType = meta.IamResourceType{
					SystemID:   res.System,
					SystemName: sys.SystemIDNameMap[res.System],
					Type:       string(res.Type),
					TypeName:   sys.ResourceTypeIDMap[res.Type],
					Instances:  make([][]meta.IamResourceInstance, 0),
				}
			}
			resType.Instances = append(resType.Instances, resource)
			permissionMap[actionID][res.Type] = resType
		}
	}

	return resTypeIDsMap, permissionMap, nil
}

// parseIamPathToAncestors parse iam path to resource's ancestor resources
func (a *Auth) parseIamPathToAncestors(iamPath []string) ([]meta.IamResourceInstance, error) {
	resources := make([]meta.IamResourceInstance, 0)
	for _, path := range iamPath {
		pathItemArr := strings.Split(strings.Trim(path, "/"), "/")
		for _, pathItem := range pathItemArr {
			typeAndID := strings.Split(pathItem, ",")
			if len(typeAndID) != 2 {
				return nil, fmt.Errorf("pathItem %s invalid", pathItem)
			}
			id := typeAndID[1]
			if id == "*" {
				continue
			}
			resources = append(resources, meta.IamResourceInstance{
				Type:     typeAndID[0],
				TypeName: sys.ResourceTypeIDMap[client.TypeID(typeAndID[0])],
				ID:       id,
			})
		}
	}
	return resources, nil
}

// getInstIDNameMap get resource id to name map by resource ids, groups by resource type
func (a *Auth) getInstIDNameMap(kt *kit.Kit, resTypeIDsMap map[client.TypeID][]string) (
	map[client.TypeID]map[string]string, error) {

	instMap := make(map[client.TypeID]map[string]string)

	for resType, ids := range resTypeIDsMap {
		if resType == sys.Biz {
			bizIDNameMap, err := a.getBizIDNameMap(kt, ids)
			if err != nil {
				return nil, err
			}
			instMap[resType] = bizIDNameMap
			continue
		}

		req := &dsproto.ListInstancesReq{
			ResourceType: resType,
			Filter:       tools.ContainersExpression("id", ids),
			Page:         &core.BasePage{Count: false, Limit: uint(len(ids))},
		}

		resp, err := a.ds.Global.Auth.ListInstances(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list auth instances failed, err: %v, type: %s, ids: %+v, rid: %s", err, resType, ids, kt.Rid)
			return nil, err
		}

		instMap[resType] = make(map[string]string)

		for _, inst := range resp.Details {
			instMap[resType][inst.ID] = inst.DisplayName
		}

	}

	return instMap, nil
}

// getBizIDNameMap get cmdb biz resource id to name map by resource ids
func (a *Auth) getBizIDNameMap(kt *kit.Kit, rawIDs []string) (map[string]string, error) {
	ids := make([]int64, len(rawIDs))
	for i, id := range rawIDs {
		intID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, errf.Newf(errf.InvalidParameter, "parse id %s failed, err: %v", id, err)
		}
		ids[i] = intID
	}

	bizReq := &cmdb.SearchBizParams{
		Fields: []string{"bk_biz_id", "bk_biz_name"},
		BizPropertyFilter: &cmdb.QueryFilter{
			Rule: &cmdb.CombinedRule{
				Condition: cmdb.ConditionAnd,
				Rules: []cmdb.Rule{
					&cmdb.AtomRule{
						Field:    "bk_biz_id",
						Operator: cmdb.OperatorIn,
						Value:    ids,
					},
				},
			},
		},
	}
	bizResp, err := a.esbCli.Cmdb().SearchBusiness(kt, bizReq)
	if err != nil {
		return nil, err
	}

	idNameMap := make(map[string]string)
	for _, biz := range bizResp.Info {
		idNameMap[strconv.FormatInt(biz.BizID, 10)] = biz.BizName
	}

	return idNameMap, nil
}

// ListAuthorizedInstances list authorized instances info.
func (a *Auth) ListAuthorizedInstances(cts *rest.Contexts) (interface{}, error) {
	if a.disableAuth {
		return client.AuthorizeList{IsAny: true}, nil
	}

	req := new(authserver.ListAuthorizedInstancesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	res := &meta.ResourceAttribute{
		Basic: &meta.Basic{
			Type:   req.Type,
			Action: req.Action,
		},
	}
	actionID, resources, err := AdaptAuthOptions(res)
	if err != nil {
		return nil, err
	}

	if len(resources) != 1 {
		logs.Errorf("auth resources(%+v) length is invalid, req: %+v, rid: %s", resources, req, cts.Kit.Rid)
		return nil, errors.New("auth resources length is not 1, cannot list authorized instances")
	}

	ops := &client.AuthOptions{
		System: sys.SystemIDHCM,
		Subject: client.Subject{
			Type: sys.UserSubjectType,
			ID:   req.User.UserName,
		},
		Action: client.Action{
			ID: string(actionID),
		},
		Resources: resources,
	}
	authorizeList, err := a.auth.ListAuthorizedInstances(cts.Kit.Ctx, ops, resources[0].Type)
	if err != nil {
		logs.Errorf("list authorized instances failed, err: %v,  ops: %+v, req: %+v, rid: %s", err, ops, req,
			cts.Kit.Rid)
		return nil, err
	}

	return authorizeList, nil
}

// RegisterResourceCreatorAction registers iam resource instance so that creator will be authorized on related actions
// 注册创建者权限, 一个资源的创建者可以拥有这个资源的关联权限，如编辑和删除权限
func (a *Auth) RegisterResourceCreatorAction(cts *rest.Contexts) (interface{}, error) {
	req := new(authserver.RegisterResourceCreatorActionReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	opts := &client.InstanceWithCreator{
		System:  sys.SystemIDHCM,
		Type:    req.Instance.Type,
		ID:      req.Instance.ID,
		Name:    req.Instance.Name,
		Creator: req.Creator,
	}

	for _, ancestor := range req.Instance.Ancestors {
		opts.Ancestors = append(opts.Ancestors, client.InstanceAncestor{
			System: sys.SystemIDHCM,
			Type:   ancestor.Type,
			ID:     ancestor.ID,
		})
	}

	policies, err := a.auth.RegisterResourceCreatorAction(cts.Kit.Ctx, opts)
	if err != nil {
		logs.Errorf("list authorized instances failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return policies, nil
}

// GetApplyPermUrl get iam apply permission url.
func (a *Auth) GetApplyPermUrl(cts *rest.Contexts) (interface{}, error) {
	req := new(meta.IamPermission)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	url, err := a.auth.GetApplyPermUrl(cts.Kit.Ctx, req)
	if err != nil {
		logs.Errorf("get iam apply permission url failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return url, nil
}
