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
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
)

// initAuditBuilder create a new audit builder instance.
func initAuditBuilder(kit *kit.Kit, res enumor.AuditResourceType, ad *audit) AuditDecorator {

	ab := &AuditBuilder{
		toAudit: &table.Audit{
			ResourceType: res,
			Operator:     kit.User,
			Rid:          kit.Rid,
			AppCode:      kit.AppCode,
		},
		ad:  ad,
		kit: kit,
	}

	if len(kit.User) == 0 {
		ab.hitErr = errors.New("invalid audit operator")
		return ab
	}

	if len(kit.Rid) == 0 {
		ab.hitErr = errors.New("invalid audit request id")
		return ab
	}

	if len(res) == 0 {
		ab.hitErr = errors.New("invalid audit resource type")
		return ab
	}

	return ab
}

// AuditDecorator is audit decorator interface, use to record audit.
type AuditDecorator interface {
	AuditCreate(txn *sqlx.Tx, cur interface{}) error
	PrepareUpdate(updatedTo interface{}) AuditDecorator
	PrepareDelete(resID uint64) AuditDecorator
	Do(txn *sqlx.Tx) error
}

// AuditBuilder is a wrapper decorator to handle all the resource's
// audit operation.
type AuditBuilder struct {
	hitErr error

	toAudit *table.Audit
	kit     *kit.Kit
	data    interface{}
	changed map[string]interface{}
	ad      *audit
}

// AuditCreate set the resource's current details.
// Note:
// 1. must call this after the resource has already been created.
// 2. cur should be a *struct.
func (ab *AuditBuilder) AuditCreate(txn *sqlx.Tx, cur interface{}) error {
	if ab.hitErr != nil {
		return ab.hitErr
	}

	ab.toAudit.Action = enumor.Create
	ab.data = cur

	switch res := cur.(type) {
	case *table.Account:
		ab.toAudit.ResourceID = res.ID
		ab.toAudit.AccountID = res.ID

	default:
		logs.Errorf("unsupported audit create resource: %s, type: %s, rid: %v", ab.toAudit.ResourceType,
			reflect.TypeOf(cur), ab.toAudit.Rid)
		return fmt.Errorf("unsupported audit create resource: %s", ab.toAudit.ResourceType)
	}

	ab.toAudit.Detail = &table.AuditBasicDetail{
		Data:    ab.data,
		Changed: nil,
	}

	return ab.ad.One(ab.kit, txn, ab.toAudit)
}

// PrepareUpdate prepare the resource's previous instance details by
// get the instance's detail from db and save it to ab.data for later use.
// Note:
// 1. call this before resource is updated.
// 2. updatedTo means 'to be updated to data', it should be a *struct.
func (ab *AuditBuilder) PrepareUpdate(updatedTo interface{}) AuditDecorator {
	if ab.hitErr != nil {
		return ab
	}

	ab.toAudit.Action = enumor.Update

	switch res := updatedTo.(type) {
	case *table.Account:
		if err := ab.decorateAccountUpdate(res); err != nil {
			ab.hitErr = err
			return ab
		}

	default:
		logs.Errorf("unsupported audit update resource: %s, type: %s, rid: %v", ab.toAudit.ResourceType,
			reflect.TypeOf(updatedTo), ab.toAudit.Rid)
		ab.hitErr = fmt.Errorf("unsupported audit update resource: %s", ab.toAudit.ResourceType)
		return ab
	}

	ab.toAudit.Detail = &table.AuditBasicDetail{
		Data:    ab.data,
		Changed: ab.changed,
	}

	return ab
}

func (ab *AuditBuilder) decorateAccountUpdate(account *table.Account) error {
	ab.toAudit.ResourceID = account.ID
	ab.toAudit.AccountID = account.ID

	prevApp, err := ab.getAccount(account.ID)
	if err != nil {
		return err
	}

	ab.data = prevApp

	changed, err := parseChangedSpecFields(prevApp, account)
	if err != nil {
		ab.hitErr = err
		return fmt.Errorf("parse account changed spec field failed, err: %v", err)
	}

	ab.changed = changed
	return nil
}

// PrepareDelete prepare the resource's previous instance details by
// get the instance's detail from db and save it to ab.data for later use.
// Note: call this before resource is deleted.
func (ab *AuditBuilder) PrepareDelete(resID uint64) AuditDecorator {
	if ab.hitErr != nil {
		return ab
	}

	ab.toAudit.Action = enumor.Delete

	switch ab.toAudit.ResourceType {
	case enumor.Account:
		account, err := ab.getAccount(resID)
		if err != nil {
			ab.hitErr = err
			return ab
		}

		ab.toAudit.ResourceID = account.ID
		ab.toAudit.AccountID = account.ID
		ab.data = account

	default:
		ab.hitErr = fmt.Errorf("unsupported audit deleted resource: %s", ab.toAudit.ResourceType)
		return ab
	}

	ab.toAudit.Detail = &table.AuditBasicDetail{
		Data:    ab.data,
		Changed: nil,
	}

	return ab
}

// Do save audit log to the db immediately.
func (ab *AuditBuilder) Do(txn *sqlx.Tx) error {

	if ab.hitErr != nil {
		return ab.hitErr
	}

	return ab.ad.One(ab.kit, txn, ab.toAudit)

}

// parseChangedSpecFields parse the changed filed with pre and cur *structs' Spec field.
// both pre and curl should be a *struct, if not, it will 'panic'.
// Note:
// 1. the pre and cur should be the same structs' pointer, and should
//    have a 'Spec' struct field.
// 2. this func only compare 'Spec' field.
// 3. if one of the cur's Spec's filed value is zero, then this filed will be ignored.
// 4. the returned update field's key is this field's 'db' tag.
func parseChangedSpecFields(pre, cur interface{}) (map[string]interface{}, error) {
	preV := reflect.ValueOf(pre)
	if preV.Kind() != reflect.Ptr {
		return nil, errors.New("parse changed spec field, but pre data is not a *struct")
	}

	curV := reflect.ValueOf(cur)
	if curV.Kind() != reflect.Ptr {
		return nil, errors.New("parse changed spec field, but cur data is not a *struct")
	}

	// make sure the pre and data is the same struct.
	if !reflect.TypeOf(pre).AssignableTo(reflect.TypeOf(cur)) {
		return nil, errors.New("parse changed spec field, but pre and cur resource type is not different, " +
			"can not be compared")
	}

	prevSpec := preV.Elem().FieldByName("Spec")
	curSpec := curV.Elem().FieldByName("Spec")
	if prevSpec.IsZero() || curSpec.IsZero() {
		return nil, errors.New("pre or cur data do not has a 'Spec' struct field")
	}

	prevSpecV := prevSpec.Elem()
	curSpecV := curSpec.Elem()
	changedField := make(map[string]interface{})

	// compare spec's detail
	for i := 0; i < prevSpecV.NumField(); i++ {
		preName := prevSpecV.Type().Field(i).Name
		curFieldV := curSpecV.FieldByName(preName)
		if curFieldV.IsZero() {
			// if this filed value is a zero value, then skip it.
			// which means it is not updated.
			continue
		}

		if reflect.DeepEqual(prevSpecV.Field(i).Interface(), curFieldV.Interface()) {
			// this field's value is not changed.
			continue
		}

		dbTag := prevSpecV.Type().Field(i).Tag.Get("db")
		if len(dbTag) == 0 {
			return nil, fmt.Errorf("filed: %s do not have a db tag, can not compare", preName)
		}

		changedField[dbTag] = curFieldV.Interface()
	}

	return changedField, nil
}

func (ab *AuditBuilder) getAccount(id uint64) (*table.Account, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d`,
		table.AccountColumns.NamedExpr(), table.AccountTable, id)

	one := new(table.Account)
	err := ab.ad.orm.Do().Get(ab.kit.Ctx, one, sql)
	if err != nil {
		return nil, fmt.Errorf("get account details failed, err: %v", err)
	}

	return one, nil
}
