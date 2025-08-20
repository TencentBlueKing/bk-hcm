<script setup lang="ts">
import { computed } from 'vue';
import CombineRequest from '@blueking/combine-request';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';

const props = defineProps<{ value: number; showId?: boolean }>();

const cloudAreaStore = useCloudAreaStore();

const displayValue = computed(() => {
  const cloudArea = cloudAreaStore.cloudAreaList.find((item) => item.id === props.value);
  if (!cloudArea) return '--';
  return props.showId ? `[${props.value}] ${cloudArea.name}` : cloudArea.name;
});

const combineRequest = CombineRequest.setup(Symbol.for('cloud-area-value'), async () => {
  cloudAreaStore.fetchAllCloudAreas();
});

combineRequest.add(null);
</script>

<template>
  {{ displayValue }}
</template>
