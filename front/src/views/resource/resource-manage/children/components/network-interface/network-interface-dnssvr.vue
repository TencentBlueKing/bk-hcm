<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { h, ref, watchEffect } from 'vue';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const columns = ref([
  {
    label: () => h('b', '应用的 DNS 服务器'),
    field: 'name',
  },
]);

const data = ref([]);

watchEffect(() => {
  data.value = props.detail.dns_settings.appliedDnsServers?.map((item: any) => ({ name: item }));
});

</script>

<template>
  <bk-table
    class="mt20"
    row-hover="auto"
    :columns="columns"
    :data="data"
  />
</template>
