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

// Package migration is the migration add_tenant_id.
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
const ID = "20250409-1000-ADD-TENANT-ID-5863"

func init() {
	register.Main.Regist(ID, "v1.8.2", "20250409100000", &migration{})
}

type migration struct{}

// Up applies the migration. Generated from external 0034_20250409_1000_add_tenant_id.sql.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "account", "idx_uk_name"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "account", "idx_name_tenant_id", []string{"name", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_exchange_rate",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "exchange_rate",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "account_bill_exchange_rate", "idx_uk_year_month_from_currency_to_currency"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "account_bill_exchange_rate", "idx_year_month_from_currency_to_currency_tenant_id", []string{"year", "month", "from_currency", "to_currency", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "account_bill_sync_record",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "adjustment_flow_id",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "account_bill_sync_record", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "application",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "delivery_detail",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "application", "idx_uk_source_sn"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "application", "idx_source_sn_tenant_id", []string{"source", "sn", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "approval_process",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "service_id",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "approval_process", "idx_uk_type"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "approval_process", "idx_type_tenant_id", []string{"application_type", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "audit",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "audit", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "cloud_selection_biz_type",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "deployment_architecture",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "cloud_selection_biz_type", "idx_uk_biz_type"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "cloud_selection_biz_type", "idx_biz_type_tenant_id", []string{"biz_type", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "cloud_selection_idc",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "region",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "cloud_selection_idc", "idx_uk_bk_biz_id_name"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "cloud_selection_idc", "idx_bk_biz_id_name_tenant_id", []string{"bk_biz_id", "name", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "cloud_selection_scheme",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "result_idc_ids",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "cloud_selection_scheme", "idx_uk_bk_biz_id_name"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "cloud_selection_scheme", "idx_bk_biz_id_name_tenant_id", []string{"bk_biz_id", "name", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "main_account",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "main_account", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "main_account", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "root_account",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "root_account", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "root_account", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "root_account_bill_config",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "root_account_bill_config", "idx_uk_vendor_account_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "root_account_bill_config", "idx_vendor_root_account_id_tenant_id", []string{"vendor", "root_account_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "user_collection",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "res_id",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "user_collection", "idx_uk_user_res_type_res_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "user_collection", "idx_user_res_type_res_id_tenant_id", []string{"user", "res_type", "res_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "argument_template",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "memo",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "argument_template", "idx_uk_bk_biz_id_cloud_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "argument_template", "idx_bk_biz_id_cloud_id_tenant_id", []string{"bk_biz_id", "cloud_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "aws_region",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "endpoint",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "aws_region", "idx_uk_account_id_region_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "aws_region", "idx_account_id_region_id_tenant_id", []string{"account_id", "region_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "aws_route",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "propagated",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "aws_route", "idx_uk_route_table_id_destination_cidr_block"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "aws_route", "idx_uk_route_table_id_destination_ipv6_cidr_block"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "aws_route", "idx_uk_route_table_id_cloud_dest_prefix_list_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "aws_route", "idx_route_table_id_destination_cidr_block_tenant_id", []string{"route_table_id", "destination_cidr_block", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "aws_route", "idx_route_table_id_destination_ipv6_cidr_block_tenant_id", []string{"route_table_id", "destination_ipv6_cidr_block", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "aws_route", "idx_route_table_id_cloud_dest_prefix_list_id_tenant_id", []string{"route_table_id", "cloud_destination_prefix_list_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "azure_region",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "paired_region_id",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "azure_region", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "azure_route",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "provisioning_state",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "azure_route", "idx_cloud_id"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "azure_route", "idx_uk_route_table_id_name"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "azure_route", "idx_uk_route_table_id_address_prefix"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "azure_route", "idx_cloud_id_tenant_id", []string{"cloud_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "azure_route", "idx_route_table_id_name_tenant_id", []string{"route_table_id", "name", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "azure_route", "idx_route_table_id_address_prefix_tenant_id", []string{"route_table_id", "address_prefix", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "azure_resource_group",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "account_id",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "azure_resource_group", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "cvm",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "cvm", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "cvm", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "disk",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "disk", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "eip",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "eip", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "eip", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "gcp_firewall_rule",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "bk_biz_id",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "gcp_firewall_rule", "idx_uk_cloud_id"); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "gcp_firewall_rule", "idx_uk_account_id_name"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "gcp_firewall_rule", "idx_cloud_id_tenant_id", []string{"cloud_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "gcp_firewall_rule", "idx_account_id_name_tenant_id", []string{"account_id", "name", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "gcp_region",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "self_link",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "gcp_region", "idx_uk_region_id_status"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "gcp_region", "idx_region_id_status_tenant_id", []string{"region_id", "status", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "gcp_route",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "memo",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "gcp_route", "idx_uk_cloud_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "gcp_route", "idx_cloud_id_tenant_id", []string{"cloud_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "huawei_region",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "locales_es_es",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "huawei_region", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "huawei_route",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "memo",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "huawei_route", "idx_uk_route_table_id_destination"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "huawei_route", "idx_route_table_id_destination_tenant_id", []string{"route_table_id", "destination", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "image",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "image", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "load_balancer",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "load_balancer", "idx_uk_cloud_id_vendor_region"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "load_balancer", "idx_cloud_id_vendor_region_tenant_id", []string{"cloud_id", "vendor", "region", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "network_interface",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "network_interface", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "network_interface", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "recycle_record",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "status",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "recycle_record", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "route_table",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "route_table", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "route_table", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "security_group",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "tags",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "security_group", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "security_group", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "ssl_cert",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "memo",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "ssl_cert", "idx_uk_bk_biz_id_cloud_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "ssl_cert", "idx_bk_biz_id_cloud_id_tenant_id", []string{"bk_biz_id", "cloud_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "subnet",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "subnet", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "subnet", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_region",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "status",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "tcloud_region", "idx_uk_region_id_status"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "tcloud_region", "idx_region_id_status_tenant_id", []string{"region_id", "status", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "tcloud_route",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "memo",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "tcloud_route", "idx_uk_cloud_route_table_id_cloud_id"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "tcloud_route", "idx_cloud_route_table_id_cloud_id_tenant_id", []string{"cloud_route_table_id", "cloud_id", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "vpc",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "vpc", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "vpc", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "zone",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.DropIndex(ctx, o, "zone", "idx_uk_cloud_id_vendor"); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "zone", "idx_cloud_id_vendor_tenant_id", []string{"cloud_id", "vendor", "tenant_id"}, true); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "async_flow",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "worker",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "async_flow", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "async_flow_task",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "result",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "async_flow_task", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	if err := util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "task_management",
		Column:  "tenant_id",
		Type:    "varchar(64)",
		NotNull: true,
		Default: util.StringDefault("default"),
		After:   "extension",
	}); err != nil {
		return err
	}
	if err := util.AddIndex(ctx, o, "task_management", "idx_tenant_id", []string{"tenant_id"}, false); err != nil {
		return err
	}
	return nil
}
