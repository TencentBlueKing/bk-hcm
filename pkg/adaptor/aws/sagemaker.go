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
	adtsm "hcm/pkg/adaptor/types/sagemaker"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	smv2 "github.com/aws/aws-sdk-go-v2/service/sagemaker"
	smtypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
)

// ListNotebookInstances lists notebook instances via SageMaker.
func (a *Aws) ListNotebookInstances(kt *kit.Kit,
	opt *adtsm.AwsListNotebookInstancesOption) (*smv2.ListNotebookInstancesOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListNotebookInstancesInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.DefaultCodeRepositoryContains != "" {
		input.DefaultCodeRepositoryContains = converter.StrNilPtr(opt.DefaultCodeRepositoryContains)
	}
	if opt.NameContains != "" {
		input.NameContains = converter.StrNilPtr(opt.NameContains)
	}
	if opt.NotebookLifecycleConfigNameContains != "" {
		input.NotebookInstanceLifecycleConfigNameContains = converter.StrNilPtr(
			opt.NotebookLifecycleConfigNameContains,
		)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.NotebookInstanceSortKey(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.NotebookInstanceSortOrder(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = smtypes.NotebookInstanceStatus(opt.StatusEquals)
	}

	resp, err := client.ListNotebookInstances(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker notebook instances failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeNotebookInstance describes a notebook instance via SageMaker.
func (a *Aws) DescribeNotebookInstance(kt *kit.Kit,
	opt *adtsm.AwsDescribeNotebookInstanceOption) (*smv2.DescribeNotebookInstanceOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeNotebookInstance(kt.Ctx, &smv2.DescribeNotebookInstanceInput{
		NotebookInstanceName: converter.StrNilPtr(opt.NotebookInstanceName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker notebook instance failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListEndpoints lists SageMaker endpoints.
func (a *Aws) ListEndpoints(kt *kit.Kit, opt *adtsm.AwsListEndpointsOption) (*smv2.ListEndpointsOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListEndpointsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = converter.StrNilPtr(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.EndpointSortKey(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.OrderKey(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = smtypes.EndpointStatus(opt.StatusEquals)
	}

	resp, err := client.ListEndpoints(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker endpoints failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeEndpoint describes a SageMaker endpoint.
func (a *Aws) DescribeEndpoint(kt *kit.Kit,
	opt *adtsm.AwsDescribeEndpointOption) (*smv2.DescribeEndpointOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeEndpoint(kt.Ctx, &smv2.DescribeEndpointInput{
		EndpointName: converter.StrNilPtr(opt.EndpointName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker endpoint failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListEndpointConfigs lists SageMaker endpoint configs.
func (a *Aws) ListEndpointConfigs(kt *kit.Kit,
	opt *adtsm.AwsListEndpointConfigsOption) (*smv2.ListEndpointConfigsOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListEndpointConfigsInput{
		CreationTimeAfter:  opt.CreationTimeAfter,
		CreationTimeBefore: opt.CreationTimeBefore,
		MaxResults:         opt.MaxResults,
		NextToken:          opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = converter.StrNilPtr(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.EndpointConfigSortKey(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.OrderKey(opt.SortOrder)
	}

	resp, err := client.ListEndpointConfigs(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker endpoint configs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeEndpointConfig describes a SageMaker endpoint config.
func (a *Aws) DescribeEndpointConfig(kt *kit.Kit,
	opt *adtsm.AwsDescribeEndpointConfigOption) (*smv2.DescribeEndpointConfigOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeEndpointConfig(kt.Ctx, &smv2.DescribeEndpointConfigInput{
		EndpointConfigName: converter.StrNilPtr(opt.EndpointConfigName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker endpoint config failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListTrainingJobs lists SageMaker training jobs.
func (a *Aws) ListTrainingJobs(kt *kit.Kit,
	opt *adtsm.AwsListTrainingJobsOption) (*smv2.ListTrainingJobsOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListTrainingJobsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = converter.StrNilPtr(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.SortBy(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.SortOrder(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = smtypes.TrainingJobStatus(opt.StatusEquals)
	}
	if opt.WarmPoolStatusEquals != "" {
		input.WarmPoolStatusEquals = smtypes.WarmPoolResourceStatus(opt.WarmPoolStatusEquals)
	}

	resp, err := client.ListTrainingJobs(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker training jobs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeTrainingJob describes a SageMaker training job.
func (a *Aws) DescribeTrainingJob(kt *kit.Kit,
	opt *adtsm.AwsDescribeTrainingJobOption) (*smv2.DescribeTrainingJobOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeTrainingJob(kt.Ctx, &smv2.DescribeTrainingJobInput{
		TrainingJobName: converter.StrNilPtr(opt.TrainingJobName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker training job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListProcessingJobs lists SageMaker processing jobs.
func (a *Aws) ListProcessingJobs(kt *kit.Kit,
	opt *adtsm.AwsListProcessingJobsOption) (*smv2.ListProcessingJobsOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListProcessingJobsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = converter.StrNilPtr(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.SortBy(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.SortOrder(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = smtypes.ProcessingJobStatus(opt.StatusEquals)
	}

	resp, err := client.ListProcessingJobs(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker processing jobs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeProcessingJob describes a SageMaker processing job.
func (a *Aws) DescribeProcessingJob(kt *kit.Kit,
	opt *adtsm.AwsDescribeProcessingJobOption) (*smv2.DescribeProcessingJobOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeProcessingJob(kt.Ctx, &smv2.DescribeProcessingJobInput{
		ProcessingJobName: converter.StrNilPtr(opt.ProcessingJobName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker processing job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListTransformJobs lists SageMaker transform jobs.
func (a *Aws) ListTransformJobs(kt *kit.Kit,
	opt *adtsm.AwsListTransformJobsOption) (*smv2.ListTransformJobsOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListTransformJobsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = converter.StrNilPtr(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.SortBy(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.SortOrder(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = smtypes.TransformJobStatus(opt.StatusEquals)
	}

	resp, err := client.ListTransformJobs(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker transform jobs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeTransformJob describes a SageMaker transform job.
func (a *Aws) DescribeTransformJob(kt *kit.Kit,
	opt *adtsm.AwsDescribeTransformJobOption) (*smv2.DescribeTransformJobOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeTransformJob(kt.Ctx, &smv2.DescribeTransformJobInput{
		TransformJobName: converter.StrNilPtr(opt.TransformJobName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker transform job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListApps lists SageMaker Studio apps.
func (a *Aws) ListApps(kt *kit.Kit, opt *adtsm.AwsListAppsOption) (*smv2.ListAppsOutput, error) {
	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListAppsInput{
		MaxResults: opt.MaxResults,
		NextToken:  opt.NextToken,
	}
	if opt.DomainIDEquals != "" {
		input.DomainIdEquals = converter.StrNilPtr(opt.DomainIDEquals)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.AppSortKey(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.SortOrder(opt.SortOrder)
	}
	if opt.SpaceNameEquals != "" {
		input.SpaceNameEquals = converter.StrNilPtr(opt.SpaceNameEquals)
	}
	if opt.UserProfileNameEquals != "" {
		input.UserProfileNameEquals = converter.StrNilPtr(opt.UserProfileNameEquals)
	}

	resp, err := client.ListApps(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeApp describes a SageMaker Studio app.
func (a *Aws) DescribeApp(kt *kit.Kit, opt *adtsm.AwsDescribeAppOption) (*smv2.DescribeAppOutput, error) {
	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.DescribeAppInput{
		AppName:  converter.StrNilPtr(opt.AppName),
		AppType:  smtypes.AppType(opt.AppType),
		DomainId: converter.StrNilPtr(opt.DomainID),
	}
	if opt.SpaceName != "" {
		input.SpaceName = converter.StrNilPtr(opt.SpaceName)
	}
	if opt.UserProfileName != "" {
		input.UserProfileName = converter.StrNilPtr(opt.UserProfileName)
	}

	resp, err := client.DescribeApp(kt.Ctx, input)
	if err != nil {
		logs.Errorf("describe aws sagemaker app failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListClusters lists SageMaker HyperPod clusters.
func (a *Aws) ListClusters(kt *kit.Kit, opt *adtsm.AwsListClustersOption) (*smv2.ListClustersOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListClustersInput{
		CreationTimeAfter:  opt.CreationTimeAfter,
		CreationTimeBefore: opt.CreationTimeBefore,
		MaxResults:         opt.MaxResults,
		NextToken:          opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = converter.StrNilPtr(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.ClusterSortBy(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.SortOrder(opt.SortOrder)
	}

	resp, err := client.ListClusters(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker hyperpod clusters failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeCluster describes a SageMaker HyperPod cluster.
func (a *Aws) DescribeCluster(kt *kit.Kit, opt *adtsm.AwsDescribeClusterOption) (*smv2.DescribeClusterOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeCluster(kt.Ctx, &smv2.DescribeClusterInput{
		ClusterName: converter.StrNilPtr(opt.ClusterName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker hyperpod cluster failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListClusterNodes lists SageMaker HyperPod cluster nodes.
func (a *Aws) ListClusterNodes(kt *kit.Kit,
	opt *adtsm.AwsListClusterNodesOption) (*smv2.ListClusterNodesOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListClusterNodesInput{
		ClusterName:        converter.StrNilPtr(opt.ClusterName),
		CreationTimeAfter:  opt.CreationTimeAfter,
		CreationTimeBefore: opt.CreationTimeBefore,
		MaxResults:         opt.MaxResults,
		NextToken:          opt.NextToken,
	}
	if opt.InstanceGroupNameContains != "" {
		input.InstanceGroupNameContains = converter.StrNilPtr(opt.InstanceGroupNameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.ClusterSortBy(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.SortOrder(opt.SortOrder)
	}

	resp, err := client.ListClusterNodes(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker hyperpod cluster nodes failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeClusterNode describes a SageMaker HyperPod cluster node.
func (a *Aws) DescribeClusterNode(kt *kit.Kit,
	opt *adtsm.AwsDescribeClusterNodeOption) (*smv2.DescribeClusterNodeOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeClusterNode(kt.Ctx, &smv2.DescribeClusterNodeInput{
		ClusterName: converter.StrNilPtr(opt.ClusterName),
		NodeId:      converter.StrNilPtr(opt.NodeID),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker hyperpod cluster node failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}
