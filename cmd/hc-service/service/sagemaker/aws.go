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
	adaptorsm "hcm/pkg/adaptor/types/sagemaker"
	proto "hcm/pkg/api/hc-service/sagemaker"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
)

// ListNotebookInstancesForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListNotebookInstancesForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListNotebookInstancesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListNotebookInstances(cts.Kit, &adaptorsm.AwsListNotebookInstancesOption{
		Region:                              req.Region,
		CreationTimeAfter:                   req.CreationTimeAfter,
		CreationTimeBefore:                  req.CreationTimeBefore,
		DefaultCodeRepositoryContains:       req.DefaultCodeRepositoryContains,
		LastModifiedTimeAfter:               req.LastModifiedTimeAfter,
		LastModifiedTimeBefore:              req.LastModifiedTimeBefore,
		MaxResults:                          req.MaxResults,
		NameContains:                        req.NameContains,
		NextToken:                           converter.StrNilPtr(req.NextToken),
		NotebookLifecycleConfigNameContains: req.NotebookLifecycleConfigNameContains,
		SortBy:                              req.SortBy,
		SortOrder:                           req.SortOrder,
		StatusEquals:                        req.StatusEquals,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker notebook instances failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetNotebookInstanceForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetNotebookInstanceForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeNotebookInstanceReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeNotebookInstance(cts.Kit, &adaptorsm.AwsDescribeNotebookInstanceOption{
		Region:               req.Region,
		NotebookInstanceName: req.NotebookInstanceName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker notebook instance failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListEndpointsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListEndpointsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListEndpointsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListEndpoints(cts.Kit, &adaptorsm.AwsListEndpointsOption{
		Region:                 req.Region,
		CreationTimeAfter:      req.CreationTimeAfter,
		CreationTimeBefore:     req.CreationTimeBefore,
		LastModifiedTimeAfter:  req.LastModifiedTimeAfter,
		LastModifiedTimeBefore: req.LastModifiedTimeBefore,
		MaxResults:             req.MaxResults,
		NameContains:           req.NameContains,
		NextToken:              converter.StrNilPtr(req.NextToken),
		SortBy:                 req.SortBy,
		SortOrder:              req.SortOrder,
		StatusEquals:           req.StatusEquals,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker endpoints failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetEndpointForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetEndpointForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeEndpointReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeEndpoint(cts.Kit, &adaptorsm.AwsDescribeEndpointOption{
		Region:       req.Region,
		EndpointName: req.EndpointName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker endpoint failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListEndpointConfigsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListEndpointConfigsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListEndpointConfigsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListEndpointConfigs(cts.Kit, &adaptorsm.AwsListEndpointConfigsOption{
		Region:             req.Region,
		CreationTimeAfter:  req.CreationTimeAfter,
		CreationTimeBefore: req.CreationTimeBefore,
		MaxResults:         req.MaxResults,
		NameContains:       req.NameContains,
		NextToken:          converter.StrNilPtr(req.NextToken),
		SortBy:             req.SortBy,
		SortOrder:          req.SortOrder,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker endpoint configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetEndpointConfigForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetEndpointConfigForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeEndpointConfigReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeEndpointConfig(cts.Kit, &adaptorsm.AwsDescribeEndpointConfigOption{
		Region:             req.Region,
		EndpointConfigName: req.EndpointConfigName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker endpoint config failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListTrainingJobsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListTrainingJobsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListTrainingJobsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListTrainingJobs(cts.Kit, &adaptorsm.AwsListTrainingJobsOption{
		Region:                 req.Region,
		CreationTimeAfter:      req.CreationTimeAfter,
		CreationTimeBefore:     req.CreationTimeBefore,
		LastModifiedTimeAfter:  req.LastModifiedTimeAfter,
		LastModifiedTimeBefore: req.LastModifiedTimeBefore,
		MaxResults:             req.MaxResults,
		NameContains:           req.NameContains,
		NextToken:              converter.StrNilPtr(req.NextToken),
		SortBy:                 req.SortBy,
		SortOrder:              req.SortOrder,
		StatusEquals:           req.StatusEquals,
		WarmPoolStatusEquals:   req.WarmPoolStatusEquals,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker training jobs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetTrainingJobForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetTrainingJobForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeTrainingJobReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeTrainingJob(cts.Kit, &adaptorsm.AwsDescribeTrainingJobOption{
		Region:          req.Region,
		TrainingJobName: req.TrainingJobName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker training job failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListProcessingJobsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListProcessingJobsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListProcessingJobsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListProcessingJobs(cts.Kit, &adaptorsm.AwsListProcessingJobsOption{
		Region:                 req.Region,
		CreationTimeAfter:      req.CreationTimeAfter,
		CreationTimeBefore:     req.CreationTimeBefore,
		LastModifiedTimeAfter:  req.LastModifiedTimeAfter,
		LastModifiedTimeBefore: req.LastModifiedTimeBefore,
		MaxResults:             req.MaxResults,
		NameContains:           req.NameContains,
		NextToken:              converter.StrNilPtr(req.NextToken),
		SortBy:                 req.SortBy,
		SortOrder:              req.SortOrder,
		StatusEquals:           req.StatusEquals,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker processing jobs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetProcessingJobForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetProcessingJobForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeProcessingJobReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeProcessingJob(cts.Kit, &adaptorsm.AwsDescribeProcessingJobOption{
		Region:            req.Region,
		ProcessingJobName: req.ProcessingJobName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker processing job failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListTransformJobsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListTransformJobsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListTransformJobsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListTransformJobs(cts.Kit, &adaptorsm.AwsListTransformJobsOption{
		Region:                 req.Region,
		CreationTimeAfter:      req.CreationTimeAfter,
		CreationTimeBefore:     req.CreationTimeBefore,
		LastModifiedTimeAfter:  req.LastModifiedTimeAfter,
		LastModifiedTimeBefore: req.LastModifiedTimeBefore,
		MaxResults:             req.MaxResults,
		NameContains:           req.NameContains,
		NextToken:              converter.StrNilPtr(req.NextToken),
		SortBy:                 req.SortBy,
		SortOrder:              req.SortOrder,
		StatusEquals:           req.StatusEquals,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker transform jobs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetTransformJobForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetTransformJobForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeTransformJobReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeTransformJob(cts.Kit, &adaptorsm.AwsDescribeTransformJobOption{
		Region:           req.Region,
		TransformJobName: req.TransformJobName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker transform job failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListAppsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListAppsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListAppsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListApps(cts.Kit, &adaptorsm.AwsListAppsOption{
		Region:                req.Region,
		DomainIDEquals:        req.DomainIDEquals,
		MaxResults:            req.MaxResults,
		NextToken:             converter.StrNilPtr(req.NextToken),
		SortBy:                req.SortBy,
		SortOrder:             req.SortOrder,
		SpaceNameEquals:       req.SpaceNameEquals,
		UserProfileNameEquals: req.UserProfileNameEquals,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker apps failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetAppForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetAppForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeAppReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeApp(cts.Kit, &adaptorsm.AwsDescribeAppOption{
		Region:          req.Region,
		DomainID:        req.DomainID,
		UserProfileName: req.UserProfileName,
		SpaceName:       req.SpaceName,
		AppType:         req.AppType,
		AppName:         req.AppName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker app failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListClustersForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListClustersForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListClustersReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListClusters(cts.Kit, &adaptorsm.AwsListClustersOption{
		Region:             req.Region,
		CreationTimeAfter:  req.CreationTimeAfter,
		CreationTimeBefore: req.CreationTimeBefore,
		MaxResults:         req.MaxResults,
		NameContains:       req.NameContains,
		NextToken:          converter.StrNilPtr(req.NextToken),
		SortBy:             req.SortBy,
		SortOrder:          req.SortOrder,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker hyperpod clusters failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetClusterForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetClusterForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeClusterReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeCluster(cts.Kit, &adaptorsm.AwsDescribeClusterOption{
		Region:      req.Region,
		ClusterName: req.ClusterName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker hyperpod cluster failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListClusterNodesForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListClusterNodesForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListClusterNodesReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListClusterNodes(cts.Kit, &adaptorsm.AwsListClusterNodesOption{
		Region:                    req.Region,
		ClusterName:               req.ClusterName,
		CreationTimeAfter:         req.CreationTimeAfter,
		CreationTimeBefore:        req.CreationTimeBefore,
		InstanceGroupNameContains: req.InstanceGroupNameContains,
		MaxResults:                req.MaxResults,
		NextToken:                 converter.StrNilPtr(req.NextToken),
		SortBy:                    req.SortBy,
		SortOrder:                 req.SortOrder,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker hyperpod cluster nodes failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetClusterNodeForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetClusterNodeForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeClusterNodeReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeClusterNode(cts.Kit, &adaptorsm.AwsDescribeClusterNodeOption{
		Region:      req.Region,
		ClusterName: req.ClusterName,
		NodeID:      req.NodeID,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker hyperpod cluster node failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListTrainingPlansForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListTrainingPlansForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListTrainingPlansReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListTrainingPlans(cts.Kit, &adaptorsm.AwsListTrainingPlansOption{
		Region:          req.Region,
		Filters:         convertTrainingPlanFilters(req.Filters),
		MaxResults:      req.MaxResults,
		NextToken:       converter.StrNilPtr(req.NextToken),
		SortBy:          req.SortBy,
		SortOrder:       req.SortOrder,
		StartTimeAfter:  req.StartTimeAfter,
		StartTimeBefore: req.StartTimeBefore,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker training plans failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetTrainingPlanForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetTrainingPlanForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeTrainingPlanReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeTrainingPlan(cts.Kit, &adaptorsm.AwsDescribeTrainingPlanOption{
		Region:           req.Region,
		TrainingPlanName: req.TrainingPlanName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker training plan failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// SearchTrainingPlanOfferingsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) SearchTrainingPlanOfferingsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerSearchTrainingPlanOfferingsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.SearchTrainingPlanOfferings(cts.Kit, &adaptorsm.AwsSearchTrainingPlanOfferingsOption{
		Region:           req.Region,
		DurationHours:    req.DurationHours,
		EndTimeBefore:    req.EndTimeBefore,
		InstanceCount:    req.InstanceCount,
		InstanceType:     req.InstanceType,
		StartTimeAfter:   req.StartTimeAfter,
		TargetResources:  req.TargetResources,
		TrainingPlanArn:  req.TrainingPlanArn,
		UltraServerCount: req.UltraServerCount,
		UltraServerType:  req.UltraServerType,
	})
	if err != nil {
		logs.Errorf("search aws assume role sagemaker training plan offerings failed, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListInferenceComponentsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListInferenceComponentsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListInferenceComponentsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListInferenceComponents(cts.Kit, &adaptorsm.AwsListInferenceComponentsOption{
		Region:                 req.Region,
		CreationTimeAfter:      req.CreationTimeAfter,
		CreationTimeBefore:     req.CreationTimeBefore,
		EndpointNameEquals:     req.EndpointNameEquals,
		LastModifiedTimeAfter:  req.LastModifiedTimeAfter,
		LastModifiedTimeBefore: req.LastModifiedTimeBefore,
		MaxResults:             req.MaxResults,
		NameContains:           req.NameContains,
		NextToken:              converter.StrNilPtr(req.NextToken),
		SortBy:                 req.SortBy,
		SortOrder:              req.SortOrder,
		StatusEquals:           req.StatusEquals,
		VariantNameEquals:      req.VariantNameEquals,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker inference components failed, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetInferenceComponentForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetInferenceComponentForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeInferenceComponentReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeInferenceComponent(cts.Kit, &adaptorsm.AwsDescribeInferenceComponentOption{
		Region:                 req.Region,
		InferenceComponentName: req.InferenceComponentName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker inference component failed, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListOptimizationJobsForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListOptimizationJobsForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListOptimizationJobsReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListOptimizationJobs(cts.Kit, &adaptorsm.AwsListOptimizationJobsOption{
		Region:                 req.Region,
		CreationTimeAfter:      req.CreationTimeAfter,
		CreationTimeBefore:     req.CreationTimeBefore,
		LastModifiedTimeAfter:  req.LastModifiedTimeAfter,
		LastModifiedTimeBefore: req.LastModifiedTimeBefore,
		MaxResults:             req.MaxResults,
		NameContains:           req.NameContains,
		NextToken:              converter.StrNilPtr(req.NextToken),
		OptimizationContains:   req.OptimizationContains,
		SortBy:                 req.SortBy,
		SortOrder:              req.SortOrder,
		StatusEquals:           req.StatusEquals,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker optimization jobs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetOptimizationJobForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetOptimizationJobForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeOptimizationJobReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeOptimizationJob(cts.Kit, &adaptorsm.AwsDescribeOptimizationJobOption{
		Region:              req.Region,
		OptimizationJobName: req.OptimizationJobName,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker optimization job failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListComputeQuotasForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListComputeQuotasForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListComputeQuotasReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListComputeQuotas(cts.Kit, &adaptorsm.AwsListComputeQuotasOption{
		Region:        req.Region,
		ClusterArn:    req.ClusterArn,
		CreatedAfter:  req.CreatedAfter,
		CreatedBefore: req.CreatedBefore,
		MaxResults:    req.MaxResults,
		NameContains:  req.NameContains,
		NextToken:     converter.StrNilPtr(req.NextToken),
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
		Status:        req.Status,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker compute quotas failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetComputeQuotaForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetComputeQuotaForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeComputeQuotaReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeComputeQuota(cts.Kit, &adaptorsm.AwsDescribeComputeQuotaOption{
		Region:              req.Region,
		ComputeQuotaID:      req.ComputeQuotaID,
		ComputeQuotaVersion: req.ComputeQuotaVersion,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker compute quota failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// GetReservedCapacityForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) GetReservedCapacityForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerDescribeReservedCapacityReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.DescribeReservedCapacity(cts.Kit, &adaptorsm.AwsDescribeReservedCapacityOption{
		Region:              req.Region,
		ReservedCapacityArn: req.ReservedCapacityArn,
	})
	if err != nil {
		logs.Errorf("describe aws assume role sagemaker reserved capacity failed, err: %v, rid: %s", err,
			cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

// ListUltraServersByReservedCapacityForAws handles the hc-service SageMaker assume-role passthrough request.
func (s *sageMakerSvc) ListUltraServersByReservedCapacityForAws(cts *rest.Contexts) (interface{}, error) {
	req := new(proto.AwsAssumeRoleSageMakerListUltraServersByReservedCapacityReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	client, err := s.assumeRoleClient(cts.Kit, req)
	if err != nil {
		return nil, err
	}
	result, err := client.ListUltraServersByReservedCapacity(
		cts.Kit,
		&adaptorsm.AwsListUltraServersByReservedCapacityOption{
			Region:              req.Region,
			ReservedCapacityArn: req.ReservedCapacityArn,
			MaxResults:          req.MaxResults,
			NextToken:           converter.StrNilPtr(req.NextToken),
		},
	)
	if err != nil {
		logs.Errorf("list aws assume role sagemaker reserved capacity ultra servers failed, err: %v, rid: %s",
			err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

func convertTrainingPlanFilters(filters []proto.AwsTrainingPlanFilter,
) []adaptorsm.AwsTrainingPlanFilter {

	if len(filters) == 0 {
		return nil
	}
	result := make([]adaptorsm.AwsTrainingPlanFilter, 0, len(filters))
	for _, filter := range filters {
		result = append(result, adaptorsm.AwsTrainingPlanFilter{Name: filter.Name, Value: filter.Value})
	}
	return result
}
