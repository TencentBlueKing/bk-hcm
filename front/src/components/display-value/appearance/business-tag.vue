<script setup lang="ts">
import FlexTag from '@/components/flex-tag/index.vue';
import { useBusinessGlobalStore } from '@/store/business-global';
import { computed } from 'vue';
import { DisplayType } from '../typings';

const props = defineProps<{
  value: number | number[];
  displayValue: string;
  displayOn?: DisplayType['on'];
}>();

const businessGlobalStore = useBusinessGlobalStore();

const list = computed(() => {
  const values = Array.isArray(props.value) ? props.value : [props.value];

  if (values?.[0] === -1) {
    return [{ name: '全部业务' }];
  }

  const names = [];
  for (const v of values) {
    if (v) {
      const name = businessGlobalStore.businessFullList.find((item) => item.id === v)?.name ?? '--';
      names.push({ name });
    }
  }
  return names;
});
</script>

<template>
  <flex-tag :is-tag-style="true" :list="list" v-if="list.length" />
  <span v-else>--</span>
</template>
