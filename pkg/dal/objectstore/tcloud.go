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
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

const (
	envTCloudCOSPrefix    = "TCLOUD_COS_PREFIX"
	envTCloudCOSSecretID  = "TCLOUD_COS_SECRET_ID"
	envTCloudCOSSecretKey = "TCLOUD_COS_SECRET_KEY"
	envTCloudCOSBucketURL = "TCLOUD_COS_BUCKET_URL"
)

// TCloudCOS tcloud cos client
type TCloudCOS struct {
	prefix string
	cli    *cos.Client
}

// NewTCloudCOS create cos client
func NewTCloudCOS() (*TCloudCOS, error) {
	prefix := os.Getenv(envTCloudCOSPrefix)
	id := os.Getenv(envTCloudCOSSecretID)
	key := os.Getenv(envTCloudCOSSecretKey)
	bucketURL := os.Getenv(envTCloudCOSBucketURL)
	if len(id) == 0 || len(key) == 0 || len(bucketURL) == 0 {
		return nil, fmt.Errorf("any of env %s, %s, %s cannot be empty",
			envTCloudCOSSecretID, envTCloudCOSSecretKey, envTCloudCOSBucketURL)
	}
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("parse bucket url %s failed, err %s", bucketURL, err.Error())
	}
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  id,
			SecretKey: key,
		},
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
		Prefix:    folderPath,
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
