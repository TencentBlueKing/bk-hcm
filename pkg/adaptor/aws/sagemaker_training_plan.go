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

// ListTrainingPlans lists SageMaker Training Plans.
func (a *Aws) ListTrainingPlans(kt *kit.Kit,
	opt *adtsm.AwsListTrainingPlansOption) (*smv2.ListTrainingPlansOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListTrainingPlansInput{
		MaxResults:      opt.MaxResults,
		NextToken:       opt.NextToken,
		StartTimeAfter:  opt.StartTimeAfter,
		StartTimeBefore: opt.StartTimeBefore,
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.TrainingPlanSortBy(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.TrainingPlanSortOrder(opt.SortOrder)
	}
	if len(opt.Filters) != 0 {
		input.Filters = make([]smtypes.TrainingPlanFilter, 0, len(opt.Filters))
		for _, filter := range opt.Filters {
			input.Filters = append(input.Filters, smtypes.TrainingPlanFilter{
				Name:  smtypes.TrainingPlanFilterName(filter.Name),
				Value: converter.StrNilPtr(filter.Value),
			})
		}
	}

	resp, err := client.ListTrainingPlans(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker training plans failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeTrainingPlan describes a SageMaker Training Plan.
func (a *Aws) DescribeTrainingPlan(kt *kit.Kit,
	opt *adtsm.AwsDescribeTrainingPlanOption) (*smv2.DescribeTrainingPlanOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeTrainingPlan(kt.Ctx, &smv2.DescribeTrainingPlanInput{
		TrainingPlanName: converter.StrNilPtr(opt.TrainingPlanName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker training plan failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// SearchTrainingPlanOfferings searches SageMaker Training Plan offerings.
func (a *Aws) SearchTrainingPlanOfferings(kt *kit.Kit,
	opt *adtsm.AwsSearchTrainingPlanOfferingsOption) (*smv2.SearchTrainingPlanOfferingsOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.SearchTrainingPlanOfferingsInput{
		DurationHours:    opt.DurationHours,
		EndTimeBefore:    opt.EndTimeBefore,
		InstanceCount:    opt.InstanceCount,
		StartTimeAfter:   opt.StartTimeAfter,
		UltraServerCount: opt.UltraServerCount,
	}
	if opt.InstanceType != "" {
		input.InstanceType = smtypes.ReservedCapacityInstanceType(opt.InstanceType)
	}
	if opt.TrainingPlanArn != "" {
		input.TrainingPlanArn = converter.StrNilPtr(opt.TrainingPlanArn)
	}
	if opt.UltraServerType != "" {
		input.UltraServerType = converter.StrNilPtr(opt.UltraServerType)
	}
	if len(opt.TargetResources) != 0 {
		input.TargetResources = make([]smtypes.SageMakerResourceName, 0, len(opt.TargetResources))
		for _, resource := range opt.TargetResources {
			input.TargetResources = append(input.TargetResources, smtypes.SageMakerResourceName(resource))
		}
	}

	resp, err := client.SearchTrainingPlanOfferings(kt.Ctx, input)
	if err != nil {
		logs.Errorf("search aws sagemaker training plan offerings failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListInferenceComponents lists SageMaker inference components.
func (a *Aws) ListInferenceComponents(kt *kit.Kit,
	opt *adtsm.AwsListInferenceComponentsOption) (*smv2.ListInferenceComponentsOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListInferenceComponentsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		EndpointNameEquals:     converter.StrNilPtr(opt.EndpointNameEquals),
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NameContains:           converter.StrNilPtr(opt.NameContains),
		NextToken:              opt.NextToken,
		VariantNameEquals:      converter.StrNilPtr(opt.VariantNameEquals),
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.InferenceComponentSortKey(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.OrderKey(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = smtypes.InferenceComponentStatus(opt.StatusEquals)
	}

	resp, err := client.ListInferenceComponents(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker inference components failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeInferenceComponent describes a SageMaker inference component.
func (a *Aws) DescribeInferenceComponent(kt *kit.Kit,
	opt *adtsm.AwsDescribeInferenceComponentOption) (*smv2.DescribeInferenceComponentOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeInferenceComponent(kt.Ctx, &smv2.DescribeInferenceComponentInput{
		InferenceComponentName: converter.StrNilPtr(opt.InferenceComponentName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker inference component failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListOptimizationJobs lists SageMaker optimization jobs.
func (a *Aws) ListOptimizationJobs(kt *kit.Kit,
	opt *adtsm.AwsListOptimizationJobsOption) (*smv2.ListOptimizationJobsOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListOptimizationJobsInput{
		CreationTimeAfter:      opt.CreationTimeAfter,
		CreationTimeBefore:     opt.CreationTimeBefore,
		LastModifiedTimeAfter:  opt.LastModifiedTimeAfter,
		LastModifiedTimeBefore: opt.LastModifiedTimeBefore,
		MaxResults:             opt.MaxResults,
		NameContains:           converter.StrNilPtr(opt.NameContains),
		NextToken:              opt.NextToken,
		OptimizationContains:   converter.StrNilPtr(opt.OptimizationContains),
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.ListOptimizationJobsSortBy(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.SortOrder(opt.SortOrder)
	}
	if opt.StatusEquals != "" {
		input.StatusEquals = smtypes.OptimizationJobStatus(opt.StatusEquals)
	}

	resp, err := client.ListOptimizationJobs(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker optimization jobs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeOptimizationJob describes a SageMaker optimization job.
func (a *Aws) DescribeOptimizationJob(kt *kit.Kit,
	opt *adtsm.AwsDescribeOptimizationJobOption) (*smv2.DescribeOptimizationJobOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeOptimizationJob(kt.Ctx, &smv2.DescribeOptimizationJobInput{
		OptimizationJobName: converter.StrNilPtr(opt.OptimizationJobName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker optimization job failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListComputeQuotas lists SageMaker HyperPod compute quotas.
func (a *Aws) ListComputeQuotas(kt *kit.Kit,
	opt *adtsm.AwsListComputeQuotasOption) (*smv2.ListComputeQuotasOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	input := &smv2.ListComputeQuotasInput{
		ClusterArn:    converter.StrNilPtr(opt.ClusterArn),
		CreatedAfter:  opt.CreatedAfter,
		CreatedBefore: opt.CreatedBefore,
		MaxResults:    opt.MaxResults,
		NameContains:  converter.StrNilPtr(opt.NameContains),
		NextToken:     opt.NextToken,
	}
	if opt.SortBy != "" {
		input.SortBy = smtypes.SortQuotaBy(opt.SortBy)
	}
	if opt.SortOrder != "" {
		input.SortOrder = smtypes.SortOrder(opt.SortOrder)
	}
	if opt.Status != "" {
		input.Status = smtypes.SchedulerResourceStatus(opt.Status)
	}

	resp, err := client.ListComputeQuotas(kt.Ctx, input)
	if err != nil {
		logs.Errorf("list aws sagemaker compute quotas failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeComputeQuota describes a SageMaker HyperPod compute quota.
func (a *Aws) DescribeComputeQuota(kt *kit.Kit,
	opt *adtsm.AwsDescribeComputeQuotaOption) (*smv2.DescribeComputeQuotaOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeComputeQuota(kt.Ctx, &smv2.DescribeComputeQuotaInput{
		ComputeQuotaId:      converter.StrNilPtr(opt.ComputeQuotaID),
		ComputeQuotaVersion: opt.ComputeQuotaVersion,
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker compute quota failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// DescribeReservedCapacity describes a SageMaker reserved capacity.
func (a *Aws) DescribeReservedCapacity(kt *kit.Kit,
	opt *adtsm.AwsDescribeReservedCapacityOption) (*smv2.DescribeReservedCapacityOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeReservedCapacity(kt.Ctx, &smv2.DescribeReservedCapacityInput{
		ReservedCapacityArn: converter.StrNilPtr(opt.ReservedCapacityArn),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker reserved capacity failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// ListUltraServersByReservedCapacity lists UltraServers in a SageMaker reserved capacity.
func (a *Aws) ListUltraServersByReservedCapacity(kt *kit.Kit,
	opt *adtsm.AwsListUltraServersByReservedCapacityOption) (*smv2.ListUltraServersByReservedCapacityOutput, error) {

	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.ListUltraServersByReservedCapacity(kt.Ctx, &smv2.ListUltraServersByReservedCapacityInput{
		ReservedCapacityArn: converter.StrNilPtr(opt.ReservedCapacityArn),
		MaxResults:          opt.MaxResults,
		NextToken:           opt.NextToken,
	})
	if err != nil {
		logs.Errorf("list aws sagemaker reserved capacity ultra servers failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}
