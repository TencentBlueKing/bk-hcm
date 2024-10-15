import { ModelPropertyColumn } from '@/model/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';
import taskDetailsViewProperties from '@/model/task/detail.view';
import { ITaskItem } from '@/store';
import { type TaskType } from '@/views/task/typings';
import { baseFieldIds, fieldIdMap, fieldRerunIdMap, fieldRerunBaseIdMap, baseColumnConfig } from './fields';

const taskActionViewProperties: ModelPropertyColumn[] = [...taskDetailsViewProperties];

export const getColumnIds = (resourceType: ResourceTypeEnum, operation: TaskType) => {
  const resourceColumnIds = fieldIdMap.get(resourceType);
  return resourceColumnIds[operation] || baseFieldIds;
};

const getColumns = (type: ResourceTypeEnum, operations?: ITaskItem['operations']) => {
  const [operation] = operations || [];
  const columnIds = getColumnIds(type, operation as TaskType);
  return columnIds.map((id) => ({
    ...taskActionViewProperties.find((item) => item.id === id),
    ...baseColumnConfig[id],
  }));
};

const getRerunColumns = (type: ResourceTypeEnum, operations?: ITaskItem['operations']) => {
  const [operation] = operations || [];
  const fields = fieldRerunIdMap.get(type);
  const opeartionFields = fields[operation as TaskType] || fieldRerunBaseIdMap;

  const columns = [];
  for (const [fieldId, setting] of Object.entries(opeartionFields)) {
    const newSetting = setting;
    if (!Object.hasOwn(newSetting, 'display')) {
      newSetting.display = {};
    }
    newSetting.display.on = 'cell';
    columns.push({
      field: taskActionViewProperties.find((item) => item.id === fieldId),
      setting: newSetting,
    });
  }
  return columns;
};

const factory = {
  getColumns,
  getRerunColumns,
};

export type FactoryType = typeof factory;

export default factory;
