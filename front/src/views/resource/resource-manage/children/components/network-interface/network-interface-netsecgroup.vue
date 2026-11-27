<script lang="ts" setup>
import type { NetworkInterface, VendorResource } from '@/typings/resource-detail';
import { extensionPath } from '@/views/resource/resource-manage/common/info/extension-path';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { ref } from 'vue';
import { FieldList } from '../../../common/info-list/types';

defineProps<{ detail?: VendorResource<NetworkInterface, 'azure'> }>();

const azureField = extensionPath<VendorResource<NetworkInterface, 'azure'>>();

const fields = ref<FieldList>([
  {
    name: '网络安全组',
    prop: azureField('extension.cloud_security_group_id'),
  },
]);
</script>

<template>
  <detail-info :detail="detail" :fields="fields" global-copyable :col="1" />
</template>
