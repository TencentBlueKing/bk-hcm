<script setup lang="ts">
import { ref } from 'vue';
import { localStorageActions } from '@/common/util';
import { MIGRATION_NOTICE_CLOSED_KEY, RESOURCE_ACCESS_OFFLINE_DATE } from '../constants';

const isClosed = ref(localStorageActions.get(MIGRATION_NOTICE_CLOSED_KEY) === '1');

const handleClose = () => {
  isClosed.value = true;
  localStorageActions.set(MIGRATION_NOTICE_CLOSED_KEY, '1');
};
</script>

<template>
  <bk-alert
    v-if="!isClosed"
    class="migration-notice"
    theme="info"
    closable
    :title="`资源接入已迁移至 资源运营 → 云资源 → 资源纳管，原菜单将于 ${RESOURCE_ACCESS_OFFLINE_DATE} 下线`"
    @close="handleClose"
  />
</template>

<style lang="scss" scoped>
.migration-notice {
  flex-shrink: 0;
}
</style>
