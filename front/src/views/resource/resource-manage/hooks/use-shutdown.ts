/**
 * 关机相关事件和状态
 */
import { ref } from 'vue';

import HostShutdown from '../children/dialog/shutdown/host-shutdown';

export default () => {
  const isShowShutdown = ref(false);

  const handleShutdown = () => {
    isShowShutdown.value = true;
  };

  return {
    isShowShutdown,
    handleShutdown,
    HostShutdown,
  };
};
