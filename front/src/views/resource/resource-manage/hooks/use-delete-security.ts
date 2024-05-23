/**
 * 分配相关事件和状态
 */
import { ref } from 'vue';

import DeleteSecurity from '../children/dialog/delete-security';

export default () => {
  const isShowSecurity = ref(false);

  const handleShowDeleteSecurity = () => {
    isShowSecurity.value = true;
  };

  return {
    isShowSecurity,
    handleShowDeleteSecurity,
    DeleteSecurity,
  };
};
