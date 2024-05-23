import type { QueryFilterType } from '@/typings';
import { QueryRuleOPEnum } from '@/typings';
import { useAuditStore } from '@/store/audit';
import { ref } from 'vue';

type SortType = {
  column: {
    field: string;
  };
  type: string;
};

export default (options: { filter: any; filterOptions: any }) => {
  // 接口
  const auditStore = useAuditStore();

  // 查询列表相关状态
  const isLoading = ref(false);
  const datas = ref([]);
  const pagination = ref({
    current: 1,
    limit: 10,
    count: 0,
  });
  const sort = ref('created_at');
  const order = ref('DESC');

  // 更新数据
  const query = () => {
    const { filter, filterOptions } = options;

    if (filterOptions.auditType === 'biz' && !filter.bk_biz_id) {
      return;
    }

    isLoading.value = true;
    filter.res_id = '';
    filter.res_name = '';
    if (filterOptions.instType === 'id') {
      filter.res_id = filterOptions.instValue;
    }
    if (filterOptions.instType === 'name') {
      filter.res_name = filterOptions.instValue;
    }

    const filterIds = Object.keys(filter);
    const queryFilter: QueryFilterType = {
      op: 'and',
      rules: [],
    };
    for (let i = 0, key; (key = filterIds[i]); i++) {
      const value = filter[key];
      if (!value || !value?.length || (Array.isArray(value) && !value.every((val) => val))) {
        continue;
      }

      if (key === 'created_at') {
        queryFilter.rules.push(
          {
            field: key,
            op: QueryRuleOPEnum.GTE,
            value: new Date(value[0]).toISOString().replace('.000Z', 'Z'),
          },
          {
            field: key,
            op: QueryRuleOPEnum.LTE,
            value: new Date(value[1]).toISOString().replace('.000Z', 'Z'),
          },
        );
        continue;
      }

      if (key === 'res_id' || key === 'res_name') {
        queryFilter.rules.push({
          field: key,
          op: filterOptions.instFuzzy ? QueryRuleOPEnum.CIS : QueryRuleOPEnum.EQ,
          value,
        });
        continue;
      }

      queryFilter.rules.push({
        field: key,
        op: Array.isArray(value) ? QueryRuleOPEnum.IN : QueryRuleOPEnum.EQ,
        value,
      });
    }

    Promise.all([
      auditStore.list(
        {
          page: {
            count: false,
            start: (pagination.value.current - 1) * pagination.value.limit,
            limit: pagination.value.limit,
            sort: sort.value,
            order: order.value,
          },
          filter: queryFilter,
        },
        filter.bk_biz_id,
      ),
      auditStore.list(
        {
          page: {
            count: true,
          },
          filter: queryFilter,
        },
        filter.bk_biz_id,
      ),
    ])
      .then(([listResult, countResult]) => {
        datas.value = listResult?.data?.details || [];
        pagination.value.count = countResult?.data?.count || 0;
      })
      .finally(() => {
        isLoading.value = false;
      });
  };

  // 页码变化发生的事件
  const handlePageChange = (current: number) => {
    pagination.value.current = current;
    query();
  };

  // 条数变化发生的事件
  const handlePageSizeChange = (limit: number) => {
    pagination.value.limit = limit;
    query();
  };

  // 排序变化发生的事件
  const handleSort = ({ column, type }: SortType) => {
    pagination.value.current = 1;
    sort.value = column.field;
    order.value = type === 'desc' ? 'DESC' : 'ASC';
    query();
  };

  return {
    query,
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  };
};
