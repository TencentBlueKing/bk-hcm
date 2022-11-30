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

package cmd

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// WithLog init and returns the log command.
func WithLog() Cmd {
	cmd := &defaultCmd{
		cmd: &Command{
			Name:  "log",
			Usage: "change log level",
			Parameters: []Parameter{{
				Name:  "v",
				Usage: "defines the log level to be changed",
				Value: new(int32),
			}},
			FromURL: true,
			Run: func(kt *kit.Kit, params map[string]interface{}) (interface{}, error) {
				v, exists := params["v"]
				if !exists {
					return nil, errf.New(errf.InvalidParameter, "v is not set")
				}

				logs.SetV(*v.(*int32))

				logs.Infof("successfully changed log level to %d, rid: %s", logs.GetV(), kt.Rid)
				return nil, nil
			},
		},
	}

	return cmd
}
