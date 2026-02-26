<script setup lang="ts">
import { ref, computed, provide } from 'vue';
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
import { DISTRIBUTE_STATUS_LIST } from '@/constants';
import EipForm from '@/views/business/forms/eip/index.vue';
import subnetForm from '@/views/business/forms/subnet/index.vue';
import securityForm from '@/views/business/forms/security/index.vue';
import firewallForm from '@/views/business/forms/firewall';
import TemplateDialog from '@/views/resource/resource-manage/children/dialog/template-dialog';
import { useRouter, useRoute } from 'vue-router';
import { RESOURCE_TYPES, VendorEnum } from '@/common/constant';
import { useI18n } from 'vue-i18n';
import { useAccountSelectorStore } from '@/store/account-selector';
import { InfoBox } from 'bkui-vue';
import { AUTH_CREATE_IAAS_RESOURCE } from '@/constants/auth-symbols';

import {
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_VPC_APPLY,
  MENU_RESOURCE_RESOURCE_LIST,
} from '@/constants/menu-symbol';

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const accountSelectorStore = useAccountSelectorStore();

// 从 route 获取 accountId、vendor
const accountId = computed(() => (route.query.accountId as string) || '');
const queryVendor = computed(() => (route.query.vendor as VendorEnum) || null);

// currentVendor: 优先 query.vendor，否则从选中账号获取
const currentVendor = computed(() => {
  if (queryVendor.value) return queryVendor.value;
  if (accountId.value) {
    const account = accountSelectorStore.authorizedResourceAccountList.find(
      (a: { id: string }) => a.id === accountId.value,
    );
    return account?.vendor || null;
  }
  return null;
});
const isOtherVendor = computed(() => currentVendor.value === VendorEnum.OTHER);

// 当前选中账号详情（用于权限 relation、sync_failed_reason 展示）
const currentAccountDetail = computed(() => {
  if (!accountId.value) return null;
  return (
    accountSelectorStore.authorizedResourceAccountList.find((a: { id: string }) => a.id === accountId.value) || null
  );
});

// 分配状态：通过 route.query.assign 驱动
const assign = computed({
  get: () => (route.query.assign as string) || 'all',
  set: (val: string | number) => {
    router.push({
      query: { ...route.query, assign: val === 'all' ? undefined : String(val) },
    });
  },
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

const formMap = {
  ip: EipForm,
  subnet: subnetForm,
  security: securityForm,
};

const renderForm = computed(() => {
  return Object.keys(formMap).reduce<Record<string, any>>((acc, cur) => {
    if (route.params.resourceType === cur) {
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
  const vendor = currentVendor.value;
  if (!vendor || vendor === VendorEnum.TCLOUD) {
    types = types.concat(specialTabTypes);
  }
  if (isOtherVendor.value) {
    types = commonTabTypes.slice(0, 1);
  }
  return RESOURCE_TYPES.filter(({ type }) => types.includes(type)).map(({ type, name }) => {
    return { name: type, type: t(name), component: componentMap[type] };
  });
});

const activeTab = computed(() => (route.params.resourceType as string) || 'host');

const handleActiveTabChange = (value: string) => {
  const { filter, ...rest } = route.query;
  router.push({
    name: MENU_RESOURCE_RESOURCE_LIST,
    params: { resourceType: value },
    query: rest,
  });
};

const handleAdd = () => {
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
      isFormDataChanged.value = false;
  }
};

const handleSecrityType = (val: 'group' | 'gcp' | 'template') => {
  securityType.value = val;
};

const handleCancel = () => {
  isShowSideSlider.value = false;
  isEdit.value = false;
};

const handleSuccess = () => {
  handleCancel();
  if (Array.isArray(componentRef.value)) componentRef.value[0].fetchComponentsData();
  else componentRef.value.fetchComponentsData();
};

const handleEdit = (detail: any) => {
  formDetail.value = detail;
  isEdit.value = true;
  isShowSideSlider.value = true;
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
</script>

<template>
  <div>
    <bk-alert
      theme="error"
      closable
      class="error-message-alert"
      v-if="currentAccountDetail?.sync_failed_reason?.length"
    >
      <template #title>
        {{ currentAccountDetail?.sync_failed_reason }}
      </template>
    </bk-alert>
    <bk-tab :active="activeTab" type="card-grid" class="resource-main g-scroller" @change="handleActiveTabChange">
      <template #setting>
        <div style="margin: 0 10px">
          <bk-select v-model="assign" :clearable="false" :filterable="false" class="w80">
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
            @handle-secrity-type="handleSecrityType"
            ref="componentRef"
            @edit="handleEdit"
          >
            <template
              v-if="['host', 'vpc', 'drive', 'security', 'subnet', 'ip', 'clb'].includes(activeTab) && !isOtherVendor"
            >
              <hcm-auth
                :sign="{ type: AUTH_CREATE_IAAS_RESOURCE, relation: [currentAccountDetail?.id] }"
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
.resource-main {
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

.w80 {
  width: 80px;
}

.error-message-alert {
  margin: -8px 0 16px;
}
</style>

<style lang="scss">
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
