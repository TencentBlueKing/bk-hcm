import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
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
    const { isResourcePage } = useWhereAmI();
    const isEmpty = !cond.cloudAccountId || !cond.vendor || !cond.region || (!isResourcePage && !cond.bizId);
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
