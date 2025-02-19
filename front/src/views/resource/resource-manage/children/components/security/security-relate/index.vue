<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue';
import {
  useSecurityGroupStore,
  type ISecurityGroupRelResCountItem,
  type ISecurityGroupDetail,
  type ISecurityGroupRelBusiness,
} from '@/store/security-group';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { SecurityGroupManageType } from './typings';

import platform from './platform.vue';
import business from './business.vue';

const props = defineProps<{ detail: ISecurityGroupDetail }>();

const { whereAmI, getBizsId } = useWhereAmI();
const securityGroupStore = useSecurityGroupStore();

const manageType = computed<SecurityGroupManageType>(() => {
  let type = SecurityGroupManageType.platform;
  // 业务管理：业务下，安全组的管理业务id与全局业务id相同
  if (whereAmI.value === Senarios.business && props.detail?.mgmt_biz_id === getBizsId()) {
    type = SecurityGroupManageType.business;
  }
  return type;
});
const comps: Record<SecurityGroupManageType, any> = { platform, business };

const relatedResourcesCountList = ref<ISecurityGroupRelResCountItem[]>([]);
const relatedBiz = ref<ISecurityGroupRelBusiness>(null);

const isLoading = ref(false);
onBeforeMount(async () => {
  if (props.detail) {
    const { id } = props.detail;
    isLoading.value = true;
    try {
      relatedResourcesCountList.value = await securityGroupStore.queryRelatedResourcesCount([id]);
      relatedBiz.value = await securityGroupStore.queryRelBusiness(id);
    } finally {
      isLoading.value = false;
    }
  }
});
</script>

<template>
  <bk-loading loading v-if="isLoading">
    <div style="width: 100%; height: 360px" />
  </bk-loading>
  <div v-else class="security-relate-page">
    <component
      :is="comps[manageType]"
      :detail="props.detail"
      :related-resources-count-list="relatedResourcesCountList"
      :related-biz="relatedBiz"
    />
  </div>
</template>

<style scoped lang="scss"></style>
