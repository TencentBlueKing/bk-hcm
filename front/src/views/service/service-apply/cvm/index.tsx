/* eslint-disable no-useless-escape */
// eslint-disable
import { computed, defineComponent, reactive, ref, watch } from 'vue';
import { Form, Input, Select, Checkbox, Button, Radio } from 'bkui-vue';
import ConditionOptions from '../components/common/condition-options.vue';
import ZoneSelector from '@/components/zone-selector/index.vue';
import MachineTypeSelector from '../components/common/machine-type-selector';
import Imagelector from '../components/common/image-selector';
import VpcSelector from '../components/common/vpc-selector';
import SubnetSelector from '../components/common/subnet-selector';
import SecurityGroupSelector from '../components/common/security-group-selector';
import CloudAreaName from '../components/common/cloud-area-name';
import {
  Plus as PlusIcon,
} from 'bkui-vue/lib/icon';
import GcpDataDiskFormDialog from './children/gcp-data-disk-form-dialog';
import './index.scss';
import { useI18n } from 'vue-i18n';

import type { IOption } from '@/typings/common';
import type { IDiskOption } from '../hooks/use-cvm-form-data';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import useCvmOptions from '../hooks/use-cvm-options';
import useCondtion from '../hooks/use-condtion';
import useCvmFormData, {
  getDataDiskDefaults,
  getGcpDataDiskDefaults,
} from '../hooks/use-cvm-form-data';
// import { useHostStore } from '@/store/host';

import { useAccountStore } from '@/store';
import CommonCard from '@/components/CommonCard';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import { useRouter } from 'vue-router';
import VpcPreviewDialog from './children/VpcPreviewDialog';
import SubnetPreviewDialog, {
  ISubnetItem,
} from './children/SubnetPreviewDialog';
import http from '@/http';
import { debounce } from 'lodash';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
// import SelectCvmBlock from './children/SelectCvmBlock';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const accountStore = useAccountStore();

const { FormItem } = Form;
const { Option } = Select;
const { Group: RadioGroup, Button: RadioButton } = Radio;

export default defineComponent({
  props: {},
  setup() {
    const { cond, isEmptyCond } = useCondtion(ResourceTypeEnum.CVM);
    const {
      formData,
      formRef,
      handleFormSubmit,
      submitting,
      resetFormItemData,
      getSaveData,
      opSystemType,
      changeOpSystemType,
    } = useCvmFormData(cond);
    const { sysDiskTypes, dataDiskTypes, billingModes, purchaseDurationUnits } =      useCvmOptions(cond, formData);
    const { t } = useI18n();
    const router = useRouter();
    const isSubmitBtnLoading = ref(false);
    const usageNum = ref(0);
    const limitNum = ref(-1);
    const refreshVpcList = ref(() => {});
    const onRefreshVpcList = (callback: () => {}) => {
      refreshVpcList.value = callback;
    };

    const dialogState = reactive({
      gcpDataDisk: {
        isShow: false,
        isEdit: false,
        editDataIndex: null,
        formData: getGcpDataDiskDefaults(),
      },
    });
    // const hostStore = useHostStore();

    const zoneSelectorRef = ref(null);
    const cloudId = ref(null);
    const vpcId = ref('');
    const machineType = ref(null);
    const subnetSelectorRef = ref(null);
    const vpcData = ref(null);
    const subnetData = ref(null);
    const isVpcPreviewDialogShow = ref(false);
    const isSubnetPreviewDialogShow = ref(false);
    const cost = ref('--');
    const { whereAmI } = useWhereAmI();

    const handleSubnetDataChange = (data: ISubnetItem) => {
      subnetData.value = data;
    };

    const handleCreateDataDisk = () => {
      const newRow: IDiskOption = getDataDiskDefaults();
      formData.data_disk.push(newRow);
    };

    const handleCreateGcpDataDisk = () => {
      dialogState.gcpDataDisk.isShow = true;
      dialogState.gcpDataDisk.isEdit = false;
      dialogState.gcpDataDisk.formData = getGcpDataDiskDefaults();
    };
    const handleAddGcpDataDisk = (data: IDiskOption) => {
      formData.data_disk.push(data);
      dialogState.gcpDataDisk.isShow = false;
    };
    const handleSaveGcpDataDisk = (data: IDiskOption) => {
      formData.data_disk[dialogState.gcpDataDisk.editDataIndex] = {
        ...formData.data_disk[dialogState.gcpDataDisk.editDataIndex],
        ...data,
      };
      dialogState.gcpDataDisk.isShow = false;
    };
    /* const handleEditGcpDataDisk = (index: number) => {
      dialogState.gcpDataDisk.isShow = true;
      dialogState.gcpDataDisk.isEdit = true;
      dialogState.gcpDataDisk.editDataIndex = index;
      dialogState.gcpDataDisk.formData = formData.data_disk[index];
    };*/

    const handleRemoveDataDisk = (index: number) => {
      formData.data_disk.splice(index, 1);
    };

    const handleZoneChange = () => {
      resetFormItemData('instance_type');
      resetFormItemData('cloud_vpc_id');
      resetFormItemData('cloud_subnet_id');
      vpcId.value = '';
    };
    const handleVpcChange = (vpc: any) => {
      vpcData.value = vpc;
      cloudId.value = vpc.bk_cloud_id;
      if (vpcId.value !== vpc.id) {
        vpcId.value = vpc.id;
        resetFormItemData('cloud_subnet_id');
      }
    };
    const handleMachineTypeChange = (machine: any) => {
      machineType.value = machine;
      resetFormItemData('cloud_image_id');

      if (cond.vendor === VendorEnum.AZURE) {
        resetFormItemData('system_disk');
        resetFormItemData('data_disk');
      }
    };

    const sysDiskSizeRules = computed(() => {
      const rules = {
        [VendorEnum.TCLOUD]: {
          validator: (value: number) => {
            return value >= 20 && value <= 1024;
          },
          message: '20-1024GB',
          trigger: 'change',
        },
        [VendorEnum.HUAWEI]: {
          validator: (value: number) => {
            return value >= 40 && value <= 1024;
          },
          message: '40-1024GB',
          trigger: 'change',
        },
        [VendorEnum.AWS]: {
          validator: (value: number) => {
            return value >= 1 && value <= 16384;
          },
          message: '1-16384GB',
          trigger: 'change',
        },
      };

      return (
        rules[cond.vendor] || {
          validator: () => true,
          message: '',
        }
      );
    });

    const dataDiskSizeRules = (item: any) => {
      const awsMinMap = {
        gp3: 1,
        gp2: 1,
        io1: 4,
        io2: 4,
        st1: 125,
        sc1: 125,
        standard: 1,
      };
      const awsMaxMap = {
        gp3: 16384,
        gp2: 16384,
        io1: 16384,
        io2: 16384,
        st1: 16384,
        sc1: 16384,
        standard: 1024,
      };
      const rules = {
        [VendorEnum.TCLOUD]: {
          validator: (value: number) => {
            return value >= 20 && value <= 32000 && value % 10 === 0;
          },
          message: '20-32,000GB且为10的倍数',
          trigger: 'change',
        },
        [VendorEnum.HUAWEI]: {
          validator: (value: number) => {
            return value >= 40 && value <= 32768;
          },
          message: '40-32,768GB',
          trigger: 'change',
        },
        [VendorEnum.AWS]: {
          validator: (value: number) => {
            return (
              value >= awsMinMap[item.disk_type]
              && value <= awsMaxMap[item.disk_type]
            );
          },
          message: `${awsMinMap[item.disk_type]}-${
            awsMaxMap[item.disk_type]
          }GB`,
          trigger: 'change',
        },
      };

      return (
        rules[cond.vendor] || {
          validator: () => true,
          message: '',
        }
      );
    };

    const dataDiskCountRules = computed(() => {
      const rules = {
        [VendorEnum.TCLOUD]: {
          min: 1,
          max: 20,
          trigger: 'change',
        },
        [VendorEnum.HUAWEI]: {
          min: 1,
          max: 23,
          trigger: 'change',
        },
        [VendorEnum.AWS]: {
          min: 1,
          max: 23,
          trigger: 'change',
        },
      };

      return (
        rules[cond.vendor] || {
          min: 1,
          max: Infinity,
        }
      );
    });

    const submitDisabled = computed(() => isEmptyCond.value);

    const formConfigDataDiskDiff = computed(() => {
      const diffs = {
        [VendorEnum.GCP]: {
          content: () => (
            <div class='form-content-list data-disk-wrap'>
              {/* {formData.data_disk.map((item: IDiskOption, index: number) => (
                <div class='flex-row'>

                  {item.disk_name}, 空白, {item.disk_size_gb}GB,{' '}
                  {dataDiskTypes.value.find((disk: IOption) => disk.id === item.disk_type)?.name || '--'}
                  <div class='btns'>
                    <Button
                      class='btn'
                      outline
                      size='small'
                      onClick={() => handleEditGcpDataDisk(index)}>
                      <EditIcon />
                    </Button>
                    <Button
                      class='btn'
                      outline
                      size='small'
                      onClick={() => handleRemoveDataDisk(index)}>
                      <CloseLineIcon />
                    </Button>
                    {index === formData.data_disk.length - 1 && (
                      <Button
                        class='btn'
                        outline
                        size='small'
                        onClick={handleCreateGcpDataDisk}>
                        <PlusIcon />
                      </Button>
                    )}
                  </div>
                </div>
              ))}*/}
              {formData.data_disk.map((item: IDiskOption, index: number) => (
                  <div class='flex-row'>
                    <FormItem
                        property={`data_disk[${index}].disk_type`}
                        rules={[]}>
                      <Select
                          v-model={item.disk_type}
                          style={{ width: '200px' }}
                          clearable={false}>
                        {dataDiskTypes.value.map(({ id, name }: IOption) => (
                            <Option key={id} value={id} label={name}></Option>
                        ))}
                      </Select>
                    </FormItem>
                    <FormItem
                        property={`data_disk[${index}].disk_size_gb`}
                        rules={[dataDiskSizeRules(item)]}
                        description={dataDiskSizeRules(item).message}>
                      <Input
                          type='number'
                          style={{ width: '160px' }}
                          v-model={item.disk_size_gb}
                          min={1}
                          suffix='GB'
                          prefix='大小'></Input>
                    </FormItem>
                    <FormItem
                        property={`data_disk[${index}].disk_count`}
                        min={dataDiskCountRules.value.min}
                        max={dataDiskCountRules.value.max}>
                      <Input
                          style={{ width: '90px' }}
                          type='number'
                          v-model={item.disk_count}
                          min={dataDiskCountRules.value.min}></Input>
                    </FormItem>
                    <div class='btns'>
                      <Button class={'btn'} onClick={handleCreateGcpDataDisk}>
                        <svg width={14} height={14} viewBox="0 0 24 24" version="1.1"
                             xmlns="http://www.w3.org/2000/svg"
                             style="fill: #c4c6cc">
                          <path d="M12 0c-6.627 0-12 5.373-12 12s5.373 12 12 12c6.627 0 12-5.373 12-12s-5.373-12-12-12zM17.25 12.75h-4.5v4.5c0 0.414-0.336 0.75-0.75 0.75s-0.75-0.336-0.75-0.75v-4.5h-4.5c-0.414 0-0.75-0.336-0.75-0.75s0.336-0.75 0.75-0.75h4.5v-4.5c0-0.414 0.336-0.75 0.75-0.75s0.75 0.336 0.75 0.75v4.5h4.5c0.414 0 0.75 0.336 0.75 0.75s-0.336 0.75-0.75 0.75z"></path>
                        </svg>
                      </Button>
                      <Button class={'btn'} onClick={() => handleRemoveDataDisk(index)}>
                        <svg width={14} height={14} viewBox="0 0 24 24" version="1.1"
                             xmlns="http://www.w3.org/2000/svg"
                             style="fill: #c4c6cc">
                          <path d="M12 0c-6.627 0-12 5.373-12 12s5.373 12 12 12c6.627 0 12-5.373 12-12s-5.373-12-12-12zM17.25 12.75h-10.5c-0.414 0-0.75-0.336-0.75-0.75s0.336-0.75 0.75-0.75h10.5c0.414 0 0.75 0.336 0.75 0.75s-0.336 0.75-0.75 0.75z"></path>
                        </svg>
                      </Button>
                    </div>
                  </div>
              ))}
              {!formData.data_disk.length && (
                <Button onClick={handleCreateGcpDataDisk}>
                  <PlusIcon />
                </Button>
              )}
            </div>
          ),
        },
      };
      return diffs[cond.vendor] || {};
    });

    // const formConfigPublicIpAssignedDiff = computed(() => {
    //   const diffs = {
    //     [VendorEnum.HUAWEI]: {
    //       label: '弹性公网IP',
    //       content: () => '暂不支持购买，请到EIP中绑定',
    //     },
    //   };
    //   return diffs[cond.vendor] || {};
    // });

    // 当前 vpc下是否有子网列表
    const subnetLength = ref(0);
    watch(
      () => formData.cloud_vpc_id,
      (val) => {
        !val && (cloudId.value = null);
        console.log(
          'subnetSelectorRef.value',
          subnetSelectorRef.value.subnetList,
        );
        subnetLength.value = subnetSelectorRef.value.subnetList?.length || 0;
      },
    );

    watch(
      () => cond.vendor,
      () => {
        formData.system_disk.disk_type = '';
      },
    );

    watch(
      () => formData,
      debounce(async () => {
        const saveData = getSaveData();
        if (
          ![VendorEnum.TCLOUD, VendorEnum.HUAWEI].includes(cond.vendor as VendorEnum)
        ) return;
        if (
          !saveData.account_id
          || !saveData.region
          || !saveData.zone
          || !saveData.name
          || !saveData.instance_type
          || !saveData.cloud_image_id
          || !saveData.cloud_vpc_id
          || !saveData.cloud_subnet_id
          || !saveData.cloud_security_group_ids
          || !saveData.system_disk?.disk_type
          || !saveData.password
          || !saveData.confirmed_password
        ) return;
        await formRef.value.validate();
        isSubmitBtnLoading.value = true;
        const res = await http.post(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/cvms/prices/inquiry`,
          {
            ...saveData,
            instance_type: cond.vendor !== VendorEnum.HUAWEI ? saveData.instance_type : `${saveData.instance_type}.${opSystemType.value}`,
          },
        );
        cost.value = res.data?.discount_price || '0';
        isSubmitBtnLoading.value = false;
      }, 300),
      {
        immediate: true,
        deep: true,
      },
    );

    watch(
      () => [cond, formData.zone, formData.instance_charge_type],
      async ([,newZone], [,oldZone]) => {
        const isBusiness = whereAmI.value === Senarios.business;
        const isTcloud = cond.vendor === VendorEnum.TCLOUD;
        if (isBusiness && !cond.bizId) return;
        if (!cond.cloudAccountId || !cond.vendor || !cond.region) return;
        if (isTcloud && !formData.zone.length) return;
        // 避免多发一次无效请求（因为监听了formData.zone的变化）
        if (newZone === oldZone) return;
        if (
          ![VendorEnum.HUAWEI, VendorEnum.GCP, VendorEnum.TCLOUD].includes(cond.vendor as VendorEnum)
        ) return;
        let url = isBusiness
          ? `/api/v1/cloud/bizs/${cond.bizId}/vendors/${cond.vendor}/accounts/${cond.cloudAccountId}/regions/quotas`
          : `/api/v1/cloud/vendors/${cond.vendor}/accounts/${cond.cloudAccountId}/regions/quotas`;
        if (cond.vendor === VendorEnum.TCLOUD) {
          url = isBusiness
            ? `/api/v1/cloud/bizs/${cond.bizId}/vendors/${cond.vendor}/accounts/${cond.cloudAccountId}/zones/quotas`
            : `/api/v1/cloud/vendors/${cond.vendor}/accounts/${cond.cloudAccountId}/zones/quotas`;
        }
        const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}${url}`, {
          bk_biz_id: isBusiness ? cond.bizId : undefined,
          account_id: cond.cloudAccountId,
          vendor: cond.vendor,
          region: cond.region,
          zone: isTcloud ? formData.zone : undefined,
        });
        switch (cond.vendor) {
          case VendorEnum.GCP:
            limitNum.value = res.data.instance.limit;
            usageNum.value = res.data.instance.usage;
            break;
          case VendorEnum.TCLOUD: {
            let dataSource = res.data.spot_paid_quota;
            if (['PREPAID'].includes(formData.instance_charge_type)) dataSource = res.data.pre_paid_quota;
            if (
              ['POSTPAID_BY_HOUR', 'postPaid'].includes(formData.instance_charge_type)
            ) dataSource = res.data.post_paid_quota_set;
            limitNum.value = dataSource.total_quota;
            usageNum.value = dataSource.used_quota;
            break;
          }
          case VendorEnum.HUAWEI:
            limitNum.value = res.data.max_total_instances;
            usageNum.value = res.data.max_total_floating_ips;
            break;
        }
      },
      {
        deep: true,
      },
    );

    // const curRegionName = computed(() => {
    //   return hostStore.regionList?.find(region => region.region_id === cond.region) || {};
    // });

    const formConfig = computed(() => [
      // {
      //   id: 'region',
      //   title: '地域',
      //   children: [
      //     {
      //       label: '可用区',
      //       required: cond.vendor === VendorEnum.AZURE ? zoneSelectorRef.value.list?.length > 0 : true,
      //       property: 'zone',
      //       rules: [{
      //         trigger: 'change',
      //       }],
      //       content: () => <ZoneSelector
      //         ref={zoneSelectorRef}
      //         v-model={formData.zone}
      //         vendor={cond.vendor}
      //         region={cond.region}
      //         onChange={handleZoneChange} />,
      //     },
      //   ],
      // },
      {
        id: 'network',
        title: '网络信息',
        children: [
          {
            label: 'VPC',
            required: true,
            property: 'cloud_vpc_id',
            content: () => (
              <div class={'component-with-detail-container'}>
                <VpcSelector
                  class={'component-with-detail'}
                  v-model={formData.cloud_vpc_id}
                  bizId={cond.bizId ? cond.bizId : accountStore.bizs}
                  accountId={cond.cloudAccountId}
                  vendor={cond.vendor}
                  region={cond.region}
                  zone={formData.zone}
                  onChange={handleVpcChange}
                  clearable={false}
                  onRefreshVpcList={onRefreshVpcList}
                />
                <Button
                  text
                  theme='primary'
                  disabled={!formData.cloud_vpc_id}
                  style={{ marginRight: '-50px' }}
                  onClick={() => {
                    isVpcPreviewDialogShow.value = true;
                  }}>
                  预览
                </Button>
              </div>
            ),
          },
          {
            label: '子网',
            required: true,
            property: 'cloud_subnet_id',
            content: () => (
              <>
                <Checkbox class='automatic-allocation-checkbox' v-model={formData.public_ip_assigned} disabled>
                  自动分配公网IP
                </Checkbox>
                <div class={'component-with-detail-container'}>
                  <SubnetSelector
                    class={'component-with-detail'}
                    v-model={formData.cloud_subnet_id}
                    bizId={cond.bizId ? cond.bizId : accountStore.bizs}
                    vpcId={vpcId.value}
                    vendor={cond.vendor}
                    region={cond.region}
                    accountId={cond.cloudAccountId}
                    zone={formData.zone}
                    resourceGroup={cond.resourceGroup}
                    ref={subnetSelectorRef}
                    clearable={false}
                    handleChange={handleSubnetDataChange}
                  />
                  <Button
                    text
                    theme='primary'
                    disabled={!formData.cloud_subnet_id}
                    style={{ marginRight: '-50px' }}
                    class={'subnet-selector-preview-btn'}
                    onClick={() => {
                      isSubnetPreviewDialogShow.value = true;
                      // if (!formData.cloud_subnet_id) return;
                      // const url = `/#/business/subnet?cloud_id=${formData.cloud_subnet_id}&bizs=${cond.bizId}`;
                      // window.open(url, '_blank');
                    }}>
                    预览
                  </Button>
                </div>
              </>
            ),
          },
          // {
          //   label: '公网IP',
          //   display: ![VendorEnum.GCP, VendorEnum.AZURE].includes(cond.vendor),
          //   required: true,
          //   description: '',
          //   property: 'public_ip_assigned',
          //   content: () => <Checkbox v-model={formData.public_ip_assigned} disabled>自动分配公网IP</Checkbox>,
          //   ...formConfigPublicIpAssignedDiff.value,
          // },
          {
            label: '管控区域',
            description: '管控区是蓝鲸可以管控的Agent网络区域，以实现跨网管理。\n一个VPC，对应一个管控区。如VPC未绑定管控区，请到资源接入-VPC-绑定管控区操作。',
            display: whereAmI.value === Senarios.business,
            content: () => (
              <>
                <CloudAreaName id={cloudId.value} />
                <span class={'instance-name-tips'}>
                  如VPC未绑定管控区，请到资源接入-VPC-绑定管控区操作
                  <Button
                    theme='primary'
                    text
                    disabled={!formData.cloud_vpc_id}
                    class={'ml6'}
                    onClick={() => {
                      refreshVpcList.value();
                    }}>
                    刷新
                  </Button>
                </span>
              </>
            ),
          },
          {
            label: '安全组',
            display: cond.vendor !== VendorEnum.GCP,
            required: true,
            property: 'cloud_security_group_ids',
            content: () => (
              <div class={'component-with-detail-container'}>
                <SecurityGroupSelector
                  class={'component-with-detail'}
                  v-model={formData.cloud_security_group_ids}
                  bizId={cond.bizId ? cond.bizId : accountStore.bizs}
                  accountId={cond.cloudAccountId}
                  region={cond.region}
                  multiple={cond.vendor !== VendorEnum.AZURE}
                  vendor={cond.vendor}
                  vpcId={vpcId.value}
                  onSelectedChange={val => (formData.cloud_security_group_ids = val)}
                />
                {/* {
                  isResourcePage
                    ? null
                    : <Button
                        text
                        theme="primary"
                        onClick={() => {
                          if (!formData.cloud_security_group_ids) return;
                          let url = `/#/business/security?bizs=${cond.bizId}&`;
                          const params = [];
                          for (const cloudId of formData.cloud_security_group_ids) {
                            params.push(`cloud_id=${cloudId}`);
                          }
                          url += params.join('&');
                          window.open(url, '_blank');
                        }}>
                        详情
                      </Button>
                 } */}
              </div>
            ),
          },
        ],
      },
      // {
      //   id: 'billing',
      //   title: '计费',
      //   display: [VendorEnum.TCLOUD, VendorEnum.HUAWEI].includes(cond.vendor),
      //   children: [
      //     {
      //       label: '计费模式',
      //       required: true,
      //       property: 'instance_charge_type',
      //       content: () =>
      //       // <Select v-model={formData.instance_charge_type} clearable={false}>{
      //       //     billingModes.value.map(({ id, name }: IOption) => (
      //       //       <Option key={id} value={id} label={name}></Option>
      //       //     ))
      //       //   }
      //       // </Select>,
      //       <RadioGroup v-model={formData.instance_charge_type}>
      //               {billingModes.value.map(item => (<RadioButton label={item.id} >{item.name}
      //             </RadioButton>))}
      //       </RadioGroup>,
      //     },
      //   ],
      // },
      {
        id: 'config',
        title: '实例配置',
        children: [
          {
            label: '机型',
            required: true,
            property: 'instance_type',
            content: () => (
              <MachineTypeSelector
                v-model={formData.instance_type}
                vendor={cond.vendor}
                accountId={cond.cloudAccountId}
                zone={formData.zone}
                region={cond.region}
                bizId={cond.bizId ? cond.bizId : accountStore.bizs}
                instanceChargeType={formData.instance_charge_type}
                clearable={false}
                onChange={handleMachineTypeChange}
              />
            ),
          },
          {
            label: '镜像',
            required: true,
            property: 'cloud_image_id',
            content: () => (
              <Imagelector
                v-model={formData.cloud_image_id}
                vendor={cond.vendor}
                region={cond.region}
                machineType={machineType.value}
                changeOpSystemType={changeOpSystemType}
              />
            ),
          },
          {
            label: '系统盘类型',
            required: true,
            content: [
              {
                property: 'system_disk.disk_type',
                required: true,
                content: () => (
                  <Select
                    v-model={formData.system_disk.disk_type}
                    style={{ width: '200px' }}
                    clearable={false}>
                    {sysDiskTypes.value.map(({ id, name }: IOption) => (
                      <Option key={id} value={id} label={name}></Option>
                    ))}
                  </Select>
                ),
              },
              {
                required: true,
                property: 'system_disk.disk_size_gb',
                rules: [sysDiskSizeRules.value],
                description: sysDiskSizeRules.value.message,
                content: () => (
                  <Input
                    type='number'
                    v-model={formData.system_disk.disk_size_gb}
                    min={1}
                    suffix='GB'
                    prefix='大小'></Input>
                ),
              },
            ],
          },
          {
            label: '数据盘',
            tips: () => (cond.vendor === VendorEnum.TCLOUD
              ? '增强型SSD云硬盘仅在部分可用区开放售卖，后续将逐步增加售卖可用区'
              : ''),
            property: 'data_disk',
            content: () => (
              <div class='form-content-list data-disk-wrap'>
                {formData.data_disk.map((item: IDiskOption, index: number) => (
                  <div class='flex-row'>
                    <FormItem
                      property={`data_disk[${index}].disk_type`}
                      rules={[]}>
                      <Select
                        v-model={item.disk_type}
                        style={{ width: '200px' }}
                        clearable={false}>
                        {dataDiskTypes.value.map(({ id, name }: IOption) => (
                          <Option key={id} value={id} label={name}></Option>
                        ))}
                      </Select>
                    </FormItem>
                    <FormItem
                      property={`data_disk[${index}].disk_size_gb`}
                      rules={[dataDiskSizeRules(item)]}
                      description={dataDiskSizeRules(item).message}>
                      <Input
                        type='number'
                        style={{ width: '160px' }}
                        v-model={item.disk_size_gb}
                        min={1}
                        suffix='GB'
                        prefix='大小'></Input>
                    </FormItem>
                    <FormItem
                      property={`data_disk[${index}].disk_count`}
                      min={dataDiskCountRules.value.min}
                      max={dataDiskCountRules.value.max}>
                      <Input
                        style={{ width: '90px' }}
                        type='number'
                        v-model={item.disk_count}
                        min={dataDiskCountRules.value.min}></Input>
                    </FormItem>
                    <div class='btns'>
                      <Button class={'btn'} onClick={handleCreateDataDisk}>
                        <svg width={14} height={14} viewBox="0 0 24 24" version="1.1"
                             xmlns="http://www.w3.org/2000/svg"
                             style="fill: #c4c6cc">
                          <path d="M12 0c-6.627 0-12 5.373-12 12s5.373 12 12 12c6.627 0 12-5.373 12-12s-5.373-12-12-12zM17.25 12.75h-4.5v4.5c0 0.414-0.336 0.75-0.75 0.75s-0.75-0.336-0.75-0.75v-4.5h-4.5c-0.414 0-0.75-0.336-0.75-0.75s0.336-0.75 0.75-0.75h4.5v-4.5c0-0.414 0.336-0.75 0.75-0.75s0.75 0.336 0.75 0.75v4.5h4.5c0.414 0 0.75 0.336 0.75 0.75s-0.336 0.75-0.75 0.75z"></path>
                        </svg>
                      </Button>
                      <Button class={'btn'} onClick={() => handleRemoveDataDisk(index)}>
                        <svg width={14} height={14} viewBox="0 0 24 24" version="1.1"
                             xmlns="http://www.w3.org/2000/svg"
                             style="fill: #c4c6cc">
                          <path d="M12 0c-6.627 0-12 5.373-12 12s5.373 12 12 12c6.627 0 12-5.373 12-12s-5.373-12-12-12zM17.25 12.75h-10.5c-0.414 0-0.75-0.336-0.75-0.75s0.336-0.75 0.75-0.75h10.5c0.414 0 0.75 0.336 0.75 0.75s-0.336 0.75-0.75 0.75z"></path>
                        </svg>
                      </Button>
                      {/* <Button
                        class='btn'
                        outline
                        size='small'
                        disabled={formData.data_disk.length !== index + 1}
                        onClick={handleCreateDataDisk}>
                        <PlusIcon />
                      </Button>
                      <Button
                        class='btn'
                        outline
                        size='small'
                        disabled={formData.data_disk.length !== index + 1}
                        onClick={() => handleRemoveDataDisk(index)}>
                        <CloseLineIcon />
                      </Button>*/}
                    </div>
                  </div>
                ))}
                {!formData.data_disk.length && (
                    <Button onClick={handleCreateDataDisk}>
                      <PlusIcon />
                    </Button>)
                }
                {
                  // (formData.data_disks.length > 0 && cond.vendor === VendorEnum.HUAWEI)
                  // && <Checkbox v-model={formData.is_quickly_initialize_data_disk}>快速初始化数据盘</Checkbox>
                }
              </div>
            ),
            ...formConfigDataDiskDiff.value,
          },
          {
            label: '密码',
            required: true,
            description: '密码必须包含3种组合：1.大写字母，2.小写字母，3. 数字或特殊字符（!@$%^-_=+[{}]:,./?）',
            content: [
              {
                property: 'username',
                display: cond.vendor === VendorEnum.AZURE,
                content: () => (
                  <Input
                    placeholder='登录用户'
                    v-model={formData.username}></Input>
                ),
              },
              {
                property: 'password',
                content: () => (
                  <Input
                    style={{ width: '249px' }}
                    type='password'
                    placeholder='密码'
                    v-model={formData.password}></Input>
                ),
              },
              {
                property: 'confirmed_password',
                content: () => (
                  <Input
                    style={{ width: '249px' }}
                    type='password'
                    placeholder='确认密码'
                    v-model={formData.confirmed_password}></Input>
                ),
              },
            ],
          },
          {
            label: '实例名称',
            required: true,
            property: 'name',
            maxlength: 60,
            description:
              '60个字符，字母、数字、“-”，且必须以字母、数字开头和结尾。\n\r 实例名称是在云上的记录名称，并不是操作系统上的主机名，以方便使用名称来搜索主机。\n\r 如申请的是1台主机，则按填写的名称命名。如申请的是多台，则填写名称是前缀，申请单会自动补充随机的后缀。',
            content: () => (
              <div>
                <Input
                  placeholder='填写实例名称，主机数量大于1时支持批量命名'
                  v-model={formData.name}
                />
                <div class={'instance-name-tips'}>
                  {'当申请数量 > 1时，该名称为前缀，申请单会自动补充随机后缀'}
                </div>
              </div>
            ),
          },
        ],
      },
      // {
      //   id: 'storage',
      //   title: '存储',
      //   children: [
      //   ],
      // },
      // {
      //   id: 'auth',
      //   title: '登录',
      //   children: [
      //     {
      //       label: '设置密码',
      //       required: true,
      //       description: '字母数字与 ()\`~!@#$%^&*-+=|{}[]:;\',.?/ 字符的组合',
      //       content: [
      //         {
      //           property: 'username',
      //           display: cond.vendor === VendorEnum.AZURE,
      //           content: () => <Input placeholder='登录用户' v-model={formData.username}></Input>,
      //         },
      //         {
      //           property: 'password',
      //           content: () => <Input type='password' placeholder='密码' v-model={formData.password}></Input>,
      //         },
      //         {
      //           property: 'confirmed_password',
      // eslint-disable-next-line max-len
      //           content: () => <Input type='password' placeholder='确认密码' v-model={formData.confirmed_password}></Input>,
      //         },
      //       ],
      //     },
      //   ],
      // },
      // {
      //   id: 'quantity',
      //   title: '数量',
      //   children: [
      //     {
      //       label: '购买数量',
      //       required: true,
      //       property: 'required_count',
      //       description: '大于0的整数，最大不能超过100',
      //       content: () => <Input type='number' min={0} max={100} v-model={formData.required_count}></Input>,
      //     },
      //     {
      //       label: '购买时长',
      //       required: true,
      //       // PREPAID：包年包月
      //       display: ['PREPAID'].includes(formData.instance_charge_type),
      //       content: [
      //         {
      //           property: 'purchase_duration.count',
      //           content: () => <Input type='number' v-model={formData.purchase_duration.count}></Input>,
      //         },
      //         {
      //           property: 'purchase_duration.unit',
      //           content: () => <Select v-model={formData.purchase_duration.unit} clearable={false}>{
      //             purchaseDurationUnits.map(({ id, name }: IOption) => (
      //               <Option key={id} value={id} label={name}></Option>
      //             ))}
      //           </Select>,
      //         },
      //         {
      //           property: 'auto_renew',
      //           content: () => <Checkbox v-model={formData.auto_renew}>自动续费</Checkbox>,
      //         },
      //       ],
      //     },
      //   ],
      // },
      {
        id: 'describe',
        title: '备注信息',
        children: [
          {
            label: '实例备注',
            property: 'memo',
            content: () => (
              <Input
                type='textarea'
                placeholder='填写实例备注'
                rows={3}
                maxlength={255}
                resize={false}
                v-model={formData.memo}></Input>
            ),
          },
          {
            label: '申请单备注',
            property: 'remark',
            content: () => (
              <Input
                type='textarea'
                placeholder='填写申请单备注'
                rows={3}
                maxlength={255}
                resize={false}
                v-model={formData.remark}></Input>
            ),
          },
        ],
      },
    ]);

    const formRules = {
      name: [
        {
          pattern: /^[a-zA-Z0-9][a-zA-Z0-9-]{0,58}[a-zA-Z0-9]$/,
          message: '60个字符，字母、数字、“-”，且必须以字母、数字开头和结尾',
          trigger: 'change',
        },
      ],
      password: [
        {
          validator: (value: string) => value.length >= 8 && value.length <= 30,
          message: '密码长度需要在8-30个字符之间',
          trigger: 'blur',
        },
        {
          validator: (value: string) => {
            /* const pattern = cond.vendor === VendorEnum.HUAWEI
               ? /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[()`~!@#$%^&*-+=|{}\
               [\]:;',.?/])[A-Za-z\d()`~!@#$%^&*\-+=|{}\[\]:;',.?/]+$/
               : /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d|.*[!@$%^\-_=+[{}\]:,./?])[A-Za-z\d!@$%^\-_=+[{}\]:,./?]+$/;
            */
            const pattern = /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d|.*[!@$%^\-_=+[{}\]:,./?])[A-Za-z\d!@$%^\-_=+[{}\]:,./?]+$/;
            return pattern.test(value);
          },
          message: '密码不符合复杂度要求',
          trigger: 'blur',
        },
        {
          validator: (value: string) => {
            // formRef.value.clearValidate('confirmed_password');
            if (formData.confirmed_password.length) {
              return value === formData.confirmed_password;
            }
            return true;
          },
          message: '两次输入的密码不一致',
          trigger: 'blur',
        },
      ],
      confirmed_password: [
        {
          validator: (value: string) => value.length >= 8 && value.length <= 30,
          message: '密码长度需要在8-30个字符之间',
          trigger: 'blur',
        },
        {
          validator: (value: string) => {
            const pattern = /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d|.*[!@$%^\-_=+[{}\]:,./?])[A-Za-z\d!@$%^\-_=+[{}\]:,./?]+$/;
            return pattern.test(value);
          },
          message: '密码不符合复杂度要求',
          trigger: 'blur',
        },
        {
          validator: (value: string) => {
            // formRef.value.clearValidate('password');
            return formData.password.length && value === formData.password;
          },
          message: '两次输入的密码不一致',
          trigger: 'blur',
        },
      ],
      username: [
        {
          validator: (value: string) => {
            const sensitives = [
              '123',
              'administrator',
              'console',
              'guest',
              'test3',
              'user1',
              'user5',
              'admin1',
              'test1',
              'john',
              'owner',
              'test',
              'user4',
              'david',
              'root',
              'support_388945a0',
              'user',
              'user2',
              '1',
              'support',
              'video',
              'a',
              'admin',
              'sys',
              'test2',
              'admin2',
              'aspnet',
              'sql',
              'user3',
              'actuser',
              'adm',
              'backup',
              'server',
            ];
            return !sensitives.includes(value);
          },
          message: '不允许使用的用户名',
          trigger: 'change',
        },
      ],
      required_count: [
        {
          max: 100,
          message: '最大不能超过100',
          trigger: 'change',
        },
      ],
      data_disk: [
        {
          validator: (disks: []) => {
            const diskNum = disks.reduce((acc: number, cur: any) => {
              acc += cur.disk_count;
              return acc;
            }, 0);
            return cond.vendor !== VendorEnum.AWS || diskNum <= 23;
          },
          message: '数据盘总数不能超过23个',
          trigger: 'change',
        },
      ],
    };

    return () => (
      <div>
        <DetailHeader>
          <p class={'purchase-cvm-header-title'}>购买主机</p>
        </DetailHeader>
        <div
          class='create-form-container cvm-wrap'
          style={whereAmI.value === Senarios.resource && { padding: 0, marginBottom: '80px' }}>
          <Form
            model={formData}
            rules={formRules}
            ref={formRef}
            onSubmit={handleFormSubmit}
            formType='vertical'>
            <ConditionOptions
              type={ResourceTypeEnum.CVM}
              v-model:bizId={cond.bizId}
              v-model:cloudAccountId={cond.cloudAccountId}
              v-model:vendor={cond.vendor}
              v-model:region={cond.region}
              v-model:resourceGroup={cond.resourceGroup}>
              {{
                default: () => (
                  <FormItem label={'可用区'} required property='zone'>
                    <ZoneSelector
                      ref={zoneSelectorRef}
                      v-model={formData.zone}
                      vendor={cond.vendor}
                      region={cond.region}
                      onChange={handleZoneChange}
                    />
                  </FormItem>
                ),
                appendix: () => ([VendorEnum.TCLOUD, VendorEnum.HUAWEI].includes(cond.vendor as VendorEnum) ? (
                  <FormItem label='计费模式' required property='instance_charge_type'>
                    <RadioGroup v-model={formData.instance_charge_type}>
                      {billingModes.value.map(item => (
                        <RadioButton label={item.id}>{item.name}</RadioButton>
                      ))}
                    </RadioGroup>
                  </FormItem>
                ) : null),
              }}
            </ConditionOptions>
            {formConfig.value
              .filter(({ display }) => display !== false)
              .map(({ title, children }) => (
                <CommonCard title={() => title} class={'mb16'}>
                  {children
                    .filter(({ display }) => display !== false)
                    .map(({
                      label,
                      description,
                      tips,
                      rules,
                      required,
                      property,
                      content,
                    }) => (
                        <FormItem
                          label={label}
                          required={required}
                          property={property}
                          rules={rules}
                          description={description}
                          class={label === '子网' && 'purchase-cvm-form-item-subnet-wrap'}
                          >
                          {Array.isArray(content) ? (
                            <div class='flex-row'>
                              {content
                                .filter(sub => sub.display !== false)
                                .map(sub => (
                                  <FormItem
                                    label={sub.label}
                                    required={sub.required}
                                    property={sub.property}
                                    rules={sub.rules}
                                    description={sub?.description}
                                    class='sub-form-item-wrap'
                                  >
                                    {sub.content()}
                                    {sub.tips && (
                                      <div class='form-item-tips'>
                                        {sub.tips()}
                                      </div>
                                    )}
                                  </FormItem>
                                ))}
                            </div>
                          ) : (
                            content()
                          )}
                          {tips && <div class='form-item-tips'>{tips()}</div>}
                        </FormItem>
                    ))}
                </CommonCard>
              ))}
            {/* <div class="action-bar">
          <Button theme='primary' loading={submitting.value}
          disabled={submitDisabled.value} onClick={handleFormSubmit}>{
            isResourcePage ? t('提交') : t('提交审批')
          }</Button>
          <Button>{ t('取消') }</Button>
        </div> */}
          </Form>
          <GcpDataDiskFormDialog
            v-model:isShow={dialogState.gcpDataDisk.isShow}
            isEdit={dialogState.gcpDataDisk.isEdit}
            dataDiskTypes={dataDiskTypes.value}
            formData={dialogState.gcpDataDisk.formData}
            onAdd={handleAddGcpDataDisk}
            onSave={handleSaveGcpDataDisk}
            onClose={() => (dialogState.gcpDataDisk.isShow = false)}
          />
        </div>
        <div class={'purchase-cvm-bottom-bar'}>
          <Form labelWidth={130} class={'purchase-cvm-bottom-bar-form'}>
            <div class='purchase-cvm-bottom-bar-form-item-wrap'>
              <FormItem
                label='数量'
                class={
                  'purchase-cvm-bottom-bar-form-count '
                  + `${limitNum.value !== -1 ? 'mb-12' : ''}`
                }>
                <Input
                  style={{ width: '150px' }}
                  type='number'
                  min={0}
                  max={100}
                  v-model={formData.required_count}></Input>
              </FormItem>

              {/* eslint-disable max-len */}
              {['PREPAID', 'prePaid'].includes(formData.instance_charge_type) ? (
                <FormItem label='时长'>
                  <div class={'purchase-cvm-time'}>
                    <Input
                      style={{ width: '160px' }}
                      type='number'
                      v-model={formData.purchase_duration.count}></Input>
                    <Select
                      style={{ width: '50px' }}
                      v-model={formData.purchase_duration.unit}
                      clearable={false}>
                      {purchaseDurationUnits.map(({ id, name }: IOption) => (
                        <Option key={id} value={id} label={name}></Option>
                      ))}
                    </Select>
                    <Checkbox class='purchase-cvm-time-checkbox' v-model={formData.auto_renew}> 自动续费 </Checkbox>
                  </div>
                </FormItem>
              ) : null}
            </div>
            {/* eslint-disable max-len */}

            <div class='purchase-cvm-bottom-bar-form-count-wrap'>
              {[VendorEnum.TCLOUD, VendorEnum.HUAWEI, VendorEnum.GCP].includes(cond.vendor as VendorEnum) && limitNum.value !== -1 ? (
                <p class={'purchase-cvm-bottom-bar-form-count-tip'}>
                  所在{VendorEnum.TCLOUD === cond.vendor ? '可用区' : '地域'}
                  配额为{' '}
                  {
                    <>
                      <span
                        class={'purchase-cvm-bottom-bar-form-count-tip-num'}>
                        {limitNum.value
                          - usageNum.value
                          - formData.required_count}
                      </span>{' '}
                      / {limitNum.value}
                    </>
                  }
                </p>
              ) : null}
            </div>
          </Form>
          <div class={'purchase-cvm-bottom-bar-info'}>
            {
              (cond.vendor === VendorEnum.TCLOUD || cond.vendor === VendorEnum.HUAWEI)
                && (
                  <div class={'purchase-cvm-cost-wrap'}>
                    <div>费用：</div>
                    <div class={'purchase-cvm-cost'}>{cost.value}</div>
                  </div>
                )
            }
            <Button
              theme='primary'
              loading={submitting.value || isSubmitBtnLoading.value}
              disabled={submitDisabled.value}
              onClick={handleFormSubmit}
              class={'mr8'}>
              立即购买
            </Button>
            <Button onClick={() => router.back()}>{t('取消')}</Button>
          </div>

          <VpcPreviewDialog
            isShow={isVpcPreviewDialogShow.value}
            data={vpcData.value}
            handleClose={() => (isVpcPreviewDialogShow.value = false)}
          />

          <SubnetPreviewDialog
            isShow={isSubnetPreviewDialogShow.value}
            data={subnetData.value}
            handleClose={() => (isSubnetPreviewDialogShow.value = false)}
          />
        </div>
      </div>
    );
  },
});
