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
	"hcm/pkg/kit"

	"github.com/jmoiron/sqlx"
)

// AuditDao supplies all the audit operations.
type AuditDao interface {
	// Decorator is used to handle the audit process as a pipeline according CUD scenarios.
	Decorator(kit *kit.Kit, res enumor.AuditResourceType) AuditDecorator
	// Insert resource's audit.
	Insert(kit *kit.Kit, txn *sqlx.Tx, audits []*table.Audit) error
}

var _ AuditDao = new(audit)

// NewAuditDao create the audit DAO
func NewAuditDao(orm orm.Interface, db *sqlx.DB) (AuditDao, error) {
	return &audit{
		orm: orm,
		db:  db,
	}, nil
}

type audit struct {
	orm orm.Interface
	// db is the audit's instance
	db *sqlx.DB
}

// Decorator return audit decorator for to record audit.
func (au *audit) Decorator(kit *kit.Kit, res enumor.AuditResourceType) AuditDecorator {
	return initAuditBuilder(kit, res, au)
}

// One audit one resource's operation.
func (au *audit) One(kit *kit.Kit, txn *sqlx.Tx, audit *table.Audit) error {
	if audit == nil {
		return errors.New("invalid input audit or opt")
	}

	if err := audit.CreateValidate(); err != nil {
		return fmt.Errorf("audit create validate failed, err: %v", err)
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, audit.TableName(),
		table.AuditColumns.ColumnExpr(), table.AuditColumns.ColonNameExpr())

	// do with the same transaction with the resource, this transaction
	// is launched by resource's owner.
	if _, err := au.orm.Txn(txn).Insert(kit.Ctx, sql, audit); err != nil {
		return fmt.Errorf("insert audit failed, err: %v", err)
	}

	return nil
}

// Insert audit resource's operation.
func (au *audit) Insert(kit *kit.Kit, txn *sqlx.Tx, audits []*table.Audit) error {
	if audits == nil {
		return errors.New("invalid input audits or opt")
	}

	for _, one := range audits {
		if err := one.CreateValidate(); err != nil {
			return fmt.Errorf("audit create validate failed, err: %v", err)
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, table.AuditTable,
		table.AuditColumns.ColumnExpr(), table.AuditColumns.ColonNameExpr())

	// do with the same transaction with the resource, this transaction
	// is launched by resource's owner.
	if _, err := au.orm.Txn(txn).BulkInsert(kit.Ctx, sql, audits); err != nil {
		return fmt.Errorf("insert audits failed, err: %v", err)
	}

	return nil
}
