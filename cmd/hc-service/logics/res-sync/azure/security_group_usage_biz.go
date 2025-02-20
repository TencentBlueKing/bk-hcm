/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package azure

import (
	usagebizrelmgr "hcm/cmd/hc-service/logics/res-sync/usage-biz-rel"
	cloudcore "hcm/pkg/api/core/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SecurityGroupUsageBiz 通过安全组关联资源的业务ID，更新安全组使用业务ID
func (cli *client) SecurityGroupUsageBiz(kt *kit.Kit, params *SyncSGUsageBizParams) error {

	mgr := usagebizrelmgr.NewUsageBizRelManager(cli.dbCli)

	for i := range params.SGList {
		sg := &params.SGList[i]
		err := mgr.SyncSecurityGroupUsageBiz(kt, sg)
		if err != nil {
			logs.Errorf("sync azure security group usage biz failed, err: %v, sg: %+v, rid: %s", err, sg, kt.Rid)
			return err
		}
	}
	return nil
}

// SyncSGUsageBizParams 同步安全组使用业务参数，使用业务只依赖本地数据
type SyncSGUsageBizParams struct {
	AccountID         string
	ResourceGroupName string
	SGList            []cloudcore.BaseSecurityGroup
}
