import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useLoadBalancerStore = defineStore('load-balance', () => {
  // state - 目标组id
  const targetGroupId = ref('');

  // action - 设置目标组id
  const setTargetGroupId = (v: string) => {
    targetGroupId.value = v;
  };

  return { targetGroupId, setTargetGroupId };
});
