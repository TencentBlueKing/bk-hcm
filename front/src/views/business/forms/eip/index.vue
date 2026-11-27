<script setup lang="ts">
import { BusinessFormFilter } from '@/typings';
import { computed, ref, useTemplateRef, watch } from 'vue';
import FormSelect from '@/views/business/components/form-select.vue';
import aws from './aws.vue';
import azure from './azure.vue';
import gcp from './gcp.vue';
import huawei from './huawei.vue';
import tcloud from './tcloud.vue';
import { useBusinessStore } from '@/store/business';
import { useAccountStore } from '@/store/account';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';

// define props
const props = defineProps<{
  isFormDataChanged: boolean;
  show: boolean;
}>();
// define emits
const emits = defineEmits(['cancel', 'success', 'update:isFormDataChanged']);

// use hooks
const businessStore = useBusinessStore();
const accountStore = useAccountStore();

// define var
const isSubmitting = ref(false);
const formRef = useTemplateRef<[() => Promise<unknown>]>('vendor-form');
const selection = ref<BusinessFormFilter>({ account_id: '', vendor: '', region: '' });
const vendorFormData = ref<Record<string, unknown>>({});
const componentMap = {
  aws,
  azure,
  gcp,
  huawei,
  tcloud,
};
const { isResourcePage } = useWhereAmI();
const { t } = useI18n();
const vendorForm = computed(() => {
  const { vendor } = selection.value;
  return Object.prototype.hasOwnProperty.call(componentMap, vendor)
    ? componentMap[vendor as keyof typeof componentMap]
    : undefined;
});
const submitRegion = computed(() => {
  if (selection.value.vendor === VendorEnum.GCP && vendorFormData.value.region === 'global') return 'global';
  return selection.value.region;
});
const submitDisabled = computed(
  () =>
    !props.show ||
    !selection.value.account_id ||
    !submitRegion.value ||
    !vendorForm.value ||
    typeof formRef.value?.[0] !== 'function',
);

watch(
  () => props.show,
  (show) => {
    if (show) return;
    selection.value = { account_id: '', vendor: '', region: '' };
    vendorFormData.value = {};
  },
);

// define method
const handleFormFilter = (value: BusinessFormFilter) => {
  if (selection.value.account_id !== value.account_id) vendorFormData.value = {};
  selection.value = { ...value };
  !props.isFormDataChanged && emits('update:isFormDataChanged', true);
};

const handleFormChange = (value: Record<string, unknown>) => {
  vendorFormData.value = value;
};

const handleSubmit = async () => {
  const form = formRef.value;
  if (submitDisabled.value || isSubmitting.value || !form) return;

  const selected = selection.value;
  const region = submitRegion.value;
  const isCurrentForm = () => props.show && form === formRef.value && selected === selection.value;
  isSubmitting.value = true;

  try {
    try {
      await form[0]();
    } catch (error) {
      if (error instanceof Error) {
        Message({ theme: 'error', message: t('弹性 IP 表单校验失败，请重试') });
      }
      return;
    }
    if (!isCurrentForm() || region !== submitRegion.value) return;
    await businessStore.addEip(
      accountStore.bizs as number,
      { ...vendorFormData.value, account_id: selected.account_id, region },
      isResourcePage,
    );
    Message({
      theme: 'success',
      message: t('新增成功'),
    });
    if (isCurrentForm()) {
      emits('success');
      handleCancel();
    }
  } catch {
    // 创建接口的错误反馈由 HTTP 层统一处理。
  } finally {
    isSubmitting.value = false;
  }
};

const handleCancel = () => {
  emits('cancel');
};
</script>

<template>
  <FormSelect v-if="props.show" @change="handleFormFilter" type="eip" :show="props.show" />
  <component
    v-if="props.show && vendorForm"
    :key="selection.account_id"
    class="mb20 pdr20"
    ref="vendor-form"
    :is="vendorForm"
    :region="String(selection.region || '')"
    :vendor="selection.vendor"
    @change="handleFormChange"
  />
  <section class="btn-group">
    <bk-button
      theme="primary"
      class="mr10 btn"
      :loading="isSubmitting"
      :disabled="submitDisabled"
      @click="handleSubmit"
    >
      提交创建
    </bk-button>
    <bk-button class="btn" :disabled="isSubmitting" @click="handleCancel">取消</bk-button>
  </section>
</template>

<style lang="scss" scoped>
.pdr20 {
  padding-right: 20px;
}

.btn-group {
  margin-left: 150px;

  .btn {
    min-width: 88px;
  }
}
</style>
