<script setup lang="ts">
import { ref, ComputedRef, inject } from 'vue';
import { ILoadBalanceDeviceCondition, ICount, DeviceTabEnum } from './typing';
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
  rsCount: 0, // 总rs数
  rsIPCount: 0, // rs下面ip数量
});
const loading = ref(false); // 条件框查询按钮loading态
const countChange = ref(false); // 总数是否变化

const handleSave = async (newCondition: ILoadBalanceDeviceCondition) => {
  // 对数字类型转换
  Object.entries(newCondition).forEach(([label, value]) => {
    const isArray = Array.isArray(value);
    if (numberField.includes(label)) newCondition[label] = isArray ? value.map(Number) : Number(value);
  });
  loading.value = true;
  try {
    // 先调总数接口
    const { listenerCount, urlCount, rsCount, rsIPCount } = await loadBalancerCountStore.getCount(
      newCondition,
      currentGlobalBusinessId.value,
    );
    count.value = {
      listenerCount,
      urlCount,
      rsCount,
      rsIPCount,
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
const handleListDone = (
  from: DeviceTabEnum,
  params: { type: 'listenerCount' | 'urlCount' | 'rsIPCount'; data: Record<string, any> },
) => {
  loading.value = false;
  const {
    type,
    data: { count: nowCount },
  } = params;
  if (nowCount !== count.value[type]) {
    countChange.value = true;
  }
};
const handleCountChange = (val: boolean) => {
  countChange.value = val;
};
</script>

<template>
  <bk-resize-layout class="device-search" :initial-divide="320" :min="320" immediate>
    <template #aside>
      <device-condition
        @save="handleSave"
        :loading="loading"
        :count-change="countChange"
        @count-change="handleCountChange"
      ></device-condition>
    </template>
    <template #main>
      <main-content :condition="condition" :count="count" @list-data-loaded="handleListDone" />
    </template>
  </bk-resize-layout>
</template>

<style scoped lang="scss">
.device-search {
  height: 100%;
}
</style>
