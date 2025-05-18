import { useBusinessGlobalStore } from './business-global';

export const preload = async () => {
  const { getFullBusiness, getAuthorizedBusiness } = useBusinessGlobalStore();

  return Promise.all([getFullBusiness(), getAuthorizedBusiness()]);
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
