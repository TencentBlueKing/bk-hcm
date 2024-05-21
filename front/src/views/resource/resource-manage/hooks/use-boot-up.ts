/**
 * 开机相关事件和状态
 */
import { ref } from 'vue';

import HostBootUp from '../children/dialog/boot-up/host-boot-up';

export default () => {
  const isShowBootUp = ref(false);

  const handleBootUp = () => {
    isShowBootUp.value = true;
  };

  return {
    isShowBootUp,
    handleBootUp,
    HostBootUp,
  };
};
