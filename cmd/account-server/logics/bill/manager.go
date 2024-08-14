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
			logs.Warnf("sync main controllers failed, err: %s", err.Error())
		}
		if err := bm.syncRootControllers(); err != nil {
			logs.Warnf("sync root controllers failed, err: %s", err.Error())
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
					logs.Warnf("sync main controllers failed, err: %s", err.Error())
				}
				if err := bm.syncRootControllers(); err != nil {
					logs.Warnf("sync root controllers failed, err: %s", err.Error())
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

	kt := getInternalKit()
	rootAccounts, err := bm.AccountList.ListAllRootAccount(kt)
	if err != nil {
		return err
	}
	existedAccountKeyMap := make(map[string]struct{})
	for _, rootAccount := range rootAccounts {
		existedAccountKeyMap[rootAccount.Key()] = struct{}{}
		_, ok := bm.CurrentRootControllers[rootAccount.Key()]
		if !ok {
			opt := RootAccountControllerOption{
				RootAccountID:      rootAccount.ID,
				RootAccountCloudID: rootAccount.CloudID,
				Vendor:             rootAccount.Vendor,
				Client:             bm.Client,
			}
			ctrl, err := NewRootAccountController(&opt)
			if err != nil {
				logs.Errorf("create root for %s %s controller failed, err: %v, rid: %s",
					opt.Vendor, opt.RootAccountID, err, kt.Rid)
				return fmt.Errorf("create root for %v controller failed, err: %s", opt, err)
			}
			if err := ctrl.Start(); err != nil {
				ctrl.Stop()
				logs.Errorf("fail to start controller, err: %s, rid: %s", err.Error(), kt.Rid)
				return fmt.Errorf("start controller failed, err: %s, rid: %s", err.Error(), kt.Rid)
			}
			bm.CurrentRootControllers[rootAccount.Key()] = ctrl
			logs.Infof("start root account controller for [%s]%s, rid: %s", opt.Vendor, opt.RootAccountID, kt.Rid)
		}
	}
	for key, controller := range bm.CurrentRootControllers {
		if _, ok := existedAccountKeyMap[key]; !ok {
			controller.Stop()
			logs.Infof("stop root account controller for %v, rid: %s", controller, kt.Rid)
			delete(bm.CurrentRootControllers, key)
		}
	}
	return nil
}

func (bm *BillManager) syncMainControllers() error {
	// TODO: 所有主账号 改为 所有未核算完成的主账号
	kt := getInternalKit()
	mainAccounts, err := bm.AccountList.ListAllMainAccount(kt)
	if err != nil {
		return err
	}

	existedAccountKeyMap := make(map[string]struct{})
	for _, mainAccount := range mainAccounts {
		existedAccountKeyMap[mainAccount.Key()] = struct{}{}
		_, ok := bm.CurrentMainControllers[mainAccount.Key()]
		if ok {
			continue
		}
		// 获取root account 信息
		rootAccount, err := bm.Client.DataService().Global.RootAccount.GetBasicInfo(kt,
			mainAccount.BaseMainAccount.ParentAccountID)
		if err != nil {
			logs.Errorf("get root account for main account controller failed, err: %v, rid: %s", err, kt.Rid)
			return fmt.Errorf("get root account for main account controller failed, err: %w", err)
		}

		opt := MainAccountControllerOption{
			RootAccountID: mainAccount.BaseMainAccount.ParentAccountID,
			MainAccountID: mainAccount.BaseMainAccount.ID,
			BkBizID:       mainAccount.BkBizID,
			ProductID:     mainAccount.OpProductID,
			Vendor:        mainAccount.Vendor,
			Client:        bm.Client,

			RootAccountCloudID: rootAccount.CloudID,
			MainAccountCloudID: mainAccount.CloudID,
		}
		ctrl, err := NewMainAccountController(&opt)
		if err != nil {
			return fmt.Errorf("create main account controller failed, err: %w", err)
		}
		if err := ctrl.Start(); err != nil {
			ctrl.Stop()
			logs.Errorf("fail to start main account controller of %s, vendor: %s, root account: %s, biz: %d,"+
				" product: %d, rid: %s",
				opt.MainAccountID, opt.Vendor, opt.RootAccountID, opt.BkBizID, opt.ProductID, kt.Rid)
			return fmt.Errorf("start main account controller failed, err: %w", err)
		}
		bm.CurrentMainControllers[mainAccount.Key()] = ctrl
		logs.Infof("start main account controller for %s, vednor: %s, root_account: %s, biz: %d, product: %d, rid: %s",
			opt.MainAccountID, opt.Vendor, opt.RootAccountID, opt.BkBizID, opt.ProductID, kt.Rid)
	}
	for key, controller := range bm.CurrentMainControllers {
		if _, ok := existedAccountKeyMap[key]; !ok {
			controller.Stop()
			logs.Infof("stop main account controller for  %s, vednor: %s, root_account: %s, biz: %d, "+
				"product: %d, rid: %s",
				controller.MainAccountID, controller.Vendor, controller.RootAccountID, controller.BkBizID,
				controller.ProductID, kt.Rid)
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
