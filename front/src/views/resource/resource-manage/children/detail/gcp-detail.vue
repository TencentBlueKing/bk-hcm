<script lang="ts" setup>
// @ts-nocheck
import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
// import GcpRelate from '../components/gcp/gcp-relate.vue';
import useDetail from '@/views/resource/resource-manage/hooks/use-detail';
import useAdd from '@/views/resource/resource-manage/hooks/use-add';
import GcpAdd from '@/views/resource/resource-manage/children/add/gcp-add';
import { GcpTypeEnum, CloudType } from '@/typings';
// import bus from '@/common/bus';

import {
  useRoute,
} from 'vue-router';

import {
  useResourceStore,
} from '@/store/resource';
import { useBusinessMapStore } from '@/store/useBusinessMap';

import {
  useI18n,
} from 'vue-i18n';

import { ref, watch } from 'vue';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { timeFormatter } from '@/common/util';

const route = useRoute();
const resourceStore = useResourceStore();
const { getNameFromBusinessMap } = useBusinessMapStore();
const { whereAmI } = useWhereAmI();

const {
  t,
} = useI18n();

const hostTabs = [
  {
    name: '基本信息',
    value: 'detail',
  },
];

const id = route.query?.id;
const gcpDetail = ref({});
const gcpLoading = ref(true);

// const authVerifyData: any = inject('authVerifyData');
// const isResourcePage: any = inject('isResourcePage');

// const actionName = computed(() => {   // 资源下没有业务ID
//   console.log('isResourcePage.value', isResourcePage.value);
//   return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
// });

const {
  loading,
  detail,
} = useDetail(
  'vendors/gcp/firewalls/rules',
  id,
);

const fetchDetail = async () => {
  gcpLoading.value = true;
  try {
    const { data } = await resourceStore.detail('vendors/gcp/firewalls/rules', id);
    data.vendorName = CloudType[data.vendor];
    // data.bk_biz_id = data.bk_biz_id === -1 ? '全部' : data.bk_biz_id;
    detail.value = {
      ...data,
      ...data.spec,
      ...data.attachment,
      ...data.revision,
    };
    gcpDetail.value = { ...detail.value };
    handleDetailData();
  } catch (error) {
    console.log(error);
  } finally {
    gcpLoading.value = false;
  }
};

watch(
  () => loading.value,
  (v) => {
    gcpLoading.value = v;
    if (!v) {
      gcpDetail.value = { ...detail.value };
      handleDetailData();
    }
  },
);


const gcpFields = [
  {
    name: t('资源ID'),
    prop: 'id',
  },
  {
    name: t('资源名称'),
    prop: 'name',
  },
  {
    name: t('账号'),
    prop: 'account_id',
  },
  {
    name: t('云资源ID'),
    prop: 'cloud_id',
  },
  {
    name: t('业务'),
    prop: 'bk_biz_id',
    render: (val: number) => (val === -1 ? '未分配' : `${getNameFromBusinessMap(val)} (${val})`),
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
    name: 'VPC',
    prop: 'vpc_id',
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
    prop: 'created_at',
    render: (val: string) => timeFormatter(val),
  },
  {
    name: t('修改时间'),
    prop: 'updated_at',
    render: (val: string) => timeFormatter(val),
  },
  {
    name: t('备注'),
    // edit: true,
    type: 'textarea',
    prop: 'memo',
  },
];
const handleDetailData = () => {
  console.log('detail', detail.value.account_id);
  detail.target_service_accounts = detail.value.target_service_accounts || [];
  detail.destination_ranges = detail.value.destination_ranges || [];
  detail.target_tags = detail.value.target_tags || [];
  detail.source_ranges = detail.value.source_ranges || [];
  detail.source_service_accounts = detail.value.source_service_accounts || [];
  detail.source_tags = detail.value.source_tags || [];
  // gcpDetail.value.bk_biz_id = detail.value.bk_biz_id === -1 ? '全部' : detail.value.bk_biz_id;
  gcpDetail.value.type = GcpTypeEnum[gcpDetail.value.type];
  gcpDetail.value.log_enable = detail?.value?.log_enable ? t('开') : t('关');
  gcpDetail.value.operate = detail.value?.allowed?.length ? t('允许') : t('拒绝');
  gcpDetail.value.disabled = detail.value?.disabled ? t('已停用') : t('已启用');
  gcpDetail.value.vendor = t('谷歌云');
  // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
  gcpDetail.value.allowed = detail.value.allowed && detail.value.allowed.reduce((p, e) => {
    p.push(`${[e.protocol]}: ${e.port ? e.port.join(',') : '--'}`);
    return p;
  }, []);
  // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
  gcpDetail.value.denied = detail.value.denied && detail.value.denied.reduce((p, e) => {
    p.push(`${[e.protocol]}: ${e.port ? e.port.join(',') : '--'}`);
    return p;
  }, []);
  gcpDetail.value.ports = gcpDetail.value.operate === t('允许') ? gcpDetail.value.allowed : gcpDetail.value.denied;
  console.log('gcpDetail.value.ports', gcpDetail.value.ports);
  // eslint-disable-next-line max-len
  gcpDetail.value.target = [...detail?.destination_ranges, ...detail?.target_service_accounts, ...detail?.target_tags].length
    ? [...detail?.destination_ranges, ...detail?.target_service_accounts, ...detail?.target_tags] : '--';
  gcpDetail.value.source = [...detail?.source_ranges, ...detail?.source_service_accounts, ...detail?.source_tags].length
    ? [...detail?.source_ranges, ...detail?.source_service_accounts, ...detail?.source_tags] : '--';
};

// const tabs = [
//   {
//     name: t('关联实例'),
//     value: 'relate',
//   },
// ];

const isShowGcpAdd = ref(false);
const gcpTitle = ref<string>(t('新增'));
const isAdd = ref(false);
const isLoading = ref(false);

// const handleGcpAdd = (add: boolean) => {
//   gcpTitle.value = add ? t('新增') : t('修改');
//   isShowGcpAdd.value = true;
//   isAdd.value = add;
//   isLoading.value = false;
// };


// 新增修改防火墙规则
const submit = async (data: any) => {
  const fetchType = data?.id ? 'vendors/gcp/firewalls/rules' : 'vendors/gcp/firewalls/rules/create';
  const {
    loading,
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
    await updateData();
    fetchDetail();
  }
  isLoading.value = loading;
  isShowGcpAdd.value = false;
};

// const isBindBusiness = computed(() => {
//   return detail.value.bk_biz_id !== -1 && isResourcePage.value;
// });

// // 权限弹窗 bus通知最外层弹出
// const showAuthDialog = (authActionName: string) => {
//   bus.$emit('auth', authActionName);
// };
</script>

<template>
  <detail-header>
    {{ t('GCP防火墙') }}：ID（{{ `${id}` }}）
    <template #right>
      <!-- <div @click="showAuthDialog(actionName)">
        <bk-button
          :disabled="isBindBusiness || !authVerifyData?.permissionAction[actionName]"
          class="w100 ml10"
          theme="primary"
          @click="handleGcpAdd(false)"
        >
          {{ t('修改') }}
        </bk-button>
      </div> -->
      <!-- <bk-button
      class="w100 ml10"
      theme="primary"
    >
      {{ t('删除') }}
    </bk-button> -->
    </template>
  </detail-header>
  <!-- <detail-info
  :fields="gcpFields"
  :detail="gcpDetail"
/> -->
  <div class="i-detail-tap-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
    <detail-tab :tabs="hostTabs">
      <template #default>
        <detail-info :fields="gcpFields" :detail="gcpDetail" />
      </template>
    </detail-tab>
  </div>
  <!-- <detail-tab
  :tabs="tabs"
>
  <gcp-relate></gcp-relate>
</detail-tab> -->
  <gcp-add v-model:is-show="isShowGcpAdd" :gcp-title="gcpTitle" :is-add="isAdd" :loading="isLoading" :detail="detail"
           @submit="submit"></gcp-add>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}

.w60 {
  width: 60px;
}

:deep(.detail-info-main .info-list-item .item-field) {
  width: 150px !important;
}
</style>
