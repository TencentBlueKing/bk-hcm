import { VendorEnum } from '@/common/constant';
import { computed, reactive } from 'vue';

export type Cond = {
  bizId: number,
  cloudAccountId: string,
  vendor: Lowercase<keyof typeof VendorEnum> | string,
  region: string,
  resourceGroup?: string
};

export default (type: string) => {
  console.log(type);

  const cond = reactive<Cond>({
    bizId: null,
    cloudAccountId: '',
    vendor: '',
    region: '',
    resourceGroup: '',
  });

  const isEmptyCond = computed(() => {
    const isEmpty = !cond.bizId || !cond.cloudAccountId || !cond.vendor || !cond.region;
    if (cond.vendor === VendorEnum.AZURE) {
      return isEmpty || !cond.resourceGroup;
    }
    return isEmpty;
  });

  return {
    cond,
    isEmptyCond,
  };
};
