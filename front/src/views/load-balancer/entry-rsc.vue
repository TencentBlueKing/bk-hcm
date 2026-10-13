<script setup lang="ts">
import { computed, provide } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import type { FilterType } from '@/typings/resource';
import { LOAD_BALANCER_SUBTYPE, type LoadBalancerSubtype } from '@/views/load-balancer/constants';
import ResourceSubtypeSwitch from '@/views/load-balancer/children/resource-subtype-switch.vue';
import ExclusiveClusterIndex from '@/views/load-balancer/exclusive-cluster/index.vue';
import LoadBalancerManage from '@/views/resource/resource-manage/children/manage/load-balancer-manage.vue';

defineOptions({ name: 'LoadBalancerEntryRsc', inheritAttrs: false });

const props = defineProps<{
  filter?: FilterType;
  isResourcePage?: boolean;
}>();

provide('currentGlobalBusinessId', {});

const route = useRoute();
const router = useRouter();

const resolveSubtype = (val: unknown): LoadBalancerSubtype =>
  val === LOAD_BALANCER_SUBTYPE.EXCLUSIVE_CLUSTER ? LOAD_BALANCER_SUBTYPE.EXCLUSIVE_CLUSTER : LOAD_BALANCER_SUBTYPE.CLB;

// 以 query.subtype 为唯一来源：先落 URL 再渲染子页，避免子页先挂载拿到旧 query 多发一次列表请求
const activeType = computed<LoadBalancerSubtype>({
  get: () => resolveSubtype(route.query.subtype),
  set: (val) => {
    const nextSubtype =
      val === LOAD_BALANCER_SUBTYPE.EXCLUSIVE_CLUSTER ? LOAD_BALANCER_SUBTYPE.EXCLUSIVE_CLUSTER : undefined;
    if (route.query.subtype === nextSubtype) return;
    const query = { ...route.query, subtype: nextSubtype };
    if (!nextSubtype) delete query.subtype;
    delete query.filter;
    delete query.page;
    router.replace({ query });
  },
});
const isExclusiveCluster = computed(() => activeType.value === LOAD_BALANCER_SUBTYPE.EXCLUSIVE_CLUSTER);
</script>

<template>
  <ExclusiveClusterIndex v-if="isExclusiveCluster" :filter="props.filter">
    <template #toolbar-prefix>
      <ResourceSubtypeSwitch v-model="activeType" />
    </template>
  </ExclusiveClusterIndex>
  <LoadBalancerManage v-else :filter="props.filter" :is-resource-page="props.isResourcePage" v-bind="$attrs">
    <template #toolbar-prefix>
      <ResourceSubtypeSwitch v-model="activeType" />
    </template>
    <slot></slot>
  </LoadBalancerManage>
</template>
