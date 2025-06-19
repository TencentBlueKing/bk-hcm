<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch, watchEffect } from 'vue';
import { useRoute } from 'vue-router';
import { ITaskCountItem, ITaskDetailItem, ITaskItem, ITaskStatusItem, useTaskStore } from '@/store';
import { ResourceTypeEnum } from '@/common/resource-constant';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import CommonCard from '@/components/CommonCard';
import taskDetailsViewProperties from '@/model/task/detail.view';
import { transformSimpleCondition } from '@/utils/search';
import BasicInfo from './children/basic-info/basic-info.vue';
import ActionList from './children/action-list/action-list.vue';
import Rerun from './children/rerun/rerun.vue';
import Cancel from './children/cancel/cancel.vue';

import { TASK_CLB_TYPE_NAME } from '../constants';
import { TaskClbType, TaskStatus, TaskDetailStatus } from '../typings';

interface ArrayDataItem {
  domain: string[];
  url: string[];
  ip: string[];
  weight: string[];
}

const taskStore = useTaskStore();
const { getBizsId } = useWhereAmI();
const route = useRoute();

const { setTitle } = useBreadcrumb();

const searchQs = useSearchQs({ key: 'filter', properties: taskDetailsViewProperties });

const { pagination, getPageParams } = usePage();

const id = computed(() => String(route.params.id));
const bizId = computed(() => getBizsId());

const taskDetailList = ref<ITaskDetailItem[]>([]);
const taskDetails = ref<ITaskItem>();

const condition = ref<Record<string, any>>({});

const selections = ref([]);
const rerunState = reactive({
  isShow: false,
});

const isSopsOperation = computed(() =>
  taskDetails.value?.operations?.some?.((op) =>
    [
      TaskClbType.DELETE_LISTENER,
      TaskClbType.MODIFY_LAYER4_RS_WEIGHT,
      TaskClbType.MODIFY_LAYER7_RS_WEIGHT,
      TaskClbType.UNBIND_LAYER4_RS,
    ].includes(op),
  ),
);

const rerunButtonDisabled = computed(() => {
  return !selections.value.length || isSopsOperation.value;
});

// 本任务的状态
const status = ref<ITaskStatusItem>();
// 本任务统计数据
const counts = ref<ITaskCountItem>();

const statusPoolIds = computed(() => {
  return taskDetailList.value
    .filter((item) => [TaskDetailStatus.INIT, TaskDetailStatus.RUNNING].includes(item.state))
    .map((item) => item.id);
});

const fetchCountAndStatus = async () => {
  // 获取当前任务状态与统计数据
  const reqs: Promise<ITaskStatusItem[] | ITaskCountItem[] | ITaskDetailItem[]>[] = [
    taskStore.getTaskStatus({ bk_biz_id: getBizsId(), ids: [id.value] }),
    taskStore.getTaskCounts({ bk_biz_id: getBizsId(), ids: [id.value] }),
  ];

  // 获取当前任务详情列表中数据的状态
  if (statusPoolIds.value.length) {
    reqs.push(taskStore.getTaskDetailListStatus(statusPoolIds.value, bizId.value));
  }

  const [statusRes, countRes, detailStatusRes] = await Promise.allSettled(reqs);

  const statusList = (statusRes as PromiseFulfilledResult<ITaskStatusItem[]>).value ?? [];
  const detailStatusList = (detailStatusRes as PromiseFulfilledResult<ITaskDetailItem[]>)?.value ?? [];
  const countList = (countRes as PromiseFulfilledResult<ITaskCountItem[]>)?.value ?? [];

  // 更新本任务的状态与统计数据
  [counts.value] = countList;
  [status.value] = statusList;
  taskDetails.value.state = status.value?.state;

  // 更新当前任务详情列表中数据的状态
  taskDetailList.value.forEach((row) => {
    const foundState = detailStatusList.find((item) => item?.id === row.id);
    if (foundState) {
      row.state = foundState.state;
      row.reason = foundState.reason;
    }
  });
};

const taskStatusPoll = useTimeoutPoll(() => {
  fetchCountAndStatus();
}, 10000);

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query);
    condition.value.task_management_id = id.value;

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    taskDetails.value = await taskStore.getTaskDetails(id.value, bizId.value);

    const { list, count } = await taskStore.getTaskDetailList({
      bk_biz_id: bizId.value,
      filter: transformSimpleCondition(condition.value, taskDetailsViewProperties),
      page: getPageParams(pagination, { sort, order }),
    });

    // TODO: 按任务类型展示不同字段来优化列表数据
    list.forEach((item) => {
      const arrayData: ArrayDataItem = {
        domain: [],
        url: [],
        ip: [],
        weight: [],
      };
      item.param.rs_list?.forEach((rs_item: any) => {
        arrayData.domain.push(rs_item?.domain);
        arrayData.url.push(rs_item?.url);
        arrayData.ip.push(rs_item?.ip);
        arrayData.weight.push(rs_item?.weight);
      });
      item.param.ip = arrayData.ip.join(',');
      item.param.url = arrayData.url.join(',');
      item.param.domain = arrayData.domain.join(',');
      item.param.weight = arrayData.weight.join(',');
    });

    taskDetailList.value = list;
    pagination.count = count;

    fetchCountAndStatus();
  },
  { immediate: true },
);

watch(status, (newStatus) => {
  // 大任务的状态不是运行中，暂停轮询
  if (newStatus.state !== TaskStatus.RUNNING) {
    taskStatusPoll.pause();
  }
});

watchEffect(async () => {
  const operations = taskDetails.value?.operations ?? [];
  const taskOps = Array.isArray(operations) ? operations : [operations];
  const title = taskOps.map((op) => TASK_CLB_TYPE_NAME[op]).join(',');

  setTitle(title);
});

const handleActionSelect = (data: any[]) => {
  selections.value = data;
};

const handleClickRerun = () => {
  if (!selections.value.length) {
    return;
  }
  rerunState.isShow = true;
};

const handleClickStatusCount = (status?: TaskDetailStatus) => {
  searchQs.set({
    state: status,
  });
};

onMounted(() => {
  taskStatusPoll.resume();
});
</script>

<template>
  <common-card class="content-card" :title="() => '基本信息'">
    <basic-info :resource="ResourceTypeEnum.CLB" :data="taskDetails"></basic-info>
  </common-card>
  <common-card class="content-card" :title="() => '操作详情'">
    <div class="toolbar">
      <bk-button
        theme="primary"
        :disabled="rerunButtonDisabled"
        v-bk-tooltips="{ content: '暂不支持', disabled: !isSopsOperation }"
        @click="handleClickRerun"
      >
        失败任务重执行
      </bk-button>
      <div class="stats">
        <span class="count-item">
          总数:
          <bk-link theme="primary" class="num" @click.prevent="handleClickStatusCount()">
            {{ counts?.total ?? '--' }}
          </bk-link>
        </span>
        <span class="count-item">
          成功:
          <bk-link theme="primary" class="num" @click.prevent="handleClickStatusCount(TaskDetailStatus.SUCCESS)">
            {{ counts?.success ?? '--' }}
          </bk-link>
        </span>
        <span class="count-item">
          失败:
          <bk-link theme="primary" class="num" @click.prevent="handleClickStatusCount(TaskDetailStatus.FAILED)">
            {{ counts?.failed ?? '--' }}
          </bk-link>
        </span>
        <span class="count-item">
          未执行:
          <bk-link theme="primary" class="num" @click.prevent="handleClickStatusCount(TaskDetailStatus.INIT)">
            {{ counts?.init }}
          </bk-link>
        </span>
        <span class="count-item">
          运行中:
          <bk-link theme="primary" class="num" @click.prevent="handleClickStatusCount(TaskDetailStatus.RUNNING)">
            {{ counts?.running ?? '--' }}
          </bk-link>
        </span>
        <span class="count-item">
          取消:
          <bk-link theme="primary" class="num" @click.prevent="handleClickStatusCount(TaskDetailStatus.CANCEL)">
            {{ counts?.cancel ?? '--' }}
          </bk-link>
        </span>
      </div>
    </div>
    <action-list
      v-bkloading="{ loading: taskStore.taskDetailListLoading }"
      :resource="ResourceTypeEnum.CLB"
      :list="taskDetailList"
      :detail="taskDetails"
      :pagination="pagination"
      @select="handleActionSelect"
    />
  </common-card>
  <template v-if="selections.length">
    <rerun
      v-model="rerunState.isShow"
      :resource="ResourceTypeEnum.CLB"
      :info="taskDetails"
      :selected="selections"
    ></rerun>
  </template>
  <cancel :resource="ResourceTypeEnum.CLB" :info="taskDetails" :status="status?.state" />
</template>

<style lang="scss" scoped>
.content-card {
  + .content-card {
    margin-top: 20px;
  }

  :deep(.common-card-content) {
    width: 100%;
  }
}

.toolbar {
  display: flex;
  align-items: center;
  margin: 16px 0;
}

.stats {
  display: flex;
  gap: 16px;
  margin-left: 24px;

  .count-item {
    display: flex;
    align-items: center;
    gap: 4px;

    .num {
      font-style: normal;
    }
  }
}
</style>
