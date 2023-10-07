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

package recyclerecord

import (
	"hcm/pkg/tools/json"
)

// BaseRecycleDetail basic recycle record detail.
type BaseRecycleDetail struct {
	ErrorMessage string `json:"error_message,omitempty"`
}

// CvmRecycleDetail 包含回收选项、disk，eip 挂载选项
type CvmRecycleDetail struct {
	CvmRecycleOptions `json:",inline"`
	DiskList          []DiskAttachInfo `json:"disk_list"`
	EipList           []EipBindInfo    `json:"eip_list"`
	ErrorMessage      string           `json:"error_message,omitempty"`
}

// DiskAttachInfo 磁盘挂载信息
type DiskAttachInfo struct {
	DiskID      string `json:"disk_id,omitempty"`
	CachingType string `json:"caching_type,omitempty"`
	DeviceName  string `json:"device_name,omitempty"`
	Err         error  `json:"-"`
}

// GetCloudID ...
func (d DiskAttachInfo) GetCloudID() string {
	return d.DiskID
}

// GetID ...
func (d DiskAttachInfo) GetID() string {
	return d.DiskID
}

func (d DiskAttachInfo) String() string {
	str, err := json.MarshalToString(d)
	if err != nil {
		return err.Error()
	}
	return str
}

// EipBindInfo eip 绑定信息
type EipBindInfo struct {
	EipID string `json:"eip_id"`
	NicID string `json:"nic_id"`
	Err   error  `json:"-"`
}

// GetCloudID ...
func (e EipBindInfo) GetCloudID() string {
	return e.EipID
}

// GetID ...
func (e EipBindInfo) GetID() string {
	return e.EipID
}

func (e EipBindInfo) String() string {
	str, err := json.MarshalToString(e)
	if err != nil {
		return err.Error()
	}
	return str
}

// CvmRecycleOptions cvm recycle record options.
type CvmRecycleOptions struct {
	WithDisk bool `json:"with_disk"`
	WithEip  bool `json:"with_eip"`
}

// DiskRecycleOptions disk recycle record options.
type DiskRecycleOptions struct{}

// DiskRelatedRecycleOpt 磁盘作为关联资源回收时的回收选项，记录关联的cvm_id
type DiskRelatedRecycleOpt struct {
	CvmID string `json:"cvm_id"`
}
