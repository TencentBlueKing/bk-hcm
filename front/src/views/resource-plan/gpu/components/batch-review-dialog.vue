<script setup lang="ts">
import { ref } from 'vue';
import { Message } from 'bkui-vue';
import { useGpuDemandStore } from '@/store/resource-plan/gpu-demand';

const isShow = defineModel<boolean>({ default: false });

const props = defineProps<{
  suborderIds: string[];
}>();

const emit = defineEmits<{
  (e: 'success'): void;
}>();

const gpuDemandStore = useGpuDemandStore();
const loading = ref(false);
const comment = ref('');

const handleConfirm = async () => {
  loading.value = true;
  try {
    await gpuDemandStore.batchUpdateSubOrders({
      suborder_data: props.suborderIds.map((id) => ({
        suborder_id: id,
        status: 'DONE',
        ...(comment.value.trim() ? { comment: [comment.value.trim()] } : {}),
      })),
    });
    Message({ theme: 'success', message: '批量评审成功' });
    isShow.value = false;
    emit('success');
  } catch {
    Message({ theme: 'error', message: '批量评审失败' });
  } finally {
    loading.value = false;
  }
};

const handleClosed = () => {
  comment.value = '';
};
</script>

<template>
  <bk-dialog
    v-model:is-show="isShow"
    title="批量评审确认"
    :is-loading="loading"
    confirm-text="确认评审"
    @confirm="handleConfirm"
    @closed="handleClosed"
  >
    <div class="batch-review-content">
      <p class="batch-review-desc">
        本次将评审
        <span class="highlight">{{ suborderIds.length }}</span>
        条数据，确认提交评审？
      </p>
      <div class="batch-review-divider" />
      <bk-form form-type="vertical">
        <bk-form-item label="评审意见">
          <bk-input v-model="comment" type="textarea" placeholder="请输入评审意见（选填）" :maxlength="100" :rows="4" />
        </bk-form-item>
      </bk-form>
    </div>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.batch-review-content {
  .batch-review-desc {
    font-size: 14px;
    line-height: 22px;
    color: #63656e;

    .highlight {
      color: #3a84ff;
      font-weight: 700;
    }
  }

  .batch-review-divider {
    height: 1px;
    background: #dcdee5;
    margin: 16px 0;
  }
}
</style>
