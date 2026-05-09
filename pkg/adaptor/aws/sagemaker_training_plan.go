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

	smv2 "github.com/aws/aws-sdk-go-v2/service/sagemaker"
	smtypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
)

// ListTrainingPlans lists SageMaker Training Plans.
func (a *Aws) ListTrainingPlans(kt *kit.Kit, opt *adtsm.AwsListTrainingPlansOption) (
	*smv2.ListTrainingPlansOutput, error,
) {
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
				Value: stringPtrV2(filter.Value),
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
func (a *Aws) DescribeTrainingPlan(kt *kit.Kit, opt *adtsm.AwsDescribeTrainingPlanOption) (
	*smv2.DescribeTrainingPlanOutput, error,
) {
	client, err := a.clientSet.sageMakerV2Client(opt.Region)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeTrainingPlan(kt.Ctx, &smv2.DescribeTrainingPlanInput{
		TrainingPlanName: stringPtrV2(opt.TrainingPlanName),
	})
	if err != nil {
		logs.Errorf("describe aws sagemaker training plan failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return resp, nil
}

// SearchTrainingPlanOfferings searches SageMaker Training Plan offerings.
func (a *Aws) SearchTrainingPlanOfferings(kt *kit.Kit, opt *adtsm.AwsSearchTrainingPlanOfferingsOption) (
	*smv2.SearchTrainingPlanOfferingsOutput, error,
) {
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
		input.TrainingPlanArn = stringPtrV2(opt.TrainingPlanArn)
	}
	if opt.UltraServerType != "" {
		input.UltraServerType = stringPtrV2(opt.UltraServerType)
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

func stringPtrV2(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
