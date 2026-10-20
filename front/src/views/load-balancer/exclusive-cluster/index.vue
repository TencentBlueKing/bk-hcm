<script setup lang="ts">
import { computed, ref, useTemplateRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import { ModelPropertyColumn } from '@/model/typings';
import { FilterType } from '@/typings/resource';
import { ISearchSelectValue, QueryFilterType, QueryRuleOPEnum } from '@/typings';
import { getModel } from '@/model/manager';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import routeQuery from '@/router/utils/query';
import {
  buildSearchSelectValueBySearchQsCondition,
  getSimpleConditionBySearchSelect,
  transformSimpleCondition,
} from '@/utils/search';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import { IExclusiveClusterItem, useLoadBalancerExclusiveClusterStore } from '@/store/load-balancer/exclusive-cluster';
import { storeToRefs } from 'pinia';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import DataList from './data-list.vue';
import { DisplayFieldExclusiveCluster } from './column';
import { SearchConditionExclusiveCluster } from './search-condition';

defineOptions({ name: 'ExclusiveClusterList' });

const props = defineProps<{
  filter?: FilterType;
}>();

const route = useRoute();
const exclusiveClusterStore = useLoadBalancerExclusiveClusterStore();
const { selectedAccountId, vendorInResourcePage } = storeToRefs(useResourceAccountStore());
const columnModel = getModel(DisplayFieldExclusiveCluster);
const columns = computed<ModelPropertyColumn[]>(() => columnModel.getProperties());
const searchFields = getModel(SearchConditionExclusiveCluster).getProperties();
const searchQs = useSearchQs({ key: 'filter', properties: searchFields });
const { pagination, getPageParams, handlePageChange, handlePageSizeChange, handleSort } = usePage();

const searchValue = ref<ISearchSelectValue>([]);
const condition = ref<Record<string, any>>({});
const sortParams = ref<{ sort?: string; order?: string }>({});
const listData = ref<IExclusiveClusterItem[]>([]);
const selections = ref<IExclusiveClusterItem[]>([]);
const batchDistributionRef = useTemplateRef<{ open: (rows: IExclusiveClusterItem[]) => void }>('batch-distribution');

const parentRules = computed(() => {
  const rules = [...(props.filter?.rules || [])];
  if (vendorInResourcePage.value && !rules.some((rule) => rule.field === 'vendor')) {
    rules.push({ field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendorInResourcePage.value });
  }
  if (selectedAccountId.value && !rules.some((rule) => rule.field === 'account_id')) {
    rules.push({ field: 'account_id', op: QueryRuleOPEnum.EQ, value: selectedAccountId.value });
  }
  return rules;
});

const handleSearch = (val: ISearchSelectValue) => {
  searchQs.set(getSimpleConditionBySearchSelect(val) || {});
};

const fetchList = async () => {
  try {
    const searchFilter = transformSimpleCondition(condition.value, searchFields) as QueryFilterType;
    const listFilter: QueryFilterType = {
      op: 'and',
      rules: [...parentRules.value, ...(searchFilter.rules || [])],
    };
    const { list, count } = await exclusiveClusterStore.getExclusiveClusterList({
      filter: listFilter,
      page: getPageParams(pagination, sortParams.value),
    });
    listData.value = list;
    pagination.count = count;
  } catch {
    listData.value = [];
    pagination.count = 0;
  }
};

const handleSelectionChange = (selection: IExclusiveClusterItem[]) => {
  selections.value = selection;
};

const handleAssign = (row: IExclusiveClusterItem) => {
  batchDistributionRef.value?.open([row]);
};

const handleRetry = () => {
  routeQuery.refresh();
};

const handleAssignSubmit = async (ids: string[], bkBizId: number) => {
  await exclusiveClusterStore.assignExclusiveClusterToBiz(ids, bkBizId);
};

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query);
    searchValue.value = buildSearchSelectValueBySearchQsCondition(condition.value, searchFields);
    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;
    sortParams.value = { sort: query.sort as string, order: query.order as string };
    await fetchList();
  },
  { immediate: true },
);

watch(
  () => props.filter,
  () => {
    routeQuery.set({
      filter: undefined,
      page: undefined,
      accountId: selectedAccountId.value || undefined,
      _t: Date.now(),
    });
  },
  { deep: true },
);
</script>

<template>
  <div class="exclusive-cluster">
    <div class="toolbar">
      <slot name="toolbar-prefix"></slot>
      <BatchDistribution
        ref="batch-distribution"
        :selections="selections"
        :type="DResourceType.exclusive_clusters"
        :submit="handleAssignSubmit"
        :get-data="handleRetry"
      />
      <div class="search-selector-container">
        <ResourceSearchSelect v-model="searchValue" :fields="searchFields" @update:model-value="handleSearch" />
      </div>
    </div>
    <DataList
      v-bkloading="{ loading: exclusiveClusterStore.exclusiveClusterListLoading }"
      :columns="columns"
      :list="listData"
      :pagination="pagination"
      @assign="handleAssign"
      @selection-change="handleSelectionChange"
      @page-change="handlePageChange"
      @page-size-change="handlePageSizeChange"
      @column-sort="handleSort"
    />
  </div>
</template>

<style lang="scss" scoped>
.exclusive-cluster {
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .toolbar {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 16px;
    flex-shrink: 0;

    // 窄屏时只让搜索框让宽，左侧操作按钮保持原宽，否则会被挤出可视区
    > *:not(.search-selector-container) {
      flex-shrink: 0;
    }

    .search-selector-container {
      margin-left: auto;
      min-width: 0;
    }
  }
}
</style>
