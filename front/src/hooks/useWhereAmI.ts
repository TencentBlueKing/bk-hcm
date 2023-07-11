import { computed } from 'vue';
import { useRoute } from 'vue-router';

export const useWhereAmI = (): {
  whereAmI: Senarios;
  isResourcePage: boolean;
  isBusinessPage: boolean;
  isServicePage: boolean;
  isWorkbenchPage: boolean;
} => {
  const route = useRoute();
  const senario = computed(() => {
    if (/^\/resource\/.+$/.test(route.path)) return Senarios.resource;
    if (/^\/business\/.+$/.test(route.path)) return Senarios.business;
    if (/^\/service\/.+$/.test(route.path)) return Senarios.service;
    if (/^\/workbench\/.+$/.test(route.path)) return Senarios.workbench;
    return Senarios.unknown;
  });
  return {
    whereAmI: senario.value,
    isResourcePage: senario.value === Senarios.resource,
    isBusinessPage: senario.value === Senarios.business,
    isServicePage: senario.value === Senarios.service,
    isWorkbenchPage: senario.value === Senarios.workbench,
  };
};

export enum Senarios {
  business = 'business',
  resource = 'resource',
  service = 'service',
  workbench = 'workbench',
  unknown = 'unknown',
}
