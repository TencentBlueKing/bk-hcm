import { computed, defineComponent, reactive, ref, watch } from 'vue';
import { Form, Input, Select, Checkbox, Button } from 'bkui-vue';
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

import type { IOption } from '@/typings/common';
import type { IDiskOption } from '../hooks/use-cvm-form-data';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import useCvmOptions from '../hooks/use-cvm-options';
import useCondtion from '../hooks/use-condtion';
import useCvmFormData, { getDataDiskDefaults, getGcpDataDiskDefaults } from '../hooks/use-cvm-form-data';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  props: {},
  setup(props, ctx) {
    const { cond, isEmptyCond } = useCondtion(ResourceTypeEnum.CVM);
    const { formData, formRef, handleFormSubmit, submitting, resetFormItemData } = useCvmFormData(cond);
    const {
      sysDiskTypes,
      dataDiskTypes,
      billingModes,
      purchaseDurationUnits,
    } = useCvmOptions(cond, formData);

    const dialogState = reactive({
      gcpDataDisk: {
        isShow: false,
        isEdit: false,
        editDataIndex: null,
        formData: getGcpDataDiskDefaults(),
      },
    });

    const cloudId = ref(null);
    const vpcId = ref('');

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
    };
    const handleVpcChange = (vpc: any) => {
      cloudId.value = vpc.bk_cloud_id;
      vpcId.value = vpc.id
      resetFormItemData('cloud_subnet_id');
    };

    const submitDisabled = computed(() => isEmptyCond.value);

    const formConfigDataDiskDiff = computed(() => {
      const diffs = {
        [VendorEnum.GCP]: {
          tips: () => <div>添加磁盘后，需要登录机器挂载和格式化，<Button text theme='primary'>参考文档</Button></div>,
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

    const formConfig = computed(() => [
      {
        id: 'region',
        title: '地域',
        children: [
          {
            label: '可用区',
            required: true,
            property: 'zone',
            rules: [{
              trigger: 'change',
            }],
            content: () => <ZoneSelector
              v-model={formData.zone}
              vendor={cond.vendor}
              region={cond.region}
              onChange={handleZoneChange} />,
          },
        ],
      },
      {
        id: 'config',
        title: '配置',
        children: [
          {
            label: '名称',
            required: true,
            property: 'name',
            maxlength: 32,
            description: '1.以小写字母开头，支持短横线，下划线。 2.最长32个字符。3.不能以连字符结尾',
            content: () => <Input placeholder='填写主机的名称' v-model={formData.name} />,
          },
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
              clearable={false} />,
          },
          {
            label: '镜像',
            required: true,
            description: '',
            property: 'cloud_image_id',
            content: () => <Imagelector v-model={formData.cloud_image_id} vendor={cond.vendor} region={cond.region} clearable={false} />,
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
            content: () => <VpcSelector
              v-model={formData.cloud_vpc_id}
              bizId={cond.bizId}
              accountId={cond.cloudAccountId}
              vendor={cond.vendor}
              region={cond.region}
              onChange={handleVpcChange}
              clearable={false} />,
          },
          {
            label: '子网',
            required: true,
            description: '',
            property: 'cloud_subnet_id',
            content: () => <SubnetSelector
              v-model={formData.cloud_subnet_id}
              bizId={cond.bizId}
              vpcId={vpcId.value}
              region={cond.region}
              clearable={false} />,
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
            label: '所属的蓝鲸云区域',
            description: '',
            content: () => <CloudAreaName id={cloudId.value} />,
          },
          {
            label: '安全组',
            display: cond.vendor !== VendorEnum.GCP,
            required: true,
            description: '',
            property: 'cloud_security_group_ids',
            content: () => <SecurityGroupSelector
              v-model={formData.cloud_security_group_ids}
              bizId={cond.bizId}
              accountId={cond.cloudAccountId}
              region={cond.region}
              multiple={cond.vendor !== VendorEnum.AZURE}
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
                content: () => <Select v-model={formData.system_disk.disk_type} clearable={false}>{
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
                description: '容量大小限制：40-1024GB',
                content: () => <Input type='number' v-model={formData.system_disk.disk_size_gb} suffix="GB"></Input>,
              },
            ],
          },
          {
            label: '数据盘',
            description: '容量大小限制：10-32768GB',
            content: () => <div class="form-content-list">
              {
                formData.data_disk.map((item: IDiskOption, index: number) => (
                  <div class="flex-row">
                    <FormItem property={`data_disk[${index}].disk_type`} rules={[]}>
                      <Select v-model={item.disk_type} clearable={false}>{
                          dataDiskTypes.value.map(({ id, name }: IOption) => (
                            <Option key={id} value={id} label={name}></Option>
                          ))
                        }
                      </Select>
                    </FormItem>
                    <FormItem label='大小' property={`data_disk[${index}].disk_size_gb`} min={10} max={32768}>
                      <Input type='number' v-model={item.disk_size_gb} suffix="GB"></Input>
                    </FormItem>
                    <FormItem label='数量' property={`data_disk[${index}].disk_count`} min={1} max={20}>
                      <Input style={{ width: '65px' }} type='number' v-model={item.disk_count}></Input>
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
        id: 'billing',
        title: '计费',
        display: [VendorEnum.TCLOUD, VendorEnum.HUAWEI].includes(cond.vendor),
        children: [
          {
            label: '计费模式',
            required: true,
            property: 'instance_charge_type',
            content: () => <Select v-model={formData.instance_charge_type} clearable={false}>{
                billingModes.value.map(({ id, name }: IOption) => (
                  <Option key={id} value={id} label={name}></Option>
                ))
              }
            </Select>,
          },
          {
            label: '购买时长',
            required: true,
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
        id: 'quantity',
        title: '数量',
        children: [
          {
            label: '购买数量',
            required: true,
            property: 'required_count',
            description: '大于0的整数，最大不能超过500',
            content: () => <Input type='number' v-model={formData.required_count}></Input>,
          },
          {
            label: '备注',
            property: 'memo',
            content: () => <Input type='textarea' rows={3} maxlength={255} v-model={formData.memo}></Input>,
          },
        ],
      },
    ]);

    const formRules = {
      name: [
        {
          pattern: /^[a-z][\w-]{0,31}(?<!-)$/,
          message: '以小写字母开头，支持短横线，下划线，不能以连字符结尾，最长32个字符',
          trigger: 'change',
        },
      ],
      password: [
        {
          validator: (value: string) => value.length >= 8 && value.length <= 30,
          message: '长度在8-30个字符之间',
          trigger: 'change',
        },
        {
          validator: (value: string) => {
            const pattern = cond.vendor === VendorEnum.HUAWEI
              ? /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[()`~!@#$%^&*-+=|{}\[\]:;',.?/])[A-Za-z\d()`~!@#$%^&*-+=|{}\[\]:;',.?/]+$/
              : /^(?=.*[A-Za-z])(?=.*\d)(?=.*[()`~!@#$%^&*-+=|{}\[\]:;',.?/])[A-Za-z\d()`~!@#$%^&*-+=|{}\[\]:;',.?/]+$/
            return pattern.test(value)
          },
          message: '不符合校验规则',
          trigger: 'change',
        },
        {
          validator: (value: string) => {
            formRef.value.clearValidate('confirmed_password');
            if (formData.confirmed_password.length) {
              return value === formData.confirmed_password
            }
            return true
          },
          message: '两次输入的密码不一致',
          trigger: 'change',
        },
      ],
      confirmed_password: [
        {
          validator: (value: string) => value.length >= 8 && value.length <= 30,
          message: '长度在8-30个字符之间',
          trigger: 'change',
        },
        {
          validator: (value: string) => {
            formRef.value.clearValidate('password');
            return formData.password.length && value === formData.password;
          },
          message: '两次输入的密码不一致',
          trigger: 'change',
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
              'actuser', 'adm', 'backup', 'server'
            ]
            return !sensitives.includes(value)
          },
          message: '不允许使用的用户名',
          trigger: 'change',
        }
      ]
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
          <Button theme='primary' loading={submitting.value} disabled={submitDisabled.value} onClick={handleFormSubmit}>提交审批</Button>
          <Button>取消</Button>
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
