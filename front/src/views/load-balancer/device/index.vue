<script setup lang="ts">
import { ref, ComputedRef, inject } from 'vue';
import { ILoadBalanceDeviceCondition, ICount } from './typing';
import { VendorEnum } from '@/common/constant';
import DeviceCondition from './search/index.vue';
import MainContent from './main-content/index.vue';
import { useLoadBalancerCountStore } from '@/store/load-balancer/device-search-count';
import routeQuery from '@/router/utils/query';

defineOptions({ name: 'device-search' });

const numberField = ['lbl_ports', 'target_ports'];

const loadBalancerCountStore = useLoadBalancerCountStore();

const currentGlobalBusinessId = inject<ComputedRef<number>>('currentGlobalBusinessId');

const condition = ref<ILoadBalanceDeviceCondition>({
  vendor: VendorEnum.TCLOUD,
  account_id: '',
});
const count = ref<ICount>({
  listenerCount: 0,
  urlCount: 0,
  rsCount: 0,
});
const loading = ref(false); // 条件框查询按钮loading态

const handleSave = async (newCondition: ILoadBalanceDeviceCondition) => {
  // 对数字类型转换
  Object.entries(newCondition).forEach(([label, value]) => {
    const isArray = Array.isArray(value);
    if (numberField.includes(label)) newCondition[label] = isArray ? value.map(Number) : Number(value);
  });
  loading.value = true;
  try {
    // 先调总数接口
    const { listenerCount, urlCount, rsCount } = await loadBalancerCountStore.getCount(
      newCondition,
      currentGlobalBusinessId.value,
    );
    count.value = {
      listenerCount,
      urlCount,
      rsCount,
    };
  } catch {
    loading.value = false;
  } finally {
    condition.value = newCondition;
    routeQuery.set({
      _t: Date.now(),
    });
  }
};
const handleListDone = () => {
  loading.value = false;
};
</script>

<template>
  <bk-resize-layout class="device-search" :initial-divide="320" :min="320" immediate>
    <template #aside>
      <device-condition @save="handleSave" :loading="loading"></device-condition>
    </template>
    <template #main>
      <main-content :condition="condition" :count="count" @get-list="handleListDone" />
    </template>
  </bk-resize-layout>
</template>

<style scoped lang="scss">
.device-search {
  height: 100%;
}
</style>
