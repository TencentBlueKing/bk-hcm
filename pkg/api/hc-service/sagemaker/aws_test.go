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

import "testing"

// TestAwsAssumeRoleSageMakerListNotebookInstancesReqValidate validates shared assume-role fields.
func TestAwsAssumeRoleSageMakerListNotebookInstancesReqValidate(t *testing.T) {
	valid := AwsAssumeRoleSageMakerListNotebookInstancesReq{
		AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
			RootAccountID: "root-account-id",
			MainAccountID: "main-account-id",
			RoleChain:     []string{"OrgAccountAccessRole"},
			Region:        "ap-southeast-1",
		},
	}

	tests := []struct {
		name    string
		mutate  func(*AwsAssumeRoleSageMakerListNotebookInstancesReq)
		wantErr bool
	}{
		{name: "valid", mutate: func(*AwsAssumeRoleSageMakerListNotebookInstancesReq) {}, wantErr: false},
		{
			name: "missing root account",
			mutate: func(req *AwsAssumeRoleSageMakerListNotebookInstancesReq) {
				req.RootAccountID = ""
			},
			wantErr: true,
		},
		{
			name: "missing main account",
			mutate: func(req *AwsAssumeRoleSageMakerListNotebookInstancesReq) {
				req.MainAccountID = ""
			},
			wantErr: true,
		},
		{
			name: "missing role chain",
			mutate: func(req *AwsAssumeRoleSageMakerListNotebookInstancesReq) {
				req.RoleChain = nil
			},
			wantErr: true,
		},
		{
			name: "missing region",
			mutate: func(req *AwsAssumeRoleSageMakerListNotebookInstancesReq) {
				req.Region = ""
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := valid
			tc.mutate(&req)
			err := req.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// TestAwsAssumeRoleSageMakerDescribeRequestsValidate validates required identifiers on get-style requests.
func TestAwsAssumeRoleSageMakerDescribeRequestsValidate(t *testing.T) {
	tests := []struct {
		name    string
		request interface{ Validate() error }
		wantErr bool
	}{
		{
			name: "describe notebook valid",
			request: &AwsAssumeRoleSageMakerDescribeNotebookInstanceReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
				NotebookInstanceName: "demo-notebook",
			},
		},
		{
			name: "describe endpoint missing endpoint name",
			request: &AwsAssumeRoleSageMakerDescribeEndpointReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
			},
			wantErr: true,
		},
		{
			name: "describe training job missing training job name",
			request: &AwsAssumeRoleSageMakerDescribeTrainingJobReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
			},
			wantErr: true,
		},
		{
			name: "describe app valid",
			request: &AwsAssumeRoleSageMakerDescribeAppReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
				DomainID:        "d-xxxxxxxxxxxx",
				UserProfileName: "demo-user",
				AppType:         "JupyterLab",
				AppName:         "default",
			},
		},
		{
			name: "describe cluster valid",
			request: &AwsAssumeRoleSageMakerDescribeClusterReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-west-2",
				},
				ClusterName: "demo-cluster",
			},
		},
		{
			name: "describe cluster node valid",
			request: &AwsAssumeRoleSageMakerDescribeClusterNodeReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-west-2",
				},
				ClusterName: "demo-cluster",
				NodeID:      "i-0123456789abcdef0",
			},
		},
		{
			name: "describe training plan valid",
			request: &AwsAssumeRoleSageMakerDescribeTrainingPlanReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
				TrainingPlanName: "demo-training-plan",
			},
		},

		{
			name: "describe inference component valid",
			request: &AwsAssumeRoleSageMakerDescribeInferenceComponentReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
				InferenceComponentName: "demo-component",
			},
		},
		{
			name: "describe inference component missing name",
			request: &AwsAssumeRoleSageMakerDescribeInferenceComponentReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
			},
			wantErr: true,
		},
		{
			name: "describe optimization job valid",
			request: &AwsAssumeRoleSageMakerDescribeOptimizationJobReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
				OptimizationJobName: "demo-optimization-job",
			},
		},
		{
			name: "describe compute quota valid",
			request: &AwsAssumeRoleSageMakerDescribeComputeQuotaReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-west-2",
				},
				ComputeQuotaID: "cq-123",
			},
		},
		{
			name: "describe reserved capacity valid",
			request: &AwsAssumeRoleSageMakerDescribeReservedCapacityReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
				ReservedCapacityArn: "arn:aws:sagemaker:us-east-1:123456789012:reserved-capacity/demo",
			},
		},
		{
			name: "list ultra servers missing reserved capacity arn",
			request: &AwsAssumeRoleSageMakerListUltraServersByReservedCapacityReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
			},
			wantErr: true,
		},
		{
			name: "describe training plan missing name",
			request: &AwsAssumeRoleSageMakerDescribeTrainingPlanReq{
				AwsAssumeRoleSageMakerBaseReq: AwsAssumeRoleSageMakerBaseReq{
					RootAccountID: "root-account-id",
					MainAccountID: "main-account-id",
					RoleChain:     []string{"OrgAccountAccessRole"},
					Region:        "us-east-1",
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
