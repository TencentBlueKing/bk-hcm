import { Ref, reactive, ref } from 'vue';
// import stores
import { useResourceStore } from '@/store';
// import types
import { QueryRuleOPEnum } from '@/typings';

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
  getOptionList: (...args: any) => any,
  handleOptionListScrollEnd: (...args: any) => any,
] => {
  // use stores
  const resourceStore = useResourceStore();

  // define data
  const pagination = reactive({
    start: 0,
    limit: 10,
    hasNext: true,
  });
  const isLoading = ref(false);
  const optionList = ref([]);

  // get option list
  const getOptionList = async () => {
    isLoading.value = true;
    try {
      const [detailRes, countRes] = await Promise.all(
        [false, true].map((isCount) =>
          resourceStore.list(
            {
              filter: {
                op: QueryRuleOPEnum.AND,
                rules,
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

  // handler - scroll end
  const handleOptionListScrollEnd = () => {
    if (!pagination.hasNext) return;
    getOptionList();
  };

  // init
  immediate && getOptionList();

  return [isLoading, optionList, getOptionList, handleOptionListScrollEnd];
};
