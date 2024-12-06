import { PaginationType } from '@/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { ITaskItem, ITaskDetailItem } from '@/store/task';

export enum TaskClbType {
  CREATE_L4_LISTENER = 'create_layer4_listener',
  CREATE_L7_LISTENER = 'create_layer7_listener',
  CREATE_L7_RULE = 'create_layer7_rule',
  DELETE_LISTENER = 'listener_delete',
  BINDING_L4_RS = 'binding_layer4_rs',
  BINDING_L7_RS = 'binding_layer7_rs',
  UNBIND_RS = 'listener_unbind_rs',
  MODIFY_RS_WEIGHT = 'listener_rs_weight',
}

export type TaskType = TaskClbType;

export enum TaskStatus {
  RUNNING = 'running',
  FAILED = 'failed',
  SUCCESS = 'success',
  DELIVER_PARTIAL = 'deliver_partial',
  CANCELED = 'cancel',
}

export enum TaskSource {
  SOPS = 'sops',
  EXCEL = 'excel',
}

export enum TaskDetailStatus {
  INIT = 'init',
  RUNNING = 'running',
  FAILED = 'failed',
  SUCCESS = 'success',
  CANCEL = 'cancel',
}

export interface ISearchConditon {
  account?: string;
  type?: TaskType;
  state?: TaskStatus;
  source?: TaskSource;
  created_at?: string;
  creator?: string;
  [key: string]: any;
}

export interface ISearchProps {
  resource: ResourceTypeEnum;
  condition: ISearchConditon;
}

export interface IDataListProps {
  resource: ResourceTypeEnum;
  list: ITaskItem[];
  pagination: PaginationType;
}

export interface IActionListProps {
  resource: ResourceTypeEnum;
  list: ITaskDetailItem[];
  detail: Partial<ITaskItem>;
  pagination: PaginationType;
}
