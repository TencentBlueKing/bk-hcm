<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { reactive, ref, watchEffect } from 'vue';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const fields = ref([
  {
    name: '网络安全组',
    prop: 'cloud_security_group_id',
  },
]);

const data = reactive({
  cloud_security_group_id: '',
});

watchEffect(() => {
  data.cloud_security_group_id = props.detail?.cloud_security_group_id;
});

</script>

<template>
  <detail-info
    :detail="data"
    :fields="fields"
  />
</template>
