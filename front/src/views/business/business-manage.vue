<script setup lang="ts">
import {
  ref,
  computed,
} from 'vue';

import HostManage from '@/views/resource/resource-manage/children/manage/host-manage.vue';
import VpcManage from '@/views/resource/resource-manage/children/manage/vpc-manage.vue';
import SubnetManage from '@/views/resource/resource-manage/children/manage/subnet-manage.vue';
import SecurityManage from '@/views/resource/resource-manage/children/manage/security-manage.vue';
import DriveManage from '@/views/resource/resource-manage/children/manage/drive-manage.vue';
import IpManage from '@/views/resource/resource-manage/children/manage/ip-manage.vue';
import RoutingManage from '@/views/resource/resource-manage/children/manage/routing-manage.vue';
import ImageManage from '@/views/resource/resource-manage/children/manage/image-manage.vue';
import NetworkInterfaceManage from '@/views/resource/resource-manage/children/manage/network-interface-manage.vue';
// forms
import EipForm from './forms/eip/index.vue';
import subnetForm from './forms/subnet/index.vue';
import securityForm from './forms/security/index.vue';

import {
  useRoute,
} from 'vue-router';

import { useAccountStore } from '@/store/account';

const isShowSideSlider = ref(false);
const componentRef = ref();

// use hooks
const route = useRoute();
const accountStore = useAccountStore();

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
const formMap = {
  ip: EipForm,
  subnet: subnetForm,
  security: securityForm,
};

const filter = ref({ op: 'and', rules: [] });

const renderComponent = computed(() => {
  return Object.keys(componentMap).reduce((acc, cur) => {
    if (route.path.includes(cur)) acc = componentMap[cur];
    return acc;
  }, {});
});

const renderForm = computed(() => {
  return Object.keys(formMap).reduce((acc, cur) => {
    if (route.path.includes(cur)) acc = formMap[cur];
    return acc;
  }, {});
});

const handleAdd = () => {
  isShowSideSlider.value = true;
};

const handleCancel = () => {
  isShowSideSlider.value = false;
};

// 新增成功 刷新列表
const handleSuccess = () => {
  handleCancel();
  componentRef.value.fetchComponentsData();
};
</script>

<template>
  <section class="business-manage-wrapper">
    <bk-loading :loading="!accountStore.bizs">
      <component
        v-if="accountStore.bizs"
        ref="componentRef"
        :is="renderComponent"
        :filter="filter"
      >
        <bk-button theme="primary" class="new-button" @click="handleAdd">新增</bk-button>
      </component>
    </bk-loading>
  </section>
  <bk-sideslider
    v-model:isShow="isShowSideSlider"
    width="800"
    title="新增"
    quick-close
  >
    <template #default>
      <component :is="renderForm" :filter="filter" @cancel="handleCancel" @success="handleSuccess"></component>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.business-manage-wrapper {
  height: calc(100% - 20px);
  overflow-y: auto;
  background-color: #fff;
  padding: 20px;
}
.new-button {
  width: 100px;
}
</style>
