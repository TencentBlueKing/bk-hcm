import { Ref, reactive, ref } from 'vue';
// import stores
import { useResourceStore } from '@/store';
// import types
import { QueryRuleOPEnum, RulesItem } from '@/typings';

/**
 * Select Option List - 支持滚动加载
 */
export default (
  // 加载的资源类型
  type: string,
  // 搜索条件
  rules: any,
  // 是否立即执行
  immediate = true,
): [
  isLoading: Ref<boolean>,
  optionList: Ref<any[]>,
  initState: (...args: any) => any,
  getOptionList: (...args: any) => any,
  handleOptionListScrollEnd: (...args: any) => any,
  isFlashLoading: Ref<boolean>,
  handleRefreshOptionList: (...args: any) => any,
] => {
  // use stores
  const resourceStore = useResourceStore();

  // define data
  const pagination = reactive({
    start: 0,
    limit: 20,
    hasNext: true,
  });
  const isLoading = ref(false);
  const optionList = ref([]);
  const isFlashLoading = ref(false); // 刷新操作loading

  /**
   * 初始化状态: optionList, pagination
   */
  const initState = () => {
    optionList.value = [];
    Object.assign(pagination, {
      start: 0,
      limit: 20,
      hasNext: true,
    });
  };

  /**
   * 请求option list
   */
  const getOptionList = async (customRules: RulesItem[] = []) => {
    isLoading.value = true;
    try {
      const [detailRes, countRes] = await Promise.all(
        [false, true].map((isCount) =>
          resourceStore.list(
            {
              filter: {
                op: QueryRuleOPEnum.AND,
                rules: [...rules, ...customRules],
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

      // 将新获取的option添加至list中
      optionList.value = [...optionList.value, ...detailRes.data.details];
      if (optionList.value.length === countRes.data.count) {
        // option列表加载完毕
        pagination.hasNext = false;
      } else {
        // option列表未加载完毕
        pagination.start += pagination.limit;
      }
    } finally {
      isLoading.value = false;
    }
  };

  /**
   * 滚动触底加载更多
   */
  const handleOptionListScrollEnd = () => {
    if (!pagination.hasNext) return;
    getOptionList();
  };

  /**
   * 刷新options list
   */
  const handleRefreshOptionList = async () => {
    initState();
    try {
      isFlashLoading.value = true;
      await getOptionList();
    } finally {
      isFlashLoading.value = false;
    }
  };

  // 立即执行
  immediate && getOptionList();

  return [
    isLoading,
    optionList,
    initState,
    getOptionList,
    handleOptionListScrollEnd,
    isFlashLoading,
    handleRefreshOptionList,
  ];
};
