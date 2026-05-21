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

import "time"

// AwsListNotebookInstancesOption defines list filters for notebook instances.
type AwsListNotebookInstancesOption struct {
	Region                              string
	CreationTimeAfter                   *time.Time
	CreationTimeBefore                  *time.Time
	DefaultCodeRepositoryContains       string
	LastModifiedTimeAfter               *time.Time
	LastModifiedTimeBefore              *time.Time
	MaxResults                          *int32
	NameContains                        string
	NextToken                           *string
	NotebookLifecycleConfigNameContains string
	SortBy                              string
	SortOrder                           string
	StatusEquals                        string
}

// AwsDescribeNotebookInstanceOption defines the notebook instance to describe.
type AwsDescribeNotebookInstanceOption struct {
	Region               string
	NotebookInstanceName string
}

// AwsListEndpointsOption defines list filters for endpoints.
type AwsListEndpointsOption struct {
	Region                 string
	CreationTimeAfter      *time.Time
	CreationTimeBefore     *time.Time
	LastModifiedTimeAfter  *time.Time
	LastModifiedTimeBefore *time.Time
	MaxResults             *int32
	NameContains           string
	NextToken              *string
	SortBy                 string
	SortOrder              string
	StatusEquals           string
}

// AwsDescribeEndpointOption defines the endpoint to describe.
type AwsDescribeEndpointOption struct {
	Region       string
	EndpointName string
}

// AwsListEndpointConfigsOption defines list filters for endpoint configs.
type AwsListEndpointConfigsOption struct {
	Region             string
	CreationTimeAfter  *time.Time
	CreationTimeBefore *time.Time
	MaxResults         *int32
	NameContains       string
	NextToken          *string
	SortBy             string
	SortOrder          string
}

// AwsDescribeEndpointConfigOption defines the endpoint config to describe.
type AwsDescribeEndpointConfigOption struct {
	Region             string
	EndpointConfigName string
}

// AwsListTrainingJobsOption defines list filters for training jobs.
type AwsListTrainingJobsOption struct {
	Region                 string
	CreationTimeAfter      *time.Time
	CreationTimeBefore     *time.Time
	LastModifiedTimeAfter  *time.Time
	LastModifiedTimeBefore *time.Time
	MaxResults             *int32
	NameContains           string
	NextToken              *string
	SortBy                 string
	SortOrder              string
	StatusEquals           string
	WarmPoolStatusEquals   string
}

// AwsDescribeTrainingJobOption defines the training job to describe.
type AwsDescribeTrainingJobOption struct {
	Region          string
	TrainingJobName string
}

// AwsListProcessingJobsOption defines list filters for processing jobs.
type AwsListProcessingJobsOption struct {
	Region                 string
	CreationTimeAfter      *time.Time
	CreationTimeBefore     *time.Time
	LastModifiedTimeAfter  *time.Time
	LastModifiedTimeBefore *time.Time
	MaxResults             *int32
	NameContains           string
	NextToken              *string
	SortBy                 string
	SortOrder              string
	StatusEquals           string
}

// AwsDescribeProcessingJobOption defines the processing job to describe.
type AwsDescribeProcessingJobOption struct {
	Region            string
	ProcessingJobName string
}

// AwsListTransformJobsOption defines list filters for transform jobs.
type AwsListTransformJobsOption struct {
	Region                 string
	CreationTimeAfter      *time.Time
	CreationTimeBefore     *time.Time
	LastModifiedTimeAfter  *time.Time
	LastModifiedTimeBefore *time.Time
	MaxResults             *int32
	NameContains           string
	NextToken              *string
	SortBy                 string
	SortOrder              string
	StatusEquals           string
}

// AwsDescribeTransformJobOption defines the transform job to describe.
type AwsDescribeTransformJobOption struct {
	Region           string
	TransformJobName string
}

// AwsListAppsOption defines list filters for Studio apps.
type AwsListAppsOption struct {
	Region                string
	DomainIDEquals        string
	MaxResults            *int32
	NextToken             *string
	SortBy                string
	SortOrder             string
	SpaceNameEquals       string
	UserProfileNameEquals string
}

// AwsDescribeAppOption defines the Studio app to describe.
type AwsDescribeAppOption struct {
	Region          string
	DomainID        string
	UserProfileName string
	SpaceName       string
	AppType         string
	AppName         string
}

// AwsListClustersOption defines list filters for HyperPod clusters.
type AwsListClustersOption struct {
	Region             string
	CreationTimeAfter  *time.Time
	CreationTimeBefore *time.Time
	MaxResults         *int32
	NameContains       string
	NextToken          *string
	SortBy             string
	SortOrder          string
}

// AwsDescribeClusterOption defines the HyperPod cluster to describe.
type AwsDescribeClusterOption struct {
	Region      string
	ClusterName string
}

// AwsListClusterNodesOption defines list filters for HyperPod cluster nodes.
type AwsListClusterNodesOption struct {
	Region                    string
	ClusterName               string
	CreationTimeAfter         *time.Time
	CreationTimeBefore        *time.Time
	InstanceGroupNameContains string
	MaxResults                *int32
	NextToken                 *string
	SortBy                    string
	SortOrder                 string
}

// AwsDescribeClusterNodeOption defines the HyperPod cluster node to describe.
type AwsDescribeClusterNodeOption struct {
	Region      string
	ClusterName string
	NodeID      string
}

// AwsTrainingPlanFilter defines a Training Plan list filter.
type AwsTrainingPlanFilter struct {
	Name  string
	Value string
}

// AwsListTrainingPlansOption defines list filters for Training Plans.
type AwsListTrainingPlansOption struct {
	Region          string
	Filters         []AwsTrainingPlanFilter
	MaxResults      *int32
	NextToken       *string
	SortBy          string
	SortOrder       string
	StartTimeAfter  *time.Time
	StartTimeBefore *time.Time
}

// AwsDescribeTrainingPlanOption defines the Training Plan to describe.
type AwsDescribeTrainingPlanOption struct {
	Region           string
	TrainingPlanName string
}

// AwsSearchTrainingPlanOfferingsOption defines search filters for Training Plan offerings.
type AwsSearchTrainingPlanOfferingsOption struct {
	Region           string
	DurationHours    *int64
	EndTimeBefore    *time.Time
	InstanceCount    *int32
	InstanceType     string
	StartTimeAfter   *time.Time
	TargetResources  []string
	TrainingPlanArn  string
	UltraServerCount *int32
	UltraServerType  string
}

// AwsListInferenceComponentsOption defines list filters for inference components.
type AwsListInferenceComponentsOption struct {
	Region                 string
	CreationTimeAfter      *time.Time
	CreationTimeBefore     *time.Time
	EndpointNameEquals     string
	LastModifiedTimeAfter  *time.Time
	LastModifiedTimeBefore *time.Time
	MaxResults             *int32
	NameContains           string
	NextToken              *string
	SortBy                 string
	SortOrder              string
	StatusEquals           string
	VariantNameEquals      string
}

// AwsDescribeInferenceComponentOption defines the inference component to describe.
type AwsDescribeInferenceComponentOption struct {
	Region                 string
	InferenceComponentName string
}

// AwsListOptimizationJobsOption defines list filters for optimization jobs.
type AwsListOptimizationJobsOption struct {
	Region                 string
	CreationTimeAfter      *time.Time
	CreationTimeBefore     *time.Time
	LastModifiedTimeAfter  *time.Time
	LastModifiedTimeBefore *time.Time
	MaxResults             *int32
	NameContains           string
	NextToken              *string
	OptimizationContains   string
	SortBy                 string
	SortOrder              string
	StatusEquals           string
}

// AwsDescribeOptimizationJobOption defines the optimization job to describe.
type AwsDescribeOptimizationJobOption struct {
	Region              string
	OptimizationJobName string
}

// AwsListComputeQuotasOption defines list filters for compute quotas.
type AwsListComputeQuotasOption struct {
	Region        string
	ClusterArn    string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	MaxResults    *int32
	NameContains  string
	NextToken     *string
	SortBy        string
	SortOrder     string
	Status        string
}

// AwsDescribeComputeQuotaOption defines the compute quota to describe.
type AwsDescribeComputeQuotaOption struct {
	Region              string
	ComputeQuotaID      string
	ComputeQuotaVersion *int32
}

// AwsDescribeReservedCapacityOption defines the reserved capacity to describe.
type AwsDescribeReservedCapacityOption struct {
	Region              string
	ReservedCapacityArn string
}

// AwsListUltraServersByReservedCapacityOption defines list filters for reserved capacity UltraServers.
type AwsListUltraServersByReservedCapacityOption struct {
	Region              string
	ReservedCapacityArn string
	MaxResults          *int32
	NextToken           *string
}
