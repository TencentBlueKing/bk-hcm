import { Ref, watch } from 'vue';
// import stores
import { useLoadBalancerStore } from '@/store';
import { storeToRefs } from 'pinia';

export default (isShow: Ref<boolean>, formData: any) => {
  // use stores
  const loadBalancerStore = useLoadBalancerStore();
  const { updateCount, currentScene } = storeToRefs(loadBalancerStore);

  watch(isShow, (val) => {
    !val && (updateCount.value = 0);
  });

  // 当目标组的基本信息发生变更时, 记录更新次数
  watch(
    [
      () => formData.account_id,
      () => formData.name,
      () => formData.protocol,
      () => formData.port,
      () => formData.region,
      () => formData.cloud_vpc_id,
    ],
    () => {
      // 记录更新次数(只有更新目标组时才需要记录, 更新目标组的初始状态下, currentScene为null)
      if (currentScene.value) return;
      if (updateCount.value > 1) return;

      // 0->1, 回显目标组基本信息, 1->2, 更新目标组基本信息
      updateCount.value += 1;
      // 变更操作场景为 edit
      updateCount.value === 2 && loadBalancerStore.setCurrentScene('edit');
    },
    {
      deep: true,
    },
  );

  return {
    updateCount,
  };
};
