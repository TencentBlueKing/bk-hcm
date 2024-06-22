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

	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
)

// GetObjectStore get object store from env
func GetObjectStore(config cc.ObjectStore) (Storage, error) {
	switch config.Type {
	case "":
		return nil, nil
	case string(enumor.TCloud):
		return NewTCloudCOS(config.ObjectStoreTCloud)
	default:
		return nil, fmt.Errorf("invalid object store type %s", config.Type)
	}
}

// Storage the interface of storage
type Storage interface {
	Upload(ctx context.Context, uploadPath string, r io.Reader) error
	Download(ctx context.Context, downloadPath string, w io.Writer) error
	ListItems(ctx context.Context, folderPath string) ([]string, error)
}
