<script setup lang="ts">
import { computed, provide, ref, useTemplateRef } from 'vue';
import { useRouter } from 'vue-router';
import { Form, Message } from 'bkui-vue';

import CommonSideslider from '@/components/common-sideslider';
import BaseInfo from './base-info/index.vue';
import IUpload from './upload/index.vue';
import Preview from './preview/index.vue';

import { useI18n } from 'vue-i18n';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import useFormModel from '@/hooks/useFormModel';
import { Action, LbBatchImportBaseInfo, LbImportPreview } from '../types';
import http from '@/http';
import { localStorageActions } from '@/common/util';

defineOptions({ name: 'LbBatchImportOperationSideslider' });
const props = defineProps<{ activeAction: Action }>();

const router = useRouter();
const { t } = useI18n();
const { whereAmI, getBizsId, getBusinessApiPath } = useWhereAmI();

// 抽屉
const isShow = ref(false);
const title = computed(() => {
  return props.activeAction === Action.CREATE_LISTENER_OR_URL_RULE ? t(`批量创建监听器及规则`) : t('批量绑定RS');
});
const show = () => (isShow.value = true);

// base-info - 表单
const lastFormModel = localStorageActions.get('bk-hcm-lb-batch-import-form-model', (value) => JSON.parse(value));
const { formModel } = useFormModel<LbBatchImportBaseInfo>(
  lastFormModel || {
    account_id: '',
    vendor: undefined,
    region_ids: [],
    operation_type: undefined,
  },
);
const bkBizId = computed(() => (whereAmI.value === Senarios.business ? getBizsId() : undefined));
provide('clbBatchImportFormModel', formModel);
provide('bk_biz_id', bkBizId);
const isBaseInfoHasEmpty = computed(() => !formModel.account_id || formModel.region_ids.length === 0);
const isBaseInfoDisabled = ref(false);

// 上传文件 - 预览
const previewData = ref<LbImportPreview>(); // 预览结果
const handlePreviewSuccess = (res: LbImportPreview) => {
  previewData.value = res;
  // 上传成功，不允许变更基本信息
  isBaseInfoDisabled.value = true;
};
const handlePreviewError = () => {
  // 预览失败，清空预览数据
  clearPreviewData();
};
const handleFileDelete = () => {
  // 手动删除文件，清空预览数据
  clearPreviewData();
};
const clearPreviewData = () => {
  previewData.value = null;
  // 未上传文件，允许变更基本信息
  isBaseInfoDisabled.value = false;
};

// 提交
const previewRef = useTemplateRef<typeof Preview>('preview');
const isSubmitDisabled = computed(
  () => isBaseInfoHasEmpty.value || !previewData.value || previewRef.value?.info.notExecutableCount !== 0,
);
const submitTooltipsOption = computed(() => {
  // 基本信息未录入
  if (isBaseInfoHasEmpty.value) {
    return { disabled: !isBaseInfoHasEmpty.value, content: t('请录入云账号、云地域、操作类型等信息') };
  }
  // 未上传文件，或预览失败
  if (!previewData.value) {
    return { disabled: !!previewData.value, content: t('请上传文件') };
  }
  // 预览成功，但存在不可执行的数据
  if (previewRef.value?.info.notExecutableCount !== 0) {
    return {
      disabled: previewRef.value?.info.notExecutableCount === 0,
      content: t('请参考参数校验结果对excel中不可执行的数据进行修正'),
    };
  }
  return { disabled: true };
});

const isSubmitLoading = ref(false);
const handleSubmit = async () => {
  isSubmitLoading.value = true;
  try {
    const { vendor, operation_type } = formModel;
    const res = await http.post(
      `/api/v1/cloud/${getBusinessApiPath()}vendor/${vendor}/load_balancers/operations/${operation_type}/submit`,
      { ...previewData.value, source: 'excel' },
    );
    Message({ theme: 'success', message: t('提交成功') });

    // 记录当前选中的云账号、云地域、操作类型 - formModel
    localStorageActions.set('bk-hcm-lb-batch-import-form-model', formModel);

    // todo: 改用name跳转
    router.push({ path: `/business/task/clb/details${res.data.task_management_id}`, query: { bizs: bkBizId.value } });
  } finally {
    isSubmitLoading.value = false;
  }
};

defineExpose({ show });
</script>

<template>
  <CommonSideslider
    v-model:is-show="isShow"
    :title="title"
    width="1280"
    :confirm-text="t('确认并提交')"
    :is-submit-loading="isSubmitLoading"
    :is-submit-disabled="isSubmitDisabled"
    :submit-tooltips-option="submitTooltipsOption"
    @handle-submit="handleSubmit"
  >
    <Form :model="formModel">
      <!-- 1. 信息录入 -->
      <BaseInfo ref="base-info" :active-action="activeAction" :global-disabled="isBaseInfoDisabled" />

      <!-- 2. 上传文件 -->
      <IUpload
        ref="upload"
        :form-model="formModel"
        @preview-success="handlePreviewSuccess"
        @preview-error="handlePreviewError"
        @file-delete="handleFileDelete"
      />

      <!-- 3. 结果预览 -->
      <Preview
        ref="preview"
        :form-model="formModel"
        :data="previewData?.details"
        :is-base-info-empty="isBaseInfoHasEmpty"
      />
    </Form>
  </CommonSideslider>
</template>

<style scoped></style>
