<script lang="ts" setup>
// @ts-nocheck
import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import GcpRelate from '../components/gcp/gcp-relate.vue';
import useDetail from '@/views/resource/resource-manage/hooks/use-detail';
import useAdd from '@/views/resource/resource-manage/hooks/use-add';
import GcpAdd from '@/views/resource/resource-manage/children/add/gcp-add';
import { GcpTypeEnum } from '@/typings';

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
    prop: 'account_id',
  },
  {
    name: t('业务'),
    prop: 'account-name',
  },
  {
    name: t('云厂商'),
    value: '谷歌云',
    prop: 'vendor',
  },
  {
    name: '日志',
    prop: 'log_enable',
  },
  {
    name: 'vpc',
    prop: 'cloud_vpc_id',
  },
  {
    name: t('优先级'),
    prop: 'priority',
  },
  {
    name: t('方向'),
    prop: 'type',
  },
  {
    name: t('对匹配项执行的操作'),
    prop: 'operate',
  },
  {
    name: t('目标'),
    prop: 'target',
  },
  {
    name: t('来源过滤条件'),
    prop: 'source',
  },
  {
    name: t('协议和端口'),
    prop: 'ports',
  },
  {
    name: t('实施'),
    prop: 'disabled',
  },
  {
    name: t('创建时间'),
    prop: 'create_at',
  },
  {
    name: t('修改时间'),
    prop: 'update_at',
  },
  {
    name: t('备注'),
    edit: true,
    type: 'textarea',
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
detail.value = { id: 1, memo: '备注', name: 'test', log_enable: false, disabled: true, type: 'egress', priority: 100, cloud_vpc_id: 1, account_id: '1111', allowed: [{
  protocol: 'tcp',
  ports: [
    '443',
  ],
}] };
const gcpDetail = { ...detail.value };
gcpDetail.type = GcpTypeEnum[gcpDetail.type];
gcpDetail.log_enable = detail?.log_enable ? t('开') : t('关');
gcpDetail.operate = detail?.allowed?.length ? t('允许') : t('拒绝');
gcpDetail.disabled = detail?.disabled ? t('已启用') : t('已停用');
gcpDetail.vendor = t('谷歌云');
detail.allowed = [{
  protocol: 'tcp',
  ports: [
    '443',
    '43',
  ],
}, {
  protocol: 'tgp',
  ports: [
    '443',
    '43',
  ],
}].reduce((p, e) => {
  p.push(`${[e.protocol]}: ${e.ports.join(',')}`);
  return p;
}, []);
gcpDetail.ports = gcpDetail.operate ? detail.allowed : detail.denied;
detail.target_service_accounts = ['https-server1'];
detail.destination_ranges = ['https-server2'];
detail.target_tags = ['https-server3'];
detail.source_ranges = ['https-server4'];
detail.source_service_accounts = ['https-server5'];
detail.source_tags = ['https-server6'];
gcpDetail.target = [...detail?.destination_ranges, ...detail?.target_service_accounts, ...detail?.target_tags];
gcpDetail.source = [...detail?.source_ranges, ...detail?.source_service_accounts, ...detail?.source_tags];

const tabs = [
  {
    name: t('关联实例'),
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
    {{t('GCP防火墙')}}：ID（{{gcpDetail.id}}）
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
    :detail="detail"
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
