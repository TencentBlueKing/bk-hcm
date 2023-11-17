import { VendorEnum } from '@/common/constant';
import { useAccountStore } from '@/store';
import { QueryRuleOPEnum } from '@/typings';
import tcloudVendor from '@/assets/image/vendor-tcloud.png';
import awsVendor from '@/assets/image/vendor-aws.png';
import azureVendor from '@/assets/image/vendor-azure.png';
import gcpVendor from '@/assets/image/vendor-gcp.png';
import huaweiVendor from '@/assets/image/vendor-huawei.png';
import { reactive, ref } from 'vue';
import { debounce } from 'lodash-es';

export const useAllVendorsAccounts = () => {
  const accountStore = useAccountStore();
  const isLoading = ref(false);

  const accountsMatrix = reactive([
    {
      vendor: VendorEnum.TCLOUD,
      count: 0,
      name: '腾讯云',
      icon: tcloudVendor,
      accounts: [] as any[],
      isExpand: false,
    },
    {
      vendor: VendorEnum.AWS,
      count: 0,
      name: '亚马逊云',
      icon: awsVendor,
      accounts: [],
      isExpand: false,
    },
    {
      vendor: VendorEnum.AZURE,
      count: 0,
      name: '微软云',
      icon: azureVendor,
      accounts: [],
      isExpand: false,
    },
    {
      vendor: VendorEnum.GCP,
      count: 0,
      name: '谷歌云',
      icon: gcpVendor,
      accounts: [],
      isExpand: false,
    },
    {
      vendor: VendorEnum.HUAWEI,
      count: 0,
      name: '华为云',
      icon: huaweiVendor,
      accounts: [],
      isExpand: false,
    },
  ]);

  const checkIsExpand = (vendorName: VendorEnum) => accountsMatrix.find(item => item.vendor === vendorName).isExpand;

  const handleExpand = (vendorName: VendorEnum) => {
    for (const item of accountsMatrix) {
      if (item.vendor === vendorName) {
        item.isExpand = !item.isExpand;
      }
    }
  };

  const getAllVendorsAccountsList = debounce(async (accountName = '') => {
    isLoading.value = true;
    const payloads = [
      VendorEnum.TCLOUD,
      VendorEnum.AWS,
      VendorEnum.AZURE,
      VendorEnum.GCP,
      VendorEnum.HUAWEI,
    ]
      .map(vendor => ({
        op: 'and',
        rules: accountName.length
          ? [
            { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
            { field: 'name', op: QueryRuleOPEnum.CS, value: accountName },
          ]
          : [{ field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor }],
      }))
      .map((filter) => {
        return [false, true].map(isCount => ({
          filter,
          page: {
            start: 0,
            limit: isCount ? 0 : 500,
            count: isCount,
          },
        }));
      });
    const detailPromises = payloads.map(payload => accountStore.getAccountList(payload[0]));
    const countPromises = payloads.map(payload => accountStore.getAccountList(payload[1]));
    const detailRes = await Promise.all(detailPromises);
    const countRes = await Promise.all(countPromises);
    accountsMatrix.forEach((obj, idx) => {
      obj.count = countRes[idx]?.data?.count || 0;
      obj.accounts = detailRes[idx]?.data?.details || [];
    });
    isLoading.value = false;
    return accountsMatrix;
  }, 500);

  return {
    accountsMatrix,
    handleExpand,
    getAllVendorsAccountsList,
    isLoading,
    checkIsExpand,
  };
};

