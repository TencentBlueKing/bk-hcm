import { IPageQuery, QueryRuleOPEnum } from '@/typings';
import { defineStore } from 'pinia';
import { reactive, ref } from 'vue';
import { useResourceStore } from './resource';

export const useLoadBalancerStore = defineStore('load-balancer', () => {
  const resourceStore = useResourceStore();

  // state - 目标组id
  const targetGroupId = ref('');
  const setTargetGroupId = (v: string) => {
    targetGroupId.value = v;
  };

  // state - lb-tree - 当前选中的资源
  const currentSelectedTreeNode = ref();
  const setCurrentSelectedTreeNode = (node: any) => {
    // 其中, node 可能为 lb, listener, domain 节点
    currentSelectedTreeNode.value = node;
  };

  // state - 目标组操作场景
  const currentScene = ref('');
  const setCurrentScene = (v: string) => {
    currentScene.value = v;
  };

  // state - 目标组左侧列表
  const allTargetGroupList = ref([]);
  const targetGroupListPageQuery = reactive<IPageQuery>({
    start: 0,
    limit: 50,
  });
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
    currentScene,
    setCurrentScene,
  };
});
