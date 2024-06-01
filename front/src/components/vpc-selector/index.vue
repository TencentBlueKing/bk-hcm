<script lang="ts" setup>
import { watch, ref, watchEffect, defineExpose } from 'vue';
import { useResourceStore } from '@/store';
import { VendorEnum } from '@/common/constant';

const props = defineProps({
  vendor: {
    type: String,
    default() {
      return 'tcloud';
    },
  },
  region: {
    type: String,
    default() {
      return '';
    },
  },
  modelValue: {
    type: String,
  },
  isDisabled: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['update:modelValue', 'handleVpcDetail']);

const resourceStore = useResourceStore();
const vpcList = ref([]);
const loading = ref(null);
const vpcPage = ref(0);
const selectedValue = ref(props.modelValue);
const hasMoreData = ref(true);

const getVpcList = async () => {
  if (!hasMoreData.value) return;
  loading.value = true;
  const rulesData = [];
  if (props.vendor) {
    if (props.vendor === 'gcp') {
      rulesData.push({ field: 'vendor', op: 'eq', value: props.vendor });
    } else {
      rulesData.push(
        { field: 'vendor', op: 'eq', value: props.vendor },
        { field: 'region', op: 'eq', value: props.region },
      );
    }
  }
  const res = await resourceStore.list(
    {
      filter: { op: 'and', rules: rulesData },
      page: {
        start: vpcPage.value * 100,
        limit: 100,
      },
    },
    'vpcs',
  );
  vpcPage.value += 1;
  vpcList.value.push(...(res?.data?.details || []));
  hasMoreData.value = res?.data?.details?.length >= 100; // 100条数据说明还有数据 可翻页
  loading.value = false;
};

watchEffect(
  void (async () => {
    getVpcList();
  })(),
);

watch(
  () => selectedValue.value,
  (val) => {
    const vpcId = val && vpcList.value.find((e: any) => e.cloud_id === val)?.id;
    emit('handleVpcDetail', vpcId);
    emit('update:modelValue', val);
  },
);

watch(
  () => props.vendor,
  () => {
    vpcPage.value = 0;
    hasMoreData.value = true;
    vpcList.value = [];
    selectedValue.value = '';
    getVpcList();
  },
);

watch(
  () => props.region,
  () => {
    vpcPage.value = 0;
    hasMoreData.value = true;
    vpcList.value = [];
    selectedValue.value = '';
    getVpcList();
  },
);

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
    :disabled="props.isDisabled"
  >
    <bk-option
      v-for="(item, index) in vpcList"
      :key="index"
      :value="item.cloud_id"
      :label="
        item.vendor === VendorEnum.AZURE ? item.name || item.cloud_id : `${item.cloud_id}（${item.name || '--'}）`
      "
    />
  </bk-select>
</template>
