import { type ComputedRef, isRef } from 'vue';
import { type LocationQuery } from 'vue-router';
import qs from 'qs';
import { ModelPropertyGeneric } from '@/model/typings';
import { findProperty } from '@/model/utils';
import routeQuery from '@/router/utils/query';
import { convertValue, getDefaultRule } from '@/utils/search';

type useSearchQsParamsType = {
  properties?: ModelPropertyGeneric[] | ComputedRef<ModelPropertyGeneric[]>;
  key?: string;
  forceUpdate?: boolean;
  resetPage?: boolean;
};

export default function useSearchQs({
  properties = [],
  key = 'filter',
  forceUpdate = true,
  resetPage = true,
}: useSearchQsParamsType = {}) {
  const build = (value: Record<string, string | number | string[] | number[]>) => {
    return qs.stringify(value, {
      arrayFormat: 'comma',
      encode: false,
    });
  };

  const set = (value: Record<string, string | number | string[] | number[]>) => {
    const queryVal = build(value);

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
    const filter = qs.parse(query[key] as string, { comma: true });
    for (const [id, val] of Object.entries(filter)) {
      const property = findProperty(id, isRef(properties) ? properties.value : properties);
      if (!property) {
        continue;
      }
      const { op } = getDefaultRule(property);
      condition[id] = convertValue(val, property, op);
    }
    return condition;
  };

  const clear = () => {
    routeQuery.delete(key);
  };

  return {
    build,
    get,
    set,
    clear,
  };
}
