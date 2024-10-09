<script setup lang="ts">
import { computed, reactive, ref, watch, watchEffect } from 'vue';
import { useRoute } from 'vue-router';
import { ITaskItem, useTaskStore } from '@/store';
import { ResourceTypeEnum } from '@/common/resource-constant';
import useBreakcrumb from '@/hooks/use-breakcrumb';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import usePage from '@/hooks/use-page';
import CommonCard from '@/components/CommonCard';
import taskDetailsViewProperties from '@/model/task/detail.view';
import { transformSimpleCondition } from '@/utils/search';
import BasicInfo from './children/basic-info/basic-info.vue';
import ActionList from './children/action-list/action-list.vue';
import Rerun from './children/rerun/rerun.vue';

import { TASKT_CLB_TYPE_NAME } from '../constants';

const taskStore = useTaskStore();
const { getBizsId } = useWhereAmI();
const route = useRoute();

const { setTitle } = useBreakcrumb();

const { pagination, pageParams } = usePage(false);

const id = computed(() => String(route.params.id));
const bizId = computed(() => getBizsId());

const taskDetails = ref<ITaskItem>();

const condition = reactive<Record<string, any>>({});

const selections = ref([]);
const rerunState = reactive({
  isShow: false,
});

watch(
  [id, bizId, pageParams],
  async ([newId, newBizId, page]) => {
    condition.task_management_id = newId;
    taskDetails.value = await taskStore.getTaskDetails(newId, newBizId);

    await taskStore.getTaskDetailList({
      bk_biz_id: newBizId,
      filter: transformSimpleCondition(condition, taskDetailsViewProperties),
      page,
    });

    pagination.count = taskStore.taskDetailListCount;
  },
  { immediate: true },
);

watchEffect(async () => {
  const operations = taskDetails.value?.operations ?? [];
  const taskOps = Array.isArray(operations) ? operations : [operations];
  const title = taskOps.map((op) => TASKT_CLB_TYPE_NAME[op]).join(',');

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
</script>

<template>
  <common-card class="content-card" :title="() => '基本信息'">
    <basic-info :resource="ResourceTypeEnum.CLB" :data="taskDetails"></basic-info>
  </common-card>
  <common-card class="content-card" :title="() => '操作详情'">
    <div class="toolbar">
      <bk-button theme="primary" :disabled="!selections.length" @click="handleClickRerun">失败任务重执行</bk-button>
    </div>
    <action-list
      v-bkloading="{ loading: taskStore.taskDetailListLoading }"
      :resource="ResourceTypeEnum.CLB"
      :list="taskStore.taskDetailList"
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
</style>
