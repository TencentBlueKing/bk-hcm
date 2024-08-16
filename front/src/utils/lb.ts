/**
 * 统一管理负载均衡需求下的 utils 函数
 */

/**
 * 异步加载负载均衡的监听器数量
 * @param lbList 负载均衡列表
 * @returns 负载均衡列表(带有listenerNum字段)
 */
const asyncGetListenerCount = async (api: (data: { lb_ids: string[] }) => any, lbList: any) => {
  // 如果lbList长度为0, 则无需请求监听器数量
  if (lbList.length === 0) return;
  // 负载均衡ids
  const lb_ids = lbList.map(({ id }: { id: string }) => id);
  // 查询负载均衡对应的监听器数量
  const res = await api({ lb_ids });
  // 构建映射关系
  const listenerCountMap = {};
  res.data.details.forEach(({ lb_id, num }: { lb_id: string; num: number }) => {
    listenerCountMap[lb_id] = num;
  });
  // 根据映射关系进行匹配, 将 num 添加到 lbList 中并返回
  return lbList.map((data: any) => {
    const { id } = data;
    return { ...data, listenerNum: listenerCountMap[id] };
  });
};

/**
 * 获取search-select组合后的过滤条件, 可用于本地表格数据的过滤
 * @param searchVal search-select 的值
 * @param resolveRuleValue 用于处理规则值的函数(比如中英文映射...)
 * @returns 过滤条件
 */
const getLocalFilterConditions = (searchVal: any[], resolveRuleValue: (rule: any) => void) => {
  if (!searchVal || searchVal.length === 0) return {};
  const filterConditions = {};
  searchVal.forEach((rule: any) => {
    // 获取规则值
    const ruleValue = resolveRuleValue(rule);

    // 组装过滤条件
    if (filterConditions[rule.id]) {
      // 如果 filterConditions[rule.id] 已经存在，则合并为一个数组
      filterConditions[rule.id] = [...filterConditions[rule.id], ruleValue];
    } else {
      filterConditions[rule.id] = [ruleValue];
    }
  });
  return filterConditions;
};

/**
 * 当负载均衡被锁定时, 可以链接至异步任务详情页面
 * @param flowId 异步任务id
 */
const goAsyncTaskDetail = async (api: (data: any, type: string) => any, flowId: string) => {
  // 1. 点击后, 先查询到 audit_id
  const { data } = await api(
    {
      page: { limit: 1, start: 0, count: false },
      filter: {
        op: 'and',
        rules: [{ field: 'detail.data.res_flow.flow_id', op: 'json_eq', value: flowId }],
      },
    },
    'audits',
  );
  const { id, res_name: name, res_id, bk_biz_id } = data.details[0];
  // 2. 新开页面查看异步任务详情
  window.open(
    `/#/business/record/detail?id=${id}&name=${name}&res_id=${res_id}&bizs=${bk_biz_id}&flow=${flowId}`,
    '_blank',
  );
};

export { asyncGetListenerCount, getLocalFilterConditions, goAsyncTaskDetail };
