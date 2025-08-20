/**
 * 分页相关状态和事件
 */
import type { FilterType } from '@/typings/resource';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import { useBusinessStore, useResourceStore } from '@/store';
import { Ref, ref, watch } from 'vue';
type SortType = {
  column: {
    field: string;
  };
  type: string;
};
type PropsType = {
  filter?: FilterType;
};
type ExtraConfigType = {
  sort?: 'string';
  order?: 'ASC' | 'DESC';
  asyncRequestApiMethod?: (datalist: any[], datalistRef: Ref<any[]>) => void; // 处理异步请求
};

export default (props: PropsType, url: Ref<string>, extraConfig?: ExtraConfigType) => {
  // 接口
  const resourceStore = useResourceStore();
  const businessStore = useBusinessStore();
  const { whereAmI } = useWhereAmI();
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
  const triggerApi = async () => {
    const method = whereAmI.value === Senarios.business ? businessStore.getCommonList : resourceStore.getCommonList;

    isLoading.value = true;
    try {
      const [listResult, countResult] = await Promise.all([
        method(
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
          { cancelPrevious: true },
        ),
        method({ page: { count: true }, filter: props.filter }, url.value, { cancelPrevious: true }),
      ]);

      const { details = [] } = listResult.data;
      const { count = 0 } = countResult.data;

      const displayDatalist =
        details?.map((item: any) => ({
          ...item,
          ...item.spec,
          ...item.attachment,
          ...item.revision,
          ...item.extension,
        })) ?? [];

      // 基础list请求
      datas.value = displayDatalist;
      pagination.value.count = count;

      // 异步请求方法
      if (extraConfig?.asyncRequestApiMethod) {
        extraConfig.asyncRequestApiMethod(displayDatalist, datas);
      }

      return details;
    } catch (error) {
      console.error(error);
      datas.value = [];
      pagination.value.count = 0;
    } finally {
      isLoading.value = false;
    }
  };

  // 页码变化发生的事件
  const handlePageChange = (current: number) => {
    pagination.value.current = current;
    triggerApi();
  };

  // 条数变化发生的事件
  const handlePageSizeChange = (limit: number) => {
    pagination.value.current = 1;
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

  watch(
    [() => props.filter, () => url],
    () => {
      pagination.value.current = 1; // 页码重置
      triggerApi();
    },
    {
      deep: true,
      flush: 'post', // DOM更新后执行
    },
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
