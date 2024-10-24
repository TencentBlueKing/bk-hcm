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

package tcloud

import (
	"sort"

	"hcm/cmd/hc-service/logics/res-sync/common"
	typeslb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataservice "hcm/pkg/api/data-service"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// lbSgRel 同步于安全组的关联关系，按lb、有序
func (cli *client) lbSgRel(kt *kit.Kit, params *SyncBaseParams, lbInfo []corelb.TCloudLoadBalancer) error {

	lbIDs := make([]string, 0, len(lbInfo))
	// lb cloud id -> lb local id
	cloudIDLbMap := make(map[string]string, len(lbInfo))
	for _, lb := range lbInfo {
		lbIDs = append(lbIDs, lb.ID)
		cloudIDLbMap[lb.CloudID] = lb.ID
	}
	// 1. 删除本地多余关联关系
	delFilter := tools.ExpressionAnd(
		tools.RuleEqual("res_type", enumor.LoadBalancerCloudResType),
		tools.RuleNotIn("res_id", lbIDs),
	)
	err := cli.dbCli.Global.SGCommonRel.BatchDeleteSgCommonRels(kt, &dataservice.BatchDeleteReq{Filter: delFilter})
	if err != nil {
		logs.Errorf("fail to del load balancer rel, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// 2. 获取云上安全组绑定信息
	sgCloudLocalIdMap, lbSgCloudMap, err := cli.getCloudLbSgBinding(kt, params, lbInfo, cloudIDLbMap)
	if err != nil {
		logs.Errorf("fail to get cloud lb sg bind for rel sync, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return cli.compareLbSgRel(kt, lbIDs, lbSgCloudMap, sgCloudLocalIdMap)
}

func (cli *client) compareLbSgRel(kt *kit.Kit, lbIDs []string, lbSgCloudMap map[string][]string,
	sgCloudLocalIdMap map[string]string) error {

	// 获取本地关联关系
	relReq := &protocloud.SGCommonRelWithSecurityGroupListReq{
		ResIDs:  lbIDs,
		ResType: enumor.LoadBalancerCloudResType,
	}
	sgRelResp, err := cli.dbCli.Global.SGCommonRel.ListWithSecurityGroup(kt, relReq)
	if err != nil {
		logs.Errorf("fail to list sg rel for lb sync, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// lb本地id-> 关联的本地sg列表
	lbLocalSgMap := make(map[string][]common.OrderedRel, len(*sgRelResp))
	for _, rel := range *sgRelResp {
		lbLocalSgMap[rel.ResID] = append(lbLocalSgMap[rel.ResID],
			common.OrderedRel{CloudResID: rel.CloudID, ResID: rel.ResID, Priority: rel.Priority})
	}
	// compare with priority
	for lbId, cloudSgList := range lbSgCloudMap {

		localSlice := lbLocalSgMap[lbId]
		// 按优先级从小到大排序
		sort.Slice(localSlice, func(i, j int) bool {
			return localSlice[i].Priority < localSlice[j].Priority
		})
		localLen := len(localSlice)
		cloudLen := len(cloudSgList)
		// 找到所有相等的列表
		var idx int
		var cloudID string
		var stayLocalIDs []string
		for ; idx < cloudLen; idx++ {
			cloudID = cloudSgList[idx]
			if idx >= localLen || localSlice[idx].CloudResID != cloudID || localSlice[idx].Priority != int64(idx+1) {
				// 剩下的全部加入新增列表里
				break
			}
			// 加入可以保留的安全组id列表中
			stayLocalIDs = append(stayLocalIDs, sgCloudLocalIdMap[cloudID])
		}
		err := cli.upsertSgRelForLb(kt, lbId, idx, stayLocalIDs, cloudSgList[idx:], sgCloudLocalIdMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cli *client) upsertSgRelForLb(kt *kit.Kit, lbId string, startIdx int, stayLocalIDs []string, sgCloudList []string,
	cloudSgMap map[string]string) error {

	createDel := &protocloud.SGCommonRelBatchUpsertReq{Rels: make([]protocloud.SGCommonRelCreate, 0)}
	// 删除所有不在给定id中的安全组，防止误删
	createDel.DeleteReq = &dataservice.BatchDeleteReq{Filter: tools.ExpressionAnd(
		tools.RuleEqual("res_type", enumor.LoadBalancerCloudResType),
		tools.RuleEqual("res_id", lbId),
	)}
	for i, cloudID := range sgCloudList {
		// 填充云上id
		createDel.Rels = append(createDel.Rels, protocloud.SGCommonRelCreate{
			SecurityGroupID: cloudSgMap[cloudID],
			Vendor:          enumor.TCloud,
			ResID:           lbId,
			ResType:         enumor.LoadBalancerCloudResType,
			Priority:        int64(i + startIdx + 1),
		})

	}
	if len(stayLocalIDs) > 0 {
		createDel.DeleteReq.Filter.Rules = append(createDel.DeleteReq.Filter.Rules,
			tools.RuleNotIn("security_group_id", stayLocalIDs))
	}
	if len(createDel.Rels) > 0 {
		// 同时需要删除和创建
		err := cli.dbCli.Global.SGCommonRel.BatchUpsertSgCommonRels(kt, createDel)
		if err != nil {
			logs.Errorf("fail to upsert lb(%s) security group rel, err: %v, req: %+v, rid: %s",
				lbId, err, createDel, kt.Rid)
			return err
		}
		return nil
	}

	// 只需要尝试删除多余关联关系即可
	err := cli.dbCli.Global.SGCommonRel.BatchDeleteSgCommonRels(kt, createDel.DeleteReq)
	if err != nil {
		logs.Errorf("fail to delete lb(%s) security group rel, err: %v, req: %+v, rid: %s",
			lbId, err, createDel.DeleteReq, kt.Rid)
		return err
	}

	return nil

}

func (cli *client) getCloudLbSgBinding(kt *kit.Kit, params *SyncBaseParams, lbInfo []corelb.TCloudLoadBalancer,
	cloudIDLbMap map[string]string) (sgCloudLocalMap map[string]string, lbSgCloudMap map[string][]string, err error) {

	lbCloudIDs := slice.Map(lbInfo, func(lb corelb.TCloudLoadBalancer) string { return lb.CloudID })
	cloudLBs, err := cli.listLBFromCloud(kt, &SyncBaseParams{
		AccountID: params.AccountID,
		Region:    params.Region,
		CloudIDs:  lbCloudIDs,
	})
	if err != nil {
		logs.Errorf("fail to list cloud load balancers for sg rel sync, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	// 1. 获取 负载均衡安全组关联关系，并组合安全组id列表
	allSgCloudIDs := make([]string, 0, len(lbCloudIDs))
	// lbLocalID-> cloud sg ids, 本地负载均衡id索引的，云上安全组id列表
	lbSgCloudMap = cvt.SliceToMap(cloudLBs, func(lb typeslb.TCloudClb) (string, []string) {
		cloudSlice := cvt.PtrToSlice(lb.SecureGroups)
		allSgCloudIDs = append(allSgCloudIDs, cloudSlice...)
		return cloudIDLbMap[lb.GetCloudID()], cloudSlice
	})
	if len(allSgCloudIDs) == 0 {
		return make(map[string]string), make(map[string][]string), nil
	}
	// 2. 获取本地id 映射
	sgReq := &protocloud.SecurityGroupListReq{
		Field:  []string{"id", "cloud_id"},
		Filter: tools.ExpressionAnd(tools.RuleIn("cloud_id", allSgCloudIDs)),
		Page:   core.NewDefaultBasePage(),
	}
	sgResp, err := cli.dbCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), sgReq)
	if err != nil {
		logs.Errorf("fail to get sg list, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}
	// cloudID->localID
	cloudSgMap := cvt.SliceToMap(sgResp.Details, func(sg cloud.BaseSecurityGroup) (string, string) {
		return sg.CloudID, sg.ID
	})
	return cloudSgMap, lbSgCloudMap, nil
}
