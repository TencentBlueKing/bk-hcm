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
	"fmt"
	"strings"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types/eip"
	"hcm/pkg/criteria/enumor"
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
			Name:                publicIp.Alias,
			CloudID:             *publicIp.Id,
			Region:              opt.Region,
			Status:              &status,
			PublicIp:            publicIp.PublicIpAddress,
			PrivateIp:           publicIp.PrivateIpAddress,
			PortID:              publicIp.PortId,
			BandwidthId:         publicIp.BandwidthId,
			BandwidthName:       publicIp.BandwidthName,
			BandwidthSize:       publicIp.BandwidthSize,
			EnterpriseProjectId: publicIp.EnterpriseProjectId,
			Type:                publicIp.Type,
			BandwidthShareType:  "",
			ChargeMode:          "",
		}
		if publicIp.BandwidthShareType != nil {
			eips[idx].BandwidthShareType = publicIp.BandwidthShareType.Value()
		}
		if publicIp.BandwidthId != nil {
			request := &model.ShowBandwidthRequest{}
			request.BandwidthId = converter.PtrToVal(publicIp.BandwidthId)
			response, err := client.ShowBandwidth(request)
			if err != nil {
				logs.Errorf("[%s] fail to ShowBandwidth, err: %v, BandwidthId: %s, rid: %s",
					enumor.HuaWei, err, request.BandwidthId, kt.Rid)
				return nil, err
			}
			if response.Bandwidth != nil {
				if response.Bandwidth.ChargeMode != nil {
					eips[idx].ChargeMode = response.Bandwidth.ChargeMode.Value()
				}
			}
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

	respPoller := poller.Poller[*HuaWei, []*eip.HuaWeiEip,
		poller.BaseDoneResult]{Handler: &associateEipPollingHandler{region: opt.Region}}
	_, err = respPoller.PollUntilDone(h, kt, []*string{&opt.CloudEipID}, nil)
	if err != nil {
		return err
	}

	return nil
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

	respPoller := poller.Poller[*HuaWei, []*eip.HuaWeiEip,
		poller.BaseDoneResult]{Handler: &disassociateEipPollingHandler{region: opt.Region}}
	_, err = respPoller.PollUntilDone(h, kt, []*string{&opt.CloudEipID}, nil)
	if err != nil {
		return err
	}

	return nil
}

// CreateEip ...
// reference: https://support.huaweicloud.com/api-eip/eip_api_0001.html
// https://support.huaweicloud.com/api-eip/eip_api_0006.html
func (h *HuaWei) CreateEip(kt *kit.Kit, opt *eip.HuaWeiEipCreateOption) (*poller.BaseDoneResult, error) {
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

		resp, err := client.CreatePrePaidPublicip(req)
		if err != nil {
			return nil, err
		}

		respPoller := poller.Poller[*HuaWei, []*eip.HuaWeiEip,
			poller.BaseDoneResult]{Handler: &createEipPollingHandler{region: opt.Region}}
		return respPoller.PollUntilDone(h, kt, []*string{resp.PublicipId}, nil)
	}
	// 按需计费
	req, err := opt.ToCreatePublicipRequest()
	if err != nil {
		return nil, err
	}

	resp, err := client.CreatePublicip(req)
	if err != nil {
		return nil, err
	}
	respPoller := poller.Poller[*HuaWei, []*eip.HuaWeiEip,
		poller.BaseDoneResult]{Handler: &createEipPollingHandler{region: opt.Region}}
	return respPoller.PollUntilDone(h, kt, []*string{resp.Publicip.Id}, nil)
}

type createEipPollingHandler struct {
	region string
}

// Done ...
func (h *createEipPollingHandler) Done(pollResult []*eip.HuaWeiEip) (bool, *poller.BaseDoneResult) {
	if len(pollResult) == 0 {
		return false, nil
	}

	successCloudIDs := make([]string, 0)
	unknownCloudIDs := make([]string, 0)

	for _, r := range pollResult {
		if converter.PtrToVal(r.Status) == "PENDING_CREATE" ||
			converter.PtrToVal(r.Status) == "NOTIFYING" {
			unknownCloudIDs = append(unknownCloudIDs, r.CloudID)
		} else {
			successCloudIDs = append(successCloudIDs, r.CloudID)
		}
	}

	isDone := false
	if len(successCloudIDs) != 0 && len(successCloudIDs) == len(pollResult) {
		isDone = true
	}

	return isDone, &poller.BaseDoneResult{
		SuccessCloudIDs: successCloudIDs,
		UnknownCloudIDs: unknownCloudIDs,
	}
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

var _ poller.PollingHandler[*HuaWei, []*eip.HuaWeiEip, poller.BaseDoneResult] = new(createEipPollingHandler)

type associateEipPollingHandler struct {
	region string
}

// Done ...
func (h *associateEipPollingHandler) Done(pollResult []*eip.HuaWeiEip) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]

	if converter.PtrToVal(r.Status) != "ACTIVE" {
		return false, nil
	}
	return true, nil
}

// Poll ...
func (h *associateEipPollingHandler) Poll(client *HuaWei, kt *kit.Kit, cloudIDs []*string) ([]*eip.HuaWeiEip, error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

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
func (h *disassociateEipPollingHandler) Done(pollResult []*eip.HuaWeiEip) (bool, *poller.BaseDoneResult) {
	r := pollResult[0]

	if converter.PtrToVal(r.Status) != "DOWN" {
		return false, nil
	}
	return true, nil
}

// Poll ...
func (h *disassociateEipPollingHandler) Poll(
	client *HuaWei,
	kt *kit.Kit,
	cloudIDs []*string,
) ([]*eip.HuaWeiEip, error) {
	if len(cloudIDs) != 1 {
		return nil, fmt.Errorf("poll only support one id param, but get %v. rid: %s", cloudIDs, kt.Rid)
	}

	cIDs := converter.PtrToSlice(cloudIDs)
	result, err := client.ListEip(kt, &eip.HuaWeiEipListOption{Region: h.region, CloudIDs: cIDs})
	if err != nil {
		return nil, err
	}

	return result.Details, nil
}
