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
	"fmt"

	cloudclient "hcm/cmd/hc-service/logics/cloud-adaptor"
	"hcm/cmd/hc-service/service/capability"
	adaptoraws "hcm/pkg/adaptor/aws"
	datacli "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

type assumeRoleReq interface {
	Validate() error
	GetRootAccountID() string
	GetMainAccountID() string
	GetRoleChain() []string
	GetExternalID() string
	GetRegion() string
}

type sageMakerSvc struct {
	adaptor *cloudclient.CloudAdaptorClient
	dataCli *datacli.Client
}

// InitService initializes SageMaker assume-role passthrough handlers.
func InitService(cap *capability.Capability) {
	svc := &sageMakerSvc{
		adaptor: cap.CloudAdaptor,
		dataCli: cap.ClientSet.DataService(),
	}

	h := rest.NewHandler()
	h.Add("ListNotebookInstancesForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/notebook_instances/list", svc.ListNotebookInstancesForAws)
	h.Add("GetNotebookInstanceForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/notebook_instances/get", svc.GetNotebookInstanceForAws)
	h.Add("ListEndpointsForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/endpoints/list", svc.ListEndpointsForAws)
	h.Add("GetEndpointForAws", "POST", "/vendors/aws/assume_role/sagemaker/endpoints/get", svc.GetEndpointForAws)
	h.Add("ListEndpointConfigsForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/endpoint_configs/list", svc.ListEndpointConfigsForAws)
	h.Add("GetEndpointConfigForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/endpoint_configs/get", svc.GetEndpointConfigForAws)
	h.Add("ListTrainingJobsForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/training_jobs/list", svc.ListTrainingJobsForAws)
	h.Add("GetTrainingJobForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/training_jobs/get", svc.GetTrainingJobForAws)
	h.Add("ListProcessingJobsForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/processing_jobs/list", svc.ListProcessingJobsForAws)
	h.Add("GetProcessingJobForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/processing_jobs/get", svc.GetProcessingJobForAws)
	h.Add("ListTransformJobsForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/transform_jobs/list", svc.ListTransformJobsForAws)
	h.Add("GetTransformJobForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/transform_jobs/get", svc.GetTransformJobForAws)

	h.Add("ListAppsForAws", "POST", "/vendors/aws/assume_role/sagemaker/apps/list", svc.ListAppsForAws)
	h.Add("GetAppForAws", "POST", "/vendors/aws/assume_role/sagemaker/apps/get", svc.GetAppForAws)
	h.Add("ListClustersForAws", "POST", "/vendors/aws/assume_role/sagemaker/clusters/list", svc.ListClustersForAws)
	h.Add("GetClusterForAws", "POST", "/vendors/aws/assume_role/sagemaker/clusters/get", svc.GetClusterForAws)
	h.Add("ListClusterNodesForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/cluster_nodes/list", svc.ListClusterNodesForAws)
	h.Add("GetClusterNodeForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/cluster_nodes/get", svc.GetClusterNodeForAws)

	h.Add("ListTrainingPlansForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/training_plans/list", svc.ListTrainingPlansForAws)
	h.Add("GetTrainingPlanForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/training_plans/get", svc.GetTrainingPlanForAws)
	h.Add("SearchTrainingPlanOfferingsForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/training_plan_offerings/search", svc.SearchTrainingPlanOfferingsForAws)
	h.Add("ListInferenceComponentsForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/inference_components/list", svc.ListInferenceComponentsForAws)
	h.Add("GetInferenceComponentForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/inference_components/get", svc.GetInferenceComponentForAws)
	h.Add("ListOptimizationJobsForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/optimization_jobs/list", svc.ListOptimizationJobsForAws)
	h.Add("GetOptimizationJobForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/optimization_jobs/get", svc.GetOptimizationJobForAws)
	h.Add("ListComputeQuotasForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/compute_quotas/list", svc.ListComputeQuotasForAws)
	h.Add("GetComputeQuotaForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/compute_quotas/get", svc.GetComputeQuotaForAws)
	h.Add("GetReservedCapacityForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/reserved_capacities/get", svc.GetReservedCapacityForAws)
	h.Add("ListUltraServersByReservedCapacityForAws", "POST",
		"/vendors/aws/assume_role/sagemaker/reserved_capacity_ultra_servers/list",
		svc.ListUltraServersByReservedCapacityForAws)
	h.Load(cap.WebService)
}

func (s *sageMakerSvc) getCloudIDFromMainAccount(kt *kit.Kit, mainAccountID, rootAccountID string) (string, error) {
	mainAccountInfo, err := s.dataCli.Aws.MainAccount.Get(kt, mainAccountID)
	if err != nil {
		logs.Errorf("get aws main account failed, main account id: %s, err: %v, rid: %s", mainAccountID, err, kt.Rid)
		return "", err
	}
	if mainAccountInfo.ParentAccountID != rootAccountID {
		logs.Errorf("main account %s does not belong to root account %s, actual parent: %s, rid: %s",
			mainAccountID, rootAccountID, mainAccountInfo.ParentAccountID, kt.Rid)
		return "", fmt.Errorf("main account '%s' does not belong to root account '%s'", mainAccountID, rootAccountID)
	}
	if mainAccountInfo.Extension == nil || mainAccountInfo.Extension.CloudMainAccountID == "" {
		logs.Errorf("main account: %s cloud main account id is empty, rid: %s", mainAccountID, kt.Rid)
		return "", fmt.Errorf("main account: %s cloud main account id is empty", mainAccountID)
	}
	return mainAccountInfo.Extension.CloudMainAccountID, nil
}

func (s *sageMakerSvc) assumeRoleClient(kt *kit.Kit, req assumeRoleReq) (*adaptoraws.Aws, error) {
	cloudID, err := s.getCloudIDFromMainAccount(kt, req.GetMainAccountID(), req.GetRootAccountID())
	if err != nil {
		return nil, err
	}
	return s.adaptor.AwsWithAssumeRole(kt, req.GetRootAccountID(), cloudID, req.GetRoleChain(), req.GetExternalID())
}
