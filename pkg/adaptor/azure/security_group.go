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

package azure

import (
	"fmt"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
)

// CreateSecurityGroup create security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) CreateSecurityGroup(kt *kit.Kit, opt *types.AzureSecurityGroupOption) (*armnetwork.SecurityGroup,
		error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	sg := armnetwork.SecurityGroup{
		Location: &opt.Region,
	}
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, opt.Name, sg, nil)
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, err := poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate result failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &resp.SecurityGroup, nil
}

// UpdateSecurityGroup update security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/create-or-update
func (az *Azure) UpdateSecurityGroup(kt *kit.Kit, opt *types.AzureSecurityGroupOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group update option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return fmt.Errorf("new security group client failed, err: %v", err)
	}

	sg := armnetwork.SecurityGroup{
		Location: &opt.Region,
	}
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, opt.Name, sg, nil)
	if err != nil {
		logs.Errorf("request to BeginCreateOrUpdate failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginCreateOrUpdate result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteSecurityGroup delete security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/delete?tabs=HTTP
func (az *Azure) DeleteSecurityGroup(kt *kit.Kit, opt *types.AzureSecurityGroupOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return fmt.Errorf("new security group client failed, err: %v", err)
	}

	poller, err := client.BeginDelete(kt.Ctx, opt.ResourceGroupName, opt.Name, nil)
	if err != nil {
		logs.Errorf("request to BeginDelete failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("pull the BeginDelete result failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSecurityGroup list security group.
// reference: https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups/list-all
func (az *Azure) ListSecurityGroup(kt *kit.Kit, opt *types.AzureSecurityGroupListOption) (
		*runtime.Pager[armnetwork.SecurityGroupsClientListResponse], error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	return client.NewListPager(opt.ResourceGroupName, nil), nil
}

func (az *Azure) getSecurityGroupByCloudID(kt *kit.Kit, resGroupName, cloudID string) (*armnetwork.SecurityGroup,
		error) {

	client, err := az.clientSet.securityGroupClient()
	if err != nil {
		return nil, fmt.Errorf("new security group client failed, err: %v", err)
	}

	pager := client.NewListPager(resGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, one := range nextResult.Value {
			if *one.ID == cloudID {
				return one, nil
			}
		}
	}

	return nil, errf.Newf(errf.RecordNotFound, "security group: %s not found", cloudID)
}
