package account

import (
	typeaccount "hcm/pkg/adaptor/types/account"
	hssubaccount "hcm/pkg/api/hc-service/sub-account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// TCloudAttachUserPolicies attaches multiple CAM policies to a TCloud sub-user.
func (svc *service) TCloudAttachUserPolicies(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudAttachUserPoliciesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for _, policyID := range req.PolicyIDs {
		if err = client.AttachUserPolicy(cts.Kit, &typeaccount.TCloudAttachUserPolicyOption{TargetUin: req.TargetUin,
			PolicyId: policyID,
		}); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// TCloudDetachUserPolicies detaches multiple CAM policies from a TCloud sub-user.
func (svc *service) TCloudDetachUserPolicies(cts *rest.Contexts) (interface{}, error) {
	req := new(hssubaccount.TCloudDetachUserPoliciesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloud(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for _, policyID := range req.PolicyIDs {
		if err = client.DetachUserPolicy(cts.Kit, &typeaccount.TCloudDetachUserPolicyOption{
			DetachUin: req.DetachUin,
			PolicyId:  policyID,
		}); err != nil {
			return nil, err
		}
	}

	return nil, nil
}
