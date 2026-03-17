<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { Message } from 'bkui-vue';
import { useCloudAccountStore, type IAccountSecretItem } from '@/store/cloud-account';
import { useWhereAmI } from '@/hooks/useWhereAmI';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

// Props 定义
const props = defineProps<{
  isEdit: boolean;
  secretData: IAccountSecretItem | null;
  accountId: string;
}>();

// Emits 定义
const emit = defineEmits<{
  success: [];
}>();

// Store 和 Hooks
const cloudAccountStore = useCloudAccountStore();
const { getBizsId } = useWhereAmI();

// 表单引用
const formRef = ref();

// 表单数据
const formData = ref({
  type: '',
  cloud_secret_id: '',
  cloud_secret_key: '',
});

// 密钥校验结果
const checkResult = ref<{
  isChecked: boolean;
  isSuccess: boolean;
  cloudMainAccountId: string;
  cloudSubAccountId: string;
}>({
  isChecked: false,
  isSuccess: false,
  cloudMainAccountId: '',
  cloudSubAccountId: '',
});

// 校验中状态
const isChecking = ref(false);

// 提交中状态
const isSubmitting = ref(false);

// 密钥类型选项
const secretTypeOptions = [
  { id: 'resource', name: '资源管理' },
  { id: 'security', name: '安全管理' },
];

// 表单校验规则
const formRules = {
  type: [{ required: true, message: '请选择密钥类型', trigger: 'change' }],
  cloud_secret_id: [
    { required: true, message: '请输入云密钥ID', trigger: 'blur' },
    { pattern: /^[A-Za-z0-9]+$/, message: '云密钥ID格式不正确', trigger: 'blur' },
  ],
  cloud_secret_key: [{ required: true, message: '请输入云密钥Key', trigger: 'blur' }],
};

// 弹窗标题
const dialogTitle = computed(() => (props.isEdit ? '编辑资源密钥' : '录入资源密钥'));

// 提交按钮是否可用
const canSubmit = computed(() => {
  return checkResult.value.isChecked && checkResult.value.isSuccess;
});

// 监听弹窗显示状态，初始化数据
watch(
  () => model.value,
  (newVal) => {
    if (newVal) {
      // 重置校验结果
      checkResult.value = {
        isChecked: false,
        isSuccess: false,
        cloudMainAccountId: '',
        cloudSubAccountId: '',
      };

      if (props.isEdit && props.secretData) {
        // 编辑模式：填充现有数据
        formData.value = {
          type: props.secretData.type,
          cloud_secret_id: props.secretData.extension?.cloud_secret_id || '',
          cloud_secret_key: '', // 密钥key不回显
        };
      } else {
        // 新增模式：清空表单
        formData.value = {
          type: '',
          cloud_secret_id: '',
          cloud_secret_key: '',
        };
      }

      // 清除表单校验状态
      nextTick(() => {
        formRef.value?.clearValidate();
      });
    }
  },
);

// 监听表单数据变化，重置校验结果
watch(
  () => [formData.value.type, formData.value.cloud_secret_id, formData.value.cloud_secret_key],
  () => {
    // 数据变化时重置校验结果
    checkResult.value = {
      isChecked: false,
      isSuccess: false,
      cloudMainAccountId: '',
      cloudSubAccountId: '',
    };
  },
);

// 密钥校验
const handleCheckSecret = async () => {
  // 先进行表单校验
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  isChecking.value = true;

  try {
    const result = await cloudAccountStore.checkAccountSecret(getBizsId(), {
      account_id: props.accountId,
      type: formData.value.type,
      extension: {
        cloud_secret_id: formData.value.cloud_secret_id,
        cloud_secret_key: formData.value.cloud_secret_key,
      },
    });

    checkResult.value = {
      isChecked: true,
      isSuccess: true,
      cloudMainAccountId: result.cloud_main_account_id,
      cloudSubAccountId: result.cloud_sub_account_id,
    };

    Message({ theme: 'success', message: '密钥校验成功' });
  } catch (error) {
    checkResult.value = {
      isChecked: true,
      isSuccess: false,
      cloudMainAccountId: '',
      cloudSubAccountId: '',
    };
    Message({ theme: 'error', message: '密钥校验失败，请检查密钥信息' });
  }

  isChecking.value = false;
};

// 提交表单
const handleSubmit = async () => {
  if (!canSubmit.value) {
    Message({ theme: 'warning', message: '请先进行密钥校验' });
    return;
  }

  isSubmitting.value = true;

  try {
    if (props.isEdit && props.secretData) {
      // 编辑密钥
      await cloudAccountStore.updateAccountSecret(getBizsId(), props.secretData.id, {
        type: formData.value.type,
        extension: {
          cloud_secret_id: formData.value.cloud_secret_id,
          cloud_secret_key: formData.value.cloud_secret_key,
        },
      });
      Message({ theme: 'success', message: '编辑成功' });
    } else {
      // 创建密钥
      await cloudAccountStore.createAccountSecret(getBizsId(), {
        account_id: props.accountId,
        type: formData.value.type,
        extension: {
          cloud_secret_id: formData.value.cloud_secret_id,
          cloud_secret_key: formData.value.cloud_secret_key,
        },
      });
      Message({ theme: 'success', message: '录入成功' });
    }
    emit('success');
  } catch (error) {
    console.error('操作失败:', error);
    Message({ theme: 'error', message: props.isEdit ? '编辑失败' : '录入失败' });
  }

  isSubmitting.value = false;
};

// 取消
const handleCancel = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" :title="dialogTitle" :width="480" :draggable="false" dialog-type="show">
    <bk-form ref="formRef" :model="formData" :rules="formRules" form-type="vertical">
      <!-- 密钥类型 -->
      <bk-form-item label="密钥类型" property="type" required>
        <bk-select v-model="formData.type" placeholder="请选择" :clearable="false">
          <bk-option v-for="item in secretTypeOptions" :key="item.id" :id="item.id" :name="item.name" />
        </bk-select>
      </bk-form-item>

      <!-- 云密钥ID -->
      <bk-form-item label="云密钥ID（SecretID）" property="cloud_secret_id" required>
        <bk-input v-model="formData.cloud_secret_id" placeholder="请输入" />
      </bk-form-item>

      <!-- 云密钥Key -->
      <bk-form-item label="云密钥（SecretKey）" property="cloud_secret_key" required>
        <bk-input v-model="formData.cloud_secret_key" type="password" placeholder="请输入" />
      </bk-form-item>

      <!-- 密钥校验按钮 -->
      <bk-form-item>
        <bk-button
          theme="primary"
          :loading="isChecking"
          :outline="checkResult.isChecked && checkResult.isSuccess"
          @click="handleCheckSecret"
        >
          密钥校验
        </bk-button>
        <!-- 校验结果状态 -->
        <span v-if="checkResult.isChecked" class="check-result">
          <template v-if="checkResult.isSuccess">
            <i class="bk-icon icon-check-circle-fill success-icon"></i>
            <span class="success-text">校验成功</span>
          </template>
          <template v-else>
            <i class="bk-icon icon-close-circle-fill error-icon"></i>
            <span class="error-text">校验失败</span>
          </template>
        </span>
      </bk-form-item>

      <!-- 校验成功后显示的信息 -->
      <div v-if="checkResult.isChecked" class="check-info-box">
        <div class="check-info-item">
          <span class="check-label">所属三级账号ID：</span>
          <span class="check-value">
            {{ checkResult.isSuccess ? checkResult.cloudSubAccountId : '密钥校验成功后自动填充' }}
          </span>
        </div>
        <div class="check-info-item">
          <span class="check-label">所属二级账号ID：</span>
          <span class="check-value">
            {{ checkResult.isSuccess ? checkResult.cloudMainAccountId : '密钥校验成功后自动填充' }}
          </span>
        </div>
      </div>
      <div v-else class="check-info-box placeholder">
        <div class="check-info-item">
          <span class="check-label">所属三级账号ID：</span>
          <span class="check-value placeholder-text">密钥校验成功后自动填充</span>
        </div>
        <div class="check-info-item">
          <span class="check-label">所属二级账号ID：</span>
          <span class="check-value placeholder-text">密钥校验成功后自动填充</span>
        </div>
      </div>
    </bk-form>

    <template #footer>
      <div class="dialog-footer">
        <bk-button theme="primary" :disabled="!canSubmit" :loading="isSubmitting" @click="handleSubmit">提交</bk-button>
        <bk-button @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.check-result {
  display: inline-flex;
  align-items: center;
  margin-left: 12px;

  .success-icon {
    color: #2dcb56;
    font-size: 16px;
    margin-right: 4px;
  }

  .error-icon {
    color: #ea3636;
    font-size: 16px;
    margin-right: 4px;
  }

  .success-text {
    color: #2dcb56;
    font-size: 12px;
  }

  .error-text {
    color: #ea3636;
    font-size: 12px;
  }
}

.check-info-box {
  background: #f5f7fa;
  border-radius: 2px;
  padding: 12px 16px;
  margin-top: 8px;

  &.placeholder {
    .check-value {
      color: #979ba5;
    }
  }

  .check-info-item {
    display: flex;
    align-items: center;
    font-size: 12px;
    line-height: 24px;

    .check-label {
      color: #3a84ff;
      min-width: 100px;
    }

    .check-value {
      color: #63656e;

      &.placeholder-text {
        color: #979ba5;
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
