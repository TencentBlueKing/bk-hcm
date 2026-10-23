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

package tcloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	adaptortcloud "hcm/pkg/adaptor/tcloud"
	typeslb "hcm/pkg/adaptor/types/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	protocloud "hcm/pkg/api/data-service/cloud"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	restcli "hcm/pkg/rest/client"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

const (
	testExclusiveClusterAccountID = "acc-exclusive-cluster"
	testExclusiveClusterRegion    = "ap-nanjing"
)

type fakeDiscover struct {
	host string
}

func (f fakeDiscover) GetServers() ([]string, error) {
	return []string{f.host}, nil
}

type exclusiveClusterCloudStub struct {
	adaptortcloud.TCloud
	clusters []typeslb.TCloudExclusiveCluster
}

func (s exclusiveClusterCloudStub) ListExclusiveClusters(_ *kit.Kit, _ *typeslb.TCloudExclusiveClusterListOption) (
	[]typeslb.TCloudExclusiveCluster, error) {

	return s.clusters, nil
}

type exclusiveClusterDSRecorder struct {
	listDetails []corelb.ExclusiveClusterRaw
	lists       []core.ListReq
	creates     []protocloud.TCloudExclusiveClusterBatchCreateReq
	updates     []protocloud.TCloudExclusiveClusterBatchUpdateReq
	deletes     []protocloud.ExclusiveClusterBatchDeleteReq
}

func TestClient_ExclusiveCluster(t *testing.T) {
	cloudNew := makeExclusiveCloudCluster("tgw-new", "new-cluster", "center_egress1", 10, 3)
	cloudUnchanged := makeExclusiveCloudCluster("tgw-same", "same-cluster", "center_egress1", 8, 2)
	cloudUpdated := makeExclusiveCloudCluster("tgw-upd", "updated-cluster", "center_egress2", 8, 2)
	dbUnchanged := makeExclusiveDBCluster("id-same", "tgw-same", "same-cluster", "center_egress1", 6)
	dbToUpdate := makeExclusiveDBCluster("id-upd", "tgw-upd", "old-cluster", "center_egress1", 6)
	dbToDelete := makeExclusiveDBCluster("id-del", "tgw-del", "deleted-cluster", "center_egress1", 4)

	testCases := []struct {
		name               string
		cloudIDs           []string
		cloud              []typeslb.TCloudExclusiveCluster
		db                 []corelb.ExclusiveClusterRaw
		wantCreateCloudIDs []string
		wantCreateClbCount []int64
		wantUpdateIDs      []string
		wantDeleteCloudIDs []string
	}{
		{
			name:               "create when cloud has and db misses",
			cloudIDs:           []string{"tgw-new"},
			cloud:              []typeslb.TCloudExclusiveCluster{cloudNew},
			wantCreateCloudIDs: []string{"tgw-new"},
			wantCreateClbCount: []int64{7},
		},
		{
			name:          "update when cloud attributes change",
			cloudIDs:      []string{"tgw-upd"},
			cloud:         []typeslb.TCloudExclusiveCluster{cloudUpdated},
			db:            []corelb.ExclusiveClusterRaw{dbToUpdate},
			wantUpdateIDs: []string{"id-upd"},
		},
		{
			name:               "delete when cloud misses and db has",
			cloudIDs:           []string{"tgw-del"},
			db:                 []corelb.ExclusiveClusterRaw{dbToDelete},
			wantDeleteCloudIDs: []string{"tgw-del"},
		},
		{
			name:     "skip write when cloud and db are equal",
			cloudIDs: []string{"tgw-same"},
			cloud:    []typeslb.TCloudExclusiveCluster{cloudUnchanged},
			db:       []corelb.ExclusiveClusterRaw{dbUnchanged},
		},
		{
			name:               "create update and delete in one batch",
			cloudIDs:           []string{"tgw-new", "tgw-upd", "tgw-del"},
			cloud:              []typeslb.TCloudExclusiveCluster{cloudNew, cloudUpdated},
			db:                 []corelb.ExclusiveClusterRaw{dbToUpdate, dbToDelete},
			wantCreateCloudIDs: []string{"tgw-new"},
			wantCreateClbCount: []int64{7},
			wantUpdateIDs:      []string{"id-upd"},
			wantDeleteCloudIDs: []string{"tgw-del"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := &exclusiveClusterDSRecorder{listDetails: tc.db}
			dsCli := newExclusiveClusterDSClient(t, recorder)
			cli := NewClient(dsCli, exclusiveClusterCloudStub{clusters: tc.cloud})

			_, err := cli.ExclusiveCluster(kit.New(), &SyncBaseParams{
				AccountID: testExclusiveClusterAccountID,
				Region:    testExclusiveClusterRegion,
				CloudIDs:  tc.cloudIDs,
			}, new(SyncExclusiveClusterOption))
			require.NoError(t, err)

			gotCreateIDs, gotCreateCounts := collectExclusiveCreate(recorder.creates)
			assert.ElementsMatch(t, tc.wantCreateCloudIDs, gotCreateIDs)
			assert.ElementsMatch(t, tc.wantCreateClbCount, gotCreateCounts)
			assert.ElementsMatch(t, tc.wantUpdateIDs, collectExclusiveUpdateIDs(recorder.updates))
			assert.ElementsMatch(t, tc.wantDeleteCloudIDs, collectExclusiveDeleteCloudIDs(recorder.deletes))

			for _, createReq := range recorder.creates {
				for _, cluster := range createReq.Clusters {
					assert.Equal(t, testExclusiveClusterAccountID, cluster.AccountID)
					assert.Equal(t, testExclusiveClusterRegion, cluster.Region)
					assert.Equal(t, enumor.TGWClusterType, cluster.ClusterType)
					assert.Equal(t, enumor.PublicClusterNetwork, cluster.Network)
				}
			}
		})
	}
}

// TestClient_ExclusiveClusterBatchSplit 校验数量超过单次上限时，新增、更新、删除按100分批，DB查询按500分批。
// 直接调用各步骤的私有方法，ExclusiveCluster入口会先把cloud_ids限制在100以内。
func TestClient_ExclusiveClusterBatchSplit(t *testing.T) {
	const total = 250
	wantBatchSizes := []int{100, 100, 50}

	cloudIDs := make([]string, 0, total)
	newClouds := make([]typeslb.TCloudExclusiveCluster, 0, total)
	changedClouds := make(map[string]typeslb.TCloudExclusiveCluster, total)
	for i := 0; i < total; i++ {
		cloudID := fmt.Sprintf("tgw-%d", i)
		cloudIDs = append(cloudIDs, cloudID)
		newClouds = append(newClouds, makeExclusiveCloudCluster(cloudID, "cluster", "center_egress1", 10, 3))
		changedClouds[fmt.Sprintf("id-%d", i)] = makeExclusiveCloudCluster(cloudID, "cluster", "center_egress2", 10, 3)
	}

	newTestClient := func(t *testing.T) (*client, *exclusiveClusterDSRecorder) {
		recorder := &exclusiveClusterDSRecorder{}
		return &client{dbCli: newExclusiveClusterDSClient(t, recorder), cloudCli: exclusiveClusterCloudStub{}}, recorder
	}

	t.Run("create in batches", func(t *testing.T) {
		cli, recorder := newTestClient(t)

		err := cli.createExclusiveCluster(kit.New(), testExclusiveClusterAccountID, testExclusiveClusterRegion,
			newClouds)
		require.NoError(t, err)

		sizes := make([]int, 0)
		for _, req := range recorder.creates {
			sizes = append(sizes, len(req.Clusters))
		}
		assert.Equal(t, wantBatchSizes, sizes)
		gotIDs, _ := collectExclusiveCreate(recorder.creates)
		assert.ElementsMatch(t, cloudIDs, gotIDs)
	})

	t.Run("update in batches", func(t *testing.T) {
		cli, recorder := newTestClient(t)

		err := cli.updateExclusiveCluster(kit.New(), testExclusiveClusterAccountID, testExclusiveClusterRegion,
			changedClouds)
		require.NoError(t, err)

		sizes := make([]int, 0)
		for _, req := range recorder.updates {
			sizes = append(sizes, len(req.Clusters))
		}
		assert.Equal(t, wantBatchSizes, sizes)
		assert.Len(t, collectExclusiveUpdateIDs(recorder.updates), total)
	})

	t.Run("delete in batches", func(t *testing.T) {
		cli, recorder := newTestClient(t)

		err := cli.deleteExclusiveCluster(kit.New(), testExclusiveClusterAccountID, testExclusiveClusterRegion,
			cloudIDs)
		require.NoError(t, err)

		sizes := make([]int, 0)
		for _, req := range recorder.deletes {
			sizes = append(sizes, len(exclusiveFilterCloudIDs(req.Filter)))
		}
		assert.Equal(t, wantBatchSizes, sizes)
		assert.ElementsMatch(t, cloudIDs, collectExclusiveDeleteCloudIDs(recorder.deletes))
	})

	t.Run("write nothing when empty", func(t *testing.T) {
		cli, recorder := newTestClient(t)
		kt, accountID, region := kit.New(), testExclusiveClusterAccountID, testExclusiveClusterRegion

		require.NoError(t, cli.createExclusiveCluster(kt, accountID, region, nil))
		require.NoError(t, cli.updateExclusiveCluster(kt, accountID, region, nil))
		require.NoError(t, cli.deleteExclusiveCluster(kt, accountID, region, nil))

		assert.Empty(t, recorder.creates)
		assert.Empty(t, recorder.updates)
		assert.Empty(t, recorder.deletes)
	})

	t.Run("list db in batches", func(t *testing.T) {
		cli, recorder := newTestClient(t)

		manyIDs := make([]string, 0, 1200)
		for i := 0; i < 1200; i++ {
			manyIDs = append(manyIDs, fmt.Sprintf("tgw-%d", i))
		}
		_, err := cli.listExclusiveClusterFromDB(kit.New(), &SyncBaseParams{
			AccountID: testExclusiveClusterAccountID,
			Region:    testExclusiveClusterRegion,
			CloudIDs:  manyIDs,
		})
		require.NoError(t, err)

		sizes := make([]int, 0)
		for _, req := range recorder.lists {
			sizes = append(sizes, len(exclusiveFilterCloudIDs(req.Filter)))
			assert.Equal(t, core.DefaultMaxPageLimit, req.Page.Limit)
		}
		assert.Equal(t, []int{500, 500, 200}, sizes)
	})
}

func newExclusiveClusterDSClient(t *testing.T, rec *exclusiveClusterDSRecorder) *dataservice.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/load_balancer_exclusive_clusters/list"):
			var req core.ListReq
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			rec.lists = append(rec.lists, req)
			writeExclusiveClusterOK(w, &core.ListResultT[corelb.ExclusiveClusterRaw]{
				Count:   uint64(len(rec.listDetails)),
				Details: rec.listDetails,
			})
		case r.Method == http.MethodPost &&
			strings.HasSuffix(r.URL.Path, "/load_balancer_exclusive_clusters/batch/create"):
			var req protocloud.TCloudExclusiveClusterBatchCreateReq
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			rec.creates = append(rec.creates, req)
			writeExclusiveClusterOK(w, &core.BatchCreateResult{IDs: []string{"id-created"}})
		case r.Method == http.MethodPatch && strings.HasSuffix(r.URL.Path, "/load_balancer_exclusive_clusters"):
			var req protocloud.TCloudExclusiveClusterBatchUpdateReq
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			rec.updates = append(rec.updates, req)
			writeExclusiveClusterOK(w, struct{}{})
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/load_balancer_exclusive_clusters/batch"):
			var req protocloud.ExclusiveClusterBatchDeleteReq
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			rec.deletes = append(rec.deletes, req)
			writeExclusiveClusterOK(w, struct{}{})
		default:
			http.Error(w, "unexpected path: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	return dataservice.NewClient(&restcli.Capability{
		Client:   http.DefaultClient,
		Discover: fakeDiscover{host: server.URL},
	}, "v1")
}

func writeExclusiveClusterOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":    errf.OK,
		"message": "ok",
		"data":    data,
	})
}

func makeExclusiveCloudCluster(cloudID, name, egress string, resourceCount, idleCount int64) typeslb.TCloudExclusiveCluster {
	return typeslb.TCloudExclusiveCluster{
		Cluster: &tclb.Cluster{
			ClusterId:         cvt.ValToPtr(cloudID),
			ClusterName:       cvt.ValToPtr(name),
			Zone:              cvt.ValToPtr("ap-nanjing-1"),
			ClusterType:       cvt.ValToPtr(string(enumor.TGWClusterType)),
			Network:           cvt.ValToPtr(string(enumor.PublicClusterNetwork)),
			Isp:               cvt.ValToPtr(string(enumor.BGPClusterIsp)),
			Egress:            cvt.ValToPtr(egress),
			IPVersion:         cvt.ValToPtr("ipv4"),
			MaxConn:           cvt.ValToPtr(int64(1000)),
			ResourceCount:     cvt.ValToPtr(resourceCount),
			IdleResourceCount: cvt.ValToPtr(idleCount),
			MaxInFlow:         cvt.ValToPtr(int64(100)),
		},
	}
}

func makeExclusiveDBCluster(id, cloudID, name, egress string, clbCount int64) corelb.ExclusiveClusterRaw {
	ext, _ := json.Marshal(corelb.TCloudExclusiveClusterExtension{MaxInFlow: 100})
	return corelb.ExclusiveClusterRaw{
		BaseExclusiveCluster: corelb.BaseExclusiveCluster{
			ID:               id,
			CloudID:          cloudID,
			Name:             name,
			Vendor:           enumor.TCloud,
			AccountID:        testExclusiveClusterAccountID,
			BkBizID:          -1,
			Region:           testExclusiveClusterRegion,
			Zone:             "ap-nanjing-1",
			ClusterType:      enumor.TGWClusterType,
			Network:          enumor.PublicClusterNetwork,
			Isp:              enumor.BGPClusterIsp,
			Egress:           egress,
			IPVersion:        "ipv4",
			MaxConn:          cvt.ValToPtr(int64(1000)),
			ClbResourceCount: clbCount,
		},
		Extension: ext,
	}
}

func collectExclusiveCreate(reqs []protocloud.TCloudExclusiveClusterBatchCreateReq) ([]string, []int64) {
	cloudIDs := make([]string, 0)
	counts := make([]int64, 0)
	for _, req := range reqs {
		for _, cluster := range req.Clusters {
			cloudIDs = append(cloudIDs, cluster.CloudID)
			counts = append(counts, cluster.ClbResourceCount)
		}
	}
	return cloudIDs, counts
}

func collectExclusiveUpdateIDs(reqs []protocloud.TCloudExclusiveClusterBatchUpdateReq) []string {
	ids := make([]string, 0)
	for _, req := range reqs {
		for _, cluster := range req.Clusters {
			ids = append(ids, cluster.ID)
		}
	}
	return ids
}

func collectExclusiveDeleteCloudIDs(reqs []protocloud.ExclusiveClusterBatchDeleteReq) []string {
	ids := make([]string, 0)
	for _, req := range reqs {
		ids = append(ids, exclusiveFilterCloudIDs(req.Filter)...)
	}
	return ids
}

// exclusiveFilterCloudIDs 取出过滤条件中cloud_id的in列表。
func exclusiveFilterCloudIDs(expr *filter.Expression) []string {
	ids := make([]string, 0)
	if expr == nil {
		return ids
	}
	for _, rule := range expr.Rules {
		atom, ok := rule.(*filter.AtomRule)
		if !ok || atom.Field != "cloud_id" {
			continue
		}
		switch val := atom.Value.(type) {
		case []string:
			ids = append(ids, val...)
		case []interface{}:
			for _, one := range val {
				ids = append(ids, fmt.Sprint(one))
			}
		}
	}
	return ids
}
