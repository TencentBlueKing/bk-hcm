<script setup lang="ts">
import { ref, useTemplateRef, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { CvmsAssignPreviewItem, useHostStore } from '@/store';

import { Message } from 'bkui-vue';
import ManualAssign from './manual-assign.vue';
import MatchHost from './match-host.vue';

const props = defineProps<{ cvm: CvmsAssignPreviewItem; reloadTable: () => void }>();
const emit = defineEmits<(e: 'hidden') => void>();

const { t } = useI18n();
const hostStore = useHostStore();

const currentCvm = ref<CvmsAssignPreviewItem>(null);

const isManualAssignShow = ref(false);

const isMatchHostShow = ref(false);
const matchHostDialogRef = useTemplateRef('match-host-dialog');
const handleManualAssign = () => {
  // 手动关联且手动分配，关闭弹框并清空form
  matchHostDialogRef.value.handleClosed();
  isManualAssignShow.value = true;
};

const handleSubmit = async (cvm: { cvm_id: string; bk_biz_id: number; bk_cloud_id: number }) => {
  await hostStore.assignCvmsToBiz([cvm]);
  Message({ theme: 'success', message: t('分配成功') });
  emit('hidden');
  props.reloadTable();
};

watchEffect(async () => {
  // loading
  isManualAssignShow.value = true;

  const [cvm] = await hostStore.getAssignPreviewList([props.cvm]);

  const { match_type } = cvm;
  // 关联配置平台主机
  if (match_type === 'manual') {
    isManualAssignShow.value = false;
    isMatchHostShow.value = true;
  } else {
    // 分配主机，预览
    isManualAssignShow.value = true;
  }
  currentCvm.value = cvm;
});
</script>

<template>
  <!-- 分配主机 -->
  <manual-assign
    v-model="isManualAssignShow"
    action="submit"
    :cvm="currentCvm"
    :is-loading="hostStore.isAssignPreviewLoading"
    @submit="handleSubmit"
  />

  <!-- 关联配置平台主机 -->
  <match-host
    v-model="isMatchHostShow"
    ref="match-host-dialog"
    action="submit"
    :cvm="currentCvm"
    @manual-assign="handleManualAssign"
    @submit="handleSubmit"
  />
</template>

<style scoped lang="scss">
.loading-status {
  margin-top: -16px;
  min-height: 180px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;

  .loading-text {
    position: relative;
    &::after {
      content: '...';
      position: absolute;
    }
  }
}
</style>
