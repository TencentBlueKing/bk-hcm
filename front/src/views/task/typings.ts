import { PaginationType } from '@/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { ITaskItem } from '@/store/task';

export enum TaskType {
  CREATE_L4_LISTENER = 'create_layer4_listener',
  CREATE_L7_LISTENER = 'create_layer7_listener',
  CREATE_URL_FILTER = 'create_url_filter',
}

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
