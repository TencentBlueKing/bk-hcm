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

package gcp

import (
	"fmt"
	"net/http"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"cloud.google.com/go/iam"
	"google.golang.org/api/cloudbilling/v1"
	resourcemanager "google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
)

const (
	// ProjectIdExistedErrorMessage error message for project id existed.
	ProjectIdExistedErrorMessage = "Requested entity already exists"
	// ProjectIdExistedErrorCode error code for project id existed.
	ProjectIdExistedErrorCode = http.StatusConflict
)

// CreateProject create project, if success, return project id.
func (g *Gcp) CreateProject(kt *kit.Kit, name string, organization string) (projectId string, err error) {
	// Note: 根据name生成合规的ID，id自动生成出于以下考虑：
	// 1.产品设计上只让用户输入name，id需要自动生成；
	// 2.gcp页面上创建也可以只输入name，自动生成一个推荐的id（规则与hcm规则一致，后附6位随机数）。
	// 3.由于没有gcp的id是否重复的校验接口，故通过2次随机降低创建失败概率
	return g.createProjectWithId(kt, formatToProjectID(name), name, organization)
}

// reference: https://cloud.google.com/resource-manager/reference/rest/v1/projects/create
// IAM permissions requires:
// resourcemanager.projects.get
// resourcemanager.projects.list
// resourcemanager.projects.create
// resourcemanager.projects.delete (optional)
// resourcemanager.projects.undelete
func (g *Gcp) createProjectWithId(kt *kit.Kit, projectId, projectName string, organization string) (string, error) {
	client, err := g.clientSet.resClient(kt)
	if err != nil {
		logs.Errorf("init gcp client failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if organization == "" {
		return "", fmt.Errorf("organization is required when create project through gcp apis")
	}

	// build project to be created
	project := &resourcemanager.Project{
		ProjectId:   projectId,
		DisplayName: projectName,
		Labels:      map[string]string{},
		// organization is requeired when create project through gcp apis
		Parent: ensureOrganizationPrefix(organization),
	}

	// create project
	operation, err := client.Projects.Create(project).Context(kt.Ctx).Do()
	if err != nil {
		// Note: 传入的项目ID可能已经存在
		// 如果已存在，则添加随机数后再次创建
		if gErr, ok := err.(*googleapi.Error); ok &&
			gErr.Code == ProjectIdExistedErrorCode &&
			gErr.Message == ProjectIdExistedErrorMessage &&
			!isRandomSuffix(projectId) {

			logs.Infof("project id existed, projectId: %s, add random suffix and retry, rid: %s", projectId, kt.Rid)
			newProjectId := addRandomSuffix(projectId)
			return g.createProjectWithId(kt, newProjectId, projectName, organization)
		}
		logs.Errorf("create project error: %s, projectId: %s, projectName: %s, rid: %s",
			err.Error(), projectId, projectName, kt.Rid)
		return "", err
	}

	handler := &createMainAccountPollingHandler{}
	respPoller := poller.Poller[*Gcp, *resourcemanager.Operation, resourcemanager.Operation]{Handler: handler}
	result, err := respPoller.PollUntilDone(g, kt, []*string{&operation.Name},
		types.NewCreateMainAccountPollerOption())
	if err != nil {
		return "", err
	}

	if result.Error != nil {
		return "", fmt.Errorf("create project failed, err: %v, rid: %s", *result.Error, kt.Rid)
	}

	_, err = client.Projects.Get(ensureProjectPrefix(projectId)).Context(kt.Ctx).Do()
	if err != nil {
		return "", err
	}

	return projectId, nil
}

// UpdateBillingInfo update project billing info.
// reference: https://cloud.google.com/billing/docs/reference/rest/v1/projects/updateBillingInfo
// IAM permissions requires (https://cloud.google.com/billing/docs/how-to/modify-project?hl=zh-cn):
// billing.resourceAssociations.list
// billing.resourceAssociations.create
// billing.resourceAssociations.delete
// resourcemanager.projects.get
// resourcemanager.projects.createBillingAssignment
// resourcemanager.projects.deleteBillingAssignment
func (g *Gcp) UpdateBillingInfo(kt *kit.Kit, projectId string, billingAccountName string) error {
	client, err := g.clientSet.billingClient(kt)
	if err != nil {
		logs.Errorf("init gcp client failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = client.Projects.UpdateBillingInfo(
		ensureProjectPrefix(projectId),
		&cloudbilling.ProjectBillingInfo{
			BillingAccountName: ensureBillingAccountPrefix(billingAccountName),
		},
	).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("update billing info error: %s, projectId: %s, billingAccountName: %s, rid: %s", err.Error(), projectId, billingAccountName, kt.Rid)
		return err
	}

	return nil
}

// BindingProjectEditor set iam policy.
// reference: https://cloud.google.com/resource-manager/reference/rest/v3/projects/setIamPolicy
func (g *Gcp) BindingProjectEditor(kt *kit.Kit, projectId, email string) error {
	client, err := g.clientSet.resClient(kt)
	if err != nil {
		logs.Errorf("init gcp client failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	policy, err := client.Projects.GetIamPolicy(
		ensureProjectPrefix(projectId),
		&resourcemanager.GetIamPolicyRequest{},
	).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("error getting IAM policy: %v", err)
		return err
	}

	policy.Bindings = append(policy.Bindings,
		&resourcemanager.Binding{
			Role:    string(iam.Editor),
			Members: []string{"user:" + email},
		},
		&resourcemanager.Binding{
			Role:    string(iam.Viewer),
			Members: []string{"user:" + email},
		},
	)

	iamReq := &resourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}

	// do request
	_, err = client.Projects.SetIamPolicy(
		ensureProjectPrefix(projectId),
		iamReq,
	).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("project create success, but binding project owner error: %s, projectId: %s, email: %s, rid: %s", err.Error(), projectId, email, kt.Rid)
		return err
	}

	return nil
}

type createMainAccountPollingHandler struct {
}

// Done ...
func (h *createMainAccountPollingHandler) Done(op *resourcemanager.Operation) (bool, *resourcemanager.Operation) {
	// Note: 没有error的情况分两种, 一种是在创建中，创建中不结束Poll，返回false，另一种是创建成功，返回true
	if op.Done {
		return true, op
	}
	return false, op
}

// Poll ...
func (h *createMainAccountPollingHandler) Poll(client *Gcp, kt *kit.Kit, opIds []*string) (*resourcemanager.Operation, error) {
	if len(opIds) == 0 {
		return nil, fmt.Errorf("operation group id is required")
	}

	opId := opIds[0]

	resClient, err := client.clientSet.resClient(kt)
	if err != nil {
		return nil, err
	}

	op, err := resClient.Operations.Get(*opId).Context(kt.Ctx).Do()
	if err != nil {
		logs.Errorf("get operation failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return nil, err
	}

	return op, nil
}
