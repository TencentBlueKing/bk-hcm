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

package azure

import (
	"fmt"

	"hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/adaptor/types/disk"
	typedisk "hcm/pkg/adaptor/types/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// CreateDisk 创建云硬盘
// reference: https://learn.microsoft.com/en-us/rest/api/compute/disks/list?source=recommendations&tabs=Go#disklist
func (a *Azure) CreateDisk(kt *kit.Kit, opt *disk.AzureDiskCreateOption) ([]string, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "azure disk create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	diskCloudIDs := make([]string, 0)

	if *opt.DiskCount == 1 {
		resp, err := a.createDisk(kt, opt, opt.DiskName)
		if err != nil {
			return nil, err
		}
		diskCloudIDs = append(diskCloudIDs, *resp.ID)
	} else {
		for i := uint64(1); i <= *opt.DiskCount; i++ {
			resp, err := a.createDisk(kt, opt, fmt.Sprintf("%s-%s", opt.DiskName, i))
			if err != nil {
				return nil, err
			}
			diskCloudIDs = append(diskCloudIDs, SPtrToLowerStr(resp.ID))
		}
	}

	return diskCloudIDs, nil
}

func (a *Azure) createDisk(kt *kit.Kit, opt *disk.AzureDiskCreateOption, diskName string) (*armcompute.Disk, error) {
	client, err := a.clientSet.diskClient()
	if err != nil {
		return nil, err
	}

	diskReq, err := opt.ToCreateDiskRequest()
	if err != nil {
		return nil, err
	}

	pollerResp, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, diskName, *diskReq, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Disk, nil
}

// GetDisk 查询单个云盘
// reference: https://learn.microsoft.com/en-us/rest/api/compute/disks/get?tabs=Go
func (a *Azure) GetDisk(kt *kit.Kit, opt *disk.AzureDiskGetOption) (*typedisk.AzureDisk, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "azure disk get option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	client, err := a.clientSet.diskClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(kt.Ctx, opt.ResourceGroupName, opt.DiskName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk: %v", err)
	}

	converterResp := &typedisk.AzureDisk{
		ID:       SPtrToLowerSPtr(resp.Disk.ID),
		Name:     SPtrToLowerSPtr(resp.Disk.Name),
		Location: SPtrToLowerNoSpaceSPtr(resp.Disk.Location),
		Type:     resp.Disk.Type,
		Status:   (*string)(resp.Disk.Properties.DiskState),
		DiskSize: resp.Disk.Properties.DiskSizeBytes,
		Zones:    resp.Disk.Zones,
	}

	return converterResp, nil
}

// ListDisk 查看云硬盘
// reference: https://learn.microsoft.com/en-us/rest/api/compute/disks/list?source=recommendations&tabs=Go#disklist
func (a *Azure) ListDisk(kt *kit.Kit, opt *disk.AzureDiskListOption) ([]*typedisk.AzureDisk, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "azure disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := a.clientSet.diskClient()
	if err != nil {
		return nil, err
	}

	disks := []*armcompute.Disk{}
	pager := client.NewListByResourceGroupPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		disks = append(disks, nextResult.Value...)
	}

	return converterDisk(disks), nil
}

// ListDiskByID 查看云硬盘
// reference: https://learn.microsoft.com/en-us/rest/api/compute/disks/list?source=recommendations&tabs=Go#disklist
func (a *Azure) ListDiskByID(kit *kit.Kit, opt *core.AzureListByIDOption) ([]*typedisk.AzureDisk, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "azure disk list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	idMap := converter.StringSliceToMap(opt.CloudIDs)

	client, err := a.clientSet.diskClient()
	if err != nil {
		return nil, err
	}

	disks := make([]*armcompute.Disk, 0, len(idMap))
	pager := client.NewListByResourceGroupPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kit.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}

		for _, one := range nextResult.Value {
			if len(opt.CloudIDs) > 0 {
				id := SPtrToLowerSPtr(one.ID)
				if _, exist := idMap[*id]; exist {
					disks = append(disks, one)
					delete(idMap, *id)

					if len(idMap) == 0 {
						return converterDisk(disks), nil
					}
				}
			} else {
				disks = append(disks, one)
			}
		}
	}

	return converterDisk(disks), nil
}

func converterDisk(disks []*armcompute.Disk) []*typedisk.AzureDisk {
	typesDisk := make([]*typedisk.AzureDisk, 0)

	for _, v := range disks {
		tmp := &typedisk.AzureDisk{
			ID:       SPtrToLowerSPtr(v.ID),
			Name:     SPtrToLowerSPtr(v.Name),
			Location: SPtrToLowerNoSpaceSPtr(v.Location),
			Type:     v.Type,
			Status:   (*string)(v.Properties.DiskState),
			DiskSize: v.Properties.DiskSizeBytes,
			Zones:    v.Zones,
			OSType:   (*string)(v.Properties.OSType),
			SKUName:  (*string)(v.SKU.Name),
			SKUTier:  v.SKU.Tier,
		}
		typesDisk = append(typesDisk, tmp)
	}

	return typesDisk
}

// DeleteDisk 删除云盘
// reference: https://learn.microsoft.com/en-us/rest/api/compute/disks/delete?tabs=Go
func (a *Azure) DeleteDisk(kt *kit.Kit, opt *disk.AzureDiskDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "azure disk delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	client, err := a.clientSet.diskClient()
	if err != nil {
		return err
	}

	pollerResp, err := client.BeginDelete(kt.Ctx, opt.ResourceGroupName, opt.DiskName, nil)
	if err != nil {
		return fmt.Errorf("failed to finish the request:  %v", err)
	}
	_, err = pollerResp.PollUntilDone(kt.Ctx, nil)

	return err
}

// AttachDisk 挂载云盘
// reference:
// https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/create-or-update?tabs=HTTP#storageprofile
func (a *Azure) AttachDisk(kt *kit.Kit, opt *disk.AzureDiskAttachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "azure disk attach option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	cvmData, err := a.GetCvm(
		kt,
		&typecvm.AzureGetOption{ResourceGroupName: opt.ResourceGroupName, Name: opt.CvmName},
	)
	if err != nil {
		return err
	}

	diskData, err := a.GetDisk(
		kt,
		&disk.AzureDiskGetOption{ResourceGroupName: opt.ResourceGroupName, DiskName: opt.DiskName})
	if err != nil {
		return err
	}

	return a.attachDisk(kt, opt, cvmData, diskData)
}

// DetachDisk 卸载云盘
// reference:
// https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/create-or-update?tabs=HTTP#storageprofile
func (a *Azure) DetachDisk(kt *kit.Kit, opt *disk.AzureDiskDetachOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "azure disk detach option is required")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	cvmData, err := a.GetCvm(
		kt,
		&typecvm.AzureGetOption{ResourceGroupName: opt.ResourceGroupName, Name: opt.CvmName},
	)
	if err != nil {
		return err
	}

	diskData, err := a.GetDisk(
		kt,
		&disk.AzureDiskGetOption{ResourceGroupName: opt.ResourceGroupName, DiskName: opt.DiskName})
	if err != nil {
		return err
	}

	return a.detachDisk(kt, opt, cvmData.StorageProfile, diskData)
}

// attachDisk 通过 vm 的 BeginCreateOrUpdate 接口完成云盘挂载
func (a *Azure) attachDisk(
	kt *kit.Kit,
	opt *disk.AzureDiskAttachOption,
	cvmData *typecvm.AzureCvm,
	diskData *typedisk.AzureDisk,
) error {
	client, err := a.clientSet.virtualMachineClient()
	if err != nil {
		return fmt.Errorf("new cvm client failed, err: %v", err)
	}

	dataDisks := cvmData.StorageProfile.DataDisks
	lun, err := genLun(dataDisks)
	if err != nil {
		return err
	}

	attachType := armcompute.DiskCreateOptionTypesAttach
	cachingType := disk.AzureCachingTypes[opt.CachingType]
	dataDisks = append(
		dataDisks,
		&armcompute.DataDisk{
			Name:         diskData.Name,
			Lun:          &lun,
			ManagedDisk:  &armcompute.ManagedDiskParameters{ID: diskData.ID},
			CreateOption: &attachType,
			Caching:      &cachingType,
		},
	)

	sp := &armcompute.StorageProfile{
		OSDisk:         cvmData.StorageProfile.OSDisk,
		ImageReference: cvmData.StorageProfile.ImageReference,
		DataDisks:      dataDisks,
	}
	vm := armcompute.VirtualMachine{Location: cvmData.Location, Properties: &armcompute.VirtualMachineProperties{
		StorageProfile: sp}}
	_, err = client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, opt.CvmName, vm, nil)
	return err
}

// attachDisk 通过 vm 的 BeginCreateOrUpdate 接口完成云盘卸载
func (a *Azure) detachDisk(
	kt *kit.Kit,
	opt *disk.AzureDiskDetachOption,
	storageProfile *armcompute.StorageProfile,
	diskData *typedisk.AzureDisk,
) error {
	client, err := a.clientSet.virtualMachineClient()
	if err != nil {
		return fmt.Errorf("new cvm client failed, err: %v", err)
	}

	var dataDisks []*armcompute.DataDisk
	for idx, d := range storageProfile.DataDisks {
		if d.Name == diskData.Name && d.ManagedDisk.ID == diskData.ID {
			dataDisks = append(storageProfile.DataDisks[:idx], storageProfile.DataDisks[idx+1:]...)
			break
		}
	}

	sp := &armcompute.StorageProfile{
		OSDisk:         storageProfile.OSDisk,
		ImageReference: storageProfile.ImageReference,
		DataDisks:      dataDisks,
	}
	vm := armcompute.VirtualMachine{Properties: &armcompute.VirtualMachineProperties{StorageProfile: sp}}
	_, err = client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, opt.CvmName, vm, nil)
	return err
}

// genLun 根据已有的 Lun 自动生成一个未被占用的
func genLun(dataDisks []*armcompute.DataDisk) (int32, error) {
	// lunUsed 用来记录已被占用的 Lun
	lunUsed := make(map[int32]bool)

	for _, d := range dataDisks {
		lunUsed[*d.Lun] = true
	}

	i := int32(0)
	for ; i < 64; i++ {
		if !lunUsed[i] {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no available lun")
}
