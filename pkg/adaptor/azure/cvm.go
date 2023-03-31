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
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// ListCvm reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/list?tabs=HTTP
func (az *Azure) ListCvm(kt *kit.Kit, opt *typecvm.AzureListOption) ([]*typecvm.AzureCvm, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.virtualMachineClient()
	if err != nil {
		return nil, fmt.Errorf("new cvm client failed, err: %v", err)
	}

	vms := make([]*armcompute.VirtualMachine, 0)
	pager := client.NewListPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		vms = append(vms, nextResult.Value...)
	}

	return converterToAzureCvm(vms), nil
}

// ListCvmByID reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/list?tabs=HTTP
func (az *Azure) ListCvmByID(kt *kit.Kit, opt *core.AzureListByIDOption) ([]*typecvm.AzureCvm, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.virtualMachineClient()
	if err != nil {
		return nil, fmt.Errorf("new cvm client failed, err: %v", err)
	}

	idMap := converter.StringSliceToMap(opt.CloudIDs)

	vms := make([]*armcompute.VirtualMachine, 0, len(idMap))
	pager := client.NewListPager(opt.ResourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(kt.Ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}

		for _, one := range nextResult.Value {
			if len(opt.CloudIDs) > 0 {
				id := SPtrToLowerSPtr(one.ID)
				if _, exist := idMap[*id]; exist {
					vms = append(vms, one)
					delete(idMap, *id)

					if len(idMap) == 0 {
						return converterToAzureCvm(vms), nil
					}
				}
			} else {
				vms = append(vms, one)
			}
		}
	}

	return converterToAzureCvm(vms), nil
}

func converterToAzureCvm(vms []*armcompute.VirtualMachine) []*typecvm.AzureCvm {
	results := make([]*typecvm.AzureCvm, 0)
	if len(vms) <= 0 {
		return results
	}

	for _, v := range vms {
		tmp := &typecvm.AzureCvm{
			ID:       SPtrToLowerSPtr(v.ID),
			Name:     SPtrToLowerSPtr(v.Name),
			Location: SPtrToLowerNoSpaceSPtr(v.Location),
			Type:     v.Type,
			Zones:    v.Zones,
		}

		if v.Properties == nil {
			results = append(results, tmp)
			continue
		}

		if v.Properties.StorageProfile != nil {
			tmp.StorageProfile = v.Properties.StorageProfile

			if v.Properties.StorageProfile.ImageReference != nil {
				tmp.CloudImageID = SPtrToLowerSPtr(v.Properties.StorageProfile.ImageReference.ID)
			}

			if v.Properties.StorageProfile.OSDisk != nil {
				if v.Properties.StorageProfile.OSDisk.ManagedDisk != nil {
					tmp.CloudOsDiskID = SPtrToLowerStr(v.Properties.StorageProfile.OSDisk.ManagedDisk.ID)
				}
			}

			tmp.CloudDataDiskIDs = make([]string, 0)
			if len(v.Properties.StorageProfile.DataDisks) > 0 {
				for _, disk := range v.Properties.StorageProfile.DataDisks {
					if disk != nil {
						tmp.CloudDataDiskIDs = append(tmp.CloudDataDiskIDs, SPtrToLowerStr(disk.ManagedDisk.ID))
					}
				}
			}
		}

		if v.Properties.OSProfile != nil {
			tmp.ComputerName = v.Properties.OSProfile.ComputerName
		}

		tmp.ProvisioningState = v.Properties.ProvisioningState

		tmp.EvictionPolicy = v.Properties.EvictionPolicy

		if v.Properties.HardwareProfile != nil {
			tmp.VMSize = v.Properties.HardwareProfile.VMSize
		}

		tmp.LicenseType = v.Properties.LicenseType

		tmp.NetworkInterfaceIDs = make([]string, 0)
		if v.Properties.NetworkProfile != nil {
			if len(v.Properties.NetworkProfile.NetworkInterfaces) > 0 {
				for _, networkInterface := range v.Properties.NetworkProfile.NetworkInterfaces {
					if networkInterface != nil {
						tmp.NetworkInterfaceIDs = append(tmp.NetworkInterfaceIDs, SPtrToLowerStr(networkInterface.ID))
					}
				}
			}
		}

		tmp.Priority = v.Properties.Priority

		if v.Properties.AdditionalCapabilities != nil {
			tmp.HibernationEnabled = v.Properties.AdditionalCapabilities.HibernationEnabled
			tmp.UltraSSDEnabled = v.Properties.AdditionalCapabilities.UltraSSDEnabled
		}

		if v.Properties.BillingProfile != nil {
			tmp.MaxPrice = v.Properties.BillingProfile.MaxPrice
		}

		if v.Properties.HardwareProfile.VMSizeProperties != nil {
			tmp.VCPUsAvailable = v.Properties.HardwareProfile.VMSizeProperties.VCPUsAvailable
			tmp.VCPUsPerCore = v.Properties.HardwareProfile.VMSizeProperties.VCPUsPerCore
		}

		tmp.TimeCreated = v.Properties.TimeCreated

		results = append(results, tmp)
	}

	return results
}

// DeleteCvm reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/delete?tabs=Go
func (az *Azure) DeleteCvm(kt *kit.Kit, opt *typecvm.AzureDeleteOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.virtualMachineClient()
	if err != nil {
		return fmt.Errorf("new cvm client failed, err: %v", err)
	}

	poller, err := client.BeginDelete(kt.Ctx, opt.ResourceGroupName, opt.Name,
		&armcompute.VirtualMachinesClientBeginDeleteOptions{ForceDeletion: to.Ptr(opt.Force)})
	if err != nil {
		logs.Errorf("begin delete cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("poll until cvm delete failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// StartCvm reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/start?tabs=HTTP
func (az *Azure) StartCvm(kt *kit.Kit, opt *typecvm.AzureStartOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "start option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.virtualMachineClient()
	if err != nil {
		return fmt.Errorf("new cvm client failed, err: %v", err)
	}

	poller, err := client.BeginStart(kt.Ctx, opt.ResourceGroupName, opt.Name, nil)
	if err != nil {
		logs.Errorf("begin start cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("poll until cvm start failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// RebootCvm reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/restart?tabs=HTTP
func (az *Azure) RebootCvm(kt *kit.Kit, opt *typecvm.AzureRebootOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "reboot option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.virtualMachineClient()
	if err != nil {
		return fmt.Errorf("new cvm client failed, err: %v", err)
	}

	poller, err := client.BeginRestart(kt.Ctx, opt.ResourceGroupName, opt.Name, nil)
	if err != nil {
		logs.Errorf("begin reboot cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("poll until cvm reboot failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// StopCvm reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/restart?tabs=HTTP
func (az *Azure) StopCvm(kt *kit.Kit, opt *typecvm.AzureStopOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "stop option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.virtualMachineClient()
	if err != nil {
		return fmt.Errorf("new cvm client failed, err: %v", err)
	}

	poller, err := client.BeginPowerOff(kt.Ctx, opt.ResourceGroupName, opt.Name,
		&armcompute.VirtualMachinesClientBeginPowerOffOptions{SkipShutdown: to.Ptr(opt.SkipShutdown)})
	if err != nil {
		logs.Errorf("begin stop cvm failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	_, err = poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("poll until cvm stop failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// CreateCvm reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/create-or-update?tabs=HTTP
func (az *Azure) CreateCvm(kt *kit.Kit, opt *typecvm.AzureCreateOption) (string, error) {
	if opt == nil {
		return "", errf.New(errf.InvalidParameter, "create option is required")
	}

	if err := opt.Validate(); err != nil {
		return "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.virtualMachineClient()
	if err != nil {
		return "", fmt.Errorf("new cvm client failed, err: %v", err)
	}

	if len(opt.DataDisk) != 0 {
		dataDisk := make([]*armcompute.DataDisk, len(opt.DataDisk))
		for index, disk := range opt.DataDisk {
			dataDisk[index] = &armcompute.DataDisk{
				Name:         to.Ptr(disk.Name),
				CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesEmpty),
				Lun:          to.Ptr(int32(index)),
				DiskSizeGB:   to.Ptr(disk.SizeGB),
				ManagedDisk: &armcompute.ManagedDiskParameters{
					StorageAccountType: converter.ValToPtr(armcompute.StorageAccountTypes(disk.Type)),
				},
			}
		}
	}

	instance := armcompute.VirtualMachine{
		Location: to.Ptr(opt.Region),
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(opt.InstanceType)),
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkAPIVersion: to.Ptr(armcompute.NetworkAPIVersionTwoThousandTwenty1101),
				NetworkInterfaceConfigurations: []*armcompute.VirtualMachineNetworkInterfaceConfiguration{
					{
						Name: to.Ptr(opt.Name + "-ni"),
						Properties: &armcompute.VirtualMachineNetworkInterfaceConfigurationProperties{
							IPConfigurations: []*armcompute.VirtualMachineNetworkInterfaceIPConfiguration{
								{
									Name: to.Ptr(opt.Name + "ni-ip-config"),
									Properties: &armcompute.VirtualMachineNetworkInterfaceIPConfigurationProperties{
										Primary: to.Ptr(true),
										Subnet: &armcompute.SubResource{
											ID: to.Ptr(opt.CloudSubnetID),
										},
									},
								},
							},
							NetworkSecurityGroup: &armcompute.SubResource{
								ID: to.Ptr(opt.CloudSecurityGroupID),
							},
							Primary: to.Ptr(true),
						},
					},
				},
			},
			OSProfile: &armcompute.OSProfile{
				AdminPassword: to.Ptr(opt.Password),
				AdminUsername: to.Ptr(opt.Username),
				ComputerName:  to.Ptr(opt.Name),
			},
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Offer:     to.Ptr(opt.Image.Offer),
					Publisher: to.Ptr(opt.Image.Publisher),
					SKU:       to.Ptr(opt.Image.Sku),
					Version:   to.Ptr(opt.Image.Version),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr(opt.OSDisk.Name),
					DiskSizeGB:   to.Ptr(opt.OSDisk.SizeGB),
					Caching:      to.Ptr(armcompute.CachingTypesReadWrite),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: converter.ValToPtr(armcompute.StorageAccountTypes(opt.OSDisk.Type)),
					},
				},
			},
		},
	}
	if len(opt.Zones) != 0 {
		instance.Zones = to.SliceOfPtrs(opt.Zones...)
	}
	poller, err := client.BeginCreateOrUpdate(kt.Ctx, opt.ResourceGroupName, opt.Name, instance, nil)
	if err != nil {
		logs.Errorf("begin create cvm failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	resp, err := poller.PollUntilDone(kt.Ctx, nil)
	if err != nil {
		logs.Errorf("poll until cvm create failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return SPtrToLowerStr(resp.ID), nil
}

// GetCvm 查询单个 cvm
// reference: https://learn.microsoft.com/en-us/rest/api/compute/virtual-machines/get?tabs=Go
func (az *Azure) GetCvm(kt *kit.Kit, opt *typecvm.AzureGetOption) (*typecvm.AzureCvm, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "get option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := az.clientSet.virtualMachineClient()
	if err != nil {
		return nil, fmt.Errorf("new cvm client failed, err: %v", err)
	}

	resp, err := client.Get(
		kt.Ctx,
		opt.ResourceGroupName,
		opt.Name,
		&armcompute.VirtualMachinesClientGetOptions{Expand: to.Ptr(armcompute.InstanceViewTypesUserData)},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cvm: %v", err)
	}

	converterResp := converterToAzureCvm([]*armcompute.VirtualMachine{&resp.VirtualMachine})
	if len(converterResp) > 0 {
		return converterResp[0], nil
	}

	return new(typecvm.AzureCvm), nil
}
