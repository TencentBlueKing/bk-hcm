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

package tableasync

import (
	"database/sql/driver"
	"strings"
	"sync"

	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

// NewShareData new share data.
func NewShareData() *ShareData {
	return &ShareData{
		Dict:  make(map[string]string),
		mutex: sync.Mutex{},
	}
}

// ShareData can read/write within all tasks and will persist it
// if you want a high performance just within same task, you can use
// ExecuteContext's Context
type ShareData struct {
	Dict map[string]string
	Save func(kt *kit.Kit, data *ShareData) error

	mutex sync.Mutex
}

// Scan is used to decode raw message which is read from db into ShareData.
func (d *ShareData) Scan(raw interface{}) error {
	d.Dict = make(map[string]string)
	return types.Scan(raw, d.Dict)
}

// Value encode the ShareData to a json raw, so that it can be stored to db with json raw.
func (d ShareData) Value() (driver.Value, error) {
	return types.Value(d.Dict)
}

// MarshalJSON used by json
func (d *ShareData) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Dict)
}

// UnmarshalJSON used by json
func (d *ShareData) UnmarshalJSON(data []byte) error {
	if d.Dict == nil {
		d.Dict = make(map[string]string)
	}
	return json.Unmarshal(data, &d.Dict)
}

// Get value from share data, it is thread-safe.
func (d *ShareData) Get(key string) (string, bool) {
	if d.Dict == nil {
		return "", false
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	v, ok := d.Dict[key]
	return v, ok
}

// Set value to share data, it is thread-safe.
func (d *ShareData) Set(kt *kit.Kit, key string, val string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Dict[key] = val
	if d.Save != nil {
		if err := d.Save(kt, d); err != nil {
			delete(d.Dict, key)
			logs.ErrorDepthf(1, "share data set key: %s, val: %s failed, err: %v, rid: %s", key, val, err, kt.Rid)
			return err
		}
	}

	return nil
}

// AppendSuccessIDs append success ids.
func (d *ShareData) AppendSuccessIDs(kt *kit.Kit, ids ...string) error {
	return d.appendIDs(kt, "success_ids", ids...)
}

// AppendFailedIDs append failed ids.
func (d *ShareData) AppendFailedIDs(kt *kit.Kit, ids ...string) error {
	return d.appendIDs(kt, "failed_ids", ids...)
}

// AppendSuccessCloudIDs append success cloud ids.
func (d *ShareData) AppendSuccessCloudIDs(kt *kit.Kit, ids ...string) error {
	return d.appendIDs(kt, "success_cloud_ids", ids...)
}

// AppendFailedCloudIDs append failed cloud ids.
func (d *ShareData) AppendFailedCloudIDs(kt *kit.Kit, ids ...string) error {
	return d.appendIDs(kt, "failed_cloud_ids", ids...)
}

// AppendUnknownCloudIDs append unknown cloud ids.
func (d *ShareData) AppendUnknownCloudIDs(kt *kit.Kit, ids ...string) error {
	return d.appendIDs(kt, "unknown_cloud_ids", ids...)
}

func (d *ShareData) appendIDs(kt *kit.Kit, key string, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	val, exist := d.Dict[key]
	var str string
	if !exist {
		str = strings.Join(ids, ",")
	} else {
		str = val + "," + strings.Join(ids, ",")
	}
	d.Dict[key] = str
	if d.Save != nil {
		if err := d.Save(kt, d); err != nil {
			delete(d.Dict, key)
			logs.ErrorDepthf(1, "share data appendIDs: %s, ids: %v failed, err: %v, rid: %s", key, ids, err, kt.Rid)
			return err
		}
	}

	return nil
}
