<script setup lang="ts">
import { computed, watchEffect } from 'vue';
import CombineRequest from '@blueking/combine-request';
import { useCloudAccountStore } from '@/store/cloud-account';
import { VendorEnum } from '@/common/constant';

const props = defineProps<{ value: string | string[]; vendor?: VendorEnum; resType: string; bizId?: number }>();

const localValue = computed(() => {
  if (!props.value) {
    return [];
  }
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((id) => {
    // 每次从全局store中查询获取
    const account = cloudAccountStore.allSecondaryAccountCacheList.get(`${id}@${resType.value}`);
    if (!account) {
      return `${id} (--)`;
    }
    return `${id} (${account.name})`;
  });
  return names?.join?.(';') || '--';
});
const vendor = computed(() => props.vendor);
const resType = computed(() => props.resType);
const bizId = computed(() => props.bizId);

const cloudAccountStore = useCloudAccountStore();

const combineRequest = CombineRequest.setup(Symbol.for('secondary-account-value'), (params: any[]) => {
  params.map(([accountIds, vendor, resType, bizId]) => {
    const uniqueIds = [...new Set((accountIds as string[][]).reduce((acc, cur) => acc.concat(cur), []))];
    return cloudAccountStore.getSecondaryAccountListByAccountIds(uniqueIds, vendor, resType, bizId);
  });
});

watchEffect(() => {
  if (!localValue.value.length || !vendor.value) {
    return;
  }
  combineRequest.add([localValue.value, vendor.value, resType.value, bizId.value]);
});
</script>

<template>
  {{ displayValue }}
</template>
