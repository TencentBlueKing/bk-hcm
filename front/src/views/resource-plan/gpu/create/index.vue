<script setup lang="ts">
import { ref, computed, onMounted, useTemplateRef } from 'vue';
import { Message, Upload } from 'bkui-vue';
import { Eye } from 'bkui-vue/lib/icon';
import type { UploadFile, UploadFiles } from 'bkui-vue/lib/upload/upload.type';
import { getModel } from '@/model/manager';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useResourcePlanStore } from '@/store/resource-plan';
import { useGpuDemandStore, type IGpuDemandItem } from '@/store/resource-plan/gpu-demand';
import { timeFormatter } from '@/common/util';
import WName from '@/components/w-name';
import ExcelPreviewDialog from '../components/excel-preview-dialog.vue';
import type { IExcelImportData } from '../hooks/use-excel-preview';
import { buildSubmitDetails } from '../hooks/use-excel-preview';
import { CreateForm } from './form-fields';

// ==================== Props / Emits ====================
type SliderMode = 'create' | 'reimport';

const model = defineModel<boolean>({ default: false });

const props = withDefaults(
  defineProps<{
    /** 弹窗模式：create=新增GPU需求, reimport=重新导入GPU需求 */
    mode?: SliderMode;
    /** 重新导入模式下，当前主单详情（用于展示上半部分信息） */
    orderDetail?: IGpuDemandItem | null;
  }>(),
  {
    mode: 'create',
    orderDetail: null,
  },
);

const emit = defineEmits<{
  hidden: [];
  success: [];
}>();

const isCreateMode = computed(() => props.mode === 'create');
const sliderTitle = computed(() => (isCreateMode.value ? '新增GPU需求' : '重新导入 GPU 需求'));

// ==================== 基础依赖 ====================
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { getBizsId } = useWhereAmI();
const resourcePlanStore = useResourcePlanStore();
const gpuDemandStore = useGpuDemandStore();

// ==================== 新增模式：表单相关 ====================
const formModel = getModel(CreateForm);
const formProperties = computed(() => formModel.getProperties());
const getField = (id: string) => formProperties.value.find((f) => f.id === id);

const formRef = ref();
const formValues = ref({
  op_product_id: undefined as number | undefined,
  op_product_name: '',
  biz_names: '',
});
const opProductLoading = ref(false);
const bizLoading = ref(false);
const isShowNoOpRelation = ref(false);

// ==================== 公共：Excel 上传与预览 ====================
const uploadRef = useTemplateRef<typeof Upload>('uploadRef');
const submitting = ref(false);
const uploadedData = ref<IExcelImportData | null>(null);
const isUploadSuccess = ref(false);
const isPreviewShow = ref(false);
let uploadFiles: UploadFiles = [];

/** 提交按钮是否可用：文件已上传成功 且 预览数据中不存在校验错误行 */
const canSubmit = computed(() => {
  if (!isUploadSuccess.value || !uploadedData.value) return false;
  const { details } = uploadedData.value;
  if (!details || details.length === 0) return false;
  return details.every((d) => d.validate_result.length === 0);
});

// 上传地址
const uploadUrl = computed(() => {
  const bizId = getBizsId();
  return `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bizId}/plans/resources/gpu/excel/import`;
});

// 处理上传响应码
const handleResCode = (response: any) => response.code === 0;

// 上传成功
const handleUploadSuccess = (response: any, file: UploadFile, fileList: UploadFiles) => {
  uploadedData.value = response.data;
  isUploadSuccess.value = true;
  const target = fileList.find((item) => item.uid === file.uid);

  // 检查上传数据中是否存在校验错误
  const details = response.data?.details ?? [];
  const hasValidateError = details.some((d: any) => d.validate_result?.length > 0);

  if (hasValidateError) {
    if (target) {
      target.status = 'fail';
      target.statusText = '上传的数据中格式有错误，请到预览结果检查';
    }
    Message({ theme: 'warning', message: '上传的数据中格式有错误，请到预览结果检查' });
  } else {
    if (target) target.statusText = '上传成功';
    Message({ theme: 'success', message: '文件上传成功' });
  }
};

// 上传失败
const handleUploadError = () => {
  uploadedData.value = null;
  Message({ theme: 'error', message: '文件上传失败' });
};

// 文件删除
const handleFileDelete = () => {
  uploadedData.value = null;
  isUploadSuccess.value = false;
};

// 上传完成
const handleUploadDone = (fileList: UploadFiles) => {
  uploadFiles = [...fileList];
};

const handlePreviewFile = () => {
  isPreviewShow.value = true;
};

const handleDownloadTemplate = () => {
  // TODO: download template file
};

// ==================== 新增模式：运营产品 / 业务初始化 ====================
const fetchOpProductAndBizList = async () => {
  const bizId = getBizsId();
  if (!bizId) return;

  opProductLoading.value = true;
  try {
    const res = await resourcePlanStore.getBizOrgRelation(bizId);
    if (res.code === 0) {
      formValues.value.op_product_name = res.data?.op_product_name ?? '';
      formValues.value.op_product_id = res.data?.op_product_id;
      isShowNoOpRelation.value = false;
      opProductLoading.value = false;

      bizLoading.value = true;
      try {
        const bizList = await resourcePlanStore.getBizsByOpProductList({
          op_product_id: res.data.op_product_id,
        });
        formValues.value.biz_names = bizList.map((item) => item.bk_biz_name).join(', ');
      } catch {
        formValues.value.biz_names = '';
      } finally {
        bizLoading.value = false;
      }
    } else {
      isShowNoOpRelation.value = true;
      opProductLoading.value = false;
    }
  } catch {
    isShowNoOpRelation.value = true;
    opProductLoading.value = false;
  }
};

onMounted(() => {
  if (isCreateMode.value) {
    fetchOpProductAndBizList();
  }
});

// ==================== 重新导入模式：主单信息展示 ====================
const demandPeriod = computed(() => {
  if (!props.orderDetail) return '-';
  return timeFormatter(props.orderDetail.created_at, 'YYYY-MM-DD');
});

// ==================== 提交 ====================
const handleSubmit = async () => {
  // 新增模式：先校验表单
  if (isCreateMode.value) {
    try {
      await formRef.value?.validate();
    } catch {
      return;
    }
  }

  // 重新导入模式：校验主单信息
  if (!isCreateMode.value && !props.orderDetail?.id) {
    Message({ theme: 'error', message: '主单信息缺失，无法提交' });
    return;
  }

  if (!uploadedData.value) {
    Message({ theme: 'warning', message: '请先上传需求数据文件' });
    return;
  }

  // 检查是否存在校验错误行
  const { details } = uploadedData.value;
  if (!details || details.length === 0) {
    Message({ theme: 'warning', message: '上传文件内容为空，请检查后重新上传' });
    return;
  }
  const errorCount = details.filter((d) => d.validate_result.length > 0).length;
  if (errorCount > 0) {
    Message({ theme: 'warning', message: `存在 ${errorCount} 条数据校验不通过，请修正后重新上传` });
    return;
  }

  // 将 Excel 预览数据转换为提交接口所需的 details 格式
  const submitDetails = buildSubmitDetails(uploadedData.value);
  if (submitDetails.length === 0) {
    Message({ theme: 'warning', message: '无有效数据可提交' });
    return;
  }

  submitting.value = true;
  try {
    if (isCreateMode.value) {
      await gpuDemandStore.createGpuDemandOrder({
        op_product_id: formValues.value.op_product_id!,
        op_product_name: formValues.value.op_product_name,
        details: submitDetails,
      });
      Message({ theme: 'success', message: '提交成功' });
    } else {
      await gpuDemandStore.overwriteGpuDemandOrder({
        order_id: props.orderDetail!.id,
        details: submitDetails,
      });
      Message({ theme: 'success', message: '重新导入成功' });
    }
    emit('success');
    model.value = false;
  } catch (error: any) {
    const defaultMsg = isCreateMode.value ? '提交失败' : '重新导入失败';
    Message({ theme: 'error', message: error?.message || defaultMsg });
  } finally {
    submitting.value = false;
  }
};

const handleCancel = () => {
  model.value = false;
};

// ==================== 关闭重置 ====================
const resetUploadState = () => {
  uploadedData.value = null;
  isUploadSuccess.value = false;
  uploadFiles.forEach((file) => {
    uploadRef.value?.handleRemove(file);
  });
  uploadFiles = [];
};

const handleClosed = () => {
  if (isCreateMode.value) {
    formValues.value = { op_product_id: undefined, op_product_name: '', biz_names: '' };
    isShowNoOpRelation.value = false;
  }
  resetUploadState();
};
</script>

<template>
  <bk-sideslider
    v-model:is-show="model"
    :title="sliderTitle"
    width="640"
    @closed="handleClosed"
    @hidden="emit('hidden')"
  >
    <template #default>
      <div class="gpu-demand-slider-form">
        <!-- ==================== 新增模式：运营产品 + 关联业务表单 ==================== -->
        <bk-form v-if="isCreateMode" form-type="vertical" ref="formRef" :model="formValues">
          <bk-form-item :label="(getField('op_product_name')?.name as string)" property="op_product_name" required>
            <bk-input
              v-bkloading="{ loading: opProductLoading, size: 'mini' }"
              :model-value="formValues.op_product_name"
              disabled
            />
            <div v-if="isShowNoOpRelation" class="no-op-relation">
              <span class="warning-text">当前业务无运营产品，</span>
              请联系
              <w-name name="ICR" alias="ICR(IEG资源服务助手)" />
              确认
            </div>
          </bk-form-item>
          <bk-form-item :label="(getField('biz_names')?.name as string)" property="biz_names" required>
            <bk-input
              v-bkloading="{ loading: bizLoading, size: 'mini' }"
              :model-value="formValues.biz_names"
              disabled
            />
          </bk-form-item>
        </bk-form>

        <!-- ==================== 重新导入模式：主单信息卡片 + 警告提示 ==================== -->
        <template v-else>
          <div class="order-info-card">
            <div class="order-info-row">
              <div class="order-info-item">
                <span class="order-info-label label-left">需求单号</span>
                <span class="order-info-colon">：</span>
                <span class="order-info-value">{{ orderDetail?.id ?? '-' }}</span>
              </div>
              <div class="order-info-item">
                <span class="order-info-label label-right">需求时间</span>
                <span class="order-info-colon">：</span>
                <span class="order-info-value">{{ demandPeriod }}</span>
              </div>
            </div>
            <div class="order-info-row">
              <div class="order-info-item">
                <span class="order-info-label label-left">提单人</span>
                <span class="order-info-colon">：</span>
                <span class="order-info-value">{{ orderDetail?.creator ?? '-' }}</span>
              </div>
              <div class="order-info-item">
                <span class="order-info-label label-right">需求卡数</span>
                <span class="order-info-colon">：</span>
                <span class="order-info-value">{{ orderDetail?.total_gpu_num ?? '-' }}</span>
              </div>
            </div>
            <div class="order-info-row">
              <div class="order-info-item">
                <span class="order-info-label label-left">提单时间</span>
                <span class="order-info-colon">：</span>
                <span class="order-info-value">{{ timeFormatter(orderDetail?.created_at, 'YYYY-MM-DD') }}</span>
              </div>
              <div class="order-info-item">
                <span class="order-info-label label-right">QPM(月调用峰值)</span>
                <span class="order-info-colon">：</span>
                <span class="order-info-value">{{ orderDetail?.total_qpm_max ?? '-' }}</span>
              </div>
            </div>
          </div>
          <bk-alert class="mt24 mb24" theme="warning" title="上传后，将覆盖原始需求，以最新需求为准" />
        </template>

        <!-- ==================== 公共：Excel 需求数据上传 ==================== -->
        <div class="upload-section">
          <div class="upload-label">
            <span class="upload-label-text">需求数据</span>
            <bk-button v-if="isUploadSuccess" theme="primary" text class="preview-btn" @click="handlePreviewFile">
              <Eye class="preview-icon" />
              预览文件
            </bk-button>
          </div>
          <bk-upload
            ref="uploadRef"
            :class="{ 'hide-trigger': isUploadSuccess }"
            :multiple="false"
            :limit="1"
            accept=".xlsx"
            :validate-name="/\.xlsx$/i"
            with-credentials
            name="file"
            :url="uploadUrl"
            :handle-res-code="handleResCode"
            @success="handleUploadSuccess"
            @error="handleUploadError"
            @delete="handleFileDelete"
            @done="handleUploadDone"
          />
          <div class="upload-tips">
            仅支持 .xlsx 格式文件，下载
            <bk-button theme="primary" text class="ml4" @click="handleDownloadTemplate">模版文件</bk-button>
          </div>
        </div>
      </div>
    </template>
    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" :loading="submitting" :disabled="!canSubmit" @click="handleSubmit">提交</bk-button>
        <bk-button @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>
  <ExcelPreviewDialog v-model="isPreviewShow" :data="uploadedData" />
</template>

<style lang="scss">
/* stylelint-disable selector-class-pattern */

.gpu-demand-slider-form {
  .bk-upload-list__item {
    max-height: 80px;
  }

  .bk-upload.hide-trigger .bk-upload-trigger--draggable {
    display: none;
  }

  .bk-upload-list__item-name,
  .bk-upload-list__item-message {
    line-height: 20px;
    font-size: 14px;
  }
}

/* stylelint-enable selector-class-pattern */
</style>

<style lang="scss" scoped>
.gpu-demand-slider-form {
  padding: 28px 40px 0;

  :deep(.bk-form-label) {
    font-size: 12px;
  }
}

/* ====== 新增模式：表单样式 ====== */
.no-op-relation {
  margin-top: 4px;
  font-size: 12px;

  .warning-text {
    color: #ff9c01;
  }
}

/* ====== 重新导入模式：主单信息卡片 ====== */
.order-info-card {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 16px 20px;
}

.order-info-row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;

  &:last-child {
    margin-bottom: 0;
  }
}

.order-info-item {
  flex: 1;
  display: flex;
  align-items: center;
  font-size: 12px;
  line-height: 20px;
}

.order-info-label {
  display: inline-block;
  color: #4d4f56;
  white-space: nowrap;
  text-align: right;
  flex-shrink: 0;

  /* 左列 label 固定宽度（最长：需求单号 = 4字） */
  &.label-left {
    width: 48px;
  }

  /* 右列 label 固定宽度（最长：QPM(月调用峰值) = 10字符+括号） */
  &.label-right {
    width: 100px;
  }
}

.order-info-colon {
  color: #4d4f56;
  flex-shrink: 0;
}

.order-info-value {
  color: #313238;
}

/* ====== 公共：上传区域 ====== */
.upload-section {
  margin-top: 8px;
}

.upload-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  position: relative;
  margin-bottom: 8px;

  .upload-label-text {
    font-size: 12px;
    color: #4d4f56;

    &::after {
      display: inline-block;
      width: 14px;
      color: #ea3636;
      text-align: center;
      content: '*';
    }
  }
}

.preview-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;

  .preview-icon {
    font-size: 14px;
  }
}

.upload-tips {
  margin-top: 8px;
  font-size: 12px;
  color: #4d4f56;
}

.sideslider-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;

  .bk-button {
    min-width: 88px;
  }
}
</style>
