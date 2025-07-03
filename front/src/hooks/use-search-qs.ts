import { type ComputedRef, isRef } from 'vue';
import { type LocationQuery } from 'vue-router';
import qs from 'qs';
import { ModelPropertyGeneric } from '@/model/typings';
import { findProperty } from '@/model/utils';
import routeQuery from '@/router/utils/query';
import { convertValue } from '@/utils/search';

type useSearchQsParamsType = {
  properties: ModelPropertyGeneric[] | ComputedRef<ModelPropertyGeneric[]>;
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
  const set = (value: Record<string, string | number | string[] | number[]>, allReplace = true) => {
    const queryVal = qs.stringify(value, {
      arrayFormat: 'comma',
      encode: false,
      allowEmptyArrays: true,
    });
    if (!allReplace) return routeQuery.set(key, queryVal, forceUpdate);
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
      const property = findProperty(id, isRef(properties) ? properties.value : properties);
      condition[id] = convertValue(val, property);
    }
    return condition;
  };

  const clear = () => {
    routeQuery.delete(key);
  };

  return {
    get,
    set,
    clear,
  };
}
