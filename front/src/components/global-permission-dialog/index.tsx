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
