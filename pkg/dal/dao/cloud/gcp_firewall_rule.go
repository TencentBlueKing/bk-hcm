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

package cloud

import (
	"github.com/jmoiron/sqlx"

	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/runtime/filter"
)

// GcpFirewallRule only used for gcp firewall rule.
type GcpFirewallRule interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []cloud.GcpFirewallRuleTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, rule *cloud.GcpFirewallRuleTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListGcpFirewallRuleDetails, error)
	Delete(kt *kit.Kit, expr *filter.Expression) error
}

var _ GcpFirewallRule = new(GcpFirewallRuleDao)

// GcpFirewallRuleDao gcp firewall rule dao.
type GcpFirewallRuleDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// BatchCreateWithTx batch create with tx.
func (g GcpFirewallRuleDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []cloud.GcpFirewallRuleTable) ([]string,
	error) {

	// TODO implement me
	panic("implement me")
}

// UpdateWithTx update with tx.
func (g GcpFirewallRuleDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
	rule *cloud.GcpFirewallRuleTable) error {

	// TODO implement me
	panic("implement me")
}

// List gcp firewall rules.
func (g GcpFirewallRuleDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListGcpFirewallRuleDetails, error) {
	// TODO implement me
	panic("implement me")
}

// Delete gcp firewall rules.
func (g GcpFirewallRuleDao) Delete(kt *kit.Kit, expr *filter.Expression) error {
	// TODO implement me
	panic("implement me")
}
