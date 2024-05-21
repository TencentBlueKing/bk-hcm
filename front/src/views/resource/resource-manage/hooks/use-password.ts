/**
 * 修改密码相关事件和状态
 */
import { ref } from 'vue';

import HostPassword from '../children/dialog/password/host-password';

export default () => {
  const isShowPassword = ref(false);

  const handlePassword = () => {
    isShowPassword.value = true;
  };

  return {
    isShowPassword,
    handlePassword,
    HostPassword,
  };
};
