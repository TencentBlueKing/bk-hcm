<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import GcpRelate from '../components/gcp/gcp-relate.vue';
import useDetail from '@/views/resource/resource-manage/hooks/use-detail';
import useAdd from '@/views/resource/resource-manage/hooks/use-add';
import GcpAdd from '@/views/resource/resource-manage/children/add/gcp-add';

import {
  useI18n,
} from 'vue-i18n';

import { ref } from 'vue';

const {
  t,
} = useI18n();

const gcpFields = [
  {
    name: t('资源ID'),
    value: '1234223',
    prop: 'id',
  },
  {
    name: t('资源名称'),
    value: '1234223',
    link: 'http://www.baidu.com',
    prop: 'name',
  },
  {
    name: t('账号'),
    value: '1234223',
    prop: 'account_id',
  },
  {
    name: t('业务'),
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: t('云厂商'),
    value: '1234223',
    prop: 'account-name',
  },
  // {
  //   name: '日志',
  //   value: '1234223',
  //   prop: 'account-name',
  // },
  {
    name: 'vpc',
    value: '1234223',
    prop: 'vpc_id',
  },
  {
    name: t('优先级'),
    value: '1234223',
    prop: 'priority',
  },
  {
    name: t('方向'),
    value: '1234223',
    prop: 'type',
  },
  {
    name: t('对匹配项执行的操作'),
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: t('目标'),
    value: '1234223',
    prop: 'target_tags',
  },
  {
    name: t('来源过滤条件'),
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: t('实施'),
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: t('创建时间'),
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: t('修改时间'),
    value: '1234223',
    prop: 'account-name',
  },
  {
    name: t('备注'),
    value: '1234223',
    edit: true,
    prop: 'memo',
  },
];

const {
  loading,
  detail,
} = useDetail(
  'vendors/gcp/firewalls/rules',
  '1',
);
const gcpDetail = { id: 1, memo: '备注', name: 'test' } || detail;

const tabs = [
  {
    name: '关联实例',
    value: 'relate',
  },
];

const isShowGcpAdd = ref(false);
const gcpTitle = ref<string>(t('新增'));
const isAdd = ref(false);

const handleGcpAdd = (add: boolean) => {
  isShowGcpAdd.value = true;
  isAdd.value = add;
};


// 新增修改防火墙规则
const submit = (data: any) => {
  const fetchType = data?.id ? 'vendors/gcp/firewalls/rules/' : 'vendors/gcp/firewalls/rules/create';
  const {
    addData,
    updateData,
  } = useAdd(
    fetchType,
    data,
    data?.id,
  );
  if (isAdd.value) {   // 新增
    addData();
  } else {
    updateData();
  }
  isShowGcpAdd.value = false;
};
</script>

<template>
  <detail-header>
    {{t('GCP防火墙')}}：ID（xxx）
    <template #right>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleGcpAdd(false)"
      >
        {{ t('修改') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
      >
        {{ t('删除') }}
      </bk-button>
    </template>
  </detail-header>
  <bk-loading
    :loading="loading"
  >
    <detail-info
      :fields="gcpFields"
      :detail="gcpDetail"
    />
  </bk-loading>

  <detail-tab
    :tabs="tabs"
  >
    <gcp-relate></gcp-relate>
  </detail-tab>
  <gcp-add
    v-model:is-show="isShowGcpAdd"
    :gcp-title="gcpTitle"
    :is-add="isAdd"
    :detail="gcpDetail"
    @submit="submit"></gcp-add>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
</style>
