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
	"time"

	"hcm/pkg/criteria/validator"
)

// AwsAssumeRoleSageMakerBaseReq defines the shared assume-role request fields for SageMaker passthrough APIs.
type AwsAssumeRoleSageMakerBaseReq struct {
	RootAccountID string   `json:"root_account_id" validate:"required"`
	MainAccountID string   `json:"main_account_id" validate:"required"`
	RoleChain     []string `json:"role_chain" validate:"required,min=1"`
	Region        string   `json:"region" validate:"required"`
	ExternalID    string   `json:"external_id,omitempty"`
}

// Validate validates the common assume-role fields.
func (req *AwsAssumeRoleSageMakerBaseReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GetRootAccountID returns the root account id.
func (req *AwsAssumeRoleSageMakerBaseReq) GetRootAccountID() string {
	return req.RootAccountID
}

// GetMainAccountID returns the main account id.
func (req *AwsAssumeRoleSageMakerBaseReq) GetMainAccountID() string {
	return req.MainAccountID
}

// GetRoleChain returns the role chain.
func (req *AwsAssumeRoleSageMakerBaseReq) GetRoleChain() []string {
	return req.RoleChain
}

// GetExternalID returns the optional external id.
func (req *AwsAssumeRoleSageMakerBaseReq) GetExternalID() string {
	return req.ExternalID
}

// GetRegion returns the target region.
func (req *AwsAssumeRoleSageMakerBaseReq) GetRegion() string {
	return req.Region
}

// AwsAssumeRoleSageMakerListNotebookInstancesReq lists notebook instances via AssumeRole.
type AwsAssumeRoleSageMakerListNotebookInstancesReq struct {
	AwsAssumeRoleSageMakerBaseReq       `json:",inline"`
	CreationTimeAfter                   *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore                  *time.Time `json:"creation_time_before,omitempty"`
	DefaultCodeRepositoryContains       string     `json:"default_code_repository_contains,omitempty"`
	LastModifiedTimeAfter               *time.Time `json:"last_modified_time_after,omitempty"`
	LastModifiedTimeBefore              *time.Time `json:"last_modified_time_before,omitempty"`
	MaxResults                          *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                        string     `json:"name_contains,omitempty"`
	NextToken                           string     `json:"next_token,omitempty"`
	NotebookLifecycleConfigNameContains string     `json:"notebook_instance_lifecycle_config_name_contains,omitempty"`
	SortBy                              string     `json:"sort_by,omitempty"`
	SortOrder                           string     `json:"sort_order,omitempty"`
	StatusEquals                        string     `json:"status_equals,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListNotebookInstancesReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeNotebookInstanceReq describes a notebook instance via AssumeRole.
type AwsAssumeRoleSageMakerDescribeNotebookInstanceReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	NotebookInstanceName          string `json:"notebook_instance_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeNotebookInstanceReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListEndpointsReq lists endpoints via AssumeRole.
type AwsAssumeRoleSageMakerListEndpointsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	LastModifiedTimeAfter         *time.Time `json:"last_modified_time_after,omitempty"`
	LastModifiedTimeBefore        *time.Time `json:"last_modified_time_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
	StatusEquals                  string     `json:"status_equals,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListEndpointsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeEndpointReq describes an endpoint via AssumeRole.
type AwsAssumeRoleSageMakerDescribeEndpointReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	EndpointName                  string `json:"endpoint_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeEndpointReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListEndpointConfigsReq lists endpoint configs via AssumeRole.
type AwsAssumeRoleSageMakerListEndpointConfigsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListEndpointConfigsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeEndpointConfigReq describes an endpoint config via AssumeRole.
type AwsAssumeRoleSageMakerDescribeEndpointConfigReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	EndpointConfigName            string `json:"endpoint_config_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeEndpointConfigReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListTrainingJobsReq lists training jobs via AssumeRole.
type AwsAssumeRoleSageMakerListTrainingJobsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	LastModifiedTimeAfter         *time.Time `json:"last_modified_time_after,omitempty"`
	LastModifiedTimeBefore        *time.Time `json:"last_modified_time_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
	StatusEquals                  string     `json:"status_equals,omitempty"`
	WarmPoolStatusEquals          string     `json:"warm_pool_status_equals,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListTrainingJobsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeTrainingJobReq describes a training job via AssumeRole.
type AwsAssumeRoleSageMakerDescribeTrainingJobReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	TrainingJobName               string `json:"training_job_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeTrainingJobReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListProcessingJobsReq lists processing jobs via AssumeRole.
type AwsAssumeRoleSageMakerListProcessingJobsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	LastModifiedTimeAfter         *time.Time `json:"last_modified_time_after,omitempty"`
	LastModifiedTimeBefore        *time.Time `json:"last_modified_time_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
	StatusEquals                  string     `json:"status_equals,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListProcessingJobsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeProcessingJobReq describes a processing job via AssumeRole.
type AwsAssumeRoleSageMakerDescribeProcessingJobReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	ProcessingJobName             string `json:"processing_job_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeProcessingJobReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListTransformJobsReq lists transform jobs via AssumeRole.
type AwsAssumeRoleSageMakerListTransformJobsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	LastModifiedTimeAfter         *time.Time `json:"last_modified_time_after,omitempty"`
	LastModifiedTimeBefore        *time.Time `json:"last_modified_time_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
	StatusEquals                  string     `json:"status_equals,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListTransformJobsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeTransformJobReq describes a transform job via AssumeRole.
type AwsAssumeRoleSageMakerDescribeTransformJobReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	TransformJobName              string `json:"transform_job_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeTransformJobReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListAppsReq lists Studio apps via AssumeRole.
type AwsAssumeRoleSageMakerListAppsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	DomainIDEquals                string `json:"domain_id_equals,omitempty"`
	MaxResults                    *int32 `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NextToken                     string `json:"next_token,omitempty"`
	SortBy                        string `json:"sort_by,omitempty"`
	SortOrder                     string `json:"sort_order,omitempty"`
	SpaceNameEquals               string `json:"space_name_equals,omitempty"`
	UserProfileNameEquals         string `json:"user_profile_name_equals,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListAppsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeAppReq describes a Studio app via AssumeRole.
type AwsAssumeRoleSageMakerDescribeAppReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	DomainID                      string `json:"domain_id" validate:"required"`
	UserProfileName               string `json:"user_profile_name,omitempty"`
	SpaceName                     string `json:"space_name,omitempty"`
	AppType                       string `json:"app_type" validate:"required"`
	AppName                       string `json:"app_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeAppReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListClustersReq lists HyperPod clusters via AssumeRole.
type AwsAssumeRoleSageMakerListClustersReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListClustersReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeClusterReq describes a HyperPod cluster via AssumeRole.
type AwsAssumeRoleSageMakerDescribeClusterReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	ClusterName                   string `json:"cluster_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeClusterReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListClusterNodesReq lists HyperPod cluster nodes via AssumeRole.
type AwsAssumeRoleSageMakerListClusterNodesReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	ClusterName                   string     `json:"cluster_name" validate:"required"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	InstanceGroupNameContains     string     `json:"instance_group_name_contains,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListClusterNodesReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeClusterNodeReq describes a HyperPod cluster node via AssumeRole.
type AwsAssumeRoleSageMakerDescribeClusterNodeReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	ClusterName                   string `json:"cluster_name" validate:"required"`
	NodeID                        string `json:"node_id" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeClusterNodeReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsTrainingPlanFilter defines a Training Plan list filter.
type AwsTrainingPlanFilter struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// AwsAssumeRoleSageMakerListTrainingPlansReq lists Training Plans via AssumeRole.
type AwsAssumeRoleSageMakerListTrainingPlansReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	Filters                       []AwsTrainingPlanFilter `json:"filters,omitempty"`
	MaxResults                    *int32                  `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NextToken                     string                  `json:"next_token,omitempty"`
	SortBy                        string                  `json:"sort_by,omitempty"`
	SortOrder                     string                  `json:"sort_order,omitempty"`
	StartTimeAfter                *time.Time              `json:"start_time_after,omitempty"`
	StartTimeBefore               *time.Time              `json:"start_time_before,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListTrainingPlansReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeTrainingPlanReq describes a Training Plan via AssumeRole.
type AwsAssumeRoleSageMakerDescribeTrainingPlanReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	TrainingPlanName              string `json:"training_plan_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeTrainingPlanReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerSearchTrainingPlanOfferingsReq searches Training Plan offerings via AssumeRole.
type AwsAssumeRoleSageMakerSearchTrainingPlanOfferingsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	DurationHours                 *int64     `json:"duration_hours,omitempty"`
	EndTimeBefore                 *time.Time `json:"end_time_before,omitempty"`
	InstanceCount                 *int32     `json:"instance_count,omitempty" validate:"omitempty,min=1"`
	InstanceType                  string     `json:"instance_type,omitempty"`
	StartTimeAfter                *time.Time `json:"start_time_after,omitempty"`
	TargetResources               []string   `json:"target_resources,omitempty"`
	TrainingPlanArn               string     `json:"training_plan_arn,omitempty"`
	UltraServerCount              *int32     `json:"ultra_server_count,omitempty" validate:"omitempty,min=1"`
	UltraServerType               string     `json:"ultra_server_type,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerSearchTrainingPlanOfferingsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListInferenceComponentsReq lists inference components via AssumeRole.
type AwsAssumeRoleSageMakerListInferenceComponentsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	EndpointNameEquals            string     `json:"endpoint_name_equals,omitempty"`
	LastModifiedTimeAfter         *time.Time `json:"last_modified_time_after,omitempty"`
	LastModifiedTimeBefore        *time.Time `json:"last_modified_time_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
	StatusEquals                  string     `json:"status_equals,omitempty"`
	VariantNameEquals             string     `json:"variant_name_equals,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListInferenceComponentsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeInferenceComponentReq describes an inference component via AssumeRole.
type AwsAssumeRoleSageMakerDescribeInferenceComponentReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	InferenceComponentName        string `json:"inference_component_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeInferenceComponentReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListOptimizationJobsReq lists optimization jobs via AssumeRole.
type AwsAssumeRoleSageMakerListOptimizationJobsReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	CreationTimeAfter             *time.Time `json:"creation_time_after,omitempty"`
	CreationTimeBefore            *time.Time `json:"creation_time_before,omitempty"`
	LastModifiedTimeAfter         *time.Time `json:"last_modified_time_after,omitempty"`
	LastModifiedTimeBefore        *time.Time `json:"last_modified_time_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	OptimizationContains          string     `json:"optimization_contains,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
	StatusEquals                  string     `json:"status_equals,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListOptimizationJobsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeOptimizationJobReq describes an optimization job via AssumeRole.
type AwsAssumeRoleSageMakerDescribeOptimizationJobReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	OptimizationJobName           string `json:"optimization_job_name" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeOptimizationJobReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListComputeQuotasReq lists compute quotas via AssumeRole.
type AwsAssumeRoleSageMakerListComputeQuotasReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	ClusterArn                    string     `json:"cluster_arn,omitempty"`
	CreatedAfter                  *time.Time `json:"created_after,omitempty"`
	CreatedBefore                 *time.Time `json:"created_before,omitempty"`
	MaxResults                    *int32     `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NameContains                  string     `json:"name_contains,omitempty"`
	NextToken                     string     `json:"next_token,omitempty"`
	SortBy                        string     `json:"sort_by,omitempty"`
	SortOrder                     string     `json:"sort_order,omitempty"`
	Status                        string     `json:"status,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListComputeQuotasReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeComputeQuotaReq describes a compute quota via AssumeRole.
type AwsAssumeRoleSageMakerDescribeComputeQuotaReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	ComputeQuotaID                string `json:"compute_quota_id" validate:"required"`
	ComputeQuotaVersion           *int32 `json:"compute_quota_version,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeComputeQuotaReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerDescribeReservedCapacityReq describes a reserved capacity via AssumeRole.
type AwsAssumeRoleSageMakerDescribeReservedCapacityReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	ReservedCapacityArn           string `json:"reserved_capacity_arn" validate:"required"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerDescribeReservedCapacityReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsAssumeRoleSageMakerListUltraServersByReservedCapacityReq lists reserved capacity UltraServers.
type AwsAssumeRoleSageMakerListUltraServersByReservedCapacityReq struct {
	AwsAssumeRoleSageMakerBaseReq `json:",inline"`
	ReservedCapacityArn           string `json:"reserved_capacity_arn" validate:"required"`
	MaxResults                    *int32 `json:"max_results,omitempty" validate:"omitempty,min=1"`
	NextToken                     string `json:"next_token,omitempty"`
}

// Validate validates the request.
func (req *AwsAssumeRoleSageMakerListUltraServersByReservedCapacityReq) Validate() error {
	return validator.Validate.Struct(req)
}
