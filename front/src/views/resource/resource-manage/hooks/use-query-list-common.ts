/**
 * 分页相关状态和事件
 */
import type { FilterType } from '@/typings/resource';

import {
  useResourceStore,
} from '@/store/resource';
import {
  Ref,
  ref,
  watch,
} from 'vue';

type SortType = {
  column: {
    field: string
  };
  type: string
};
type PropsType = {
  filter?: FilterType
};

export default (props: PropsType, url: Ref<string>, extraConfig?: any) => {
  // 接口
  const resourceStore = useResourceStore();

  // 查询列表相关状态
  const isLoading = ref(false);
  const datas = ref([]);
  const pagination = ref({
    current: 1,
    limit: 20,
    count: 0,
  });
  const sort = ref('created_at');
  const order = ref('DESC');

  // 更新数据
  const triggerApi = () => {
    isLoading.value = true;

    Promise
      .all([
        resourceStore.getCommonList(
          {
            page: {
              count: false,
              start: (pagination.value.current - 1) * pagination.value.limit,
              limit: pagination.value.limit,
              sort: extraConfig?.sort ? extraConfig.sort : sort.value,
              order: extraConfig?.order ? extraConfig.order : order.value,
            },
            filter: props.filter,
          },
          url.value,
        ),
        resourceStore.getCommonList(
          {
            page: {
              count: true,
            },
            filter: props.filter,
          },
          url.value,
        ),
      ]).then(([listResult, countResult]: [any, any]) => {
        datas.value = (listResult?.data?.details || []).map((item: any) => {
          return {
            ...item,
            ...item.spec,
            ...item.attachment,
            ...item.revision,
            ...item.extension,
          };
        });
        pagination.value.count = countResult?.data?.count || 0;
      })
      .finally(() => {
        isLoading.value = false;
      });
  };

  // 页码变化发生的事件
  const handlePageChange = (current: number) => {
    pagination.value.current = current;
    triggerApi();
  };

  // 条数变化发生的事件
  const handlePageSizeChange = (limit: number) => {
    pagination.value.limit = limit;
    triggerApi();
  };

  // 排序变化发生的事件
  const handleSort = ({ column, type }: SortType) => {
    pagination.value.current = 1;
    sort.value = column.field;
    order.value = type === 'desc' ? 'DESC' : 'ASC';
    triggerApi();
  };

  // 过滤发生变化的时候，获取数据
  watch(
    () => props.filter,
    triggerApi,
    {
      deep: true,
    },
  );

  // 切换tab重新获取数据
  watch(
    () => url,
    () => {
      triggerApi();
    },
    { deep: true },
  );


  const getList = () => {
    triggerApi();
  };

  return {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
    getList,
  };
};
