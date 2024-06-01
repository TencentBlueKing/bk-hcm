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

package consumer

import (
	"errors"
	"fmt"

	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
)

const (
	// VirtualTaskRootID ...
	VirtualTaskRootID = "virtual_root"
)

// TaskTree task tree
type TaskTree struct {
	Root *TaskNode
	Flow *Flow
}

// TaskNode task node
type TaskNode struct {
	TaskID string
	State  enumor.TaskState

	children []*TaskNode
	parents  []*TaskNode
}

// NewTaskNode new task node
func NewTaskNode(task *Task) *TaskNode {
	return &TaskNode{
		TaskID: task.ID,
		State:  task.State,
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
	return t.State == enumor.TaskSuccess
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

// Executable check can executable: no parent or all parents CanExecuteChild(state == TaskSuccess )
func (t *TaskNode) Executable() bool {
	if t.State != enumor.TaskPending && t.State != enumor.TaskRollback {
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

// ComputeState compute state
func (t *TaskNode) ComputeState() (state enumor.FlowState) {
	walkNode(t, func(node *TaskNode) (proceed bool) {
		switch node.State {
		// 如果Task存在失败节点，无法继续遍历当前节点子节点。
		case enumor.TaskCancel:
			state = enumor.FlowCancel
			return false
		case enumor.TaskFailed:
			state = enumor.FlowFailed
			return false
		// 如果当前节点运行成功，继续遍历当前节点子节点。
		case enumor.TaskSuccess:
			state = enumor.FlowSuccess
			return true

		// 如果当前节点处于其他运行中间状态，无法继续遍历当前节点子节点。
		default:
			state = enumor.FlowRunning
			return false
		}
	})

	return
}

// GetExecutableTasks get executable task nodes
func (t *TaskNode) GetExecutableTasks() (executables []string) {
	walkNode(t, func(node *TaskNode) bool {
		if node.Executable() {
			executables = append(executables, node.TaskID)
		}
		return true
	})

	return
}

// GetExecStateTasks 获取执行状态的节点
func (t *TaskNode) GetExecStateTasks() (ids []string) {
	walkNode(t, func(node *TaskNode) bool {
		if node.State == enumor.TaskRunning || node.State == enumor.TaskRollback {
			ids = append(ids, node.TaskID)
		}
		return true
	})

	return
}

// GetNextExecutableTaskNodes get next executable task nodes
func (t *TaskNode) GetNextExecutableTaskNodes(completedOrRetryTask *Task) (executable []string) {
	walkNode(t, func(node *TaskNode) (proceed bool) {
		if node.TaskID == completedOrRetryTask.ID {
			node.State = completedOrRetryTask.State
			// running 和 rollback 都要放回去执行
			if node.State == enumor.TaskRunning || node.State == enumor.TaskRollback {
				executable = append(executable, node.TaskID)
				return false
			}

			if !node.CanExecuteChild() {
				return false
			}

			for i := range node.children {
				if node.children[i].Executable() {
					executable = append(executable, node.children[i].TaskID)
				}
			}

			return false
		}
		return true
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
			if _, ok := visited[p.TaskID]; !ok {
				return false
			}
		}
		return true
	}

	for i := 0; i < queueLen; i++ {
		cur := waitQueue[i]
		if !isParentCompleted(cur) {
			incomplete[cur.TaskID] = cur
			continue
		}

		visited[cur.TaskID] = struct{}{}

		delete(incomplete, cur.TaskID)

		waitQueue = append(waitQueue, cur.children...)
	}

	waitQueue = waitQueue[queueLen:]

	bfsCheckCycle(waitQueue, visited, incomplete)
}

// BuildTaskRoot build task root
func BuildTaskRoot(tasks []*Task) (*TaskNode, error) {
	root := &TaskNode{
		TaskID: VirtualTaskRootID,
		State:  enumor.TaskSuccess,
	}

	m, err := buildTaskNodeMap(tasks)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if len(task.DependOn) == 0 {
			n := m[task.ActionID]
			n.AppendParent(root)
			root.children = append(root.children, n)
		}

		if len(task.DependOn) > 0 {
			for _, dependId := range task.DependOn {
				parent, ok := m[dependId]
				if !ok {
					return nil, fmt.Errorf("does not find task[%s] depend: %s", task.ID, dependId)
				}
				parent.AppendChild(m[task.ActionID])
				m[task.ActionID].AppendParent(parent)
			}
		}
	}

	if len(root.children) == 0 {
		return nil, errors.New("here is no start nodes")
	}

	if cycleStart := root.HasCycle(); cycleStart != nil {
		return nil, fmt.Errorf("has cycle at: %s", cycleStart.TaskID)
	}

	return root, nil
}

func buildTaskNodeMap(tasks []*Task) (map[action.ActIDType]*TaskNode, error) {
	m := map[action.ActIDType]*TaskNode{}

	for i := range tasks {
		if _, ok := m[tasks[i].ActionID]; ok {
			return nil, fmt.Errorf("task actionID is repeat, actionID: %s", tasks[i].ActionID)
		}
		m[tasks[i].ActionID] = NewTaskNode(tasks[i])
	}

	return m, nil
}

func walkNode(root *TaskNode, walkFunc func(node *TaskNode) (proceed bool)) {
	dfsWalk(root, walkFunc)
}

// dfsWalk 从某个节点进行深度优先遍历并依次遍历它的子节点
// Note: walkFunc
func dfsWalk(root *TaskNode, walkFunc func(node *TaskNode) (proceed bool)) (stop bool) {
	if root.TaskID != VirtualTaskRootID {
		if !walkFunc(root) {
			return false
		}
	}

	// 当前节点是否可以遍历子节点，如果不能遍历子节点，达到当前路线叶子节点
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
