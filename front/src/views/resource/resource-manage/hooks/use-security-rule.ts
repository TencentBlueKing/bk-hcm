/**
 * 添加规则事件和状态
 */
import { ref } from 'vue';

import SecurityRule from '../children/dialog/security-rule/add-rule';

export default () => {
  const isShowSecurityRule = ref(false);

  const handleSecurityRule = () => {
    isShowSecurityRule.value = true;
  };

  return {
    isShowSecurityRule,
    handleSecurityRule,
    SecurityRule,
  };
};
