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

// Package cos ...
package cos

import (
	"fmt"
	"net/url"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// UrlType cos url type
type UrlType string

const (
	// NormalUrl cos normal url
	NormalUrl UrlType = "cos_normal_url"
	// UrlWithNameAndRegion cos url with name and region
	UrlWithNameAndRegion UrlType = "cos_url_with_name_and_region"
)

// ClientOpt defines cos client options.
type ClientOpt struct {
	UrlType         UrlType
	Region          string
	BucketNameAppID string
}

// GetUrl get cos url
func (o *ClientOpt) GetUrl(urlMap map[UrlType]string) (*cos.BaseURL, error) {
	switch o.UrlType {
	case NormalUrl:
		serviceUrl, err := url.Parse(urlMap[NormalUrl])
		if err != nil {
			return nil, err
		}
		return &cos.BaseURL{ServiceURL: serviceUrl}, nil
	case UrlWithNameAndRegion:
		bucketUrl, err := url.Parse(fmt.Sprintf(urlMap[UrlWithNameAndRegion], o.BucketNameAppID, o.Region))
		if err != nil {
			return nil, err
		}
		return &cos.BaseURL{BucketURL: bucketUrl}, nil
	default:
		return nil, fmt.Errorf("unknown url type: %s", o.UrlType)
	}
}

// TCloudBucketCreateOption defines tencent cloud create bucket options.
type TCloudBucketCreateOption struct {
	Name   string `json:"name" validate:"required"`
	Region string `json:"region" validate:"required"`
	// AppID is the application ID for the bucket.
	AppID uint64 `json:"-" validate:"required"`

	// XCosACL 定义存储桶的访问控制列表（ACL）属性, 默认为 private
	XCosACL string `json:"x_cos_acl" validate:"omitempty"`
	// XCosGrantRead 赋予被授权者读取存储桶的权限
	XCosGrantRead string `json:"x_cos_grant_read" validate:"omitempty"`
	// XCosGrantWrite 赋予被授权者写入存储桶的权限
	XCosGrantWrite string `json:"x_cos_grant_write" validate:"omitempty"`
	// XCosGrantFullControl 赋予被授权者操作存储桶的所有权限
	XCosGrantFullControl string `json:"x_cos_grant_full_control" validate:"omitempty"`
	// XCosGrantReadACP 赋予被授权者读取存储桶的访问控制列表（ACL）和存储桶策略（Policy）的权限
	XCosGrantReadACP string `json:"x_cos_grant_read_acp" validate:"omitempty"`
	// XCosGrantWriteACP 赋予被授权者写入存储桶的访问控制列表（ACL）和存储桶策略（Policy）的权限
	XCosGrantWriteACP string `json:"x_cos_grant_write_acp" validate:"omitempty"`
	// XCosTagging 在创建存储桶的同时，为存储桶添加标签，最多可设置50个标签。例如 key1=value1&key2=value2
	XCosTagging string `json:"x_cos_tagging" validate:"omitempty"`
	// CreateBucketConfiguration 包含 PUT Bucket 操作的所有请求信息
	CreateBucketConfiguration *CreateBucketConfiguration `json:"create_bucket_configuration" validate:"omitempty"`
}

// Validate TCloudBucketCreateOption.
func (c TCloudBucketCreateOption) Validate() error {
	if c.CreateBucketConfiguration != nil {
		if err := c.CreateBucketConfiguration.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(c)
}

// CreateBucketConfiguration defines tencent cloud create bucket configuration.
type CreateBucketConfiguration struct {
	BucketAZConfig enumor.BucketAZConfig `json:"bucket_az_config" validate:"required"`
}

// Validate CreateBucketConfiguration.
func (c CreateBucketConfiguration) Validate() error {
	if err := c.BucketAZConfig.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(c)
}

// TCloudBucketDeleteOption defines tencent cloud delete bucket options.
type TCloudBucketDeleteOption struct {
	Name   string `json:"name" validate:"required"`
	Region string `json:"region" validate:"required"`
	// AppID is the application ID for the bucket.
	AppID uint64 `json:"-" validate:"required"`
}

// Validate TCloudBucketDeleteOption.
func (c TCloudBucketDeleteOption) Validate() error {
	return validator.Validate.Struct(c)
}

// TCloudBucketListOption defines tencent cloud list bucket options.
type TCloudBucketListOption struct {
	TagKey     *string `json:"tag_key" validate:"omitempty"`
	TagValue   *string `json:"tag_value" validate:"omitempty"`
	MaxKeys    *int64  `json:"max_keys" validate:"omitempty"`
	Marker     *string `json:"marker" validate:"omitempty"`
	Range      *string `json:"range" validate:"omitempty"`
	CreateTime *int64  `json:"create_time" validate:"omitempty"`
	Region     *string `json:"region" validate:"omitempty"`
}

// Validate TCloudBucketListOption.
func (c TCloudBucketListOption) Validate() error {
	return validator.Validate.Struct(c)
}

// TCloudBucketListResult defines tencent cloud list bucket result.
type TCloudBucketListResult struct {
	Owner       *Owner   `json:"owner"`
	Buckets     []Bucket `json:"buckets"`
	Marker      string   `json:"marker"`
	NextMarker  string   `json:"next_marker"`
	IsTruncated bool     `json:"is_truncated"`
}

// Owner defines Bucket/Object's owner
type Owner struct {
	UIN         string `json:"uin"`
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// Bucket defines tencent cloud bucket.
type Bucket struct {
	// CloudName 腾讯云接口返回的bucket名称（带AppID后缀）
	CloudName    string `json:"cloud_name"`
	Region       string `json:"region"`
	CreationDate string `json:"creation_date"`
	BucketType   string `json:"bucket_type"`
}
