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

package bkcc

import (
	"fmt"
	"sync"

	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/api-gateway/cmdb"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Watcher sync cc host operator
type Watcher struct {
	CliSet        *client.ClientSet
	EtcdCli       *clientv3.Client
	leaseOp       *leaseOp
	ccHostPoolBiz int64
}

// NewWatcher create cc Watcher
func NewWatcher(cliSet *client.ClientSet, etcdCli *clientv3.Client) (Watcher, error) {
	op := &leaseOp{cli: clientv3.NewLease(etcdCli), leaseMap: make(map[string]clientv3.LeaseID)}
	// todo ccHostPoolBiz后续使用cc提供的api获取
	return Watcher{CliSet: cliSet, EtcdCli: etcdCli, leaseOp: op, ccHostPoolBiz: cc.CloudServer().CCHostPoolBiz}, nil
}

// Watch cc event
func (w *Watcher) Watch(sd serviced.ServiceDiscover) {
	go w.WatchHostEvent(sd)
	go w.WatchHostRelationEvent(sd)
}

func getCursorKey(cursorType cmdb.CursorType) string {
	return fmt.Sprintf("/hcm/event/cc/%s", cursorType)
}

func (w *Watcher) getEventCursor(kt *kit.Kit, cursorType cmdb.CursorType) (string, error) {
	key := getCursorKey(cursorType)
	resp, err := w.EtcdCli.Get(kt.Ctx, key)
	if err != nil {
		logs.Errorf("get cmdb event cursor from etcd fail, err: %v, key: %s, rid: %s", err, key, kt.Rid)
		return "", err
	}

	// 从etcd里拿不到cursor，返回空字符串，从当前时间watch
	if len(resp.Kvs) == 0 {
		logs.Warnf("can not get cmdb event cursor from etcd, key: %s, rid: %s", key, kt.Rid)
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

func (w *Watcher) setEventCursor(kt *kit.Kit, cursorType cmdb.CursorType, cursor string) error {
	key := getCursorKey(cursorType)

	leaseID, err := w.leaseOp.getLeaseID(kt, key)
	if err != nil {
		logs.Errorf("get lease id failed, err: %v, key: %s, rid: %s", err, key, kt.Rid)
		return err
	}

	if _, err = w.EtcdCli.Put(kt.Ctx, key, cursor, clientv3.WithLease(leaseID)); err != nil {
		logs.Errorf("set etcd error, err: %v, key: %s, val: %s, rid: %s", err, key, cursor, kt.Rid)
		return err
	}

	return nil
}

type leaseOp struct {
	sync.Mutex
	cli      clientv3.Lease
	leaseMap map[string]clientv3.LeaseID
}

func (l *leaseOp) getLeaseID(kt *kit.Kit, key string) (clientv3.LeaseID, error) {
	l.Lock()
	defer l.Unlock()

	leaseID, ok := l.leaseMap[key]
	var err error
	if ok {
		if _, err = l.cli.KeepAliveOnce(kt.Ctx, leaseID); err != nil {
			logs.Errorf("keep lease alive failed, err: %v, key: %s, leaseID: %v, rid: %s", err, key, leaseID, kt.Rid)
		}
	}

	if !ok || err != nil {
		var seconds int64 = 60 * 60
		leaseResp, err := l.cli.Grant(kt.Ctx, seconds)
		if err != nil {
			logs.Errorf("grant lease failed, err: %v, key: %s, rid: %s", err, key, kt.Rid)
			return 0, err
		}

		l.leaseMap[key] = leaseResp.ID
	}

	return l.leaseMap[key], nil
}
