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
	"sort"
	"strings"

	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"
)

// dedupAndSortStrings 对字符串切片去重并按字典序升序排列，输入为空时返回空切片 []string{}。
func dedupAndSortStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	boolMap := cvt.StringSliceToMapBool(items)
	unique := cvt.MapKeyToStringSlice(boolMap)
	sort.Strings(unique)
	return unique
}

// validateAndSortBlocks 检查所有节点的 BlockID 是否唯一、依赖是否有环，并按拓扑顺序排好返回。
// 排序规则：先按拓扑层级升序；同一层内按 import -> data -> resource 排列；同类型再按 BlockID 字典序排列。
//
// 举例：
//
//	假设有 4 个节点：
//	- tencentcloud_subnet.app（依赖 tencentcloud_vpc.main，拓扑层级 1）
//	- tencentcloud_vpc.main（无依赖，拓扑层级 0）
//	- import.tencentcloud_vpc.legacy（无依赖，拓扑层级 0）
//	- data.tencentcloud_images.ubuntu（无依赖，拓扑层级 0）
//	排完序的结果依次是：
//	1) import.tencentcloud_vpc.legacy
//	2) data.tencentcloud_images.ubuntu
//	3) tencentcloud_vpc.main
//	4) tencentcloud_subnet.app
func validateAndSortBlocks(blocks []*Block) ([]*Block, error) {
	if len(blocks) == 0 {
		return []*Block{}, nil
	}
	blockMap := make(map[string]*Block, len(blocks))
	for _, blk := range blocks {
		if _, dup := blockMap[blk.BlockID]; dup {
			return nil, fmt.Errorf("duplicate block_id %q", blk.BlockID)
		}
		blk.DependsOn = dedupAndSortStrings(blk.DependsOn)
		blockMap[blk.BlockID] = blk
	}

	levels, err := computeTopologicalLevelsAndDetectCycles(blocks, blockMap)
	if err != nil {
		return nil, err
	}

	sorted := make([]*Block, len(blocks))
	copy(sorted, blocks)
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if levels[a.BlockID] != levels[b.BlockID] {
			return levels[a.BlockID] < levels[b.BlockID]
		}
		pa, pb := kindSortPriority(a.Kind), kindSortPriority(b.Kind)
		if pa != pb {
			return pa < pb
		}
		return a.BlockID < b.BlockID
	})
	return sorted, nil
}

// computeTopologicalLevelsAndDetectCycles 用 DFS 三色标记法检查节点依赖有没有环，同时算出每个节点的拓扑层级。
// 没有依赖的节点层级为 0，有依赖的节点层级等于所有上游依赖的最大层级加 1。
func computeTopologicalLevelsAndDetectCycles(blocks []*Block, blockMap map[string]*Block) (map[string]int, error) {
	// 0: 未访问, 1: 正在当前递归栈中, 2: 已算完
	color := make(map[string]int, len(blocks))
	levels := make(map[string]int, len(blocks))
	var path []string

	var dfs func(id string) (int, error)
	dfs = func(id string) (int, error) {
		if color[id] == 2 {
			return levels[id], nil
		}
		if color[id] == 1 {
			cycleStr := strings.Join(append(path, id), " -> ")
			return 0, fmt.Errorf("dependency cycle detected: %s", cycleStr)
		}
		blk, exists := blockMap[id]
		if !exists {
			return 0, fmt.Errorf("referenced dependency block %q does not exist", id)
		}
		color[id] = 1
		path = append(path, id)
		maxDepLevel := -1
		for _, depID := range blk.DependsOn {
			depLvl, err := dfs(depID)
			if err != nil {
				return 0, err
			}
			if depLvl > maxDepLevel {
				maxDepLevel = depLvl
			}
		}
		path = path[:len(path)-1]
		color[id] = 2
		levels[id] = maxDepLevel + 1
		return levels[id], nil
	}

	ids := make([]string, len(blocks))
	for i, b := range blocks {
		ids[i] = b.BlockID
	}
	sort.Strings(ids)
	for _, id := range ids {
		if _, err := dfs(id); err != nil {
			return nil, err
		}
	}
	return levels, nil
}

// kindSortPriority 返回同一拓扑层级下的节点类型优先级：import(0) < data(1) < resource(2)。
func kindSortPriority(kind enumor.BlockKind) int {
	switch kind {
	case enumor.BlockKindImport:
		return 0
	case enumor.BlockKindData:
		return 1
	default:
		return 2
	}
}
