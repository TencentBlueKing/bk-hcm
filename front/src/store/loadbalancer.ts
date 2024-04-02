import { IPageQuery, QueryRuleOPEnum } from '@/typings';
import { defineStore } from 'pinia';
import { Ref, reactive, ref } from 'vue';
import { useResourceStore } from './resource';

export interface ITreeNode {
  [key: string]: any;
  lb: Record<string, any>;  // 当前域名节点所属的负载均衡信息，非域名节点时不生效
  listener: Record<string, any>; // 当前域名节点所属的监听器信息，非域名节点时不生效
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
  const currentSelectedTreeNode: Ref<ITreeNode> = ref({} as ITreeNode);
  // state - 目标组id

  // action - lb-tree - 设置当前选中的资源
  const setCurrentSelectedTreeNode = (node: ITreeNode) => {
    // 其中, node 可能为 lb, listener, domain 节点
    currentSelectedTreeNode.value = node;
    console.log(666666, currentSelectedTreeNode.value);
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

  return {
    targetGroupId,
    setTargetGroupId,
    currentSelectedTreeNode,
    setCurrentSelectedTreeNode,
    getTargetGroupList,
    allTargetGroupList,
  };
});
