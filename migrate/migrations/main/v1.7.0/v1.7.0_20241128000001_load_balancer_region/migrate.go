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

// Package migration is the migration load_balancer_region.
package migration

import (
	"context"

	"hcm/migrate/register"
	"hcm/migrate/util"
	"hcm/pkg/dal/dao/orm"
)

// ID is the unique identifier of this migration file, in the form
// <date>-<time>-<desc>-<random>. It must never change once executed.
// Migrations that share an ID run only the earlier one in execution order.
const ID = "20241128-0000-LOAD-BALANCER-REGION-96C6"

func init() {
	register.Main.Regist(ID, "v1.7.0", "20241128000001", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0029_20241128_load_balancer_region.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddConstraint(ctx, o, "load_balancer", "idx_uk_cloud_id_vendor_region", "ALTER TABLE load_balancer add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region)"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "load_balancer", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "load_balancer_listener",
		Column:  "region",
		Type:    "varchar(20)",
		NotNull: true,
		Default: util.StringDefault(""),
		After:   "default_domain",
	}); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "load_balancer_listener", "idx_uk_cloud_id_vendor_region", "ALTER TABLE load_balancer_listener add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region)"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "load_balancer_listener", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "load_balancer_listener", "idx_lb_id_cloud_id", []string{"lb_id", "cloud_id", "id"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "load_balancer_listener", "idx_vendor_account_id_bk_biz_id_cloud_lb_id", []string{"vendor", "account_id", "bk_biz_id", "cloud_lb_id"}, false); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "load_balancer_target_group", "idx_uk_cloud_id_vendor_region", "ALTER TABLE load_balancer_target_group add constraint idx_uk_cloud_id_vendor_region unique (cloud_id, vendor, region)"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "load_balancer_target_group", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "load_balancer_target_group", "idx_bk_biz_id", []string{"bk_biz_id", "id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_lb_url_rule",
		Column:  "region",
		Type:    "varchar(20)",
		NotNull: true,
		Default: util.StringDefault(""),
		After:   "cloud_target_group_id",
	}); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "tcloud_lb_url_rule", "idx_uk_cloud_id_cloud_lbl_id_region", "ALTER TABLE tcloud_lb_url_rule add constraint idx_uk_cloud_id_cloud_lbl_id_region unique (cloud_id, cloud_lbl_id, region)"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "tcloud_lb_url_rule", "idx_uk_cloud_id_lbl_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "tcloud_lb_url_rule", "idx_cloud_lbl_id_rule_type", []string{"lbl_id", "rule_type"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "tcloud_lb_url_rule", "idx_lb_id_rule_type", []string{"lb_id", "rule_type"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "load_balancer_target",
		Column:  "target_group_region",
		Type:    "varchar(20)",
		NotNull: true,
		Default: util.StringDefault(""),
		After:   "inst_name",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "load_balancer_target", "idx_target_group_id", []string{"target_group_id", "id"}, false); err != nil {
		return err
	}
	if err := util.AddConstraint(ctx, o, "load_balancer_target", "idx_uk_cloud_target_group_id_ip_port_target_group_region", "ALTER TABLE load_balancer_target add constraint idx_uk_cloud_target_group_id_ip_port_target_group_region unique (cloud_target_group_id, ip, port, target_group_region)"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "load_balancer_target", "idx_uk_cloud_target_group_id_ip_port_cloud_inst_id"); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:  "load_balancer",
		Column: "tags",
		Type:   "JSON",
		After:  "cloud_expired_time",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "target_group_listener_rule_rel", "idx_listener_rule_id", []string{"listener_rule_id"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "target_group_listener_rule_rel", "idx_lb_id_cloud_rule_id", []string{"lb_id", "cloud_listener_rule_id"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "target_group_listener_rule_rel", "idx_lb_id_cloud_lbl_id", []string{"lb_id", "cloud_lbl_id", "id"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "target_group_listener_rule_rel", "idx_lbl_id", []string{"lbl_id", "id"}, false); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "audit", "idx_bk_biz_id", []string{"bk_biz_id", "id"}, false); err != nil {
		return err
	}
	return nil
}
