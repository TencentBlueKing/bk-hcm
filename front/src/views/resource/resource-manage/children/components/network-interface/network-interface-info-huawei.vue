<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { h, ref, watchEffect } from 'vue';
import { CloudType } from '@/typings';
import { useRegionsStore } from '@/store/useRegionsStore';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const { getRegionName } = useRegionsStore();

const fields = ref([
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
    name: '网络接口ID',
    prop: 'id',
  },
  {
    name: '账号',
    prop: 'account_id',
    link(val: string) {
      return `/#/resource/account/detail/?id=${val}`;
    },
  },
  {
    name: '地域',
    prop: 'region',
    render: (val: string) => getRegionName(props?.detail?.vendor, val)
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
    render(val: number) {
      return val === -1 ? '--' : val;
    },
  },
  {
    name: '状态',
    prop: 'portState',
  },
  {
    name: '所属网络(VPC)',
    prop: 'cloud_vpc_id',
    render(val: string) {
      if (!val) {
        return '--';
      }
      return h('div', { class: 'cell-content-list' }, val?.split(';')
        .map(item => h('p', { class: 'cell-content-item' }, item?.split('/')?.pop())));
    },
  },
  {
    name: '所属子网',
    prop: 'cloud_subnet_id',
    render(val: string) {
      if (!val) {
        return '--';
      }
      return h('div', { class: 'cell-content-list' }, val?.split(';')
        .map(item => h('p', { class: 'cell-content-item' }, item?.split('/')?.pop())));
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
    <detail-info
      class="mt20"
      :detail="data"
      :fields="fields"
    />
  </div>
</template>

<style lang="scss" scoped>
.field-list {
  :deep(.cell-content-list) {
    line-height: normal;
    .cell-content-item {
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}
</style>
