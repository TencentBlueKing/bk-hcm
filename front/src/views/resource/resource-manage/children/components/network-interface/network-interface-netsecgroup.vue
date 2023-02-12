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
    prop: 'networkSecurityGroup',
  },
]);

const data = reactive({
  networkSecurityGroup: '',
});

watchEffect(() => {
  data.networkSecurityGroup = props.detail.network_security_group.name;
});

</script>

<template>
  <detail-info
    class="mt20"
    :detail="data"
    :fields="fields"
  />
</template>
