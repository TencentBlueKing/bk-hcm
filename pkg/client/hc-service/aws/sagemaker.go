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

package aws

import (
	stdjson "encoding/json"

	proto "hcm/pkg/api/hc-service/sagemaker"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// SageMakerClient is hc-service aws SageMaker passthrough client.
type SageMakerClient struct {
	client rest.ClientInterface
}

// NewSageMakerClient creates a new SageMaker client.
func NewSageMakerClient(client rest.ClientInterface) *SageMakerClient {
	return &SageMakerClient{client: client}
}

// ListNotebookInstances lists notebook instances via AssumeRole cross-account access.
func (c *SageMakerClient) ListNotebookInstances(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListNotebookInstancesReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/notebook_instances/list")
}

// GetNotebookInstance gets a notebook instance via AssumeRole cross-account access.
func (c *SageMakerClient) GetNotebookInstance(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeNotebookInstanceReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/notebook_instances/get")
}

// ListEndpoints lists endpoints via AssumeRole cross-account access.
func (c *SageMakerClient) ListEndpoints(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListEndpointsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/endpoints/list")
}

// GetEndpoint gets an endpoint via AssumeRole cross-account access.
func (c *SageMakerClient) GetEndpoint(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeEndpointReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/endpoints/get")
}

// ListEndpointConfigs lists endpoint configs via AssumeRole cross-account access.
func (c *SageMakerClient) ListEndpointConfigs(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListEndpointConfigsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/endpoint_configs/list")
}

// GetEndpointConfig gets an endpoint config via AssumeRole cross-account access.
func (c *SageMakerClient) GetEndpointConfig(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeEndpointConfigReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/endpoint_configs/get")
}

// ListTrainingJobs lists training jobs via AssumeRole cross-account access.
func (c *SageMakerClient) ListTrainingJobs(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListTrainingJobsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/training_jobs/list")
}

// GetTrainingJob gets a training job via AssumeRole cross-account access.
func (c *SageMakerClient) GetTrainingJob(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeTrainingJobReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/training_jobs/get")
}

// ListProcessingJobs lists processing jobs via AssumeRole cross-account access.
func (c *SageMakerClient) ListProcessingJobs(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListProcessingJobsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/processing_jobs/list")
}

// GetProcessingJob gets a processing job via AssumeRole cross-account access.
func (c *SageMakerClient) GetProcessingJob(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeProcessingJobReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/processing_jobs/get")
}

// ListTransformJobs lists transform jobs via AssumeRole cross-account access.
func (c *SageMakerClient) ListTransformJobs(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListTransformJobsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/transform_jobs/list")
}

// GetTransformJob gets a transform job via AssumeRole cross-account access.
func (c *SageMakerClient) GetTransformJob(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeTransformJobReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/transform_jobs/get")
}

// ListApps lists Studio apps via AssumeRole cross-account access.
func (c *SageMakerClient) ListApps(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListAppsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/apps/list")
}

// GetApp gets a Studio app via AssumeRole cross-account access.
func (c *SageMakerClient) GetApp(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeAppReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/apps/get")
}

// ListClusters lists HyperPod clusters via AssumeRole cross-account access.
func (c *SageMakerClient) ListClusters(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListClustersReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/clusters/list")
}

// GetCluster gets a HyperPod cluster via AssumeRole cross-account access.
func (c *SageMakerClient) GetCluster(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeClusterReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/clusters/get")
}

// ListClusterNodes lists HyperPod cluster nodes via AssumeRole cross-account access.
func (c *SageMakerClient) ListClusterNodes(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListClusterNodesReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/cluster_nodes/list")
}

// GetClusterNode gets a HyperPod cluster node via AssumeRole cross-account access.
func (c *SageMakerClient) GetClusterNode(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeClusterNodeReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/cluster_nodes/get")
}

// ListTrainingPlans lists Training Plans via AssumeRole cross-account access.
func (c *SageMakerClient) ListTrainingPlans(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListTrainingPlansReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/training_plans/list")
}

// GetTrainingPlan gets a Training Plan via AssumeRole cross-account access.
func (c *SageMakerClient) GetTrainingPlan(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeTrainingPlanReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/training_plans/get")
}

// SearchTrainingPlanOfferings searches Training Plan offerings via AssumeRole cross-account access.
func (c *SageMakerClient) SearchTrainingPlanOfferings(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerSearchTrainingPlanOfferingsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/training_plan_offerings/search")
}

// ListInferenceComponents lists inference components via AssumeRole cross-account access.
func (c *SageMakerClient) ListInferenceComponents(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListInferenceComponentsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/inference_components/list")
}

// GetInferenceComponent gets an inference component via AssumeRole cross-account access.
func (c *SageMakerClient) GetInferenceComponent(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeInferenceComponentReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/inference_components/get")
}

// ListOptimizationJobs lists optimization jobs via AssumeRole cross-account access.
func (c *SageMakerClient) ListOptimizationJobs(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListOptimizationJobsReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/optimization_jobs/list")
}

// GetOptimizationJob gets an optimization job via AssumeRole cross-account access.
func (c *SageMakerClient) GetOptimizationJob(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeOptimizationJobReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/optimization_jobs/get")
}

// ListComputeQuotas lists compute quotas via AssumeRole cross-account access.
func (c *SageMakerClient) ListComputeQuotas(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListComputeQuotasReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/compute_quotas/list")
}

// GetComputeQuota gets a compute quota via AssumeRole cross-account access.
func (c *SageMakerClient) GetComputeQuota(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeComputeQuotaReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/compute_quotas/get")
}

// GetReservedCapacity gets a reserved capacity via AssumeRole cross-account access.
func (c *SageMakerClient) GetReservedCapacity(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerDescribeReservedCapacityReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/reserved_capacities/get")
}

// ListUltraServersByReservedCapacity lists reserved capacity UltraServers via AssumeRole cross-account access.
func (c *SageMakerClient) ListUltraServersByReservedCapacity(kt *kit.Kit,
	request *proto.AwsAssumeRoleSageMakerListUltraServersByReservedCapacityReq) (stdjson.RawMessage, error) {

	return requestRaw(c.client, kt, request, "/assume_role/sagemaker/reserved_capacity_ultra_servers/list")
}

// requestRaw sends a passthrough POST request and returns the raw response body.
func requestRaw[T any](cli rest.ClientInterface, kt *kit.Kit, req *T, path string) (stdjson.RawMessage, error) {
	resp, err := common.Request[T, stdjson.RawMessage](cli, rest.POST, kt, req, path)
	if err != nil {
		return nil, err
	}
	return *resp, nil
}
