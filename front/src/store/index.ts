// @ts-check
import { acceptHMRUpdate } from 'pinia';

export * from './staff';
export * from './user';
export * from './account';
export * from './departments';
export * from './common';


// @ts-ignore
if (import.meta.hot) {
  // @ts-ignore
  import.meta.hot.accept(acceptHMRUpdate(useCartStore, import.meta.hot));
}
