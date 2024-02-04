/**
 * 卸载云硬盘相关事件和状态
 */
import { ref, h } from 'vue';

import UninstallDrive from '../children/dialog/uninstall-drive/uninstall-drive';

export default () => {
  const disk = ref({});
  const isShowUninstallDrive = ref(false);

  const handleUninstallDrive = (data: any) => {
    disk.value = data;
    isShowUninstallDrive.value = true;
  };

  return {
    isShowUninstallDrive,
    handleUninstallDrive,
    UninstallDrive: (props: any, { emit }: any) => {
      return h(UninstallDrive, {
        ...props,
        data: disk.value,
        onSuccess() {
          emit('success');
        },
      });
    },
  };
};
