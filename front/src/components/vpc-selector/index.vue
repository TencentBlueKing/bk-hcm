<script lang="ts" setup>
import { watch, ref, watchEffect, defineExpose } from 'vue';
import {
  useResourceStore,
} from '@/store';

const props = defineProps({
  modelValue: {
    type: String,
  },
});

const emit = defineEmits(['update:modelValue']);

const resourceStore = useResourceStore();
const vpcList = ref([]);
const loading = ref(null);
const vpcPage = ref(0);
const selectedValue = ref(props.modelValue);
const hasMoreData = ref(true);

const getVpcList = async () => {
  if (!hasMoreData.value) return;
  loading.value = true;
  const res = await resourceStore.list({
    filter: { op: 'and', rules: [] },
    page: {
      start: vpcPage.value * 100,
      limit: 100,
    },
  }, 'vpcs');
  vpcPage.value += 1;
  vpcList.value.push(...res?.data?.details || []);
  hasMoreData.value = res?.data?.details?.length >= 100;   // 100条数据说明还有数据 可翻页
  loading.value = false;
};

watchEffect(void (async () => {
  getVpcList();
})());

watch(() => selectedValue.value, (val) => {
  emit('update:modelValue', val);
});

defineExpose({
  vpcList,
});
</script>

<template>
  <bk-select
    v-model="selectedValue"
    filterable
    @scroll-end="getVpcList"
    :loading="loading"
  >
    <bk-option
      v-for="(item, index) in vpcList"
      :key="index"
      :value="item.id"
      :label="item.name"
    />
  </bk-select>
</template>
