<script setup lang="ts">
import { ref, computed, provide, watch, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { InfoBox } from 'bkui-vue';
// 共享资源管理组件（业务视图与资源视图共用）
import HostManage from '@/views/resource/resource-manage/children/manage/host-manage.vue';
import VpcManage from '@/views/resource/resource-manage/children/manage/vpc-manage.vue';
import SubnetManage from '@/views/resource/resource-manage/children/manage/subnet-manage.vue';
import SecurityManage from '@/views/resource/resource-manage/children/manage/security-manage.vue';
import DriveManage from '@/views/resource/resource-manage/children/manage/drive-manage.vue';
import IpManage from '@/views/resource/resource-manage/children/manage/ip-manage.vue';
import RoutingManage from '@/views/resource/resource-manage/children/manage/routing-manage.vue';
import ImageManage from '@/views/resource/resource-manage/children/manage/image-manage.vue';
import NetworkInterfaceManage from '@/views/resource/resource-manage/children/manage/network-interface-manage.vue';
import LoadBalancerManage from '@/views/load-balancer/entry-rsc.vue';
import CertManager from '@/views/business/cert-manager';
import EipForm from '@/views/business/forms/eip/index.vue';
import subnetForm from '@/views/business/forms/subnet/index.vue';
import securityForm from '@/views/business/forms/security/index.vue';
import firewallForm from '@/views/business/forms/firewall';
import TemplateDialog from '@/views/resource/resource-manage/children/dialog/template-dialog';
import { DISTRIBUTE_STATUS_LIST } from '@/constants';
import { RESOURCE_TYPES, VendorEnum } from '@/common/constant';
import type { FilterType } from '@/typings/resource';
import { useAccountStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { AUTH_CREATE_IAAS_RESOURCE } from '@/constants/auth-symbols';
import routerAction from '@/router/utils/action';
import {
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_VPC_APPLY,
  MENU_RESOURCE_RESOURCE_LIST,
} from '@/constants/menu-symbol';
import AccountSummary from './account/account-summary.vue';

const { t } = useI18n();
const route = useRoute();
const accountStore = useAccountStore();
const resourceAccountStore = useResourceAccountStore();

// 资源下没有业务 ID
const isResourcePage = computed(() => !accountStore.bizs);
const isOtherVendor = computed(() => resourceAccountStore.vendorInResourcePage === VendorEnum.OTHER);

// 列表筛选沿用现有机制：账号 / 云厂商来自云账号树写入的 store，分配状态为页内状态
const filter = ref({ op: 'and', rules: [] });
const accountId = ref((route.query.accountId as string) || '');
const status = ref('all');
const accountFilter = ref<FilterType>({
  op: 'and',
  rules: [{ field: 'type', op: 'eq', value: 'resource' }],
});

const isShowSideSlider = ref(false);
const componentRef = ref();
const securityType = ref('group');
const isEdit = ref(false);
const formDetail = ref({});
const isTemplateDialogShow = ref(false);
const isTemplateDialogEdit = ref(false);
const templateDialogPayload = ref({});

// 用于判断 sideslider 中的表单数据是否改变
const isFormDataChanged = ref(false);

provide('securityType', securityType);
provide('isOtherVendor', isOtherVendor);

const activeTab = computed(() => (route.params.resourceType as string) || 'host');

const formMap = {
  ip: EipForm,
  subnet: subnetForm,
  security: securityForm,
};

const renderForm = computed(() => {
  return Object.keys(formMap).reduce<Record<string, any>>((acc, cur) => {
    if (activeTab.value === cur) {
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

const handleActiveTabChange = (value: string) => {
  routerAction.redirect(
    {
      name: MENU_RESOURCE_RESOURCE_LIST,
      params: { resourceType: value },
      query: { accountId: accountId.value || undefined },
    },
    { replace: true },
  );
};

// 地址里的资源类型不存在或当前云厂商不支持时，回到第一个可用页签
watch(
  [tabs, activeTab],
  ([tabList, active]) => {
    if (!tabList.length || tabList.some((item) => item.name === active)) return;
    handleActiveTabChange(tabList[0].name);
  },
  { immediate: true },
);

const filterData = (key: string, val: string | number) => {
  // val === 1 表示已分配，需要 bk_biz_id !== -1，即 op = 'neq'
  // val === -1 表示未分配，需要 bk_biz_id === -1，即 op = 'eq'
  const currentOp = val === 1 ? 'neq' : 'eq';

  const existingRuleIndex = filter.value.rules.findIndex((e: any) => e.field === key);

  if (existingRuleIndex !== -1) {
    filter.value.rules[existingRuleIndex].op = currentOp;
  } else {
    filter.value.rules.push({
      field: key,
      op: currentOp,
      value: -1,
    });
  }
};

const applyRouteMap: Record<string, symbol> = {
  host: MENU_RESOURCE_HOST_APPLY,
  vpc: MENU_RESOURCE_VPC_APPLY,
  drive: MENU_RESOURCE_DISK_APPLY,
  subnet: MENU_RESOURCE_SUBNET_APPLY,
  clb: MENU_RESOURCE_LOAD_BALANCER_APPLY,
};

const handleAdd = () => {
  if (activeTab.value === 'security' && securityType.value === 'template') {
    isTemplateDialogShow.value = true;
    isTemplateDialogEdit.value = false;
    templateDialogPayload.value = {};
    return;
  }
  const applyRouteName = applyRouteMap[activeTab.value];
  if (applyRouteName) {
    routerAction.redirect({ name: applyRouteName, query: route.query }, { history: true });
    return;
  }
  isShowSideSlider.value = true;
  // 标记初始化
  isFormDataChanged.value = false;
};

const handleSecrityType = (val: 'group' | 'gcp' | 'template') => {
  securityType.value = val;
};

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

// 搜索栏「云账号」的选项来自 accountStore.accountList
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
    accountStore.updateAccountList(res?.data?.details);
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
  <div class="resource-list-page">
    <AccountSummary />
    <bk-alert
      theme="error"
      closable
      class="error-message-alert"
      v-if="resourceAccountStore.resourceAccount?.sync_failed_reason?.length"
    >
      <template #title>
        {{ resourceAccountStore.resourceAccount?.sync_failed_reason }}
      </template>
    </bk-alert>
    <bk-tab :active="activeTab" type="card-grid" class="resource-main" @change="handleActiveTabChange">
      <template #setting>
        <div class="assign-container">
          <bk-select v-model="status" :clearable="false" :filterable="false" class="assign-select">
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
      :handle-close="() => (isTemplateDialogShow = false)"
      :handle-success="
        () => {
          isTemplateDialogShow = false;
          handleSuccess();
        }
      "
    />
  </div>
</template>

<style lang="scss" scoped>
.resource-list-page {
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.error-message-alert {
  margin: 0 24px;
}

.resource-main {
  flex: 1;
  overflow: hidden;
  margin: 0 24px 24px;

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

.assign-container {
  display: flex;
  align-items: center;
  height: 100%;
}

.assign-select {
  width: 80px;
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
