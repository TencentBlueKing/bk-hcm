<template>
  <bk-dialog
    :is-show="model"
    :title="title"
    :quick-close="false"
    :is-loading="loading"
    @confirm="handleConfirm"
    @closed="handleClosed"
    @hidden="emit('hidden')"
  >
    <bk-alert theme="info" :title="desc" class="mb16" />
    <bk-form ref="sync-form" form-type="vertical" :model="formModel" :rules="rules">
      <bk-form-item :label="t('云账号')" property="account_id" required>
        <account-selector
          v-model="formModel.account_id"
          :biz-id="businessId"
          :disabled="!!initialModel?.account_id"
          :resource-type="resourceType"
          @change="(resource) => (formModel.vendor = resource?.vendor)"
        />
      </bk-form-item>
      <bk-form-item :label="t('云地域')" property="regions" required>
        <region-selector
          v-model="formModel.regions"
          :vendor="formModel.vendor"
          :account-id="formModel.account_id"
          :multiple="multipleRegion"
          :disabled="!!initialModel?.regions"
        />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref, useTemplateRef, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { useResourceStore } from '@/store';

import { Message, Form } from 'bkui-vue';
import AccountSelector from '@/components/account-selector/index-new.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector';
import { ResourceTypeEnum } from '@/common/constant';

interface IProps {
  resourceType: ResourceTypeEnum;
  title: string;
  desc: string;
  resourceName: string;
  businessId?: number;
  multipleRegion?: boolean;
  initialModel?: IModel;
}

interface IModel {
  account_id: string;
  vendor: string;
  regions: string | string[];
  cloud_ids?: string[];
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits<{ success: []; hidden: [] }>();

const { t } = useI18n();
const resourceStore = useResourceStore();

const syncForm = useTemplateRef<typeof Form>('sync-form');
const rules = {
  regions: [
    {
      validator: (val: string[]) => (props.multipleRegion ? val.length <= 5 : !!val),
      message: t('最多选择5个地域'),
      trigger: 'change',
    },
  ],
  cloud_ids: [
    {
      validator: (val: string[]) => val.length <= 20,
      message: t('最多选择20个资源'),
      trigger: 'change',
    },
  ],
};
const formModel = reactive<IModel>({
  account_id: '',
  vendor: '',
  regions: props.multipleRegion ? [] : '',
});
watchEffect(() => {
  if (props.initialModel) {
    Object.assign(formModel, props.initialModel);
  }
});
const reset = () => {
  Object.assign(formModel, { account_id: '', vendor: '', regions: [] });
  nextTick(() => syncForm.value.clearValidate());
};

const loading = ref(false);
const handleConfirm = async () => {
  const { resourceName } = props;
  const { account_id: accountId, vendor, regions, cloud_ids: cloudIds } = formModel;

  await syncForm.value.validate();
  loading.value = true;
  try {
    await resourceStore.syncResource(vendor, accountId, resourceName, {
      regions: props.multipleRegion ? (regions as string[]) : [regions as string],
      cloud_ids: cloudIds,
    });
    Message({ theme: 'success', message: t('已提交同步任务，请等待同步结果') });
    reset();
    handleClosed();
    emit('success');
  } catch (error) {
    console.error(error);
    return Promise.reject();
  } finally {
    loading.value = false;
  }
};

const handleClosed = () => {
  model.value = false;
};
</script>
