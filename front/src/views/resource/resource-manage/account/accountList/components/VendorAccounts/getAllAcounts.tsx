import { VendorEnum } from '@/common/constant';
import { useAccountStore } from '@/store';

export const getAllAccounts = async () => {
  const accountStore = useAccountStore();
  const promises = [VendorEnum.TCLOUD, VendorEnum.HUAWEI, VendorEnum.GCP, VendorEnum.AWS, VendorEnum.AZURE]
    .map((vendor) => ({
      op: 'and',
      rules: [{ field: 'vendor', op: 'eq', value: vendor }],
    }))
    .map((filter) => ({
      filter,
      page: {
        start: 0,
        limit: 100,
      },
    }))
    .map((params) => accountStore.getAccountList(params));
  const arr = Promise.all(promises);
  return arr;
};
