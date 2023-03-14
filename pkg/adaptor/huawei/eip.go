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

package huawei

import (
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eip/v2/model"
)

// ListEip ...
// reference: https://support.huaweicloud.com/api-eip/eip_api_0003.html
func (h *HuaWei) ListEip(kt *kit.Kit, opt *eip.HuaWeiEipListOption) (*eip.HuaWeiEipListResult, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := h.clientSet.eipClient(opt.Region)
	if err != nil {
		return nil, err
	}

	req := new(model.ListPublicipsRequest)

	if len(opt.CloudIDs) > 0 {
		req.Id = &opt.CloudIDs
	}

	if len(opt.Ips) > 0 {
		req.PublicIpAddress = &opt.Ips
	}

	if opt.Limit != nil {
		req.Limit = opt.Limit
	}

	if opt.Marker != nil {
		req.Marker = opt.Marker
	}

	resp, err := client.ListPublicips(req)
	if err != nil {
		if strings.Contains(err.Error(), ErrDataNotFound) {
			return new(eip.HuaWeiEipListResult), nil
		}
		logs.Errorf("list huawei eip failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	eips := make([]*eip.HuaWeiEip, len(*resp.Publicips))
	for idx, publicIp := range *resp.Publicips {
		status := publicIp.Status.Value()
		eips[idx] = &eip.HuaWeiEip{
			CloudID:       *publicIp.Id,
			Region:        opt.Region,
			Status:        &status,
			PublicIp:      publicIp.PublicIpAddress,
			PrivateIp:     publicIp.PrivateIpAddress,
			PortID:        publicIp.PortId,
			BandwidthId:   publicIp.BandwidthId,
			BandwidthName: publicIp.BandwidthName,
			BandwidthSize: publicIp.BandwidthSize,
		}
	}

	return &eip.HuaWeiEipListResult{Details: eips}, nil
}

// DeleteEip ...
// reference: https://support.huaweicloud.com/api-eip/eip_api_0005.html
func (h *HuaWei) DeleteEip(kt *kit.Kit, opt *eip.HuaWeiEipDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "huawei eip delete option is required")
	}

	req, err := opt.ToDeletePublicipRequest()
	if err != nil {
		return err
	}

	client, err := h.clientSet.eipClient(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.DeletePublicip(req)
	if err != nil {
		logs.Errorf("delete huawei eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// AssociateEip ...
// reference: https://support.huaweicloud.com/api-eip/eip_api_0004.html
func (h *HuaWei) AssociateEip(kt *kit.Kit, opt *eip.HuaWeiEipAssociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "huawei eip associate option is required")
	}

	req, err := opt.ToUpdatePublicipRequest()
	if err != nil {
		return err
	}

	client, err := h.clientSet.eipClient(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.UpdatePublicip(req)
	if err != nil {
		logs.Errorf("associate huawei eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*HuaWei, []*eip.HuaWeiEip]{Handler: &associateEipPollingHandler{region: opt.Region}}
	return respPoller.PollUntilDone(h, kt, []*string{&opt.CloudEipID})
}

// DisassociateEip ...
// reference: https://support.huaweicloud.com/api-eip/eip_api_0004.html
func (h *HuaWei) DisassociateEip(kt *kit.Kit, opt *eip.HuaWeiEipDisassociateOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "huawei eip disassociate option is required")
	}

	req, err := opt.ToUpdatePublicipRequest()
	if err != nil {
		return err
	}

	client, err := h.clientSet.eipClient(opt.Region)
	if err != nil {
		return err
	}

	_, err = client.UpdatePublicip(req)
	if err != nil {
		logs.Errorf("disassociate huawei eip failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	respPoller := poller.Poller[*HuaWei, []*eip.HuaWeiEip]{Handler: &disassociateEipPollingHandler{region: opt.Region}}
	return respPoller.PollUntilDone(h, kt, []*string{&opt.CloudEipID})
}

// CreateEip ...
// reference: https://support.huaweicloud.com/api-eip/eip_api_0001.html
// https://support.huaweicloud.com/api-eip/eip_api_0006.html
func (h *HuaWei) CreateEip(kt *kit.Kit, opt *eip.HuaWeiEipCreateOption) (*string, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "huawei eip create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := h.clientSet.eipClient(opt.Region)
	if err != nil {
		return nil, err
	}

	// 包年包月
	if opt.InternetChargeType == "prePaid" {
		req, err := opt.ToCreatePrePaidPublicipRequest()
		if err != nil {
			return nil, err
		}

		response, err := client.CreatePrePaidPublicip(req)
		return response.Publicip.Id, err
	}
	// 按需计费
	req, err := opt.ToCreatePublicipRequest()
	if err != nil {
		return nil, err
	}

	resp, err := client.CreatePublicip(req)

	respPoller := poller.Poller[*HuaWei, []*eip.HuaWeiEip]{Handler: &createEipPollingHandler{region: opt.Region}}
	err = respPoller.PollUntilDone(h, kt, []*string{resp.Publicip.Id})
	if err != nil {
		return nil, err
	}

	return resp.Publicip.Id, err
}

type createEipPollingHandler struct {
	region string
}

// Done ...
func (h *createEipPollingHandler) Done(pollResult []*eip.HuaWeiEip) bool {
	for _, r := range pollResult {
		if r.Status == nil || *r.Status == "PENDING_CREATE" || *r.Status == "NOTIFYING" {
			return false
		}
	}
	return true
}

// Poll ...
func (h *createEipPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]*eip.HuaWeiEip, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.HuaWeiEipListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

var _ poller.PollingHandler[*HuaWei, []*eip.HuaWeiEip] = new(createEipPollingHandler)

type associateEipPollingHandler struct {
	region string
}

// Done ...
func (h *associateEipPollingHandler) Done(pollResult []*eip.HuaWeiEip) bool {
	for _, r := range pollResult {
		if *r.Status != "ACTIVE" {
			return false
		}
	}
	return true
}

// Poll ...
func (h *associateEipPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]*eip.HuaWeiEip, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.HuaWeiEipListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}

type disassociateEipPollingHandler struct {
	region string
}

// Done ...
func (h *disassociateEipPollingHandler) Done(pollResult []*eip.HuaWeiEip) bool {
	for _, r := range pollResult {
		if *r.Status != "DOWN" {
			return false
		}
	}
	return true
}

// Poll ...
func (h *disassociateEipPollingHandler) Poll(
	client *HuaWei,
	kt *kit.Kit,
	cloudIDs []*string,
) ([]*eip.HuaWeiEip, error) {
	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.HuaWeiEipListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}
