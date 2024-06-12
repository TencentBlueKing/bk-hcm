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

// Package types ...
package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/client"
	"hcm/pkg/iam/sys"
	"hcm/pkg/runtime/filter"
)

const (
	// SuccessCode blueking iam success resp code.
	SuccessCode = 0
	// UnauthorizedErrorCode iam token authorized failed error code.
	UnauthorizedErrorCode = 401

	// ListAttrMethod query the list of properties that a resource type can use to configure permissions.
	ListAttrMethod Method = "list_attr"
	// ListAttrValueMethod gets a list of values for an attribute of a resource type.
	ListAttrValueMethod Method = "list_attr_value"
	// ListInstanceMethod query instances based on filter criteria.
	ListInstanceMethod Method = "list_instance"
	// FetchInstanceInfoMethod obtain resource instance details in batch.
	FetchInstanceInfoMethod Method = "fetch_instance_info"
	// ListInstanceByPolicyMethod query resource instances based on policy expressions.
	ListInstanceByPolicyMethod Method = "list_instance_by_policy"
	// SearchInstanceMethod query instances based on filter criteria and search keywords.
	SearchInstanceMethod Method = "search_instance"

	// IDField instance id field name.
	IDField = "id"
	// NameField instance display name.
	NameField = "display_name"
	// ResTopology resource topology level. e.g: "/biz,1/set,1/module,1/"
	ResTopology = "_bk_iam_path_"
)

// Method pull resource method.
type Method string

// PullResourceReq blueking iam pull resource request.
type PullResourceReq struct {
	Type   client.TypeID `json:"type"`
	Method Method        `json:"method"`
	Filter interface{}   `json:"filter,omitempty"`
	Page   Page          `json:"page,omitempty"`
}

// UnmarshalJSON unmarshal json to PullResourceReq.
func (req *PullResourceReq) UnmarshalJSON(raw []byte) error {
	data := struct {
		Type   client.TypeID   `json:"type"`
		Method Method          `json:"method"`
		Filter json.RawMessage `json:"filter,omitempty"`
		Page   Page            `json:"page,omitempty"`
	}{}
	err := json.Unmarshal(raw, &data)
	if err != nil {
		return err
	}
	req.Type = data.Type
	req.Method = data.Method
	req.Page = data.Page
	if data.Filter == nil || len(data.Filter) == 0 {
		return nil
	}
	switch data.Method {
	case ListAttrValueMethod:
		opt := ListAttrValueFilter{}
		err := json.Unmarshal(data.Filter, &opt)
		if err != nil {
			return err
		}
		req.Filter = opt
	case ListInstanceMethod, SearchInstanceMethod:
		opt := ListInstanceFilter{}
		err := json.Unmarshal(data.Filter, &opt)
		if err != nil {
			return err
		}
		req.Filter = opt
	case FetchInstanceInfoMethod:
		opt := FetchInstanceInfoFilter{}
		err := json.Unmarshal(data.Filter, &opt)
		if err != nil {
			return err
		}
		req.Filter = opt
	case ListInstanceByPolicyMethod:
		opt := ListInstanceByPolicyFilter{}
		err := json.Unmarshal(data.Filter, &opt)
		if err != nil {
			return err
		}
		req.Filter = opt
	default:
		return fmt.Errorf("method %s is not supported", data.Method)
	}
	return nil
}

// Page blueking iam pull resource page.
type Page struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

// ListAttrValueFilter list attr value filter.
type ListAttrValueFilter struct {
	Attr    string `json:"attr"`
	Keyword string `json:"keyword,omitempty"`
	// id type is string, int or bool
	IDs []interface{} `json:"ids,omitempty"`
}

// ListInstanceFilter list instance filter.
type ListInstanceFilter struct {
	Parent  *ParentFilter `json:"parent,omitempty"`
	Keyword string        `json:"keyword,omitempty"`
}

// GetFilter get filter from list instance filter.
func (f *ListInstanceFilter) GetFilter() (*filter.Expression, error) {
	expr := &filter.Expression{
		Op:    filter.And,
		Rules: make([]filter.RuleFactory, 0),
	}

	if f.Parent != nil {
		field, err := getResourceIDField(f.Parent.Type)
		if err != nil {
			return nil, err
		}

		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: field,
			Op:    filter.Equal.Factory(),
			Value: f.Parent.ID.InstanceID,
		})
	}

	if len(f.Keyword) != 0 {
		expr.Rules = append(expr.Rules, &filter.AtomRule{
			Field: "name",
			Op:    filter.ContainsInsensitive.Factory(),
			Value: f.Keyword,
		})
	}

	return expr, nil
}

// getResourceIDField get the query instance id field corresponding to the resource type.
func getResourceIDField(resType client.TypeID) (string, error) {
	switch resType {
	case sys.Account:
		return "id", nil
	case sys.CloudSelectionScheme:
		return "id", nil
	case sys.MainAccount:
		return "id", nil
	case sys.RootAccount:
		return "id", nil

	default:
		return "", errf.New(errf.InvalidParameter, "resource type not support")
	}
}

// ParentFilter parent filter.
type ParentFilter struct {
	Type client.TypeID `json:"type"`
	ID   InstanceID    `json:"id"`
}

// ResourceTypeChainFilter resource type chain filter.
type ResourceTypeChainFilter struct {
	SystemID string        `json:"system_id"`
	ID       client.TypeID `json:"id"`
}

// FetchInstanceInfoFilter fetch instance info filter.
type FetchInstanceInfoFilter struct {
	IDs   []InstanceID `json:"ids"`
	Attrs []string     `json:"attrs,omitempty"`
}

// ListInstanceByPolicyFilter list instance by policy filter.
type ListInstanceByPolicyFilter struct {
	// Expression *operator.Policy `json:"expression"`
}

// AttrResource attr resource.
type AttrResource struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ListAttrValueResult list attr value result.
type ListAttrValueResult struct {
	Count   int64               `json:"count"`
	Results []AttrValueResource `json:"results"`
}

// AttrValueResource attr value resource.
type AttrValueResource struct {
	// id type is string, int or bool
	ID          interface{} `json:"id"`
	DisplayName string      `json:"display_name"`
}

// ListInstanceResult list instance result.
type ListInstanceResult struct {
	Count   uint64             `json:"count"`
	Results []InstanceResource `json:"results"`
}

// InstanceResource instance resource.
type InstanceResource struct {
	ID          InstanceID `json:"id"`
	DisplayName string     `json:"display_name"`
}

// InstanceID is iam resource id.
type InstanceID struct {
	// InstanceID is hcm resource instance id.
	InstanceID string
}

// UnmarshalJSON unmarshal json raw data to instance id.
func (i *InstanceID) UnmarshalJSON(raw []byte) error {
	id := strings.Trim(strings.TrimSpace(string(raw)), `\"`)
	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "instance id is empty")
	}

	i.InstanceID = id

	return nil
}

// MarshalJSON marshal instance id to json string.
func (i InstanceID) MarshalJSON() ([]byte, error) {
	if len(i.InstanceID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "instance id is required")
	}

	return json.Marshal(i.InstanceID)
}
