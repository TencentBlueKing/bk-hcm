<script setup lang="ts">
import { ref, inject, type Ref, computed } from 'vue';
import { Message } from 'bkui-vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useTertiaryAccountStore, type ISubAccountItem } from '@/store/cloud-account-manage/tertiary-account';
import { VendorEnum } from '@/common/constant';
import { ACCOUNT_TYPE_OPTIONS } from '../../constants';

const model = defineModel<boolean>();

const props = defineProps<{
  accountData: ISubAccountItem | null;
}>();

const emit = defineEmits<{
  (e: 'success'): void;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const tertiaryAccountStore = useTertiaryAccountStore();
const { getBizsId } = useWhereAmI();

const confirmed = ref(false);
const isSubmitting = ref(false);

const isConfirmDisabled = computed(() => !confirmed.value);

const handleClose = () => {
  model.value = false;
  confirmed.value = false;
};

const handleConfirm = async () => {
  if (!props.accountData?.id) return;

  isSubmitting.value = true;
  try {
    await tertiaryAccountStore.deleteSubAccount(getBizsId(), currentVendor.value, [props.accountData.id]);
    Message({ theme: 'success', message: '删除申请提交成功' });
    handleClose();
    emit('success');
  } catch (error) {
    console.error('删除三级账号失败:', error);
  } finally {
    isSubmitting.value = false;
  }
};

const formatTime = (time?: string) => {
  if (!time) return '--';
  return time.replace('T', ' ').replace('Z', '');
};

const getAccountTypeText = (consoleLogin?: number) => {
  if (consoleLogin === undefined || consoleLogin === null) return '--';
  return ACCOUNT_TYPE_OPTIONS[consoleLogin] || '--';
};
</script>

<template>
  <bk-dialog
    :is-show="model"
    title="删除三级账号"
    :close-icon="true"
    width="500"
    @closed="handleClose"
    @confirm="handleConfirm"
  >
    <template #default>
      <div v-if="accountData" class="delete-dialog-content">
        <bk-alert theme="warning" class="warning-alert">
          <template #title>
            <p>删除三级账号后，账号不可恢复，请确认该账号不再使用。</p>
            <p>点击确认后，将提交申请单进行删除。</p>
          </template>
        </bk-alert>

        <div class="account-info">
          <div class="info-item">
            <span class="info-label">三级账号ID：</span>
            <span class="info-value">{{ accountData.cloud_id || '--' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">三级账号名称：</span>
            <span class="info-value">{{ accountData.name || '--' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">账号类型：</span>
            <span class="info-value">{{ getAccountTypeText(accountData.extension?.console_login) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">创建时间：</span>
            <span class="info-value">{{ formatTime(accountData.cloud_created_at) }}</span>
          </div>
        </div>

        <div class="confirm-check">
          <bk-checkbox v-model="confirmed">我已确认删除此账号</bk-checkbox>
        </div>
      </div>
    </template>
    <template #footer>
      <bk-button theme="primary" :disabled="isConfirmDisabled" :loading="isSubmitting" @click="handleConfirm">
        确认
      </bk-button>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.delete-dialog-content {
  .warning-alert {
    margin-bottom: 16px;
  }

  .account-info {
    padding: 16px;
    background: #f5f7fa;
    border-radius: 2px;
    margin-bottom: 16px;

    .info-item {
      display: flex;
      align-items: center;
      font-size: 14px;
      line-height: 32px;

      .info-label {
        color: #63656e;
        white-space: nowrap;
        min-width: 98px;
        flex-shrink: 0;
        text-align: right;
        margin-right: 8px;
      }

      .info-value {
        color: #313238;
      }
    }
  }

  .confirm-check {
    padding-top: 8px;
  }
}
</style>
