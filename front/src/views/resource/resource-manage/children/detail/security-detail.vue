<script lang="ts" setup>
import { Message } from 'bkui-vue';
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import SecurityInfo from '../components/security/security-info.vue';
import SecurityRelate from '../components/security/security-relate/index.vue';
import SecurityRule from '../components/security/security-rule.vue';
import Confirm from '@/components/confirm';
import { useI18n } from 'vue-i18n';

import { watch, ref, reactive, computed, provide } from 'vue';

import { useRoute } from 'vue-router';
import useDetail from '../../hooks/use-detail';
import { QueryRuleOPEnum } from '@/typings';
import { useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { SecurityGroupManageType } from '@/constants/security-group';

const { t } = useI18n();

const route = useRoute();
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
const { whereAmI, getBizsId } = useWhereAmI();
const resoureStore = useResourceStore();

const { loading, detail, getDetail } = useDetail('security_groups', securityId.value as string);

const tabs = [
  {
    name: t('基本信息'),
    value: 'detail',
  },
  {
    name: t('安全组规则'),
    value: 'rule',
  },
  {
    name: t('关联实例'),
    value: 'relate',
  },
];

const handleTabsChange = (val: string) => {
  if (val === 'rule') getRelatedSecurityGroups(detail.value);
};

watch(
  () => detail.value,
  (val: { account_id: string; region: string }) => {
    getRelatedSecurityGroups(val);
    getTemplateData(val);
  },
);

const isAssigned = computed(() => detail.value.bk_biz_id !== -1);
// 资源下已分配，不可以update安全组
const hasEditScopeInResource = computed(() => whereAmI.value === Senarios.resource && !isAssigned.value);
// 业务下未分配、非业务管理、管理业务!==当前业务，不可以update安全组
const hasEditScopeInBusiness = computed(
  () =>
    whereAmI.value === Senarios.business &&
    isAssigned.value &&
    detail.value?.mgmt_type === SecurityGroupManageType.BIZ &&
    detail.value?.mgmt_biz_id === getBizsId(),
);
const operateTooltipsOption = computed(() => {
  const isResourcePage = whereAmI.value === Senarios.resource;
  const isBusinessPage = whereAmI.value === Senarios.business;
  const isPlatformManage = detail.value?.mgmt_type === SecurityGroupManageType.PLATFORM;
  const isCurrentBizManage = detail.value?.mgmt_biz_id === getBizsId();

  if (isResourcePage && isAssigned.value) {
    return { content: t('安全组已分配，请到业务下操作'), disabled: !(isResourcePage && isAssigned.value) };
  }
  if (isBusinessPage && isPlatformManage) {
    return {
      content: t('该安全组的管理类型为平台管理，不允许在业务下操作'),
      disabled: !(isBusinessPage && isPlatformManage),
    };
  }
  if (isBusinessPage && !isAssigned.value) {
    return {
      content: t('该安全组当前处于未分配状态，不允许在业务下进行管理配置安全组规则等操作'),
      disabled: !(isBusinessPage && !isAssigned.value),
    };
  }
  if (isBusinessPage && !isCurrentBizManage) {
    return { content: t('该安全组不在当前业务管理，不允许操作'), disabled: !(isBusinessPage && !isCurrentBizManage) };
  }
  return { disabled: true };
});

const getRelatedSecurityGroups = async (detail: { account_id: string; region: string }) => {
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
  const res = await resourceStore.getCommonList(
    {
      page: {
        count: false,
        start: 0,
        limit: 100,
      },
      filter,
    },
    url,
  );
  relatedSecurityGroups.value = res?.data?.details;
};

const getTemplateData = async (detail: { account_id: string }) => {
  const [ipListPromise, ipGroupListPromise, portListPromise, portGroupListPromise] = [
    'address',
    'address_group',
    'service',
    'service_group',
  ].map((type) =>
    resoureStore.getCommonList(
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
    ),
  );
  const res = await Promise.all([ipListPromise, ipGroupListPromise, portListPromise, portGroupListPromise]);
  templateData.ipList = res[0]?.data?.details;
  templateData.ipGroupList = res[1]?.data?.details;
  templateData.portList = res[2]?.data?.details;
  templateData.portGroupList = res[3]?.data?.details;
};

const handleSync = async () => {
  const { account_id, vendor, cloud_id, region, resource_group_name: resourceGroupName } = detail.value;
  const isAzureVendor = vendor === 'azure';
  Confirm(t('同步单个安全组'), t('从云上同步该安全组、安全组规则、关联的实例信息等'), async () => {
    await resourceStore.syncResource(vendor, account_id, 'security_group', {
      cloud_ids: [cloud_id],
      regions: isAzureVendor ? undefined : [region],
      resource_group_names: isAzureVendor ? [resourceGroupName] : undefined,
    });
    Message({ theme: 'success', message: t('已提交同步任务，请等待同步结果') });
  });
};

provide('isAssigned', isAssigned);
provide('hasEditScopeInResource', hasEditScopeInResource);
provide('hasEditScopeInBusiness', hasEditScopeInBusiness);
provide('operateTooltipsOption', operateTooltipsOption);
</script>

<template>
  <detail-header>
    {{ t('安全组') }}：ID（{{ `${securityId}` }}）
    <template #right>
      <bk-button @click="handleSync">{{ t('同步') }}</bk-button>
    </template>
  </detail-header>

  <div class="i-detail-tap-wrap" :style="whereAmI === Senarios.resource && 'padding: 0;'">
    <detail-tab :tabs="tabs" :active="activeTab" :on-change="handleTabsChange">
      <template #default="type">
        <security-info
          :id="securityId"
          :vendor="vendor"
          v-if="type === 'detail'"
          :loading="loading"
          :detail="detail"
          :get-detail="getDetail"
        />
        <security-rule
          v-else-if="type === 'rule'"
          :id="securityId"
          :vendor="vendor"
          :related-security-groups="relatedSecurityGroups"
          :template-data="templateData"
        />
        <security-relate v-else :detail="detail" />
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
