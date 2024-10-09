import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { QueryRuleOPEnum, IListResData, QueryBuilderType } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { enableCount, onePageParams } from '@/utils/search';
import { TaskDetailStatus, TaskStatus, TaskType } from '@/views/task/typings';

export interface ITaskItem {
  id: string;
  bk_biz_id: number;
  source: string;
  vendor: VendorEnum;
  state: TaskStatus;
  account_id: string;
  operations: TaskType | TaskType[];
  flow_ids: string | string[];
  extension: Record<string, any>;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  count_total?: number | string;
  count_success?: number | string;
  count_failed?: number | string;
  [k: string]: any;
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

type ITaskDetailParam = {
  clb_vip_domain?: string;
  cloud_clb_id?: string;
  protocol?: string;
  listener_port: string[];
  ssl_mode?: string;
  domain?: string;
  url_path?: string;
  health_check?: boolean;
  session?: number;
};

export interface ITaskDetailItem {
  id: string;
  task_management_id: string;
  flow_id: string;
  task_action_ids: string[];
  operation: string;
  param: ITaskDetailParam;
  result: object;
  state: TaskDetailStatus;
  reason: string;
  extension: object;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  [k: string]: any;
}

export const useTaskStore = defineStore('task', () => {
  const taskList = ref<ITaskItem[]>([]);
  const taskDetailList = ref<ITaskDetailItem[]>([]);
  const taskListCount = ref<number>(0);
  const taskDetailListCount = ref<number>(0);

  const taskListLoading = ref(false);
  const taskStatusLoading = ref(false);
  const taskCountLoading = ref(false);
  const taskDetailListLoading = ref(false);
  const taskDetailsLoading = ref(false);

  const getTaskList = async (params: QueryBuilderType & { bk_biz_id: number }) => {
    const { bk_biz_id, ...data } = params;
    taskListLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/task_managements/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ITaskItem[]>>, Promise<IListResData<ITaskItem[]>>]
      >([http.post(api, enableCount(data, false)), http.post(api, enableCount(data, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
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

  const getTaskDetailList = async (params: QueryBuilderType & { bk_biz_id: number }) => {
    const { bk_biz_id, ...data } = params;
    taskDetailListLoading.value = true;
    const api = `/api/v1/cloud/bizs/${bk_biz_id}/task_details/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ITaskDetailItem[]>>, Promise<IListResData<ITaskDetailItem[]>>]
      >([http.post(api, enableCount(data, false)), http.post(api, enableCount(data, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      taskDetailList.value = list;
      taskDetailListCount.value = count;
    } catch {
      taskDetailList.value = [];
      taskDetailListCount.value = 0;
    } finally {
      taskDetailListLoading.value = false;
    }
  };

  const getTaskDetails = async (id: ITaskItem['id'], bizId: number): Promise<ITaskItem> => {
    taskDetailsLoading.value = true;
    try {
      const res: IListResData<ITaskItem[]> = await http.post(`/api/v1/cloud/bizs/${bizId}/task_managements/list`, {
        bk_biz_id: bizId,
        filter: {
          op: 'and',
          rules: [{ field: 'id', op: QueryRuleOPEnum.EQ, value: id }],
        },
        page: onePageParams(),
      });
      return res?.data?.details?.[0];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      taskDetailsLoading.value = false;
    }
  };

  return {
    taskList,
    taskListCount,
    taskListLoading,
    taskStatusLoading,
    taskCountLoading,
    taskDetailList,
    taskDetailListCount,
    taskDetailListLoading,
    getTaskList,
    getTaskStatus,
    getTaskCounts,
    getTaskDetails,
    getTaskDetailList,
  };
});
