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
	"fmt"
	"hcm/pkg/cc"
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	cos "github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
)

// TCloudCOS tcloud cos client
type TCloudCOS struct {
	prefix string
	cli    *cos.Client
}

// NewTCloudCOS create cos client
func NewTCloudCOS(config cc.ObjectStoreTCloud) (*TCloudCOS, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	prefix := config.COSPrefix
	id := config.COSSecretID
	key := config.COSSecretKey
	bucketURL := config.COSBucketURL
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("parse bucket url %s failed, err %s", bucketURL, err.Error())
	}
	b := &cos.BaseURL{BucketURL: u}
	transport := &cos.AuthorizationTransport{
		SecretID:  id,
		SecretKey: key,
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
	_, err = client.Bucket.Head(context.Background())
	if err != nil {
		return nil, fmt.Errorf("check bucket failed, err %s", err.Error())
	}
	return &TCloudCOS{
		prefix: prefix,
		cli:    client,
	}, nil
}

// Upload put object to path
func (t *TCloudCOS) Upload(ctx context.Context, uploadPath string, r io.Reader) error {
	uploadPath = filepath.Join(t.prefix, uploadPath)
	resp, err := t.cli.Object.Put(ctx, uploadPath, r, nil)
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
func (t *TCloudCOS) Download(ctx context.Context, downloadPath string, w io.Writer) error {
	downloadPath = filepath.Join(t.prefix, downloadPath)
	resp, err := t.cli.Object.Get(ctx, downloadPath, nil)
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
func (t *TCloudCOS) ListItems(ctx context.Context, folderPath string) ([]string, error) {
	folderPath = filepath.Join(t.prefix, folderPath)
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
		v, resp, err := t.cli.Bucket.Get(context.Background(), opt)
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
