<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import SecurityInfo from '../components/security/security-info.vue';
import SecurityRelate from '../components/security/security-relate.vue';
import SecurityRule from '../components/security/security-rule.vue';
import {
  useI18n,
} from 'vue-i18n';

import { watch, ref, reactive } from 'vue';


import {
  useRoute,
} from 'vue-router';
import useDetail from '../../hooks/use-detail';
import { QueryRuleOPEnum } from '@/typings';
import { useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const {
  t,
} = useI18n();

const route = useRoute();
const filter = ref({ op: 'and', rules: [{ field: 'type', op: 'eq', value: 'ingress' }] });
const activeTab = ref(route.query?.activeTab);
const securityId = ref(route.query?.id);
const vendor = ref(route.query?.vendor);
const resourceStore = useResourceStore();
const relatedSecurityGroups = ref([]);
const templateData = reactive({
  ipList: [],
  ipGroupList: [],
  portList: [],
  portGroupList: [],
});
const { whereAmI } = useWhereAmI();
const resoureStore = useResourceStore();

const {
  loading,
  detail,
  getDetail,
} = useDetail(
  'security_groups',
  securityId.value as string,
);

const tabs = [
  {
    name: t('基本信息'),
    value: 'detail',
  },
  // {
  //   name: t('关联实例'),
  //   value: 'relate',
  // },
  {
    name: t('安全组规则'),
    value: 'rule',
  },
];

const handleTabsChange = (val: string) => {
  if (val === 'rule') getRelatedSecurityGroups(detail.value);
};

watch(
  () => detail.value,
  (val: { account_id: string; region: string; }) => {
    getRelatedSecurityGroups(val);
    getTemplateData(val);
  },
);

const getRelatedSecurityGroups = async (detail: { account_id: string; region: string; }) => {
  const url = 'security_groups/list';
  const filter = {
    op: QueryRuleOPEnum.AND,
    rules: [
      {
        field: 'account_id',
        op: QueryRuleOPEnum.CS,
        value: detail.account_id,
      },
      {
        field: 'region',
        op: QueryRuleOPEnum.CS,
        value: detail.region,
      },
    ],
  };
  const res = await resourceStore.getCommonList({
    page: {
      count: false,
      start: 0,
      limit: 100,
    },
    filter,
  }, url);
  relatedSecurityGroups.value = res?.data?.details;
};

const getTemplateData = async (detail: { account_id: string;}) => {
  const [
    ipListPromise,
    ipGroupListPromise,
    portListPromise,
    portGroupListPromise,
  ] = [
    'address',
    'address_group',
    'service',
    'service_group',
  ].map(type => resoureStore.getCommonList(
    {
      filter: {
        op: 'and',
        rules: [
          {
            field: 'vendor',
            op: 'eq',
            value: 'tcloud',
          },
          {
            field: 'account_id',
            op: QueryRuleOPEnum.CS,
            value: detail.account_id,
          },
          {
            field: 'type',
            op: 'eq',
            value: type,
          },
        ],
      },
      page: {
        start: 0,
        limit: 500,
      },
    },
    'argument_templates/list',
  ));
  const res = await Promise.all([
    ipListPromise,
    ipGroupListPromise,
    portListPromise,
    portGroupListPromise,
  ]);
  templateData.ipList = res[0]?.data?.details;
  templateData.ipGroupList = res[1]?.data?.details;
  templateData.portList = res[2]?.data?.details;
  templateData.portGroupList = res[3]?.data?.details;
};

</script>

<template>
  <detail-header>
    {{ t('安全组') }}：ID（{{ `${securityId}` }}）
  </detail-header>

  <div class="i-detail-tap-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
    <detail-tab :tabs="tabs" :active="activeTab" :on-change="handleTabsChange">
      <template #default="type">
        <security-info :id="securityId" :vendor="vendor" v-if="type === 'detail'" :loading="loading" :detail="detail"
                       :get-detail="getDetail" />
        <security-relate v-if="type === 'relate'" />
        <security-rule v-if="type === 'rule'" :filter="filter" :id="securityId" :vendor="vendor"
                       :related-security-groups="relatedSecurityGroups" :template-data="templateData" />
      </template>
    </detail-tab>
  </div>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}

.w60 {
  width: 60px;
}
</style>
