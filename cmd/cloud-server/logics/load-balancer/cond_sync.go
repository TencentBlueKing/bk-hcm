/*
 *
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

package lblogic

import (
	"time"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// CondSyncLoadBalancerOption 负载均衡条件同步的厂商无关参数
type CondSyncLoadBalancerOption struct {
	// Vendor 云厂商，当前支持 tcloud
	Vendor enumor.Vendor
	// AccountID 账号ID
	AccountID string
	// BkBizID 业务ID，未分配业务时为 constant.UnassignedBiz
	BkBizID int64
	// Regions 本次同步的地域列表
	Regions []string
	// CloudIDs 本次同步指定的云上ID，为空表示地域全量
	CloudIDs []string
	// TagFilters 标签过滤条件
	TagFilters core.MultiValueTagMap
}

// LoadBalancerRegionDiff 单个地域云上与DB的差异，create/update 来自云上，delete 来自DB
type LoadBalancerRegionDiff struct {
	Region string
	Create []corelb.LoadBalancerBrief
	Update []corelb.LoadBalancerBrief
	Delete []corelb.LoadBalancerBrief
}

// ListAndDiffLoadBalancerByRegion 查询单个地域云上与DB的负载均衡并计算差异
func ListAndDiffLoadBalancerByRegion(kt *kit.Kit, cliSet *client.ClientSet, opt *CondSyncLoadBalancerOption,
	region string) (*LoadBalancerRegionDiff, error) {

	listOpt := &ListLoadBalancerBriefOption{
		Vendor:     opt.Vendor,
		AccountID:  opt.AccountID,
		Region:     region,
		CloudIDs:   opt.CloudIDs,
		TagFilters: opt.TagFilters,
	}
	// 腾讯云上没有业务概念，业务入口下先按 bk_biz_id 查DB，再用查出的云上ID收敛云上查询范围。
	isBizEntry := opt.BkBizID != 0 && opt.BkBizID != constant.UnassignedBiz
	if isBizEntry {
		listOpt.BkBizID = cvt.ValToPtr(opt.BkBizID)
	}
	// DB查询先于云上查询，一是收敛云上范围需要DB结果，二是同步期间入库的实例只会被判为新增而非删除。
	dbStartedAt := time.Now()
	dbLBs, err := ListLoadBalancerBriefFromDB(kt, cliSet.DataService(), listOpt)
	if err != nil {
		logs.Errorf("list db load balancer brief failed, err: %v, account: %s, region: %s, rid: %s",
			err, opt.AccountID, region, kt.Rid)
		return nil, err
	}
	logs.Infof("list db load balancer brief done, account: %s, region: %s, count: %d, cost: %s, rid: %s",
		opt.AccountID, region, len(dbLBs), time.Since(dbStartedAt), kt.Rid)

	if isBizEntry && opt.Vendor == enumor.TCloud {
		// 业务下没有负载均衡时无需查询云上，云上多出来的实例不属于本业务，其新建由资源侧入口负责。
		if len(dbLBs) == 0 {
			return &LoadBalancerRegionDiff{Region: region}, nil
		}
		listOpt.CloudIDs = slice.Map(dbLBs, func(one corelb.LoadBalancerBrief) string { return one.CloudID })
	}

	cloudStartedAt := time.Now()
	cloudLBs, err := ListLoadBalancerBriefFromCloud(kt, cliSet, listOpt)
	if err != nil {
		logs.Errorf("list cloud load balancer brief failed, err: %v, account: %s, region: %s, rid: %s",
			err, opt.AccountID, region, kt.Rid)
		return nil, err
	}
	logs.Infof("list cloud load balancer brief done, account: %s, region: %s, count: %d, cost: %s, rid: %s",
		opt.AccountID, region, len(cloudLBs), time.Since(cloudStartedAt), kt.Rid)

	diff := DiffLoadBalancerBrief(region, cloudLBs, dbLBs)
	logs.Infof("diff conditional sync load balancer done, account: %s, region: %s, create: %d, update: %d, "+
		"delete: %d, rid: %s", opt.AccountID, region, len(diff.Create), len(diff.Update), len(diff.Delete), kt.Rid)

	return diff, nil
}

// DiffLoadBalancerBrief 按云上ID对比云上与DB的负载均衡，得到单个地域的差异。
func DiffLoadBalancerBrief(region string, cloudLBs, dbLBs []corelb.LoadBalancerBrief) *LoadBalancerRegionDiff {
	dbLBMap := cvt.SliceToMap(dbLBs, func(one corelb.LoadBalancerBrief) (string, corelb.LoadBalancerBrief) {
		return one.CloudID, one
	})

	diff := &LoadBalancerRegionDiff{Region: region}
	cloudIDs := make(map[string]struct{}, len(cloudLBs))
	for _, one := range cloudLBs {
		if _, exists := cloudIDs[one.CloudID]; exists {
			continue
		}
		cloudIDs[one.CloudID] = struct{}{}

		if _, exists := dbLBMap[one.CloudID]; exists {
			diff.Update = append(diff.Update, one)
			continue
		}
		diff.Create = append(diff.Create, one)
	}

	for _, one := range dbLBs {
		if _, exists := cloudIDs[one.CloudID]; !exists {
			diff.Delete = append(diff.Delete, one)
		}
	}

	return diff
}
