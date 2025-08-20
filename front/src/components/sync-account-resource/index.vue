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
          :resource-type="resourceType"
          :disabled="globalDisabled"
          @change="handleAccountChange"
        />
      </bk-form-item>
      <bk-form-item
        v-if="formModel.vendor === VendorEnum.AZURE"
        :label="t('资源组')"
        property="resource_group_names"
        required
      >
        <resource-group-selector
          v-model="formModel.resource_group_names"
          :account-id="formModel.account_id"
          :vendor="formModel.vendor"
          :multiple="multipleResourceGroup"
          :disabled="globalDisabled"
        />
      </bk-form-item>
      <bk-form-item v-else :label="t('云地域')" property="regions" required>
        <region-selector
          v-model="formModel.regions"
          :vendor="formModel.vendor"
          :multiple="multipleRegion"
          :disabled="globalDisabled"
        />
      </bk-form-item>
      <bk-form-item v-if="globalDisabled" :label="t('名称')">
        <bk-input :value="initialModel.name" disabled />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, useTemplateRef, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { useResourceStore } from '@/store';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import { IAccountItem } from '@/typings';

import { Message, Form } from 'bkui-vue';
import AccountSelector from '@/components/account-selector/index-new.vue';
import ResourceGroupSelector from '@/views/service/service-apply/components/common/resource-group-selector.vue';
import RegionSelector from '@/views/service/service-apply/components/common/region-selector.vue';

interface IProps {
  title: string;
  desc: string;
  resourceType?: ResourceTypeEnum; // 用于账号过滤
  businessId?: number; // 用于账号过滤
  resourceName: string; // load_balancer | security_group
  multipleRegion?: boolean;
  multipleResourceGroup?: boolean;
  initialModel?: IModel;
  errorHandler?: (error: any) => void; // 业务报错catch
}

interface IModel {
  name?: string;
  account_id: string;
  vendor: string;
  regions?: string | string[]; // 指定资源同步地域，最少1，最大5
  cloud_ids?: string[]; // 资源id，数量上限20
  resource_group_names?: string | string[]; // 指定资源同步的资源组，最少1，最大5（azure）
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits<{ success: []; hidden: [] }>();

const { t } = useI18n();
const resourceStore = useResourceStore();

const globalDisabled = computed(() => Boolean(props.initialModel?.cloud_ids?.length));

const syncForm = useTemplateRef<typeof Form>('sync-form');
const rules = {
  regions: [
    {
      validator: (val: string | string[]) => (props.multipleRegion ? (val as string[]).length <= 5 : !!(val as string)),
      message: t('指定资源同步地域，最少1，最大5'),
      trigger: 'change',
    },
  ],
  resource_group_names: [
    {
      validator: (val: string | string[]) =>
        props.multipleResourceGroup ? (val as string[]).length <= 5 : !!(val as string),
      message: t('指定资源同步的资源组，最少1，最大5'),
      trigger: 'change',
    },
  ],
};

const formModel = reactive<IModel>({ account_id: '', vendor: '' });
const handleAccountChange = (resource: IAccountItem) => {
  formModel.vendor = resource?.vendor;
  // 如果不是全局禁用，则清空地域和资源组
  if (!globalDisabled.value) {
    formModel.regions = props.multipleRegion ? [] : '';
    formModel.resource_group_names = props.multipleResourceGroup ? [] : '';
  }
  nextTick(() => syncForm.value.clearValidate());
};
watchEffect(() => {
  if (props.initialModel) {
    Object.assign(formModel, props.initialModel);
  }
});

const loading = ref(false);
const buildRequestBody = () => {
  const { vendor, regions, resource_group_names: resourceGroupNames, cloud_ids: cloudIds } = formModel;
  const extensionParams =
    VendorEnum.AZURE === vendor
      ? { resource_group_names: (props.multipleResourceGroup ? resourceGroupNames : [resourceGroupNames]) as string[] }
      : { regions: (props.multipleRegion ? regions : [regions]) as string[] };

  return { cloud_ids: cloudIds, ...extensionParams };
};
const handleConfirm = async () => {
  const { resourceName } = props;
  const { account_id: accountId, vendor } = formModel;

  await syncForm.value.validate();
  loading.value = true;
  const requestConfig = props.errorHandler ? { globalError: false } : {};
  try {
    const res = await resourceStore.syncResource(vendor, accountId, resourceName, buildRequestBody(), requestConfig);

    if (res.code !== 0 && props.errorHandler) {
      props.errorHandler(res);
      return;
    }

    Message({ theme: 'success', message: t('已同步成功') });
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
