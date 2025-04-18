<script setup lang="ts">
import { useResourceStore } from '@/store';
import DeleteButton from './single-delete-button.vue';

const props = withDefaults(defineProps<{ id: string; disabled: boolean }>(), {
  disabled: true,
});

const emit = defineEmits<(e: 'success') => void>();

const resourceStore = useResourceStore();
const loading = defineModel('loading', { default: false });

const handleDelete = async () => {
  loading.value = true;
  try {
    await resourceStore.deleteBatch('security_groups', { ids: [props.id] });
    emit('success');
  } finally {
    loading.value = false;
  }
};
</script>
<template>
  <delete-button :disabled="disabled" :loading="loading" @del="handleDelete"></delete-button>
</template>
