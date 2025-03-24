<template>
  <div class="wrap">
    <bk-select
      v-model="zones"
      prefix="主"
      :loading="isDataLoad"
      :scroll-loading="isDataLoad"
      @scroll-end="handleScrollEnd"
      @change="backupZones = ''"
    >
      <bk-option v-for="option in dataList" :key="option.id" :id="option.name" :name="option.name_cn || option.name" />
    </bk-select>
    <bk-select
      v-model="backupZones"
      prefix="备"
      :disabled="!displayBackupZoneList.length"
      v-bk-tooltips="{ content: '当前可用区不支持主备模式', disabled: displayBackupZoneList.length }"
    >
      <bk-option
        v-for="option in displayBackupZoneList"
        :key="option.id"
        :id="option.name"
        :name="option.name_cn || option.name"
      />
    </bk-select>
  </div>
</template>

<script setup lang="ts">
import { VendorEnum } from '@/common/constant';
import { useSingleList } from '@/hooks/useSingleList';
import { QueryRuleOPEnum } from '@/typings';
import { computed } from 'vue';

interface IProps {
  vendor: VendorEnum;
  region: string;
  currentResourceListMap: Record<string, any>;
}

const props = defineProps<IProps>();
const zones = defineModel<string>('zones');
const backupZones = defineModel<string>('backupZones');

const { dataList, isDataLoad, handleScrollEnd } = useSingleList({
  url: () => `/api/v1/cloud/vendors/${props.vendor}/regions/${props.region}/zones/list`,
  rules: () => [
    { op: QueryRuleOPEnum.EQ, field: 'vendor', value: props.vendor },
    { op: QueryRuleOPEnum.EQ, field: 'state', value: 'AVAILABLE' },
  ],
  immediate: true,
});

const displayBackupZoneList = computed(() => {
  // 确保 zones.value 和 dataList.value 存在且有效
  if (!zones.value || !Array.isArray(dataList.value)) return [];

  const slaveZoneSet = getSlaveZonesSet(zones.value);

  return dataList.value.filter((item) => slaveZoneSet.has(item.name));
});

// 获取备可用区选项
const getSlaveZonesSet = (masterZone: string): Set<string> => {
  if (!masterZone || typeof masterZone !== 'string') return new Set();
  if (!props.currentResourceListMap || typeof props.currentResourceListMap !== 'object') return new Set();

  const options: string[] = [];
  Object.keys(props.currentResourceListMap).forEach((key) => {
    // 校验 key 格式是否符合预期
    const [master, slave] = key.split('|');
    if (master === masterZone && slave) {
      options.push(slave);
    }
  });

  // 去重并返回
  return new Set(options);
};
</script>

<style scoped lang="scss">
.wrap {
  display: flex;
  align-items: center;
}
</style>
