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

package handler

import (
	"errors"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

func checkBizMatch(urlBizID int64, resBizID int64, usageBizIDs []int64) bool {
	// 检查分配业务和访问业务是否匹配
	if resBizID == 0 || resBizID == urlBizID {
		return true
	}

	// 检查使用业务和访问业务是否匹配
	if len(usageBizIDs) > 0 {
		for _, id := range usageBizIDs {
			// 使用业务为全部业务，均可访问
			if id == constant.AttachedAllBiz {
				return true
			}
			if id == urlBizID {
				return true
			}
		}
	}

	return false
}

// BizOperateAuth 业务下校验操作合法性, 校验传入的basicInfo是否在url中的业务下
func BizOperateAuth(cts *rest.Contexts, opt *ValidWithAuthOption) error {
	bizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return err
	}

	if bizID <= 0 {
		return errf.New(errf.InvalidParameter, "resource is not in biz")
	}

	// authorize one resource
	if opt.BasicInfo != nil {
		if opt.BasicInfos == nil {
			opt.BasicInfos = map[string]types.CloudResourceBasicInfo{}
		}
		opt.BasicInfos[opt.BasicInfo.ID] = *opt.BasicInfo
	}
	// batch authorize resource
	total := len(opt.BasicInfos)
	if total == 0 {
		return errors.New("no resource to authorize")
	}
	authRes := make([]meta.ResourceAttribute, 0, total)
	notMatchedIDs, recycledIDs, notRecycledIDS := make([]string, 0), make([]string, 0, total), make([]string, 0, total)
	unAssignIDs := make([]string, 0)
	for id, info := range opt.BasicInfos {
		if !checkBizMatch(bizID, info.BkBizID, info.UsageBizIDs) {
			if info.BkBizID == constant.UnassignedBiz {
				unAssignIDs = append(unAssignIDs, id)
			} else {
				notMatchedIDs = append(notMatchedIDs, id)
			}
		}
		if info.RecycleStatus == enumor.RecycleStatus {
			recycledIDs = append(recycledIDs, id)
		} else {
			notRecycledIDS = append(notRecycledIDS, id)
		}

		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: opt.ResType, Action: opt.Action},
			BizID: bizID})
	}

	if !opt.DisableBizIDEqual && len(notMatchedIDs) > 0 {
		return errf.Newf(errf.PermissionDenied, "resources(ids: %+v) not matches url biz", notMatchedIDs)
	}
	if len(unAssignIDs) > 0 {
		return errf.Newf(errf.ResourceUnassigned, "resources(ids: %+v) are unassigned", unAssignIDs)
	}

	// 恢复或删除已回收资源, 要求资源必须在已回收状态下
	if opt.Action == meta.Destroy || opt.Action == meta.Recover {
		if len(notRecycledIDS) > 0 {
			return errf.Newf(errf.InvalidParameter, "resources(ids: %+v) are not in recycle bin", notRecycledIDS)
		}
	} else {
		// 其他操作要求资源不能在回收状态下
		if len(recycledIDs) > 0 {
			return errf.Newf(errf.InvalidParameter, "resources(ids: %+v) are in recycle bin", recycledIDs)
		}
	}

	return opt.Authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
}

// ResOperateAuth 资源下校验操作合法性
func ResOperateAuth(cts *rest.Contexts, opt *ValidWithAuthOption) error {
	// authorize one resource
	if opt.BasicInfo != nil {
		if opt.BasicInfos == nil {
			opt.BasicInfos = map[string]types.CloudResourceBasicInfo{}
		}
		opt.BasicInfos[opt.BasicInfo.ID] = *opt.BasicInfo
	}

	total := len(opt.BasicInfos)
	if total == 0 {
		return errors.New("no resource to authorize")
	}
	// batch authorize resource
	authRes := make([]meta.ResourceAttribute, 0, total)
	typeAssignedIDMap, recycledIDs, notRecycledIDS := make(map[enumor.CloudResourceType][]string),
		make([]string, 0, total), make([]string, 0, total)
	for id, info := range opt.BasicInfos {
		// validate if resource is not in biz for write operation
		if opt.Action != meta.Find && info.BkBizID != constant.UnassignedBiz && info.BkBizID != 0 {
			typeAssignedIDMap[info.ResType] = append(typeAssignedIDMap[info.ResType], id)
		}

		if info.RecycleStatus == enumor.RecycleStatus {
			recycledIDs = append(recycledIDs, id)
		} else {
			notRecycledIDS = append(notRecycledIDS, id)
		}

		authRes = append(authRes, meta.ResourceAttribute{Basic: &meta.Basic{Type: opt.ResType, Action: opt.Action,
			ResourceID: info.AccountID}})
	}

	// 资源下，不允许操作业务下的资源
	if len(typeAssignedIDMap) > 0 {
		return errf.Newf(errf.InvalidParameter, "resources(%v) are already assigned", typeAssignedIDMap)
	}

	// 恢复或删除已回收资源, 要求资源必须在已回收状态下
	if opt.Action == meta.Destroy || opt.Action == meta.Recover {
		if len(notRecycledIDS) > 0 {
			return errf.Newf(errf.InvalidParameter, "resources(ids: %+v) are not in recycle bin", notRecycledIDS)
		}
	} else {
		// 其他操作要求资源不能在回收状态下
		if len(recycledIDs) > 0 {
			return errf.Newf(errf.InvalidParameter, "resources(ids: %+v) are in recycle bin", recycledIDs)
		}
	}

	return opt.Authorizer.AuthorizeWithPerm(cts.Kit, authRes...)
}
