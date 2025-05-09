<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue';
import {
  useSecurityGroupStore,
  type ISecurityGroupRelResCountItem,
  type ISecurityGroupDetail,
  type ISecurityGroupRelBusiness,
} from '@/store/security-group';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { SecurityGroupManageType } from '@/constants/security-group';

import platform from './platform.vue';
import biz from './biz.vue';

const props = defineProps<{ detail: ISecurityGroupDetail }>();

const { whereAmI, getBizsId } = useWhereAmI();
const securityGroupStore = useSecurityGroupStore();

// 1、账号下，只展示 platform 组件，展示所有业务的 related_resources。安全组未分配的情况下，可以绑定、解绑关联实例。
// 2、业务下，分配业务 === 当前业务：展示 biz 组件，展示所有业务的 rel_res，其他业务 rel_res 只读，未分配 rel_res 只读。
// 3、业务下，分配业务 !== 当前业务：展示 platform 组件，只展示当前业务的 rel_res，支持绑定、解绑 rel_res。
const viewType = computed<SecurityGroupManageType>(() => {
  let type = SecurityGroupManageType.PLATFORM;
  // 业务管理：业务下，安全组的分配业务id与当前业务id相同
  if (whereAmI.value === Senarios.business && props.detail?.bk_biz_id === getBizsId()) {
    type = SecurityGroupManageType.BIZ;
  }
  return type;
});
const comps: Record<SecurityGroupManageType, any> = {
  [SecurityGroupManageType.PLATFORM]: platform,
  [SecurityGroupManageType.BIZ]: biz,
  [SecurityGroupManageType.UNKNOWN]: null,
};

const relatedResourcesCountList = ref<ISecurityGroupRelResCountItem[]>([]);
const relatedBiz = ref<ISecurityGroupRelBusiness>(null);

const getRelatedInfo = () => {
  const { id } = props.detail;
  if (whereAmI.value === Senarios.business) {
    // 业务下，关联资源list请求前置接口
    securityGroupStore.queryRelBusiness(id).then((data) => (relatedBiz.value = data));
  }
  securityGroupStore.queryRelatedResourcesCount([id]).then((data) => (relatedResourcesCountList.value = data));
};
onBeforeMount(async () => {
  if (props.detail) {
    getRelatedInfo();
  }
});
</script>

<template>
  <div class="security-relate-page">
    <component
      :is="comps[viewType]"
      :detail="props.detail"
      :related-resources-count-list="relatedResourcesCountList"
      :related-biz="relatedBiz"
      :get-related-info="getRelatedInfo"
      :rel-biz-loading="securityGroupStore.isQueryRelBusinessLoading"
    />
  </div>
</template>

<style scoped lang="scss"></style>
