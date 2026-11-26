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
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"hcm/pkg/criteria/enumor"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// segmentIndex 表示地址段后面跟着的一个方括号下标（如 [0] 或 ["gz"]）。
type segmentIndex struct {
	IsInt  bool
	IntVal int
	StrVal string
}

// String 把单个下标还原成 "[0]"、"[\"gz\"]" 或 "[*]"。
func (idx segmentIndex) String() string {
	if idx.IsInt {
		return fmt.Sprintf("[%d]", idx.IntVal)
	}
	if idx.StrVal == "*" {
		return "[*]"
	}
	return fmt.Sprintf("[%s]", strconv.Quote(idx.StrVal))
}

// addrSegment 表示点号分隔地址中的一段，包含名字和可选的方括号下标列表。
//
// 举例：
//   - 普通名字 "module" -> addrSegment{Name: "module"}
//   - count 整数下标 "sub[0]" -> addrSegment{Name: "sub", Indices: []segmentIndex{{IsInt: true, IntVal: 0}}}
//   - for_each 字符串下标 vpc["gz"] -> addrSegment{Name: "vpc", Indices: []segmentIndex{{StrVal: "gz"}}}
//   - 连续多下标 az_map["gz"][0] ->
//     addrSegment{Name: "az_map", Indices: []segmentIndex{{StrVal: "gz"}, {IsInt: true, IntVal: 0}}}
type addrSegment struct {
	Name    string
	Indices []segmentIndex
}

// HasIndex 判断当前段是否带有方括号下标。
func (s addrSegment) HasIndex() bool {
	return len(s.Indices) > 0
}

// String 把 addrSegment 还原成标准地址段字符串，比如 "network"、"sub[0]"、"vpc[\"gz\"]"、"az_map[\"gz\"][0]"。
func (s addrSegment) String() string {
	if len(s.Indices) == 0 {
		return s.Name
	}
	var sb strings.Builder
	sb.WriteString(s.Name)
	for _, idx := range s.Indices {
		sb.WriteString(idx.String())
	}
	return sb.String()
}

// parsedAddress 保存从完整引用地址里拆出来的模块前缀、节点本身和属性路径。
//
// 举例 1（子模块和资源名字都叫 module 的情况）：
//
//	输入：module.net.module.module.tencentcloud_vpc.module.id
//	ModuleHopSegments: [module, net, module, module]
//	NodeSegments:      [tencentcloud_vpc, module]
//	AttrSegments:      [id]
//	BlockID:           "module.net.module.module.tencentcloud_vpc.module"
//
// 举例 2（子模块里的 data 节点）：
//
//	输入：module.net[0].data.tencentcloud_images.ubuntu.images[0].image_id
//	ModuleHopSegments: [module, net[0]]
//	NodeSegments:      [data, tencentcloud_images, ubuntu]
//	AttrSegments:      [images[0], image_id]
//	BlockID:           "module.net[0].data.tencentcloud_images.ubuntu"
type parsedAddress struct {
	// ModuleHopSegments 成对的 [module, <模块名>] 前缀段
	ModuleHopSegments []addrSegment
	// NodeSegments 节点本身（resource 占 2 段，data 和 import 占 3 段）
	NodeSegments []addrSegment
	// AttrSegments 节点后面的属性访问路径，比如 [id] 或 [tags, "env"]
	AttrSegments []addrSegment
	// BlockID 模块前缀加节点本身拼出来的完整节点 ID
	BlockID  string
	Kind     enumor.BlockKind
	TypeName string
}

// parseAddressString 利用 HCL 原生语法解析器把点分地址拆成 addrSegment 列表。
// 比如输入 "module.vpc[\"gz.01\"].tencentcloud_subnet.sub[0].id"，
// 拆成 [module, vpc["gz.01"], tencentcloud_subnet, sub[0], id]。
func parseAddressString(addr string) ([]addrSegment, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, fmt.Errorf("address is empty")
	}
	trav, diags := hclsyntax.ParseTraversalAbs([]byte(addr), "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("invalid address %q: %s", addr, diags.Error())
	}
	return traversalToSegments(trav)
}

// traversalToSegments 把 HCL 原生的 hcl.Traversal 转成 []addrSegment。
// 比如 module.vpc["gz"].subnet_id 转成 [module, vpc["gz"], subnet_id]。
func traversalToSegments(trav hcl.Traversal) ([]addrSegment, error) {
	var segs []addrSegment
	for _, step := range trav {
		switch s := step.(type) {
		// 开头的根名称，如 module.vpc["gz"].subnet_id 里的 module
		case hcl.TraverseRoot:
			segs = append(segs, addrSegment{Name: s.Name})
		// 点号后面的名字，如 .vpc、.subnet_id
		case hcl.TraverseAttr:
			segs = append(segs, addrSegment{Name: s.Name})
		// 方括号下标，如 ["gz"] 或 [0]，直接挂到前一段名字上
		case hcl.TraverseIndex:
			lastIdx := len(segs) - 1
			if lastIdx < 0 {
				return nil, fmt.Errorf("invalid index traversal")
			}
			indexedSeg, err := attachCtyIndexToSegment(segs[lastIdx], s.Key)
			if err != nil {
				return nil, err
			}
			segs[lastIdx] = indexedSeg
		default:
			return nil, fmt.Errorf("unsupported traversal step %T", step)
		}
	}
	return segs, nil
}

// attachCtyIndexToSegment 把求值出来的整数或字符串下标追加到 addrSegment 上并返回更新后的段，比如把 sub 变成 sub[0] 或 vpc["gz"]。
func attachCtyIndexToSegment(seg addrSegment, key cty.Value) (addrSegment, error) {
	if !key.IsKnown() || key.IsNull() {
		return addrSegment{}, fmt.Errorf("index key on %q cannot be null or unknown", seg.Name)
	}
	if key.Type() == cty.String {
		seg.Indices = append(seg.Indices, segmentIndex{StrVal: key.AsString()})
		return seg, nil
	}
	if key.Type() == cty.Number {
		bf := key.AsBigFloat()
		if i64, acc := bf.Int64(); acc == big.Exact {
			seg.Indices = append(seg.Indices, segmentIndex{IsInt: true, IntVal: int(i64)})
			return seg, nil
		}
	}
	return addrSegment{}, fmt.Errorf("index key on %q must be integer or string", seg.Name)
}

// consumePairwiseAddress 从地址段里拆出模块前缀、节点名和属性路径。
// 因为子模块前缀固定是 "module.<模块名>" 两段一组，所以从左往右每次跳过两段，
// 遇到第一个不是 "module" 的段就是节点本身（data/import 占 3 段，普通 resource 占 2 段），剩下的就是属性名。
// 这样就算模块名或资源名叫 module，也不会解析错。
//
// 举例：
//
//	输入: module.net.module.module.tencentcloud_vpc.module.id
//	模块前缀: module.net.module.module（两层子模块，哪怕第二层模块名叫 module 也不会认错）
//	节点部分: tencentcloud_vpc.module
//	属性路径: id
func consumePairwiseAddress(segs []addrSegment) (*parsedAddress, error) {
	i := 0
	var modHops []addrSegment
	for i+1 < len(segs) && segs[i].Name == "module" && !segs[i].HasIndex() {
		modHops = append(modHops, segs[i], segs[i+1])
		i += 2
	}
	if i >= len(segs) {
		return nil, fmt.Errorf("address %q contains only module prefixes without node identifier", formatSegments(segs))
	}

	head := segs[i]
	if (head.Name == string(enumor.BlockKindData) || head.Name == string(enumor.BlockKindImport)) && !head.HasIndex() {
		if i+2 >= len(segs) {
			return nil, fmt.Errorf("incomplete %s node address in %q", head.Name, formatSegments(segs))
		}
		typeSeg := segs[i+1]
		if !strings.Contains(typeSeg.Name, "_") || typeSeg.HasIndex() {
			return nil, fmt.Errorf("invalid type_name %q in %q: must contain underscore", typeSeg.Name,
				formatSegments(segs))
		}
		nodeSegs := segs[i : i+3]
		attrSegs := segs[i+3:]
		blockID := formatSegments(append(append([]addrSegment{}, modHops...), nodeSegs...))
		return &parsedAddress{
			ModuleHopSegments: modHops,
			NodeSegments:      nodeSegs,
			AttrSegments:      attrSegs,
			BlockID:           blockID,
			Kind:              enumor.BlockKind(head.Name),
			TypeName:          typeSeg.Name,
		}, nil
	}

	if !strings.Contains(head.Name, "_") || head.HasIndex() {
		return nil, fmt.Errorf("invalid resource type_name %q in %q: must contain underscore", head.Name,
			formatSegments(segs))
	}
	if i+1 >= len(segs) {
		return nil, fmt.Errorf("incomplete resource node address in %q", formatSegments(segs))
	}
	nodeSegs := segs[i : i+2]
	attrSegs := segs[i+2:]
	blockID := formatSegments(append(append([]addrSegment{}, modHops...), nodeSegs...))
	return &parsedAddress{
		ModuleHopSegments: modHops,
		NodeSegments:      nodeSegs,
		AttrSegments:      attrSegs,
		BlockID:           blockID,
		Kind:              enumor.BlockKindResource,
		TypeName:          head.Name,
	}, nil
}

// formatSegments 把地址段切片用 "." 拼回完整地址字符串。
func formatSegments(segs []addrSegment) string {
	if len(segs) == 0 {
		return ""
	}
	parts := make([]string, len(segs))
	for i, s := range segs {
		parts[i] = s.String()
	}
	return strings.Join(parts, ".")
}

// formatAttrSuffix 把属性段拼成带前导 "." 的后缀字符串，比如 [tags, 0, value] 拼成 ".tags.0.value"，空切片返回 ""。
func formatAttrSuffix(segs []addrSegment) string {
	if len(segs) == 0 {
		return ""
	}
	return "." + formatSegments(segs)
}

// joinModulePrefix 把模块前缀和模块内的节点名拼成完整 BlockID，
// 比如 "module.net" 和 "tencentcloud_vpc.main" 拼成 "module.net.tencentcloud_vpc.main"。
func joinModulePrefix(modulePrefix, localID string) string {
	if modulePrefix == "" {
		return localID
	}
	return modulePrefix + "." + localID
}

// extractPluginFromTypeName 从 type_name 里截取第一个下划线前面的插件名，比如 "tencentcloud_subnet" 提取出 "tencentcloud"。
func extractPluginFromTypeName(typeName string) (string, error) {
	idx := strings.IndexByte(typeName, '_')
	if idx <= 0 || idx == len(typeName)-1 {
		return "", fmt.Errorf("invalid type_name %q: must be in <plugin>_<type> format containing underscore", typeName)
	}
	return typeName[:idx], nil
}
