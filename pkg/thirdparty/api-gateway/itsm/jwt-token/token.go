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

// Package jwttoken ...
package jwttoken

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"hcm/pkg/api/core"
	cgconf "hcm/pkg/api/core/global-config"
	datagconf "hcm/pkg/api/data-service/global_config"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"

	"github.com/golang-jwt/jwt/v4"
)

func init() {
	parser = new(defaultParser)
}

var parser Parser

// Init init jwt parser.
func Init(disableITSMToken bool, dataCli *dataservice.Client) error {
	if disableITSMToken {
		logs.Warnf("disable itsm callback jwt authorize may cause security problems!!!")
		return nil
	}

	kt := core.NewBackendKit()
	sk, err := getJWTTokenSecret(kt, dataCli)
	if err != nil {
		return err
	}

	// init http request parser.
	parser = &jwtParser{
		SecretKey: sk,
	}

	return nil
}

// getJWTTokenSecret get the secret of itsm jwt token. And create if not found.
func getJWTTokenSecret(kt *kit.Kit, dataCli *dataservice.Client) ([]byte, error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", JWTSecretKeyConfigType),
			tools.RuleEqual("config_key", JWTSecretKeyConfigKey),
		),
		Page: core.NewDefaultBasePage(),
	}

	list, err := dataCli.Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get jwt secret key for itsm callback token, err: %v, filter: %+v, rid: %s",
			err, listReq.Filter, kt.Rid)
		return nil, err
	}
	if len(list.Details) == 0 {
		// create new hmac key
		hmacKey, err := generateHMACKey()
		if err != nil {
			logs.Errorf("failed to generate hmac key, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		keyBase64 := base64.StdEncoding.EncodeToString(hmacKey)
		err = createSecretKeyToGlobalConfig(kt, keyBase64, dataCli)
		if err != nil {
			logs.Errorf("failed to update jwt secret key to global config, err: %v, key: %s, rid: %s", err,
				keyBase64, kt.Rid)
			return nil, err
		}

		return hmacKey, nil
	}

	var keyBase64Str string
	err = json.UnmarshalFromString(string(list.Details[0].ConfigValue), &keyBase64Str)
	if err != nil {
		logs.Errorf("failed to unmarshal hmac key from global config, err: %v, key: %s, rid: %s", err,
			list.Details[0].ConfigValue, kt.Rid)
	}
	hmacKey, err := base64.StdEncoding.DecodeString(keyBase64Str)
	if err != nil {
		logs.Errorf("failed to decode hmac key from base64, err: %v, key: %s, rid: %s", err,
			string(list.Details[0].ConfigValue), kt.Rid)
		return nil, err
	}

	return hmacKey, nil
}

func createSecretKeyToGlobalConfig(kt *kit.Kit, key string, dataCli *dataservice.Client) error {
	dataReq := &datagconf.BatchCreateReq{
		Configs: []cgconf.GlobalConfig{
			{
				ConfigKey:   JWTSecretKeyConfigKey,
				ConfigValue: key,
				ConfigType:  JWTSecretKeyConfigType,
			},
		},
	}

	res, err := dataCli.Global.GlobalConfig.BatchCreate(kt, dataReq)
	if err != nil {
		logs.Errorf("failed to create global config, err: %v, req: %+v, rid: %s", err, *dataReq, kt.Rid)
		return err
	}

	if len(res.IDs) != 1 {
		logs.Errorf("failed to create global config, ids: %v, rid: %s", res.IDs, kt.Rid)
		return errf.New(errf.InvalidParameter, "failed to create global config")
	}

	return nil
}

func generateHMACKey() ([]byte, error) {
	key := make([]byte, 32) // HS256对应32字节密钥，HS384需48字节，HS512需64字节
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate HMAC key: %v", err)
	}
	return key, nil
}

// GenerateToken generate token.
func GenerateToken(userName, workflowID, title string) (string, error) {
	token, err := parser.GenerateToken(userName, workflowID, title)
	if err != nil {
		return "", err
	}
	return token, nil
}

// ParseToken parse token, return user_name and workflow_id, title.
func ParseToken(token string) (string, string, string, error) {
	return parser.ParseToken(token)
}

// Parser is request header parser.
type Parser interface {
	GenerateToken(userName, workflowID, title string) (string, error)
	ParseToken(token string) (string, string, string, error)
}

// defaultParser used to parse itsm callback token directly in the scenario.
type defaultParser struct{}

// GenerateToken generate jwt token.
func (p *defaultParser) GenerateToken(userName, workflowID, title string) (string, error) {
	return fmt.Sprintf("%s/%s/%s", userName, workflowID, title), nil
}

// ParseToken parse jwt token.
func (p *defaultParser) ParseToken(token string) (string, string, string, error) {
	tokens := strings.Split(token, "/")
	if len(tokens) != 3 {
		return "", "", "", errf.New(errf.InvalidParameter, "invalid jwt token")
	}

	return tokens[0], tokens[1], tokens[2], nil
}

// jwtParser used to parse requests from blueking api-gateway.
type jwtParser struct {
	// SecretKey used to parse jwt token in itsm callback request.
	SecretKey []byte
}

// GenerateToken generate jwt token.
func (p *jwtParser) GenerateToken(userName, workflowID, title string) (string, error) {
	c := claims{
		Ticket: &ticket{
			WorkflowID: workflowID,
			Title:      title,
			Verified:   true,
		},
		User: &user{
			UserName: userName,
			Verified: true,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(p.SecretKey)
}

// ParseToken parse jwt token, return user_name, workflow_id, title.
func (p *jwtParser) ParseToken(token string) (string, string, string, error) {
	tokenClaims, err := p.parseToken(token, p.SecretKey)
	if err != nil {
		return "", "", "", err
	}

	if err := tokenClaims.Validate(); err != nil {
		return "", "", "", err
	}

	return tokenClaims.User.UserName, tokenClaims.Ticket.WorkflowID, tokenClaims.Ticket.Title, nil
}

// ticket itsm ticket info.
type ticket struct {
	WorkflowID string `json:"workflow_id"`
	Title      string `json:"title"`
	Verified   bool   `json:"verified"`
}

// Validate app.
func (t *ticket) Validate() error {
	if !t.Verified {
		return errf.New(errf.InvalidParameter, "ticket not verified")
	}
	return nil
}

// user itsm user info.
type user struct {
	UserName string `json:"username"`
	Verified bool   `json:"verified"`
}

// Validate user.
func (u *user) Validate() error {
	if !u.Verified {
		return errf.New(errf.InvalidParameter, "user not verified")
	}
	return nil
}

// claims itsm callback jwt token struct.
type claims struct {
	Ticket *ticket `json:"ticket"`
	User   *user   `json:"user"`
	jwt.RegisteredClaims
}

// Validate claims.
func (c *claims) Validate() error {
	if c.Ticket == nil {
		return errf.New(errf.InvalidParameter, "ticket info is required")
	}

	if err := c.Ticket.Validate(); err != nil {
		return err
	}

	if c.User == nil {
		return errf.New(errf.InvalidParameter, "user info is required")
	}

	if err := c.User.Validate(); err != nil {
		return err
	}

	return nil
}

// parseToken parse token by jwt token and secret.
func (p *jwtParser) parseToken(token string, jwtSecret []byte) (*claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if tokenClaims == nil {
		return nil, errors.New("can not get token from parse with claims")
	}

	claims, ok := tokenClaims.Claims.(*claims)
	if !ok {
		return nil, errors.New("token claims type error")
	}

	if !tokenClaims.Valid {
		return nil, errors.New("token claims valid failed")
	}

	return claims, nil
}
