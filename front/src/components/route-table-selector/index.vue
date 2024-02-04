<script lang="ts" setup>
import { ref, watchEffect, defineExpose, watch } from 'vue';
import { useBusinessStore, useAccountStore } from '@/store';

const props = defineProps({
  modelValue: {
    type: String,
  },
  cloudVpcId: {
    type: String,
  },
});

const emit = defineEmits(['update:modelValue']);

const businessStore = useBusinessStore();
const accountStore = useAccountStore();
const tableList = ref([]);
const loading = ref(null);
const page = ref(0);
const selectedValue = ref(props.modelValue);
const hasMoreData = ref(true);

console.log(accountStore.bizs);
const getSelectData = async () => {
  if (!hasMoreData.value || !props.cloudVpcId) return;
  loading.value = true;
  const res = await businessStore.getRouteTableList({
    // { field: 'bk_biz_id', op: 'eq', value: accountStore.bizs },
    filter: { op: 'and', rules: [{ field: 'cloud_vpc_id', op: 'eq', value: props.cloudVpcId }] },
    page: {
      start: page.value * 100,
      limit: 100,
    },
  });
  page.value += 1;
  tableList.value.push(...(res?.data?.details || []));
  hasMoreData.value = res?.data?.details?.length >= 100; // 100条数据说明还有数据 可翻页
  loading.value = false;
};

watchEffect(
  void (async () => {
    getSelectData();
  })(),
);

watch(
  () => selectedValue.value,
  (val) => {
    emit('update:modelValue', val);
  },
);

watch(
  () => props.cloudVpcId,
  () => {
    page.value = 0;
    hasMoreData.value = true;
    selectedValue.value = '';
    tableList.value = [];
    getSelectData();
  },
  { deep: true },
);

defineExpose({
  tableList,
});
</script>

<template>
  <bk-select v-model="selectedValue" filterable @scroll-end="getSelectData" :loading="loading">
    <bk-option v-for="(item, index) in tableList" :key="index" :value="item.cloud_id" :label="item.name" />
  </bk-select>
</template>
