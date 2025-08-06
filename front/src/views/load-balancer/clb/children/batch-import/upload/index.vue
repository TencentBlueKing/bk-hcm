<script setup lang="ts">
import { computed, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { UploadFile, UploadFiles } from 'bkui-vue/lib/upload/upload.type';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import {
  ILoadBalancerBatchImportModel,
  ILoadBalancerImportPreview,
  LoadBalancerImportPreviewResData,
} from '../typings';
import { LoadBalancerBatchImportOperationType } from '@/views/load-balancer/constants';

import { Upload } from 'bkui-vue';
import Step from '../../step.vue';

defineOptions({ name: 'LbBatchImportUploadComp' });

const props = defineProps<{
  formModel: ILoadBalancerBatchImportModel;
}>();

const emit = defineEmits<{
  (e: 'previewSuccess', data: ILoadBalancerImportPreview): void;
  (e: 'previewError' | 'fileDelete'): void;
}>();

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const { t } = useI18n();
const { getBusinessApiPath } = useWhereAmI();
let files: UploadFiles = [];
const uploadRef = useTemplateRef<typeof Upload>('upload');

const url = computed(() => {
  const { vendor, operation_type } = props.formModel;
  return `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/load_balancers/operations/${operation_type}/preview`;
});
const formDataAttributes = computed(() => {
  const { account_id, region_ids } = props.formModel;

  return [
    { name: 'account_id', value: account_id },
    { name: 'region_ids', value: JSON.stringify(region_ids) },
  ];
});
const isDisabled = computed(() => {
  const { account_id, region_ids, operation_type } = props.formModel;
  return !(account_id && region_ids.length && operation_type);
});

const handleSuccess = (response: LoadBalancerImportPreviewResData, file: UploadFile, fileList: UploadFiles) => {
  emit('previewSuccess', response.data);
  fileList.find((item) => item.uid === file.uid).statusText = t('上传成功');
};
const handleError = () => {
  emit('previewError');
};
const handleDelete = () => {
  emit('fileDelete');
};
const handleDone = (fileList: UploadFiles) => {
  files = [...fileList];
};

// 下载模板文件
const handleDownloadTemplateFile = async () => {
  const filenameMap = {
    [LoadBalancerBatchImportOperationType.create_layer4_listener]: '1_hcm_clb_tcp_udp_listener_template.xlsx',
    [LoadBalancerBatchImportOperationType.create_layer7_listener]: '2_hcm_clb_http_https_listener_template.xlsx',
    [LoadBalancerBatchImportOperationType.create_layer7_rule]: '3_hcm_clb_url_rule_template.xlsx',
    [LoadBalancerBatchImportOperationType.binding_layer4_rs]: '4_hcm_clb_bind_rs_tcp_udp_template.xlsx',
    [LoadBalancerBatchImportOperationType.binding_layer7_rs]: '5_hcm_clb_bind_rs_http_https_template.xlsx',
  };
  http.download({ url: `/api/v1/web/templates/${filenameMap[props.formModel.operation_type]}`, method: 'get' });
};

const clearFiles = () => {
  files.forEach((file) => {
    uploadRef.value.handleRemove(file);
  });
  files = [];
};

defineExpose({ clearFiles });
</script>

<template>
  <step :step="2" :title="t('上传文件')">
    <bk-upload
      ref="upload"
      :multiple="false"
      :limit="1"
      accept=".xlsx"
      :validate-name="/\.xlsx$/i"
      with-credentials
      name="file"
      :url="url"
      :form-data-attributes="formDataAttributes"
      :disabled="isDisabled"
      @success="handleSuccess"
      @error="handleError"
      @delete="handleDelete"
      @done="handleDone"
    />
    <section class="tips">
      {{ t('仅支持.xlsx格式的文件，不能超过5千行，下载') }}
      <bk-button theme="primary" text @click="handleDownloadTemplateFile" class="ml4">{{ t('模板文件') }}</bk-button>
    </section>
  </step>
</template>

<style scoped lang="scss">
.tips {
  margin-top: 8px;
  font-size: 12px;
}
</style>
