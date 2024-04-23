import { QueryRuleOPEnum } from '@/typings';
import { defineStore } from 'pinia';
import { Ref, reactive, ref } from 'vue';
import { useResourceStore } from './resource';

export interface ITreeNode {
  [key: string]: any;
  lb: Record<string, any>; // 当前域名节点所属的负载均衡信息，非域名节点时不生效
  listener: Record<string, any>; // 当前域名节点所属的监听器信息，非域名节点时不生效
}
// 目标组视角 - 操作场景
export type TGOperationScene =
  | 'add' // 新建目标组
  | 'edit' // 编辑目标组基本信息
  | 'BatchDelete' // 批量删除目标组
  | 'AddRs' // 添加rs
  | 'BatchAddRs' // 批量添加rs
  | 'BatchDeleteRs' // 批量删除rs
  | 'port' // 批量修改端口
  | 'weight'; // 批量修改权重

export const useLoadBalancerStore = defineStore('load-balancer', () => {
  const resourceStore = useResourceStore();

  // state - 目标组id
  const targetGroupId = ref('');
  const setTargetGroupId = (v: string) => {
    targetGroupId.value = v;
  };

  // state - lb-tree - 当前选中的资源
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
  const currentSelectedTreeNode: Ref<ITreeNode> = ref({} as ITreeNode);
  const setCurrentSelectedTreeNode = (node: ITreeNode) => {
    // 其中, node 可能为 lb, listener, domain 节点
    currentSelectedTreeNode.value = node;
  };

  // state - 目标组操作场景
  const updateCount = ref(0); // 记录修改次数: 当值为2时, 重新对场景进行判断(判断第一次回显数据的影响)
  const setUpdateCount = (v: number) => {
    updateCount.value = v;
  };
  const currentScene = ref<TGOperationScene>(); // 记录当前操作类型
  const setCurrentScene = (v: TGOperationScene) => {
    currentScene.value = v;
  };

  // state - 目标组左侧列表
  const allTargetGroupList = ref([]);
  const targetGroupListPageQuery = reactive({
    start: 0,
    limit: 50,
    count: 0,
  });
  const getTargetGroupList = async () => {
    // 获取数据
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
    allTargetGroupList.value = [...allTargetGroupList.value, ...detailRes.data.details];
    targetGroupListPageQuery.count = countRes.data.count;
  };
  // 目标组list滚动触底, 获取下一页数据
  const getNextTargetGroupList = async () => {
    // 判断是否有下一页数据
    if (allTargetGroupList.value.length >= targetGroupListPageQuery.count) return;
    // 累加 start
    targetGroupListPageQuery.start += targetGroupListPageQuery.limit;
    // 获取数据
    await getTargetGroupList();
  };

  /**
   * 刷新目标组列表
   */
  const refreshTargetGroupList = async () => {
    // 重置分页参数
    Object.assign(targetGroupListPageQuery, {
      start: 0,
      limit: 50,
      count: 0,
    });
    // 重置数据
    allTargetGroupList.value = [];
    // 获取数据
    await getTargetGroupList();
  };

  return {
    targetGroupId,
    setTargetGroupId,
    currentSelectedTreeNode,
    setCurrentSelectedTreeNode,
    getTargetGroupList,
    getNextTargetGroupList,
    refreshTargetGroupList,
    allTargetGroupList,
    updateCount,
    setUpdateCount,
    currentScene,
    setCurrentScene,
  };
});
