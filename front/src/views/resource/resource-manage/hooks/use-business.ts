/**
 * 分配相关事件和状态
 */
import { ref } from 'vue';

import ResourceBusiness from '../children/dialog/business/resource-business';

export default () => {
  const isShowDistribution = ref(false);

  const handleDistribution = () => {
    isShowDistribution.value = true;
  };

  return {
    isShowDistribution,
    handleDistribution,
    ResourceBusiness,
  };
};
