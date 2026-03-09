/**
 * @deprecated 已废弃。仅被废弃的 components/global-permission-dialog 使用。
 * 全局权限弹窗已迁移到 app.vue 中通过 hooks/use-permission-dialog.ts 管理。
 */
import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useGlobalPermissionDialog = defineStore('useGlobalPermissionDialog', () => {
  const isShow = ref(false);

  const setShow = (val: boolean) => {
    isShow.value = val;
  };

  return {
    isShow,
    setShow,
  };
});
