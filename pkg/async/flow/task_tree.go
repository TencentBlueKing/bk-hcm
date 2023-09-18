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
	"errors"
	"fmt"
	"sync"

	"hcm/pkg/async/task"
	"hcm/pkg/criteria/enumor"
)

const (
	VirtualTaskRootID          = "virtual_root"
	VirtualTaskRootFlowID      = "vir_flow_id"
	VirtualTaskRootFlowName    = "vir_flow_name"
	VirtualTaskRootTimeoutSecs = 20
)

// TaskTree task tree
type TaskTree struct {
	Root             *TaskNode
	RunTaskNodesLock *sync.Mutex
	RunTaskNodes     map[string]*TaskNode
	FlowState        enumor.FlowState
	Reason           string
}

// NewTaskTree new task tree
func NewTaskTree() *TaskTree {
	return &TaskTree{
		RunTaskNodesLock: &sync.Mutex{},
		RunTaskNodes:     make(map[string]*TaskNode),
		FlowState:        enumor.FlowPending,
	}
}

// TaskNode task node
type TaskNode struct {
	Task *task.Task

	children []*TaskNode
	parents  []*TaskNode
}

// NewTaskNode new task node
func NewTaskNode(task *task.Task) *TaskNode {
	return &TaskNode{
		Task: task,
	}
}

// AppendChild append child
func (t *TaskNode) AppendChild(task *TaskNode) {
	t.children = append(t.children, task)
}

// AppendParent append parent
func (t *TaskNode) AppendParent(task *TaskNode) {
	t.parents = append(t.parents, task)
}

// GetChildren get children
func (t *TaskNode) GetChildren() []*TaskNode {
	return t.children
}

// GetParents get parents
func (t *TaskNode) GetParents() []*TaskNode {
	return t.parents
}

// CanExecuteChild can execute child
func (t *TaskNode) CanExecuteChild() bool {
	return t.Task.State == enumor.TaskSuccess
}

// CanBeExecuted check whether task could be executed
func (t *TaskNode) CanBeExecuted() bool {
	if len(t.parents) == 0 {
		return true
	}

	for _, p := range t.parents {
		if !p.CanExecuteChild() {
			return false
		}
	}
	return true
}

// Executable check can executable
func (t *TaskNode) Executable() bool {
	if t.Task.State != enumor.TaskPending &&
		t.Task.State != enumor.TaskBeforeSuccess &&
		t.Task.State != enumor.TaskBeforeFailed {
		return false
	}

	if len(t.parents) == 0 {
		return true
	}

	for i := range t.parents {
		if !t.parents[i].CanExecuteChild() {
			return false
		}
	}

	return true
}

// ComputeStatus compute status
func (t *TaskNode) ComputeStatus() (state enumor.FlowState) {
	walkNode(t, func(node *TaskNode) bool {
		switch node.Task.State {
		case enumor.TaskFailed:
			state = enumor.FlowFailed
			return true
		case enumor.TaskSuccess:
			state = enumor.FlowSuccess
			return true
		default:
			state = enumor.FlowRunning
			return false
		}
	})

	return
}

// GetExecutableTaskNodes get executable task nodes
func (t *TaskNode) GetExecutableTaskNodes() (executables []*TaskNode) {
	walkNode(t, func(node *TaskNode) bool {
		if node.Executable() {
			executables = append(executables, node)
		}
		return true
	})

	return
}

// GetNextTaskNodes get next task nodes
func (t *TaskNode) GetNextTaskNodes() (executable []*TaskNode, find bool) {
	walkNode(t, func(node *TaskNode) bool {
		find = true

		if node.Task.State == enumor.TaskPending {
			executable = append(executable, node)
			return false
		}

		if !node.CanExecuteChild() {
			return false
		}

		for i := range node.children {
			if node.children[i].Executable() {
				executable = append(executable, node.children[i])
			}
		}

		return false
	})

	return
}

// HasCycle check has cycle
func (t *TaskNode) HasCycle() (cycleStart *TaskNode) {
	visited, incomplete := map[string]struct{}{}, map[string]*TaskNode{}
	waitQueue := []*TaskNode{t}

	bfsCheckCycle(waitQueue, visited, incomplete)

	if len(incomplete) > 0 {
		for k := range incomplete {
			return incomplete[k]
		}
	}

	return
}

func bfsCheckCycle(waitQueue []*TaskNode, visited map[string]struct{}, incomplete map[string]*TaskNode) {
	queueLen := len(waitQueue)
	if queueLen == 0 {
		return
	}

	isParentCompleted := func(node *TaskNode) bool {
		for _, p := range node.parents {
			if _, ok := visited[p.Task.ID]; !ok {
				return false
			}
		}
		return true
	}

	for i := 0; i < queueLen; i++ {
		cur := waitQueue[i]
		if !isParentCompleted(cur) {
			incomplete[cur.Task.ID] = cur
			continue
		}

		visited[cur.Task.ID] = struct{}{}

		delete(incomplete, cur.Task.ID)

		waitQueue = append(waitQueue, cur.children...)
	}

	waitQueue = waitQueue[queueLen:]

	bfsCheckCycle(waitQueue, visited, incomplete)
}

// BuildTaskRoot build task root
func BuildTaskRoot(tasks []task.Task) (*TaskNode, error) {
	root := &TaskNode{
		Task: &task.Task{
			ID:          VirtualTaskRootID,
			FlowID:      VirtualTaskRootFlowID,
			FlowName:    VirtualTaskRootFlowName,
			ActionName:  string(enumor.VirRoot),
			State:       enumor.TaskSuccess,
			TimeoutSecs: VirtualTaskRootTimeoutSecs,
			DependOn:    []string{},
		},
	}

	m, err := buildTaskNodeMap(tasks)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if len(task.DependOn) == 0 {
			n := m[task.ID]
			n.AppendParent(root)
			root.children = append(root.children, n)
		}

		if len(task.DependOn) > 0 {
			for _, dependId := range task.DependOn {
				parent, ok := m[dependId]
				if !ok {
					return nil, fmt.Errorf("does not find task[%s] depend: %s", task.ID, dependId)
				}
				parent.AppendChild(m[task.ID])
				m[task.ID].AppendParent(parent)
			}
		}
	}

	if len(root.children) == 0 {
		return nil, errors.New("here is no start nodes")
	}

	if cycleStart := root.HasCycle(); cycleStart != nil {
		return nil, fmt.Errorf("has cycle at: %s", cycleStart.Task.ID)
	}

	return root, nil
}

func buildTaskNodeMap(tasks []task.Task) (map[string]*TaskNode, error) {
	m := map[string]*TaskNode{}

	for i := range tasks {
		if _, ok := m[tasks[i].ID]; ok {
			return nil, fmt.Errorf("task id is repeat, id: %s", tasks[i].ID)
		}
		m[tasks[i].ID] = NewTaskNode(&tasks[i])
	}

	return m, nil
}

func walkNode(root *TaskNode, walkFunc func(node *TaskNode) bool) {
	dfsWalk(root, walkFunc)
}

// dfsWalk 从某个节点进行深度优先遍历并依次遍历它的子节点
func dfsWalk(root *TaskNode, walkFunc func(node *TaskNode) bool) bool {
	if root.Task.ID != VirtualTaskRootID {
		if !walkFunc(root) {
			return false
		}
	}

	// we cannot execute children, but should execute brother nodes
	if !root.CanExecuteChild() {
		return true
	}

	for _, c := range root.children {
		// if children's parent is not just root, we must check it
		if len(c.parents) > 1 && !c.CanBeExecuted() {
			continue
		}

		if !dfsWalk(c, walkFunc) {
			return false
		}
	}

	return true
}
