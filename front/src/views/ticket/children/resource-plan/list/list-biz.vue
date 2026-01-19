<script setup lang="ts">
import { computed, provide, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { transformFlatCondition } from '@/utils/search';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import routerAction from '@/router/utils/action';
import { type IResourcePlanTicketItem, useResourcePlanTicketStore } from '@/store/ticket/resource-plan';
import type { ISearchCondition } from '../typings';
import { properties as conditionProperties } from './children/search/condition';
import { properties as columnProperties } from './children/data-list/column';
import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import { MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS } from '@/constants/menu-symbol';

const route = useRoute();

const resourcePlanTicketStore = useResourcePlanTicketStore();

const { pagination, getPageParams } = usePage();

const bizId = computed(() => Number(route.query[GLOBAL_BIZS_KEY]));

const searchFields = computed(() =>
  conditionProperties.filter((prop) => !['bk_biz_ids', 'op_product_ids', 'plan_product_ids'].includes(prop.id)),
);
const dataListColumns = computed(() =>
  columnProperties.filter((prop) => !['bk_biz_name', 'op_product_name', 'plan_product_name'].includes(prop.id)),
);

const searchQs = useSearchQs({ properties: conditionProperties });

const condition = ref<Record<string, any>>({});
const ticketList = ref<IResourcePlanTicketItem[]>([]);

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {});

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = query.sort as string;
    const order = query.order as string;

    const { list, count } = await resourcePlanTicketStore.getTicketList(
      {
        ...transformFlatCondition(condition.value, conditionProperties),
        page: getPageParams(pagination, { sort, order }),
      },
      bizId.value,
    );

    ticketList.value = list;
    pagination.count = count;
  },
  { immediate: true },
);

const handleSearch = (vals: ISearchCondition) => {
  searchQs.set(vals);
};

const handleReset = () => {
  searchQs.clear();
};

const handleViewDetails = (row: IResourcePlanTicketItem) => {
  routerAction.redirect(
    {
      name: MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS,
      query: {
        id: row.id,
        type: 'resource_plan',
        [GLOBAL_BIZS_KEY]: row.bk_biz_id,
      },
    },
    {
      history: true,
    },
  );
};

provide('isBusinessPage', true);
</script>
<template>
  <search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />
  <data-list
    v-bkloading="{ loading: resourcePlanTicketStore.ticketListLoading }"
    :columns="dataListColumns"
    :list="ticketList"
    :pagination="pagination"
    @view-details="handleViewDetails"
  />
</template>
