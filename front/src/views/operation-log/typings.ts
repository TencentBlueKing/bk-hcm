import { OPERATION_LOG_RESOURCE_TYPE, OPERATION_LOG_ACTION, OPERATION_LOG_SOURCE } from './constants';

export type OperationLogResourceType = (typeof OPERATION_LOG_RESOURCE_TYPE)[keyof typeof OPERATION_LOG_RESOURCE_TYPE];

export type OperationLogAction = (typeof OPERATION_LOG_ACTION)[keyof typeof OPERATION_LOG_ACTION];

export type OperationLogSource = (typeof OPERATION_LOG_SOURCE)[keyof typeof OPERATION_LOG_SOURCE];

export interface ISearchCondition {
  [key: string]: any;
}
