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

// Package sys ...
package sys

import (
	"reflect"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/iam"
)

// Sys iam system related operate.
type Sys struct {
	client iam.Client
}

// NewSys create sys to iam sys related operate.
func NewSys(client iam.Client) (*Sys, error) {
	if client == nil {
		return nil, errf.New(errf.InvalidParameter, "client is nil")
	}

	sys := &Sys{
		client: client,
	}
	return sys, nil
}

// GetSystemToken get system token from iam, used to validate if request is from iam.
func (s *Sys) GetSystemToken(kt *kit.Kit) (string, error) {
	return s.client.GetSystemToken(kt)
}

/**
1. 资源间的依赖关系为 Action 依赖 InstanceSelection 依赖 ResourceType，对资源的增删改操作需要按照这个依赖顺序调整
2. ActionGroup、ResCreatorAction、CommonAction 依赖于 Action，这些资源的增删操作始终放在最后
3. 因为资源的名称在系统中是唯一的，所以可能遇到循环依赖的情况（如两个资源分别更新成对方的名字），此时需要引入一个中间变量进行二次更新

综上，具体操作顺序如下：
	1. 注册hcm系统信息
	2. 删除Action。该操作无依赖
	3. 更新ResourceType，先更新名字冲突的(包括需要删除的)为中间值，再更新其它的。该操作无依赖
	4. 新增ResourceType。该操作依赖于上一步中同名的ResourceType均已更新
	5. 更新InstanceSelection，先更新名字冲突的(包括需要删除的)为中间值，再更新其它的。该操作依赖于上一步中的ResourceType均已新增
	6. 新增InstanceSelection。该操作依赖于上一步中同名的InstanceSelection均已更新+第4步中的ResourceType均已新增
	7. 更新ResourceAction，先更新名字冲突的为中间值，再更新其它的。该操作依赖于第2步中同名Action已删除+上一步中InstanceSelection已新增
	8. 新增ResourceAction。该操作依赖于上一步中同名的ResourceAction均已更新+第6步中的InstanceSelection均已新增
	9. 删除InstanceSelection。该操作依赖于第2步和第7步中的原本依赖了这些InstanceSelection的Action均已删除和更新
	10. 删除ResourceType。该操作依赖于第5步和第9步中的原本依赖了这些ResourceType的InstanceSelection均已删除和更新
	11. 注册ActionGroup、ResCreatorAction、CommonAction信息
*/

// Register register auth model to iam.
func (s *Sys) Register(kt *kit.Kit, host string) error {
	system, err := s.registerSystem(kt, host)
	if err != nil {
		return err
	}

	newResTypes, updateResTypes, removedResTypeIDs := s.crossCompareResTypes(system.ResourceTypes)
	newInstSelections, updateInstSelections, removedInstSelectionIDs := s.crossCompareInstSelections(
		system.InstanceSelections)
	newResActions, updateResActions, removedResActionIDs := s.crossCompareResActions(system.Actions)

	if err = s.removeResActions(kt, removedResActionIDs); err != nil {
		return err
	}

	for _, resourceType := range updateResTypes {
		if err = s.client.UpdateResourcesType(kt, resourceType); err != nil {
			logs.Errorf("update resource type(%+v) failed, err: %v", resourceType, err)
			return err
		}
	}

	if err = s.client.RegisterResourcesTypes(kt, newResTypes); err != nil {
		logs.Errorf("register resource types(%+v) failed, err: %v", newResTypes, err)
		return err
	}

	for _, instanceSelection := range updateInstSelections {
		if err = s.client.UpdateInstanceSelection(kt, instanceSelection); err != nil {
			logs.Errorf("update instance selection(%+v) failed, err: %v", instanceSelection, err)
			return err
		}
	}

	if err = s.client.RegisterInstanceSelections(kt, newInstSelections); err != nil {
		logs.Errorf("register instance selections(%+v) failed, err: %v", newInstSelections, err)
		return err
	}

	for _, resourceAction := range updateResActions {
		if err = s.client.UpdateAction(kt, resourceAction); err != nil {
			logs.Errorf("update resource action(%+v) failed, err: %v", resourceAction, err)
			return err
		}
	}

	if err = s.client.RegisterActions(kt, newResActions); err != nil {
		logs.Errorf("register resource actions(%+v) failed, err: %v", newResActions, err)
		return err
	}

	if err = s.client.DeleteInstanceSelections(kt, removedInstSelectionIDs); err != nil {
		logs.Errorf("delete instance selections(%+v) failed, err: %v", removedInstSelectionIDs, err)
		return err
	}

	if err = s.client.DeleteResourcesTypes(kt, removedResTypeIDs); err != nil {
		logs.Errorf("delete resource types(%+v) failed, err: %v", removedResTypeIDs, err)
		return err
	}

	if err := s.registerActionGroups(kt, system); err != nil {
		return err
	}

	if err := s.registerResCreatorActions(kt, system); err != nil {
		return err
	}

	if err := s.registerCommonActions(kt, system); err != nil {
		return err
	}
	return nil
}

// registerSystem register or update system to iam.
func (s *Sys) registerSystem(kt *kit.Kit, host string) (*iam.RegisteredSystemInfo, error) {
	resp, err := s.client.GetSystemInfo(kt, []iam.SystemQueryField{})
	if err != nil && err != iam.ErrNotFound {
		logs.Errorf("get system info failed, err: %v", err)
		return nil, err
	}

	sys := iam.System{
		ID:          SystemIDHCM,
		Name:        SystemNameHCM,
		EnglishName: SystemNameHCMEn,
		Clients:     SystemIDHCM,
		ProviderConfig: &iam.SysConfig{
			Host: host,
			Auth: "basic",
		},
	}

	// if iam hcm system has not been registered, register system
	if err == iam.ErrNotFound {
		if err = s.client.RegisterSystem(kt, &sys); err != nil {
			logs.Errorf("register system failed, system: %v, err: %v", sys, err)
			return nil, err
		}

		if logs.V(5) {
			logs.Infof("register new system succeed, system: %v", sys)
		}

		return &resp.Data, nil
	}

	// update system config
	if err = s.client.UpdateSystemConfig(kt, sys); err != nil {
		logs.Errorf("update system host config failed, host: %s, err: %v", host, err)
		return nil, err
	}

	if logs.V(5) {
		logs.Infof("update system succeed, old: %+v, new: %+v", resp.Data.BaseInfo, sys)
	}

	return &resp.Data, nil
}

// iamName record iam name and english name to find if name conflicts
type iamName struct {
	Name   string
	NameEn string
}

// crossCompareResTypes cross compare resource types to get need create/update/delete ones
func (s *Sys) crossCompareResTypes(registeredResourceTypes []iam.ResourceType) (
	[]iam.ResourceType, []iam.ResourceType, []iam.TypeID) {

	registeredResTypeMap := make(map[iam.TypeID]iam.ResourceType)
	for _, resourceType := range registeredResourceTypes {
		registeredResTypeMap[resourceType.ID] = resourceType
	}

	// record the name and resource type id mapping to get the resource types whose name conflicts
	resNameMap, resNameEnMap := make(map[string]iam.TypeID), make(map[string]iam.TypeID)
	updateResPrevNameMap := make(map[iam.TypeID]iamName)
	newResTypes, updateResTypes := make([]iam.ResourceType, 0), make([]iam.ResourceType, 0)

	for _, resourceType := range GenerateStaticResourceTypes() {
		resNameMap[resourceType.Name] = resourceType.ID
		resNameEnMap[resourceType.NameEn] = resourceType.ID
		// if current resource type is not registered, register it, otherwise, update it if its version is changed
		registeredResType, exists := registeredResTypeMap[resourceType.ID]
		if exists {
			// registered resource type exists in current resource types, should not be removed
			delete(registeredResTypeMap, resourceType.ID)
			if s.compareResType(registeredResType, resourceType) {
				continue
			}
			updateResPrevNameMap[resourceType.ID] = iamName{Name: registeredResType.Name,
				NameEn: registeredResType.NameEn}
			updateResTypes = append(updateResTypes, resourceType)
			continue
		}

		newResTypes = append(newResTypes, resourceType)
	}

	// if to update resource type previous name conflict with a valid one, change its name to an intermediate one first
	conflictResTypes := make([]iam.ResourceType, 0)
	for _, updateResType := range updateResTypes {
		prevName := updateResPrevNameMap[updateResType.ID]
		isConflict := false

		if resNameMap[prevName.Name] != updateResType.ID {
			isConflict = true
			updateResType.Name = prevName.Name + "_"
		}

		if resNameEnMap[prevName.NameEn] != updateResType.ID {
			isConflict = true
			updateResType.NameEn = prevName.NameEn + "_"
		}
		if isConflict {
			conflictResTypes = append(conflictResTypes, updateResType)
		}
	}

	// remove the resource types that are not exist in new resource types
	removedResTypeIDs := make([]iam.TypeID, len(registeredResTypeMap))
	idx := 0
	for resTypeID, resType := range registeredResTypeMap {
		removedResTypeIDs[idx] = resTypeID
		idx++
		// if to remove resource type name conflicts with a valid one, change its name to an intermediate one first
		isConflict := false

		if _, exists := resNameMap[resType.Name]; exists {
			resType.Name += "_"
			isConflict = true
		}
		if _, exists := resNameEnMap[resType.NameEn]; exists {
			resType.NameEn += "_"
			isConflict = true
		}
		if isConflict {
			conflictResTypes = append(conflictResTypes, resType)
		}
	}

	return newResTypes, append(conflictResTypes, updateResTypes...), removedResTypeIDs
}

// compareResType compare if registered resource type that iam returns is the same with the new resource type
func (s *Sys) compareResType(registeredResType, resType iam.ResourceType) bool {
	if registeredResType.ID != resType.ID ||
		registeredResType.Name != resType.Name ||
		registeredResType.NameEn != resType.NameEn ||
		registeredResType.Description != resType.Description ||
		registeredResType.DescriptionEn != resType.DescriptionEn ||
		registeredResType.Version < resType.Version ||
		registeredResType.ProviderConfig.Path != resType.ProviderConfig.Path {
		return false
	}

	if len(registeredResType.Parents) != len(resType.Parents) {
		return false
	}
	for idx, parent := range registeredResType.Parents {
		resTypeParent := resType.Parents[idx]
		if parent.ResourceID != resTypeParent.ResourceID || parent.SystemID != resTypeParent.SystemID {
			return false
		}
	}

	return true
}

// crossCompareInstSelections cross compare instance selections to get need create/update/delete ones
func (s *Sys) crossCompareInstSelections(registeredInstanceSelections []iam.InstanceSelection) (
	[]iam.InstanceSelection, []iam.InstanceSelection, []iam.InstanceSelectionID) {

	registeredInstSelectionMap := make(map[iam.InstanceSelectionID]iam.InstanceSelection)
	for _, instanceSelection := range registeredInstanceSelections {
		registeredInstSelectionMap[instanceSelection.ID] = instanceSelection
	}

	// record the name and instance selection id mapping to get the instance selections whose name conflicts
	selectionNameMap := make(map[string]iam.InstanceSelectionID)
	selectionNameEnMap := make(map[string]iam.InstanceSelectionID)
	updateSelectionPrevNameMap := make(map[iam.InstanceSelectionID]iamName)

	newInstSelections, updateInstSelections := make([]iam.InstanceSelection, 0), make([]iam.InstanceSelection, 0)

	for _, instanceSelection := range GenerateStaticInstanceSelections() {
		selectionNameMap[instanceSelection.Name] = instanceSelection.ID
		selectionNameEnMap[instanceSelection.NameEn] = instanceSelection.ID

		selection, exists := registeredInstSelectionMap[instanceSelection.ID]

		// if current instance selection is not registered, register it, otherwise, update it if it is changed
		if exists {
			// registered instance selection exists in current instance selections, should not be removed
			delete(registeredInstSelectionMap, instanceSelection.ID)

			if reflect.DeepEqual(selection, instanceSelection) {
				continue
			}

			updateSelectionPrevNameMap[instanceSelection.ID] = iamName{Name: selection.Name, NameEn: selection.NameEn}
			updateInstSelections = append(updateInstSelections, instanceSelection)
			continue
		}
		newInstSelections = append(newInstSelections, instanceSelection)
	}

	// if to update selection previous name conflict with a valid one, change its name to an intermediate one first
	conflictSelections := make([]iam.InstanceSelection, 0)
	for _, updateSelection := range updateInstSelections {
		prevName := updateSelectionPrevNameMap[updateSelection.ID]
		isConflict := false

		if selectionNameMap[prevName.Name] != updateSelection.ID {
			updateSelection.Name = prevName.Name + "_"
			isConflict = true
		}

		if selectionNameEnMap[prevName.NameEn] != updateSelection.ID {
			updateSelection.NameEn = prevName.NameEn + "_"
			isConflict = true
		}

		if isConflict {
			conflictSelections = append(conflictSelections, updateSelection)
		}
	}

	// remove the resource types that are not exist in new resource types
	removedInstSelectionIDs := make([]iam.InstanceSelectionID, len(registeredInstSelectionMap))
	idx := 0
	for selectionID, selection := range registeredInstSelectionMap {
		removedInstSelectionIDs[idx] = selectionID
		idx++
		// if to remove selection name conflicts with a valid one, change its name to an intermediate one first
		isConflict := false
		if _, exists := selectionNameMap[selection.Name]; exists {
			selection.Name += "_"
			isConflict = true
		}
		if _, exists := selectionNameEnMap[selection.NameEn]; exists {
			selection.NameEn += "_"
			isConflict = true
		}
		if isConflict {
			conflictSelections = append(conflictSelections, selection)
		}
	}
	return newInstSelections, append(conflictSelections, updateInstSelections...), removedInstSelectionIDs
}

// crossCompareResActions cross compare resource actions to get need create/update/delete ones
func (s *Sys) crossCompareResActions(registeredActions []iam.ResourceAction) (
	[]iam.ResourceAction, []iam.ResourceAction, []iam.ActionID) {

	registeredResActionMap := make(map[iam.ActionID]iam.ResourceAction)
	for _, resourceAction := range registeredActions {
		registeredResActionMap[resourceAction.ID] = resourceAction
	}

	// record the name and resource action id mapping to get the instance selections whose name conflicts
	actionNameMap := make(map[string]iam.ActionID)
	actionNameEnMap := make(map[string]iam.ActionID)
	updateActionPrevNameMap := make(map[iam.ActionID]iamName)

	newResActions := make([]iam.ResourceAction, 0)
	updateResActions := make([]iam.ResourceAction, 0)

	for _, resourceAction := range GenerateStaticActions() {
		actionNameMap[resourceAction.Name] = resourceAction.ID
		actionNameEnMap[resourceAction.NameEn] = resourceAction.ID

		// if current resource action is not registered, register it, otherwise, update it if its version is changed
		action, exists := registeredResActionMap[resourceAction.ID]
		if exists {
			// registered resource action exist in current resource actions, should not be removed
			delete(registeredResActionMap, resourceAction.ID)

			if s.compareResAction(action, resourceAction) {
				continue
			}

			updateActionPrevNameMap[action.ID] = iamName{
				Name:   action.Name,
				NameEn: action.NameEn,
			}
			updateResActions = append(updateResActions, resourceAction)
			continue
		}
		newResActions = append(newResActions, resourceAction)
	}

	// if to update action previous name conflict with a valid one, change its name to an intermediate one first
	conflictActions := make([]iam.ResourceAction, 0)
	for _, updateAction := range updateResActions {
		prevName := updateActionPrevNameMap[updateAction.ID]
		isConflict := false

		if actionNameMap[prevName.Name] != updateAction.ID {
			updateAction.Name = prevName.Name + "_"
			isConflict = true
		}

		if actionNameEnMap[prevName.NameEn] != updateAction.ID {
			updateAction.NameEn = prevName.NameEn + "_"
			isConflict = true
		}

		if isConflict {
			conflictActions = append(conflictActions, updateAction)
		}
	}

	removedResActionIDs := make([]iam.ActionID, len(registeredResActionMap))
	idx := 0
	for resourceActionID := range registeredResActionMap {
		removedResActionIDs[idx] = resourceActionID
		idx++
	}

	return newResActions, append(conflictActions, updateResActions...), removedResActionIDs
}

// compareResAction compare if registered resource action that iam returns is the same with the new resource action
func (s *Sys) compareResAction(registeredAction, action iam.ResourceAction) bool {
	if registeredAction.ID != action.ID ||
		registeredAction.Name != action.Name ||
		registeredAction.NameEn != action.NameEn ||
		registeredAction.Type != action.Type ||
		registeredAction.Version < action.Version {
		return false
	}

	if len(registeredAction.RelatedResourceTypes) != len(action.RelatedResourceTypes) {
		return false
	}

	for idx, registeredResType := range registeredAction.RelatedResourceTypes {
		if !s.compareRelatedResType(registeredResType, action.RelatedResourceTypes[idx]) {
			return false
		}
	}

	if len(registeredAction.RelatedActions) != len(action.RelatedActions) {
		return false
	}

	for idx, actionID := range registeredAction.RelatedActions {
		if actionID != action.RelatedActions[idx] {
			return false
		}
	}

	return true
}

// compareRelatedResType compare if registered related resource type that iam returns is the same with the new one
func (s *Sys) compareRelatedResType(registeredResType, resType iam.RelateResourceType) bool {
	// iam default selection mode is "instance"
	if resType.SelectionMode == "" {
		resType.SelectionMode = iam.ModeInstance
	}

	if registeredResType.ID != resType.ID || registeredResType.SelectionMode != resType.SelectionMode {
		return false
	}

	if registeredResType.Scope == nil && resType.Scope == nil {
		return true
	}

	if registeredResType.Scope == nil && resType.Scope != nil ||
		registeredResType.Scope != nil && resType.Scope == nil {
		return false
	}

	if registeredResType.Scope.Op != resType.Scope.Op {
		return false
	}

	if len(registeredResType.Scope.Content) != len(resType.Scope.Content) {
		return false
	}

	for index, registeredContent := range registeredResType.Scope.Content {
		content := resType.Scope.Content[index]
		if registeredContent.Op != content.Op || registeredContent.Value != content.Value ||
			registeredContent.Field != content.Field {
			return false
		}
	}

	// since iam returns no related selections & we use matching type & selection, skip this comparison
	return true
}

// removeResActions remove resource actions and related policies
func (s *Sys) removeResActions(kt *kit.Kit, actionIDs []iam.ActionID) error {
	if len(actionIDs) == 0 {
		return nil
	}

	// before deleting action, the dependent action policies must be deleted
	for _, resourceActionID := range actionIDs {
		if err := s.client.DeleteActionPolicies(kt, resourceActionID); err != nil {
			logs.Errorf("delete action %s policies failed, err: %v", resourceActionID, err)
			return err
		}
	}

	if err := s.client.DeleteActions(kt, actionIDs); err != nil {
		logs.Errorf("delete resource actions(%+v) failed, err: %v", actionIDs, err)
		return err
	}

	return nil
}

// registerActionGroups register or update resource action groups
func (s *Sys) registerActionGroups(kt *kit.Kit, system *iam.RegisteredSystemInfo) error {
	actionGroups := GenerateStaticActionGroups()

	if len(system.ActionGroups) == 0 {
		if len(actionGroups) == 0 {
			return nil
		}

		if err := s.client.RegisterActionGroups(kt, actionGroups); err != nil {
			logs.Errorf("register action groups(%+v) failed, err: %v", actionGroups, err)
			return err
		}
		return nil
	}

	if reflect.DeepEqual(system.ActionGroups, actionGroups) {
		return nil
	}

	if len(actionGroups) == 0 {
		logs.Warnf("action groups can not be updated to empty, update to one")
		actionGroups = system.ActionGroups[:1]
	}

	if err := s.client.UpdateActionGroups(kt, actionGroups); err != nil {
		logs.Errorf("update action groups(%+v) failed, err: %v", actionGroups, err)
		return err
	}
	return nil
}

// registerResCreatorActions register or update resource creator actions
func (s *Sys) registerResCreatorActions(kt *kit.Kit, system *iam.RegisteredSystemInfo) error {
	rcActions := GenerateResourceCreatorActions()

	if len(system.ResourceCreatorActions.Config) == 0 {
		if len(rcActions.Config) == 0 {
			return nil
		}
		if err := s.client.RegisterResourceCreatorActions(kt, rcActions); err != nil {
			logs.Errorf("register resource creator actions(%+v) failed, err: %v", rcActions, err)
			return err
		}
		return nil
	}

	if reflect.DeepEqual(system.ResourceCreatorActions, rcActions) {
		return nil
	}

	if len(rcActions.Config) == 0 {
		logs.Warnf("resource creator actions can not be updated to empty, update to one")
		rcActions.Config = system.ResourceCreatorActions.Config[:1]
	}

	if err := s.client.UpdateResourceCreatorActions(kt, rcActions); err != nil {
		logs.Errorf("update resource creator actions(%+v) failed, err: %v", rcActions, err)
		return err
	}
	return nil
}

// registerCommonActions register or update common actions
func (s *Sys) registerCommonActions(kt *kit.Kit, system *iam.RegisteredSystemInfo) error {
	commonActions := GenerateCommonActions()

	if len(system.CommonActions) == 0 {
		if len(commonActions) == 0 {
			return nil
		}
		if err := s.client.RegisterCommonActions(kt, commonActions); err != nil {
			logs.Errorf("register common actions(%+v) failed, err: %v", commonActions, err)
			return err
		}
		return nil
	}

	if reflect.DeepEqual(system.CommonActions, commonActions) {
		return nil
	}

	if len(commonActions) == 0 {
		logs.Warnf("common actions can not be updated to empty, update to one")
		commonActions = system.CommonActions[:1]
	}

	if err := s.client.UpdateCommonActions(kt, commonActions); err != nil {
		logs.Errorf("update common actions(%+v) failed, err: %v", commonActions, err)
		return err
	}
	return nil
}
