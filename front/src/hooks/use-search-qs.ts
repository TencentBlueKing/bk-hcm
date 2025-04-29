import { type LocationQuery } from 'vue-router';
import qs from 'qs';
import { ModelProperty } from '@/model/typings';
import { findProperty, findSearchData } from '@/model/utils';
import routeQuery from '@/router/utils/query';
import { convertValue } from '@/utils/search';
import { ISearchSelectValue } from '@/typings';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';

type useSearchQsParamsType = {
  properties: ModelProperty[];
  key?: string;
  forceUpdate?: boolean;
  resetPage?: boolean;
};

export default function useSearchQs({
  properties,
  key = 'filter',
  forceUpdate = true,
  resetPage = true,
}: useSearchQsParamsType) {
  const set = (value: Record<string, string | number | string[] | number[]>) => {
    const queryVal = qs.stringify(value, {
      arrayFormat: 'comma',
      encode: false,
      allowEmptyArrays: true,
    });

    const updateQuery = { [key]: queryVal };
    if (resetPage) {
      updateQuery.page = undefined;
    }
    routeQuery.set(updateQuery, null, forceUpdate);
  };

  const get = (query: LocationQuery, defaults?: Record<string, any>) => {
    if (!Object.hasOwn(query, key)) {
      return { ...defaults };
    }
    const condition: Record<string, any> = {};
    const filter = qs.parse(query[key] as string, { comma: true, allowEmptyArrays: true });
    for (const [id, val] of Object.entries(filter)) {
      const property = findProperty(id, properties);
      condition[id] = convertValue(val, property);
    }
    return condition;
  };

  const buildSearchValue = (searchDataConfig: ISearchItem[], query: LocationQuery, defaults?: Record<string, any>) => {
    // 获取值的显示名称，优先从children中查找，找不到则返回原值
    const getDisplayName = (value: any, children: ISearchItem['children']) => {
      return children?.find((item) => item.id === value)?.name || value;
    };

    const condition = get(query, defaults);
    const searchValue: ISearchSelectValue = [];

    for (const [id, val] of Object.entries(condition)) {
      const searchData = findSearchData(id, searchDataConfig);
      if (!searchData) continue;

      const { name, multiple, children } = searchData;

      if (Array.isArray(val)) {
        if (multiple) {
          // 处理多选数组情况
          searchValue.push({
            id,
            name,
            values: val.map((item) => ({ id: item, name: getDisplayName(item, children) })),
          });
        } else {
          // 处理单选数组情况(展开为多个搜索项)
          searchValue.push(
            ...val.map((item: any) => ({
              id,
              name,
              values: [{ id: item, name: getDisplayName(item, children) }],
            })),
          );
        }
      } else {
        // 处理单值情况
        searchValue.push({ id, name, values: [{ id: val, name: getDisplayName(val, children) }] });
      }
    }

    return searchValue;
  };

  const clear = () => {
    routeQuery.delete(key);
  };

  return {
    get,
    set,
    buildSearchValue,
    clear,
  };
}
