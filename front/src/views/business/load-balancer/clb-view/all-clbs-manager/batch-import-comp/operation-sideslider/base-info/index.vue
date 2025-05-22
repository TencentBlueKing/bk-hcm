<script setup lang="ts">
import { computed, inject, nextTick, Ref, watch } from 'vue';
import { Alert, Form } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { VendorEnum } from '@/common/constant';
import { IAccountItem } from '@/typings/account';
import Step from '../components/step.vue';
import AccountSelector from '@/components/account-selector/index-new.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector.vue';

import { useI18n } from 'vue-i18n';
import { Action, LbBatchImportBaseInfo, Operation } from '../../types';
import { accountFilter } from '@/views/service/service-apply/components/common/condition-options/account-filter.plugin';

defineOptions({ name: 'LbBatchImportBaseInfoComp' });
const props = defineProps<{
  activeAction: Action;
  // 文件上传成功后，禁用基本信息的表单
  globalDisabled: boolean;
  formRef: InstanceType<typeof Form>;
}>();

const { t } = useI18n();

const { FormItem } = Form;

const formModel = inject<LbBatchImportBaseInfo>('clbBatchImportFormModel');
const bkBizId = inject<Ref<number>>('bk_biz_id');

const operationTypes = computed(() => {
  if (props.activeAction === Action.CREATE_LISTENER_OR_URL_RULE) {
    return [
      { type: Operation.create_layer4_listener, label: t('创建TCP/UDP监听器') },
      { type: Operation.create_layer7_listener, label: t('创建HTTP/HTTPS监听器') },
      { type: Operation.create_layer7_rule, label: t('创建URL规则') },
    ];
  }
  return [
    { type: Operation.binding_layer4_rs, label: t('四层监听器绑定RS') },
    { type: Operation.binding_layer7_rs, label: t('七层监听器绑定RS') },
  ];
});

// 默认operation type
watch(
  () => props.activeAction,
  (val) => {
    let defaultOperationType = Operation.create_layer4_listener;
    if (val === Action.BIND_RS) defaultOperationType = Operation.binding_layer4_rs;
    formModel.operation_type = defaultOperationType;
  },
  { immediate: true },
);

const handleAccountChange = (
  account: IAccountItem,
  oldAccount: IAccountItem,
  vendorAccountMap: Map<VendorEnum, IAccountItem[]>,
) => {
  const accountList = vendorAccountMap.get(account?.vendor);
  // 本地缓存中的账号可能在当前业务下不存在，当不存在时重置账号数据避免无法重新选择
  if (accountList?.some((item) => item.id === account.id)) {
    formModel.vendor = account.vendor;
  } else {
    formModel.account_id = '';
    formModel.vendor = undefined;
  }
  formModel.region_ids = [];
  nextTick(() => props.formRef.clearValidate());
};
</script>

<template>
  <Step :step="1" :title="t('信息录入')">
    <FormItem :label="t('云账号')" property="account_id" required>
      <AccountSelector
        v-model="formModel.account_id"
        :biz-id="bkBizId"
        :filter="accountFilter"
        :disabled="globalDisabled"
        @change="handleAccountChange"
      />
    </FormItem>

    <FormItem :label="t('云地域')" property="region_ids" required>
      <!-- todo: 如果是华为云，需要确定一下 type props 的值 -->
      <RegionSelector
        v-model="formModel.region_ids"
        multiple
        :vendor="formModel.vendor"
        :is-disabled="!formModel.account_id || globalDisabled"
      />
    </FormItem>

    <FormItem :label="t('操作类型')" property="operation_type" required>
      <BkRadioGroup v-model="formModel.operation_type" :disabled="globalDisabled">
        <BkRadioButton v-for="{ type, label } in operationTypes" :key="type" :label="type">
          {{ label }}
        </BkRadioButton>
      </BkRadioGroup>
    </FormItem>

    <Alert theme="warning">
      <template #title>
        <template v-if="Action.CREATE_LISTENER_OR_URL_RULE === activeAction">
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
    </Alert>
  </Step>
</template>

<style scoped></style>
