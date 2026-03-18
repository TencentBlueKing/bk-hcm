<script setup lang="ts">
import { ref, computed } from 'vue';
import { Message } from 'bkui-vue';
import { useGpuDemandStore } from '@/store/resource-plan/gpu-demand';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';

export interface IReviewDetailItem {
  label: string;
  value: string | number;
}

const isShow = defineModel<boolean>({ default: false });

const props = defineProps<{
  row: Record<string, any> | null;
  detailItems: IReviewDetailItem[];
}>();

const emit = defineEmits<{
  (e: 'success'): void;
}>();

const gpuDemandStore = useGpuDemandStore();
const loading = ref(false);
const conclusion = ref<'DONE' | 'REJECT'>('DONE');
const comment = ref('');

const commentLabel = computed(() => (conclusion.value === 'DONE' ? '评审意见' : '驳回原因'));
const commentPlaceholder = computed(() =>
  conclusion.value === 'DONE' ? '请输入评审意见（选填）' : '请输入驳回原因（选填）',
);

const handleConfirm = async () => {
  if (!props.row) return;
  loading.value = true;
  try {
    const item: { suborder_id: string; status: string; comment?: string[] } = {
      suborder_id: props.row._id,
      status: conclusion.value,
    };
    if (comment.value.trim()) {
      item.comment = [comment.value.trim()];
    }
    await gpuDemandStore.batchUpdateSubOrders({ suborder_data: [item] });
    Message({ theme: 'success', message: conclusion.value === 'DONE' ? '评审通过' : '已驳回' });
    isShow.value = false;
    emit('success');
  } catch {
    Message({ theme: 'error', message: '评审提交失败' });
  } finally {
    loading.value = false;
  }
};

const handleClosed = () => {
  conclusion.value = 'DONE';
  comment.value = '';
};
</script>

<template>
  <bk-dialog
    v-model:is-show="isShow"
    title="评审"
    :width="680"
    :is-loading="loading"
    @confirm="handleConfirm"
    @closed="handleClosed"
  >
    <div class="review-dialog-content">
      <p class="review-desc">请确认以下数据信息，确认后将提交评审：</p>
      <grid-container :column="2" bordered :label-width="160" :content-min-width="80">
        <grid-item v-for="item in detailItems" :key="item.label" :label="item.label">
          {{ item.value }}
        </grid-item>
      </grid-container>
      <bk-form class="review-form" form-type="vertical">
        <bk-form-item label="评审结论" required>
          <bk-radio-group v-model="conclusion">
            <bk-radio label="DONE">通过</bk-radio>
            <bk-radio label="REJECT">驳回</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item :label="commentLabel">
          <bk-input v-model="comment" type="textarea" :placeholder="commentPlaceholder" :maxlength="100" :rows="3" />
        </bk-form-item>
      </bk-form>
    </div>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.review-dialog-content {
  .review-desc {
    font-size: 14px;
    color: #63656e;
    margin-bottom: 16px;
  }

  .review-form {
    margin-top: 16px;
  }
}
</style>
