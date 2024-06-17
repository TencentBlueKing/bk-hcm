export const fetchData = async (params: any) => {
  const { api, props, pagination, sort, order, filter, type } = params;
  const { requestOption } = props;
  const { apiMethod, extension, full } = requestOption;

  // type api
  if (requestOption.type) {
    // 判断是业务下, 还是资源下
    const fetchApi = async (page: any) =>
      api({ page, filter: { op: filter.op, rules: filter.rules }, ...extension }, type || requestOption.type);

    // 请求数据
    return await Promise.all([
      fetchApi({
        limit: pagination.limit,
        start: pagination.start,
        sort: sort.value,
        order: order.value,
        count: false,
      }),
      fetchApi({ count: true }),
    ]);
  }

  // apiMethod api
  if (full) {
    // 非分页请求
    return [await apiMethod(extension), null];
  }

  const fetchApi = async (page: any) =>
    apiMethod({ page, filter: { op: filter.op, rules: filter.rules }, ...extension });

  // 分页请求
  return await Promise.all([
    fetchApi({
      limit: pagination.limit,
      start: pagination.start,
      sort: sort.value,
      order: order.value,
      count: false,
    }),
    fetchApi({ count: true }),
  ]);
};
