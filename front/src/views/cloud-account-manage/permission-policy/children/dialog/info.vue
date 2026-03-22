<script setup lang="ts">
import { computed } from 'vue';
import JSON from '@/views/cloud-account-manage/components/json.vue';

interface IProps {
  show: boolean;
  json: string;
  accountId: string;
  id: string;
  name: string;
}

const props = withDefaults(defineProps<IProps>(), {
  show: false,
  json: '',
  accountId: '',
  id: '',
  name: '',
});

const emit = defineEmits(['close']);

const show = computed(() => props.show);

const handleClose = () => {
  emit('close');
};
</script>

<template>
  <bk-dialog
    v-model:is-show="show"
    title="云上权限模板详情"
    quick-close
    class="model-info-dialog"
    @closed="handleClose"
  >
    <div class="model-info-item">二级账号: {{ props.accountId }}</div>
    <div class="model-info-item">模板名称: {{ props.name }}</div>
    <div class="model-info-item">云上策略ID: {{ props.id }}</div>
    <div class="model-info-json">
      <JSON :content="props.json" />
    </div>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.model-info-dialog {
  .model-info-json {
    background: hsl(216, 33%, 97%);
    height: 350px;
    overflow-y: auto;
    padding: 20px;
  }
  .model-info-item {
    font-weight: 400;
    font-size: 12px;
    line-height: 20px;
    margin-bottom: 8px;
  }
  :deep(.bk-modal-footer) {
    display: none;
  }
}
</style>
