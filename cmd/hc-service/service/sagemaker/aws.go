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
)

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
		NextToken:                           stringPtr(req.NextToken),
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
		NextToken:              stringPtr(req.NextToken),
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
		NextToken:          stringPtr(req.NextToken),
		SortBy:             req.SortBy,
		SortOrder:          req.SortOrder,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker endpoint configs failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

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
		NextToken:              stringPtr(req.NextToken),
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
		NextToken:              stringPtr(req.NextToken),
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
		NextToken:              stringPtr(req.NextToken),
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

func stringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

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
		NextToken:             stringPtr(req.NextToken),
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
		NextToken:          stringPtr(req.NextToken),
		SortBy:             req.SortBy,
		SortOrder:          req.SortOrder,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker hyperpod clusters failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

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
		NextToken:                 stringPtr(req.NextToken),
		SortBy:                    req.SortBy,
		SortOrder:                 req.SortOrder,
	})
	if err != nil {
		logs.Errorf("list aws assume role sagemaker hyperpod cluster nodes failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return result, nil
}

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
