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

package async

import (
	"errors"
	"fmt"

	_ "hcm/cmd/task-server/service/async/tasks"
	"hcm/cmd/task-server/service/capability"
	taskserver "hcm/pkg/api/task-server"
	pasync "hcm/pkg/async"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	mtasks "github.com/RichardKnop/machinery/v1/tasks"
)

// InitAsyncService initial the async service
func InitAsyncService(cap *capability.Capability) {
	svc := &async{
		cs: cap.ApiClient,
		as: cap.Async,
	}

	h := rest.NewHandler()

	h.Add("CreateAsyncTask", "POST", "/async/tasks/create", svc.CreateAsyncTask)

	h.Load(cap.WebService)
}

type async struct {
	cs *client.ClientSet
	as *pasync.TaskServer
}

// CreateAsyncTask create aysnc task
func (a async) CreateAsyncTask(cts *rest.Contexts) (interface{}, error) {
	req := new(taskserver.AsyncTask)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("async create async task request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tasks, err := a.changeAsyncTaskToMachineryTask(req.Steps)
	if err != nil {
		logs.Errorf("async change asyncTask to machineryTask failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errors.New("async change asyncTask to machineryTask failed")
	}

	resp := &taskserver.AsyncTaskResp{
		TaskType: string(req.TaskType),
	}

	switch req.TaskType {
	case enumor.SingleAsyncTask:
		taskID, err := a.as.SendSingleTask(tasks[0])
		if err != nil {
			logs.Errorf("async send singleTask failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errors.New("async send singleTask failed")
		}
		resp.TaskID = taskID
	case enumor.GroupAsyncTask:
		groupID, err := a.as.SendGroupTasks(tasks...)
		if err != nil {
			logs.Errorf("async send groupTask failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errors.New("async send groupTask failed")
		}
		resp.GroupID = groupID
	case enumor.ChordAsyncTask:
		tasks, err := a.changeAsyncTaskToMachineryTask([]*taskserver.Step{req.CallBackTask})
		if err != nil {
			logs.Errorf("async change asyncTask to machineryTask failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errors.New("async change asyncTask to machineryTask failed")
		}
		groupID, err := a.as.SendChordTasks(tasks[0], tasks...)
		if err != nil {
			logs.Errorf("async send chordTask failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errors.New("async send chordTask failed")
		}
		resp.GroupID = groupID
	case enumor.ChainAsyncTask:
		taskID, err := a.as.SendChainTasks(tasks...)
		if err != nil {
			logs.Errorf("async send chainTask failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, errors.New("async send chainTask failed")
		}
		resp.TaskID = taskID
	case enumor.CronAsyncTask:
	default:
		return nil, fmt.Errorf("async task type: %s not support", req.TaskType)
	}

	return resp, nil
}

// changeAsyncTaskToMachineryTask change asyncTask to machineryTask
func (a async) changeAsyncTaskToMachineryTask(steps []*taskserver.Step) ([]*mtasks.Signature, error) {
	if len(steps) <= 0 {
		return nil, errors.New("async not set steps")
	}

	tasks := make([]*mtasks.Signature, 0, len(steps))

	for _, one := range steps {
		if err := one.Validate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		args := make([]mtasks.Arg, 0, len(one.Args))
		if len(one.Args) > 0 {
			var err error
			args, err = a.changeArgsToMachineryArgs(one.Args)
			if err != nil {
				return nil, errors.New("async change args to machineryArgs failed")
			}
		}

		onSuccess := make([]*mtasks.Signature, 0, len(one.OnSuccess))
		if len(one.OnSuccess) > 0 {
			var err error
			onSuccess, err = a.changeAsyncTaskToMachineryTask(one.OnSuccess)
			if err != nil {
				return nil, errors.New("async change asyncTask to machineryTask failed")
			}
		}

		onError := make([]*mtasks.Signature, 0, len(one.OnError))
		if len(one.OnError) > 0 {
			var err error
			onError, err = a.changeAsyncTaskToMachineryTask(one.OnError)
			if err != nil {
				return nil, errors.New("async change asyncTask to machineryTask failed")
			}
		}

		task := &mtasks.Signature{
			Immutable:      one.Immutable,
			UUID:           one.UUID,
			Name:           one.TaskName,
			Priority:       uint8(one.TaskPriority),
			Args:           args,
			RetryCount:     one.RetryCount,
			GroupUUID:      one.GroupUUID,
			GroupTaskCount: one.GroupTaskCount,
			OnSuccess:      onSuccess,
			OnError:        onError,
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// changeArgsToMachineryArgs change args to machineryArgs
func (a async) changeArgsToMachineryArgs(args []taskserver.Arg) ([]mtasks.Arg, error) {
	if len(args) <= 0 {
		return nil, errors.New("async step not set arg")
	}

	mArgs := make([]mtasks.Arg, 0, len(args))

	for _, one := range args {
		if err := one.Validate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}

		arg := mtasks.Arg{
			Type:  one.Type,
			Value: one.Value,
		}

		mArgs = append(mArgs, arg)
	}

	return mArgs, nil
}
