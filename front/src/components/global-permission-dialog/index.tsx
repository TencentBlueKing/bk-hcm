/**
 * @deprecated 已废弃。全局权限弹窗已迁移到 app.vue 中的 PermissionApplyDialog 组件，
 * 通过 window.hcmPermissionDialog.show(permission) 调用。
 * 此组件仅被废弃的 views/home/index.tsx 使用。
 */
import { defineComponent } from 'vue';
import PermissionDialog from '../permission-dialog';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';

export default defineComponent({
  setup() {
    const globalPermissionDialogStore = useGlobalPermissionDialog();
    const { permissionParams, handlePermissionDialog, handlePermissionConfirm } = useVerify();

    return () => (
      <PermissionDialog
        v-model:isShow={globalPermissionDialogStore.isShow}
        params={permissionParams.value}
        onCancel={() => {
          globalPermissionDialogStore.setShow(false);
          handlePermissionDialog();
        }}
        onConfirm={(val) => {
          globalPermissionDialogStore.setShow(false);
          handlePermissionConfirm(val);
        }}
      />
    );
  },
});
