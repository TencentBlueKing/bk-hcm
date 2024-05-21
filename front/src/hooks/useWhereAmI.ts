import { Ref, computed } from 'vue';
import { useRoute } from 'vue-router';

export const useWhereAmI = (): {
  whereAmI: Ref<Senarios>;
  isResourcePage: boolean;
  isBusinessPage: boolean;
  isServicePage: boolean;
  isWorkbenchPage: boolean;
  isSchemePage: boolean;
} => {
  const route = useRoute();
  const senario = computed(() => {
    if (!route) return;
    if (/^\/resource\/.+$/.test(route?.path)) return Senarios.resource;
    if (/^\/business\/.+$/.test(route.path)) return Senarios.business;
    if (/^\/service\/.+$/.test(route.path)) return Senarios.service;
    if (/^\/workbench\/.+$/.test(route.path)) return Senarios.workbench;
    if (/^\/scheme\/.+$/.test(route.path)) return Senarios.scheme;
    return Senarios.unknown;
  });
  return {
    whereAmI: senario,
    isResourcePage: senario.value === Senarios.resource,
    isBusinessPage: senario.value === Senarios.business,
    isServicePage: senario.value === Senarios.service,
    isWorkbenchPage: senario.value === Senarios.workbench,
    isSchemePage: senario.value === Senarios.scheme,
  };
};

export enum Senarios {
  business = 'business',
  resource = 'resource',
  service = 'service',
  workbench = 'workbench',
  scheme = 'scheme',
  unknown = 'unknown',
}
