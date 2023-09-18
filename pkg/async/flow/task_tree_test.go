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

package flow

import (
	"fmt"
	"testing"

	"hcm/pkg/async/task"
	"hcm/pkg/criteria/enumor"

	"github.com/stretchr/testify/assert"
)

// TestExecutable test Executable func
func TestExecutable(t *testing.T) {
	tests := []struct {
		caseDesc     string
		giveTaskNode *TaskNode
		wantRet      bool
	}{
		{
			caseDesc: "pending task",
			giveTaskNode: &TaskNode{
				Task: &task.Task{
					State: enumor.TaskPending,
				},
			},
			wantRet: true,
		},
		{
			caseDesc: "running task",
			giveTaskNode: &TaskNode{
				Task: &task.Task{
					State: enumor.TaskRunning,
				},
			},
			wantRet: false,
		},
		{
			caseDesc: "failed task",
			giveTaskNode: &TaskNode{
				Task: &task.Task{
					State: enumor.TaskFailed,
				},
			},
			wantRet: false,
		},
		{
			caseDesc: "before success task",
			giveTaskNode: &TaskNode{
				Task: &task.Task{
					State: enumor.TaskBeforeSuccess,
				},
			},
			wantRet: true,
		},
		{
			caseDesc: "pending task has succeed parents",
			giveTaskNode: &TaskNode{
				Task: &task.Task{
					State: enumor.TaskPending,
				},
				parents: []*TaskNode{
					{
						Task: &task.Task{
							State: enumor.TaskSuccess,
						},
					},
					{
						Task: &task.Task{
							State: enumor.TaskSuccess,
						},
					},
				},
			},
			wantRet: true,
		},
		{
			caseDesc: "pending task has failed parents",
			giveTaskNode: &TaskNode{
				Task: &task.Task{
					State: enumor.TaskPending,
				},
				parents: []*TaskNode{
					{
						Task: &task.Task{
							State: enumor.TaskSuccess,
						},
					},
					{
						Task: &task.Task{
							State: enumor.TaskFailed,
						},
					},
				},
			},
			wantRet: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.caseDesc, func(t *testing.T) {
			assert.Equal(t, tc.wantRet, tc.giveTaskNode.Executable())
		})
	}
}

// TestComputeStatus test ComputeStatus func
func TestComputeStatus(t *testing.T) {
	tests := []struct {
		caseDesc    string
		giveTaskIns []task.Task
		wantStatus  enumor.FlowState
	}{
		{
			caseDesc: "flow success",
			giveTaskIns: []task.Task{
				{
					ID:    "task1",
					State: enumor.TaskSuccess,
				},
				{
					ID:       "task2",
					DependOn: []string{"task1"},
					State:    enumor.TaskSuccess,
				},
				{
					ID:       "task3",
					DependOn: []string{"task1"},
					State:    enumor.TaskSuccess,
				},
			},
			wantStatus: enumor.FlowSuccess,
		},
		{
			caseDesc: "flow failed",
			giveTaskIns: []task.Task{
				{
					ID:    "task1",
					State: enumor.TaskSuccess,
				},
				{
					ID:       "task2",
					DependOn: []string{"task1"},
					State:    enumor.TaskSuccess,
				},
				{
					ID:       "task3",
					DependOn: []string{"task1"},
					State:    enumor.TaskFailed,
				},
				{
					ID:       "task4",
					DependOn: []string{"task2", "task3"},
					State:    enumor.TaskSuccess,
				},
			},
			wantStatus: enumor.FlowFailed,
		},
		{
			caseDesc: "flow running",
			giveTaskIns: []task.Task{
				{
					ID:    "task1",
					State: enumor.TaskSuccess,
				},
				{
					ID:       "task2",
					DependOn: []string{"task1"},
					State:    enumor.TaskFailed,
				},
				{
					ID:       "task3",
					DependOn: []string{"task1"},
					State:    enumor.TaskPending,
				},
			},
			wantStatus: enumor.FlowRunning,
		},
	}

	for _, tc := range tests {
		t.Run(tc.caseDesc, func(t *testing.T) {
			root, _ := BuildTaskRoot(tc.giveTaskIns)
			status := root.ComputeStatus()
			assert.Equal(t, tc.wantStatus, status)
		})
	}
}

// TestBuildRootNode test BuildRootNode func
func TestBuildRootNode(t *testing.T) {
	tests := []struct {
		caseDesc  string
		giveTasks []task.Task
		wantRoot  *TaskNode
		wantErr   error
	}{
		{
			caseDesc: "no start nodes",
			giveTasks: []task.Task{
				{
					ID:       "child1",
					DependOn: []string{"child1"},
				},
				{
					ID:       "child2",
					DependOn: []string{"child1"},
				},
			},
			wantRoot: nil,
			wantErr:  fmt.Errorf("here is no start nodes"),
		},
		{
			caseDesc: "parent not existed",
			giveTasks: []task.Task{
				{
					ID: "root1",
				},
				{
					ID:       "r1-child1",
					DependOn: []string{"root2"},
				},
			},
			wantRoot: nil,
			wantErr:  fmt.Errorf("does not find task[r1-child1] depend: root2"),
		},
		{
			caseDesc: "has cycle",
			giveTasks: []task.Task{
				{
					ID: "root",
				},
				{
					ID:       "child1",
					DependOn: []string{"root", "child3"},
				},
				{
					ID:       "child2",
					DependOn: []string{"child1"},
				},
				{
					ID:       "child3",
					DependOn: []string{"child2"},
				},
			},
			wantRoot: nil,
			wantErr:  fmt.Errorf("has cycle at: child1"),
		},
		{
			caseDesc: "normal",
			giveTasks: []task.Task{
				{
					ID: "root1",
				},
				{
					ID:       "r1-child1",
					DependOn: []string{"root1"},
				},
				{
					ID:       "r1-child2",
					DependOn: []string{"root1"},
				},
				{
					ID:       "c1-child1",
					DependOn: []string{"r1-child2"},
				},
				{
					ID: "root2",
				},
				{
					ID:       "r2-child1",
					DependOn: []string{"root2"},
				},
			},
			wantRoot: &TaskNode{
				Task: &task.Task{
					ID:          VirtualTaskRootID,
					ActionName:  string(enumor.VirRoot),
					State:       enumor.TaskSuccess,
					TimeoutSecs: VirtualTaskRootTimeoutSecs,
					FlowID:      VirtualTaskRootFlowID,
					DependOn:    []string{},
				},
				children: []*TaskNode{
					{
						Task: &task.Task{
							ID: "root1",
						},
						children: []*TaskNode{
							{
								Task: &task.Task{
									ID: "r1-child1",
								},
							},
							{
								Task: &task.Task{
									ID: "r1-child2",
								},
								children: []*TaskNode{
									{
										Task: &task.Task{
											ID: "c1-child1",
										},
									},
								},
							},
						},
					},
					{
						Task: &task.Task{
							ID: "root2",
						},
						children: []*TaskNode{
							{
								Task: &task.Task{
									ID: "r2-child1",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.caseDesc, func(t *testing.T) {
			ret, err := BuildTaskRoot(tc.giveTasks)
			assert.Equal(t, tc.wantErr, err)
			if err == nil {
				checkParentAndRemoveIt(t, ret, nil)
				assert.Equal(t, tc.wantRoot.Task.ID, ret.Task.ID)
				assert.Equal(t, tc.wantRoot.children[0].Task.ID, ret.children[0].Task.ID)
				assert.Equal(t, tc.wantRoot.children[0].children[0].Task.ID, ret.children[0].children[0].Task.ID)
				assert.Equal(t, tc.wantRoot.children[0].children[1].Task.ID, ret.children[0].children[1].Task.ID)
				assert.Equal(t, tc.wantRoot.children[0].children[1].children[0].Task.ID,
					ret.children[0].children[1].children[0].Task.ID)
				assert.Equal(t, tc.wantRoot.children[1].Task.ID, ret.children[1].Task.ID)
				assert.Equal(t, tc.wantRoot.children[1].children[0].Task.ID, ret.children[1].children[0].Task.ID)
			}
		})
	}
}

func checkParentAndRemoveIt(t *testing.T, node, pNode *TaskNode) {
	if pNode != nil {
		find := false
		var newParents []*TaskNode
		for _, n := range node.parents {
			if n == pNode {
				find = true
				continue
			}
			newParents = append(newParents, n)
		}
		node.parents = newParents
		if !find {
			assert.Fail(t, "parent node is not contain")
			return
		}
	}

	for _, c := range node.children {
		checkParentAndRemoveIt(t, c, node)
	}
}
