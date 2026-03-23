<script setup lang="ts">
import { ref } from 'vue';
import { Message } from 'bkui-vue';
import { useGpuDemandStore } from '@/store/resource-plan/gpu-demand';

const isShow = defineModel<boolean>({ default: false });

const props = defineProps<{
  suborderIds: string[];
}>();

const emit = defineEmits<{
  (e: 'success', subOrderIds: string[]): void;
}>();

const gpuDemandStore = useGpuDemandStore();
const loading = ref(false);
const reason = ref('');

const handleConfirm = async () => {
  loading.value = true;
  try {
    await gpuDemandStore.batchUpdateSubOrderStatus({
      suborder_ids: props.suborderIds,
      status: 'REJECT',
      ...(reason.value.trim() ? { comment: [reason.value.trim()] } : {}),
    });
    Message({ theme: 'success', message: '批量驳回成功' });
    isShow.value = false;
    emit('success', [...props.suborderIds]);
  } catch {
    Message({ theme: 'error', message: '批量驳回失败' });
  } finally {
    loading.value = false;
  }
};

const handleClosed = () => {
  reason.value = '';
};
</script>

<template>
  <bk-dialog
    v-model:is-show="isShow"
    title="驳回确认"
    theme="danger"
    :is-loading="loading"
    confirm-text="确认驳回"
    @confirm="handleConfirm"
    @closed="handleClosed"
  >
    <div class="batch-reject-content">
      <p class="batch-reject-desc">
        本次将驳回
        <span class="highlight">{{ suborderIds.length }}</span>
        条数据，确认驳回？
      </p>
      <div class="batch-reject-divider" />
      <bk-form form-type="vertical">
        <bk-form-item label="驳回原因">
          <bk-input v-model="reason" type="textarea" placeholder="请输入驳回原因（选填）" :maxlength="100" :rows="4" />
        </bk-form-item>
      </bk-form>
    </div>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.batch-reject-content {
  .batch-reject-desc {
    font-size: 14px;
    line-height: 22px;
    color: #63656e;

    .highlight {
      color: #ea3636;
      font-weight: 700;
    }
  }

  .batch-reject-divider {
    height: 1px;
    background: #dcdee5;
    margin: 16px 0;
  }
}
</style>
