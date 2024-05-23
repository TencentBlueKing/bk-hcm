<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { h, ref, watchEffect } from 'vue';
import { CloudType } from '@/typings';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const { getRegionName } = useRegionsStore();
const { getNameFromBusinessMap } = useBusinessMapStore();

const fields = ref([
  {
    name: '资源ID',
    prop: 'id',
  },
  {
    name: '云资源ID',
    prop: 'cloud_id',
  },
  {
    name: '云厂商',
    prop: 'vendor',
    render(cell: string) {
      return CloudType[cell] || '--';
    },
  },
  {
    name: '网络接口名称',
    prop: 'name',
  },
  {
    name: '账号',
    prop: 'account_id',
    link(val: string) {
      return `/#/resource/account/detail/?accountId=${val}&id=${val}`;
    },
  },
  {
    name: '地域',
    prop: 'region',
    render: (val: string) => getRegionName(props?.detail?.vendor, val),
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
    render: (val: number) => (val === -1 ? '未分配' : `${getNameFromBusinessMap(val)} (${val})`),
  },
  {
    name: '状态',
    prop: 'portState',
  },
  {
    name: '所属VPC',
    prop: 'cloud_vpc_id',
    render(val: string) {
      if (!val) {
        return '--';
      }
      return h(
        'div',
        { class: 'cell-content-list' },
        val?.split(';').map((item) => h('p', { class: 'cell-content-item' }, item?.split('/')?.pop())),
      );
    },
  },
  {
    name: '所属子网',
    prop: 'cloud_subnet_id',
    render(val: string) {
      if (!val) {
        return '--';
      }
      return h(
        'div',
        { class: 'cell-content-list' },
        val?.split(';').map((item) => h('p', { class: 'cell-content-item' }, item?.split('/')?.pop())),
      );
    },
  },
  {
    name: '已关联到主机ID',
    prop: 'instance_id',
  },
  {
    name: '安全组ID',
    prop: 'cloud_security_group_ids', // cloud_security_group_ids
    render(cell: string) {
      return cell || '--';
    },
  },
  {
    name: 'MAC地址',
    prop: 'mac_addr',
  },
]);

const data = ref([]);
watchEffect(() => {
  data.value = {
    ...props.detail,
    portState: props.detail?.port_state,
  };
});
</script>

<template>
  <div class="field-list">
    <detail-info :detail="data" :fields="fields" />
  </div>
</template>
