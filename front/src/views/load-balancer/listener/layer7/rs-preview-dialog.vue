<script setup lang="ts">
import { computed, inject, onMounted, Ref, ref } from 'vue';
import { ITargetGroupDetails, useLoadBalancerTargetGroupStore } from '@/store/load-balancer/target-group';

import Search from '../../children/search/search.vue';
import RsPreviewTable from '../children/rs-preview-table.vue';
import { ConditionKeyType, SearchConditionFactory } from '../../children/search/condition-factory';
import { ISearchSelectValue } from '@/typings';
import { getLocalFilterFnBySearchSelect } from '@/utils/search';

interface IProps {
  targetGroupId: string;
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();

const loadBalancerTargetGroupStore = useLoadBalancerTargetGroupStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const conditionProperties = SearchConditionFactory.createModel(ConditionKeyType.RS).getProperties();

const searchValue = ref<ISearchSelectValue>([]);
const list = ref<ITargetGroupDetails['target_list']>([]);
const displayList = computed(() => {
  return list.value.filter(getLocalFilterFnBySearchSelect(searchValue.value));
});

onMounted(async () => {
  const details = await loadBalancerTargetGroupStore.getTargetGroupDetails(
    props.targetGroupId,
    currentGlobalBusinessId.value,
  );
  list.value = details.target_list;
});

const handleSearch = (v: ISearchSelectValue) => {
  searchValue.value = v;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" title="预览目标组RS信息" :width="960" dialog-type="show" class="rs-preview-dialog">
    <search class="mb16" :fields="conditionProperties" @search="handleSearch" />
    <RsPreviewTable :list="displayList" :loading="loadBalancerTargetGroupStore.targetGroupDetailsLoading" />
  </bk-dialog>
</template>

<style scoped lang="scss"></style>
