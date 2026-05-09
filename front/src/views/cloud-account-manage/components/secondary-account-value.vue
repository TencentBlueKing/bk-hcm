<script setup lang="ts">
import { computed, watchEffect } from 'vue';
import CombineRequest from '@blueking/combine-request';
import { useSecondaryAccountStore } from '@/store/cloud-account-manage/secondary-account';
import { SecondaryAccountResourceTypeEnum, VendorEnum } from '@/common/constant';

const props = defineProps<{ value: string | string[]; vendor: VendorEnum; resType: string; bizId?: number }>();

const localValue = computed(() => {
  if (!props.value) {
    return [];
  }
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((id) => {
    // 每次从全局store中查询获取
    const account = secondaryAccountStore.allSecondaryAccountCacheList.get(`${id}@${resType.value}@${bizId.value}`);
    if (!account) {
      return '';
    }
    return `${account?.extension?.cloud_main_account_id} (${account.name})`;
  });
  return names?.join?.(';') || '--';
});
const vendor = computed(() => props.vendor);
const resType = computed(() => props.resType);
const bizId = computed(() => props.bizId || 0);

const secondaryAccountStore = useSecondaryAccountStore();

const combineRequest = CombineRequest.setup(Symbol.for('secondary-account-value'), (params: any[]) => {
  const requestIdsMap = new Map<string, string[]>();
  params.forEach(([accountIds, vendor, resType, bizId]) => {
    const uniqueIds = [...new Set((accountIds as string[][]).reduce((acc, cur) => acc.concat(cur), []))];
    const key = `${bizId}@${vendor}@${resType}`;
    const value = requestIdsMap.get(key) ?? [];
    requestIdsMap.set(key, [...value, ...uniqueIds]);
  });
  // 将map数据拆解出来通过key去调取接口
  for (const [key, value] of requestIdsMap) {
    const [bizId, vendor, resType] = key.split('@');
    secondaryAccountStore.getSecondaryAccountListByAccountIds(
      value,
      vendor as VendorEnum,
      resType as SecondaryAccountResourceTypeEnum,
      +bizId,
    );
  }
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
