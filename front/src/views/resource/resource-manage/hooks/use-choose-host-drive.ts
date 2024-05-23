/**
 * 挂载云硬盘相关事件和状态
 */
import { ref } from 'vue';

import MountedDrive from '../children/dialog/choose-host-drive/mounted-drive';

export default () => {
  const isShowMountedDrive = ref(false);

  const handleMountedDrive = () => {
    isShowMountedDrive.value = true;
  };

  return {
    isShowMountedDrive,
    handleMountedDrive,
    MountedDrive,
  };
};
