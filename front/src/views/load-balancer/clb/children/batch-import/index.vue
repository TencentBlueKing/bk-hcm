<script setup lang="ts">
import { computed, inject, provide, reactive, ref, Ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { LoadBalancerActionType } from '@/views/load-balancer/constants';
import { localStorageActions } from '@/common/util';
import { ILoadBalancerBatchImportModel, ILoadBalancerImportPreview } from './typings';
import { ResourceTypeEnum } from '@/common/constant';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';

import { Form, Message } from 'bkui-vue';
import BaseInfo from './base-info/index.vue';
import Upload from './upload/index.vue';
import Preview from './preview/index.vue';
import ModalFooter from '@/components/modal/modal-footer.vue';

const model = defineModel<boolean>();
const props = defineProps<{
  activeAction: LoadBalancerActionType.CREATE_LISTENER_OR_RULES | LoadBalancerActionType.BIND_RS;
}>();

const { t } = useI18n();
const loadBalancerClbStore = useLoadBalancerClbStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const title = computed(() => {
  return props.activeAction === LoadBalancerActionType.CREATE_LISTENER_OR_RULES
    ? t(`批量创建监听器及规则`)
    : t('批量绑定RS');
});

const formRef = useTemplateRef<typeof Form>('batch-import-clb-form');
const lastFormModel = localStorageActions.get('bk-hcm-lb-batch-import-form-model', (value) => JSON.parse(value));
const formModel = reactive<ILoadBalancerBatchImportModel>(
  lastFormModel || { account_id: '', vendor: undefined, region_ids: [], operation_type: undefined },
);
provide('clbBatchImportFormModel', formModel);

const isBaseInfoHasEmpty = computed(() => !formModel.account_id || formModel.region_ids.length === 0);
const isBaseInfoDisabled = ref(false);

// 上传文件 - 预览
const previewData = ref<ILoadBalancerImportPreview>(); // 预览结果
const handlePreviewSuccess = (res: ILoadBalancerImportPreview) => {
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
const submitTooltips = computed(() => {
  // 基本信息未录入
  if (isBaseInfoHasEmpty.value) {
    return { disabled: !isBaseInfoHasEmpty.value, content: t('请录入云账号、云地域、操作类型等信息') };
  }
  // 未上传文件，或预览失败
  if (!previewData.value) {
    return { disabled: !!previewData.value, content: t('请上传文件') };
  }
  // 文件解析失败
  if (!previewData.value?.details) {
    return { disabled: !!previewData.value?.details, content: t('预览失败，请检查文件内容格式') };
  }
  // 预览成功，但存在不可执行的数据
  if (previewRef.value?.info.notExecutableCount !== 0) {
    return {
      disabled: previewRef.value?.info.notExecutableCount === 0,
      content: t('请参考参数校验结果对excel中不可执行的数据进行修正'),
    };
  }
  return { disabled: true, content: '' };
});

const handleSubmit = async () => {
  const { account_id, region_ids, vendor, operation_type } = formModel;
  const data = { account_id, region_ids, source: 'excel', ...previewData.value };

  const res = await loadBalancerClbStore.batchImportLoadBalancer(
    vendor,
    operation_type,
    data,
    currentGlobalBusinessId.value,
  );

  Message({ theme: 'success', message: t('提交成功') });

  // 记录当前选中的云账号、云地域、操作类型 - formModel
  localStorageActions.set('bk-hcm-lb-batch-import-form-model', formModel);

  routerAction.redirect({
    name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
    query: { bizs: currentGlobalBusinessId.value },
    params: { resourceType: ResourceTypeEnum.CLB, id: res.data.task_management_id },
  });
};

const handleClosed = () => {
  model.value = false;
};

const uploadRef = useTemplateRef<typeof Upload>('upload');
watch(model, (val) => {
  if (!val) {
    uploadRef.value?.clearFiles();
    clearPreviewData();
  }
});
</script>

<template>
  <bk-sideslider v-model:is-show="model" :title="title" width="1280" class="batch-import-clb-sideslider">
    <bk-form ref="batch-import-clb-form" :model="formModel">
      <base-info
        ref="base-info"
        :active-action="activeAction"
        :global-disabled="isBaseInfoDisabled"
        :form-ref="formRef"
      />
      <upload
        ref="upload"
        :form-model="formModel"
        @preview-success="handlePreviewSuccess"
        @preview-error="handlePreviewError"
        @file-delete="handleFileDelete"
      />
      <preview
        ref="preview"
        :form-model="formModel"
        :data="previewData?.details"
        :is-base-info-empty="isBaseInfoHasEmpty"
      />
    </bk-form>
    <template #footer>
      <modal-footer
        :confirm-text="t('确认并提交')"
        :loading="loadBalancerClbStore.batchImportLoadBalancerLoading"
        :disabled="isSubmitDisabled"
        :tooltips="submitTooltips"
        @confirm="handleSubmit"
        @closed="handleClosed"
      />
    </template>
  </bk-sideslider>
</template>

<style scoped lang="scss">
.batch-import-clb-sideslider {
  :deep(.bk-sideslider-content) {
    padding: 24px 40px;
  }
}
</style>
