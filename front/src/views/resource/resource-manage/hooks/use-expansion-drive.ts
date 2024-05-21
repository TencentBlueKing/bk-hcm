/**
 * 扩容云硬盘相关事件和状态
 */
import { ref } from 'vue';

import ExpansionDrive from '../children/dialog/expansion-drive/expansion-drive';

export default () => {
  const isShowExpansionDrive = ref(false);

  const handleExpansionDrive = () => {
    isShowExpansionDrive.value = true;
  };

  return {
    isShowExpansionDrive,
    handleExpansionDrive,
    ExpansionDrive,
  };
};
