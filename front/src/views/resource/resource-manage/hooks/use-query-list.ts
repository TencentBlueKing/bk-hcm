/**
 * 分页相关状态和事件
 */
import type { FilterType } from '@/typings/resource';

import {
  useResourceStore,
} from '@/store/resource';
import {
  ref,
  // onMounted,
  watch,
} from 'vue';

type SortType = {
  column: {
    prop: string
  };
  type: string
};
type PropsType = {
  filter?: FilterType
};

export default async (props: PropsType, type: string) => {
  // 接口
  const resourceStore = useResourceStore();

  // 查询列表相关状态
  const isLoading = ref(false);
  const datas = ref([]);
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0,
  });
  const sort = ref();
  const order = ref();

  // 更新数据
  const triggerApi = async () => {
    console.log(222);
    isLoading.value = true;
    const listCount = await resourceStore
      .list(
        {
          page: {
            count: true,
          },
          filter: props.filter,
        },
        type,
      );
    pagination.value.count = listCount.data.count;
    const listData = await resourceStore
      .list(
        {
          page: {
            count: false,
            start: (pagination.value.current - 1) * pagination.value.limit,
            limit: pagination.value.limit,
            sort: sort.value,
            order: order.value,
          },
          filter: props.filter,
        },
        type,
      );
    datas.value = (listData.data.details || []).map((item: any) => {
      return {
        ...item,
        ...item.spec,
        ...item.attachment,
        ...item.revision,
      };
    });
    isLoading.value = false;
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
    sort.value = column.prop;
    order.value = type;
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
  watch(() => type, async (value) => {
    console.log('value', value);
    triggerApi();
  });

  await triggerApi();
  return {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  };
};
