import { computed, toRefs, Ref } from 'vue';
import { RequirementType } from '@/store/config/requirement';
import type { ICvmDeviceTypeFormData } from '../../typings';

export const useChargeTypeDefault = (props: {
  requireType: RequirementType;
  selectedDeviceType: Ref<ICvmDeviceTypeFormData['deviceTypeList'][number]>;
}) => {
  const { requireType, selectedDeviceType } = toRefs(props);

  // 常规项目-包年包月，专用机型默认包4年
  const isDefaultFourYears = computed(
    () =>
      requireType.value === RequirementType.Regular && selectedDeviceType.value?.device_type_class === 'SpecialType',
  );

  // GPU机型默认包6年
  const isGpuDeviceType = computed(
    () =>
      selectedDeviceType.value?.device_type_class === 'SpecialType' &&
      selectedDeviceType.value?.device_family.includes('GPU'),
  );

  return {
    isDefaultFourYears,
    isGpuDeviceType,
  };
};
