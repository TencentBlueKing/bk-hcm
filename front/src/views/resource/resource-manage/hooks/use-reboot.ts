/**
 * 重启相关事件和状态
 */
import { ref } from 'vue';

import HostReboot from '../children/dialog/reboot/host-reboot';

export default () => {
  const isShowReboot = ref(false);

  const handleReboot = () => {
    isShowReboot.value = true;
  };

  return {
    isShowReboot,
    handleReboot,
    HostReboot,
  };
};
