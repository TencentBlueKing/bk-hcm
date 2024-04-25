<script lang="ts" setup>
import { ref, watchEffect, defineExpose, watch, computed } from 'vue';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { useBusinessStore } from '@/store';

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
  // 暂时用 delayed 来取消 props.vendor 的即时监听
  delayed: {
    type: Boolean,
    default: false,
  },
  // 是否加载中
  isLoading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['update:modelValue']);

const businessStore = useBusinessStore();
const zonesList = ref([]);
const loading = ref(null);
const zonePage = ref(0);
const selectedValue = computed({
  get() {
    return props.modelValue;
  },
  set(val) {
    emit('update:modelValue', val);
  },
});
const hasMoreData = ref(true);

const filter = ref<QueryFilterType>({
  op: 'and',
  rules: [],
});

const getZonesData = async () => {
  if (!hasMoreData.value || !props.vendor || !props.region) return;
  loading.value = true;
  const res = await businessStore.getZonesList({
    vendor: props.vendor,
    region: props.region,
    data: {
      filter: filter.value,
      page: {
        start: zonePage.value * 100,
        limit: 100,
      },
    },
  });
  zonePage.value += 1;
  zonesList.value.push(...(res?.data?.details || []));
  hasMoreData.value = res?.data?.details?.length >= 100; // 100条数据说明还有数据 可翻页
  loading.value = false;
};

const resetData = () => {
  zonePage.value = 0;
  hasMoreData.value = true;
  zonesList.value = [];
  selectedValue.value = '';
};

watchEffect(
  void (async () => {
    getZonesData();
  })(),
);

watch(
  () => props.vendor,
  (val) => {
    switch (val) {
      case VendorEnum.TCLOUD:
        filter.value.rules = [
          {
            field: 'vendor',
            op: QueryRuleOPEnum.EQ,
            value: val,
          },
          {
            field: 'state',
            op: QueryRuleOPEnum.EQ,
            value: 'AVAILABLE',
          },
        ];
        break;
      default:
        filter.value.rules = [];
      /* case VendorEnum.AWS:
      filter.value.rules = [
        {
          field: 'state',
          op: QueryRuleOPEnum.EQ,
          value: 'available',
        },
      ];
      break;
    case VendorEnum.GCP:
      filter.value.rules = [
        {
          field: 'state',
          op: QueryRuleOPEnum.EQ,
          value: 'UP',
        },
      ];
      break;*/
    }
    resetData();
    getZonesData();
  },
  { immediate: !props.delayed },
);

watch(
  () => props.region,
  () => {
    resetData();
    getZonesData();
  },
);

defineExpose({
  zonesList,
});
</script>

<template>
  <bk-select v-model="selectedValue" filterable @scroll-end="getZonesData" :loading="loading || isLoading">
    <bk-option v-for="(item, index) in zonesList" :key="index" :value="item.name" :label="item.name_cn || item.name" />
  </bk-select>
</template>
