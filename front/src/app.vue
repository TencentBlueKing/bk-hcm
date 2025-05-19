<script lang="ts" setup>
import { onMounted, useTemplateRef } from 'vue';
import { provideBreadcrumb } from '@/hooks/use-breadcrumb';
import { providePermissionDialog } from '@/hooks/use-permission-dialog';
import Home from '@/views/home';
import Notice from '@/views/notice/index.vue';
import PermissionApplyDialog from '@/components/permission/apply-dialog.vue';

const { ENABLE_NOTICE } = window.PROJECT_CONFIG;

// 面包屑
provideBreadcrumb();

// 权限申请弹窗
const permissionDialogContext = providePermissionDialog();

const permissionDialogRef = useTemplateRef<InstanceType<typeof PermissionApplyDialog>>('permission-dialog');

onMounted(() => {
  window.hcmPermissionDialog = permissionDialogRef.value;
});
</script>

<template>
  <div class="full-page flex-column">
    <Notice v-if="ENABLE_NOTICE === 'true'" />
    <Home class="flex-1"></Home>
  </div>
  <PermissionApplyDialog
    ref="permission-dialog"
    v-model="permissionDialogContext.isShow"
    :permission="permissionDialogContext.permission"
    :done="permissionDialogContext.done"
  />
</template>
