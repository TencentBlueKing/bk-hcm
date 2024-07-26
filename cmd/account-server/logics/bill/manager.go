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

	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
)

// BillManager bill manager
type BillManager struct {
	Sd                     serviced.ServiceDiscover
	Client                 *client.ClientSet
	CurrentMainControllers map[string]*MainAccountController
	CurrentRootControllers map[string]*RootAccountController
	AccountList            AccountLister
}

// Run bill manager
func (bm *BillManager) Run(ctx context.Context) {
	if bm.Sd.IsMaster() {
		if err := bm.syncMainControllers(); err != nil {
			logs.Warnf("sync main controllers failed, err %s", err.Error())
		}
		if err := bm.syncRootControllers(); err != nil {
			logs.Warnf("sync root controllers failed, err %s", err.Error())
		}
	} else {
		bm.stopControllers()
	}

	ticker := time.NewTicker(*cc.AccountServer().Controller.ControllerSyncDuration)
	for {
		select {
		case <-ticker.C:
			if bm.Sd.IsMaster() {
				if err := bm.syncMainControllers(); err != nil {
					logs.Warnf("sync main controllers failed, err %s", err.Error())
				}
				if err := bm.syncRootControllers(); err != nil {
					logs.Warnf("sync root controllers failed, err %s", err.Error())
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

func (bm *BillManager) syncRootControllers() error {
	list, err := bm.AccountList.ListAllRootAccount(getInternalKit())
	if err != nil {
		return err
	}
	existedAccountKeyMap := make(map[string]struct{})
	for _, item := range list {
		existedAccountKeyMap[item.Key()] = struct{}{}
		_, ok := bm.CurrentRootControllers[item.Key()]
		if !ok {
			opt := RootAccountControllerOption{
				RootAccountID: item.BaseRootAccount.ID,
				Vendor:        item.Vendor,
				Client:        bm.Client,
				Sd:            bm.Sd,
			}
			ctrl, err := NewRootAccountController(&opt)
			if err != nil {
				return fmt.Errorf("create root for %v controller failed, err %s", opt, err)
			}
			if err := ctrl.Start(); err != nil {
				ctrl.Stop()
				return fmt.Errorf("start controller failed, err %s", err.Error())
			}
			bm.CurrentRootControllers[item.Key()] = ctrl
			logs.Infof("start root account controller for %+v", opt)
		}
	}
	for key, controller := range bm.CurrentRootControllers {
		if _, ok := existedAccountKeyMap[key]; !ok {
			controller.Stop()
			logs.Infof("stop root account controller for %v", controller)
			delete(bm.CurrentRootControllers, key)
		}
	}
	return nil
}

func (bm *BillManager) syncMainControllers() error {
	// TODO: 所有主账号 改为 所有未核算完成的主账号
	list, err := bm.AccountList.ListAllMainAccount(getInternalKit())
	if err != nil {
		return err
	}
	existedAccountKeyMap := make(map[string]struct{})
	for _, item := range list {
		existedAccountKeyMap[item.Key()] = struct{}{}
		_, ok := bm.CurrentMainControllers[item.Key()]
		if !ok {
			opt := MainAccountControllerOption{
				RootAccountID: item.BaseMainAccount.ParentAccountID,
				MainAccountID: item.BaseMainAccount.ID,
				BkBizID:       item.BkBizID,
				ProductID:     item.OpProductID,
				Vendor:        item.Vendor,
				Client:        bm.Client,
				Sd:            bm.Sd,
			}
			ctrl, err := NewMainAccountController(&opt)
			if err != nil {
				return fmt.Errorf("create controller failed, err %s", err)
			}
			if err := ctrl.Start(); err != nil {
				ctrl.Stop()
				return fmt.Errorf("start controller failed, err %s", err.Error())
			}
			bm.CurrentMainControllers[item.Key()] = ctrl
			logs.Infof("start main account controller for %v", opt)
		}
	}
	for key, controller := range bm.CurrentMainControllers {
		if _, ok := existedAccountKeyMap[key]; !ok {
			controller.Stop()
			logs.Infof("stop main account controller for %v", controller)
			delete(bm.CurrentMainControllers, key)
		}
	}
	return nil
}

func (bm *BillManager) stopControllers() {
	for key, ctrl := range bm.CurrentRootControllers {
		logs.Warnf("stop root account controller %s", key)
		ctrl.Stop()
		delete(bm.CurrentRootControllers, key)
	}

	for key, ctrl := range bm.CurrentMainControllers {
		logs.Warnf("stop main account controller %s", key)
		ctrl.Stop()
		delete(bm.CurrentMainControllers, key)
	}
}
