/**
 * 退回相关事件和状态
 */
import { ref } from 'vue';

import HostRefund from '../children/dialog/refund/host-refund';

export default () => {
  const isShowRefund = ref(false);

  const handleRefund = () => {
    isShowRefund.value = true;
  };

  return {
    isShowRefund,
    handleRefund,
    HostRefund,
  };
};
