<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';

import { PropType } from 'vue';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import { FieldList } from '@/views/resource/resource-manage/common/info-list/types';

const props = defineProps({
  data: {
    type: Object as PropType<any>,
  },
});

const cvmInfo: FieldList = [
  {
    name: '实例名称',
    prop: 'name',
  },
  {
    name: '实例ID',
    prop: 'cloud_id',
  },
  {
    name: '账号',
    prop: 'account_id',
    render: () => '内置账号',
  },
  {
    name: '云厂商',
    prop: 'vendorName',
  },
  {
    name: '可用区域',
    prop: 'zone',
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
    render() {
      return props.data.bk_biz_id === -1 ? '未分配' : `${props.data.bk_biz_id_name} (${props.data.bk_biz_id})`;
    },
  },
  {
    name: '当前状态',
    prop: 'status',
    cls(val: string) {
      return `status-${val}`;
    },
    render() {
      return CLOUD_HOST_STATUS[props.data.status];
    },
  },
  {
    name: '备注',
    prop: 'memo',
  },
];
</script>

<template>
  <h3 class="info-title">实例信息</h3>
  <div class="wrap-info">
    <detail-info :fields="cvmInfo" :detail="props.data" global-copyable></detail-info>
  </div>
</template>
