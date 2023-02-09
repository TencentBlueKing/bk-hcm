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

package gcp

import (
	"hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"google.golang.org/api/compute/v1"
)

// ListCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
func (g *Gcp) ListCvm(kt *kit.Kit, opt *typecvm.GcpListOption) (*compute.InstanceList, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return nil, err
	}

	request := client.Instances.List(g.CloudProjectID(), "").Context(kt.Ctx)

	var call *compute.InstancesListCall
	if opt.Page != nil {
		call = request.MaxResults(opt.Page.PageSize).PageToken(opt.Page.PageToken)
	} else {
		call = request.MaxResults(core.GcpQueryLimit)
	}

	resp, err := call.Do()
	if err != nil {
		logs.Errorf("list instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return nil, err
	}

	return resp, nil
}

// DeleteCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/delete
func (g *Gcp) DeleteCvm(kt *kit.Kit, opt *typecvm.GcpDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.Delete(g.CloudProjectID(), opt.Zone, opt.Name).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("delete instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// StopCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/stop
func (g *Gcp) StopCvm(kt *kit.Kit, opt *typecvm.GcpStopOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "stop option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.Stop(g.CloudProjectID(), opt.Zone, opt.Name).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("stop instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// StartCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/start
func (g *Gcp) StartCvm(kt *kit.Kit, opt *typecvm.GcpStartOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "start option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.Start(g.CloudProjectID(), opt.Zone, opt.Name).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("start instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// ResetCvm reference: https://cloud.google.com/compute/docs/reference/rest/v1/instances/reset
func (g *Gcp) ResetCvm(kt *kit.Kit, opt *typecvm.GcpResetOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "reset option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := g.clientSet.computeClient(kt)
	if err != nil {
		return err
	}

	_, err = client.Instances.Start(g.CloudProjectID(), opt.Zone, opt.Name).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("reset instance failed, err: %v, opt: %v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}
