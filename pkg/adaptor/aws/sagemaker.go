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

	"github.com/aws/aws-sdk-go/aws"
	awssm "github.com/aws/aws-sdk-go/service/sagemaker"
)

// ListNotebookInstances lists notebook instances via SageMaker.
func (a *Aws) ListNotebookInstances(kt *kit.Kit, opt *adtsm.AwsListNotebookInstancesOption) (
	*awssm.ListNotebookInstancesOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListNotebookInstancesInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.DefaultCodeRepositoryContains != "" {
		input.DefaultCodeRepositoryContains = aws.String(opt.DefaultCodeRepositoryContains)
	}
	if opt.NameContains != "" {
		input.NameContains = aws.String(opt.NameContains)
	}
	if opt.NotebookLifecycleConfigNameContains != "" {
		input.NotebookInstanceLifecycleConfigNameContains = aws.String(opt.NotebookLifecycleConfigNameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = aws.String(opt.StatusEquals)
	}

	resp, err := client.ListNotebookInstancesWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker notebook instances failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeNotebookInstance describes a notebook instance via SageMaker.
func (a *Aws) DescribeNotebookInstance(kt *kit.Kit, opt *adtsm.AwsDescribeNotebookInstanceOption) (
	*awssm.DescribeNotebookInstanceOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeNotebookInstanceWithContext(kt.Ctx, &awssm.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(opt.NotebookInstanceName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker notebook instance failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListEndpoints lists SageMaker endpoints.
func (a *Aws) ListEndpoints(kt *kit.Kit, opt *adtsm.AwsListEndpointsOption) (*awssm.ListEndpointsOutput, error) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListEndpointsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = aws.String(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = aws.String(opt.StatusEquals)
	}

	resp, err := client.ListEndpointsWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker endpoints failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeEndpoint describes a SageMaker endpoint.
func (a *Aws) DescribeEndpoint(kt *kit.Kit, opt *adtsm.AwsDescribeEndpointOption) (
	*awssm.DescribeEndpointOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeEndpointWithContext(kt.Ctx, &awssm.DescribeEndpointInput{
		EndpointName: aws.String(opt.EndpointName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker endpoint failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListEndpointConfigs lists SageMaker endpoint configs.
func (a *Aws) ListEndpointConfigs(kt *kit.Kit, opt *adtsm.AwsListEndpointConfigsOption) (
	*awssm.ListEndpointConfigsOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListEndpointConfigsInput{
		CreationTimeAfter:  opt.CreationTimeAfter,
		CreationTimeBefore: opt.CreationTimeBefore,
		MaxResults:         opt.MaxResults,
		NextToken:          opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = aws.String(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}

	resp, err := client.ListEndpointConfigsWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker endpoint configs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeEndpointConfig describes a SageMaker endpoint config.
func (a *Aws) DescribeEndpointConfig(kt *kit.Kit, opt *adtsm.AwsDescribeEndpointConfigOption) (
	*awssm.DescribeEndpointConfigOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeEndpointConfigWithContext(kt.Ctx, &awssm.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(opt.EndpointConfigName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker endpoint config failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListTrainingJobs lists SageMaker training jobs.
func (a *Aws) ListTrainingJobs(kt *kit.Kit, opt *adtsm.AwsListTrainingJobsOption) (
	*awssm.ListTrainingJobsOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListTrainingJobsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = aws.String(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = aws.String(opt.StatusEquals)
	}
	if opt.WarmPoolStatusEquals != "" {
		input.WarmPoolStatusEquals = aws.String(opt.WarmPoolStatusEquals)
	}

	resp, err := client.ListTrainingJobsWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker training jobs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeTrainingJob describes a SageMaker training job.
func (a *Aws) DescribeTrainingJob(kt *kit.Kit, opt *adtsm.AwsDescribeTrainingJobOption) (
	*awssm.DescribeTrainingJobOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeTrainingJobWithContext(kt.Ctx, &awssm.DescribeTrainingJobInput{
		TrainingJobName: aws.String(opt.TrainingJobName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker training job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListProcessingJobs lists SageMaker processing jobs.
func (a *Aws) ListProcessingJobs(kt *kit.Kit, opt *adtsm.AwsListProcessingJobsOption) (
	*awssm.ListProcessingJobsOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListProcessingJobsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = aws.String(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = aws.String(opt.StatusEquals)
	}

	resp, err := client.ListProcessingJobsWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker processing jobs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeProcessingJob describes a SageMaker processing job.
func (a *Aws) DescribeProcessingJob(kt *kit.Kit, opt *adtsm.AwsDescribeProcessingJobOption) (
	*awssm.DescribeProcessingJobOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeProcessingJobWithContext(kt.Ctx, &awssm.DescribeProcessingJobInput{
		ProcessingJobName: aws.String(opt.ProcessingJobName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker processing job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListTransformJobs lists SageMaker transform jobs.
func (a *Aws) ListTransformJobs(kt *kit.Kit, opt *adtsm.AwsListTransformJobsOption) (
	*awssm.ListTransformJobsOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListTransformJobsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NextToken:              opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = aws.String(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = aws.String(opt.StatusEquals)
	}

	resp, err := client.ListTransformJobsWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker transform jobs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeTransformJob describes a SageMaker transform job.
func (a *Aws) DescribeTransformJob(kt *kit.Kit, opt *adtsm.AwsDescribeTransformJobOption) (
	*awssm.DescribeTransformJobOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeTransformJobWithContext(kt.Ctx, &awssm.DescribeTransformJobInput{
		TransformJobName: aws.String(opt.TransformJobName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker transform job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListApps lists SageMaker Studio apps.
func (a *Aws) ListApps(kt *kit.Kit, opt *adtsm.AwsListAppsOption) (*awssm.ListAppsOutput, error) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListAppsInput{
		MaxResults: opt.MaxResults,
		NextToken:  opt.NextToken,
	}
	if opt.DomainIDEquals != "" {
		input.DomainIdEquals = aws.String(opt.DomainIDEquals)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}
	if opt.SpaceNameEquals != "" {
		input.SpaceNameEquals = aws.String(opt.SpaceNameEquals)
	}
	if opt.UserProfileNameEquals != "" {
		input.UserProfileNameEquals = aws.String(opt.UserProfileNameEquals)
	}

	resp, err := client.ListAppsWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeApp describes a SageMaker Studio app.
func (a *Aws) DescribeApp(kt *kit.Kit, opt *adtsm.AwsDescribeAppOption) (*awssm.DescribeAppOutput, error) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.DescribeAppInput{
		AppName:  aws.String(opt.AppName),
		AppType:  aws.String(opt.AppType),
		DomainId: aws.String(opt.DomainID),
	}
	if opt.SpaceName != "" {
		input.SpaceName = aws.String(opt.SpaceName)
	}
	if opt.UserProfileName != "" {
		input.UserProfileName = aws.String(opt.UserProfileName)
	}

	resp, err := client.DescribeAppWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("describe aws sagemaker app failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListClusters lists SageMaker HyperPod clusters.
func (a *Aws) ListClusters(kt *kit.Kit, opt *adtsm.AwsListClustersOption) (*awssm.ListClustersOutput, error) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListClustersInput{
		CreationTimeAfter:  opt.CreationTimeAfter,
		CreationTimeBefore: opt.CreationTimeBefore,
		MaxResults:         opt.MaxResults,
		NextToken:          opt.NextToken,
	}
	if opt.NameContains != "" {
		input.NameContains = aws.String(opt.NameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}

	resp, err := client.ListClustersWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker hyperpod clusters failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeCluster describes a SageMaker HyperPod cluster.
func (a *Aws) DescribeCluster(kt *kit.Kit, opt *adtsm.AwsDescribeClusterOption) (*awssm.DescribeClusterOutput, error) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeClusterWithContext(kt.Ctx, &awssm.DescribeClusterInput{
		ClusterName: aws.String(opt.ClusterName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker hyperpod cluster failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListClusterNodes lists SageMaker HyperPod cluster nodes.
func (a *Aws) ListClusterNodes(kt *kit.Kit, opt *adtsm.AwsListClusterNodesOption) (
	*awssm.ListClusterNodesOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &awssm.ListClusterNodesInput{
		ClusterName:        aws.String(opt.ClusterName),
		CreationTimeAfter:  opt.CreationTimeAfter,
		CreationTimeBefore: opt.CreationTimeBefore,
		MaxResults:         opt.MaxResults,
		NextToken:          opt.NextToken,
	}
	if opt.InstanceGroupNameContains != "" {
		input.InstanceGroupNameContains = aws.String(opt.InstanceGroupNameContains)
	}
	if opt.SortBy != "" {
		input.SortBy = aws.String(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = aws.String(opt.SortOrder)
	}

	resp, err := client.ListClusterNodesWithContext(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker hyperpod cluster nodes failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeClusterNode describes a SageMaker HyperPod cluster node.
func (a *Aws) DescribeClusterNode(kt *kit.Kit, opt *adtsm.AwsDescribeClusterNodeOption) (
	*awssm.DescribeClusterNodeOutput, error,
) {
	client, err := a.clientSet.sageMakerClient(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeClusterNodeWithContext(kt.Ctx, &awssm.DescribeClusterNodeInput{
		ClusterName: aws.String(opt.ClusterName),
		NodeId:      aws.String(opt.NodeID),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker hyperpod cluster node failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}
