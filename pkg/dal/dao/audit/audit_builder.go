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
	"reflect"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/audit"
	"hcm/pkg/dal/table/cloud"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
)

// initAuditBuilder create a new auditDao builder instance.
func initAuditBuilder(kt *kit.Kit, resourceType enumor.AuditResourceType, ad *auditDao) AuditDecorator {
	ab := new(AuditBuilder)

	if len(kt.User) == 0 {
		ab.hitErr = errors.New("invalid auditDao operator")
		return ab
	}

	if len(kt.Rid) == 0 {
		ab.hitErr = errors.New("invalid auditDao request id")
		return ab
	}

	if len(resourceType) == 0 {
		ab.hitErr = errors.New("invalid auditDao resource type")
		return ab
	}

	ab.ResourceType = resourceType
	ab.Operator = kt.User
	ab.Rid = kt.Rid
	ab.AppCode = kt.AppCode
	ab.ad = ad
	ab.kt = kt

	return ab
}

// AuditDecorator is auditDao decorator interface, use to record auditDao.
type AuditDecorator interface {
	AuditCreate(txn *sqlx.Tx, cur interface{}) error
	AuditBatchCreate(txn *sqlx.Tx, cur interface{}) error
	PrepareUpdate(whereExpr string, toUpdate map[string]interface{}) AuditDecorator
	PrepareDelete(whereExpr string) AuditDecorator

	Do(txn *sqlx.Tx) error
}

// AuditBuilder is a wrapper decorator to handle all the resource's
// auditDao operation.
type AuditBuilder struct {
	hitErr error

	kt *kit.Kit
	ad *auditDao

	ResourceType enumor.AuditResourceType
	Operator     string
	Rid          string
	AppCode      string

	audits []*audit.Audit
}

// Audit return auditDao struct, that include auditDao builder auditDao basic information.
func (ab *AuditBuilder) Audit() *audit.Audit {
	return &audit.Audit{
		ResourceType: ab.ResourceType,
		Operator:     ab.Operator,
		Rid:          ab.Rid,
		AppCode:      ab.AppCode,
	}
}

// Audits return auditDao list, that include auditDao builder auditDao basic information.
func (ab *AuditBuilder) Audits(len int) []*audit.Audit {
	list := make([]*audit.Audit, len)

	for index := range list {
		list[index] = &audit.Audit{
			ResourceType: ab.ResourceType,
			Operator:     ab.Operator,
			Rid:          ab.Rid,
			AppCode:      ab.AppCode,
		}
	}

	return list
}

// AuditCreate set the resource's current details.
// Note:
// 1. must call this after the resource has already been created.
// 2. cur should be a *struct.
func (ab *AuditBuilder) AuditCreate(txn *sqlx.Tx, cur interface{}) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	one := ab.Audit()
	one.Action = enumor.Create
	one.Detail = &audit.AuditBasicDetail{
		Data: cur,
	}

	switch res := cur.(type) {
	case *tablecloud.AccountTable:
		one.ResourceID = res.ID
		one.AccountID = res.ID

	default:
		logs.Errorf("unsupported auditDao create resource: %s, type: %s, rid: %v", ab.ResourceType,
			reflect.TypeOf(cur), ab.Rid)
		return fmt.Errorf("unsupported auditDao create resource: %s", ab.ResourceType)
	}

	return ab.ad.One(ab.kt, txn, one)
}

// AuditBatchCreate set the resource's current details.
// Note:
// 1. must call this after the resource has already been created.
// 2. cur should be a slice.
func (ab *AuditBuilder) AuditBatchCreate(txn *sqlx.Tx, cur interface{}) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	var audits []*audit.Audit
	switch ress := cur.(type) {
	// TODO: 账号不会存在批量创建操作，之后添加新的审计时，将该逻辑删除
	case []cloud.AccountTable:
		audits = ab.Audits(len(ress))
		for index := range audits {
			audits[index].Action = enumor.Create
			audits[index].ResourceID = ress[index].ID
			audits[index].AccountID = ress[index].ID
			audits[index].Detail = &audit.AuditBasicDetail{
				Data: ress[index],
			}
		}

	default:
		logs.Errorf("unsupported auditDao batch create resource: %s, type: %s, rid: %v", ab.ResourceType,
			reflect.TypeOf(cur), ab.Rid)
		return fmt.Errorf("unsupported auditDao batch create resource: %s", ab.ResourceType)
	}

	return ab.ad.Insert(ab.kt, txn, audits)
}

// PrepareUpdate prepare the resource's previous instance details by
// get the instance's detail from db and save it to ab.data for later use.
// Note:
// 1. call this before resource is updated.
// 2. updatedTo means 'to be updated to data', it should be a *struct.
func (ab *AuditBuilder) PrepareUpdate(whereExpr string, toUpdate map[string]interface{}) AuditDecorator {
	if ab.hitErr != nil {
		return ab
	}

	var audits []*audit.Audit
	var err error
	switch ab.ResourceType {
	case enumor.Account:
		audits, err = ab.decorateAccountUpdate(whereExpr, toUpdate)
		if err != nil {
			ab.hitErr = err
			return ab
		}

	default:
		logs.Errorf("unsupported auditDao update resource type: %s, rid: %v", ab.ResourceType, ab.Rid)
		ab.hitErr = fmt.Errorf("unsupported auditDao update resource type: %s", ab.ResourceType)
		return ab
	}

	ab.audits = audits
	return ab
}

func (ab *AuditBuilder) decorateAccountUpdate(whereExpr string, toUpdate map[string]interface{}) (
	[]*audit.Audit, error,
) {
	accounts, err := ab.listAccount(whereExpr)
	if err != nil {
		return nil, err
	}

	audits := ab.Audits(len(accounts))
	for index, one := range accounts {
		audits[index].Action = enumor.Update
		audits[index].ResourceID = one.ID
		audits[index].AccountID = one.ID
		audits[index].Detail = &audit.AuditBasicDetail{
			Data:    one,
			Changed: toUpdate,
		}
	}

	return audits, nil
}

// PrepareDelete prepare the resource's previous instance details by
// get the instance's detail from db and save it to ab.data for later use.
// Note: call this before resource is deleted.
func (ab *AuditBuilder) PrepareDelete(whereExpr string) AuditDecorator {
	if ab.hitErr != nil {
		return ab
	}

	var audits []*audit.Audit
	var err error
	switch ab.ResourceType {
	case enumor.Account:
		audits, err = ab.decorateAccountDelete(whereExpr)
		if err != nil {
			ab.hitErr = err
			return ab
		}

	default:
		ab.hitErr = fmt.Errorf("unsupported auditDao deleted resource: %s", ab.ResourceType)
		return ab
	}

	ab.audits = audits
	return ab
}

func (ab *AuditBuilder) decorateAccountDelete(whereExpr string) (
	[]*audit.Audit, error,
) {
	accounts, err := ab.listAccount(whereExpr)
	if err != nil {
		return nil, err
	}

	audits := ab.Audits(len(accounts))
	for index, one := range accounts {
		audits[index].Action = enumor.Delete
		audits[index].ResourceID = one.ID
		audits[index].AccountID = one.ID
		audits[index].Detail = &audit.AuditBasicDetail{
			Data: one,
		}
	}

	return audits, nil
}

// Do save auditDao log to the db immediately.
func (ab *AuditBuilder) Do(txn *sqlx.Tx) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	if ab.audits == nil || len(ab.audits) == 0 {
		ab.hitErr = fmt.Errorf("insert auditDao is empty")
	}

	return ab.ad.Insert(ab.kt, txn, ab.audits)
}

func (ab *AuditBuilder) listAccount(whereExpr string) ([]*cloud.AccountTable, error) {
	// sql := fmt.Sprintf(`SELECT %s FROM %s %s`, cloud.AccountColumns.NamedExpr(), "account", whereExpr)
	//
	// list := make([]*cloud.AccountTable, 0)
	// err := ab.ad.orm.Do().Select(ab.kt.Ctx, &list, sql)
	// if err != nil {
	// 	return nil, fmt.Errorf("select account failed, err: %v", err)
	// }

	return nil, nil
}
