<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Message } from 'bkui-vue';
import { useCloudSecretStore } from '@/store/cloud-account-manage/cloud-secret';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { SECRET_ACTION_CONFIG } from '../../constants';
import type { ICloudSecretItem, SecretActionType } from '../../typings';

const model = defineModel<boolean>();

const props = defineProps<{
  actionType: SecretActionType;
  secretData: ICloudSecretItem | null;
  vendor: string;
}>();

const emit = defineEmits<{
  success: [];
}>();

const cloudSecretStore = useCloudSecretStore();
const { getBizsId } = useWhereAmI();

const isAcknowledged = ref(false);
const isSubmitting = ref(false);
const actionConfig = computed(() => SECRET_ACTION_CONFIG[props.actionType]);
const isConfirmDisabled = computed(() => !isAcknowledged.value || isSubmitting.value);

const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return '--';
  return dateStr.replace('T', ' ').replace('Z', '');
};

watch(model, (val) => {
  if (!val) {
    isAcknowledged.value = false;
    isSubmitting.value = false;
  }
});

const handleCancel = () => {
  model.value = false;
};

const handleConfirm = async () => {
  if (!props.secretData || !isAcknowledged.value) return;

  isSubmitting.value = true;

  try {
    const bkBizId = getBizsId();

    if (props.actionType === 'delete') {
      await cloudSecretStore.deleteSubAccountSecret(bkBizId, props.vendor, [props.secretData.id]);
      Message({ theme: 'success', message: '删除密钥申请已提交' });
    } else {
      const newStatus = props.actionType === 'enable' ? 'enabled' : 'disabled';
      await cloudSecretStore.updateSubAccountSecretStatus(bkBizId, props.vendor, [
        {
          id: props.secretData.id,
          status: newStatus,
        },
      ]);
      Message({ theme: 'success', message: `${props.actionType === 'enable' ? '启用' : '禁用'}密钥申请已提交` });
    }

    model.value = false;
    emit('success');
  } catch (error) {
    console.error('操作失败:', error);
    Message({ theme: 'error', message: '操作失败，请稍后重试' });
  } finally {
    isSubmitting.value = false;
  }
};
</script>

<template>
  <bk-dialog v-model:is-show="model" :width="480" header-align="center" footer-align="center" :quick-close="false">
    <template #header>
      <div class="dialog-header">
        <svg class="icon svg-icon">
          <use xlink:href="#bkhcm-icon-tishi"></use>
        </svg>
        <span>{{ actionConfig?.title }}</span>
      </div>
    </template>

    <div class="secret-action-dialog">
      <bk-alert :theme="actionConfig?.alertType" class="alert-box">
        <template #title>
          <span>
            {{ actionConfig?.alertMessage }}
          </span>
        </template>
      </bk-alert>

      <div class="secret-info">
        <div class="info-item">
          <span class="label">密钥ID：</span>
          <span class="value">{{ secretData?.cloud_secret_id || secretData?.extension?.cloud_secret_id || '--' }}</span>
        </div>
        <div class="info-item">
          <span class="label">创建时间：</span>
          <span class="value">{{ formatDateTime(secretData?.cloud_created_at) }}</span>
        </div>
        <div class="info-item">
          <span class="label">最近访问时间：</span>
          <span class="value">{{ formatDateTime(secretData?.last_used_time) }}</span>
        </div>
      </div>

      <bk-checkbox v-model="isAcknowledged">
        <span class="acknowledge-box-label">已知晓变更影响，仍需变更</span>
      </bk-checkbox>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <bk-button
          :theme="actionConfig?.confirmTheme"
          :disabled="isConfirmDisabled"
          :loading="isSubmitting"
          @click="handleConfirm"
        >
          {{ actionConfig?.confirmText }}
        </bk-button>
        <bk-button @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
:deep(.bk-dialog-content) {
  padding: 0 32px;
}

:deep(.bk-dialog-header) {
  padding: 24px 32px 0;
}

:deep(.bk-alert-icon-info) {
  margin-top: 3px;
}

.secret-action-dialog {
  .alert-box {
    margin-bottom: 16px;

    span {
      white-space: pre-wrap;
      word-break: break-all;
      line-height: 20px;
    }
  }

  .secret-info {
    background: #f5f7fa;
    border-radius: 2px;
    padding: 14px 24px;
    margin-bottom: 16px;

    .info-item {
      display: flex;
      align-items: center;
      line-height: 20px;
      font-size: 12px;
      margin-bottom: 12px;
      font-weight: 400;

      &:last-child {
        margin-bottom: 0;
      }

      .label {
        min-width: 84px;
        text-align: right;
        color: #313238;
      }

      .value {
        margin-left: 8px;
        word-break: break-all;
      }
    }
  }

  .acknowledge-box-label {
    font-size: 12px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: center;
  gap: 8px;

  .bk-button {
    min-width: 88px;
  }
}

.dialog-header {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: #4d4f56;

  .svg-icon {
    width: 42px;
    height: 42px;
    margin-bottom: 11px;
  }
}
</style>
