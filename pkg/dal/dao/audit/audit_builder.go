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
	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
)

// initAuditBuilder create a new audit builder instance.
func initAuditBuilder(kt *kit.Kit, resourceType enumor.AuditResourceType, ad *audit) AuditDecorator {
	ab := new(AuditBuilder)

	if len(kt.User) == 0 {
		ab.hitErr = errors.New("invalid audit operator")
		return ab
	}

	if len(kt.Rid) == 0 {
		ab.hitErr = errors.New("invalid audit request id")
		return ab
	}

	if len(resourceType) == 0 {
		ab.hitErr = errors.New("invalid audit resource type")
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

// AuditDecorator is audit decorator interface, use to record audit.
type AuditDecorator interface {
	AuditCreate(txn *sqlx.Tx, cur interface{}) error
	AuditBatchCreate(txn *sqlx.Tx, cur interface{}) error
	PrepareUpdate(whereExpr string, toUpdate map[string]interface{}) AuditDecorator
	PrepareDelete(whereExpr string) AuditDecorator

	Do(txn *sqlx.Tx) error
}

// AuditBuilder is a wrapper decorator to handle all the resource's
// audit operation.
type AuditBuilder struct {
	hitErr error

	kt *kit.Kit
	ad *audit

	ResourceType enumor.AuditResourceType
	Operator     string
	Rid          string
	AppCode      string

	audits []*table.Audit
}

// Audit return audit struct, that include audit builder audit basic information.
func (ab *AuditBuilder) Audit() *table.Audit {
	return &table.Audit{
		ResourceType: ab.ResourceType,
		Operator:     ab.Operator,
		Rid:          ab.Rid,
		AppCode:      ab.AppCode,
	}
}

// Audits return audit list, that include audit builder audit basic information.
func (ab *AuditBuilder) Audits(len int) []*table.Audit {
	list := make([]*table.Audit, len)

	for index := range list {
		list[index] = &table.Audit{
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
	one.Detail = &table.AuditBasicDetail{
		Data: cur,
	}

	switch res := cur.(type) {
	case *tablecloud.AccountModel:
		one.ResourceID = res.ID
		one.AccountID = res.ID

	default:
		logs.Errorf("unsupported audit create resource: %s, type: %s, rid: %v", ab.ResourceType,
			reflect.TypeOf(cur), ab.Rid)
		return fmt.Errorf("unsupported audit create resource: %s", ab.ResourceType)
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

	var audits []*table.Audit
	switch ress := cur.(type) {
	// TODO: 账号不会存在批量创建操作，之后添加新的审计时，将该逻辑删除
	case []table.Account:
		audits = ab.Audits(len(ress))
		for index := range audits {
			audits[index].Action = enumor.Create
			audits[index].ResourceID = ress[index].ID
			audits[index].AccountID = ress[index].ID
			audits[index].Detail = &table.AuditBasicDetail{
				Data: ress[index],
			}
		}

	default:
		logs.Errorf("unsupported audit batch create resource: %s, type: %s, rid: %v", ab.ResourceType,
			reflect.TypeOf(cur), ab.Rid)
		return fmt.Errorf("unsupported audit batch create resource: %s", ab.ResourceType)
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

	var audits []*table.Audit
	var err error
	switch ab.ResourceType {
	case enumor.Account:
		audits, err = ab.decorateAccountUpdate(whereExpr, toUpdate)
		if err != nil {
			ab.hitErr = err
			return ab
		}

	default:
		logs.Errorf("unsupported audit update resource type: %s, rid: %v", ab.ResourceType, ab.Rid)
		ab.hitErr = fmt.Errorf("unsupported audit update resource type: %s", ab.ResourceType)
		return ab
	}

	ab.audits = audits
	return ab
}

func (ab *AuditBuilder) decorateAccountUpdate(whereExpr string, toUpdate map[string]interface{}) (
	[]*table.Audit, error,
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
		audits[index].Detail = &table.AuditBasicDetail{
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

	var audits []*table.Audit
	var err error
	switch ab.ResourceType {
	case enumor.Account:
		audits, err = ab.decorateAccountDelete(whereExpr)
		if err != nil {
			ab.hitErr = err
			return ab
		}

	default:
		ab.hitErr = fmt.Errorf("unsupported audit deleted resource: %s", ab.ResourceType)
		return ab
	}

	ab.audits = audits
	return ab
}

func (ab *AuditBuilder) decorateAccountDelete(whereExpr string) (
	[]*table.Audit, error,
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
		audits[index].Detail = &table.AuditBasicDetail{
			Data: one,
		}
	}

	return audits, nil
}

// Do save audit log to the db immediately.
func (ab *AuditBuilder) Do(txn *sqlx.Tx) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	if ab.audits == nil || len(ab.audits) == 0 {
		ab.hitErr = fmt.Errorf("insert audit is empty")
	}

	return ab.ad.Insert(ab.kt, txn, ab.audits)
}

func (ab *AuditBuilder) listAccount(whereExpr string) ([]*table.Account, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s %s`, table.AccountColumns.NamedExpr(), "account", whereExpr)

	list := make([]*table.Account, 0)
	err := ab.ad.orm.Do().Select(ab.kt.Ctx, &list, sql)
	if err != nil {
		return nil, fmt.Errorf("select account failed, err: %v", err)
	}

	return list, nil
}
