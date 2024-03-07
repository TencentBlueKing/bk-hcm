package clb

import (
	"encoding/json"
	"fmt"

	cloudserver "hcm/pkg/api/cloud-server"
	protoclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

func (svc *clbSvc) BatchCreateCLB(cts *rest.Contexts) (any, error) {

	req := new(cloudserver.ResourceCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		logs.Errorf("create clb request decode failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	// 权限校验
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{
		Type:       meta.Clb,
		Action:     meta.Create,
		ResourceID: req.AccountID,
	}}
	if err := svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes); err != nil {
		logs.Errorf("create load balancer auth failed, err: %v, account id: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	accountInfo, err := svc.client.DataService().Global.Cloud.
		GetResBasicInfo(cts.Kit, enumor.AccountCloudResType, req.AccountID)
	if err != nil {
		logs.Errorf("get account basic info failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch accountInfo.Vendor {
	case enumor.TCloud:
		return svc.batchCreateTCloudCLB(cts.Kit, req.Data)
	default:
		return nil, fmt.Errorf("vendor: %s not support", accountInfo.Vendor)
	}

}

func (svc *clbSvc) batchCreateTCloudCLB(kt *kit.Kit, rawReq json.RawMessage) (any, error) {
	req := new(protoclb.TCloudBatchCreateReq)
	if err := json.Unmarshal(rawReq, req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	// 参数校验
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.client.HCService().TCloud.Clb.BatchCreate(kt, req)
}
