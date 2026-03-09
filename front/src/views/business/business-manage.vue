<script setup lang="ts">
import { ref, computed, provide, type Component } from 'vue';
import HostManage from '@/views/business/host/host-manage.vue';
import VpcManage from '@/views/resource/resource-manage/children/manage/vpc-manage.vue';
import SubnetManage from '@/views/resource/resource-manage/children/manage/subnet-manage.vue';
import SecurityManage from '@/views/resource/resource-manage/children/manage/security-manage.vue';
import DriveManage from '@/views/resource/resource-manage/children/manage/drive-manage.vue';
import IpManage from '@/views/resource/resource-manage/children/manage/ip-manage.vue';
import RoutingManage from '@/views/resource/resource-manage/children/manage/routing-manage.vue';
import ImageManage from '@/views/resource/resource-manage/children/manage/image-manage.vue';
import NetworkInterfaceManage from '@/views/resource/resource-manage/children/manage/network-interface-manage.vue';
import useAdd from '@/views/resource/resource-manage/hooks/use-add';
import GcpAdd from '@/views/resource/resource-manage/children/add/gcp-add';
// forms
import EipForm from './forms/eip/index.vue';
import subnetForm from './forms/subnet/index.vue';
import securityForm from './forms/security/index.vue';
import firewallForm from './forms/firewall';
import TemplateDialog from '@/views/resource/resource-manage/children/dialog/template-dialog';

import { useRoute } from 'vue-router';

import { InfoBox } from 'bkui-vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { AUTH_BIZ_CREATE_IAAS_RESOURCE } from '@/constants/auth-symbols';
import routerAction from '@/router/utils/action';
import {
  MENU_BUSINESS_HOST_MANAGEMENT,
  MENU_BUSINESS_DISK_MANAGEMENT,
  MENU_BUSINESS_VPC_MANAGEMENT,
  MENU_BUSINESS_SUBNET_MANAGEMENT,
  MENU_BUSINESS_EIP_MANAGEMENT,
  MENU_BUSINESS_IMAGE_MANAGEMENT,
  MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT,
  MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT,
  MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT,
  MENU_BUSINESS_APPLY_CVM,
  MENU_BUSINESS_APPLY_VPC,
  MENU_BUSINESS_APPLY_DISK,
  MENU_BUSINESS_APPLY_SUBNET,
} from '@/constants/menu-symbol';

const isShowSideSlider = ref(false);
const isShowGcpAdd = ref(false);
const componentRef = ref();
const securityType = ref('group');

const isTemplateDialogShow = ref(false);
const isTemplateDialogEdit = ref(false);
const templateDialogPayload = ref({});

// use hooks
const route = useRoute();
const { getBizsId } = useWhereAmI();

const gcpTitle = ref<string>('新增');
const isAdd = ref(true);
const isLoading = ref(false);
const formDetail = ref({});
const isEdit = ref(false);

provide('securityType', securityType); // 将数据传入孙组件

// 用于判断 sideslider 中的表单数据是否改变
const isFormDataChanged = ref(false);

const componentMap = new Map<symbol, Component>([
  [MENU_BUSINESS_HOST_MANAGEMENT, HostManage],
  [MENU_BUSINESS_DISK_MANAGEMENT, DriveManage],
  [MENU_BUSINESS_VPC_MANAGEMENT, VpcManage],
  [MENU_BUSINESS_SUBNET_MANAGEMENT, SubnetManage],
  [MENU_BUSINESS_EIP_MANAGEMENT, IpManage],
  [MENU_BUSINESS_IMAGE_MANAGEMENT, ImageManage],
  [MENU_BUSINESS_NETWORK_INTERFACE_MANAGEMENT, NetworkInterfaceManage],
  [MENU_BUSINESS_ROUTEING_TABLE_MANAGEMENT, RoutingManage],
  [MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT, SecurityManage],
]);

const formMap = new Map<symbol, Component>([
  [MENU_BUSINESS_EIP_MANAGEMENT, EipForm],
  [MENU_BUSINESS_SUBNET_MANAGEMENT, subnetForm],
  [MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT, securityForm],
]);

const activeKey = computed(() => route.name as symbol);

const renderComponent = computed(() => componentMap.get(activeKey.value));

const renderForm = computed(() => {
  if (activeKey.value === MENU_BUSINESS_SECURITY_GROUP_MANAGEMENT && securityType.value === 'gcp') {
    return firewallForm;
  }
  return formMap.get(activeKey.value);
});

const handleAdd = () => {
  if (securityType.value === 'template' && renderComponent.value === SecurityManage) {
    isTemplateDialogShow.value = true;
    isTemplateDialogEdit.value = false;
    return;
  }
  const applyRouteMap = new Map<Component, symbol>([
    [DriveManage, MENU_BUSINESS_APPLY_DISK],
    [HostManage, MENU_BUSINESS_APPLY_CVM],
    [VpcManage, MENU_BUSINESS_APPLY_VPC],
    [SubnetManage, MENU_BUSINESS_APPLY_SUBNET],
  ]);
  const applyRouteName = applyRouteMap.get(renderComponent.value!);
  if (applyRouteName) {
    routerAction.redirect({ name: applyRouteName }, { history: true });
    return;
  }
  isEdit.value = false;
  isShowSideSlider.value = true;
  isFormDataChanged.value = false;
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
  <div class="business-manage-wrapper">
    <div class="common-card-wrap">
      <component
        ref="componentRef"
        :is="renderComponent"
        @handle-secrity-type="handleSecrityType"
        @edit-template="handleEditTemplate"
        @edit="handleEdit"
      >
        <span>
          <hcm-auth :sign="{ type: AUTH_BIZ_CREATE_IAAS_RESOURCE, relation: [getBizsId()] }" v-slot="{ noPerm }">
            <bk-button theme="primary" class="mw64" :disabled="noPerm" @click="handleAdd">
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
      </component>
    </div>
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
          :detail="formDetail"
          :is-edit="isEdit"
          :show="isShowSideSlider"
          v-model:is-form-data-changed="isFormDataChanged"
        ></component>
      </template>
    </bk-sideslider>

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
