<script lang="ts" setup>
import { ref, watch, PropType } from 'vue';
import type { FilterType } from '@/typings/resource';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';

import { useI18n } from 'vue-i18n';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

const { t } = useI18n();

const cvmColumns = [
  {
    label: '实例 ID',
    field: 'id',
  },
  {
    label: '所属 VPC',
    field: 'id',
  },
  {
    label: 'IP 信息',
    field: 'id',
  },
];
const cvmData = ref<any>([
  {
    id: 233,
  },
]);

const networkData = ref<any>([
  {
    id: 233,
  },
]);

const networkColumns = [
  {
    label: '实例 ID',
    field: 'id',
  },
  {
    label: '所属 VPC',
    field: 'id',
  },
  {
    label: '可用区',
    field: 'id',
  },
  {
    label: 'IP 信息',
    field: 'id',
  },
  {
    label: '关联主机',
    field: 'id',
  },
];
// tab 信息
const types = [
  { name: 'cvm', label: 'CVM' },
  { name: 'network', label: t('网络接口') },
];
const activeType = ref('cvm');

const fetchList = (fetchType: string) => {
  const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryList(
    props,
    fetchType,
  );
  return {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  };
};

watch(
  () => activeType.value,
  (v) => {
    console.log('value', v);
    if (v === 'cvm') {
      const { datas } = fetchList('security_groups/cvms/relations');
      cvmData.value = datas;
    } else if (v === 'network') {
      const { datas } = fetchList('security_groups/subnets/relations');
      networkData.value = datas;
    }
  },
  { immediate: true },
);
</script>

<template>
  <bk-radio-group class="mt20" v-model="activeType">
    <bk-radio-button v-for="item in types" :key="item.name" :label="item.name">
      {{ item.label }}
    </bk-radio-button>
  </bk-radio-group>

  <bk-table
    v-if="activeType === 'cvm'"
    class="mt20"
    row-hover="auto"
    :columns="cvmColumns"
    :data="cvmData.value"
    show-overflow-tooltip
  />

  <bk-table
    v-if="activeType === 'network'"
    class="mt20"
    row-hover="auto"
    :columns="networkColumns"
    :data="networkData.value"
    show-overflow-tooltip
  />
</template>

<style lang="scss" scoped>
.info-title {
  font-size: 14px;
  margin-bottom: 8px;
}
</style>
