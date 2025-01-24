import { useUserStore } from './user';
import { useBusinessGlobalStore } from './business-global';

export const preload = async () => {
  const { userInfo } = useUserStore();
  const { getFullBusiness, getAuthorizedBusiness } = useBusinessGlobalStore();

  return Promise.all([userInfo(), getFullBusiness(), getAuthorizedBusiness()]);
};

export * from './staff';
export * from './user';
export * from './account';
export * from './departments';
export * from './business';
export * from './resource';
export * from './common';
export * from './host';
export * from './scheme';
export * from './loadbalancer';
export * from './task';
