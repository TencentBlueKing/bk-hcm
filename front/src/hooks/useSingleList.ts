import { reactive, ref, watch } from 'vue';
import { defaults, get } from 'lodash';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import http from '@/http';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

/**
 * hooks 函数 - 适用于单列表
 * @param options 配置项, 如 url, immediate, rules, apiMethod 等...
 * @returns dataList, pagination, isDataLoad, isDataRefresh, loadDataList, handleScrollEnd, handleReset, handleRefresh
 */
export function useSingleList(options?: {
  // 请求url（固定的 url 可以使用字符串形式；path带参数的 url 用函数形式，如 :id, :vendor 等）
  url: string | ((...args: any) => string);
  // 是否立即加载数据, 值为 true 时可以省略在 onMounted 中加载初始数据
  immediate?: boolean;
  // 初始搜索条件（数组参数形式推荐 id 等不变值搜索条件；函数参数形式可以支持响应式的搜索条件）
  rules?: RulesItem[] | ((...args: any) => RulesItem[]);
  // 请求载荷
  payload?: object | ((...args: any) => object);
  // 自定义的 api 方法（请求全量数据时使用）
  apiMethod?: (...args: any) => Promise<any[]>;
  // 自定义参数路径
  path?: { data: string; count: string };
  // 禁用排序
  disableSort?: boolean;
  // 初始分页参数
  pagination?: { start?: number; limit: number; count?: number };
}) {
  // 设置 options 默认值
  defaults(options, {
    immediate: false,
    rules: [],
    payload: {},
    apiMethod: null,
    path: {},
    pagination: {},
    disableSort: false,
  });
  // 设置 path 默认值
  defaults(options.path, { data: 'details', count: 'count' });
  // 设置 pagination 默认值
  defaults(options.pagination, { start: 0, limit: 50, count: 0 });

  const getDefaultPagination = () => options.pagination;

  const dataList = ref([]);
  const pagination = reactive(getDefaultPagination());
  const isDataLoad = ref(false);
  const isDataRefresh = ref(false);

  const loadDataList = (customRules: RulesItem[] = []) => {
    /**
     * @returns 返回一个默认的 list api
     */
    const getDefaultApiMethod = () => {
      const url = typeof options?.url === 'function' ? options?.url() : options?.url;

      return Promise.all(
        [false, true].map((isCount) =>
          http.post(BK_HCM_AJAX_URL_PREFIX + url, {
            filter: {
              op: QueryRuleOPEnum.AND,
              rules: [...(typeof options?.rules === 'function' ? options?.rules() : options?.rules), ...customRules],
            },
            page: {
              count: isCount,
              start: isCount ? 0 : pagination.start,
              limit: isCount ? 0 : pagination.limit,
              ...(options.disableSort
                ? {}
                : { sort: isCount ? undefined : 'created_at', order: isCount ? undefined : 'DESC' }),
            },
            ...(typeof options.payload === 'function' ? options.payload() : options.payload),
          }),
        ),
      );
    };

    // 选择 api
    const apiMethod = options?.apiMethod || getDefaultApiMethod;

    // 开启 loading 效果
    isDataLoad.value = true;
    return apiMethod()
      .then(([detailRes, countRes]) => {
        // 加载数据
        dataList.value = [...dataList.value, ...get(detailRes.data, options.path.data, [])];
        // 更新分页参数
        pagination.count = get(countRes.data, options.path.count, 0);
        // 将加载数据后的 dataList 作为 then 函数的返回值, 用以支持对新的 dataList 做额外的处理
        return dataList.value;
      })
      .finally(() => {
        // 关闭 loading 效果
        isDataLoad.value = false;
      });
  };

  const handleScrollEnd = () => {
    // 判断是否有下一页数据
    if (dataList.value.length >= pagination.count) return;
    // 累加 start
    pagination.start += pagination.limit;
    // 获取数据
    loadDataList();
  };

  const handleReset = () => {
    dataList.value = [];
    Object.assign(pagination, getDefaultPagination());
  };

  const handleRefresh = async () => {
    handleReset();
    isDataRefresh.value = true;
    await loadDataList();
    isDataRefresh.value = false;
  };

  // 立即执行
  options?.immediate && loadDataList();

  // url 变更时, 刷新列表
  watch(
    () => options?.url,
    () => {
      handleRefresh();
    },
  );

  return {
    /**
     * 数据列表
     */
    dataList,
    /**
     * 分页参数
     */
    pagination,
    /**
     * loading - 数据加载
     */
    isDataLoad,
    /**
     * loading - 数据刷新
     */
    isDataRefresh,
    /**
     * 加载数据
     * @param customRules 自定义查询规则
     * @returns 返回一个 Promise 实例, 其成功状态下的结果值为加载数据后的 dataList.
     */
    loadDataList,
    /**
     * 滚动触底 - 加载数据
     */
    handleScrollEnd,
    /**
     * 数据重置
     */
    handleReset,
    /**
     * 数据刷新
     */
    handleRefresh,
  };
}
