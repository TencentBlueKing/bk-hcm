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

package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"hcm/pkg/api/core"
	coreplugin "hcm/pkg/api/core/plugin"
	dataplugin "hcm/pkg/api/data-service/plugin"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/cryptography"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableplugin "hcm/pkg/dal/table/plugin"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

func (svc *service) initIacPluginInstanceRoutes(h *rest.Handler) {
	h.Add("BatchCreateIacPluginInstance", http.MethodPost,
		"/iac_plugin_instance/batch/create", svc.BatchCreateIacPluginInstance)
	h.Add("BatchUpdateIacPluginInstance", http.MethodPatch,
		"/iac_plugin_instance/batch/update", svc.BatchUpdateIacPluginInstance)
	h.Add("BatchDeleteIacPluginInstance", http.MethodDelete,
		"/iac_plugin_instance/batch", svc.BatchDeleteIacPluginInstance)
	h.Add("ListIacPluginInstance", http.MethodPost, "/iac_plugin_instance/list", svc.ListIacPluginInstance)
}

// BatchCreateIacPluginInstance 批量创建 iac_plugin_instance 记录，入库前对 secret_attrs 标量叶子值加密。
func (svc *service) BatchCreateIacPluginInstance(cts *rest.Contexts) (any, error) {
	req := new(dataplugin.IacPluginInstanceBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch create iac plugin instance decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch create iac plugin instance validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	models := make([]tableplugin.IacPluginInstanceTable, 0, len(req.Items))
	for _, item := range req.Items {
		attrsField, err := marshalJSONMap(item.Attrs)
		if err != nil {
			logs.Errorf("marshal iac plugin instance attrs failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		secretAttrsField, err := encryptSecretAttrs(item.SecretAttrs, svc.cipher)
		if err != nil {
			logs.Errorf("encrypt iac plugin instance secret_attrs failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.Aborted, err)
		}

		extField, err := marshalJSONMap(item.Extension)
		if err != nil {
			logs.Errorf("marshal iac plugin instance extension failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		models = append(models, tableplugin.IacPluginInstanceTable{
			PluginName:  item.PluginName,
			Name:        item.Name,
			Attrs:       attrsField,
			SecretAttrs: secretAttrsField,
			Extension:   extField,
			Memo:        item.Memo,
			Creator:     cts.Kit.User,
			Reviser:     cts.Kit.User,
		})
	}

	createdIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		ids, err := svc.dao.IacPluginInstance().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			logs.Errorf("batch create iac plugin instance in db failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create iac plugin instance transaction failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	ids, ok := createdIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create iac plugin instance return ids type is not []string: %v",
			reflect.TypeOf(createdIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

// BatchUpdateIacPluginInstance 批量更新 iac_plugin_instance 记录，若传入 secret_attrs 则对标量叶子值加密后全量覆盖。
func (svc *service) BatchUpdateIacPluginInstance(cts *rest.Contexts) (any, error) {
	req := new(dataplugin.IacPluginInstanceBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch update iac plugin instance decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch update iac plugin instance validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		for _, item := range req.Items {
			model, convErr := svc.buildUpdateInstanceModel(cts, item)
			if convErr != nil {
				return nil, convErr
			}

			if err := svc.dao.IacPluginInstance().UpdateWithTx(cts.Kit, txn,
				tools.EqualExpression("id", item.ID), model); err != nil {
				logs.Errorf("update iac plugin instance in db failed, id: %s, err: %v, rid: %s",
					item.ID, err, cts.Kit.Rid)
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update iac plugin instance transaction failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

func (svc *service) buildUpdateInstanceModel(cts *rest.Contexts,
	item dataplugin.IacPluginInstanceUpdateField) (*tableplugin.IacPluginInstanceTable, error) {

	model := &tableplugin.IacPluginInstanceTable{
		Name:    item.Name,
		Memo:    item.Memo,
		Reviser: cts.Kit.User,
	}

	if item.Attrs != nil {
		attrsField, err := marshalJSONMap(item.Attrs)
		if err != nil {
			logs.Errorf("marshal update iac plugin instance attrs failed, id: %s, err: %v, rid: %s",
				item.ID, err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		model.Attrs = attrsField
	}

	if item.SecretAttrs != nil {
		secretAttrsField, err := encryptSecretAttrs(item.SecretAttrs, svc.cipher)
		if err != nil {
			logs.Errorf("encrypt update iac plugin instance secret_attrs failed, id: %s, err: %v, rid: %s",
				item.ID, err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.Aborted, err)
		}
		model.SecretAttrs = secretAttrsField
	}

	if item.Extension != nil {
		extField, err := marshalJSONMap(item.Extension)
		if err != nil {
			logs.Errorf("marshal update iac plugin instance extension failed, id: %s, err: %v, rid: %s",
				item.ID, err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
		model.Extension = extField
	}

	return model, nil
}

// BatchDeleteIacPluginInstance 批量删除 iac_plugin_instance 记录。
func (svc *service) BatchDeleteIacPluginInstance(cts *rest.Contexts) (any, error) {
	req := new(dataplugin.IacPluginInstanceBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("batch delete iac plugin instance decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("batch delete iac plugin instance validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		if err := svc.dao.IacPluginInstance().DeleteWithTx(cts.Kit, txn, req.Filter); err != nil {
			logs.Errorf("delete iac plugin instance in db failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch delete iac plugin instance transaction failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return nil, nil
}

// ListIacPluginInstance 列表查询 iac_plugin_instance 记录并解密 secret_attrs 标量叶子值。
func (svc *service) ListIacPluginInstance(cts *rest.Contexts) (any, error) {
	req := new(dataplugin.IacPluginInstanceListReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("list iac plugin instance decode request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("list iac plugin instance validate request failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	res, err := svc.dao.IacPluginInstance().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list iac plugin instance failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	if req.Page.Count {
		return &dataplugin.IacPluginInstanceListResult{Count: res.Count}, nil
	}

	details := make([]coreplugin.IacPluginInstance, 0, len(res.Details))
	for i := range res.Details {
		inst, err := convertTableToIacPluginInstance(&res.Details[i], svc.cipher)
		if err != nil {
			logs.Errorf("convert iac plugin instance in list failed, id: %s, err: %v, rid: %s",
				res.Details[i].ID, err, cts.Kit.Rid)
			return nil, errf.NewFromErr(errf.Aborted, err)
		}
		details = append(details, *inst)
	}

	return &dataplugin.IacPluginInstanceListResult{Details: details}, nil
}

func convertTableToIacPluginInstance(row *tableplugin.IacPluginInstanceTable,
	cipher cryptography.Crypto) (*coreplugin.IacPluginInstance, error) {

	attrs, err := unmarshalJSONMap(row.Attrs)
	if err != nil {
		return nil, fmt.Errorf("unmarshal attrs failed: %w", err)
	}

	secretAttrs, err := decryptSecretAttrs(row.SecretAttrs, cipher)
	if err != nil {
		return nil, fmt.Errorf("decrypt secret_attrs failed: %w", err)
	}

	ext, err := unmarshalJSONMap(row.Extension)
	if err != nil {
		return nil, fmt.Errorf("unmarshal extension failed: %w", err)
	}

	return &coreplugin.IacPluginInstance{
		ID:          row.ID,
		PluginName:  row.PluginName,
		Name:        row.Name,
		Attrs:       attrs,
		SecretAttrs: secretAttrs,
		Extension:   ext,
		Memo:        row.Memo,
		TenantID:    row.TenantID,
		Revision: &core.Revision{
			Creator:   row.Creator,
			Reviser:   row.Reviser,
			CreatedAt: row.CreatedAt.String(),
			UpdatedAt: row.UpdatedAt.String(),
		},
	}, nil
}

func marshalJSONMap(m map[string]any) (tabletypes.JsonField, error) {
	if len(m) == 0 {
		return tabletypes.JsonField("{}"), nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return tabletypes.JsonField(raw), nil
}

func unmarshalJSONMap(raw tabletypes.JsonField) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "{}" || string(raw) == "null" {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	if m == nil {
		return map[string]any{}, nil
	}
	return m, nil
}

// encryptSecretAttrs 对 secret_attrs 树中的所有标量叶子值执行 EncryptToBase64(string(json.Marshal(leafVal)))，
// 保持外层嵌套 map key、数组结构及 {} 占位符明文不变。
func encryptSecretAttrs(secretAttrs map[string]any, cipher cryptography.Crypto) (tabletypes.JsonField, error) {
	if len(secretAttrs) == 0 {
		return tabletypes.JsonField("{}"), nil
	}
	if cipher == nil {
		return "", fmt.Errorf("cipher is nil")
	}

	encryptedMap, err := encryptNodeMap(secretAttrs, cipher)
	if err != nil {
		return "", err
	}

	raw, err := json.Marshal(encryptedMap)
	if err != nil {
		return "", err
	}
	return tabletypes.JsonField(raw), nil
}

func encryptNodeMap(m map[string]any, cipher cryptography.Crypto) (map[string]any, error) {
	out := make(map[string]any, len(m))
	for k, v := range m {
		encVal, err := encryptNodeValue(v, cipher)
		if err != nil {
			return nil, fmt.Errorf("key %s: %w", k, err)
		}
		out[k] = encVal
	}
	return out, nil
}

func encryptNodeValue(v any, cipher cryptography.Crypto) (any, error) {
	switch typed := v.(type) {
	case map[string]any:
		return encryptNodeMap(typed, cipher)
	case []any:
		out := make([]any, len(typed))
		for i, elem := range typed {
			encElem, err := encryptNodeValue(elem, cipher)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			out[i] = encElem
		}
		return out, nil
	default:
		leafJSON, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("marshal leaf value failed: %w", err)
		}
		return cipher.EncryptToBase64(string(leafJSON)), nil
	}
}

// decryptSecretAttrs 将 secret_attrs 树中加密的标量叶子值通过 DecryptFromBase64 + json.Unmarshal 还原为原始标量类型。
func decryptSecretAttrs(raw tabletypes.JsonField, cipher cryptography.Crypto) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "{}" || string(raw) == "null" {
		return map[string]any{}, nil
	}
	if cipher == nil {
		return nil, fmt.Errorf("cipher is nil")
	}

	var encryptedMap map[string]any
	if err := json.Unmarshal([]byte(raw), &encryptedMap); err != nil {
		return nil, err
	}
	if len(encryptedMap) == 0 {
		return map[string]any{}, nil
	}

	return decryptNodeMap(encryptedMap, cipher)
}

func decryptNodeMap(m map[string]any, cipher cryptography.Crypto) (map[string]any, error) {
	out := make(map[string]any, len(m))
	for k, v := range m {
		decVal, err := decryptNodeValue(v, cipher)
		if err != nil {
			return nil, fmt.Errorf("key %s: %w", k, err)
		}
		out[k] = decVal
	}
	return out, nil
}

func decryptNodeValue(v any, cipher cryptography.Crypto) (any, error) {
	switch typed := v.(type) {
	case map[string]any:
		return decryptNodeMap(typed, cipher)
	case []any:
		out := make([]any, len(typed))
		for i, elem := range typed {
			decElem, err := decryptNodeValue(elem, cipher)
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			out[i] = decElem
		}
		return out, nil
	case string:
		plainJSON, err := cipher.DecryptFromBase64(typed)
		if err != nil {
			return nil, fmt.Errorf("decrypt leaf ciphertext failed: %w", err)
		}
		var val any
		if err := json.Unmarshal([]byte(plainJSON), &val); err != nil {
			return nil, fmt.Errorf("unmarshal decrypted leaf json failed: %w", err)
		}
		return val, nil
	default:
		return nil, fmt.Errorf("unexpected non-string encrypted leaf value of type %T", v)
	}
}
