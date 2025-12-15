import { computed, ref, Ref, watchEffect } from 'vue';
import { RequirementType } from '@/store/config/requirement';
import { useCvmDeviceStore, type ICvmChargeTypDevicetypeItem } from '@/store/cvm/device';
import { storeToRefs } from 'pinia';

export type AvailableDeviceTypeMap = Map<
  ICvmChargeTypDevicetypeItem['charge_type'],
  Map<
    ICvmChargeTypDevicetypeItem['device_types'][number]['device_type'],
    ICvmChargeTypDevicetypeItem['device_types'][number]
  >
>;

export const useDeviceTypePlan = (params: {
  bizId: Ref<number | string>;
  region: Ref<string>;
  requireType: Ref<RequirementType>;
}) => {
  const cvmDeviceStore = useCvmDeviceStore();

  const availableDeviceTypeMap = ref<AvailableDeviceTypeMap>(new Map());

  const { bizId, region, requireType } = params;

  const isSpringPool = computed(() => requireType.value === RequirementType.SpringResPool);

  watchEffect(async () => {
    // 非预测需求类型，不获取机型的预测数据
    const isNonPlanType = [RequirementType.RollServer, RequirementType.GreenChannel].includes(requireType.value);

    // 获取机型的预测数据，除非条件变化，预期只获取一次
    if (!bizId.value || !requireType.value || !region.value || isNonPlanType) {
      return;
    }

    availableDeviceTypeMap.value.clear();

    const { list } = await cvmDeviceStore.getChargeTypeDeviceTypeList({
      bk_biz_id: isSpringPool.value ? 931 : Number(params.bizId.value),
      require_type: params.requireType.value,
      region: params.region.value,
    });

    list.forEach(({ charge_type, device_types }) => {
      const currentMap = availableDeviceTypeMap.value.get(charge_type) || new Map();
      device_types.forEach((item) => {
        if (item.available) {
          currentMap.set(item.device_type, item);
        }
      });
      // 以计费模式为key，值为可用机型的map以机型为key
      availableDeviceTypeMap.value.set(charge_type, currentMap);
    });
  });

  return { availableDeviceTypeMap, loading: storeToRefs(cvmDeviceStore).chargeTypeDeviceTypeListLoading };
};
