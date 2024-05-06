import { reactive, ref } from 'vue';
import { useBusinessStore, useResourceStore } from '@/store';
import { useWhereAmI, Senarios } from './useWhereAmI';
import { QueryRuleOPEnum, RulesItem } from '@/typings';

export default (type: string, rules: RulesItem[] | ((...args: any) => RulesItem[]), immediate = true) => {
  // 业务下使用 businessStore, 资源下使用 resourceStore
  const resourceStore = useResourceStore();
  const businessStore = useBusinessStore();
  const { whereAmI } = useWhereAmI();

  // 列表数据
  const list = ref([]);
  // 分页信息
  const pagination = reactive({
    start: 0,
    limit: 50,
    count: 0,
  });

  // 滚动加载loading
  const isLoading = ref(false);

  /**
   * 获取列表数据
   * @param customRules 自定义查询规则
   */
  const getList = async (customRules: RulesItem[] = []) => {
    const method = whereAmI.value === Senarios.business ? businessStore.list : resourceStore.list;
    isLoading.value = true;
    try {
      const [detailRes, countRes] = await Promise.all(
        [false, true].map((isCount) =>
          method(
            {
              filter: {
                op: QueryRuleOPEnum.AND,
                rules: [...(typeof rules === 'function' ? rules() : rules), ...customRules],
              },
              page: {
                count: isCount,
                start: isCount ? 0 : pagination.start,
                limit: isCount ? 0 : pagination.limit,
              },
            },
            type,
          ),
        ),
      );
      list.value = [...list.value, ...detailRes.data.details];
      pagination.count = countRes.data.count;
    } finally {
      isLoading.value = false;
    }
  };

  /**
   * 滚动触底
   */
  const handleScrollEnd = async () => {
    // 判断是否有下一页数据
    if (list.value.length >= pagination.count) return;
    // 累加 start
    pagination.start += pagination.limit;
    // 获取数据
    await getList();
  };

  /**
   * 重置
   */
  const reset = () => {
    list.value = [];
    Object.assign(pagination, {
      start: 0,
      limit: 50,
      count: 0,
    });
  };

  /**
   * 刷新列表数据
   */
  const refresh = async () => {
    // 重置
    reset();
    await getList();
  };

  immediate && getList();

  return {
    list,
    pagination,
    getList,
    handleScrollEnd,
    reset,
    refresh,
    isLoading,
  };
};
