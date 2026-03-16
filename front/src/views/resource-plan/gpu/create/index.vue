<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getModel } from '@/model/manager';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useResourcePlanStore } from '@/store/resource-plan';
import WName from '@/components/w-name';
import { CreateForm } from './form-fields';

const model = defineModel<boolean>({ default: false });
const emit = defineEmits<{
  hidden: [];
  success: [];
}>();

const { getBizsId } = useWhereAmI();
const resourcePlanStore = useResourcePlanStore();

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
const submitting = ref(false);
const isShowNoOpRelation = ref(false);

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

onMounted(fetchOpProductAndBizList);

const handleSubmit = async () => {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }
  submitting.value = true;
  try {
    // TODO: call create API
    emit('success');
    model.value = false;
  } finally {
    submitting.value = false;
  }
};

const handleCancel = () => {
  model.value = false;
};

const handleClosed = () => {
  formValues.value = { op_product_id: undefined, op_product_name: '', biz_names: '' };
  isShowNoOpRelation.value = false;
};

const handleDownloadTemplate = () => {
  // TODO: download template file
};
</script>

<template>
  <bk-sideslider
    v-model:is-show="model"
    title="新增GPU需求"
    width="640"
    @closed="handleClosed"
    @hidden="emit('hidden')"
  >
    <template #default>
      <div class="create-form">
        <bk-form form-type="vertical" ref="formRef" :model="formValues">
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
          <bk-form-item label="需求数据" required>
            <bk-upload :multiple="false" :limit="1" accept=".xlsx" :validate-name="/\.xlsx$/i" with-credentials />
            <div class="upload-tips">
              仅支持 .xlsx 格式文件，下载
              <bk-button theme="primary" text class="ml4" @click="handleDownloadTemplate">模版文件</bk-button>
            </div>
          </bk-form-item>
        </bk-form>
      </div>
    </template>
    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" :loading="submitting" @click="handleSubmit">提交</bk-button>
        <bk-button @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.create-form {
  padding: 28px 40px 0;
}

.no-op-relation {
  margin-top: 4px;
  font-size: 12px;

  .warning-text {
    color: #ff9c01;
  }
}

.upload-tips {
  margin-top: 8px;
  font-size: 12px;
  color: #979ba5;
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
