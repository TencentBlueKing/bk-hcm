<script setup lang="ts">
import FlexTag from '@/components/flex-tag/index.vue';
import { useBusinessGlobalStore } from '@/store/business-global';
import { computed } from 'vue';

const props = defineProps<{ value: number[] }>();

const businessGlobalStore = useBusinessGlobalStore();

const list = computed(() => {
  if (props.value?.[0] === -1) {
    return [{ name: '全部业务' }];
  }

  const names = [];
  for (const value of props.value) {
    const name = businessGlobalStore.businessFullList.find((item) => item.id === value)?.name ?? '--';
    names.push({ name });
  }
  return names;
});
</script>

<template>
  <flex-tag :is-tag-style="true" :list="list" v-if="value?.length" />
  <span v-else>--</span>
</template>
