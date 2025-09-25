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

// Package subnet defines subnet logics.
package subnet

import (
	"errors"
	"fmt"

	cloudclient "hcm/cmd/hc-service/logics/cloud-adaptor"
	syncaws "hcm/cmd/hc-service/logics/res-sync/aws"
	syncazure "hcm/cmd/hc-service/logics/res-sync/azure"
	syncgcp "hcm/cmd/hc-service/logics/res-sync/gcp"
	synchuawei "hcm/cmd/hc-service/logics/res-sync/huawei"
	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/pkg/adaptor/types/subnet"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcservice "hcm/pkg/api/hc-service/subnet"
	"hcm/pkg/client"
	dataclient "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// Subnet logics.
type Subnet struct {
	client  *client.ClientSet
	adaptor *cloudclient.CloudAdaptorClient
}

// NewSubnet new subnet logics.
func NewSubnet(client *client.ClientSet, adaptor *cloudclient.CloudAdaptorClient) *Subnet {
	return &Subnet{
		client:  client,
		adaptor: adaptor,
	}
}

// SubnetCreateOptions create subnet options.
type SubnetCreateOptions[T hcservice.SubnetCreateExt] struct {
	BkBizID    int64                          `validate:"required"`
	AccountID  string                         `validate:"required"`
	Region     string                         `validate:"required"`
	CloudVpcID string                         `validate:"required"`
	CreateReqs []hcservice.SubnetCreateReq[T] `validate:"min=1,max=100"`
}

// Validate SubnetCreateReq.
func (c SubnetCreateOptions[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// AzureSubnetSyncOptions sync azure subnet options.
type AzureSubnetSyncOptions struct {
	BkBizID       int64                    `validate:"required"`
	AccountID     string                   `validate:"required"`
	CloudVpcID    string                   `validate:"required"`
	ResourceGroup string                   `validate:"required"`
	Subnets       []adtysubnet.AzureSubnet `validate:"min=1,max=100"`
}

// Validate AzureSubnetSyncOptions.
func (c AzureSubnetSyncOptions) Validate() error {
	return validator.Validate.Struct(c)
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
		return nil, fmt.Errorf("some vpc can not sync to database, cloudIDs: %v", notExistCloudID)
	}

	for cloudID, id := range convVpcCloudIDMap(notExistResult) {
		existVpcMap[cloudID] = id
	}

	return existVpcMap, nil
}

func convVpcCloudIDMap(result *protocloud.VpcListResult) map[string]string {
	m := make(map[string]string, len(result.Details))
	for _, one := range result.Details {
		m[one.CloudID] = one.ID
	}
	return m
}

type vpcMeta struct {
	CloudID string
	ID      string
}

// QueryVpcIDsAndSyncForGcp 查询vpc，如果不存在则同步完再进行查询.
func QueryVpcIDsAndSyncForGcp(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client, accountID string, selfLinks []string) (map[string]vpcMeta, error) {

	if len(selfLinks) == 0 {
		return nil, errors.New("self_links is required")
	}

	sls := slice.Unique(selfLinks)
	result, err := listGcpVpcBySelfLink(kt, dataCli, sls)
	if err != nil {
		logs.Errorf("fail to list gcp vpc, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	existVpcMap := convVpcSelfLinkMap(result)

	// 如果相等，则Vpc全部同步到了db
	if len(result.Details) == len(sls) {
		return existVpcMap, nil
	}

	notExistSelfLink := make([]string, 0)
	notExistCloudIDs := make([]string, 0)
	for _, cloudID := range sls {
		if vpcData, exist := existVpcMap[cloudID]; !exist {
			notExistSelfLink = append(notExistSelfLink, cloudID)
			notExistCloudIDs = append(notExistCloudIDs, vpcData.CloudID)
		}
	}

	if len(notExistSelfLink) == 0 {
		return existVpcMap, nil
	}

	gcp, err := adaptor.Gcp(kt, accountID)
	if err != nil {
		return nil, err
	}

	syncClient := syncgcp.NewClient(dataCli, gcp)

	params := &syncgcp.SyncBaseParams{
		AccountID: accountID,
		CloudIDs:  notExistCloudIDs,
	}

	_, err = syncClient.Vpc(kt, params, &syncgcp.SyncVpcOption{})
	if err != nil {
		logs.Errorf("sync gcp vpc failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 同步完，二次查询
	notExistResult, err := listGcpVpcBySelfLink(kt, dataCli, notExistSelfLink)
	if err != nil {
		logs.Errorf("list vpc from db after sync failed, err: %v, cloudIDs: %v, rid: %s", err, notExistSelfLink, kt.Rid)
		return nil, err
	}

	if len(notExistResult.Details) != len(sls) {
		return nil, fmt.Errorf("some vpc can not sync, selfLinks: %v", notExistSelfLink)
	}

	for cloudID, id := range convVpcSelfLinkMap(notExistResult) {
		existVpcMap[cloudID] = id
	}

	return existVpcMap, nil
}

func listGcpVpcBySelfLink(kt *kit.Kit, dataCli *dataclient.Client, selfLinks []string) (
	*protocloud.VpcExtListResult[cloudcore.GcpVpcExtension], error) {

	listReq := &core.ListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				filter.AtomRule{Field: "extension.self_link", Op: filter.JSONIn.Factory(), Value: selfLinks},
			},
		},
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "cloud_id", "extension"},
	}
	result, err := dataCli.Gcp.Vpc.ListVpcExt(kt, listReq)
	if err != nil {
		logs.Errorf("list vpc from db failed, err: %v, selfLinks: %v, rid: %s", err, selfLinks, kt.Rid)
		return nil, err
	}
	return result, nil
}

func convVpcSelfLinkMap(result *protocloud.VpcExtListResult[cloudcore.GcpVpcExtension]) map[string]vpcMeta {
	m := make(map[string]vpcMeta, len(result.Details))
	for _, one := range result.Details {
		m[one.Extension.SelfLink] = vpcMeta{
			CloudID: one.CloudID,
			ID:      one.ID,
		}
	}
	return m
}

// QuerySecurityGroupIDsAndSync 查询安全组，如果不存在则同步完再进行查询.
func QuerySecurityGroupIDsAndSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client, opt *QuerySecurityGroupIDsAndSyncOption) (map[string]string, error) {
	if len(opt.CloudSecurityGroupIDs) <= 0 {
		return make(map[string]string), nil
	}

	cloudIDs := slice.Unique(opt.CloudSecurityGroupIDs)

	listReq := &protocloud.SecurityGroupListReq{
		Filter: tools.ContainersExpression("cloud_id", cloudIDs),
		Page:   core.NewDefaultBasePage(),
		Field:  []string{"id", "cloud_id"},
	}
	result, err := dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("logics list security group from db failed, err: %v, cloudIDs: %v, rid: %s",
			err, cloudIDs, kt.Rid)
		return nil, err
	}

	existMap := convSecurityGroupCloudIDMap(result)

	// 如果相等，则全部同步到了db
	if len(result.Details) == len(cloudIDs) {
		return existMap, nil
	}

	notExistCloudIDs := make([]string, 0)
	for _, cloudID := range cloudIDs {
		if _, exist := existMap[cloudID]; !exist {
			notExistCloudIDs = append(notExistCloudIDs, cloudID)
		}
	}

	// 如果有部分不存在，则触发同步
	err = batchSecurityGroupSync(kt, adaptor, dataCli, opt, notExistCloudIDs)
	if err != nil {
		return nil, err
	}

	// 同步完，二次查询
	listReq = &protocloud.SecurityGroupListReq{
		Filter: tools.ContainersExpression("cloud_id", notExistCloudIDs),
		Page:   core.NewDefaultBasePage(),
		Field:  []string{"id", "cloud_id"},
	}
	notExistResult, err := dataCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("logics list security group from db failed, err: %v, notExistCloudIDs: %v, rid: %s",
			err, notExistCloudIDs, kt.Rid)
		return nil, err
	}

	for cloudID, id := range convSecurityGroupCloudIDMap(notExistResult) {
		existMap[cloudID] = id
	}

	return existMap, nil
}

func batchSecurityGroupSync(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QuerySecurityGroupIDsAndSyncOption, cloudIDs []string) error {

	switch opt.Vendor {
	case enumor.Aws:
		return syncAwsSecurityGroup(kt, adaptor, dataCli, opt, cloudIDs)

	case enumor.TCloud:
		return syncTCloudSecurityGroup(kt, adaptor, dataCli, opt, cloudIDs)
	case enumor.HuaWei:
		return syncHuaWeiSecurityGroup(kt, adaptor, dataCli, opt, cloudIDs)

	case enumor.Azure:
		return syncAzureSecurityGroup(kt, adaptor, dataCli, opt, cloudIDs)
	default:
		return fmt.Errorf("unknown %s vendor", opt.Vendor)
	}
}

func syncAzureSecurityGroup(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QuerySecurityGroupIDsAndSyncOption, cloudIDs []string) error {

	azure, err := adaptor.Azure(kt, opt.AccountID)
	if err != nil {
		return err
	}

	syncClient := syncazure.NewClient(dataCli, azure)

	params := &syncazure.SyncBaseParams{
		AccountID:         opt.AccountID,
		ResourceGroupName: opt.ResourceGroupName,
		CloudIDs:          cloudIDs,
	}

	_, err = syncClient.SecurityGroup(kt, params, &syncazure.SyncSGOption{})
	if err != nil {
		logs.Errorf("sync azure sg failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func syncHuaWeiSecurityGroup(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QuerySecurityGroupIDsAndSyncOption, cloudIDs []string) error {

	huawei, err := adaptor.HuaWei(kt, opt.AccountID)
	if err != nil {
		return err
	}

	syncClient := synchuawei.NewClient(dataCli, huawei)

	params := &synchuawei.SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.SecurityGroup(kt, params, &synchuawei.SyncSGOption{})
	if err != nil {
		logs.Errorf("sync huawei sg failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func syncTCloudSecurityGroup(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QuerySecurityGroupIDsAndSyncOption, cloudIDs []string) error {

	tcloud, err := adaptor.TCloud(kt, opt.AccountID)
	if err != nil {
		return err
	}

	syncClient := synctcloud.NewClient(dataCli, tcloud)

	params := &synctcloud.SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.SecurityGroup(kt, params, &synctcloud.SyncSGOption{})
	if err != nil {
		logs.Errorf("sync tcloud sg failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func syncAwsSecurityGroup(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient, dataCli *dataclient.Client,
	opt *QuerySecurityGroupIDsAndSyncOption, cloudIDs []string) error {

	aws, err := adaptor.Aws(kt, opt.AccountID)
	if err != nil {
		return err
	}

	syncClient := syncaws.NewClient(dataCli, aws)

	params := &syncaws.SyncBaseParams{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.SecurityGroup(kt, params, &syncaws.SyncSGOption{})
	if err != nil {
		logs.Errorf("sync aws sg failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func convSecurityGroupCloudIDMap(result *protocloud.SecurityGroupListResult) map[string]string {
	m := make(map[string]string, len(result.Details))
	for _, one := range result.Details {
		m[one.CloudID] = one.ID
	}
	return m
}

func syncVpc(kt *kit.Kit, adaptor *cloudclient.CloudAdaptorClient,
	dataCli *dataclient.Client, opt *QueryVpcIDsAndSyncOption, notExistCloudID []string) error {

	switch opt.Vendor {
	case enumor.Aws:
		return syncAwsVpc(kt, adaptor, dataCli, opt, notExistCloudID)
	case enumor.TCloud:
		return syncTCloudVpc(kt, adaptor, dataCli, opt, notExistCloudID)
	case enumor.HuaWei:
		return syncHuaWeiVpc(kt, adaptor, dataCli, opt, notExistCloudID)
	case enumor.Azure:
		return syncAzureVpc(kt, adaptor, dataCli, opt, notExistCloudID)
	default:
		return fmt.Errorf("unknown %s vendor", opt.Vendor)
	}
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
	if err != nil {
		logs.Errorf("sync azure vpc with res failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
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
	if err != nil {
		logs.Errorf("sync huawei vpc with res failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
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
	if err != nil {
		logs.Errorf("sync tcloud vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
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
	if err != nil {
		logs.Errorf("sync aws vpc failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
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
