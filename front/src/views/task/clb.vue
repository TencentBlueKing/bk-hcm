<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import { useUserStore, useTaskStore, ITaskStatusItem, ITaskCountItem, ITaskItem } from '@/store';
import routerAction from '@/router/utils/action';
import { TaskStatus, type ISearchConditon } from '@/views/task/typings';
import accountProperties from '@/model/account/properties';
import taskProperties from '@/model/task/properties';
import { transformSimpleCondition, getDateRange } from '@/utils/search';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';

const route = useRoute();
const userStore = useUserStore();
const taskStore = useTaskStore();
const { getBizsId } = useWhereAmI();

const taskViewProperties = [...accountProperties, ...taskProperties];

const searchQs = useSearchQs({ key: 'filter', properties: taskViewProperties });
const { pagination, getPageParams } = usePage();

const taskList = ref<ITaskItem[]>([]);
const condition = ref<Record<string, any>>({});

const fetchCountAndStatus = async (ids?: ITaskItem['id'][]) => {
  const fetchIds = !ids ? taskList.value.map((item) => item.id) : ids;
  if (!fetchIds.length) {
    return;
  }

  const [statusRes, countRes] = await Promise.allSettled([
    taskStore.getTaskStatus({ bk_biz_id: getBizsId(), ids: fetchIds }),
    taskStore.getTaskCounts({ bk_biz_id: getBizsId(), ids: fetchIds }),
  ]);

  if (statusRes.status === 'rejected' && countRes.status === 'rejected') {
    return;
  }

  const statusList = (statusRes as PromiseFulfilledResult<ITaskStatusItem[]>).value ?? [];
  const countList = (countRes as PromiseFulfilledResult<ITaskCountItem[]>).value ?? [];

  taskList.value.forEach((row) => {
    const foundState = statusList.find((item) => item?.id === row.id);
    const foundCount = countList.find((item) => item?.id === row.id);
    if (foundState) {
      row.state = foundState.state;
    }
    if (foundCount) {
      row.count_total = foundCount.total;
      row.count_success = foundCount.success;
      row.count_failed = foundCount.failed;
    }
  });
};

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      created_at: getDateRange('last7d'),
      creator: userStore.username,
    });

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    const { list, count } = await taskStore.getTaskList({
      bk_biz_id: getBizsId(),
      filter: transformSimpleCondition(condition.value, taskViewProperties),
      page: getPageParams(pagination, { sort, order }),
    });

    taskList.value = list;

    // 设置页码总条数
    pagination.count = count;

    // 获取数量和状态
    fetchCountAndStatus();
  },
  { immediate: true },
);

const taskStatusPoll = useTimeoutPoll(() => {
  const ids = taskList.value.filter((item) => [TaskStatus.RUNNING].includes(item.state)).map((item) => item.id);
  fetchCountAndStatus(ids);
}, 5000);

const handleSearch = (vals: ISearchConditon) => {
  searchQs.set(vals);
};

const handleReset = () => {
  searchQs.clear();
};

const handleViewDetails = (id: string) => {
  routerAction.redirect(
    {
      name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
      params: { id, resourceType: ResourceTypeEnum.CLB },
      query: { bizs: getBizsId() },
    },
    {
      history: true,
    },
  );
};

onMounted(() => {
  taskStatusPoll.resume();
});
</script>

<template>
  <search :resource="ResourceTypeEnum.CLB" :condition="condition" @search="handleSearch" @reset="handleReset" />
  <data-list
    v-bkloading="{ loading: taskStore.taskListLoading }"
    :resource="ResourceTypeEnum.CLB"
    :list="taskList"
    :pagination="pagination"
    @view-details="handleViewDetails"
  />
</template>
