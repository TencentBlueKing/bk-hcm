import { computed, defineComponent } from 'vue';
import { Form, Input, Select, Checkbox, Button, Radio } from 'bkui-vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import ConditionOptions from '../components/common/condition-options.vue';
import CommonCard from '@/components/CommonCard';
import ZoneSelector from '../components/common/zone-selector';

import type { IOption } from '@/typings/common';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import useDiskOptions from '../hooks/use-disk-options';
import useCondtion from '../hooks/use-condtion';
import useDiskFormData from '../hooks/use-disk-form-data';
import './index.scss';
import { useI18n } from 'vue-i18n';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useRouter } from 'vue-router';

const { FormItem } = Form;
const { Option } = Select;
const { Group: RadioGroup, Button: RadioButton } = Radio;

export default defineComponent({
  props: {},
  setup() {
    const { cond, isEmptyCond } = useCondtion(ResourceTypeEnum.DISK);
    const { formData, formRef, handleFormSubmit, submitting } = useDiskFormData(cond);
    const { diskTypes, billingModes, purchaseDurationUnits } = useDiskOptions(cond, formData);
    const { t } = useI18n();
    const { isResourcePage } = useWhereAmI();
    const router = useRouter();

    const submitDisabled = computed(() => isEmptyCond.value);

    const nameRules = computed(() => {
      const rules = {
        [VendorEnum.TCLOUD]: {
          pattern: /^\S{1,60}$/,
          message: '最多60个字符',
          trigger: 'change',
        },
        [VendorEnum.HUAWEI]: {
          pattern: /^\S{1,100}$/,
          message: '最多100个字符',
          trigger: 'change',
        },
        [VendorEnum.AWS]: {
          pattern: /^\S{1,255}$/,
          message: '最多255个字符',
          trigger: 'change',
        },
        [VendorEnum.GCP]: {
          pattern: /^[a-z][a-z0-9-]{0,61}(?<!-)$/,
          message: '必须以小写字母开头，后面最多可跟 62 个小写字母、数字或连字符，但不能以连字符结尾',
          trigger: 'change',
        },
        [VendorEnum.AZURE]: {
          pattern: /^[a-z0-9][\w-.]{0,79}(?<!-)$/,
          message:
            '名称必须以字母或数字开头，以字母、数字或下划线结尾，并且只能包含字母、数字、下划线、句点或连字符。该值的长度不得超过 80',
          trigger: 'change',
        },
      };

      return rules[cond.vendor] || {};
    });

    const dataDiskSizeRules = (disk_type: string) => {
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
      const min = awsMinMap[disk_type] || 20;
      const max = awsMaxMap[disk_type] || 32000;
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
            return value >= min && value <= max;
          },
          message: `${disk_type === '' ? '请选择云硬盘类型' : `${min}-${max}GB`}`,
          trigger: 'change',
        },
      };

      return {
        rules: rules[cond.vendor] || {
          validator: () => true,
          message: '',
        },
        min,
        max,
      };
    };

    const formConfig = computed(() => [
      {
        id: 'base',
        title: '硬盘信息',
        children: [
          {
            label: '名称',
            display: ![VendorEnum.AWS].includes(cond.vendor),
            required: true,
            property: 'disk_name',
            rules: [nameRules.value],
            // description: nameRules.value.message,
            content: () => <Input placeholder='填写硬盘的名称' v-model={formData.disk_name} />,
          },
          {
            label: '可用区',
            required: true,
            property: 'zone',
            content: () => <ZoneSelector v-model={formData.zone} vendor={cond.vendor} region={cond.region} />,
          },
          {
            label: '云硬盘类型',
            required: true,
            property: 'disk_type',
            content: () => (
              <Select v-model={formData.disk_type} clearable={false} onChange={() => (formData.disk_size = null)}>
                {diskTypes.value.map(({ id, name }: IOption) => (
                  <Option key={id} value={id} label={name}></Option>
                ))}
              </Select>
            ),
          },
          {
            label: '大小',
            required: true,
            property: 'disk_size',
            rules: [dataDiskSizeRules(formData.disk_type).rules],
            content: () => (
              <Input
                type='number'
                min={dataDiskSizeRules(formData.disk_type).min}
                max={dataDiskSizeRules(formData.disk_type).max}
                step={10}
                v-model={formData.disk_size}
                suffix='GB'></Input>
            ),
          },
          {
            label: '购买数量',
            required: true,
            property: 'disk_count',
            content: () => <Input type='number' min={1} max={500} v-model={formData.disk_count}></Input>,
          },
          {
            label: '计费模式',
            display: [VendorEnum.TCLOUD, VendorEnum.HUAWEI].includes(cond.vendor),
            required: true,
            property: 'disk_charge_type',
            content: () => (
              <RadioGroup v-model={formData.disk_charge_type}>
                {billingModes.value.map(({ id, name }: IOption) => (
                  <RadioButton label={id}>{name}</RadioButton>
                ))}
              </RadioGroup>
            ),
          },
          {
            label: '购买时长',
            display: ['PREPAID', 'prePaid'].includes(formData.disk_charge_type),
            required: true,
            content: [
              {
                property: 'purchase_duration.size',
                content: () => <Input type='number' min={1} v-model={formData.purchase_duration.count}></Input>,
              },
              {
                property: 'purchase_duration.unit',
                content: () => (
                  <Select v-model={formData.purchase_duration.unit} clearable={false}>
                    {purchaseDurationUnits.map(({ id, name }: IOption) => (
                      <Option key={id} value={id} label={name}></Option>
                    ))}
                  </Select>
                ),
              },
            ],
          },
          {
            label: '自动续费',
            display: ['PREPAID', 'prePaid'].includes(formData.disk_charge_type),
            required: true,
            property: 'auto_renew',
            content: () => <Checkbox v-model={formData.auto_renew}>账号余额足够时，设备到期后按月自动续费</Checkbox>,
          },
          {
            label: '申请单备注',
            property: 'memo',
            content: () => (
              <Input
                type='textarea'
                placeholder='简要描述申请硬盘的理由，30字以内。'
                rows={2}
                maxlength={30}
                resize={false}
                v-model={formData.memo}></Input>
            ),
          },
        ],
      },
    ]);

    const formRules = {};

    return () => (
      <div>
        <DetailHeader>
          <p class={'purchase-dish-header-title'}>购买硬盘</p>
        </DetailHeader>
        <div class='create-form-container disk-wrap' style={isResourcePage && { padding: 0 }}>
          <Form model={formData} rules={formRules} ref={formRef} onSubmit={handleFormSubmit} formType='vertical'>
            <ConditionOptions
              type={ResourceTypeEnum.DISK}
              v-model:bizId={cond.bizId}
              v-model:cloudAccountId={cond.cloudAccountId}
              v-model:vendor={cond.vendor}
              v-model:region={cond.region}
              v-model:resourceGroup={cond.resourceGroup}
            />
            {formConfig.value
              .filter(({ display }) => display !== false)
              .map(({ title, children }) => (
                <CommonCard title={() => title} class={'mb16'}>
                  {children
                    .filter(({ display }) => display !== false)
                    .map(({ label, description, tips, rules, required, property, content }) => (
                      <FormItem
                        label={label}
                        required={required}
                        property={property}
                        rules={rules}
                        description={description}>
                        {Array.isArray(content) ? (
                          <div class='flex-row'>
                            {content
                              .filter((sub) => sub.display !== false)
                              .map((sub) => (
                                <FormItem
                                  label={sub.label}
                                  required={sub.required}
                                  property={sub.property}
                                  description={sub?.description}
                                  class={'mr8'}>
                                  {sub.content()}
                                  {sub.tips && <div class='form-item-tips'>{sub.tips()}</div>}
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
            <div class='action-bar' style={{ paddingLeft: isResourcePage && 'calc(15% + 24px)' }}>
              <Button
                theme='primary'
                loading={submitting.value}
                disabled={submitDisabled.value}
                class={'mr8'}
                onClick={handleFormSubmit}>
                {isResourcePage ? t('提交') : t('提交审批')}
              </Button>
              <Button onClick={() => router.back()}>{t('取消')}</Button>
            </div>
          </Form>
        </div>
      </div>
    );
  },
});
