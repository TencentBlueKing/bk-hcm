/**
 * 分配相关事件和状态
 */
import {
  ref,
} from 'vue';

import DeleteVPC from '../children/dialog/delete-vpc/delete-vpc';

export default () => {
  const isShowVPC = ref(false);

  const handleShowDeleteVPC = () => {
    isShowVPC.value = true;
  };

  return {
    isShowVPC,
    handleShowDeleteVPC,
    DeleteVPC,
  };
};
