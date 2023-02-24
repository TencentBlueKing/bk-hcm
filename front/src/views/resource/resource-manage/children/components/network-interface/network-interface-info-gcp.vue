<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { h, ref, watchEffect } from 'vue';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const fields = ref([
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
  },
  {
    name: '可用区域',
    prop: 'zone',
  },
  {
    name: '业务',
    prop: 'bk_biz_id',
    render(val: number) {
      return val === -1 ? '--' : val;
    },
  },
  {
    name: '内网IP',
    prop: 'internal_ip',
  },
  {
    name: '公网IP',
    prop: 'internal_ip',
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
    name: '已关联到',
    prop: 'instance_id',
  },
  {
    name: '网络层级',
    prop: 'networkTier',
    render(val: string) {
      const vals = { PREMIUM: '高级', STANDARD: '标准' };
      return vals[val];
    },
  },
  {
    name: 'IP转发',
    prop: 'can_ip_forward',
    render(val: boolean) {
      return  val ? '开启' : '关闭';
    },
  },
]);

const data = ref([]);

watchEffect(() => {
  data.value = {
    ...props.detail,
    networkTier: props.detail?.access_configs?.[0]?.network_tier,
  };
});
</script>

<template>
  <div class="field-list">
    <detail-info
      class="field-list mt20"
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
