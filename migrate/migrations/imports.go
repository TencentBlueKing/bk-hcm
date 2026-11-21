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

// Package migrations pulls every migration package into the build. Each blank
// import runs that migration's init, which registers it. The lines are kept by
// scripts/new-migrate.sh and scripts/release-migrate.sh, and checked by
// scripts/check-migrate.sh.
package migrations

import (
	_ "hcm/migrate/migrations/main/v1.0.0/v1.0.0_20230227204500_init_db"
	_ "hcm/migrate/migrations/main/v1.1.0/v1.1.0_20230329151000_azure_sg_rule_uk_and_cleanup"
	_ "hcm/migrate/migrations/main/v1.1.18/v1.1.18_20230710160000_utf8mb4_bin_and_fk_cascade"
	_ "hcm/migrate/migrations/main/v1.1.20/v1.1.20_20230727104000_gcp_firewall_rule_uk"
	_ "hcm/migrate/migrations/main/v1.1.21/v1.1.21_20230727192200_security_group_name_len"
	_ "hcm/migrate/migrations/main/v1.1.22/v1.1.22_20230731111700_aws_sg_rule_memo"
	_ "hcm/migrate/migrations/main/v1.1.25/v1.1.25_20230821194900_sg_rule_memo_cvm_zone"
	_ "hcm/migrate/migrations/main/v1.1.27/v1.1.27_20231019201000_tcloud_route_uk"
	_ "hcm/migrate/migrations/main/v1.1.3/v1.1.3_20230516160000_account_bill_config"
	_ "hcm/migrate/migrations/main/v1.1.5/v1.1.5_20230526141000_tcloud_sg_rule_memo"
	_ "hcm/migrate/migrations/main/v1.1.7/v1.1.7_20230530210000_approval_process_managers"
	_ "hcm/migrate/migrations/main/v1.2.1/v1.2.1_20231130160400_sub_account_async_flow"
	_ "hcm/migrate/migrations/main/v1.3.0/v1.3.0_20231211101500_cloud_selection"
	_ "hcm/migrate/migrations/main/v1.3.3/v1.3.3_20231228160100_sub_account_uk"
	_ "hcm/migrate/migrations/main/v1.4.0/v1.4.0_20240305100000_argument_template"
	_ "hcm/migrate/migrations/main/v1.4.4/v1.4.4_20240516163000_application_source"
	_ "hcm/migrate/migrations/main/v1.5.0/v1.5.0_20240521160000_tcloud_cert"
	_ "hcm/migrate/migrations/main/v1.5.0/v1.5.0_20240521170000_tcloud_clb"
	_ "hcm/migrate/migrations/main/v1.5.2/v1.5.2_20240617000000_cross_region_rs"
	_ "hcm/migrate/migrations/main/v1.6.0/v1.6.0_20240531164800_bill"
	_ "hcm/migrate/migrations/main/v1.6.0/v1.6.0_20240603170000_account"
	_ "hcm/migrate/migrations/main/v1.6.0/v1.6.0_20240716175700_application_biz_id"
	_ "hcm/migrate/migrations/main/v1.6.1/v1.6.1_20240814114500_bill_account_cloud_id"
	_ "hcm/migrate/migrations/main/v1.6.10/v1.6.10_20241030000000_async_table_optimize"
	_ "hcm/migrate/migrations/main/v1.6.11/v1.6.11_20241017000000_security_group"
	_ "hcm/migrate/migrations/main/v1.6.8/v1.6.8_20241009000000_multiple_month_task"
	_ "hcm/migrate/migrations/main/v1.7.0/v1.7.0_20241128000000_task_management"
	_ "hcm/migrate/migrations/main/v1.7.0/v1.7.0_20241128000001_load_balancer_region"
	_ "hcm/migrate/migrations/main/v1.7.2/v1.7.2_20250108110000_global_config"
	_ "hcm/migrate/migrations/main/v1.8.0/v1.8.0_20250311000000_security_group_mgmt"
	_ "hcm/migrate/migrations/main/v1.8.0/v1.8.0_20250311000001_assign_host"
	_ "hcm/migrate/migrations/main/v1.8.1/v1.8.1_20250612000000_tenant"
	_ "hcm/migrate/migrations/main/v1.8.11/v1.8.11_20251219000000_image"
	_ "hcm/migrate/migrations/main/v1.8.11/v1.8.11_20260319100000_permission_policy_library"
	_ "hcm/migrate/migrations/main/v1.8.12/v1.8.12_20260415000000_application_operation"
	_ "hcm/migrate/migrations/main/v1.8.2/v1.8.2_20250409100000_add_tenant_id"
	_ "hcm/migrate/migrations/main/v1.8.2/v1.8.2_20250718174322_account_biz_id"
	_ "hcm/migrate/migrations/main/v1.8.2/v1.8.2_20250729103729_clb_import"
	_ "hcm/migrate/migrations/main/v1.8.2/v1.8.2_20250729103730_load_balancer_band_width"
	_ "hcm/migrate/migrations/main/v1.8.2/v1.8.2_20250729103731_load_balancer_sync_time"
	_ "hcm/migrate/migrations/main/v1.8.3/v1.8.3_20250731152719_account"
	_ "hcm/migrate/migrations/main/v1.8.5/v1.8.5_20250814205541_ssl_cert"
	_ "hcm/migrate/migrations/main/v1.8.7/v1.8.7_20250820000000_add_bk_biz_id"
	_ "hcm/migrate/migrations/main/v1.8.8/v1.8.8_20251013000000_extract_bk_asset_id"
	_ "hcm/migrate/migrations/main/v1.8.9/v1.8.9_20260121000000_zone"
	_ "hcm/migrate/migrations/main/v1.9.0/v1.9.0_20260130180000_account_secret"
	_ "hcm/migrate/migrations/main/v1.9.0/v1.9.0_20260131180000_sub_account_secret"
	_ "hcm/migrate/migrations/main/v1.9.0/v1.9.0_20260202095800_account_fields"
	_ "hcm/migrate/migrations/main/v1.9.0/v1.9.0_20260202100000_sub_account_fields"
	_ "hcm/migrate/migrations/main/v1.9.0/v1.9.0_20260324100100_permission_template"
	_ "hcm/migrate/migrations/main/v1.9.0/v1.9.0_20260518210016_account_bill_adjustment_item_add_res_class"
	_ "hcm/migrate/migrations/main/v1.9.2/v1.9.2_20260609000000_cvm_add_gpu"
)
