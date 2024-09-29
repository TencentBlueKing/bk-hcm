import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryBuilderType } from '@/typings';
import { enableCount } from '@/utils/search';
import { TaskStatus } from '@/views/task/typings';

export interface ITaskItem {
  id: string;
  bk_biz_id: number;
  source: string;
  vendor: string;
  state: TaskStatus;
  account_id: string;
  operations: string | string[];
  flow_ids: string | string[];
  extension: Record<string, any>;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  count_total?: number | string;
  count_success?: number | string;
  count_failed?: number | string;
}

export interface ITaskStatusItem {
  id: string;
  state: TaskStatus;
}

export interface ITaskCountItem {
  id: string;
  success: number;
  failed: number;
  init: number;
  running: number;
  cancel: number;
  total: number;
}

export const useTaskStore = defineStore('task', () => {
  const taskList = ref<ITaskItem[]>([]);
  const taskListCount = ref<number>(0);
  const taskListLoading = ref(false);
  const taskStatusLoading = ref(false);
  const taskCountLoading = ref(false);

  const getTaskList = async (params: QueryBuilderType & { bk_biz_id: number }) => {
    const { bk_biz_id, ...data } = params;
    taskListLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/task_managements/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ITaskItem[]>>, Promise<IListResData<ITaskItem[]>>]
      >([http.post(api, enableCount(data, false)), http.post(api, enableCount(data, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      taskList.value = list;
      taskListCount.value = count;
    } catch {
      taskList.value = [];
      taskListCount.value = 0;
    } finally {
      taskListLoading.value = false;
    }
  };

  const getTaskStatus = async (params: { bk_biz_id: number; ids: ITaskItem['id'][] }) => {
    const { bk_biz_id, ids } = params;
    taskStatusLoading.value = true;
    try {
      const res: IListResData<ITaskStatusItem[]> = await http.post(
        `/api/v1/cloud/bizs/${bk_biz_id}/task_managements/state/list`,
        { ids },
      );
      return res?.data?.details ?? [];
    } finally {
      taskStatusLoading.value = false;
    }
  };

  const getTaskCounts = async (params: { bk_biz_id: number; ids: ITaskItem['id'][] }) => {
    const { bk_biz_id, ids } = params;
    taskCountLoading.value = true;
    try {
      const res: IListResData<ITaskCountItem[]> = await http.post(
        `/api/v1/cloud/bizs/${bk_biz_id}/task_details/state/count`,
        { ids },
      );
      return res?.data?.details ?? [];
    } finally {
      taskCountLoading.value = false;
    }
  };

  return {
    taskList,
    taskListCount,
    taskListLoading,
    taskStatusLoading,
    taskCountLoading,
    getTaskList,
    getTaskStatus,
    getTaskCounts,
  };
});
