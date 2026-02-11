/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tjjapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
)

// TjjClientInterface tjj api interface
type TjjClientInterface interface {
	// GetPwd get device password
	GetPwd(ctx context.Context, header http.Header, ip string) (string, error)
}

// NewTjjClientInterface creates a tjj api instance
func NewTjjClientInterface(opts cc.TjjCli, reg prometheus.Registerer) (TjjClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "tjj api",
			servers: []string{opts.TjjApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	tjj := &tjjApi{
		client: rest.NewClient(c, "/"),
		opts:   &opts,
	}

	return tjj, nil
}

// tjjApi tjj api interface implementation
type tjjApi struct {
	client rest.ClientInterface
	opts   *cc.TjjCli
}

func (t *tjjApi) sign(rtx string) (tokenString string, err error) {
	// The token content.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  t.opts.SecretID,
		"user": rtx,
		"iat":  time.Now().Add(-1 * time.Hour).Unix(),
	})
	// Sign the token with the specified secret.
	tokenString, err = token.SignedString([]byte(t.opts.SecretKey))
	return
}

// GetPwd get device password
func (t *tjjApi) GetPwd(ctx context.Context, header http.Header, ip string) (string, error) {
	subPath := "/thirdpartyapi/tjj/devicepassword"

	token, signErr := t.sign(t.opts.Operator)
	if signErr != nil {
		return "", signErr
	}
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+token)

	req := &GetPwdReq{
		IpList:    []string{ip},
		Decrypted: true,
	}

	resp := new(GetPwdResp)
	err := t.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return "", err
	}

	if resp.Code != 0 {
		return "", fmt.Errorf("failed to get tjj pwd, code: %d, msg: %s", resp.Code, resp.Message)
	}

	pwd, ok := resp.Data[ip]
	if !ok {
		return "", fmt.Errorf("get no tjj pwd by ip %s", ip)
	}

	return pwd, err
}
