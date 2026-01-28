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

package ziyan

import (
	"fmt"
	"net/http"

	"hcm/pkg/adaptor/metric"
	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/adaptor/types"
	typescos "hcm/pkg/adaptor/types/cos"
	"hcm/pkg/criteria/constant"
	bpaas "hcm/pkg/thirdparty/tencentcloud/bpaas/v20181217"
	"hcm/pkg/tools/converter"

	billing "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/billing/v20180709"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	regionsdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/region/v20220627"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	tag "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tag/v20180813"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type clientSet struct {
	credential common.CredentialIface
	profile    *profile.ClientProfile
}

// 自研云客户端集
func newClientSet(s *types.BaseSecret, profile *profile.ClientProfile) *clientSet {
	return &clientSet{
		// 自研云使用随机秘钥
		credential: types.WrapZiyanMultiSecret(s),
		profile:    profile,
	}

}

// SetRateLimitRetryWithConstInterval ...
func (c *clientSet) SetRateLimitRetryWithConstInterval() {
	c.profile.RateLimitExceededMaxRetries = constant.MaxRetries
	c.profile.RateLimitExceededRetryDuration = tcloud.RandomDurationFunc()
	return
}

func (c *clientSet) getProfileUseSpecifiedEndpoint(endpoint string) *profile.ClientProfile {
	newHttpProfile := converter.PtrToVal(c.profile.HttpProfile)
	newHttpProfile.Endpoint = endpoint

	newProfile := converter.PtrToVal(c.profile)
	newProfile.HttpProfile = converter.ValToPtr(newHttpProfile)

	return converter.ValToPtr(newProfile)
}

// CamServiceClient tcloud ziyan sdk cam client
func (c *clientSet) CamServiceClient(region string) (*cam.Client, error) {
	// 使用内部域名
	client, err := cam.NewClient(c.credential, region, c.getProfileUseSpecifiedEndpoint(constant.InternalCamEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// CvmClient tcloud ziyan sdk cvm client
func (c *clientSet) CvmClient(region string) (*cvm.Client, error) {
	// 使用内部域名
	client, err := cvm.NewClient(c.credential, region, c.getProfileUseSpecifiedEndpoint(constant.InternalCvmEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// CbsClient tcloud ziyan sdk cbs client
func (c *clientSet) CbsClient(region string) (*cbs.Client, error) {
	// 使用内部域名
	client, err := cbs.NewClient(c.credential, region, c.getProfileUseSpecifiedEndpoint(constant.InternalCbsEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// VpcClient tcloud ziyan sdk vpc client
func (c *clientSet) VpcClient(region string) (*vpc.Client, error) {
	// 使用内部域名
	client, err := vpc.NewClient(c.credential, region, c.getProfileUseSpecifiedEndpoint(constant.InternalVpcEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// BillClient tcloud ziyan sdk bill client
func (c *clientSet) BillClient() (*billing.Client, error) {
	// 使用内部域名
	client, err := billing.NewClient(c.credential, "",
		c.getProfileUseSpecifiedEndpoint(constant.InternalBillingEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// ClbClient tcloud clb client
func (c *clientSet) ClbClient(region string) (*clb.Client, error) {
	// 使用内部域名
	client, err := clb.NewClient(c.credential, region, c.getProfileUseSpecifiedEndpoint(constant.InternalClbEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// TagClient tcloud tag client
func (c *clientSet) TagClient() (*tag.Client, error) {
	// 使用内部域名
	client, err := tag.NewClient(c.credential, "", c.getProfileUseSpecifiedEndpoint(constant.InternalTagEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// CertClient tcloud cert client
func (c *clientSet) CertClient() (*ssl.Client, error) {
	// 使用内部域名
	client, err := ssl.NewClient(c.credential, "", c.getProfileUseSpecifiedEndpoint(constant.InternalCertEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// BPaasClient tcloud ziyan sdk bpaas client
func (c *clientSet) BPaasClient() (*bpaas.Client, error) {
	// 使用内部域名
	client, err := bpaas.NewClient(c.credential, "", c.getProfileUseSpecifiedEndpoint(constant.InternalBPaasEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

// RegionClient tcloud ziyan sdk region client
func (c *clientSet) RegionClient() (*regionsdk.Client, error) {
	// 使用内部域名
	client, err := regionsdk.NewClient(c.credential, "",
		c.getProfileUseSpecifiedEndpoint(constant.InternalRegionEndpoint))
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetZiyanRecordRoundTripper(nil))

	return client, nil
}

var cosUrlMap = map[typescos.UrlType]string{
	typescos.NormalUrl:            constant.InternalCosEndpoint,
	typescos.UrlWithNameAndRegion: constant.InternalCosEndpointWithNameAndRegion,
}

// CosClient tcloud ziyan cos client
func (c *clientSet) CosClient(opt *typescos.ClientOpt) (*cos.Client, error) {
	if opt == nil {
		return nil, fmt.Errorf("cos client opt is nil")
	}

	cosUrl, err := opt.GetUrl(cosUrlMap)
	if err != nil {
		return nil, err
	}
	client := cos.NewClient(
		cosUrl,
		&http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  c.credential.GetSecretId(),
				SecretKey: c.credential.GetSecretKey(),
			},
		},
	)

	return client, nil
}

// MonitorClient tcloud monitor client
func (c *clientSet) MonitorClient(region string) (*monitor.Client, error) {
	// TODO 仅占位，暂未验证，应该需要内部域名
	client, err := monitor.NewClient(c.credential, region, c.profile)
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(metric.GetTCloudRecordRoundTripper(nil))

	return client, nil
}
