<script setup lang="ts">
import { ref, watch } from 'vue';
import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';
import hintIcon from '@/assets/image/hint.svg';

const show = defineModel<boolean>({ default: false });

defineProps<{
  data: IPermissionTemplateItem;
  loading?: boolean;
}>();

const emit = defineEmits<{
  confirm: [];
  ioaVerify: [];
}>();

const isConfirmed = ref(false);

watch(show, (val) => {
  if (!val) {
    isConfirmed.value = false;
  }
});

const handleConfirm = () => {
  emit('confirm');
};
</script>

<template>
  <bk-dialog v-model:is-show="show" dialog-type="show" header-align="center" :quick-close="false" width="480">
    <div class="delete-dialog-content">
      <div class="delete-header">
        <img :src="hintIcon" class="delete-icon" />
        <div class="delete-title">确认删除权限模板？</div>
      </div>

      <bk-alert theme="error">
        <template #title>
          <div>删除此权限模板后，对应用户的权限将自动取消。</div>
          <div>此操作不可恢复，请谨慎操作。</div>
        </template>
      </bk-alert>

      <div class="info-panel">
        <div class="info-row">
          <span class="info-label">权限模板名称：</span>
          <span class="info-value">{{ data?.name }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">所属二级账号：</span>
          <span class="info-value">{{ data?.account_id }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">关联三级账号数：</span>
          <span class="info-value">{{ data?.associated_sub_account_count }}</span>
        </div>
      </div>

      <bk-checkbox v-model="isConfirmed" class="confirm-checkbox">已知晓变更影响，仍需变更</bk-checkbox>
    </div>
    <div class="delete-dialog-footer">
      <bk-button theme="danger" :loading="loading" :disabled="!isConfirmed" @click="handleConfirm">删除</bk-button>
      <bk-button @click="show = false">取消</bk-button>
    </div>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.delete-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;

  .delete-header {
    display: flex;
    align-items: center;
    flex-direction: column;
    gap: 16px;
    margin-top: 8px;
  }

  .delete-icon {
    width: 42px;
    height: 42px;
  }

  .delete-title {
    font-size: 20px;
    color: #313238;
    line-height: 28px;
  }

  .info-panel {
    background: #f5f7fa;
    border-radius: 2px;
    padding: 8px 24px;

    .info-row {
      display: flex;
      align-items: center;
      height: 32px;
      font-size: 12px;

      .info-label {
        width: 98px;
        text-align: right;
        color: #4d4f56;
        flex-shrink: 0;
        margin-right: 8px;
      }

      .info-value {
        color: #313238;
      }
    }
  }

  .confirm-checkbox {
    font-size: 12px;
  }
}

.delete-dialog-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 16px;

  .bk-button {
    min-width: 88px;
  }
}
</style>
