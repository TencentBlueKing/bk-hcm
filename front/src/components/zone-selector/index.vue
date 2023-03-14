<script lang="ts" setup>
import {  ref, watchEffect, defineExpose, watch } from 'vue';
import {
  useBusinessStore,
} from '@/store';

const props = defineProps({
  vendor: {
    type: String,
  },
  region: {
    type: String,
  },
  modelValue: {
    type: String,
  },
});

const emit = defineEmits(['update:modelValue']);

const businessStore = useBusinessStore();
const zonesList = ref([]);
const loading = ref(null);
const zonePage = ref(0);
const selectedValue = ref(props.modelValue);
const hasMoreData = ref(true);

const getZonesData = async () => {
  if (!hasMoreData.value || !props.vendor || !props.region) return;
  loading.value = true;
  const res = await businessStore.getZonesList({
    vendor: props.vendor,
    region: props.region,
    page: {
      start: zonePage.value * 100,
      limit: 100,
    },
  });
  zonePage.value += 1;
  zonesList.value.push(...res?.data?.details || []);
  hasMoreData.value = res?.data?.details?.length >= 100;   // 100条数据说明还有数据 可翻页
  loading.value = false;
};

watchEffect(void (async () => {
  getZonesData();
})());

watch(() => props.vendor, () => {
  resetData();
  getZonesData();
});

watch(() => props.region, () => {
  resetData();
  getZonesData();
});

watch(() => selectedValue.value, (val) => {
  emit('update:modelValue', val);
});

const resetData = () => {
  zonePage.value = 0;
  hasMoreData.value = true;
  zonesList.value = [];
  selectedValue.value = '';
};

defineExpose({
  zonesList,
});
</script>

<template>
  <bk-select
    v-model="selectedValue"
    filterable
    @scroll-end="getZonesData"
    :loading="loading"
  >
    <bk-option
      v-for="(item, index) in zonesList"
      :key="index"
      :value="item.name"
      :label="item.name_cn || item.name"
    />
  </bk-select>
</template>
