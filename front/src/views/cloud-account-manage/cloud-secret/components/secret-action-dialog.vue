<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { Message } from 'bkui-vue';
import { useCloudAccountStore } from '@/store/cloud-account';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { SECRET_ACTION_CONFIG } from '../constants';
import type { ICloudSecretItem, SecretActionType } from '../typings';

const props = defineProps<{
  modelValue: boolean;
  actionType: SecretActionType;
  secretData: ICloudSecretItem | null;
  vendor: string;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  success: [];
}>();

const cloudAccountStore = useCloudAccountStore();
const { getBizsId } = useWhereAmI();

const isShow = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

const isAcknowledged = ref(false);
const isSubmitting = ref(false);
const actionConfig = computed(() => SECRET_ACTION_CONFIG[props.actionType]);
const isConfirmDisabled = computed(() => !isAcknowledged.value || isSubmitting.value);

const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return '--';
  return dateStr.replace('T', ' ').replace('Z', '');
};

watch(isShow, (val) => {
  if (!val) {
    isAcknowledged.value = false;
    isSubmitting.value = false;
  }
});

const handleCancel = () => {
  isShow.value = false;
};

const handleConfirm = async () => {
  if (!props.secretData || !isAcknowledged.value) return;

  isSubmitting.value = true;

  try {
    const bkBizId = getBizsId();

    if (props.actionType === 'delete') {
      await cloudAccountStore.deleteSubAccountSecret(bkBizId, props.vendor, [props.secretData.id]);
      Message({ theme: 'success', message: '删除密钥申请已提交' });
    } else {
      const newStatus = props.actionType === 'enable' ? 'enabled' : 'disabled';
      await cloudAccountStore.updateSubAccountSecretStatus(bkBizId, props.vendor, [
        {
          id: props.secretData.id,
          status: newStatus,
        },
      ]);
      Message({ theme: 'success', message: `${props.actionType === 'enable' ? '启用' : '禁用'}密钥申请已提交` });
    }

    isShow.value = false;
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
  <bk-dialog
    v-model:is-show="isShow"
    :title="actionConfig?.title"
    :width="480"
    header-align="center"
    footer-align="center"
    :quick-close="false"
  >
    <div class="secret-action-dialog">
      <bk-alert :theme="actionConfig?.alertType" :title="actionConfig?.alertMessage" class="alert-box">
        <template v-if="actionConfig?.alertDescription" #description>
          {{ actionConfig.alertDescription }}
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

      <div class="acknowledge-box">
        <bk-checkbox v-model="isAcknowledged">已知晓变更影响，仍需变更</bk-checkbox>
      </div>
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
.secret-action-dialog {
  .alert-box {
    margin-bottom: 16px;
  }

  .secret-info {
    background: #f5f7fa;
    border-radius: 2px;
    padding: 16px 24px;
    margin-bottom: 16px;

    .info-item {
      display: flex;
      align-items: center;
      font-size: 12px;
      line-height: 28px;

      .label {
        color: #979ba5;
        min-width: 100px;
        text-align: right;
      }

      .value {
        color: #313238;
        margin-left: 8px;
      }
    }
  }

  .acknowledge-box {
    padding: 8px 0;
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
</style>
