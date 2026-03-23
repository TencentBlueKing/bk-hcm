<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { Message } from 'bkui-vue';
import BusinessSelector from '@/components/business-selector/business.vue';
import type { IPermissionPolicyItem } from '../../typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { isJSON } from '@/utils';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

// Props 定义
const props = defineProps<{
  isEdit?: boolean;
  accountData?: IPermissionPolicyItem | null;
}>();

// Emits 定义
const emit = defineEmits<{
  success: [updatedData?: IPermissionPolicyItem];
}>();

const { getBizsId } = useWhereAmI();

// 表单引用
const formRef = ref();

// 提交加载状态
const submitLoading = ref(false);

// 是否是粘贴得数据
const isPaste = ref(false);

// 表单数据
const formData = ref<IPermissionPolicyItem>({
  id: '',
  name: '',
  description: '',
  bk_biz_id: getBizsId() as number | undefined,
  usage_biz_ids: getBizsId() ? [getBizsId()] : ([] as number[]),
  json: '',
});

// 侧栏标题
const sidesliderTitle = computed(() => (props.isEdit ? '编辑权限策略库' : '新建权限策略库'));

// 表单校验规则
const formRules = {
  name: [{ required: true, message: '请输入权限策略库名称', trigger: 'blur' }],
  description: [{ required: true, message: '请输入权限策略库描述', trigger: 'blur' }],
  usage_biz_ids: [{ required: true, message: '请选择使用业务', trigger: 'blur' }],
  json: [
    { required: true, message: '请输入权限策略', trigger: 'blur' },
    { validator: isJSON, trigger: 'change', message: '请输入正确得JSON' },
  ],
};

// 重置表单
const resetForm = () => {
  formData.value = {
    id: '',
    name: '',
    description: '',
    bk_biz_id: getBizsId(),
    usage_biz_ids: getBizsId() ? [getBizsId()] : [],
    json: '',
  };
  nextTick(() => {
    formRef.value?.clearValidate();
  });
};

// 填充编辑数据
const fillEditData = () => {
  if (props.isEdit && props.accountData) {
    const data = props.accountData;
    formData.value = {
      id: data.id || '',
      name: data.name || '',
      description: data.description || '',
      bk_biz_id: data.bk_biz_id,
      usage_biz_ids: data.usage_biz_ids || [],
      json: data.json || '',
    };
  }
};

// 监听显示状态变化
watch(
  () => model.value,
  (newVal) => {
    if (newVal && props.isEdit) {
      fillEditData();
    } else {
      resetForm();
    }
  },
);

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  submitLoading.value = true;

  try {
    // 真实接口调用
    if (props.isEdit) {
      // 编辑接口
      // TODO: 替换为真实API调用
      // const list = await permissionPolicyStore.updatePermissionPolicy(getBizsId(), vendorFilter);
      Message({ theme: 'success', message: '编辑成功' });
      // 编辑模式时，返回更新后的数据
      const updatedData: IPermissionPolicyItem = {
        ...props.accountData!,
        name: formData.value.name,
        bk_biz_id: formData.value.bk_biz_id || 0,
        updated_at: new Date().toISOString(),
      };
      model.value = false;
      emit('success', updatedData);
    } else {
      // 录入接口
      // TODO: 替换为真实API调用
      // const list = await permissionPolicyStore.createPermissionPolicy(getBizsId(), vendorFilter);
      Message({ theme: 'success', message: '新建成功' });
      model.value = false;
      emit('success');
    }
  } catch (error) {
    console.error('提交失败:', error);
    Message({ theme: 'error', message: '提交失败，请重试' });
  } finally {
    submitLoading.value = false;
  }
};

const handlePaste = () => {
  isPaste.value = true;
};
const handleInput = (val: string) => {
  if (isPaste.value) {
    if (isJSON(val)) {
      formData.value.json = JSON.stringify(JSON.parse(val), null, 2);
    }
    isPaste.value = false;
    return;
  }
};

// 取消
const handleCancel = () => {
  model.value = false;
};
</script>

<template>
  <bk-sideslider v-model:is-show="model" :title="sidesliderTitle" :width="640" quick-close>
    <template #default>
      <div class="policy-form-container">
        <bk-form ref="formRef" :model="formData" :rules="formRules" form-type="vertical">
          <!-- 权限策略库名称 -->
          <bk-form-item label="权限策略库名称" property="name" required>
            <bk-input v-model="formData.name" placeholder="请输入权限策略库名称" :maxlength="128" show-word-limit />
          </bk-form-item>

          <!-- 允许使用业务 -->
          <bk-form-item label="允许使用业务" property="usage_biz_ids">
            <BusinessSelector
              v-model="formData.usage_biz_ids"
              placeholder="请选择允许使用的业务，支持全部，多选"
              multiple
              clearable
              collapse-tags
            />
          </bk-form-item>

          <!-- 权限策略库描述 -->
          <bk-form-item label="权限策略库描述" property="description" required>
            <bk-input v-model="formData.name" placeholder="请输入权限策略库描述" />
          </bk-form-item>

          <!-- 权限策略 -->
          <bk-form-item label="权限策略" property="json">
            <bk-input
              v-model="formData.json"
              type="textarea"
              placeholder="请输入权限策略"
              style="height: 350px; overflow-y: auto"
              @paste="handlePaste"
              @input="handleInput"
            />
          </bk-form-item>
        </bk-form>
      </div>
    </template>

    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" :loading="submitLoading" @click="handleSubmit">提交</bk-button>
        <bk-button :disabled="submitLoading" @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.policy-form-container {
  padding: 24px 40px 0;

  :deep(.bk-form-item) {
    margin-bottom: 24px;
  }

  :deep(.bk-radio-button-input:checked + .bk-radio-button-label) {
    background-color: #e1ecff;
    border-color: #3a84ff;
    color: #3a84ff;
  }
}

.sideslider-footer {
  display: flex;
  align-items: center;
  padding: 0 18px;

  .bk-button {
    min-width: 88px;
    margin-right: 8px;
  }
}
</style>
