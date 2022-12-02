/**
 * 卸载云硬盘相关事件和状态
 */
import {
  ref,
} from 'vue';

import UninstallDrive from '../children/dialog/uninstall-drive/uninstall-drive';

export default () => {
  const isShowUninstallDrive = ref(false);

  const handleUninstallDrive = () => {
    isShowUninstallDrive.value = true;
  };

  return {
    isShowUninstallDrive,
    handleUninstallDrive,
    UninstallDrive,
  };
};
