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

package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	cos "github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

// TCloudCOS tcloud cos client
type TCloudCOS struct {
	prefix string
	config cc.ObjectStoreTCloud
	cli    *cos.Client
	sts    *sts.Client
}

// URLToken 通过 tag 的方式，用户可以将请求参数或者请求头部放进签名中。
type URLToken struct {
	SessionToken string `url:"x-cos-security-token,omitempty" header:"-"`
}

// NewTCloudCOS create cos client
func NewTCloudCOS(config cc.ObjectStoreTCloud) (*TCloudCOS, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	prefix := config.COSPrefix
	ak := config.COSSecretID
	sk := config.COSSecretKey
	bucketURL := config.COSBucketURL
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("parse bucket url %s failed, err %s", bucketURL, err.Error())
	}
	b := &cos.BaseURL{BucketURL: u}
	transport := &cos.AuthorizationTransport{
		SecretID:  ak,
		SecretKey: sk,
	}
	if config.COSIsDebug {
		transport.Transport = &debug.DebugRequestTransport{
			RequestHeader:  true,
			RequestBody:    true,
			ResponseHeader: true,
			ResponseBody:   true,
		}
	}
	client := cos.NewClient(b, &http.Client{
		Transport: transport,
	})
	stsCli := sts.NewClient(ak, sk, &http.Client{})
	_, err = client.Bucket.Head(context.Background())
	if err != nil {
		return nil, fmt.Errorf("check bucket failed, err %s", err.Error())
	}
	return &TCloudCOS{
		prefix: prefix,
		cli:    client,
		sts:    stsCli,
		config: config,
	}, nil
}

// Upload put object to path
func (t *TCloudCOS) Upload(kt *kit.Kit, uploadPath string, r io.Reader) error {
	uploadPath = t.prependPrefix(uploadPath)
	resp, err := t.cli.Object.Put(kt.Ctx, uploadPath, r, nil)
	if err != nil {
		return fmt.Errorf("put to path %s failed, err %s", uploadPath, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("put to path %s failed, http status code %v", uploadPath, resp.StatusCode)
	}
	return nil
}

// Download get object from path
func (t *TCloudCOS) Download(kt *kit.Kit, downloadPath string, w io.Writer) error {
	downloadPath = t.prependPrefix(downloadPath)
	resp, err := t.cli.Object.Get(kt.Ctx, downloadPath, nil)
	if err != nil {
		return fmt.Errorf("get from path %s failed, err %s", downloadPath, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get from path %s failed, http status code %v", downloadPath, resp.StatusCode)
	}
	// 将响应主体复制到文件
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed writing response, err %s", err.Error())
	}
	return nil
}

// ListItems list items under path
func (t *TCloudCOS) ListItems(kt *kit.Kit, folderPath string) ([]string, error) {
	folderPath = t.prependPrefix(folderPath)
	opt := &cos.BucketGetOptions{
		// filepath join之后，最后的斜杠会被去掉，这里需要加上，不然查不出来
		Prefix:    folderPath + "/",
		Delimiter: "/",
		MaxKeys:   1000,
	}
	var marker string
	var retList []string
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, resp, err := t.cli.Bucket.Get(kt.Ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list item for path %s failed, err %s", folderPath, err.Error())
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("list item for path %s failed, http status code %v", folderPath, resp.StatusCode)
		}

		for _, content := range v.Contents {
			retList = append(retList, content.Key)
		}
		isTruncated = v.IsTruncated // 是否还有数据
		marker = v.NextMarker       // 设置下次请求的起始 key
	}
	return retList, nil
}

// Delete delete object by path
func (t *TCloudCOS) Delete(kt *kit.Kit, path string) error {
	deletePath := filepath.Join(t.prefix, path)
	_, err := t.cli.Object.Delete(kt.Ctx, deletePath)
	if err != nil {
		return err
	}
	return nil
}

// GetPreSignedURL 获取预签名URL
func (t *TCloudCOS) GetPreSignedURL(kt *kit.Kit, action OperateAction, ttl time.Duration,
	path string) (tempCred *sts.Credentials, url string, err error) {
	var cosActions []CosAction
	var httpMethod string
	switch action {
	case DownloadOperateAction:
		cosActions = []CosAction{CosActionGet}
		httpMethod = http.MethodGet
	case UploadOperateAction:
		cosActions = []CosAction{CosActionPost, CosActionPut}
		httpMethod = http.MethodPut
	default:
		return nil, "", errors.New("invalid action for get presigned url: " + string(action))
	}
	path = t.prependPrefix(path)
	tempCred, err = t.GetTemporalSecret(kt, t.config.CosBucketRegion, ttl, path, cosActions, nil)
	if err != nil {
		logs.Errorf("fail to get temporal secret for action: %s url, err: %s, ttl: %f, path: %s, rid: %s",
			action, err.Error(), ttl.Seconds(), path, kt.Rid)
		return nil, "", err
	}

	// 构造临时下载链接的访问Token
	urlToken := &URLToken{
		SessionToken: tempCred.SessionToken,
	}
	presigned, err := t.cli.Object.GetPresignedURL(kt.Ctx, httpMethod, path, tempCred.TmpSecretID,
		tempCred.TmpSecretKey, ttl, urlToken)
	if err != nil {
		logs.Errorf("fail to get presigned url for action: %s, err: %s, httpMethod: %s, ttl: %f, path: %s, rid: %s",
			action, err.Error(), httpMethod, ttl.Seconds(), path, kt.Rid)
		return nil, "", err
	}
	return tempCred, presigned.String(), nil
}

// GetTemporalSecret 获取临时密钥
func (t *TCloudCOS) GetTemporalSecret(kt *kit.Kit, region string, ttl time.Duration, path string, actions []CosAction,
	allowIps []string) (*sts.Credentials, error) {

	var strActions = make([]string, len(actions))
	for i, action := range actions {
		strActions[i] = string(action)
	}

	statement := sts.CredentialPolicyStatement{
		Action:   strActions,
		Effect:   "allow",
		Resource: []string{},
	}
	// 添加存储桶的Policy
	statement.Resource = append(statement.Resource, t.getResourcePolicy(region, path))

	if len(allowIps) > 0 {
		// 开始构建生效条件 condition
		// 关于 condition 的详细设置规则和COS支持的condition类型可以参考https://cloud.tencent.com/document/product/436/71306
		statement.Condition = map[string]map[string]interface{}{
			"ip_equal": {
				"qcs:ip": allowIps,
			},
		}
	}

	// 策略概述 https://cloud.tencent.com/document/product/436/18023
	opt := &sts.CredentialOptions{
		DurationSeconds: int64(ttl.Seconds()),
		Region:          region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{statement},
		},
	}
	credential, err := t.sts.GetCredential(opt)
	if err != nil {
		logs.Errorf("fail to get credential, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if credential.Credentials == nil {
		return nil, errors.New("credential is nil")
	}
	return credential.Credentials, nil
}

func (t *TCloudCOS) prependPrefix(path string) string {
	return filepath.Join(t.prefix, path)
}

// getResourcePolicy cos资源格式: qcs::cos:{region}:uid/{appid}:{bucket}/{path}
func (t *TCloudCOS) getResourcePolicy(region, path string) string {
	return fmt.Sprintf("qcs::cos:%s:uid/%s:%s/%s", region, t.config.UIN, t.config.CosBucketName, path)
}

// CosAction allowed post action
type CosAction string

const (
	// CosActionPost cos 表单上传操作
	CosActionPost CosAction = "name/cos:PostObject"
	// CosActionGet cos 下载操作
	CosActionGet CosAction = "name/cos:GetObject"
	// CosActionPut cos 上传操作
	CosActionPut CosAction = "name/cos:PutObject"
)

// OperateAction operate action
type OperateAction string

const (
	// DownloadOperateAction 操作Action-下载
	DownloadOperateAction OperateAction = "download"
	// UploadOperateAction 操作Action-上传
	UploadOperateAction OperateAction = "upload"
)
