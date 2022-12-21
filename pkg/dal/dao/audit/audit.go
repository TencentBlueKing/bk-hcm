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

package audit

import (
	"errors"
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/audit"
	"hcm/pkg/kit"

	"github.com/jmoiron/sqlx"
)

// AuditDao supplies all the auditDao operations.
type AuditDao interface {
	// Decorator is used to handle the auditDao process as a pipeline according CUD scenarios.
	Decorator(kit *kit.Kit, res enumor.AuditResourceType) AuditDecorator
	// Insert resource's auditDao.
	Insert(kit *kit.Kit, txn *sqlx.Tx, audits []*audit.Audit) error
}

var _ AuditDao = new(auditDao)

// NewAuditDao create the auditDao DAO
func NewAuditDao(orm orm.Interface, db *sqlx.DB) (AuditDao, error) {
	return &auditDao{
		orm: orm,
		db:  db,
	}, nil
}

type auditDao struct {
	orm orm.Interface
	// db is the auditDao's instance
	db *sqlx.DB
}

// Decorator return auditDao decorator for to record auditDao.
func (au *auditDao) Decorator(kit *kit.Kit, res enumor.AuditResourceType) AuditDecorator {
	return initAuditBuilder(kit, res, au)
}

// One auditDao one resource's operation.
func (au *auditDao) One(kt *kit.Kit, txn *sqlx.Tx, one *audit.Audit) error {
	if one == nil {
		return errors.New("invalid input auditDao or opt")
	}

	if err := one.CreateValidate(); err != nil {
		return fmt.Errorf("auditDao create validate failed, err: %v", err)
	}

	one.TenantID = kt.TenantID
	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, one.TableName(),
		audit.AuditColumns.ColumnExpr(), audit.AuditColumns.ColonNameExpr())

	// do with the same transaction with the resource, this transaction
	// is launched by resource's owner.
	if err := au.orm.Txn(txn).Insert(kt.Ctx, sql, one); err != nil {
		return fmt.Errorf("insert auditDao failed, err: %v", err)
	}

	return nil
}

// Insert auditDao resource's operation.
func (au *auditDao) Insert(kt *kit.Kit, txn *sqlx.Tx, audits []*audit.Audit) error {
	if audits == nil {
		return errors.New("invalid input audits or opt")
	}

	for idx := range audits {
		if err := audits[idx].CreateValidate(); err != nil {
			return fmt.Errorf("auditDao create validate failed, err: %v", err)
		}
		audits[idx].TenantID = kt.TenantID
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, table.AuditTable,
		audit.AuditColumns.ColumnExpr(), audit.AuditColumns.ColonNameExpr())

	// do with the same transaction with the resource, this transaction
	// is launched by resource's owner.
	if err := au.orm.Txn(txn).BulkInsert(kt.Ctx, sql, audits); err != nil {
		return fmt.Errorf("insert audits failed, err: %v", err)
	}

	return nil
}
