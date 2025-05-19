<!-- eslint-disable no-nested-ternary -->
<script setup lang="ts">
import { Dialog, Form, Message } from 'bkui-vue';
import { ref, computed, h, reactive, watch, useTemplateRef } from 'vue';
import { useResourceStore, useBusinessStore } from '@/store';

import UsageBizValue from '../../components/security/usage-biz-value.vue';
import SecurityGroupManagerSelector from '@/views/resource/resource-manage/children/components/security/manager-selector/index.vue';
import { useI18n } from 'vue-i18n';
import {
  azureSourceAddressTypes,
  AzureSourceTypeArr,
  azureTargetAddressTypes,
  AzureTargetTypeArr,
} from '@/views/resource/resource-manage/children/components/security/add-rule/vendors/azure';
import {
  awsSourceAddressTypes,
  AwsSourceTypeArr,
} from '@/views/resource/resource-manage/children/components/security/add-rule/vendors/aws';
import {
  tcloudSourceAddressTypes,
  TcloudSourceTypeArr,
} from '@/views/resource/resource-manage/children/components/security/add-rule/vendors/tcloud';
import { huaweiSourceAddressTypes } from '@/views/resource/resource-manage/children/components/security/add-rule/vendors/huawei';
import { SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum, QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';

export interface IData {
  [key: string]: any;
}
export interface ICloneSecurityProps {
  isShow: boolean;
  data: IData;
}

const props = defineProps<ICloneSecurityProps>();

const emit = defineEmits(['update:isShow', 'success']);
const { t } = useI18n();
// tab 信息
const types = [
  { name: 'ingress', label: t('入站规则') },
  { name: 'egress', label: t('出站规则') },
];

const states = reactive<any>({
  dataList: [],
  isLoading: true,
});
const filter = ref({ op: QueryRuleOPEnum.AND, rules: [{ field: 'type', op: QueryRuleOPEnum.EQ, value: 'ingress' }] });
const personSelectorRef = ref(null);
const formModel = reactive({ name: `${props.data.name}-copy` });
const formRef = useTemplateRef<typeof Form>('formRef');
const rules = {
  name: [{ validator: (val: string) => val.length <= 60, message: t('名称超过60个字符的长度限制，请调整后重试') }],
};

const vendor = computed(() => props?.data?.vendor);

const inColumns: any = computed(() =>
  [
    {
      label: t('名称'),
      field: 'name',
      isShow: vendor.value === 'azure',
    },
    {
      label: t('优先级'),
      field: 'priority',
      isShow: vendor.value === 'huawei' || vendor.value === 'azure',
    },
    {
      label: t('源地址类型'),
      render({ row }: any) {
        const nowVendor = (vendor.value as VendorEnum) || VendorEnum.TCLOUD;
        const sourceMap: any = {
          [VendorEnum.AWS]: {
            types: awsSourceAddressTypes,
            arr: AwsSourceTypeArr,
          },
          [VendorEnum.AZURE]: {
            types: azureSourceAddressTypes,
            arr: AzureSourceTypeArr,
          },
          [VendorEnum.HUAWEI]: {
            types: huaweiSourceAddressTypes,
            arr: TcloudSourceTypeArr,
          },
          [VendorEnum.TCLOUD]: {
            types: tcloudSourceAddressTypes,
            arr: TcloudSourceTypeArr,
          },
        };
        const { types } = sourceMap[nowVendor];
        const { arr } = sourceMap[nowVendor];
        const map = new Map(types.map((item: { value: string; label: string }) => [item.value, item.label]));
        let k = '';
        arr.forEach((type: string) => row[type] && (k = type));
        return map.get(k) || '--';
      },
      isShow: true,
    },
    {
      label: t('源地址'),
      render({ row }: any) {
        return h('span', {}, [
          row.cloud_address_group_id ||
            row.cloud_address_id ||
            row.cloud_service_group_id ||
            row.cloud_target_security_group_id ||
            row.ipv4_cidr ||
            row.ipv6_cidr ||
            row.cloud_remote_group_id ||
            row.remote_ip_prefix ||
            (row.source_address_prefix === '*' ? t('ALL') : row.source_address_prefix) ||
            row.source_address_prefixes ||
            row.cloud_source_security_group_ids ||
            row.destination_address_prefix ||
            row.destination_address_prefixes ||
            row.cloud_destination_security_group_ids ||
            (row?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
        ]);
      },
      isShow: true,
    },
    {
      label: t('源端口'),
      render({ row }: any) {
        return (row.source_port_range === '*' ? 'ALL' : row.source_port_range) || '--';
      },
      isShow: vendor.value === 'azure',
    },
    {
      label: t('目标地址类型'),
      render({ row }: any) {
        const map = new Map(
          azureTargetAddressTypes.map((item: { value: string; label: string }) => [item.value, item.label]),
        );
        let k = '';
        AzureTargetTypeArr.forEach((type: string) => row[type] && (k = type));
        return map.get(k) || '--';
      },
      isShow: vendor.value === 'azure',
    },
    {
      label: t('类型'),
      field: 'ethertype',
      isShow: vendor.value === 'huawei',
    },

    {
      label: t('目标地址'),
      render({ row }: any) {
        return (
          (row.destination_address_prefix === '*' ? t('ALL') : row.destination_address_prefix) ||
          row.destination_address_prefixes ||
          row.cloud_destination_security_group_ids
        );
      },
      isShow: vendor.value === 'azure',
    },
    {
      label: vendor.value === 'azure' ? t('目标端口协议类型') : t('协议'),
      render({ row }: any) {
        return h('span', {}, [
          row.cloud_service_id ||
            (vendor.value === 'aws' && row.protocol === '-1'
              ? t('ALL')
              : vendor.value === 'huawei' && !row.protocol
              ? t('ALL')
              : vendor.value === 'azure' && row.protocol === '*'
              ? t('ALL')
              : `${row.protocol}`),
        ]);
      },
      isShow: true,
    },
    {
      label: vendor.value === 'azure' ? t('目标协议端口') : t('端口'),
      render({ row }: any) {
        return h('span', {}, [
          row.cloud_service_id ||
            (vendor.value === 'aws' && row.to_port === -1
              ? t('ALL')
              : vendor.value === 'huawei' && !row.port
              ? t('ALL')
              : vendor.value === 'azure' && row.destination_port_range === '*'
              ? t('ALL')
              : `${row.port || row.to_port || row.destination_port_range || row.destination_port_ranges || '--'}`),
        ]);
      },
      isShow: true,
    },
    {
      label: t('策略'),
      render({ row }: any) {
        return h('span', {}, [
          vendor.value === 'huawei'
            ? HuaweiSecurityRuleEnum[row.action]
            : vendor.value === 'azure'
            ? AzureSecurityRuleEnum[row.access]
            : vendor.value === 'aws'
            ? t('允许')
            : SecurityRuleEnum[row.action] || '--',
        ]);
      },
      isShow: vendor.value !== 'aws',
    },
    {
      label: t('备注'),
      field: 'memo',
      render: ({ row }: any) => row.memo || '--',
      isShow: true,
    },
  ].filter(({ isShow }) => !!isShow),
);

// 出站规则列字段
const outColumns: any = computed(() =>
  [
    {
      label: t('名称'),
      field: 'name',
      isShow: vendor.value === 'azure',
    },
    {
      label: t('优先级'),
      field: 'priority',
      isShow: vendor.value === 'huawei' || vendor.value === 'azure',
    },
    {
      label: t('源地址类型'),
      render({ row }: any) {
        const map = new Map(
          azureSourceAddressTypes.map((item: { value: string; label: string }) => [item.value, item.label]),
        );
        let k = '';
        AzureSourceTypeArr.forEach((type: string) => row[type] && (k = type));
        return map.get(k) || '--';
      },
      isShow: vendor.value === 'azure',
    },
    {
      label: t('源地址'),
      render({ row }: any) {
        return h('span', {}, [
          row.cloud_address_group_id ||
            row.cloud_address_id ||
            row.cloud_service_group_id ||
            row.cloud_target_security_group_id ||
            row.ipv4_cidr ||
            row.ipv6_cidr ||
            row.cloud_remote_group_id ||
            row.remote_ip_prefix ||
            (row.source_address_prefix === '*' ? t('ALL') : row.source_address_prefix) ||
            row.source_address_prefixes ||
            row.cloud_source_security_group_ids ||
            row.destination_address_prefix ||
            row.destination_address_prefixes ||
            row.cloud_destination_security_group_ids ||
            (row?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
        ]);
      },
      isShow: vendor.value === 'azure',
    },
    {
      label: t('源端口'),
      render({ row }: any) {
        return (row.source_port_range === '*' ? 'ALL' : row.source_port_range) || '--';
      },
      isShow: vendor.value === 'azure',
    },
    {
      label: t('目标地址类型'),
      render({ row }: any) {
        const nowVendor = vendor.value as VendorEnum;
        const targetMap: any = {
          [VendorEnum.AWS]: {
            types: awsSourceAddressTypes,
            arr: AwsSourceTypeArr,
          },
          [VendorEnum.AZURE]: {
            types: azureTargetAddressTypes,
            arr: AzureTargetTypeArr,
          },
          [VendorEnum.HUAWEI]: {
            types: huaweiSourceAddressTypes,
            arr: TcloudSourceTypeArr,
          },
          [VendorEnum.TCLOUD]: {
            types: tcloudSourceAddressTypes,
            arr: TcloudSourceTypeArr,
          },
        };
        const { types } = targetMap[nowVendor];
        const { arr } = targetMap[nowVendor];
        const map = new Map(types.map((item: { value: string; label: string }) => [item.value, item.label]));
        let k = '';
        arr.forEach((type: string) => row[type] && (k = type));
        return map.get(k) || '--';
      },
      isShow: true,
    },
    {
      label: t('目标地址'),
      render({ row }: any) {
        return h('span', {}, [
          row.cloud_address_group_id ||
            row.cloud_address_id ||
            row.cloud_service_group_id ||
            row.cloud_target_security_group_id ||
            row.ipv4_cidr ||
            row.ipv6_cidr ||
            row.cloud_remote_group_id ||
            row.remote_ip_prefix ||
            row.cloud_source_security_group_ids ||
            (row.destination_address_prefix === '*' ? t('ALL') : row.destination_address_prefix) ||
            row.destination_address_prefixes ||
            row.cloud_destination_security_group_ids ||
            (row?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
        ]);
      },
      isShow: true,
    },
    {
      label: t('类型'),
      field: 'ethertype',
      isShow: vendor.value === 'huawei',
    },
    {
      label: vendor.value === 'azure' ? t('目标端口协议类型') : t('协议'),
      render({ row }: any) {
        return h('span', {}, [
          row.cloud_service_id ||
            (vendor.value === 'aws' && row.protocol === '-1'
              ? t('ALL')
              : vendor.value === 'huawei' && !row.protocol
              ? t('ALL')
              : vendor.value === 'azure' && row.protocol === '*'
              ? t('ALL')
              : `${row.protocol}`),
        ]);
      },
      isShow: true,
    },
    {
      label: vendor.value === 'azure' ? t('目标协议端口') : t('端口'),
      render({ row }: any) {
        return h('span', {}, [
          row.cloud_service_id ||
            (vendor.value === 'aws' && row.to_port === -1
              ? t('ALL')
              : vendor.value === 'huawei' && !row.port
              ? t('ALL')
              : vendor.value === 'azure' && row.destination_port_range === '*'
              ? t('ALL')
              : `${row.port || row.to_port || row.destination_port_range || '--'}`),
        ]);
      },
      isShow: true,
    },
    {
      label: t('策略'),
      render({ row }: any) {
        return h('span', {}, [
          vendor.value === 'huawei'
            ? HuaweiSecurityRuleEnum[row.action]
            : vendor.value === 'azure'
            ? AzureSecurityRuleEnum[row.access]
            : vendor.value === 'aws'
            ? t('允许')
            : SecurityRuleEnum[row.action] || '--',
        ]);
      },
      isShow: vendor.value !== 'aws',
    },
    {
      label: t('备注'),
      field: 'memo',
      render: ({ row }: any) => row.memo || '--',
      isShow: true,
    },
  ].filter(({ isShow }) => !!isShow),
);

const activeType = ref('ingress');
const useBusiness = useBusinessStore();
const resourceStore = useResourceStore();
const getList = async () => {
  states.isLoading = true;
  try {
    const list = await resourceStore.getAllSort({
      id: props?.data?.id,
      vendor: vendor.value,
      filter: filter.value,
    });
    states.dataList = list;
    return list;
  } catch {
    states.dataList = [];
  } finally {
    states.isLoading = false;
  }
};
const handleClose = () => {
  emit('update:isShow', false);
  personSelectorRef?.value?.reset?.();
};
const isConfirmLoading = ref(false);
const handleConfirm = async () => {
  const { id } = props.data;
  const { formData: personSelectorParams, validate } = personSelectorRef.value;
  const { bak_manager, manager } = personSelectorParams;
  await validate();
  await formRef.value.validate();
  isConfirmLoading.value = true;
  try {
    await useBusiness.cloneSecurity({
      id,
      name: formModel.name,
      bak_manager,
      manager,
    });
    Message({ theme: 'success', message: t('克隆成功！') });
    handleClose();
    emit('success');
  } finally {
    isConfirmLoading.value = false;
  }
};

watch(
  () => props.isShow,
  (val: boolean) => {
    if (val) getList();
  },
  {
    immediate: true,
  },
);
watch(
  () => activeType.value,
  (val: string) => {
    states.isLoading = true;
    filter.value.rules[0].value = val;
    getList();
  },
);
</script>

<template>
  <Dialog
    width="960"
    class="clone-security-dialog"
    :is-show="props.isShow"
    :title="t('克隆安全组')"
    theme="primary"
    @closed="handleClose"
    @confirm="handleConfirm"
    :is-loading="isConfirmLoading"
  >
    <div class="security-info">
      <div class="info-wrap">
        <span class="label">{{ t('管理业务：') }}</span>
        <display-value :property="{ type: 'business' }" :value="props.data.mgmt_biz_id" />
      </div>
      <div>
        <div class="info-wrap usage-bizs">
          <span class="label">{{ t('使用业务：') }}</span>
          <usage-biz-value :value="props.data.usage_biz_ids" />
        </div>
      </div>
    </div>
    <bk-form ref="formRef" :model="formModel" :rules="rules" form-type="vertical">
      <bk-form-item :label="t('安全组名称')" property="name" label-width="100">
        <bk-input v-model.trim="formModel.name" />
      </bk-form-item>
    </bk-form>
    <SecurityGroupManagerSelector
      ref="personSelectorRef"
      :manager="props?.data?.manager"
      :bak-manager="props?.data?.bak_manager"
    ></SecurityGroupManagerSelector>
    <div class="security-rule">
      <div class="title">{{ t('安全组规则') }}</div>
      <section class="rule-main">
        <bk-radio-group v-model="activeType">
          <bk-radio-button v-for="{ name, label } in types" :key="name" :label="name">
            {{ label }}
          </bk-radio-button>
        </bk-radio-group>
      </section>
      <bk-table
        class="mt20"
        row-hover="auto"
        :columns="activeType === 'ingress' ? inColumns : outColumns"
        :data="states.dataList"
        show-overflow-tooltip
        max-height="300"
        v-bkloading="{ loading: states.isLoading }"
      >
        <template #empty>
          <div class="security-empty-container">
            <bk-exception
              class="exception-wrap-item exception-part"
              type="empty"
              scene="part"
              :description="t('无规则，默认拒绝所有流量')"
            />
          </div>
        </template>
      </bk-table>
    </div>
  </Dialog>
</template>

<style scoped lang="scss">
.security-info {
  margin-bottom: 24px;
  display: flex;
  align-items: center;
  gap: 100px;
  font-size: 14px;
  color: #4d4f56;

  .info-wrap {
    display: flex;
    align-items: center;

    .label {
      width: 80px;
    }

    &.usage-bizs {
      width: 320px;

      :deep(.flex-tag) {
        width: calc(100% - 80px);
      }
    }
  }
}

.security-rule {
  .title {
    margin-bottom: 16px;
    font-weight: 700;
    font-size: 14px;
    color: #4d4f56;
    line-height: 22px;
  }
}
</style>
