/**
 * 分配相关事件和状态
 */
import { ref } from 'vue';

import ResourceDistribution from '../children/dialog/steps/resource-distribution';

export default () => {
  const isShowDistribution = ref(false);

  const handleDistribution = () => {
    isShowDistribution.value = true;
  };

  return {
    isShowDistribution,
    handleDistribution,
    ResourceDistribution,
  };
};
