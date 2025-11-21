<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useConfigQcloudResourceStore } from '@/store/config/qcloud-resource';

defineOptions({ name: 'qcloud-zone-value' });

const props = defineProps<{
  value: string | string[];
}>();

const list = ref([]);

const localValue = computed(() => {
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((zone) => {
    if (zone === 'all') {
      return '全部可用区';
    }
    return list.value.find((item) => item.zone === zone)?.zone_cn;
  });
  return names?.join?.(', ');
});

const configQcloudRegionStore = useConfigQcloudResourceStore();
watchEffect(async () => {
  // display场景，拉取全量，减少缓存及请求数量
  list.value = await configQcloudRegionStore.getQcloudZoneList();
});
</script>

<template>
  {{ displayValue }}
</template>
