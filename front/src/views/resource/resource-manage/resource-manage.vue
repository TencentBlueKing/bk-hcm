<script setup lang="ts">
import { ref, watch, computed, provide, onMounted } from 'vue';
import HostManage from './children/manage/host-manage.vue';
import VpcManage from './children/manage/vpc-manage.vue';
import SubnetManage from './children/manage/subnet-manage.vue';
import SecurityManage from './children/manage/security-manage.vue';
import DriveManage from './children/manage/drive-manage.vue';
import IpManage from './children/manage/ip-manage.vue';
import RoutingManage from './children/manage/routing-manage.vue';
import ImageManage from './children/manage/image-manage.vue';
import NetworkInterfaceManage from './children/manage/network-interface-manage.vue';
import LoadBalancerManage from '@/views/load-balancer/entry-rsc.vue';
import CertManager from '@/views/business/cert-manager';
// import AccountSelector from '@/components/account-selector/index.vue';
import { DISTRIBUTE_STATUS_LIST } from '@/constants';
import EipForm from '@/views/business/forms/eip/index.vue';
import subnetForm from '@/views/business/forms/subnet/index.vue';
import securityForm from '@/views/business/forms/security/index.vue';
import firewallForm from '@/views/business/forms/firewall';
import TemplateDialog from '@/views/resource/resource-manage/children/dialog/template-dialog';
import BkTab, { BkTabPanel } from 'bkui-vue/lib/tab';
import { RouterView, useRouter, useRoute } from 'vue-router';
import { RESOURCE_TYPES, RESOURCE_TABS, VendorEnum } from '@/common/constant';
import { useI18n } from 'vue-i18n';
import type { FilterType } from '@/typings/resource';
import { useAccountStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { InfoBox } from 'bkui-vue';
import { AUTH_CREATE_IAAS_RESOURCE } from '@/constants/auth-symbols';
import routeQuery from '@/router/utils/query';
import {
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_VPC_APPLY,
} from '@/constants/menu-symbol';

// use hooks
const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const accountStore = useAccountStore();

const resourceAccountStore = useResourceAccountStore();

const isResourcePage = computed(() => {
  // 资源下没有业务ID
  return !accountStore.bizs;
});

const isOtherVendor = computed(() => resourceAccountStore.vendorInResourcePage === VendorEnum.OTHER);

// 账号 extension 信息
const headerExtensionMap = computed(() => {
  const map = { firstLabel: '', firstField: '', secondLabel: '', secondField: '' };
  switch (resourceAccountStore.resourceAccount.vendor) {
    case VendorEnum.TCLOUD:
      Object.assign(map, {
        firstLabel: '主账号ID',
        firstField: 'cloud_main_account_id',
        secondLabel: '子账号ID',
        secondField: 'cloud_sub_account_id',
      });
      break;
    case VendorEnum.AWS:
      Object.assign(map, {
        firstLabel: '云账号ID',
        firstField: 'cloud_account_id',
        secondLabel: '云iam用户名',
        secondField: 'cloud_iam_username',
      });
      break;
    case VendorEnum.AZURE:
      Object.assign(map, {
        firstLabel: '云租户ID',
        firstField: 'cloud_tenant_id',
        secondLabel: '云订阅名称',
        secondField: 'cloud_subscription_name',
      });
      break;
    case VendorEnum.GCP:
      Object.assign(map, {
        firstLabel: '云项目ID',
        firstField: 'cloud_project_id',
        secondLabel: '云项目名称',
        secondField: 'cloud_project_name',
      });
      break;
    case VendorEnum.HUAWEI:
      Object.assign(map, {
        firstLabel: '子账号ID',
        firstField: 'cloud_sub_account_id',
        secondLabel: '云子账号名称',
        secondField: 'cloud_sub_account_name',
      });
      break;
  }
  return map;
});

// 搜索过滤相关数据
const filter = ref({ op: 'and', rules: [] });
const accountId = ref((route.query.accountId as string) || '');
const status = ref('all');
const op = ref('eq');
const accountFilter = ref<FilterType>({
  op: 'and',
  rules: [{ field: 'type', op: 'eq', value: 'resource' }],
});
const isShowSideSlider = ref(false);
const componentRef = ref();
const securityType = ref('group');
const isEdit = ref(false);
const formDetail = ref({});
const activeResourceTab = ref(RESOURCE_TABS[0].key);
const isTemplateDialogShow = ref(false);
const isTemplateDialogEdit = ref(false);
const templateDialogPayload = ref({});

provide('securityType', securityType);
provide('isOtherVendor', isOtherVendor);

const handleTabChange = (path: string) => {
  router.push({ path, query: { ...route.query } });
};

// 用于判断 sideslider 中的表单数据是否改变
const isFormDataChanged = ref(false);

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
    }
    return acc;
  }, {});
});

// 组件map
const componentMap: Record<string, any> = {
  host: HostManage,
  vpc: VpcManage,
  subnet: SubnetManage,
  security: SecurityManage,
  drive: DriveManage,
  ip: IpManage,
  routing: RoutingManage,
  image: ImageManage,
  'network-interface': NetworkInterfaceManage,
  clb: LoadBalancerManage,
  certs: CertManager,
};

// 获取账号详情失败时不会执行resourceAccountStore.setResourceAccount，则会导致计算无效
const topTabs = computed(() =>
  RESOURCE_TABS.filter(({ key }) => !(isOtherVendor.value && key === '/resource/resource/account')),
);

// 标签相关数据
const commonTabTypes = ['host', 'vpc', 'subnet', 'security', 'drive', 'ip', 'routing', 'image', 'network-interface'];
const specialTabTypes = ['clb', 'certs'];
const tabs = computed(() => {
  let types = commonTabTypes;
  // 未选云厂商或腾讯云，展示clb和证书管理tab
  const vendor = resourceAccountStore.vendorInResourcePage;
  if (!vendor || vendor === VendorEnum.TCLOUD) {
    types = types.concat(specialTabTypes);
  }
  // 其他云厂商只展示主机tab
  if (isOtherVendor.value) {
    types = commonTabTypes.slice(0, 1);
  }
  return RESOURCE_TYPES.filter(({ type }) => types.includes(type)).map(({ type, name }) => {
    return { name: type, type: t(name), component: componentMap[type] };
  });
});
const activeTab = ref((route.query.type as string) || tabs.value[0].type);
const handleActiveTabChange = (value: string) => {
  router.replace({ query: { type: value, accountId: accountId.value || undefined } });
};

const filterData = (key: string, val: string | number) => {
  if (!filter.value.rules.length) {
    if (val === 1) {
      // 已分配标志
      op.value = 'neq';
    }
    filter.value.rules.push({
      field: key,
      op: op.value,
      value: -1,
    });
  } else {
    filter.value.rules.forEach((e: any) => {
      if (e.field === key) {
        e.op = val === 1 ? 'neq' : 'eq';
        return;
      }
      if (filter.value.rules.length === 2) return;
      if (val === 1) {
        // 已分配标志
        op.value = 'neq';
      }
      filter.value.rules.push({
        field: key,
        op: op.value,
        value: -1,
      });
    });
  }
};

const handleAdd = () => {
  // ['host', 'vpc', 'drive', ||| 'security', 'subnet', 'ip']
  if (activeTab.value === 'security' && securityType.value === 'template') {
    isTemplateDialogShow.value = true;
    isTemplateDialogEdit.value = false;
    templateDialogPayload.value = {};
    return;
  }
  switch (activeTab.value) {
    case 'host':
      router.push({ name: MENU_RESOURCE_HOST_APPLY, query: route.query });
      break;
    case 'vpc':
      router.push({ name: MENU_RESOURCE_VPC_APPLY, query: route.query });
      break;
    case 'drive':
      router.push({ name: MENU_RESOURCE_DISK_APPLY, query: route.query });
      break;
    case 'subnet':
      router.push({ name: MENU_RESOURCE_SUBNET_APPLY, query: route.query });
      break;
    case 'clb':
      router.push({ name: MENU_RESOURCE_LOAD_BALANCER_APPLY, query: route.query });
      break;
    default:
      isShowSideSlider.value = true;
      // 标记初始化
      isFormDataChanged.value = false;
  }
};

const handleSecrityType = (val: 'group' | 'gcp' | 'template') => {
  securityType.value = val;
};

const handleRouteDone = () => {
  routeQuery.set('type', activeTab.value);
};

watch(
  () => route.path,
  (path) => {
    topTabs.value.forEach(({ key }) => {
      const reg = new RegExp(key);
      if (reg.test(path)) {
        activeResourceTab.value = key;
      }
    });
  },
  {
    immediate: true,
  },
);

// 搜索数据
watch(
  () => accountId.value,
  (val) => {
    if (val) {
      if (!filter.value.rules.length) {
        filter.value.rules.push({
          field: 'account_id',
          op: 'eq',
          value: val,
        });
      } else {
        filter.value.rules.forEach((e: any) => {
          if (e.field === 'account_id') {
            e.value = val;
          } else {
            if (filter.value.rules.length === 2) return;
            filter.value.rules.push({
              field: 'account_id',
              op: 'eq',
              value: val,
            });
          }
        });
      }
    } else {
      filter.value.rules = filter.value.rules.filter((e: any) => e.field !== 'account_id');
    }
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

// 选择账号时，会触发selectedAccountId重新计算，优先使用账号列表中已有的list数据，其次再使用details数据
// 在这里设置accountId会触发watch accountId改变filter.value
// 最后触发use-query-list中的triggerApi
watch(
  () => resourceAccountStore.selectedAccountId,
  (id: string) => (accountId.value = id),
);

// 选择账号或云厂商时，会设置currentVendor
watch(
  () => resourceAccountStore.currentVendor,
  (vendor: VendorEnum) => {
    if (vendor) {
      const vendorRuleIdx = filter.value.rules.findIndex((e: any) => e.field === 'vendor');
      if (vendorRuleIdx === -1) {
        filter.value.rules.push({
          field: 'vendor',
          op: 'eq',
          value: vendor,
        });
      } else {
        filter.value.rules[vendorRuleIdx].value = vendor;
      }
    } else {
      filter.value.rules = filter.value.rules.filter((e: any) => e.field !== 'vendor');
    }
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
  } catch (error) {}
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
  // 初始化标记
  isFormDataChanged.value = false;
};

const handleBeforeClose = () => {
  if (isFormDataChanged.value) {
    InfoBox({
      title: '请确认是否关闭侧栏？',
      subTitle: '关闭后，内容需要重新填写！',
      quickClose: false,
      onConfirm() {
        handleCancel();
      },
    });
  } else {
    handleCancel();
  }
};
const computedSecurityText = computed(() => {
  if (!['security'].includes(activeTab.value)) return '新建';
  switch (securityType.value) {
    case 'template':
      return '新建模板';
    case 'gcp':
      return '新建GCP防火墙规则';
    default:
      return '新建安全组';
  }
});
onMounted(() => {
  getResourceAccountList();
});
</script>

<template>
  <div>
    <div class="navigation-resource">
      <div class="card-layout">
        <p class="resource-title">
          <span class="main-account-name">
            {{ resourceAccountStore?.resourceAccount?.name || '全部账号' }}
          </span>
          <template v-if="resourceAccountStore?.resourceAccount?.extension && !isOtherVendor">
            <div class="extension">
              <span>
                {{ headerExtensionMap.firstLabel }}：
                <span class="info-text">
                  {{ resourceAccountStore.resourceAccount.extension[headerExtensionMap.firstField] }}
                </span>
              </span>
              <span>
                {{ headerExtensionMap.secondLabel }}：
                <span class="info-text">
                  {{ resourceAccountStore.resourceAccount.extension[headerExtensionMap.secondField] }}
                </span>
              </span>
            </div>
          </template>
        </p>
        <BkTab
          class="resource-tab-wrap ml15"
          type="unborder-card"
          v-model:active="activeResourceTab"
          @change="handleTabChange"
        >
          <BkTabPanel v-for="item of topTabs" :label="item.label" :key="item.key" :name="item.key" />
        </BkTab>
      </div>
    </div>

    <div v-if="activeResourceTab === '/resource/resource/'">
      <bk-alert
        theme="error"
        closable
        class="error-message-alert"
        v-if="resourceAccountStore?.resourceAccount?.sync_failed_reason?.length"
      >
        <template #title>
          {{ resourceAccountStore?.resourceAccount?.sync_failed_reason }}
        </template>
      </bk-alert>
      <bk-tab
        v-model:active="activeTab"
        type="card-grid"
        class="resource-main g-scroller"
        @change="handleActiveTabChange"
      >
        <template #setting>
          <div style="margin: 0 10px">
            <bk-select v-model="status" :clearable="false" :filterable="false" class="w80">
              <bk-option
                v-for="(item, index) in DISTRIBUTE_STATUS_LIST"
                :key="index"
                :value="item.value"
                :label="item.label"
              />
            </bk-select>
          </div>
        </template>
        <template v-for="item in tabs" :key="item.name">
          <bk-tab-panel :name="item.name" :label="item.type">
            <component
              v-if="item.name === activeTab"
              :is="item.component"
              :filter="filter"
              :where-am-i="activeTab"
              :is-resource-page="isResourcePage"
              @handle-secrity-type="handleSecrityType"
              @route-done="handleRouteDone"
              ref="componentRef"
              @edit="handleEdit"
              v-model:is-form-data-changed="isFormDataChanged"
            >
              <template
                v-if="['host', 'vpc', 'drive', 'security', 'subnet', 'ip', 'clb'].includes(activeTab) && !isOtherVendor"
              >
                <hcm-auth
                  :sign="{ type: AUTH_CREATE_IAAS_RESOURCE, relation: [resourceAccountStore.resourceAccount?.id] }"
                  v-slot="{ noPerm }"
                >
                  <bk-button theme="primary" class="mw64" :disabled="noPerm" @click="handleAdd">
                    {{ ['host', 'clb'].includes(activeTab) ? '购买' : computedSecurityText }}
                  </bk-button>
                </hcm-auth>
              </template>
            </component>
          </bk-tab-panel>
        </template>
      </bk-tab>

      <bk-sideslider
        v-model:is-show="isShowSideSlider"
        width="800"
        title="新增"
        quick-close
        :before-close="handleBeforeClose"
      >
        <template #default>
          <component
            :is="renderForm"
            :filter="filter"
            @cancel="handleCancel"
            @success="handleSuccess"
            :is-edit="isEdit"
            :detail="formDetail"
            :show="isShowSideSlider"
            @edit="handleEdit"
            v-model:is-form-data-changed="isFormDataChanged"
          ></component>
        </template>
      </bk-sideslider>

      <TemplateDialog
        :is-show="isTemplateDialogShow"
        :is-edit="isTemplateDialogEdit"
        :payload="templateDialogPayload"
        :handle-close="
          () => {
            isTemplateDialogShow = false;
          }
        "
        :handle-success="
          () => {
            isTemplateDialogShow = false;
            handleSuccess();
          }
        "
      />
    </div>

    <RouterView v-else></RouterView>
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
  // margin-top: 20px;
  box-shadow: 1px 2px 3px 0 rgb(0 0 0 / 5%);
  height: calc(100vh - 200px);

  :deep(.bk-tab-header) {
    line-height: normal !important;

    .bk-tab-header-item {
      padding: 0 24px;
      height: 42px;
    }
  }

  :deep(.bk-tab-content) {
    height: calc(100% - 42px);
    padding: 16px 24px;

    & > .bk-tab-panel > .bk-nested-loading {
      height: 100%;

      .bk-table {
        margin-top: 16px;
        max-height: calc(100% - 52px);
      }
    }
  }
}

.search-filter {
  width: 500px;
}

.new-button {
  width: 64px;
}

.w80 {
  width: 80px;
}

.navigation-resource {
  min-height: 88px;
  margin: -24px -24px 24px;
}

.card-layout {
  background: #fff;
  border-bottom: 1px solid #dcdee5;
}

.resource-title {
  font-family: MicrosoftYaHei;
  font-size: 16px;
  color: #313238;
  letter-spacing: 0;
  line-height: 24px;
  padding: 14px 0 9px 24px;
  display: flex;
  align-items: center;

  .extension {
    font-size: 14px;
    color: #63656e;

    & > span {
      margin-left: 20px;

      .info-text {
        color: #313238;
      }
    }
  }
}

.bk-tab-content {
  padding: 0 !important;
}

.error-message-alert {
  margin: -8px 0 16px;
}
</style>

<style lang="scss">
.delete-resource-infobox,
.recycle-resource-infobox {
  .bk-info-sub-title {
    word-break: break-all;
  }
}

.mw64 {
  min-width: 64px;
}

.mw88 {
  min-width: 88px;
}

.table-new-row td {
  background-color: #f2fff4 !important;
}
</style>
