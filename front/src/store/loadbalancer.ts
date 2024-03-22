import { defineStore } from 'pinia';
import { Ref, ref } from 'vue';

export interface ILoadBalancer {
  id: string; // 负载均衡资源ID
  account_id: string; // 关联的账号ID
}

export const useLoadBalancerStore = defineStore('load-balancer', () => {
  // state - lb-tree - 当前选中的资源
  const currentSelectedTreeNode = ref();
  // state - 目标组id
  const targetGroupId = ref('');
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

  const setLB = (obj: ILoadBalancer) => {
    lb.value = obj;
  };

  return { targetGroupId, setTargetGroupId, lb, setLB, currentSelectedTreeNode, setCurrentSelectedTreeNode };
});
