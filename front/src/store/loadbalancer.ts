import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useLoadBalancerStore = defineStore('load-balancer', () => {
  // state - lb-tree - 当前选中的资源
  const currentSelectedTreeNode = ref();
  // state - 目标组id
  const targetGroupId = ref('');

  // action - lb-tree - 设置当前选中的资源
  const setCurrentSelectedTreeNode = (node: any) => {
    // 其中, node 可能为 lb, listener, domain 节点
    currentSelectedTreeNode.value = node;
  };
  // action - 设置目标组id
  const setTargetGroupId = (v: string) => {
    targetGroupId.value = v;
  };

  return { currentSelectedTreeNode, setCurrentSelectedTreeNode, targetGroupId, setTargetGroupId };
});
