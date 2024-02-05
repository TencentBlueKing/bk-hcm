<script lang="ts" setup>
import { ref, watchEffect, defineExpose, watch } from 'vue';
import { useBusinessStore } from '@/store';

const props = defineProps({
  modelValue: {
    type: String,
  },
  vendor: {
    type: String,
  },
});

const emit = defineEmits(['update:modelValue']);

const businessStore = useBusinessStore();
const resourceGroupList = ref([]);
const loading = ref(null);
const zonePage = ref(0);
const selectedValue = ref(props.modelValue);
const hasMoreData = ref(true);

const getResourceGroupData = async () => {
  if (!hasMoreData.value) return;
  loading.value = true;
  const res = await businessStore.getResourceGroupList({
    filter: { op: 'and', rules: [] },
    page: {
      start: zonePage.value * 100,
      limit: 100,
    },
  });
  zonePage.value += 1;
  resourceGroupList.value.push(...(res?.data?.details || []));
  hasMoreData.value = res?.data?.details?.length >= 100; // 100条数据说明还有数据 可翻页
  loading.value = false;
};

watchEffect(
  void (async () => {
    getResourceGroupData();
  })(),
);

watch(
  () => selectedValue.value,
  (val) => {
    emit('update:modelValue', val);
  },
);

defineExpose({
  resourceGroupList,
});
</script>

<template>
  <bk-select v-model="selectedValue" filterable @scroll-end="getResourceGroupData" :loading="loading">
    <bk-option
      v-for="(item, index) in resourceGroupList"
      :key="index"
      :value="props.vendor === 'azure' ? item.name : item.id"
      :label="item.name"
    />
  </bk-select>
</template>
