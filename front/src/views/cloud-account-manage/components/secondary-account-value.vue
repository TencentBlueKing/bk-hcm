<script setup lang="ts">
import { computed, watchEffect } from 'vue';
import CombineRequest from '@blueking/combine-request';
import { useCloudAccountStore } from '@/store/cloud-account';

const props = defineProps<{ value: string | string[]; bizId: number }>();

const localValue = computed(() => {
  if (!props.value) {
    return [];
  }
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((id) => {
    // 每次从全局store中查询获取
    const account = cloudAccountStore.allSecondaryAccountCacheList.get(id);
    if (!account) {
      return `${id} (--)`;
    }
    return `${id} (${account.name})`;
  });
  return names?.join?.(';') || '--';
});

const cloudAccountStore = useCloudAccountStore();

const combineRequest = CombineRequest.setup(Symbol.for('secondary-account-value'), (accountIds) => {
  const uniqueIds = [...new Set((accountIds as string[][]).reduce((acc, cur) => acc.concat(cur), []))];
  cloudAccountStore.getSecondaryAccountListByAccountIds(uniqueIds, props.bizId);
});

watchEffect(() => {
  if (!localValue.value.length) {
    return;
  }
  combineRequest.add(localValue.value);
});
</script>

<template>
  {{ displayValue }}
</template>
