import { defineStore } from 'pinia';
import { Ref, ref } from 'vue';

export interface ILoadBalancer {
  id: string; // 负载均衡资源ID
  account_id: string; // 关联的账号ID
}

export const useLoadBalancerStore = defineStore('load-balance', () => {
  // state - 目标组id
  const targetGroupId = ref('');
  const lb: Ref<ILoadBalancer> = ref({
    id: '',
    account_id: '',
  });

  // action - 设置目标组id
  const setTargetGroupId = (v: string) => {
    targetGroupId.value = v;
  };

  const setLB = (obj: ILoadBalancer) => {
    lb.value = obj;
  };

  return { targetGroupId, setTargetGroupId, lb, setLB };
});
