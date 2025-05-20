import { VendorEnum } from '@/common/constant';
import { useAccountStore } from '@/store';
import { QueryRuleOPEnum } from '@/typings';
import { reactive, ref } from 'vue';
import { debounce } from 'lodash';
import tcloudVendor from '@/assets/image/vendor-tcloud.svg';
import awsVendor from '@/assets/image/vendor-aws.svg';
import azureVendor from '@/assets/image/vendor-azure.svg';
import gcpVendor from '@/assets/image/vendor-gcp.svg';
import huaweiVendor from '@/assets/image/vendor-huawei.svg';

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
      hasNext: true,
    },
    {
      vendor: VendorEnum.AWS,
      count: 0,
      name: '亚马逊云',
      icon: awsVendor,
      accounts: [],
      isExpand: false,
      hasNext: true,
    },
    {
      vendor: VendorEnum.AZURE,
      count: 0,
      name: '微软云',
      icon: azureVendor,
      accounts: [],
      isExpand: false,
      hasNext: true,
    },
    {
      vendor: VendorEnum.GCP,
      count: 0,
      name: '谷歌云',
      icon: gcpVendor,
      accounts: [],
      isExpand: false,
      hasNext: true,
    },
    {
      vendor: VendorEnum.HUAWEI,
      count: 0,
      name: '华为云',
      icon: huaweiVendor,
      accounts: [],
      isExpand: false,
      hasNext: true,
    },
  ]);

  const pageOptions = reactive({
    [VendorEnum.TCLOUD]: {
      start: 0,
    },
    [VendorEnum.HUAWEI]: {
      start: 0,
    },
    [VendorEnum.AWS]: {
      start: 0,
    },
    [VendorEnum.AZURE]: {
      start: 0,
    },
    [VendorEnum.GCP]: {
      start: 0,
    },
  });

  const vendorAccountMatrixMap = {
    [VendorEnum.TCLOUD]: accountsMatrix[0],
    [VendorEnum.AWS]: accountsMatrix[1],
    [VendorEnum.AZURE]: accountsMatrix[2],
    [VendorEnum.GCP]: accountsMatrix[3],
    [VendorEnum.HUAWEI]: accountsMatrix[4],
  };

  const checkIsExpand = (vendorName: VendorEnum) => accountsMatrix.find((item) => item.vendor === vendorName)?.isExpand;

  const handleExpand = (vendorName: VendorEnum) => {
    for (const item of accountsMatrix) {
      if (item.vendor === vendorName) {
        item.isExpand = !item.isExpand;
      }
    }
  };

  const getAllVendorsAccountsList = debounce(async (accountName = '') => {
    isLoading.value = true;
    const payloads = [VendorEnum.TCLOUD, VendorEnum.AWS, VendorEnum.AZURE, VendorEnum.GCP, VendorEnum.HUAWEI]
      .map((vendor) => ({
        op: 'and',
        rules: accountName.length
          ? [
              { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
              { field: 'name', op: QueryRuleOPEnum.CS, value: accountName },
              { field: 'type', op: QueryRuleOPEnum.EQ, value: 'resource' },
            ]
          : [
              { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
              { field: 'type', op: QueryRuleOPEnum.EQ, value: 'resource' },
            ],
      }))
      .map((filter) => {
        return [false, true].map((isCount) => ({
          filter,
          page: {
            start: 0,
            limit: isCount ? 0 : 50,
            count: isCount,
          },
        }));
      });
    const detailPromises = payloads.map((payload) => accountStore.getAccountList(payload[0]));
    const countPromises = payloads.map((payload) => accountStore.getAccountList(payload[1]));
    const detailRes = await Promise.all(detailPromises);
    const countRes = await Promise.all(countPromises);
    accountsMatrix.forEach((obj, idx) => {
      obj.count = countRes[idx]?.data?.count || 0;
      obj.accounts = detailRes[idx]?.data?.details || [];
      obj.hasNext = countRes[idx]?.data?.count !== obj.accounts.length;
    });
    isLoading.value = false;
    return accountsMatrix;
  }, 500);

  const getVendorAccountList = async (vendor: VendorEnum, accountName = '') => {
    isLoading.value = true;
    pageOptions[vendor].start += 50;
    const [res1, res2] = await Promise.all(
      [false, true].map((isCount) =>
        accountStore.getAccountList({
          filter: {
            op: 'and',
            rules: accountName.length
              ? [
                  { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
                  { field: 'name', op: QueryRuleOPEnum.CS, value: accountName },
                  { field: 'type', op: QueryRuleOPEnum.EQ, value: 'resource' },
                ]
              : [
                  { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
                  { field: 'type', op: QueryRuleOPEnum.EQ, value: 'resource' },
                ],
          },
          page: {
            start: isCount ? 0 : pageOptions[vendor].start,
            limit: isCount ? 0 : 50,
            count: isCount,
          },
        }),
      ),
    );
    isLoading.value = false;
    vendorAccountMatrixMap[vendor].accounts = [...vendorAccountMatrixMap[vendor].accounts, ...res1.data.details];
    vendorAccountMatrixMap[vendor].count = res2.data.count;
    vendorAccountMatrixMap[vendor].hasNext = vendorAccountMatrixMap[vendor].accounts.length !== res2.data.count;
  };

  return {
    accountsMatrix,
    handleExpand,
    getAllVendorsAccountsList,
    getVendorAccountList,
    isLoading,
    checkIsExpand,
  };
};
