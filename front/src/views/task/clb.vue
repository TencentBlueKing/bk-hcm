<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import { useUserStore, useTaskStore, ITaskStatusItem, ITaskCountItem } from '@/store';
import { TaskStatus, type ISearchConditon } from '@/views/task/typings';
import accountProperties from '@/model/account/properties';
import taskProperties from '@/model/task/properties';
import { transformSimpleCondition, getDateRange } from '@/utils/search';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';

const route = useRoute();
const userStore = useUserStore();
const taskStore = useTaskStore();
const { getBizsId } = useWhereAmI();

const taskViewProperties = [...accountProperties, ...taskProperties];

const searchQs = useSearchQs({ key: 'filter', properties: taskViewProperties });
const { pagination, getPageParams } = usePage();

const handleSearch = (vals: ISearchConditon) => {
  searchQs.set(vals);
};
const handleReset = () => {
  searchQs.clear();
  taskStatusPoll.pause();
};

const taskStatusPoll = useTimeoutPoll(
  async () => {
    const ids = taskStore.taskList.filter((item) => [TaskStatus.SUCCESS].includes(item.state)).map((item) => item.id);
    if (!ids.length) {
      return;
    }

    const [statusRes, countRes] = await Promise.allSettled([
      taskStore.getTaskStatus({ bk_biz_id: getBizsId(), ids }),
      taskStore.getTaskCounts({ bk_biz_id: getBizsId(), ids }),
    ]);

    if (statusRes.status === 'rejected' && countRes.status === 'rejected') {
      return;
    }

    const statusList = (statusRes as PromiseFulfilledResult<ITaskStatusItem[]>).value ?? [];
    const countList = (countRes as PromiseFulfilledResult<ITaskCountItem[]>).value ?? [];

    taskStore.taskList.forEach((row) => {
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
  },
  5000,
  { immediate: false },
);

const condition = ref<Record<string, any>>({});

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      created_at: getDateRange('last7d'),
      creator: userStore.username,
    });

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    await taskStore.getTaskList({
      bk_biz_id: getBizsId(),
      filter: transformSimpleCondition(condition.value, taskViewProperties),
      page: getPageParams(pagination, { sort: query.sort as string, order: query.order as string }),
    });

    pagination.count = taskStore.taskListCount;
  },
  { immediate: true },
);

onMounted(() => {
  taskStatusPoll.resume();
});
</script>

<template>
  <div>
    <search :resource="ResourceTypeEnum.CLB" :condition="condition" @search="handleSearch" @reset="handleReset" />
    <data-list
      v-bkloading="{ loading: taskStore.taskListLoading }"
      :resource="ResourceTypeEnum.CLB"
      :list="taskStore.taskList"
      :pagination="pagination"
    />
  </div>
</template>
