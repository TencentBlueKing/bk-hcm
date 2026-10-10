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
	"errors"
	"fmt"

	typelb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// ListExclusiveClusters 查询独占集群列表
// reference: https://cloud.tencent.com/document/api/214/49278
func (t *TCloudImpl) ListExclusiveClusters(kt *kit.Kit, opt *typelb.TCloudExclusiveClusterListOption) (
	[]typelb.TCloudExclusiveCluster, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud clb client failed, region: %s, err: %v", opt.Region, err)
	}

	req := tclb.NewDescribeExclusiveClustersRequest()
	if opt.Page != nil {
		req.Offset = cvt.ValToPtr(opt.Page.Offset)
		req.Limit = cvt.ValToPtr(opt.Page.Limit)
	}
	if len(opt.CloudIDs) != 0 {
		req.Filters = append(req.Filters, &tclb.Filter{
			Name:   cvt.ValToPtr("cluster-id"),
			Values: cvt.SliceToPtr(opt.CloudIDs),
		})
	}
	if len(opt.Network) != 0 {
		req.Filters = append(req.Filters, &tclb.Filter{
			Name:   cvt.ValToPtr("network"),
			Values: cvt.SliceToPtr([]string{string(opt.Network)}),
		})
	}
	if len(opt.ClusterTypes) != 0 {
		values := make([]string, len(opt.ClusterTypes))
		for i, one := range opt.ClusterTypes {
			values[i] = string(one)
		}
		req.Filters = append(req.Filters, &tclb.Filter{
			Name:   cvt.ValToPtr("cluster-type"),
			Values: cvt.SliceToPtr(values),
		})
	}

	resp, err := NetworkErrRetry(client.DescribeExclusiveClustersWithContext, kt, req)
	if err != nil {
		logs.Errorf("fail to describe exclusive clusters from tcloud, err: %v, region: %s, offset: %d, "+
			"limit: %d, rid: %s", err, opt.Region, cvt.PtrToVal(req.Offset), cvt.PtrToVal(req.Limit), kt.Rid)
		return nil, err
	}

	if resp == nil || resp.Response == nil {
		return nil, errors.New("empty exclusive cluster response from tcloud")
	}

	clusters := make([]typelb.TCloudExclusiveCluster, 0, len(resp.Response.ClusterSet))
	for _, one := range resp.Response.ClusterSet {
		clusters = append(clusters, typelb.TCloudExclusiveCluster{Cluster: one})
	}

	return clusters, nil
}
