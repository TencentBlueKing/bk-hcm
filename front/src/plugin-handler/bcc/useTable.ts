import http from '@/http';
import { FetchDataType, fetchData as defaultFetchData } from '../useTable';

export const fetchData: FetchDataType = async (params: any) => {
  const { props, BK_HCM_AJAX_URL_PREFIX, pagination, sort, order } = params;

  let detailsRes;
  let countRes;

  if (typeof props.scrConfig === 'function') {
    const { url, payload, pageEnableCountKey = 'enable_count', clearRules = true } = props.scrConfig();
    // 处理排序格式
    let sortVal = {};
    if (props.requestOption.sortOption.legacy) {
      sortVal = { sort: `${sort.value}:${order.value === 'ASC' ? 1 : -1}` };
    } else {
      sortVal = {
        sort: sort.value,
        order: order.value,
      };
    }
    [detailsRes, countRes] = await Promise.all(
      [false, true].map((isCount) => {
        if (isCount) sortVal = {};
        return http.post(
          BK_HCM_AJAX_URL_PREFIX + url,
          Object.assign(clearRules && payload?.filter?.rules.length === 0 ? {} : payload, {
            page: {
              start: isCount ? 0 : pagination.start,
              limit: isCount ? 0 : pagination.limit,
              [pageEnableCountKey]: isCount,
              ...sortVal,
            },
          }),
        );
      }),
    );
  } else {
    [detailsRes, countRes] = await defaultFetchData(params);
  }

  return [detailsRes, countRes];
};
