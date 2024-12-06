<script setup lang="ts">
import { computed } from 'vue';
import { Button, Upload } from 'bkui-vue';
import Step from '../components/step.vue';

import { useI18n } from 'vue-i18n';
import { LbBatchImportBaseInfo, LbImportPreview, LbImportPreviewResData, Operation } from '../../types';
import { UploadFile, UploadFiles } from 'bkui-vue/lib/upload/upload.type';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

defineOptions({ name: 'LbBatchImportUploadComp' });
const props = defineProps<{
  formModel: LbBatchImportBaseInfo;
}>();
const emit = defineEmits<{
  (e: 'previewSuccess', data: LbImportPreview): void;
  (e: 'previewError' | 'fileDelete'): void;
}>();

const { t } = useI18n();
const { getBusinessApiPath } = useWhereAmI();

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

const handleSuccess = (response: LbImportPreviewResData, file: UploadFile, fileList: UploadFiles) => {
  emit('previewSuccess', response.data);
  fileList.find((item) => item.uid === file.uid).statusText = t('上传成功');
};
const handleError = () => {
  emit('previewError');
};
const handleDelete = () => {
  emit('fileDelete');
};

// 下载模板文件
const handleDownloadTemplateFile = async () => {
  const filenameMap = {
    [Operation.create_layer4_listener]: '1_hcm_clb_tcp_udp_listener_template.xlsx',
    [Operation.create_layer7_listener]: '2_hcm_clb_http_https_listener_template.xlsx',
    [Operation.create_layer7_rule]: '3_hcm_clb_bind_rs_url_ruler_template.xlsx',
    [Operation.binding_layer4_rs]: '4_hcm_clb_bind_rs_tcp_udp_template.xlsx',
    [Operation.binding_layer7_rs]: '5_hcm_clb_url_rule_http_https_template.xlsx',
  };
  http.download({ url: `/api/v1/web/templates/${filenameMap[props.formModel.operation_type]}`, method: 'get' });
};
</script>

<template>
  <Step :step="2" :title="t('上传文件')">
    <Upload
      ref="upload"
      :multiple="false"
      :limit="1"
      accept=".xlsx"
      :validate-name="/\.xlsx$/i"
      with-credentials
      name="file"
      :url="url"
      :form-data-attributes="formDataAttributes"
      @success="handleSuccess"
      @error="handleError"
      @delete="handleDelete"
    />
    <section class="tips">
      {{ t('仅支持.xlsx格式的文件，不能超过5千行，下载') }}
      <Button theme="primary" text @click="handleDownloadTemplateFile" class="ml4">{{ t('模板文件') }}</Button>
    </section>
  </Step>
</template>

<style scoped lang="scss">
.tips {
  margin-top: 8px;
  font-size: 12px;
}
</style>
