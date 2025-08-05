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

package tcloud

import (
	typescos "hcm/pkg/adaptor/types/cos"
	"hcm/pkg/api/core"
	corecos "hcm/pkg/api/core/cloud/cos"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func (cli *client) Cos(kt *kit.Kit, params *SyncBaseParams, opt *SyncCosOption) (*SyncResult,
	error) {
	// cosFromCloud, err := cli.listCosFromCloud(kt, params)
	// if err != nil {
	// 	return nil, fmt.Errorf("sync cos from tcloud failed:%s", err.Error())
	// }

	// cosFromDB, err := cli.dbCli.TCloud.Cos.List(kt, &core.ListReq{
	// 	Filter: &filter.Expression{
	// 		Op: filter.And,
	// 		Rules: []filter.RuleFactory{
	// 			&filter.AtomRule{
	// 				Field: "region",
	// 				Op:    filter.Equal.Factory(),
	// 				Value: params.Region,
	// 			},
	// 		},
	// 	},
	// })
	//isCosChange := false
	//addList, updateList, deleteList := common.Diff[typescos.Bucket, 
	// corecos.TCloudCos](cosFromCloud, cosFromDB, isCosChange)
	return &SyncResult{}, nil
}

// SyncCosOption sync cos option
type SyncCosOption struct {
	Name string
}

// listCosFromDB list cos from database
func (cli *client) listCosFromDB(kt *kit.Kit, params *SyncBaseParams) ([]corecos.TCloudCos, error) {

	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("account_id", params.AccountID),
			tools.RuleEqual("region", params.Region),
		),
		Page: core.NewDefaultBasePage(),
	}
	// 支持查询所有，或者指定cloud_id
	if len(params.CloudIDs) > 0 {
		req.Filter.Rules = append(req.Filter.Rules, tools.RuleIn("cloud_id", params.CloudIDs))
	}
	result, err := cli.dbCli.TCloud.Cos.List(kt, req)
	if err != nil {
		logs.Errorf("[%s] list cos from db failed, err: %v, account: %s, req: %v, rid: %s",
			enumor.TCloud, err, params.AccountID, req, kt.Rid)
		return nil, err
	}

	return result.Details, nil
}

// listCosFromCloud list load balancer from cloud vendor
func (cli *client) listCosFromCloud(kt *kit.Kit, params *SyncBaseParams) (result []typescos.Bucket, err error) {
	opt := &typescos.TCloudBucketListOption{
		Region: &params.Region,
	}
	// 指定id时一次只能查询20个
	batch, err := cli.cloudCli.ListBuckets(kt, opt)
	if err != nil {
		logs.Errorf("[%s] list cos from cloud failed, err: %v, account: %s, opt: %v, rid: %s",
			enumor.TCloud, err, params.AccountID, opt, kt.Rid)
		return nil, err
	}
	result = append(result, batch.Buckets...)
	return result, nil

}
