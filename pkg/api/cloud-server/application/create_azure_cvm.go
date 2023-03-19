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

package application

import (
	"fmt"
	"strings"

	"hcm/pkg/criteria/validator"

	"github.com/TencentBlueKing/gopkg/collection/set"
)

var (
	sensitiveUsernames = set.NewStringSetWithValues([]string{
		"1", "123", "a", "actuser", "adm", "admin", "admin1", "admin2", "administrator", "aspnet",
		"backup", "console", "david", "guest", "john", "owner", "root", "server", "sql", "support_388945a0",
		"support", "sys", "test", "test1", "test2", "test3", "user", "user1", "user2",
		// Note: 以上是linxu和windows都敏感的用户名，以下是linux比windows多出的
		"user3", "user4", "user5", "video",
	})
)

// AzureCvmCreateReq ...
type AzureCvmCreateReq struct {
	BkBizID               int64    `json:"bk_biz_id" validate:"required,min=1"`
	AccountID             string   `json:"account_id" validate:"required"`
	ResourceGroupName     string   `json:"resource_group_name" validate:"required"`
	Region                string   `json:"region" validate:"required"`
	Zone                  string   `json:"zone" validate:"required"`
	Name                  string   `json:"name" validate:"required,min=1,max=60"`
	InstanceType          string   `json:"instance_type" validate:"required"`
	CloudImageID          string   `json:"cloud_image_id" validate:"required"`
	CloudVpcID            string   `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID         string   `json:"cloud_subnet_id" validate:"required"`
	CloudSecurityGroupIDs []string `json:"cloud_security_group_ids" validate:"required,min=1,max=1"`

	SystemDisk struct {
		// TODO: 硬盘类型待hc-service支持
		DiskType   string `json:"disk_type" validate:"required"`
		DiskSizeGB int64  `json:"disk_size_gb" validate:"required,min=20,max=36767"`
	} `json:"system_disk" validate:"required"`

	DataDisk []struct {
		// TODO: 硬盘类型待hc-service支持
		DiskType   string `json:"disk_type" validate:"required"`
		DiskSizeGB int64  `json:"disk_size_gb" validate:"required,min=20,max=36767"`
		DiskCount  int64  `json:"disk_count" validate:"required,min=1"`
	} `json:"data_disk" validate:"required"`

	// Note: 不同系统对用户名和密码要求不一样，这里暂时以Linux为主
	// https://learn.microsoft.com/en-us/azure/virtual-machines/linux/faq
	// https://learn.microsoft.com/en-us/azure/virtual-machines/windows/faq
	Username          string `json:"username" validate:"required,min=1,max=32"`
	Password          string `json:"password" validate:"required"`
	ConfirmedPassword string `json:"confirmed_password" validate:"eqfield=Password"`

	RequiredCount int64 `json:"required_count" validate:"required,min=1,max=500"`

	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *AzureCvmCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	// 校验用户名
	if err := req.validateUsername(); err != nil {
		return err
	}

	// TODO: 密码校验比较复杂，暂时不支持

	return nil
}

func (req *AzureCvmCreateReq) validateUsername() error {
	// 不能含空格
	if len(strings.Trim(req.Username, " ")) != len(req.Username) {
		return fmt.Errorf("username should not contain spaces")
	}

	// 不能是敏感用户名
	if sensitiveUsernames.Has(req.Username) {
		return fmt.Errorf(
			"username should not contain sensitive username which is [%s]",
			sensitiveUsernames.ToString(","),
		)
	}

	// 用户名最长为 20 个字符，不能以句点（“.”）结尾
	if strings.HasSuffix(req.Username, ".") {
		return fmt.Errorf("username cannot end in a period")
	}

	return nil
}
