<script setup lang="ts">
import { computed, inject, nextTick, Ref, watch } from 'vue';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import { IAccountItem } from '@/typings/account';
import { useI18n } from 'vue-i18n';
import { LoadBalancerActionType, LoadBalancerBatchImportOperationType } from '@/views/load-balancer/constants';
import { ILoadBalancerBatchImportModel } from '../typings';

import { Form } from 'bkui-vue';
import Step from '../../step.vue';
import AccountSelector from '@/components/account-selector/index-new.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector.vue';

defineOptions({ name: 'LbBatchImportBaseInfoComp' });
const props = defineProps<{
  activeAction: LoadBalancerActionType;
  // 文件上传成功后，禁用基本信息的表单
  globalDisabled: boolean;
  formRef: typeof Form;
}>();

const { t } = useI18n();

const formModel = inject<ILoadBalancerBatchImportModel>('clbBatchImportFormModel');
const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const operationTypes = computed(() => {
  if (props.activeAction === LoadBalancerActionType.CREATE_LISTENER_OR_RULES) {
    return [
      { type: LoadBalancerBatchImportOperationType.create_layer4_listener, label: t('创建TCP/UDP监听器') },
      { type: LoadBalancerBatchImportOperationType.create_layer7_listener, label: t('创建HTTP/HTTPS监听器') },
      { type: LoadBalancerBatchImportOperationType.create_layer7_rule, label: t('创建URL规则') },
    ];
  }
  return [
    { type: LoadBalancerBatchImportOperationType.binding_layer4_rs, label: t('四层监听器绑定RS') },
    { type: LoadBalancerBatchImportOperationType.binding_layer7_rs, label: t('七层监听器绑定RS') },
  ];
});

// 默认operation type
watch(
  () => props.activeAction,
  (val) => {
    let defaultOperationType = LoadBalancerBatchImportOperationType.create_layer4_listener;
    if (val === LoadBalancerActionType.BIND_RS) {
      defaultOperationType = LoadBalancerBatchImportOperationType.binding_layer4_rs;
    }
    formModel.operation_type = defaultOperationType;
  },
  { immediate: true },
);

const handleAccountChange = (
  account: IAccountItem,
  _oldAccount: IAccountItem,
  vendorAccountMap: Map<VendorEnum, IAccountItem[]>,
) => {
  const accountList = vendorAccountMap.get(account?.vendor);
  // 本地缓存中的账号可能在当前业务下不存在，当不存在时重置账号数据避免无法重新选择
  if (accountList?.some((item) => item.id === account.id)) {
    formModel.vendor = account.vendor;
  } else {
    formModel.account_id = '';
    formModel.vendor = undefined;
    formModel.region_ids = [];
  }
  nextTick(() => props.formRef.clearValidate());
};
</script>

<template>
  <step :step="1" :title="t('信息录入')">
    <bk-form-item :label="t('云账号')" property="account_id" required>
      <account-selector
        v-model="formModel.account_id"
        :biz-id="currentGlobalBusinessId"
        :resource-type="ResourceTypeEnum.CLB"
        :disabled="globalDisabled"
        @change="handleAccountChange"
      />
    </bk-form-item>

    <bk-form-item :label="t('云地域')" property="region_ids" required>
      <!-- TODO-CLB: 如果是华为云，需要确定一下 type props 的值 -->
      <region-selector
        v-model="formModel.region_ids"
        multiple
        :vendor="formModel.vendor"
        :is-disabled="!formModel.account_id || globalDisabled"
      />
    </bk-form-item>

    <bk-form-item :label="t('操作类型')" property="operation_type" required>
      <bk-radio-group v-model="formModel.operation_type" :disabled="globalDisabled">
        <bk-radio-button v-for="{ type, label } in operationTypes" :key="type" :label="type">
          {{ label }}
        </bk-radio-button>
      </bk-radio-group>
    </bk-form-item>

    <bk-alert theme="warning">
      <template #title>
        <template v-if="LoadBalancerActionType.CREATE_LISTENER_OR_RULES === activeAction">
          <div>{{ t('1.导入的数据中，必须匹配所选的账号、云地域。') }}</div>
          <div>{{ t('2.当导入的数据中存在监听器，请处理数据后再重新导入。') }}</div>
        </template>
        <template v-else>
          <div class="font-bold">{{ t('绑定RS的场景有2种：') }}</div>
          <div class="pl10">{{ t('1.绑定RS到TCP、UDP的监听器中') }}</div>
          <div class="pl10">{{ t('2.绑定RS到HTTP、HTTPS的URL规则上') }}</div>
          <div>{{ t('注意：如监听器或URL中已存在RS，新的RS会追加绑定，已有的RS仍存留不受影响。') }}</div>
        </template>
      </template>
    </bk-alert>
  </step>
</template>
