/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package dao implements data access object which provides an abstract interface to database
package dao

// DaoSet defines all the DAO to be operated.
type DaoSet interface {
	PoolHost() PoolHost
	LaunchTask() LaunchTask
	RecallTask() RecallTask
	RecallOrder() RecallOrder
	RecallDetail() RecallDetail
	OpRecord() OpRecord
	GradeCfg() GradeCfg
	Zone() Zone
}

type set struct {
	poolHost     PoolHost
	launchTask   LaunchTask
	recallTask   RecallTask
	recallOrder  RecallOrder
	recallDetail RecallDetail
	opRecord     OpRecord
	gradeCfg     GradeCfg
	zone         Zone
}

var daoSet *set

func init() {
	daoSet = &set{
		poolHost:     &poolHostDao{},
		launchTask:   &launchTaskDao{},
		recallTask:   &recallTaskDao{},
		recallOrder:  &recallOrderDao{},
		recallDetail: &recallDetailDao{},
		opRecord:     &opRecordDao{},
		gradeCfg:     &gradeCfgDao{},
		zone:         &zoneDao{},
	}
}

// Set return all dao set interface
func Set() *set {
	return daoSet
}

// PoolHost get resource pool host operation interface
func (s *set) PoolHost() PoolHost {
	return s.poolHost
}

// LaunchTask get resource pool launch task operation interface
func (s *set) LaunchTask() LaunchTask {
	return s.launchTask
}

// RecallTask get resource pool recall task operation interface
func (s *set) RecallTask() RecallTask {
	return s.recallTask
}

// RecallOrder get resource pool recall order operation interface
func (s *set) RecallOrder() RecallOrder {
	return s.recallOrder
}

// RecallDetail get resource pool recall task detail operation interface
func (s *set) RecallDetail() RecallDetail {
	return s.recallDetail
}

// OpRecord get resource pool operation record operation interface
func (s *set) OpRecord() OpRecord {
	return s.opRecord
}

// GradeCfg get resource pool grade config operation interface
func (s *set) GradeCfg() GradeCfg {
	return s.gradeCfg
}

// Zone get qcloud zone config operation interface
func (s *set) Zone() Zone {
	return s.zone
}
