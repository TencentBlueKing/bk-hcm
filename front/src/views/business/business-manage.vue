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
import EipTcloudForm from './forms/eip/tcloud.vue'

import {
  useRoute,
} from 'vue-router';

const isShowSideSlider = ref(false);

// use hooks
const route = useRoute();

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
  ip: EipTcloudForm
}

const filter = ref({ op: 'and', rules: [] });

const renderComponent = computed(() => {
  return Object.keys(componentMap).reduce((acc, cur) => {
    if (route.path.includes(cur)) acc = componentMap[cur]
    return acc
  }, {})
})

const renderForm = computed(() => {
  return Object.keys(formMap).reduce((acc, cur) => {
    if (route.path.includes(cur)) acc = formMap[cur]
    return acc
  }, {})
})

const handleAdd = () => {
  isShowSideSlider.value = true
}
</script>

<template>
  <section class="business-manage-wrapper">
    <component
      :is="renderComponent"
      :filter="filter"
    >
      <bk-button theme="primary" class="new-button" @click="handleAdd">新增</bk-button>
    </component>
  </section>
  <bk-sideslider
    v-model:isShow="isShowSideSlider"
    width="800"
    title="新增"
    quick-close
  >
    <template #default>
      <component :is="renderForm"></component>
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
