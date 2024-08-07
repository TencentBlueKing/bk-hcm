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

package gcp

import (
	"fmt"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/kit"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/bigquery"
	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"google.golang.org/api/cloudbilling/v1"
	res "google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/compute/v1"
	iam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type clientSet struct {
	credential *types.GcpCredential
}

func newClientSet(credential *types.GcpCredential) *clientSet {
	return &clientSet{credential}
}

func (c *clientSet) assetClient(kt *kit.Kit) (*asset.Client, error) {
	opt := option.WithCredentialsJSON(c.credential.Json)
	client, err := asset.NewClient(kt.Ctx, opt)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *clientSet) iamClient(kt *kit.Kit) (*credentials.IamCredentialsClient, error) {
	opt := option.WithCredentialsJSON(c.credential.Json)
	client, err := credentials.NewIamCredentialsClient(kt.Ctx, opt)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *clientSet) computeClient(kt *kit.Kit) (*compute.Service, error) {
	opt := option.WithCredentialsJSON(c.credential.Json)
	service, err := compute.NewService(kt.Ctx, opt)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (c *clientSet) bigQueryClient(kt *kit.Kit) (*bigquery.Client, error) {
	opt := option.WithCredentialsJSON(c.credential.Json)
	service, err := bigquery.NewClient(kt.Ctx, c.credential.CloudProjectID, opt)
	if err != nil {
		return nil, fmt.Errorf("gcp.bigquery.NewClient, projectID: %s, err: %+v",
			c.credential.CloudProjectID, err)
	}
	defer service.Close()

	return service, nil
}

func (c *clientSet) resClient(kt *kit.Kit) (*res.Service, error) {
	opt := option.WithCredentialsJSON(c.credential.Json)
	service, err := res.NewService(kt.Ctx, opt)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (c *clientSet) iamServiceClient(kt *kit.Kit) (*iam.Service, error) {
	opt := option.WithCredentialsJSON(c.credential.Json)
	service, err := iam.NewService(kt.Ctx, opt)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (c *clientSet) billingClient(kt *kit.Kit) (*cloudbilling.APIService, error) {
	opt := option.WithCredentialsJSON(c.credential.Json)
	service, err := cloudbilling.NewService(kt.Ctx, opt)
	if err != nil {
		return nil, err
	}
	return service, nil
}
