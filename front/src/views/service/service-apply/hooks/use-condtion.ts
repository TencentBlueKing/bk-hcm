import { VendorEnum } from '@/common/constant';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { computed, reactive } from 'vue';

export type Cond = {
  bizId: number;
  cloudAccountId: string;
  vendor: Lowercase<keyof typeof VendorEnum> | string;
  region: string;
  resourceGroup?: string;
};

export function hasEmptyResourceCondition(accountId: string, vendor: string, region: string, resourceGroup?: string) {
  return !accountId || !vendor || !region || (vendor === VendorEnum.AZURE && !resourceGroup);
}

export default () => {
  const { getBizsId, whereAmI } = useWhereAmI();

  const cond = reactive<Cond>({
    bizId: whereAmI.value === Senarios.business ? getBizsId() : 0,
    cloudAccountId: '',
    vendor: '',
    region: '',
    resourceGroup: '',
  });

  const isEmptyCond = computed(() => {
    const isResourcePage = whereAmI.value === Senarios.resource;
    return (
      hasEmptyResourceCondition(cond.cloudAccountId, cond.vendor, cond.region, cond.resourceGroup) ||
      (!isResourcePage && !cond.bizId)
    );
  });

  return {
    cond,
    isEmptyCond,
  };
};
