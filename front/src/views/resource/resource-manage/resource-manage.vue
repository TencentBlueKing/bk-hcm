<script setup lang="ts">
import {
  ref,
  watch,
  computed,
  provide,
} from 'vue';

import HostManage from './children/manage/host-manage.vue';
import VpcManage from './children/manage/vpc-manage.vue';
import SubnetManage from './children/manage/subnet-manage.vue';
import SecurityManage from './children/manage/security-manage.vue';
import DriveManage from './children/manage/drive-manage.vue';
import IpManage from './children/manage/ip-manage.vue';
import RoutingManage from './children/manage/routing-manage.vue';
import ImageManage from './children/manage/image-manage.vue';
import NetworkInterfaceManage from './children/manage/network-interface-manage.vue';
import AccountSelector from '@/components/account-selector/index.vue';
import { DISTRIBUTE_STATUS_LIST } from '@/constants';
import { useDistributionStore } from '@/store/distribution';
import EipForm from '@/views/business/forms/eip/index.vue';
import subnetForm from '@/views/business/forms/subnet/index.vue';
import securityForm from '@/views/business/forms/security/index.vue';
import firewallForm from '@/views/business/forms/firewall';

import {
  RESOURCE_TYPES,
} from '@/common/constant';

import {
  useI18n,
} from 'vue-i18n';
import {
  useRouter,
  useRoute,
} from 'vue-router';
import useSteps from './hooks/use-steps';

import type {
  FilterType,
} from '@/typings/resource';

import {
  useAccountStore,
} from '@/store';

import { useVerify } from '@/hooks';

// use hooks
const {
  t,
} = useI18n();
const router = useRouter();
const route = useRoute();
const accountStore = useAccountStore();
const {
  isShowDistribution,
  handleDistribution,
  ResourceDistribution,
} = useSteps();

const isResourcePage = computed(() => {   // 资源下没有业务ID
  return !accountStore.bizs;
});

// 权限hook
const {
  showPermissionDialog,
  handlePermissionConfirm,
  handlePermissionDialog,
  handleAuth,
  permissionParams,
  authVerifyData,
} = useVerify();

// 搜索过滤相关数据
const filter = ref({ op: 'and', rules: [] });
const accountId = ref('');
const status = ref('');
const op = ref('eq');
const accountFilter = ref<FilterType>({ op: 'and', rules: [{ field: 'type', op: 'eq', value: 'resource' }] });
const isShowSideSlider = ref(false);
const componentRef = ref();
const securityType = ref('group');
const isEdit = ref(false);
const formDetail = ref({});

provide('securityType', securityType);

const formMap = {
  ip: EipForm,
  subnet: subnetForm,
  security: securityForm,
};

const renderForm = computed(() => {
  return Object.keys(formMap).reduce((acc, cur) => {
    if (route.query.type === cur) {
      if (cur === 'security' && securityType.value === 'gcp') acc = firewallForm;
      else acc = formMap[cur];
    };
    return acc;
  }, {});
});

// 组件map
const componentMap = {
  host: HostManage,
  vpc: VpcManage,
  subnet: SubnetManage,
  security: SecurityManage,
  drive: DriveManage,
  ip: IpManage,
  routing: RoutingManage,
  image: ImageManage,
  'network-interface': NetworkInterfaceManage,
};

// 标签相关数据
const tabs = RESOURCE_TYPES.map((type) => {
  return {
    name: type.type,
    type: t(type.name),
    component: componentMap[type.type],
  };
});
const activeTab = ref(route.query.type || tabs[0].type);

const filterData = (key: string, val: string | number) => {
  if (!filter.value.rules.length) {
    if (val === 1) {    // 已分配标志
      op.value = 'neq';
    }
    filter.value.rules.push({
      field: key, op: op.value, value: -1,
    });
  } else {
    filter.value.rules.forEach((e: any) => {
      console.log(e.field, key, e.field === key);
      if (e.field === key) {
        e.op = val === 1 ? 'neq' : 'eq';
        return;
      }
      if (filter.value.rules.length === 2) return;
      if (val === 1) {    // 已分配标志
        op.value = 'neq';
      }
      filter.value.rules.push({
        field: key, op: op.value, value: -1,
      });
    });
  }
};

const handleAdd = () => {
  // ['host', 'vpc', 'drive', ||| 'security', 'subnet', 'ip']
  switch (activeTab.value) {
    case 'host':
      router.push({ path: '/resource/service-apply/cvm' });
      break;
    case 'vpc':
      router.push({ path: '/resource/service-apply/vpc' });
      break;
    case 'drive':
      router.push({ path: '/resource/service-apply/disk' });
      break;
    default:
      isShowSideSlider.value = true;
  }
};

const handleTabChange = (val: 'group' | 'gcp') => {
  securityType.value = val;
};

// 搜索数据
watch(
  () => accountId.value,
  (val) => {
    if (val) {
      if (!filter.value.rules.length) {
        filter.value.rules.push({
          field: 'account_id', op: 'eq', value: val,
        });
      } else {
        filter.value.rules.forEach((e: any) => {
          if (e.field === 'account_id') {
            e.value = val;
          } else {
            if (filter.value.rules.length === 2) return;
            filter.value.rules.push({
              field: 'account_id', op: 'eq', value: val,
            });
          }
        });
      }
    } else {
      filter.value.rules = filter.value.rules.filter((e: any) => e.field !== 'account_id');
    }
    useDistributionStore().setCloudAccountId(val);
  },
  {
    immediate: true,
  },
);

watch(
  () => status.value,
  (val) => {
    if (val === 'all' || !val) {
      filter.value.rules = filter.value.rules.filter((e: any) => e.field !== 'bk_biz_id');
    } else {
      filterData('bk_biz_id', val);
    }
  },
);

// 状态保持
watch(
  activeTab,
  () => {
    router.replace({
      query: {
        type: activeTab.value,
      },
    });
  },
);

const getResourceAccountList = async () => {
  try {
    const params = {
      filter: accountFilter.value,
      page: {
        start: 0,
        limit: 100,
      },
    };
    const res = await accountStore.getAccountList(params);
    accountStore.updateAccountList(res?.data?.details); // 账号数据   用于筛选
  } catch (error) {

  }
};

const handleCancel = () => {
  isShowSideSlider.value = false;
  isEdit.value = false;
};

// 新增成功 刷新列表
const handleSuccess = () => {
  handleCancel();
  if (Array.isArray(componentRef.value)) componentRef.value[0].fetchComponentsData();
  else componentRef.value.fetchComponentsData();
};

const handleEdit = (detail: any) => {
  formDetail.value = detail;
  isEdit.value = true;
  isShowSideSlider.value = true;
};

getResourceAccountList();


</script>

<template>
  <div>
    <section class="flex-center resource-header">
      <section class="flex-center" v-if="activeTab !== 'image'">
        <div class="mr10">{{t('云账号')}}</div>
        <div class="mr20">
          <account-selector
            :is-resource-page="isResourcePage"
            :filter="accountFilter"
            v-model="accountId"
          />
        </div>
      </section>
      <section class="flex-center" v-if="activeTab !== 'image'">
        <div class="mr10">{{t('分配状态')}}</div>
        <div class="mr20">
          <bk-select
            v-model="status"
          >
            <bk-option
              v-for="(item, index) in DISTRIBUTE_STATUS_LIST"
              :key="index"
              :value="item.value"
              :label="item.label"
            />
          </bk-select>
        </div>
      </section>
      <section class="flex-center">
        <bk-button
          theme="primary"
          class="ml10"
          @click="handleDistribution"
        >
          {{ t('快速分配') }}
        </bk-button>
      </section>
    <!-- <section class="flex-center">
      <bk-checkbox
        v-model="isAccurate"
      >
        {{ t('精确') }}
      </bk-checkbox>
      <bk-search-select
        class="search-filter ml10"
        clearable
        :data="searchData"
        v-model="searchValue"
      />
    </section> -->
    </section>
    <bk-tab
      v-model:active="activeTab"
      type="card"
      class="resource-main g-scroller"
    >
      <bk-tab-panel
        v-for="item in tabs"
        :key="item.name"
        :name="item.name"
        :label="item.type"
      >
        <component
          v-if="item.name === activeTab"
          :is="item.component"
          :filter="filter"
          :where-am-i="activeTab"
          :is-resource-page="isResourcePage"
          :auth-verify-data="authVerifyData"
          @auth="(val: string) => {
            handleAuth(val)
          }"
          @tabchange="handleTabChange"
          ref="componentRef"
          @edit="handleEdit"
        >
          <span
            @click="handleAuth('biz_iaas_resource_create')"
            v-if="['host', 'vpc', 'drive', 'security', 'subnet', 'ip'].includes(activeTab)"
          >
            <bk-button
              theme="primary"
              class="new-button"
              :disabled="!authVerifyData?.permissionAction?.biz_iaas_resource_create"
              @click="handleAdd">
              新建
            </bk-button>
          </span>
        </component>
      </bk-tab-panel>
    </bk-tab>

    <bk-sideslider
      v-model:isShow="isShowSideSlider"
      width="800"
      title="新增"
      quick-close
    >
      <template #default>
        <component
          :is="renderForm"
          :filter="filter"
          @cancel="handleCancel"
          @success="handleSuccess"
          :is-edit="isEdit"
          :detail="formDetail"></component>
      </template>
    </bk-sideslider>

    <resource-distribution
      v-model:is-show="isShowDistribution"
      :choose-resource-type="true"
      :title="t('快速分配')"
      :data="[]"
    />

    <permission-dialog
      v-model:is-show="showPermissionDialog"
      :params="permissionParams"
      @cancel="handlePermissionDialog"
      @confirm="handlePermissionConfirm"
    ></permission-dialog>
  </div>
</template>

<style lang="scss" scoped>
.flex-center {
  display: flex;
  align-items: center;
}
.resource-header {
  background: #fff;
  box-shadow: 1px 2px 3px 0 rgb(0 0 0 / 5%);
  padding: 20px;
}
.resource-main {
  margin-top: 20px;
  background: #fff;
  box-shadow: 1px 2px 3px 0 rgb(0 0 0 / 5%);
  height: calc(100vh - 270px);
  :deep(.bk-tab-content) {
    border-left: 1px solid #dcdee5;;
    border-right: 1px solid #dcdee5;;
    border-bottom: 1px solid #dcdee5;;
    padding: 20px;
  }
}
.search-filter {
  width: 500px;
}
.new-button {
  width: 100px;
}
</style>
