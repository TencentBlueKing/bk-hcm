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

package cvmrelmgr

import (
	"fmt"
	"strings"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/api/core"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// NewCvmRelManager new cvm rel manger.
func NewCvmRelManager(dataCli *dataservice.Client) *CvmRelManger {
	return &CvmRelManger{
		cvmAssResMap:           make(map[string]map[enumor.CloudResourceType][]string),
		assResWithParentResMap: make(map[enumor.CloudResourceType]map[string] /*parentID*/ []string),
		dataCli:                dataCli,
	}
}

// CvmRelManger 为了同步主机和关联资源而定义的关联关系存储结构
type CvmRelManger struct {
	cvmAssResMap           map[string] /*cvmCloudID*/ map[enumor.CloudResourceType][]string /*cvmAssResCloudID*/
	assResWithParentResMap map[enumor.CloudResourceType]map[string] /*parentID*/ []string
	dataCli                *dataservice.Client
}

// SyncRelOption ...
type SyncRelOption struct {
	Vendor  enumor.Vendor            `json:"vendor" validate:"required"`
	ResType enumor.CloudResourceType `json:"res_type" validate:"required"`
}

// Validate ...
func (opt SyncRelOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// SyncRel 同步主机和关联资源关系表的关联关系
func (mgr *CvmRelManger) SyncRel(kt *kit.Kit, opt *SyncRelOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	if err := mgr.validateSyncRelParams(opt.Vendor, opt.ResType); err != nil {
		return err
	}

	cvmMap, err := mgr.getCvmMap(kt)
	if err != nil {
		logs.Errorf("get cvm map failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	switch opt.ResType {
	case enumor.DiskCloudResType:
		err = mgr.syncCvmDiskRel(kt, cvmMap, opt)
	case enumor.SecurityGroupCloudResType:
		err = mgr.syncCvmSGRel(kt, cvmMap, opt)
	case enumor.EipCloudResType:
		err = mgr.syncCvmEipRel(kt, cvmMap, opt)
	case enumor.NetworkInterfaceCloudResType:
		err = mgr.syncCvmNetworkInterfaceRel(kt, cvmMap, opt)
	}
	if err != nil {
		logs.Errorf("sync cvm_%s_rel failed, err: %v, rid: %s", opt.ResType, err, kt.Rid)
		return err
	}

	return nil
}

// Sync resource framework.
func (mgr *CvmRelManger) Sync(kt *kit.Kit, resType enumor.CloudResourceType,
	syncFunc func(kt *kit.Kit, cloudIDs []string) error) error {

	cloudIDs := make([]string, 0)
	if resType == enumor.CvmCloudResType {
		for cvmCloudID := range mgr.cvmAssResMap {
			cloudIDs = append(cloudIDs, cvmCloudID)
		}
	} else {
		cloudIDs = mgr.getAllCvmAssResCloudIDs(resType)
	}

	if len(cloudIDs) == 0 {
		return nil
	}

	split := slice.Split(cloudIDs, constant.CloudResourceSyncMaxLimit)
	for _, partCloudIDs := range split {
		if err := syncFunc(kt, partCloudIDs); err != nil {
			return err
		}
	}

	return nil
}

// SyncForAzure resource framework.
func (mgr *CvmRelManger) SyncForAzure(kt *kit.Kit, resType enumor.CloudResourceType,
	syncFunc func(kt *kit.Kit, resGroupName string, cloudIDs []string) error) error {

	cloudIDs := make([]string, 0)
	if resType == enumor.CvmCloudResType {
		for cvmCloudID := range mgr.cvmAssResMap {
			cloudIDs = append(cloudIDs, cvmCloudID)
		}
	} else {
		for _, resCloudIDMap := range mgr.cvmAssResMap {
			tmp, exist := resCloudIDMap[resType]
			if !exist {
				continue
			}

			cloudIDs = append(cloudIDs, tmp...)
		}
	}

	if len(cloudIDs) == 0 {
		return nil
	}

	resGroupNameMap := common.CloudIDClassByResGroupName(cloudIDs)

	for resGroupName, ids := range resGroupNameMap {
		split := slice.Split(ids, constant.CloudResourceSyncMaxLimit)
		for _, partCloudIDs := range split {
			if err := syncFunc(kt, resGroupName, partCloudIDs); err != nil {
				return err
			}
		}
	}

	return nil
}

// SyncDependParentRes 同步依赖父资源资源
func (mgr *CvmRelManger) SyncDependParentRes(kt *kit.Kit, resType enumor.CloudResourceType,
	syncFunc func(kt *kit.Kit, parentID string, cloudIDs []string) error) error {

	relMap, exist := mgr.assResWithParentResMap[resType]
	if !exist {
		return nil
	}

	for parentID, childIDs := range relMap {
		split := slice.Split(childIDs, constant.CloudResourceSyncMaxLimit)
		for _, partCloudIDs := range split {
			if err := syncFunc(kt, parentID, partCloudIDs); err != nil {
				return err
			}
		}
	}

	return nil
}

// SyncDependParentResForAzure 同步依赖父资源资源
func (mgr *CvmRelManger) SyncDependParentResForAzure(kt *kit.Kit, resType enumor.CloudResourceType,
	syncFunc func(kt *kit.Kit, resGroupName, parentID string, cloudIDs []string) error) error {

	relMap, exist := mgr.assResWithParentResMap[resType]
	if !exist {
		return nil
	}

	for parentID, childIDs := range relMap {
		tmp := parentID[strings.Index(parentID, "resourcegroups/")+15:]
		resGroupName := tmp[:strings.Index(tmp, "/")]

		split := slice.Split(childIDs, constant.CloudResourceSyncMaxLimit)
		for _, partCloudIDs := range split {
			if err := syncFunc(kt, resGroupName, parentID, partCloudIDs); err != nil {
				return err
			}
		}
	}

	return nil
}

// CvmAppendAssResCloudID 指定主机添加关联资源的云ID
func (mgr *CvmRelManger) CvmAppendAssResCloudID(cvmCloudID string,
	assResType enumor.CloudResourceType, assResCloudID string) {

	if _, exist := mgr.cvmAssResMap[cvmCloudID]; !exist {
		mgr.cvmAssResMap[cvmCloudID] = make(map[enumor.CloudResourceType][]string)
	}

	if _, exist := mgr.cvmAssResMap[cvmCloudID][assResType]; !exist {
		mgr.cvmAssResMap[cvmCloudID][assResType] = make([]string, 0)
	}

	// 校验关联关系是否已经建立，如果建立不再重复录入
	tmpMap := converter.StringSliceToMap(mgr.cvmAssResMap[cvmCloudID][assResType])
	if _, exist := tmpMap[assResCloudID]; exist {
		return
	}

	mgr.cvmAssResMap[cvmCloudID][assResType] = append(mgr.cvmAssResMap[cvmCloudID][assResType], assResCloudID)

	return
}

// AddAssParentWithChildRes 添加关联资源的父子资源关系，因为有部分子资源同步依赖父资源
func (mgr *CvmRelManger) AddAssParentWithChildRes(assResType enumor.CloudResourceType, relMap map[string][]string) {

	// 对关系进行去重
	for key, ids := range relMap {
		relMap[key] = slice.Unique(ids)
	}

	mgr.assResWithParentResMap[assResType] = relMap

	return
}

func (mgr *CvmRelManger) getAllCvmAssResCloudIDs(resType enumor.CloudResourceType) []string {
	cloudIDs := make([]string, 0)
	for _, resCloudIDMap := range mgr.cvmAssResMap {
		tmp, exist := resCloudIDMap[resType]
		if !exist {
			continue
		}

		cloudIDs = append(cloudIDs, tmp...)
	}

	return cloudIDs
}

type cvmRelInfo struct {
	// RelID 关系表
	RelID    uint64
	AssResID string
	CvmID    string
}

func (mgr *CvmRelManger) getCvmIDWithAssResIDMap(resType enumor.CloudResourceType,
	cvmMap, assResMap map[string]string) ([]string, map[string][]string, error) {

	result := make(map[string][]string)
	cvmIDs := make([]string, 0, len(mgr.cvmAssResMap))
	for cvmCloudID, valueMap := range mgr.cvmAssResMap {
		cvmID, exist := cvmMap[cvmCloudID]
		if !exist {
			return nil, nil, fmt.Errorf("cvm: %s not found", cvmCloudID)
		}

		cvmIDs = append(cvmIDs, cvmID)
		result[cvmID] = make([]string, 0)

		assResCloudIDs, exist := valueMap[resType]
		if !exist {
			continue
		}

		for _, one := range assResCloudIDs {
			id, exist := assResMap[one]
			if !exist {
				return nil, nil, fmt.Errorf("%s: %s not found", resType, one)
			}

			result[cvmID] = append(result[cvmID], id)
		}
	}

	return cvmIDs, result, nil
}

// validateSyncRelParams 不同vendor的主机可关联的资源有所不同.
func (mgr *CvmRelManger) validateSyncRelParams(vendor enumor.Vendor, resType enumor.CloudResourceType) error {
	switch vendor {
	case enumor.TCloud:
		switch resType {
		case enumor.SecurityGroupCloudResType, enumor.DiskCloudResType, enumor.EipCloudResType:
		default:
			return fmt.Errorf("vendor: %s cvm and %s are not associated", vendor, resType)
		}
	case enumor.Aws:
		switch resType {
		case enumor.SecurityGroupCloudResType, enumor.DiskCloudResType, enumor.EipCloudResType:
		default:
			return fmt.Errorf("vendor: %s cvm and %s are not associated", vendor, resType)
		}
	case enumor.Gcp:
		switch resType {
		case enumor.DiskCloudResType, enumor.EipCloudResType, enumor.NetworkInterfaceCloudResType:
		default:
			return fmt.Errorf("vendor: %s cvm and %s are not associated", vendor, resType)
		}
	case enumor.HuaWei:
		switch resType {
		case enumor.SecurityGroupCloudResType, enumor.DiskCloudResType, enumor.EipCloudResType,
			enumor.NetworkInterfaceCloudResType:
		default:
			return fmt.Errorf("vendor: %s cvm and %s are not associated", vendor, resType)
		}
	case enumor.Azure:
		switch resType {
		case enumor.SecurityGroupCloudResType, enumor.DiskCloudResType, enumor.EipCloudResType,
			enumor.NetworkInterfaceCloudResType:
		default:
			return fmt.Errorf("vendor: %s cvm and %s are not associated", vendor, resType)
		}
	default:
		return fmt.Errorf("vendor: %s not support", vendor)
	}

	return nil
}

func (mgr *CvmRelManger) getCvmMap(kt *kit.Kit) (map[string]string, error) {
	cloudIDs := make([]string, 0, len(mgr.cvmAssResMap))
	for key := range mgr.cvmAssResMap {
		cloudIDs = append(cloudIDs, key)
	}

	cvmMap := make(map[string]string)
	split := slice.Split(cloudIDs, int(core.DefaultMaxPageLimit))
	for _, partCloudIDs := range split {
		req := &core.ListReq{
			Fields: []string{"id", "cloud_id"},
			Filter: tools.ContainersExpression("cloud_id", partCloudIDs),
			Page:   core.NewDefaultBasePage(),
		}
		result, err := mgr.dataCli.Global.Cvm.ListCvm(kt, req)
		if err != nil {
			logs.Errorf("list cvm failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		for _, one := range result.Details {
			cvmMap[one.CloudID] = one.ID
		}
	}

	return cvmMap, nil
}
