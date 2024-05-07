/**
 * 统一管理负载均衡需求下的 utils 函数
 */

// import stores
import { useBusinessStore } from '@/store';
const businessStore = useBusinessStore();

/**
 * 异步加载负载均衡的监听器数量
 * @param lbList 负载均衡列表
 * @returns 负载均衡列表(带有listenerNum字段)
 */
const asyncGetListenerCount = async (lbList: any) => {
  // 如果lbList长度为0, 则无需请求监听器数量
  if (lbList.length === 0) return;
  // 负载均衡ids
  const lb_ids = lbList.map(({ id }: { id: string }) => id);
  // 查询负载均衡对应的监听器数量
  const res = await businessStore.asyncGetListenerCount({ lb_ids });
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
 * 获取负载均衡的ip地址
 * @param lb 负载均衡
 * @returns 负载均衡的ip地址
 */
const getLbVip = (lb: any) => {
  const { private_ipv4_addresses, private_ipv6_addresses, public_ipv4_addresses, public_ipv6_addresses } = lb;
  if (public_ipv4_addresses.length > 0) return public_ipv4_addresses.join(',');
  if (public_ipv6_addresses.length > 0) return public_ipv6_addresses.join(',');
  if (private_ipv4_addresses.length > 0) return private_ipv4_addresses.join(',');
  if (private_ipv6_addresses.length > 0) return private_ipv6_addresses.join(',');
  return '--';
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

export { asyncGetListenerCount, getLbVip, getLocalFilterConditions };
