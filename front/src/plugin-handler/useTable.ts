export const fetchData = async (params: any) => {
  const { api, pagination, sort, order, filter, props, type } = params;

  // 请求数据
  const [detailsRes, countRes] = await Promise.all(
    [false, true].map((isCount) =>
      api(
        {
          page: {
            limit: isCount ? 0 : pagination.limit,
            start: isCount ? 0 : pagination.start,
            sort: isCount ? undefined : sort.value,
            order: isCount ? undefined : order.value,
            count: isCount,
          },
          filter: { op: filter.op, rules: filter.rules },
          ...props.requestOption.extension,
        },
        type ? type : props.requestOption.type,
      ),
    ),
  );

  return [detailsRes, countRes];
};
