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

package huawei

import (
	"fmt"

	"hcm/pkg/adaptor/types"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	bssintl "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2"
	bssintlv2region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/region"
	dcs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dcs/v2"
	dcsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dcs/v2/region"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
	eip "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2"
	eipregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/region"
	eipv3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v3"
	eipv3region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v3/region"
	evs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2"
	evsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/region"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	iamregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	ims "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ims/v2"
	vpcv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
	vpc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3"
	vpcregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/region"
)

const (
	ErrDataNotFound = "readObjectStart: expect { or n, but found"
	CvmDeleteStatus = "DELETED"
)

type clientSet struct {
	credentials       *basic.Credentials
	globalCredentials *global.Credentials
}

func newClientSet(secret *types.BaseSecret) *clientSet {
	return &clientSet{
		credentials: basic.NewCredentialsBuilder().
			WithAk(secret.CloudSecretID).
			WithSk(secret.CloudSecretKey).
			Build(),
		globalCredentials: global.NewCredentialsBuilder().
			WithAk(secret.CloudSecretID).
			WithSk(secret.CloudSecretKey).
			Build(),
	}
}

func (c *clientSet) iamClient(region *region.Region) (client *iam.IamClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	client = iam.NewIamClient(
		iam.IamClientBuilder().
			WithRegion(region).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return client, nil
}

func (c *clientSet) iamRegionClient(region string) (client *iam.IamClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()
	client = iam.NewIamClient(
		iam.IamClientBuilder().
			WithRegion(iamregion.ValueOf(region)).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return client, nil
}

func (c *clientSet) evsClient(region string) (client *evs.EvsClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()
	client = evs.NewEvsClient(
		evs.EvsClientBuilder().
			WithRegion(evsregion.ValueOf(region)).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return client, nil
}

func (c *clientSet) vpcClient(regionID string) (cli *vpc.VpcClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	client := vpc.NewVpcClient(
		vpc.VpcClientBuilder().
			WithRegion(vpcregion.ValueOf(regionID)).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return client, nil
}

func (c *clientSet) vpcClientV2(regionID string) (cli *vpcv2.VpcClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	client := vpcv2.NewVpcClient(
		vpcv2.VpcClientBuilder().
			WithRegion(vpcregion.ValueOf(regionID)).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return client, nil
}

func (c *clientSet) imsClientV2(region *region.Region) (cli *ims.ImsClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	cli = ims.NewImsClient(
		ims.ImsClientBuilder().
			WithRegion(region).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return cli, nil
}

func (c *clientSet) ecsClient(regionID string) (cli *ecs.EcsClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	client := ecs.NewEcsClient(
		ecs.EcsClientBuilder().
			WithRegion(ecsregion.ValueOf(regionID)).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return client, nil
}

func (c *clientSet) dcsClient(regionID string) (cli *dcs.DcsClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	client := dcs.NewDcsClient(
		dcs.DcsClientBuilder().
			WithRegion(dcsregion.ValueOf(regionID)).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return client, nil
}

func (c *clientSet) eipClient(regionID string) (cli *eip.EipClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	cli = eip.NewEipClient(
		eip.EipClientBuilder().
			WithRegion(eipregion.ValueOf(regionID)).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return cli, nil
}

func (c *clientSet) eipV3Client(regionID string) (cli *eipv3.EipClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	cli = eipv3.NewEipClient(
		eipv3.EipClientBuilder().
			WithRegion(eipv3region.ValueOf(regionID)).
			WithCredential(c.credentials).
			WithHttpConfig(config.DefaultHttpConfig()).
			Build())

	return cli, nil
}

func (c *clientSet) bssintlClient() (cli *bssintl.BssintlClient, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("huawei error recovered, err: %v", p)
		}
	}()

	client := bssintl.NewBssintlClient(
		bssintl.BssintlClientBuilder().
			WithRegion(bssintlv2region.ValueOf(bssintlv2region.AP_SOUTHEAST_1.Id)).
			WithCredential(c.globalCredentials).
			Build())

	return client, nil
}
