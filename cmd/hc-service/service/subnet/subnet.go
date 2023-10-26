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

// Package subnet defines subnet service.
package subnet

import (
	"errors"
	"fmt"

	syncaws "hcm/cmd/hc-service/logics/res-sync/aws"
	syncazure "hcm/cmd/hc-service/logics/res-sync/azure"
	synchuawei "hcm/cmd/hc-service/logics/res-sync/huawei"
	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	subnetlogics "hcm/cmd/hc-service/logics/subnet"
	"hcm/cmd/hc-service/service/capability"
	cloudadaptor "hcm/cmd/hc-service/service/cloud-adaptor"
	cloudclient "hcm/cmd/hc-service/service/cloud-adaptor"
	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// InitSubnetService initial the subnet service
func InitSubnetService(cap *capability.Capability) {
	s := &subnet{
		ad:     cap.CloudAdaptor,
		cs:     cap.ClientSet,
		subnet: subnetlogics.NewSubnet(cap.ClientSet, cap.CloudAdaptor),
	}

	h := rest.NewHandler()

	h.Add("TCloudSubnetBatchCreate", "POST", "/vendors/tcloud/subnets/batch/create", s.TCloudSubnetBatchCreate)
	h.Add("AwsSubnetCreate", "POST", "/vendors/aws/subnets/create", s.AwsSubnetCreate)
	h.Add("HuaWeiSubnetCreate", "POST", "/vendors/huawei/subnets/create", s.HuaWeiSubnetCreate)
	h.Add("GcpSubnetCreate", "POST", "/vendors/gcp/subnets/create", s.GcpSubnetCreate)
	h.Add("AzureSubnetCreate", "POST", "/vendors/azure/subnets/create", s.AzureSubnetCreate)

	h.Add("TCloudSubnetUpdate", "PATCH", "/vendors/tcloud/subnets/{id}", s.TCloudSubnetUpdate)
	h.Add("AwsSubnetUpdate", "PATCH", "/vendors/aws/subnets/{id}", s.AwsSubnetUpdate)
	h.Add("HuaWeiSubnetUpdate", "PATCH", "/vendors/huawei/subnets/{id}", s.HuaWeiSubnetUpdate)
	h.Add("GcpSubnetUpdate", "PATCH", "/vendors/gcp/subnets/{id}", s.GcpSubnetUpdate)
	h.Add("AzureSubnetUpdate", "PATCH", "/vendors/azure/subnets/{id}", s.AzureSubnetUpdate)

	h.Add("TCloudSubnetDelete", "DELETE", "/vendors/tcloud/subnets/{id}", s.TCloudSubnetDelete)
	h.Add("AwsSubnetDelete", "DELETE", "/vendors/aws/subnets/{id}", s.AwsSubnetDelete)
	h.Add("HuaWeiSubnetDelete", "DELETE", "/vendors/huawei/subnets/{id}", s.HuaWeiSubnetDelete)
	h.Add("GcpSubnetDelete", "DELETE", "/vendors/gcp/subnets/{id}", s.GcpSubnetDelete)
	h.Add("AzureSubnetDelete", "DELETE", "/vendors/azure/subnets/{id}", s.AzureSubnetDelete)

	// count subnet available ips
	h.Add("TCloudListSubnetCountIP", "POST", "/vendors/tcloud/subnets/ips/count/list", s.TCloudListSubnetCountIP)
	h.Add("AwsListSubnetCountIP", "POST", "/vendors/aws/subnets/ips/count/list", s.AwsListSubnetCountIP)
	h.Add("AzureListSubnetCountIP", "POST", "/vendors/azure/subnets/ips/count/list", s.AzureListSubnetCountIP)
	h.Add("HuaWeiSubnetCountIP", "POST", "/vendors/huawei/subnets/{id}/ips/count", s.HuaWeiSubnetCountIP)
	h.Add("GcpSubnetCountIP", "POST", "/vendors/gcp/subnets/ips/count/list", s.GcpSubnetCountIP)

	h.Load(cap.WebService)
}

type subnet struct {
	ad     *cloudadaptor.CloudAdaptorClient
	cs     *client.ClientSet
	subnet *subnetlogics.Subnet
}

// QueryVpcIDsAndSyncOption ...
type QueryVpcIDsAndSyncOption struct {
	Vendor            enumor.Vendor `json:"vendor" validate:"required"`
	AccountID         string        `json:"account_id" validate:"required"`
	CloudVpcIDs       []string      `json:"cloud_vpc_ids" validate:"required"`
	ResourceGroupName string        `json:"resource_group_name" validate:"omitempty"`
	Region            string        `json:"region" validate:"omitempty"`
}

// Validate QueryVpcIDsAndSyncOption
func (opt *QueryVpcIDsAndSyncOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudVpcIDs) == 0 {
		return errors.New("CloudVpcIDs is required")
	}

	if len(opt.CloudVpcIDs) > int(core.DefaultMaxPageLimit) {
		return fmt.Errorf("cloudIDs should <= %d", core.DefaultMaxPageLimit)
	}

	return nil
}

// QueryVpcIDsAndSync 查询vpc，如果不存在则同步完再进行查询.
func QueryVpcIDsAndSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client, opt *QueryVpcIDsAndSyncOption) (map[string]string, error) {

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	cloudVpcIDs := slice.Unique(opt.CloudVpcIDs)
	result, err := getVpcsFromDB(kt, dataCli, cloudVpcIDs)
	if err != nil {
		return nil, err
	}

	existVpcMap := convVpcCloudIDMap(result)

	// 如果相等，则Vpc全部同步到了db
	if len(result.Details) == len(cloudVpcIDs) {
		return existVpcMap, nil
	}

	notExistCloudID := make([]string, 0)
	for _, cloudID := range cloudVpcIDs {
		if _, exist := existVpcMap[cloudID]; !exist {
			notExistCloudID = append(notExistCloudID, cloudID)
		}
	}

	// 如果有部分vpc不存在，则触发vpc同步
	err = syncVpc(kt, adaptor, dataCli, opt, notExistCloudID)
	if err != nil {
		return nil, err
	}

	// 同步完，二次查询
	notExistResult, err := getVpcsFromDB(kt, dataCli, notExistCloudID)
	if err != nil {
		return nil, err
	}

	if len(notExistResult.Details) != len(cloudVpcIDs) {
		return nil, fmt.Errorf("some vpc can not sync, cloudIDs: %v", notExistCloudID)
	}

	for cloudID, id := range convVpcCloudIDMap(notExistResult) {
		existVpcMap[cloudID] = id
	}

	return existVpcMap, nil
}

func syncVpc(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client, opt *QueryVpcIDsAndSyncOption, notExistCloudID []string) error {

	var err error
	switch opt.Vendor {
	case enumor.Aws:
		err = syncAwsVpc(kt, adaptor, dataCli, opt, notExistCloudID)

	case enumor.TCloud:
		err = syncTCloudVpc(kt, adaptor, dataCli, opt, notExistCloudID)

	case enumor.HuaWei:
		err = syncHuaWeiVpc(kt, adaptor, dataCli, opt, notExistCloudID)

	case enumor.Azure:
		err = syncAzureVpc(kt, adaptor, dataCli, opt, notExistCloudID)

	default:
		return fmt.Errorf("unknown %s vendor", opt.Vendor)
	}
	if err != nil {
		logs.Errorf("sync %s vpc failed, err: %v, rid: %s", opt.Vendor, err, kt.Rid)
		return err
	}

	return nil
}

func syncAzureVpc(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QueryVpcIDsAndSyncOption, notExistCloudID []string) error {

	azure, err := adaptor.Azure(kt, opt.AccountID)
	if err != nil {
		return err
	}

	syncClient := syncazure.NewClient(dataCli, azure)

	params := &syncazure.SyncBaseParams{
		AccountID:         opt.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          notExistCloudID,
	}

	_, err = syncClient.Vpc(kt, params, &syncazure.SyncVpcOption{})
	return err
}

func syncHuaWeiVpc(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QueryVpcIDsAndSyncOption, notExistCloudID []string) error {

	huawei, err := adaptor.HuaWei(kt, opt.AccountID)
	if err != nil {
		return err
	}

	syncClient := synchuawei.NewClient(dataCli, huawei)

	params := &synchuawei.SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  notExistCloudID,
	}

	_, err = syncClient.Vpc(kt, params, &synchuawei.SyncVpcOption{})
	return err
}

func syncTCloudVpc(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QueryVpcIDsAndSyncOption, notExistCloudID []string) error {

	tcloud, err := adaptor.TCloud(kt, opt.AccountID)
	if err != nil {
		return err
	}

	syncClient := synctcloud.NewClient(dataCli, tcloud)

	params := &synctcloud.SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  notExistCloudID,
	}

	_, err = syncClient.Vpc(kt, params, &synctcloud.SyncVpcOption{})
	return err
}

func syncAwsVpc(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QueryVpcIDsAndSyncOption, notExistCloudID []string) error {

	aws, err := adaptor.Aws(kt, opt.AccountID)
	if err != nil {
		return err
	}

	syncClient := syncaws.NewClient(dataCli, aws)

	params := &syncaws.SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  notExistCloudID,
	}

	_, err = syncClient.Vpc(kt, params, &syncaws.SyncVpcOption{})
	return err
}

func getVpcsFromDB(kt *kit.Kit, dataCli *dataclient.Client,
	cloudVpcIDs []string) (*protocloud.VpcListResult, error) {

	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("cloud_id", cloudVpcIDs),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "cloud_id"},
	}
	result, err := dataCli.Global.Vpc.List(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list vpc from db failed, err: %v, cloudIDs: %v, rid: %s", err, cloudVpcIDs, kt.Rid)
		return nil, err
	}

	return result, nil
}

func convVpcCloudIDMap(result *protocloud.VpcListResult) map[string]string {
	m := make(map[string]string, len(result.Details))
	for _, one := range result.Details {
		m[one.CloudID] = one.ID
	}
	return m
}
