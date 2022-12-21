<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import DetailInfo from '../../common/info/detail-info';
import SubnetRoute from '../../children/components/subnet/subnet-route.vue';

import {
  useI18n,
} from 'vue-i18n';
import useColumns from '../../hooks/use-columns';
import useDetail from '../../hooks/use-detail';
import useDelete from '../../hooks/use-delete';

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
  {
    name: '路由策略',
    value: 'network',
  },
];

const settingFields = [
  {
    name: '账号',
    prop: 'account_id',
    link: 'http://www.baidu.com',
  },
  {
    name: '资源 ID',
    prop: 'id',
  },
  {
    name: '资源名称',
    prop: 'name',
  },
  {
    name: '云厂商',
    prop: 'vendor',
  },
  {
    name: '业务',
    prop: '1234223',
  },
  {
    name: '状态',
    prop: 'status',
  },
  {
    name: '是否默认子网',
    prop: 'is_default',
  },
  {
    name: '所属 VPC',
    prop: 'is_default',
  },
  {
    name: 'IPv4 CIDR',
    prop: 'ipv4_cidr',
  },
  {
    name: '可用 IPv4 地址数',
    prop: 'ipv4_cidr',
  },
  {
    name: 'IPv6 CIDR',
    prop: 'ipv6_cidr',
  },
  {
    name: '可用 IPv6 地址数',
    prop: 'ipv4_cidr',
  },
  {
    name: '地域 ID',
    prop: 'region',
  },
  {
    name: '地域名称',
    prop: '12312321',
  },
  {
    name: '可用区 ID',
    prop: '1234223',
  },
  {
    name: '可用区名称',
    prop: '1234223',
  },
  {
    name: '创建时间',
    prop: 'create_at',
  },
  {
    name: '备注',
    prop: 'memo',
    edit: true,
  },
];

const {
  t,
} = useI18n();

const columns = useColumns('subnet');

const {
  loading,
  detail,
} = useDetail(
  'subnet',
  '1',
);

const {
  handleShowDelete,
  DeleteDialog,
} = useDelete(
  columns,
  [detail],
  'subnet',
  t('删除子网'),
);
</script>

<template>
  <bk-loading
    :loading="loading"
  >
    <detail-header>
      子网：ID（xxx）
      <template #right>
        <bk-button
          class="w100 ml10"
          theme="primary"
          @click="handleShowDelete"
        >
          {{ t('删除') }}
        </bk-button>
      </template>
    </detail-header>

    <detail-tab
      :tabs="hostTabs"
    >
      <template #default="type">
        <detail-info
          v-if="type === 'detail'"
          :fields="settingFields"
          :detail="detail"
        />
        <subnet-route
          v-if="type === 'network'"
        />
      </template>
    </detail-tab>
  </bk-loading>

  <delete-dialog>
    {{ t('请注意该子网包含一个或多个资源，在释放这些资源前，无法删除VPC') }}<br />
    {{ t('CVM：{count} 个', { count: 5 }) }}
  </delete-dialog>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
</style>
