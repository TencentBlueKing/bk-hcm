import { ref } from 'vue';
import usePagination from './usePagination';
import http from '@/http';
import { defaults } from 'lodash';
import { QueryRuleOPEnum, RulesItem } from '@/typings';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default function useRemoteTable(
  url: string | (() => string),
  options?: {
    rules?: RulesItem[] | (() => RulesItem[]);
    sortOption?: { sort: string; order: 'ASC' | 'DESC' };
    extApi?: {
      url: string | (() => string);
      rules?: RulesItem[] | ((dataList: any[]) => RulesItem[]);
    };
    immediate?: boolean;
  },
) {
  // 设置默认参数
  defaults(options, { extApi: null, rules: [], immediate: false });

  const isLoading = ref(false);
  const dataList = ref([]);
  const { pagination, handlePageLimitChange, handlePageValueChange } = usePagination(() => getDataList());
  const sort = ref(options.sortOption ? options.sortOption.sort : 'created_at');
  const order = ref(options.sortOption ? options.sortOption.order : 'DESC');

  const buildApiMethod = (fetchUrl: string, rules: RulesItem[]) => {
    return Promise.all(
      [false, true].map((isCount) =>
        http.post(fetchUrl, {
          filter: { op: QueryRuleOPEnum.AND, rules },
          page: {
            limit: isCount ? 0 : pagination.limit,
            start: isCount ? 0 : pagination.start,
            sort: isCount ? undefined : sort.value,
            order: isCount ? undefined : order.value,
            count: isCount,
          },
        }),
      ),
    )
      .then(([detailsRes, countRes]) => {
        dataList.value = detailsRes.data.details;
        pagination.count = countRes.data.count;
        return dataList.value;
      })
      .finally(() => {
        isLoading.value = false;
      });
  };

  const getDataList = () => {
    isLoading.value = true;

    // 请求基础数据
    const promise = buildApiMethod(
      BK_HCM_AJAX_URL_PREFIX + (typeof url === 'string' ? url : url()),
      typeof options.rules === 'function' ? options.rules() : options.rules,
    );

    // 如果存在扩展接口, 则根据基础数据再次请求
    if (options?.extApi) {
      promise.then((dataList) => {
        buildApiMethod(
          BK_HCM_AJAX_URL_PREFIX +
            (typeof options?.extApi.url === 'string' ? options?.extApi.url : options?.extApi.url()),
          typeof options?.extApi.rules === 'function' ? options?.extApi.rules(dataList) : options?.extApi.rules,
        );
      });
    }

    return promise;
  };

  // 表头排序
  const handleSort = ({ column, type }: any) => {
    pagination.start = 0;
    sort.value = column.field;
    order.value = type === 'asc' ? 'ASC' : 'DESC';
    // 如果type为null，则默认排序
    if (type === 'null') {
      sort.value = options.sortOption ? options.sortOption.sort : 'created_at';
      order.value = options.sortOption ? options.sortOption.order : 'DESC';
    }
    buildApiMethod(
      BK_HCM_AJAX_URL_PREFIX + (typeof options?.extApi.url === 'string' ? options?.extApi.url : options?.extApi.url()),
      // 此处的排序条件, 应根据目标渲染数据(比如cvm)重新构造
      [{ op: QueryRuleOPEnum.IN, field: 'id', value: dataList.value.map((item) => item.id) }],
    );
  };

  options?.immediate && getDataList();

  return {
    isLoading,
    dataList,
    getDataList,
    pagination,
    handlePageLimitChange,
    handlePageValueChange,
    handleSort,
  };
}
