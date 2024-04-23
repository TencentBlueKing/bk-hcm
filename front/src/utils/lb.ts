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

export { asyncGetListenerCount };
