import { defineStore } from 'pinia';
import http from '@/http';
import { IQueryResData } from '@/typings';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';

type CvmChargeTypes = ReturnType<typeof useCvmChargeType>['cvmChargeTypes'];
type ChargeType = CvmChargeTypes[keyof CvmChargeTypes];
type SpringResPoolChargeTypeResponse = IQueryResData<{ charge_type: ChargeType }>;

export const useConfigSpringResPoolStore = defineStore('config-spring-res-pool', () => {
  const { cvmChargeTypes } = useCvmChargeType();

  // 按bizId缓存请求结果
  const chargeTypeCache = new Map<number, ChargeType>();

  const getSpringResPoolChargeType = async (bizId: number) => {
    if (chargeTypeCache.has(bizId)) {
      return chargeTypeCache.get(bizId);
    }

    const res: SpringResPoolChargeTypeResponse = await http.get(
      `/api/v1/woa/bizs/${bizId}/config/spring_res_pool/charge_type`,
    );

    // 后端兜底默认值为 POSTPAID_BY_HOUR，这里再做一次容错
    const chargeType = res?.data?.charge_type ?? cvmChargeTypes.POSTPAID_BY_HOUR;
    chargeTypeCache.set(bizId, chargeType);
    return chargeType;
  };

  return {
    getSpringResPoolChargeType,
  };
});
