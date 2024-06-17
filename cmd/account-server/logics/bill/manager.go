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

package bill

import (
	"context"
	"fmt"
	"time"

	"hcm/pkg/client"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
)

// BillManager bill manager
type BillManager struct {
	Sd                 serviced.ServiceDiscover
	Client             *client.ClientSet
	CurrentControllers map[string]*MainAccountController
	AccountList        AccountLister
}

// Run bill manager
func (bm *BillManager) Run(ctx context.Context) {
	if err := bm.syncControllers(); err != nil {
		logs.Warnf("sync controllers failed, err %s", err.Error())
	}
	ticker := time.NewTicker(defaultControllerSyncDuration)
	for {
		select {
		case <-ticker.C:
			if bm.Sd.IsMaster() {
				if err := bm.syncControllers(); err != nil {
					logs.Warnf("sync controllers failed, err %s", err.Error())
				}
			} else {
				bm.stopControllers()
			}

		case <-ctx.Done():
			logs.Infof("bill manager context done")
			return
		}
	}
}

func (bm *BillManager) syncControllers() error {
	list, err := bm.AccountList.ListAllAccount(getInternalKit())
	if err != nil {
		return err
	}
	for _, item := range list {
		_, ok := bm.CurrentControllers[item.Key()]
		if !ok {
			ctrl, err := NewMainAccountController(&MainAccountControllerOption{
				RootAccountID: item.BaseMainAccount.ParentAccountID,
				MainAccountID: item.BaseMainAccount.CloudID,
				BkBizID:       item.BkBizID,
				ProductID:     item.OpProductID,
				Vendor:        item.Vendor,
				Client:        bm.Client,
			})
			if err != nil {
				return fmt.Errorf("create controller failed, err %s", err)
			}
			if err := ctrl.Start(); err != nil {
				ctrl.Stop()
				return fmt.Errorf("start controller failed, err %s", err.Error())
			}
			bm.CurrentControllers[item.Key()] = ctrl
		}
	}
	return nil
}

func (bm *BillManager) stopControllers() {
	for key, ctrl := range bm.CurrentControllers {
		logs.Warnf("stop controller %s", key)
		ctrl.Stop()
		delete(bm.CurrentControllers, key)
	}
}
