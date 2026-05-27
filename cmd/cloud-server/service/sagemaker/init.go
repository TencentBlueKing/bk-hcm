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
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

type service struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}

// InitService initializes cloud-server SageMaker passthrough routes.
func InitService(c *capability.Capability) {
	svc := &service{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()
	h.Add("ListNotebookInstancesInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/notebook_instances/list", svc.ListNotebookInstancesInRes)
	h.Add("GetNotebookInstanceInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/notebook_instances/get", svc.GetNotebookInstanceInRes)
	h.Add("ListEndpointsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/endpoints/list", svc.ListEndpointsInRes)
	h.Add("GetEndpointInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/endpoints/get", svc.GetEndpointInRes)
	h.Add("ListEndpointConfigsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/endpoint_configs/list", svc.ListEndpointConfigsInRes)
	h.Add("GetEndpointConfigInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/endpoint_configs/get", svc.GetEndpointConfigInRes)
	h.Add("ListTrainingJobsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/training_jobs/list", svc.ListTrainingJobsInRes)
	h.Add("GetTrainingJobInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/training_jobs/get", svc.GetTrainingJobInRes)
	h.Add("ListProcessingJobsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/processing_jobs/list", svc.ListProcessingJobsInRes)
	h.Add("GetProcessingJobInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/processing_jobs/get", svc.GetProcessingJobInRes)
	h.Add("ListTransformJobsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/transform_jobs/list", svc.ListTransformJobsInRes)
	h.Add("GetTransformJobInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/transform_jobs/get", svc.GetTransformJobInRes)

	h.Add("ListAppsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/apps/list", svc.ListAppsInRes)
	h.Add("GetAppInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/apps/get", svc.GetAppInRes)
	h.Add("ListClustersInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/clusters/list", svc.ListClustersInRes)
	h.Add("GetClusterInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/clusters/get", svc.GetClusterInRes)
	h.Add("ListClusterNodesInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/cluster_nodes/list", svc.ListClusterNodesInRes)
	h.Add("GetClusterNodeInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/cluster_nodes/get", svc.GetClusterNodeInRes)

	h.Add("ListTrainingPlansInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/training_plans/list", svc.ListTrainingPlansInRes)
	h.Add("GetTrainingPlanInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/training_plans/get", svc.GetTrainingPlanInRes)
	h.Add("SearchTrainingPlanOfferingsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/training_plan_offerings/search", svc.SearchTrainingPlanOfferingsInRes)
	h.Add("ListInferenceComponentsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/inference_components/list", svc.ListInferenceComponentsInRes)
	h.Add("GetInferenceComponentInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/inference_components/get", svc.GetInferenceComponentInRes)
	h.Add("ListOptimizationJobsInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/optimization_jobs/list", svc.ListOptimizationJobsInRes)
	h.Add("GetOptimizationJobInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/optimization_jobs/get", svc.GetOptimizationJobInRes)
	h.Add("ListComputeQuotasInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/compute_quotas/list", svc.ListComputeQuotasInRes)
	h.Add("GetComputeQuotaInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/compute_quotas/get", svc.GetComputeQuotaInRes)
	h.Add("GetReservedCapacityInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/reserved_capacities/get", svc.GetReservedCapacityInRes)
	h.Add("ListUltraServersByReservedCapacityInRes", http.MethodPost,
		"/vendors/aws/assume_role/sagemaker/reserved_capacity_ultra_servers/list",
		svc.ListUltraServersByReservedCapacityInRes)

	h.Load(c.WebService)
}
