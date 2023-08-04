/* eslint-disable no-useless-escape */
import { computed, defineComponent, reactive, ref, watch } from 'vue';
import { Form, Input, Select, Checkbox, Button, Radio } from 'bkui-vue';
import ContentContainer from '../components/common/content-container.vue';
import ConditionOptions from '../components/common/condition-options.vue';
import FormGroup from '../components/common/form-group.vue';
import ZoneSelector from '../components/common/zone-selector';
import MachineTypeSelector from '../components/common/machine-type-selector';
import Imagelector from '../components/common/image-selector';
import VpcSelector from '../components/common/vpc-selector';
import SubnetSelector from '../components/common/subnet-selector';
import SecurityGroupSelector from '../components/common/security-group-selector';
import CloudAreaName from '../components/common/cloud-area-name';
import { Plus as PlusIcon, CloseLine as CloseLineIcon, EditLine as EditIcon } from 'bkui-vue/lib/icon';
import GcpDataDiskFormDialog from './children/gcp-data-disk-form-dialog';
import './index.scss';
import { useI18n } from 'vue-i18n';

import type { IOption } from '@/typings/common';
import type { IDiskOption } from '../hooks/use-cvm-form-data';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import useCvmOptions from '../hooks/use-cvm-options';
import useCondtion from '../hooks/use-condtion';
import useCvmFormData, { getDataDiskDefaults, getGcpDataDiskDefaults } from '../hooks/use-cvm-form-data';
// import { useHostStore } from '@/store/host';

import { useAccountStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const accountStore = useAccountStore();

const { FormItem } = Form;
const { Option } = Select;
const { Group: RadioGroup, Button: RadioButton } = Radio;

export default defineComponent({
  props: {},
  setup() {
    const { cond, isEmptyCond } = useCondtion(ResourceTypeEnum.CVM);
    const { formData, formRef, handleFormSubmit, submitting, resetFormItemData } = useCvmFormData(cond);
    const {
      sysDiskTypes,
      dataDiskTypes,
      billingModes,
      purchaseDurationUnits,
    } = useCvmOptions(cond, formData);
    const { t } = useI18n();
    const { isResourcePage } = useWhereAmI();

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
    const handleEditGcpDataDisk = (index: number) => {
      dialogState.gcpDataDisk.isShow = true;
      dialogState.gcpDataDisk.isEdit = true;
      dialogState.gcpDataDisk.editDataIndex = index;
      dialogState.gcpDataDisk.formData = formData.data_disk[index];
    };

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
      cloudId.value = vpc.bk_cloud_id;
      vpcId.value = vpc.id;
      resetFormItemData('cloud_subnet_id');
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

      return rules[cond.vendor] || {
        validator: () => true,
        message: '',
      };
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
            return value >= awsMinMap[item.disk_type] && value <= awsMaxMap[item.disk_type];
          },
          message: `${awsMinMap[item.disk_type]}-${awsMaxMap[item.disk_type]}GB`,
          trigger: 'change',
        },
      };

      return rules[cond.vendor] || {
        validator: () => true,
        message: '',
      };
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

      return rules[cond.vendor] || {
        min: 1,
        max: Infinity,
      };
    });

    const submitDisabled = computed(() => isEmptyCond.value);

    const formConfigDataDiskDiff = computed(() => {
      const diffs = {
        [VendorEnum.GCP]: {
          content: () => <div class="form-content-list details">
              {
                formData.data_disk.map((item: IDiskOption, index: number) => (
                  <div class="flex-row">
                    {item.disk_name}, 空白, {item.disk_size_gb}GB, {dataDiskTypes.value.find((disk: IOption) => disk.id === item.disk_type)?.name || '--'}
                    <div class="btns">
                      <Button class="btn" outline size="small" onClick={() => handleEditGcpDataDisk(index)}><EditIcon /></Button>
                      <Button class="btn" outline size="small" onClick={() => handleRemoveDataDisk(index)}><CloseLineIcon /></Button>
                      {
                        (index === formData.data_disk.length - 1)
                        && <Button class="btn" outline size="small" onClick={handleCreateGcpDataDisk}><PlusIcon /></Button>
                      }
                    </div>
                  </div>
                ))
              }
              { !formData.data_disk.length && <div class="btns"><Button class="btn" onClick={handleCreateGcpDataDisk}><PlusIcon /></Button></div> }
            </div>,
        },
      };
      return diffs[cond.vendor] || {};
    });

    const formConfigPublicIpAssignedDiff = computed(() => {
      const diffs = {
        [VendorEnum.HUAWEI]: {
          label: '弹性公网IP',
          content: () => '暂不支持购买，请到EIP中绑定',
        },
      };
      return diffs[cond.vendor] || {};
    });

    // 当前 vpc下是否有子网列表
    const subnetLength = ref(0);
    watch(() => formData.cloud_vpc_id, () => {
      console.log('subnetSelectorRef.value', subnetSelectorRef.value.subnetList);
      subnetLength.value = subnetSelectorRef.value.subnetList?.length || 0;
    });

    watch(() => cond.vendor, () => {
      formData.system_disk.disk_type = '';
    });

    // const curRegionName = computed(() => {
    //   return hostStore.regionList?.find(region => region.region_id === cond.region) || {};
    // });

    const formConfig = computed(() => [
      {
        id: 'region',
        title: '地域',
        children: [
          {
            label: '可用区',
            required: cond.vendor === VendorEnum.AZURE ? zoneSelectorRef.value.list?.length > 0 : true,
            property: 'zone',
            rules: [{
              trigger: 'change',
            }],
            content: () => <ZoneSelector
              ref={zoneSelectorRef}
              v-model={formData.zone}
              vendor={cond.vendor}
              region={cond.region}
              onChange={handleZoneChange} />,
          },
        ],
      },
      {
        id: 'network',
        title: '网络',
        children: [
          {
            label: 'VPC',
            required: true,
            property: 'cloud_vpc_id',
            description: '',
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
                />
                {isResourcePage ? null : (
                  <Button
                    text
                    theme='primary'
                    onClick={() => {
                      if (!formData.cloud_vpc_id) return;
                      const url = `/#/business/vpc?cloud_id=${formData.cloud_vpc_id}&bizs=${cond.bizId}`;
                      window.open(url, '_blank');
                    }}>
                    详情
                  </Button>
                )}
              </div>
            ),
          },
          {
            label: '子网',
            required: true,
            description: '',
            property: 'cloud_subnet_id',
            content: () => (
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
                />
                {
                  isResourcePage
                    ? null
                    : <Button
                        text
                        theme="primary"
                        onClick={() => {
                          if (!formData.cloud_subnet_id) return;
                          const url = `/#/business/subnet?cloud_id=${formData.cloud_subnet_id}&bizs=${cond.bizId}`;
                          window.open(url, '_blank');
                        }}>
                        详情
                      </Button>
                }
              </div>
            ),
          },
          {
            label: '公网IP',
            display: ![VendorEnum.GCP, VendorEnum.AZURE].includes(cond.vendor),
            required: true,
            description: '',
            property: 'public_ip_assigned',
            content: () => <Checkbox v-model={formData.public_ip_assigned} disabled>自动分配公网IP</Checkbox>,
            ...formConfigPublicIpAssignedDiff.value,
          },
          {
            label: '管控区域',
            description: '',
            content: () => <CloudAreaName id={cloudId.value} />,
          },
          {
            label: '安全组',
            display: cond.vendor !== VendorEnum.GCP,
            required: true,
            description: '',
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
                  clearable={false}
                />
                {
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
                 }
              </div>
            ),
          },
        ],
      },
      {
        id: 'billing',
        title: '计费',
        display: [VendorEnum.TCLOUD, VendorEnum.HUAWEI].includes(cond.vendor),
        children: [
          {
            label: '计费模式',
            required: true,
            property: 'instance_charge_type',
            content: () =>
            // <Select v-model={formData.instance_charge_type} clearable={false}>{
            //     billingModes.value.map(({ id, name }: IOption) => (
            //       <Option key={id} value={id} label={name}></Option>
            //     ))
            //   }
            // </Select>,
            <RadioGroup v-model={formData.instance_charge_type}>
                    {billingModes.value.map(item => (<RadioButton label={item.id} >{item.name}
                  </RadioButton>))}
            </RadioGroup>,
          },
        ],
      },
      {
        id: 'config',
        title: '配置',
        children: [
          {
            label: '机型',
            required: true,
            description: '',
            property: 'instance_type',
            content: () => <MachineTypeSelector
              v-model={formData.instance_type}
              vendor={cond.vendor}
              accountId={cond.cloudAccountId}
              zone={formData.zone?.[0]}
              region={cond.region}
              bizId={cond.bizId ? cond.bizId : accountStore.bizs}
              instanceChargeType={formData.instance_charge_type}
              clearable={false}
              onChange={handleMachineTypeChange} />,
          },
          {
            label: '镜像',
            required: true,
            description: '',
            property: 'cloud_image_id',
            content: () => <Imagelector
              v-model={formData.cloud_image_id}
              vendor={cond.vendor}
              region={cond.region}
              machineType={machineType.value}
              clearable={false} />,
          },
        ],
      },
      {
        id: 'storage',
        title: '存储',
        children: [
          {
            label: '系统盘类型',
            required: true,
            content: [
              {
                property: 'system_disk.disk_type',
                required: true,
                content: () => <Select v-model={formData.system_disk.disk_type} style={{ width: '200px' }} clearable={false}>{
                    sysDiskTypes.value.map(({ id, name }: IOption) => (
                      <Option key={id} value={id} label={name}></Option>
                    ))
                  }
                </Select>,
              },
              {
                label: '大小',
                required: true,
                property: 'system_disk.disk_size_gb',
                rules: [sysDiskSizeRules.value],
                description: sysDiskSizeRules.value.message,
                content: () => <Input type='number' v-model={formData.system_disk.disk_size_gb} suffix="GB"></Input>,
              },
            ],
          },
          {
            label: '数据盘',
            tips: () => (cond.vendor === VendorEnum.TCLOUD ? '增强型SSD云硬盘仅在部分可用区开放售卖，后续将逐步增加售卖可用区' : ''),
            property: 'data_disk',
            content: () => <div class="form-content-list">
              {
                formData.data_disk.map((item: IDiskOption, index: number) => (
                  <div class="flex-row">
                    <FormItem property={`data_disk[${index}].disk_type`} rules={[]}>
                      <Select v-model={item.disk_type} style={{ width: '200px' }} clearable={false}>{
                          dataDiskTypes.value.map(({ id, name }: IOption) => (
                            <Option key={id} value={id} label={name}></Option>
                          ))
                        }
                      </Select>
                    </FormItem>
                    <FormItem
                      label='大小'
                      property={`data_disk[${index}].disk_size_gb`}
                      rules={[dataDiskSizeRules(item)]}
                      description={dataDiskSizeRules(item).message}
                    >
                      <Input type='number' style={{ width: '160px' }} v-model={item.disk_size_gb} suffix="GB"></Input>
                    </FormItem>
                    <FormItem
                      label='数量'
                      property={`data_disk[${index}].disk_count`}
                      min={dataDiskCountRules.value.min}
                      max={dataDiskCountRules.value.max}>
                      <Input style={{ width: '90px' }} type='number' v-model={item.disk_count}></Input>
                    </FormItem>
                    <div class="btns">
                      <Button class="btn" outline size="small" onClick={() => handleRemoveDataDisk(index)}><CloseLineIcon /></Button>
                      {
                        (index === formData.data_disk.length - 1)
                        && <Button class="btn" outline size="small" onClick={handleCreateDataDisk}><PlusIcon /></Button>
                      }
                    </div>
                  </div>
                ))
              }
              { !formData.data_disk.length && <div class="btns"><Button class="btn" onClick={handleCreateDataDisk}><PlusIcon /></Button></div> }
              {
                // (formData.data_disks.length > 0 && cond.vendor === VendorEnum.HUAWEI)
                // && <Checkbox v-model={formData.is_quickly_initialize_data_disk}>快速初始化数据盘</Checkbox>
              }
            </div>,
            ...formConfigDataDiskDiff.value,
          },
        ],
      },
      {
        id: 'auth',
        title: '登录',
        children: [
          {
            label: '设置密码',
            required: true,
            description: '字母数字与 ()\`~!@#$%^&*-+=|{}[]:;\',.?/ 字符的组合',
            content: [
              {
                property: 'username',
                display: cond.vendor === VendorEnum.AZURE,
                content: () => <Input placeholder='登录用户' v-model={formData.username}></Input>,
              },
              {
                property: 'password',
                content: () => <Input type='password' placeholder='密码' v-model={formData.password}></Input>,
              },
              {
                property: 'confirmed_password',
                content: () => <Input type='password' placeholder='确认密码' v-model={formData.confirmed_password}></Input>,
              },
            ],
          },
        ],
      },
      {
        id: 'quantity',
        title: '数量',
        children: [
          {
            label: '购买数量',
            required: true,
            property: 'required_count',
            description: '大于0的整数，最大不能超过100',
            content: () => <Input type='number' min={0} max={100} v-model={formData.required_count}></Input>,
          },
          {
            label: '购买时长',
            required: true,
            // PREPAID：包年包月
            display: ['PREPAID'].includes(formData.instance_charge_type),
            content: [
              {
                property: 'purchase_duration.count',
                content: () => <Input type='number' v-model={formData.purchase_duration.count}></Input>,
              },
              {
                property: 'purchase_duration.unit',
                content: () => <Select v-model={formData.purchase_duration.unit} clearable={false}>{
                  purchaseDurationUnits.map(({ id, name }: IOption) => (
                    <Option key={id} value={id} label={name}></Option>
                  ))}
                </Select>,
              },
              {
                property: 'auto_renew',
                content: () => <Checkbox v-model={formData.auto_renew}>自动续费</Checkbox>,
              },
            ],
          },
        ],
      },
      {
        id: 'describe',
        title: '主机描述',
        children: [
          {
            label: '实例名称',
            required: true,
            property: 'name',
            maxlength: 60,
            description: '60个字符，字母、数字、“-”，且必须以字母、数字开头和结尾。\n\r 实例名称是在云上的记录名称，并不是操作系统上的主机名，以方便使用名称来搜索主机。\n\r 如申请的是1台主机，则按填写的名称命名。如申请的是多台，则填写名称是前缀，申请单会自动补充随机的后缀。',
            content: () => <Input placeholder='填写实例名称，主机数量大于1时支持批量命名' v-model={formData.name} />,
          },
          {
            label: '实例备注',
            property: 'memo',
            content: () => <Input type='textarea' placeholder='填写实例备注' rows={3} maxlength={255} v-model={formData.memo}></Input>,
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
            const pattern = cond.vendor === VendorEnum.HUAWEI
              ? /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[()`~!@#$%^&*-+=|{}\[\]:;',.?/])[A-Za-z\d()`~!@#$%^&*\-+=|{}\[\]:;',.?/]+$/
              : /^(?=.*[A-Za-z])(?=.*\d)(?=.*[()`~!@#$%^&*-+=|{}\[\]:;',.?/])[A-Za-z\d()`~!@#$%^&*\-+=|{}\[\]:;',.?/]+$/;
            return pattern.test(value);
          },
          message: '密码复杂度不符合要求',
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
              '123', 'administrator', 'console', 'guest', 'test3',
              'user1', 'user5', 'admin1', 'test1', 'john', 'owner',
              'test', 'user4', 'david', 'root', 'support_388945a0',
              'user', 'user2', '1', 'support', 'video', 'a', 'admin',
              'sys', 'test2', 'admin2', 'aspnet', 'sql', 'user3',
              'actuser', 'adm', 'backup', 'server',
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

    return () => <ContentContainer>
      <ConditionOptions
        type={ResourceTypeEnum.CVM}
        v-model:bizId={cond.bizId}
        v-model:cloudAccountId={cond.cloudAccountId}
        v-model:vendor={cond.vendor}
        v-model:region={cond.region}
        v-model:resourceGroup={cond.resourceGroup}
      />
      <Form model={formData} rules={formRules} ref={formRef} onSubmit={handleFormSubmit}>
        {
          formConfig.value
            .filter(({ display }) => display !== false)
            .map(({ title, children }) => (
              <FormGroup title={title}>
                {
                  children
                    .filter(({ display }) => display !== false)
                    .map(({ label, description, tips, rules, required, property, content }) => (
                    <FormItem
                      label={label}
                      required={required}
                      property={property}
                      rules={rules}
                      description={description}
                    >
                      {
                        Array.isArray(content)
                          ? <div class="flex-row">
                            {
                              content
                                .filter(sub => sub.display !== false)
                                .map(sub => (
                                  <FormItem
                                    label={sub.label}
                                    required={sub.required}
                                    property={sub.property}
                                    rules={sub.rules}
                                    description={sub?.description}
                                  >
                                    {sub.content()}
                                    { sub.tips && <div class="form-item-tips">{sub.tips()}</div> }
                                  </FormItem>
                                ))
                            }
                          </div>
                          : content()
                      }
                      { tips && <div class="form-item-tips">{tips()}</div> }
                    </FormItem>
                    ))
                }
              </FormGroup>
            ))
        }
        <div class="action-bar">
          <Button theme='primary' loading={submitting.value} disabled={submitDisabled.value} onClick={handleFormSubmit}>{
            isResourcePage ? t('提交') : t('提交审批')
          }</Button>
          <Button>{ t('取消') }</Button>
        </div>
      </Form>
      <GcpDataDiskFormDialog
        v-model:isShow={dialogState.gcpDataDisk.isShow}
        isEdit={dialogState.gcpDataDisk.isEdit}
        dataDiskTypes={dataDiskTypes.value}
        formData={dialogState.gcpDataDisk.formData}
        onAdd={handleAddGcpDataDisk}
        onSave={handleSaveGcpDataDisk}
        onClose={() => dialogState.gcpDataDisk.isShow = false}
      />
    </ContentContainer>;
  },
});
