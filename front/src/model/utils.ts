import { ModelPropertyGeneric, ModelPropertyColumn } from '@/model/typings';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';

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

export const findSearchData = (id: ISearchItem['id'], searchData: ISearchItem[], key?: keyof ISearchItem) => {
  // 先按默认的规则找
  let found = searchData.find((data) => data.id === id);

  // 找不到同时指定了key则再根据key再找一次
  if (!found && key) {
    found = searchData.find((data) => data[key] === id);
  }

  return found;
};
