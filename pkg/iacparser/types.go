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

package iacparser

import (
	"errors"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
)

// ParseReq 阶段一解析请求参数。
//
// 举例：
//
//	req := &ParseReq{
//	    Modules: map[string]map[string][]byte{
//	        ".": {"main.tf": []byte("module \"net\" { source = \"./modules/vpc\" }")},
//	        "modules/vpc": {
//	            "vpc.tf": []byte("resource \"tencentcloud_vpc\" \"main\" { cidr_block = \"127.0.0.1/16\" }"),
//	        },
//	    },
//	    Vars: map[string]any{"bk_biz_id": 213, "region": "ap-guangzhou"},
//	}
type ParseReq struct {
	// Modules 按模块路径存放对应的 .tf 文件内容，"." 表示根模块，子模块如 "modules/network"
	Modules map[string]map[string][]byte `json:"modules"`
	// Vars 传给根模块的外部变量值，对应根模块里的 variable 定义
	Vars map[string]any `json:"vars,omitempty"`
}

// Validate 检查请求是否合法，根模块 "." 必须存在且至少有一个 .tf 文件。
func (req *ParseReq) Validate() error {
	if req == nil {
		return errors.New("parse request is nil")
	}
	if len(req.Modules) == 0 {
		return errors.New("parse request modules is empty")
	}
	rootFiles, ok := req.Modules[constant.RootModuleKey]
	if !ok || len(rootFiles) == 0 {
		return errors.New("missing or empty root module \".\"")
	}
	return nil
}

// ParseResult 阶段一解析结果。
type ParseResult struct {
	// Blocks 展开所有子模块和多实例、并按依赖关系拓扑排序后的节点列表
	Blocks []*Block `json:"blocks"`
}

// Block 表示解析出的单个执行节点（只包含 resource、data、import 三种）。
//
// 举例：
//
//	Block{
//	    BlockID:          "module.app.tencentcloud_instance.web[0]",
//	    Kind:             enumor.BlockKindResource,
//	    Plugin:           "tencentcloud",
//	    PluginInstanceID: "tcloud-prod-account-01",
//	    TypeName:         "tencentcloud_instance",
//	    DependsOn:        []string{"module.network.tencentcloud_subnet.app"},
//	    Config: map[string]any{
//	        "instance_name": "web-0",
//	        "tags": []any{
//	            map[string]any{"key": "subnet", "value": nil}, // 依赖其他资源的字段先用 nil 占位
//	        },
//	    },
//	    PendingExprs: map[string]string{
//	        "subnet_id":    "module.network.tencentcloud_subnet.app.id",
//	        "tags.0.value": "module.network.tencentcloud_subnet.app.id",
//	    },
//	}
type Block struct {
	// BlockID 节点唯一标识，格式与 Terraform 地址一致，比如 "tencentcloud_vpc.main"、"module.net.tencentcloud_subnet.sub[0]"
	BlockID string `json:"block_id"`
	// Kind 节点类型
	Kind enumor.BlockKind `json:"kind"`
	// Plugin 插件名，默认取 TypeName 第一个下划线前缀，比如 "tencentcloud"
	Plugin string `json:"plugin"`
	// PluginInstanceID 关联的插件实例 ID，来自 provider 块里的 plugin_instance_id
	PluginInstanceID string `json:"plugin_instance_id"`
	// TypeName 资源或数据源类型名，比如 "tencentcloud_subnet"
	TypeName string `json:"type_name"`
	// DependsOn 依赖的上游节点 BlockID 列表（已去重并按字典序排列）
	DependsOn []string `json:"depends_on"`
	// Config 阶段一算出来的静态配置参数；如果嵌套结构里有动态字段，对应位置会先填 nil 占位
	Config map[string]any `json:"config"`
	// PendingExprs 留到阶段二再算的动态表达式，key 是顶层字段名或点号路径（如 "subnet_id"、"tags.0.value"），value 是改写后的表达式
	PendingExprs map[string]string `json:"pending_exprs,omitempty"`
	// Lifecycle 从块内提取出来的 lifecycle 配置，不会放进业务 Config
	Lifecycle *Lifecycle `json:"lifecycle,omitempty"`
}

// Lifecycle 资源生命周期控制参数，对应 HCL 里的 lifecycle {} 子块。
type Lifecycle struct {
	// CreateBeforeDestroy 替换资源时是否先建新资源再删旧资源
	CreateBeforeDestroy bool `json:"create_before_destroy,omitempty"`
	// PreventDestroy 是否禁止删除该资源
	PreventDestroy bool `json:"prevent_destroy,omitempty"`
	// IgnoreChanges 更新时忽略对比的字段路径列表，比如 ["tags.created_at", "description"] 或 ["all"]
	IgnoreChanges []string `json:"ignore_changes,omitempty"`
	// ReplaceTriggeredBy 触发当前资源重建的上游资源或属性引用列表，同时会自动加进 Block.DependsOn
	ReplaceTriggeredBy []string `json:"replace_triggered_by,omitempty"`
	// Recompute 更新时需要重置为 null 让插件重新计算默认值的字段名列表
	Recompute []string `json:"recompute,omitempty"`
}

// EvalReq 阶段二单节点求值请求参数。
//
// 举例：
//
//	evalReq := &EvalReq{
//	    Config: map[string]any{
//	        "name": "app-subnet",
//	        "tags": []any{map[string]any{"key": "vpc", "value": nil}},
//	    },
//	    PendingExprs: map[string]string{
//	        "vpc_id":       "tencentcloud_vpc.main.id",
//	        "tags.0.value": "tencentcloud_vpc.main.id",
//	    },
//	    DependencyStates: map[string]map[string]any{
//	        "tencentcloud_vpc.main": {"id": "vpc-12345"},
//	    },
//	}
type EvalReq struct {
	// Config 阶段一生成的静态配置，EvalBlock 内部会先做深拷贝，不会改入参
	Config map[string]any `json:"config"`
	// PendingExprs 待求值的表达式映射，key 为顶层字段名或点号路径
	PendingExprs map[string]string `json:"pending_exprs"`
	// DependencyStates 上游依赖节点的运行结果，key 是上游节点的 BlockID，value 是该节点的属性字典
	DependencyStates map[string]map[string]any `json:"dependency_states"`
}

// Validate 检查阶段二请求参数是否为空。
func (req *EvalReq) Validate() error {
	if req == nil {
		return errors.New("eval request is nil")
	}
	return nil
}
