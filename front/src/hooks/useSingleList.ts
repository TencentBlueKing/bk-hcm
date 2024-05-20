import { reactive, ref } from 'vue';
import { useBusinessStore, useResourceStore } from '@/store';
import { defaults } from 'lodash';
import { Senarios, useWhereAmI } from './useWhereAmI';
import { QueryRuleOPEnum, RulesItem } from '@/typings';

/**
 * hooks 函数 - 适用于单列表
 * @param type 请求的资源类型, 如 cvm, drive, vpc 等...
 * @param options 其他配置项, 如 immediate, rules, apiMethod 等...
 * @returns dataList, pagination, isDataLoad, isDataRefresh, loadDataList, handleScrollEnd, handleReset, handleRefresh
 */
export function useSingleList(
  type: string,
  options?: {
    // 是否立即加载数据, 值为 true 时可以省略在 onMounted 中加载初始数据
    immediate?: boolean;
    // 初始搜索条件, 数组参数形式推荐 id 等不变值搜索条件, 函数参数形式可以支持响应式的搜索条件
    rules?: RulesItem[] | ((...args: any) => RulesItem[]);
    // 自定义的 api 方法, 可以替换默认的 list api
    apiMethod?: (...args: any) => Promise<any[]>;
  },
) {
  // 设置 options 默认值
  defaults(options, { immediate: false, rules: [], apiMethod: null });
  const getDefaultPagination = () => ({ start: 0, limit: 50, count: 0 });

  // 引入可能会用到的 hooks 函数以及 store
  const { whereAmI } = useWhereAmI();
  const resourceStore = useResourceStore();
  const businessStore = useBusinessStore();

  const dataList = ref([]);
  const pagination = reactive(getDefaultPagination());
  const isDataLoad = ref(false);
  const isDataRefresh = ref(false);

  const loadDataList = (customRules: RulesItem[] = []) => {
    /**
     * @returns 返回一个默认的 list api
     */
    const getDefaultApiMethod = () => {
      const apiMethod = whereAmI.value === Senarios.business ? businessStore.list : resourceStore.list;
      return Promise.all(
        [false, true].map((isCount) =>
          apiMethod(
            {
              filter: {
                op: QueryRuleOPEnum.AND,
                rules: [...(typeof options?.rules === 'function' ? options?.rules() : options?.rules), ...customRules],
              },
              page: {
                count: isCount,
                start: isCount ? 0 : pagination.start,
                limit: isCount ? 0 : pagination.limit,
                sort: isCount ? undefined : 'created_at',
                order: isCount ? undefined : 'DESC',
              },
            },
            type,
          ),
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
        dataList.value = [...dataList.value, ...detailRes.data.details];
        // 更新分页参数
        pagination.count = countRes.data.count;
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
