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

package sagemaker

import (
	proto "hcm/pkg/api/hc-service/sagemaker"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListNotebookInstancesInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListNotebookInstancesInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListNotebookInstancesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListNotebookInstances(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker notebook instances failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetNotebookInstanceInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetNotebookInstanceInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeNotebookInstanceReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetNotebookInstance(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker notebook instance failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListEndpointsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListEndpointsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListEndpointsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListEndpoints(cts.Kit, req)
	if err != nil {
		logs.Errorf("call hc-service to list aws assume role sagemaker endpoints failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return data, nil
}

// GetEndpointInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetEndpointInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeEndpointReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetEndpoint(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker endpoint failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListEndpointConfigsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListEndpointConfigsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListEndpointConfigsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListEndpointConfigs(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker endpoint configs failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetEndpointConfigInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetEndpointConfigInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeEndpointConfigReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetEndpointConfig(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker endpoint config failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListTrainingJobsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListTrainingJobsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListTrainingJobsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListTrainingJobs(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker training jobs failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetTrainingJobInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetTrainingJobInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeTrainingJobReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetTrainingJob(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker training job failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListProcessingJobsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListProcessingJobsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListProcessingJobsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListProcessingJobs(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker processing jobs failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetProcessingJobInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetProcessingJobInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeProcessingJobReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetProcessingJob(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker processing job failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListTransformJobsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListTransformJobsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListTransformJobsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListTransformJobs(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker transform jobs failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetTransformJobInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetTransformJobInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeTransformJobReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetTransformJob(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker transform job failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListAppsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListAppsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListAppsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListApps(cts.Kit, req)
	if err != nil {
		logs.Errorf("call hc-service to list aws assume role sagemaker apps failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return data, nil
}

// GetAppInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetAppInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeAppReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetApp(cts.Kit, req)
	if err != nil {
		logs.Errorf("call hc-service to describe aws assume role sagemaker app failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return data, nil
}

// ListClustersInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListClustersInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListClustersReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListClusters(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker hyperpod clusters failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetClusterInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetClusterInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeClusterReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetCluster(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker hyperpod cluster failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListClusterNodesInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListClusterNodesInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListClusterNodesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListClusterNodes(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker hyperpod cluster nodes failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetClusterNodeInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetClusterNodeInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeClusterNodeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetClusterNode(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker hyperpod cluster node failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListTrainingPlansInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListTrainingPlansInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListTrainingPlansReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListTrainingPlans(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker training plans failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetTrainingPlanInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetTrainingPlanInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeTrainingPlanReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetTrainingPlan(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker training plan failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// SearchTrainingPlanOfferingsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) SearchTrainingPlanOfferingsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerSearchTrainingPlanOfferingsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.SearchTrainingPlanOfferings(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to search aws assume role sagemaker training plan offerings failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListInferenceComponentsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListInferenceComponentsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListInferenceComponentsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListInferenceComponents(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker inference components failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetInferenceComponentInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetInferenceComponentInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeInferenceComponentReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetInferenceComponent(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker inference component failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListOptimizationJobsInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListOptimizationJobsInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListOptimizationJobsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListOptimizationJobs(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker optimization jobs failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetOptimizationJobInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetOptimizationJobInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeOptimizationJobReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetOptimizationJob(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker optimization job failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListComputeQuotasInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListComputeQuotasInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListComputeQuotasReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListComputeQuotas(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker compute quotas failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetComputeQuotaInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetComputeQuotaInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeComputeQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetComputeQuota(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker compute quota failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// GetReservedCapacityInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) GetReservedCapacityInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeReservedCapacityReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.GetReservedCapacity(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to describe aws assume role sagemaker reserved capacity failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

// ListUltraServersByReservedCapacityInRes handles the cloud-server SageMaker assume-role passthrough request.
func (svc *service) ListUltraServersByReservedCapacityInRes(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListUltraServersByReservedCapacityReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if err := svc.authRootAccount(cts, req.RootAccountID); err != nil {
		return nil, err
	}
	data, err := svc.client.HCService().Aws.SageMaker.ListUltraServersByReservedCapacity(cts.Kit, req)
	if err != nil {
		logs.Errorf(
			"call hc-service to list aws assume role sagemaker reserved capacity ultra servers failed, err: %v, rid: %s",
			err, cts.Kit.Rid,
		)
		return nil, err
	}
	return data, nil
}

func (svc *service) authRootAccount(cts *rest.Contexts, accountID string) error {
	err := handler.ResOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer:        svc.authorizer,
		ResType:           meta.CloudResource,
		Action:            meta.Find,
		DisableBizIDEqual: true,
		BasicInfo:         &types.CloudResourceBasicInfo{AccountID: accountID},
	})
	if err != nil {
		logs.Errorf(
			"auth aws assume role sagemaker root account failed, accountID: %s, err: %v, rid: %s",
			accountID, err, cts.Kit.Rid,
		)
		return err
	}
	return nil
}
