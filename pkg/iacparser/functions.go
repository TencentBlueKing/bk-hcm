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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// safeFunctions 是允许在 HCL 表达式里调用的内置函数表。
// 这里只放数学、字符串、集合、类型转换、编解码等纯内存计算函数，不涉及任何文件读写或网络操作，多协程只读无需加锁。
var safeFunctions = buildSafeFunctions()

// forbiddenFunctions 是明确禁用的宿主机文件和路径类函数黑名单。
// 如果用户在 HCL 里写了 file("config.txt") 或 templatefile(...)，解析时会直接报错拦截。
var forbiddenFunctions = map[string]bool{
	"file":         true,
	"fileexists":   true,
	"fileset":      true,
	"filebase64":   true,
	"templatefile": true,
	"pathexpand":   true,
	"abspath":      true,
}

// buildSafeFunctions 初始化内置函数表。
//
// 常用函数举例：
//   - 数学计算：max(10, 20) -> 20，min(1, 2) -> 1
//   - 字符串处理：join("-", ["web", "prod"]) -> "web-prod"，format("app-%02d", 1) -> "app-01"
//   - 集合处理：merge({a = "1"}, {b = "2"}) -> {"a" = "1", "b" = "2"}，flatten([["a"], ["b"]]) -> ["a", "b"]
//   - 编码转换：jsonencode({k = "v"}) -> "{\"k\":\"v\"}"
func buildSafeFunctions() map[string]function.Function {
	funcs := map[string]function.Function{
		// 数学计算
		"abs":      stdlib.AbsoluteFunc,
		"ceil":     stdlib.CeilFunc,
		"floor":    stdlib.FloorFunc,
		"log":      stdlib.LogFunc,
		"max":      stdlib.MaxFunc,
		"min":      stdlib.MinFunc,
		"parseint": stdlib.ParseIntFunc,
		"pow":      stdlib.PowFunc,
		"signum":   stdlib.SignumFunc,
		// 字符串处理
		"chomp":      stdlib.ChompFunc,
		"format":     stdlib.FormatFunc,
		"formatlist": stdlib.FormatListFunc,
		"indent":     stdlib.IndentFunc,
		"join":       stdlib.JoinFunc,
		"lower":      stdlib.LowerFunc,
		"regex":      stdlib.RegexFunc,
		"regexall":   stdlib.RegexAllFunc,
		"replace":    stdlib.ReplaceFunc,
		"split":      stdlib.SplitFunc,
		"strrev":     stdlib.ReverseFunc,
		"substr":     stdlib.SubstrFunc,
		"title":      stdlib.TitleFunc,
		"trim":       stdlib.TrimFunc,
		"trimprefix": stdlib.TrimPrefixFunc,
		"trimsuffix": stdlib.TrimSuffixFunc,
		"trimspace":  stdlib.TrimSpaceFunc,
		"upper":      stdlib.UpperFunc,
		// 集合操作
		"chunklist":       stdlib.ChunklistFunc,
		"coalesce":        stdlib.CoalesceFunc,
		"coalescelist":    stdlib.CoalesceListFunc,
		"compact":         stdlib.CompactFunc,
		"concat":          stdlib.ConcatFunc,
		"contains":        stdlib.ContainsFunc,
		"distinct":        stdlib.DistinctFunc,
		"element":         stdlib.ElementFunc,
		"flatten":         stdlib.FlattenFunc,
		"keys":            stdlib.KeysFunc,
		"length":          makeLengthFunc(),
		"merge":           stdlib.MergeFunc,
		"range":           stdlib.RangeFunc,
		"reverse":         stdlib.ReverseListFunc,
		"setintersection": stdlib.SetIntersectionFunc,
		"setproduct":      stdlib.SetProductFunc,
		"setsubtract":     stdlib.SetSubtractFunc,
		"setunion":        stdlib.SetUnionFunc,
		"slice":           stdlib.SliceFunc,
		"sort":            stdlib.SortFunc,
		"values":          stdlib.ValuesFunc,
		"zipmap":          stdlib.ZipmapFunc,
		// 编解码与时间格式化
		"csvdecode":  stdlib.CSVDecodeFunc,
		"formatdate": stdlib.FormatDateFunc,
		"jsondecode": stdlib.JSONDecodeFunc,
		"jsonencode": stdlib.JSONEncodeFunc,
	}
	funcs = registerTerraformExtFunctions(funcs)
	return funcs
}

// registerTerraformExtFunctions 补充注册 Terraform 常用、但不在 go-cty 标准库直接导出变量里的内置函数。
//
// 举例：
//   - 类型转换：tostring(213) -> "213"，tonumber("80") -> 80，toset(["a", "a"]) -> ["a"]
//   - 字典查找：lookup({env = "prod"}, "region", "ap-guangzhou") -> "ap-guangzhou"
//   - 哈希与编码：base64encode("hello") -> "aGVsbG8="，md5("test") -> "098f6bcd4621d373cade4e832627b4f6"
func registerTerraformExtFunctions(funcs map[string]function.Function) map[string]function.Function {
	funcs["tostring"] = stdlib.MakeToFunc(cty.String)
	funcs["tonumber"] = stdlib.MakeToFunc(cty.Number)
	funcs["tobool"] = stdlib.MakeToFunc(cty.Bool)
	funcs["tolist"] = stdlib.MakeToFunc(cty.List(cty.DynamicPseudoType))
	funcs["toset"] = stdlib.MakeToFunc(cty.Set(cty.DynamicPseudoType))
	funcs["tomap"] = stdlib.MakeToFunc(cty.Map(cty.DynamicPseudoType))
	funcs["lookup"] = stdlib.LookupFunc
	funcs["try"] = makeTryFunc()
	funcs["can"] = makeCanFunc()
	funcs["index"] = makeIndexFunc()
	funcs["base64encode"] = makeStringTransformFunc(func(s string) (string, error) {
		return base64.StdEncoding.EncodeToString([]byte(s)), nil
	})
	funcs["base64decode"] = makeStringTransformFunc(func(s string) (string, error) {
		b, err := base64.StdEncoding.DecodeString(s)
		return string(b), err
	})
	funcs["md5"] = makeStringTransformFunc(func(s string) (string, error) {
		sum := md5.Sum([]byte(s))
		return hex.EncodeToString(sum[:]), nil
	})
	funcs["sha1"] = makeStringTransformFunc(func(s string) (string, error) {
		sum := sha1.Sum([]byte(s))
		return hex.EncodeToString(sum[:]), nil
	})
	funcs["sha256"] = makeStringTransformFunc(func(s string) (string, error) {
		sum := sha256.Sum256([]byte(s))
		return hex.EncodeToString(sum[:]), nil
	})
	funcs["urlencode"] = makeStringTransformFunc(func(s string) (string, error) {
		return url.QueryEscape(s), nil
	})
	return funcs
}

// makeLengthFunc 实现 Terraform 的 length(value) 函数，同时支持求字符串长度和集合元素个数。
// 比如 length("hello") 返回 5，length(["a", "b"]) 返回 2。
func makeLengthFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{{Name: "value", Type: cty.DynamicPseudoType}},
		Type:   function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			if args[0].Type() == cty.String {
				return stdlib.StrlenFunc.Call(args)
			}
			return stdlib.LengthFunc.Call(args)
		},
	})
}

// makeTryFunc 实现 try(expressions...) 函数，返回第一个求值成功的参数。
func makeTryFunc() function.Function {
	return function.New(&function.Spec{
		VarParam: &function.Parameter{Name: "expressions", Type: cty.DynamicPseudoType, AllowNull: true},
		Type: func(args []cty.Value) (cty.Type, error) {
			if len(args) == 0 {
				return cty.NilType, fmt.Errorf("try requires at least one argument")
			}
			return args[0].Type(), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return args[0], nil
		},
	})
}

// makeCanFunc 实现 can(expression) 函数，判断参数表达式能否正常求值。
func makeCanFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{{Name: "expression", Type: cty.DynamicPseudoType, AllowNull: true}},
		Type:   function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return cty.BoolVal(args[0].IsKnown()), nil
		},
	})
}

// makeIndexFunc 实现 index(list, value) 函数，返回元素在列表里第一次出现的下标，比如 index(["gz", "sh"], "sh") 返回 1。
func makeIndexFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{Name: "list", Type: cty.DynamicPseudoType},
			{Name: "value", Type: cty.DynamicPseudoType},
		},
		Type: function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			listVal, target := args[0], args[1]
			if !listVal.Type().IsTupleType() && !listVal.Type().IsListType() {
				return cty.NilVal, fmt.Errorf("index requires a list or tuple")
			}
			idx := int64(0)
			it := listVal.ElementIterator()
			for it.Next() {
				_, elem := it.Element()
				eq := elem.Equals(target)
				if eq.IsKnown() && eq.True() {
					return cty.NumberIntVal(idx), nil
				}
				idx++
			}
			return cty.NilVal, fmt.Errorf("item not found in list")
		},
	})
}

// makeStringTransformFunc 包装单字符串输入、单字符串输出的转换函数（如 base64、md5、sha256、urlencode）。
func makeStringTransformFunc(fn func(string) (string, error)) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{{Name: "str", Type: cty.String}},
		Type:   function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			out, err := fn(args[0].AsString())
			if err != nil {
				return cty.NilVal, err
			}
			return cty.StringVal(out), nil
		},
	})
}
