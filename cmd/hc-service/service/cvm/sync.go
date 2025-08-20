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

package cvm

import (
	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/pkg/adaptor/tcloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// syncTCloudCvmWithRelRes sync tcloud cvm with related resources.
func (svc *cvmSvc) syncTCloudCvmWithRelRes(kt *kit.Kit, tcloud tcloud.TCloud, accountID, region string,
	cloudIDs []string) error {

	syncClient := synctcloud.NewClient(svc.dataCli, tcloud)
	params := &synctcloud.SyncBaseParams{
		AccountID: accountID,
		Region:    region,
		CloudIDs:  cloudIDs,
	}

	_, err := syncClient.CvmWithRelRes(kt, params, &synctcloud.SyncCvmWithRelResOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm with res failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}
