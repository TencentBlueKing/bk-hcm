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
	"fmt"

	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
)

// GetMonitorData get monitor data from tcloud.
// reference: https://cloud.tencent.com/document/product/248/31014
func (t *TCloudImpl) GetMonitorData(kt *kit.Kit, opt *typecvm.TCloudMonitorDataOption) (
	*typecvm.TCloudMonitorDataResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "monitor data option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.MonitorClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("new tcloud monitor client failed, err: %v", err)
	}

	req := monitor.NewGetMonitorDataRequest()
	req.Namespace = common.StringPtr(constant.TCloudCvmNamespace)
	req.MetricName = common.StringPtr(opt.MetricName)
	req.Period = common.Uint64Ptr(uint64(opt.Period))
	req.StartTime = common.StringPtr(opt.StartTime)
	req.EndTime = common.StringPtr(opt.EndTime)

	// 构建Instances参数
	instances := make([]*monitor.Instance, 0, len(opt.InstanceIDs))
	for _, instID := range opt.InstanceIDs {
		instances = append(instances, &monitor.Instance{
			Dimensions: []*monitor.Dimension{
				{
					Name:  common.StringPtr(constant.TCloudCvmInstanceIDKey),
					Value: common.StringPtr(instID),
				},
			},
		})
	}
	req.Instances = instances

	resp, err := client.GetMonitorDataWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("get tcloud monitor data failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 转换响应数据
	result := &typecvm.TCloudMonitorDataResult{
		DataPoints: make([]*typecvm.MonitorDataPoint, 0, len(resp.Response.DataPoints)),
	}

	for _, dp := range resp.Response.DataPoints {
		dataPoint := &typecvm.MonitorDataPoint{
			Dimensions: make([]*typecvm.MonitorDimension, 0, len(dp.Dimensions)),
			Timestamps: make([]int64, 0, len(dp.Timestamps)),
			Values:     make([]float64, 0, len(dp.Values)),
		}

		// 转换Dimensions
		for _, dim := range dp.Dimensions {
			dataPoint.Dimensions = append(dataPoint.Dimensions, &typecvm.MonitorDimension{
				Name:  cvt.PtrToVal(dim.Name),
				Value: cvt.PtrToVal(dim.Value),
			})
		}

		// 转换Timestamps
		for _, ts := range dp.Timestamps {
			dataPoint.Timestamps = append(dataPoint.Timestamps, int64(cvt.PtrToVal(ts)))
		}

		// 转换Values
		for _, val := range dp.Values {
			dataPoint.Values = append(dataPoint.Values, cvt.PtrToVal(val))
		}

		result.DataPoints = append(result.DataPoints, dataPoint)
	}

	return result, nil
}
