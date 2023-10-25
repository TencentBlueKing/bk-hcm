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

package azure

import (
	"hcm/cmd/hc-service/logics/res-sync/azure"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// AzurePublisherAndOffer ...
type AzurePublisherAndOffer struct {
	Publisher string
	Offer     string
}

// SyncImage ....
func (svc *service) SyncImage(cts *rest.Contexts) (interface{}, error) {
	req := new(sync.AzureImageReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	syncCli, err := svc.syncCli.Azure(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	filters := make([]AzurePublisherAndOffer, 0)
	filters = append(filters, AzurePublisherAndOffer{Publisher: "openlogic", Offer: "CentOS"})
	filters = append(filters, AzurePublisherAndOffer{Publisher: "microsoftwindowsserver", Offer: "WindowsServer"})

	for _, one := range filters {
		if _, err := syncCli.Image(cts.Kit, &azure.SyncImageOption{AccountID: req.AccountID,
			Region: req.Region, Publisher: one.Publisher, Offer: one.Offer}); err != nil {
			logs.Errorf("sync azure image failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}
