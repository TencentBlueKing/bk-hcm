import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { QueryRuleOPEnum, IListResData, QueryBuilderType } from '@/typings';
import { VendorEnum } from '@/common/constant';
import { enableCount, maxPageParams, onePageParams } from '@/utils/search';
import { TaskDetailStatus, TaskStatus, TaskType } from '@/views/task/typings';

export interface ITaskItem {
  id: string;
  bk_biz_id: number;
  source: string;
  vendor: VendorEnum;
  state: TaskStatus;
  account_id: string;
  operations: TaskType[];
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

export interface ITaskDetailParam {
  clb_vip_domain?: string;
  cloud_clb_id?: string;
  protocol?: string;
  listener_port?: string[];
  ssl_mode?: string;
  domain?: string;
  url_path?: string;
  health_check?: boolean;
  region_id?: string;
  session?: number;
  status?: 'executable' | 'not_executable' | 'existing';
  validate_result?: string;
  [k: string]: any;
}

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

export interface ITaskRerunParams {
  bk_biz_id: number;
  vendor: string;
  operation_type: string;
  data: {
    account_id: string;
    region_ids: string[];
    source?: string;
    details: ITaskDetailParam[];
  };
}

export const useTaskStore = defineStore('task', () => {
  const taskListLoading = ref(false);
  const taskStatusLoading = ref(false);
  const taskCountLoading = ref(false);
  const taskDetailListLoading = ref(false);
  const taskDetailsLoading = ref(false);
  const taskRerunValidateLoading = ref(false);
  const taskRerunSubmitLoading = ref(false);

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
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
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

  const getTaskDetailListStatus = async (ids: ITaskItem['id'][], bizId: number): Promise<ITaskDetailItem[]> => {
    try {
      const res: IListResData<ITaskDetailItem[]> = await http.post(`/api/v1/cloud/bizs/${bizId}/task_details/list`, {
        bk_biz_id: bizId,
        filter: {
          op: 'and',
          rules: [{ field: 'id', op: QueryRuleOPEnum.IN, value: ids }],
        },
        fields: ['state'],
        page: maxPageParams(),
      });
      return res?.data?.details;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const taskRerunValidate = async (params: ITaskRerunParams) => {
    taskRerunValidateLoading.value = true;
    const { bk_biz_id, vendor, operation_type, data } = params;
    try {
      const res = await http.post(
        `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/load_balancers/operations/${operation_type}/validate`,
        data,
      );
      return res?.data?.details;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      taskRerunValidateLoading.value = false;
    }
  };

  const taskRerunSubmit = async (params: ITaskRerunParams) => {
    taskRerunSubmitLoading.value = true;

    const { bk_biz_id, vendor, operation_type, data } = params;
    data.source = 'excel';

    try {
      const res = await http.post(
        `/api/v1/cloud/bizs/${bk_biz_id}/vendors/${vendor}/load_balancers/operations/${operation_type}/submit`,
        data,
      );
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      taskRerunSubmitLoading.value = false;
    }
  };

  const taskCancel = async (ids: ITaskItem['id'][], bizId: number) => {
    try {
      await http.post(`/api/v1/cloud/bizs/${bizId}/task_managements/cancel`, { ids });
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    taskListLoading,
    taskStatusLoading,
    taskCountLoading,
    taskDetailListLoading,
    taskRerunValidateLoading,
    taskRerunSubmitLoading,
    getTaskList,
    getTaskStatus,
    getTaskCounts,
    getTaskDetails,
    getTaskDetailList,
    getTaskDetailListStatus,
    taskRerunValidate,
    taskRerunSubmit,
    taskCancel,
  };
});
