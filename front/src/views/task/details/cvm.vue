<script setup lang="ts">
import { computed, onMounted, ref, watch, watchEffect } from 'vue';
import { useRoute } from 'vue-router';
import { ITaskCountItem, ITaskDetailItem, ITaskItem, useTaskStore } from '@/store';
import { ResourceTypeEnum } from '@/common/resource-constant';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import CommonCard from '@/components/CommonCard';
import taskDetailsViewProperties from '@/model/task/detail.view';
import { transformSimpleCondition } from '@/utils/search';
import { getPrivateIPs } from '@/utils/common';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import BasicInfo from './children/basic-info/basic-info.vue';
import ActionList from './children/action-list/action-list.vue';
import Cancel from './children/cancel/cancel.vue';
import ExecuteRecord from '@/views/ziyanScr/host-recycle/execute-record/index.vue';

import { TASK_TYPE_NAME } from '../constants';
import { TaskCvmType, TaskDetailStatus } from '../typings';

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

// 本任务统计数据
const counts = ref<ITaskCountItem>();

const statusPoolIds = computed(() => {
  return taskDetailList.value
    .filter((item) => [TaskDetailStatus.INIT, TaskDetailStatus.RUNNING].includes(item.state))
    .map((item) => item.id);
});

const fetchCountAndStatus = async () => {
  // 获取当前任务状态与统计数据
  const reqs: Promise<ITaskItem | ITaskCountItem[] | ITaskDetailItem[]>[] = [
    taskStore.getTaskDetails(id.value, getBizsId()),
    taskStore.getTaskCounts({ bk_biz_id: getBizsId(), ids: [id.value] }),
  ];

  // 获取当前任务详情列表中数据的状态
  if (statusPoolIds.value.length) {
    reqs.push(taskStore.getTaskDetailListStatus(statusPoolIds.value, bizId.value));
  }

  const [statusRes, countRes, detailStatusRes] = await Promise.allSettled(reqs);

  // 更新本任务的状态与统计数据
  taskDetails.value = (statusRes as PromiseFulfilledResult<ITaskItem>).value;
  [counts.value] = (countRes as PromiseFulfilledResult<ITaskCountItem[]>)?.value ?? [];
  const detailStatusList = (detailStatusRes as PromiseFulfilledResult<ITaskDetailItem[]>)?.value ?? [];

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

    taskDetailList.value = list;
    pagination.count = count;

    fetchCountAndStatus();
  },
  { immediate: true },
);

watchEffect(async () => {
  const operations = taskDetails.value?.operations ?? [];
  const taskOps = Array.isArray(operations) ? operations : [operations];
  const title = taskOps.map((op) => TASK_TYPE_NAME[op]).join(',');

  setTitle(`主机${title}`);
});

const getStatusIPs = async (status?: TaskDetailStatus) => {
  const list = (await taskStore.getTaskDetailListAll({
    bk_biz_id: bizId.value,
    filter: transformSimpleCondition(
      {
        task_management_id: id.value,
        state: status,
      },
      taskDetailsViewProperties,
    ),
    fields: ['param'],
  })) as ITaskItem[];

  const ips = list.map((item) => getPrivateIPs(item.param)).join('\n');

  return ips;
};

const handleClickStatusCount = (status?: TaskDetailStatus) => {
  searchQs.set({
    state: status,
  });
};

const showActions = computed(() => {
  return [TaskCvmType.IDLE_CHECK].includes(taskDetails.value?.operations[0] as TaskCvmType);
});
const executeDataInfo = ref({
  ip: '',
  suborderId: '',
  page: {
    start: 0,
    limit: 20,
  },
});
const openRecordDetail = ref(false);
const handleIdelCheckDetail = (row: ITaskDetailItem) => {
  executeDataInfo.value = {
    ip: row.param?.ip,
    suborderId: row.param?.sub_order_id,
    page: {
      start: 0,
      limit: 20,
    },
  };
  openRecordDetail.value = true;
};

onMounted(() => {
  taskStatusPoll.resume();
});
</script>

<template>
  <common-card class="content-card" :title="() => '基本信息'">
    <basic-info :resource="ResourceTypeEnum.CVM" :data="taskDetails"></basic-info>
  </common-card>
  <common-card class="content-card" :title="() => '操作详情'">
    <div class="toolbar">
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
      <div class="action-buttons">
        <copy-to-clipboard :disabled="!counts?.total" :content="() => getStatusIPs()">
          <template #default="{ disabled, loading }">
            <bk-button theme="primary" :disabled="disabled" :loading="loading">复制全部IP</bk-button>
          </template>
        </copy-to-clipboard>
        <copy-to-clipboard :disabled="!counts?.success" :content="() => getStatusIPs(TaskDetailStatus.SUCCESS)">
          <template #default="{ disabled, loading }">
            <bk-button :disabled="disabled" :loading="loading">复制成功IP</bk-button>
          </template>
        </copy-to-clipboard>
        <copy-to-clipboard :disabled="!counts?.failed" :content="() => getStatusIPs(TaskDetailStatus.FAILED)">
          <template #default="{ disabled, loading }">
            <bk-button :disabled="disabled" :loading="loading">复制失败IP</bk-button>
          </template>
        </copy-to-clipboard>
      </div>
    </div>
    <action-list
      v-bkloading="{ loading: taskStore.taskDetailListLoading }"
      :resource="ResourceTypeEnum.CVM"
      :list="taskDetailList"
      :detail="taskDetails"
      :pagination="pagination"
      :selectable="false"
    >
      <template v-if="showActions" #action="{ row }">
        <bk-button theme="primary" text @click="handleIdelCheckDetail(row)">详情</bk-button>
      </template>
    </action-list>
  </common-card>
  <cancel :resource="ResourceTypeEnum.CVM" :info="taskDetails" :status="taskDetails?.state" />
  <execute-record v-model="openRecordDetail" :data-info="executeDataInfo" />
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
  flex-direction: column;
  gap: 12px;
  margin: 16px 0;
}

.stats {
  display: flex;
  align-items: center;
  gap: 16px;

  .count-item {
    display: flex;
    align-items: center;
    gap: 4px;

    .num {
      font-style: normal;
    }
  }
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
