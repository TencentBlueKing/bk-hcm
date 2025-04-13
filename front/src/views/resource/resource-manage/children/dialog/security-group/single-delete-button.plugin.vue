<script setup lang="ts">
import { useAttrs } from 'vue';
import { useResourceStore } from '@/store';
import DeleteButton from './single-delete-button.vue';

const attrs = useAttrs();
const emit = defineEmits<(e: 'success') => void>();

const resourceStore = useResourceStore();
const loading = defineModel('loading', { default: false });

const handleDelete = async () => {
  loading.value = true;
  try {
    await resourceStore.deleteBatch('security_groups', { ids: [attrs.id] });
    emit('success');
  } finally {
    loading.value = false;
  }
};
</script>
<template>
  <!-- 将透传attrs -->
  <delete-button @del="handleDelete" :loading="loading"></delete-button>
</template>
