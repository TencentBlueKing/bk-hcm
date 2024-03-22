import { IPageQuery, QueryRuleOPEnum } from '@/typings';
import { defineStore } from 'pinia';
import { reactive, ref, Ref } from 'vue';
import { useResourceStore } from './resource';

export interface ILoadBalancer {
  id: string; // 负载均衡资源ID
  account_id: string; // 关联的账号ID
}

export const useLoadBalancerStore = defineStore('load-balancer', () => {
  const targetGroupListPageQuery = reactive<IPageQuery>({
    start: 0,
    limit: 50,
  });
  const resourceStore = useResourceStore();

  // state - 目标组id
  const targetGroupId = ref('');
  const allTargetGroupList = ref([]);
  // state - lb-tree - 当前选中的资源
  const currentSelectedTreeNode = ref();
  // state - 目标组id
  const lb: Ref<ILoadBalancer> = ref({
    id: '',
    account_id: '',
  });

  // action - lb-tree - 设置当前选中的资源
  const setCurrentSelectedTreeNode = (node: any) => {
    // 其中, node 可能为 lb, listener, domain 节点
    currentSelectedTreeNode.value = node;
  };
  // action - 设置目标组id
  const setTargetGroupId = (v: string) => {
    targetGroupId.value = v;
  };

  const getTargetGroupList = async () => {
    const [detailRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        resourceStore.list(
          {
            filter: {
              op: QueryRuleOPEnum.AND,
              rules: [],
            },
            page: {
              count: isCount,
              start: isCount ? 0 : targetGroupListPageQuery.start,
              limit: isCount ? 0 : targetGroupListPageQuery.limit,
            },
          },
          'target_groups',
        ),
      ),
    );
    allTargetGroupList.value = detailRes.data.details;
    targetGroupListPageQuery.count = countRes.data.count;
  };

  const setLB = (obj: ILoadBalancer) => {
    lb.value = obj;
  };

  return {
    targetGroupId,
    setTargetGroupId,
    lb,
    setLB,
    currentSelectedTreeNode,
    setCurrentSelectedTreeNode,
    getTargetGroupList,
    allTargetGroupList,
  };
});
