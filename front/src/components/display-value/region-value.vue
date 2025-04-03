<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useRegionStore } from '@/store/region';
import { ResourceTypeEnum } from '@/common/resource-constant';

const props = defineProps<{
  value: string | string[];
  vendor: string;
  resourceType?: ResourceTypeEnum.CVM | ResourceTypeEnum.VPC | ResourceTypeEnum.DISK | ResourceTypeEnum.SUBNET;
}>();

const list = ref([]);

const localValue = computed(() => {
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((id) => {
    return list.value.find((item) => item.id === id)?.name;
  });
  return names?.join?.(', ');
});

const regionStore = useRegionStore();

watchEffect(async () => {
  if (props.vendor) {
    list.value = await regionStore.getRegionList({ vendor: props.vendor, resourceType: props.resourceType });
  }
});
</script>

<template>
  {{ displayValue }}
</template>
