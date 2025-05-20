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

package ressync

import (
	cloudclient "hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/logics/res-sync/aws"
	"hcm/cmd/hc-service/logics/res-sync/azure"
	"hcm/cmd/hc-service/logics/res-sync/gcp"
	"hcm/cmd/hc-service/logics/res-sync/huawei"
	"hcm/cmd/hc-service/logics/res-sync/other"
	"hcm/cmd/hc-service/logics/res-sync/tcloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
)

// Interface sync support vendor.
type Interface interface {
	TCloud(kt *kit.Kit, accountID string) (tcloud.Interface, error)
	Aws(kt *kit.Kit, accountID string) (aws.Interface, error)
	HuaWei(kt *kit.Kit, accountID string) (huawei.Interface, error)
	Gcp(kt *kit.Kit, accountID string) (gcp.Interface, error)
	Azure(kt *kit.Kit, accountID string) (azure.Interface, error)
	Other(kt *kit.Kit, accountID string) (other.Interface, error)
}

var _ Interface = new(client)

// NewClient new client.
func NewClient(ad *cloudclient.CloudAdaptorClient, dataCli *dataservice.Client) Interface {
	return &client{
		ad:      ad,
		dataCli: dataCli,
	}
}

// client sync client.
type client struct {
	ad      *cloudclient.CloudAdaptorClient
	dataCli *dataservice.Client
}

// TCloud ...
func (cli *client) TCloud(kt *kit.Kit, accountID string) (tcloud.Interface, error) {
	cloudCli, err := cli.ad.TCloud(kt, accountID)
	if err != nil {
		return nil, err
	}

	return tcloud.NewClient(cli.dataCli, cloudCli), nil
}

// Aws ...
func (cli *client) Aws(kt *kit.Kit, accountID string) (aws.Interface, error) {
	cloudCli, err := cli.ad.Aws(kt, accountID)
	if err != nil {
		return nil, err
	}

	return aws.NewClient(cli.dataCli, cloudCli), nil
}

// Gcp ...
func (cli *client) Gcp(kt *kit.Kit, accountID string) (gcp.Interface, error) {
	cloudCli, err := cli.ad.Gcp(kt, accountID)
	if err != nil {
		return nil, err
	}

	return gcp.NewClient(cli.dataCli, cloudCli), nil
}

// HuaWei ...
func (cli *client) HuaWei(kt *kit.Kit, accountID string) (huawei.Interface, error) {
	cloudCli, err := cli.ad.HuaWei(kt, accountID)
	if err != nil {
		return nil, err
	}

	return huawei.NewClient(cli.dataCli, cloudCli), nil
}

// Azure ...
func (cli *client) Azure(kt *kit.Kit, accountID string) (azure.Interface, error) {
	cloudCli, err := cli.ad.Azure(kt, accountID)
	if err != nil {
		return nil, err
	}

	return azure.NewClient(cli.dataCli, cloudCli), nil
}

// Other ...
func (cli *client) Other(kt *kit.Kit, accountID string) (other.Interface, error) {
	return other.NewClient(cli.dataCli, accountID), nil
}
