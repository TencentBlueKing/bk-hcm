import { Ref, computed } from 'vue';
import { useRoute } from 'vue-router';

export const useWhereAmI = (): {
  whereAmI: Ref<Senarios>;
  isResourcePage: boolean;
  isBusinessPage: boolean;
  isServicePage: boolean;
  isSchemePage: boolean;
  getBusinessApiPath: () => string;
  getBizsId: () => number;
} => {
  const route = useRoute();
  const senario = computed(() => {
    if (!route) return;
    // bill 模块现已迁移到 /resource/bill/... 下，需优先匹配
    if (/^\/resource\/bill\/.+$/.test(route?.path)) return Senarios.bill;
    if (/^\/resource\/.+$/.test(route?.path)) return Senarios.resource;
    if (/^\/business\/.+$/.test(route.path)) return Senarios.business;
    if (/^\/service\/.+$/.test(route.path)) return Senarios.service;
    if (/^\/scheme\/.+$/.test(route.path)) return Senarios.scheme;
    // 兼容旧 /bill/ 路径
    if (/^\/bill\/.+$/.test(route.path)) return Senarios.bill;
    return Senarios.unknown;
  });

  const getBizsId = () => {
    return Number(route.params.bizId);
  };

  /**
   * @returns 业务下需要拼接的 API 路径
   */
  const getBusinessApiPath = () => {
    return senario.value === Senarios.business ? `bizs/${getBizsId()}/` : '';
  };

  return {
    whereAmI: senario,
    isResourcePage: senario.value === Senarios.resource,
    isBusinessPage: senario.value === Senarios.business,
    isServicePage: senario.value === Senarios.service,
    isSchemePage: senario.value === Senarios.scheme,
    getBusinessApiPath,
    getBizsId,
  };
};

export enum Senarios {
  business = 'business',
  resource = 'resource',
  service = 'service',
  scheme = 'scheme',
  bill = 'bill',
  unknown = 'unknown',
  unauthorized = 'unauthorized',
}
