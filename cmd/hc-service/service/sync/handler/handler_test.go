/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package handler

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

type TestHandler struct {
	idx            int
	total          int
	listBatchSize  int
	syncConcurrent uint
	waitSynced     sync.Map
}

func (t *TestHandler) String() string {
	return fmt.Sprintf("TestHandler{idx: %d, total: %d, listBatchSize: %d, wait sync: %d}",
		t.idx, t.total, t.listBatchSize, t.WaitSyncCount())
}

func (t *TestHandler) TestName() string {
	return fmt.Sprintf("total_%d_batch_%d_concurrent_%d", t.total, t.listBatchSize, t.syncConcurrent)
}

func (t *TestHandler) Prepare(cts *rest.Contexts) error {
	return nil
}
func (t *TestHandler) GenerateResID(idx int) string {
	return fmt.Sprintf("test-%08d", idx)
}

func (t *TestHandler) Next(kt *kit.Kit) ([]common.TestCloudRes, error) {

	payload := make([]common.TestCloudRes, 0, t.listBatchSize)
	for i := 0; i < t.listBatchSize && t.idx < t.total; i++ {
		resCloudID := t.GenerateResID(t.idx)
		t.waitSynced.Store(resCloudID, struct{}{})
		payload = append(payload, common.TestCloudRes{CloudID: resCloudID})
		t.idx++
	}
	time.Sleep(5 * time.Millisecond)
	return payload, nil
}

func (t *TestHandler) Sync(kt *kit.Kit, instances []common.TestCloudRes) error {
	for _, instance := range instances {
		_, loaded := t.waitSynced.LoadAndDelete(instance.CloudID)
		if !loaded {
			return fmt.Errorf("instance not found for sync: %s", instance.CloudID)
		}
	}
	time.Sleep(5 * time.Millisecond)
	return nil
}

func (t *TestHandler) WaitSyncCount() int {
	count := 0
	t.waitSynced.Range(func(key, value interface{}) (goon bool) {
		count++
		return true
	})
	return count
}

func (t *TestHandler) RemoveDeletedFromCloud(kt *kit.Kit, allCloudIDMap map[string]struct{}) error {
	return nil
}

func (t *TestHandler) Resource() enumor.CloudResourceType {
	return "test_res"
}

func (t *TestHandler) SyncConcurrent() uint {
	return max(t.syncConcurrent, 1)
}

func (t *TestHandler) Describe() string {
	return "test"
}

func TestResourceSyncV2(t *testing.T) {
	logCfg := logs.LogConfig{
		LogDir:             ".",
		LogMaxSize:         1,
		LogLineMaxSize:     1,
		LogMaxNum:          10,
		RestartNoScrolling: false,
		ToStdErr:           true,
		AlsoToStdErr:       false,
		Verbosity:          3,
	}
	logs.InitLogger(logCfg)
	th := &TestHandler{
		idx:            0,
		total:          20,
		listBatchSize:  200,
		syncConcurrent: uint(6),
	}
	cts := &rest.Contexts{Kit: kit.New()}
	t.Run(th.TestName(), func(t *testing.T) {
		if err := ResourceSyncV2[common.TestCloudRes](cts, th); err != nil {
			t.Errorf("ResourceSyncV2() error = %v, handler: %s", err, th)
		}
		if th.WaitSyncCount() != 0 {
			t.Errorf("synced count not match, handler: %s", th)
		}
	})
	for _, total := range []int{10, 100, 120, 250, 340, 1130} {
		for _, batch := range []int{100, 200, 2000} {
			for _, concurrent := range []int{1, 3, 5, 8, 10} {
				th := &TestHandler{
					idx:            0,
					total:          total,
					listBatchSize:  batch,
					syncConcurrent: uint(concurrent),
				}
				cts := &rest.Contexts{Kit: kit.New()}
				t.Run(th.TestName(), func(t *testing.T) {
					if err := ResourceSyncV2[common.TestCloudRes](cts, th); err != nil {
						t.Errorf("ResourceSyncV2() error = %v, handler: %s", err, th)
					}
					if th.WaitSyncCount() != 0 {
						t.Errorf("synced count not match, handler: %s", th)
					}
				})
			}
		}
	}
}
