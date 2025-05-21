import { ModelPropertyGeneric, ModelPropertyColumn } from '@/model/typings';

export const findProperty = (
  id: ModelPropertyGeneric['id'],
  properties: ModelPropertyGeneric[],
  key?: keyof ModelPropertyGeneric,
) => {
  // 先按默认的规则找
  let found = properties.find((property) => property.id === id);

  // 找不到同时指定了key则再根据key再找一次
  if (!found && key) {
    found = properties.find((property) => property[key] === id);
  }

  return found;
};

export const getColumnName = (property: ModelPropertyColumn, options?: { showUnit: boolean }) => {
  const { showUnit = true } = options || {};
  const { name, unit } = property;
  return `${name}${showUnit && unit ? `（${unit}）` : ''}`;
};
