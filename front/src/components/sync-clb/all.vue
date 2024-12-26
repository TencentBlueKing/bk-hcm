<script lang="ts" setup>
import { reactive, ref, nextTick } from 'vue';
import { Message, Form, Dialog, Button } from 'bkui-vue';
import { useResourceStore } from '@/store';
import AccountSelector from '@/components/account-selector/index-new.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector';
import { useI18n } from 'vue-i18n';

export interface syncResourceForm {
  account_id: string;
  vendor: string;
  region: string;
  isShow: boolean;
}
export interface AllLoadBalancerProps {
  disabled: boolean;
}

defineOptions({ name: 'AllLoadBalancer' });

withDefaults(defineProps<AllLoadBalancerProps>(), {
  disabled: false,
});

const resourceStore = useResourceStore();
const { t } = useI18n();
const formRef = ref(null);
const { FormItem } = Form;

const formData = reactive<syncResourceForm>({
  account_id: '',
  vendor: '',
  region: '',
  isShow: false,
});

const handleSetFormDataInit = () => {
  formData.account_id = '';
  formData.vendor = '';
  formData.region = '';
  formData.isShow = false;
  nextTick(() => formRef.value?.clearValidate());
};
const handleConfirm = async () => {
  const { account_id, vendor, region } = formData;
  await formRef.value?.validate();
  resourceStore
    .syncResource({
      account_id,
      vendor,
      regions: [region],
      resource: 'load_balancer',
    })
    .then(() => {
      Message({ theme: 'success', message: t('已提交同步任务，请等待同步结果') });
      handleSetFormDataInit();
    });
};
</script>

<template>
  <Button class="mw88 mr8" @click="() => (formData.isShow = true)" :disabled="disabled">
    {{ t('同步负载均衡') }}
  </Button>
  <Dialog
    :is-show="formData.isShow"
    :title="t('同步负载均衡列表')"
    :quick-close="false"
    @confirm="handleConfirm"
    @closed="handleSetFormDataInit"
  >
    <Form form-type="vertical" ref="formRef" :model="formData">
      <FormItem :label="t('选择云账号')" required property="account_id">
        <account-selector v-model="formData.account_id" @change="(resource) => (formData.vendor = resource.vendor)" />
      </FormItem>
      <FormItem :label="t('云地域')" required property="region">
        <region-selector v-model="formData.region" :vendor="formData.vendor" :account-id="formData.account_id" />
      </FormItem>
      <div>{{ t('从云上同步该业务的所有负载均衡数据，包括负载均衡，监听器等') }}</div>
    </Form>
  </Dialog>
</template>
