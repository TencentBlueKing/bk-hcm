// @ts-check
import { acceptHMRUpdate } from 'pinia';

export * from './user';


// @ts-ignore
if (import.meta.hot) {
  // @ts-ignore
  import.meta.hot.accept(acceptHMRUpdate(useCartStore, import.meta.hot));
}
