<script setup lang="ts">
import { ref, computed, provide } from 'vue';
import { QueryRuleOPEnum } from '@/typings/common';
import HostManage from '@/views/business/host/host-manage.vue';
import VpcManage from '@/views/resource/resource-manage/children/manage/vpc-manage.vue';
import SubnetManage from '@/views/resource/resource-manage/children/manage/subnet-manage.vue';
import SecurityManage from '@/views/resource/resource-manage/children/manage/security-manage.vue';
import DriveManage from '@/views/resource/resource-manage/children/manage/drive-manage.vue';
import IpManage from '@/views/resource/resource-manage/children/manage/ip-manage.vue';
import RoutingManage from '@/views/resource/resource-manage/children/manage/routing-manage.vue';
import ImageManage from '@/views/resource/resource-manage/children/manage/image-manage.vue';
import NetworkInterfaceManage from '@/views/resource/resource-manage/children/manage/network-interface-manage.vue';
import recyclebinManage from '@/views/resource/recyclebin-manager/recyclebin-manager.vue';
import { useVerify } from '@/hooks';
import useAdd from '@/views/resource/resource-manage/hooks/use-add';
import GcpAdd from '@/views/resource/resource-manage/children/add/gcp-add';
// forms
import EipForm from './forms/eip/index.vue';
import subnetForm from './forms/subnet/index.vue';
import securityForm from './forms/security/index.vue';
import firewallForm from './forms/firewall';
import TemplateDialog from '@/views/resource/resource-manage/children/dialog/template-dialog';

import { useRoute, useRouter } from 'vue-router';

import { useAccountStore } from '@/store/account';
import { InfoBox } from 'bkui-vue';
import { AUTH_BIZ_CREATE_IAAS_RESOURCE } from '@/constants/auth-symbols';

const isShowSideSlider = ref(false);
const isShowGcpAdd = ref(false);
const componentRef = ref();
const securityType = ref('group');

const isTemplateDialogShow = ref(false);
const isTemplateDialogEdit = ref(false);
const templateDialogPayload = ref({});

// use hooks
const route = useRoute();
const router = useRouter();
const accountStore = useAccountStore();

const gcpTitle = ref<string>('新增');
const isAdd = ref(true);
const isLoading = ref(false);
const formDetail = ref({});
const isEdit = ref(false);

provide('securityType', securityType); // 将数据传入孙组件

// 用于判断 sideslider 中的表单数据是否改变
const isFormDataChanged = ref(false);

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
  recyclebin: recyclebinManage,
};
const formMap = {
  ip: EipForm,
  subnet: subnetForm,
  security: securityForm,
};

const renderComponent = computed(() => {
  return Object.keys(componentMap).reduce((acc, cur) => {
    if (route.path.includes(cur)) acc = componentMap[cur];
    return acc;
  }, {});
});

const renderForm = computed(() => {
  return Object.keys(formMap).reduce((acc, cur) => {
    if (route.path.includes(cur)) {
      if (cur === 'security') acc = securityType.value === 'gcp' ? firewallForm : securityForm;
      else acc = formMap[cur];
    }
    return acc;
  }, {});
});

const filter = computed(() => {
  if (renderComponent.value === HostManage) {
    return {
      op: 'and',
      rules: [
        {
          op: QueryRuleOPEnum.NEQ,
          field: 'recycle_status',
          value: 'recycling',
        },
      ],
    };
  }
  return { op: 'and', rules: [] };
});

const isResourcePage = computed(() => {
  // 资源下没有业务ID
  return !accountStore.bizs;
});

const handleAdd = () => {
  if (securityType.value === 'template' && renderComponent.value === SecurityManage) {
    isTemplateDialogShow.value = true;
    isTemplateDialogEdit.value = false;
    return;
  }
  const { bizs } = route.query;
  if (renderComponent.value === DriveManage) {
    router.push({
      path: '/business/service/service-apply/disk',
      query: { bizs },
    });
  } else if (renderComponent.value === HostManage) {
    router.push({
      path: '/business/service/service-apply/cvm',
      query: { bizs },
    });
  } else if (renderComponent.value === VpcManage) {
    router.push({
      path: '/business/service/service-apply/vpc',
      query: { bizs },
    });
  } else if (renderComponent.value === SubnetManage) {
    router.push({
      path: '/business/service/service-apply/subnet',
      query: { bizs },
    });
  } else {
    isEdit.value = false;
    isShowSideSlider.value = true;
    // 标记初始化
    isFormDataChanged.value = false;
  }
};

const handleCancel = () => {
  isShowSideSlider.value = false;
};

const handleEdit = (detail: any) => {
  isShowSideSlider.value = true;
  formDetail.value = detail;
  isEdit.value = true;
};

// 新增成功 刷新列表
const handleSuccess = () => {
  handleCancel();
  componentRef.value.fetchComponentsData();
};

const handleSecrityType = (val: string) => {
  securityType.value = val;
};

// 新增修改防火墙规则
const submit = async (data: any) => {
  const fetchType = 'vendors/gcp/firewalls/rules/create';
  const { addData, updateData } = useAdd(fetchType, data, data?.id);
  if (isAdd.value) {
    // 新增
    addData();
  } else {
    await updateData();
  }
  isLoading.value = false;
};

// const handleToPage = () => {
//   const isHostManagePage = route.path.includes('/business/host');
//   const isDriveManagePage = route.path.includes('/business/drive');
//   let destination = '';
//   if (isHostManagePage) destination = '/business/host/recyclebin/cvm';
//   if (isDriveManagePage) destination = '/business/drive/recyclebin/disk';
//   router.push({ path: destination });
// };

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

// 权限hook
const {
  showPermissionDialog,
  handlePermissionConfirm,
  handlePermissionDialog,
  handleAuth,
  permissionParams,
  authVerifyData,
} = useVerify();
const computedSecurityText = computed(() => {
  if (renderComponent.value !== SecurityManage) return '新建';
  switch (securityType.value) {
    case 'template':
      return '新建模板';
    case 'gcp':
      return '新建GCP防火墙规则';
    default:
      return '新建安全组';
  }
});
const handleEditTemplate = (payload: any) => {
  isTemplateDialogShow.value = true;
  isTemplateDialogEdit.value = true;
  templateDialogPayload.value = payload;
};
</script>

<template>
  <div
    class="business-manage-wrapper"
    :class="[
      route.path === '/business/host' ? 'is-host-page' : '',
      route.path === '/business/recyclebin' ? 'is-recycle-page' : '',
    ]"
  >
    <bk-loading class="common-card-wrap" :loading="!accountStore.bizs">
      <component
        v-if="accountStore.bizs"
        ref="componentRef"
        :is="renderComponent"
        :filter="filter"
        :is-resource-page="isResourcePage"
        :auth-verify-data="authVerifyData"
        @auth="(val: string) => {
          handleAuth(val)
        }"
        @handleSecrityType="handleSecrityType"
        @editTemplate="handleEditTemplate"
        @edit="handleEdit"
        v-model:isFormDataChanged="isFormDataChanged"
      >
        <span>
          <hcm-auth :sign="{ type: AUTH_BIZ_CREATE_IAAS_RESOURCE, relation: [accountStore.bizs] }" v-slot="{ noPerm }">
            <bk-button theme="primary" class="mw64 mr10" :disabled="noPerm" @click="handleAdd">
              {{
                renderComponent === DriveManage ||
                renderComponent === HostManage ||
                renderComponent === SubnetManage ||
                renderComponent === VpcManage
                  ? '申请'
                  : computedSecurityText
              }}
            </bk-button>
          </hcm-auth>
        </span>

        <template #recycleHistory>
          <!-- <bk-button class="f-right" theme="primary" @click="handleToPage">
            {{ '回收记录' }}
          </bk-button> -->
        </template>
      </component>
    </bk-loading>
    <bk-sideslider
      v-model:isShow="isShowSideSlider"
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
          :detail="formDetail"
          :is-edit="isEdit"
          v-model:isFormDataChanged="isFormDataChanged"
        ></component>
      </template>
    </bk-sideslider>
    <permission-dialog
      v-model:is-show="showPermissionDialog"
      :params="permissionParams"
      @cancel="handlePermissionDialog"
      @confirm="handlePermissionConfirm"
    ></permission-dialog>

    <gcp-add
      v-model:is-show="isShowGcpAdd"
      :gcp-title="gcpTitle"
      :is-add="isAdd"
      :loading="isLoading"
      :detail="{}"
      @submit="submit"
    ></gcp-add>

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
.business-manage-wrapper {
  padding: 24px;
  height: 100%;
  overflow-y: auto;

  .common-card-wrap {
    padding: 16px 24px;
    height: 100%;
    background-color: #fff;

    & > :deep(.bk-nested-loading) {
      height: 100%;
      .bk-table {
        margin-top: 16px;
        max-height: calc(100% - 48px);
      }
    }
  }

  &.is-host-page {
    padding-bottom: 0;
  }

  &.is-recycle-page .common-card-wrap {
    padding: 0;
    background-color: transparent;

    :deep(.recycle-manager-page) {
      height: 100%;
      .bk-tab {
        height: 100%;
      }
    }
  }
}
</style>

<style lang="scss">
.mw64 {
  min-width: 64px;
}
.mw88 {
  min-width: 88px;
}
</style>
