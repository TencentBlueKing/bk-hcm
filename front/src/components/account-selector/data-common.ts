import { Ref, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useAccountSelectorStore } from '@/store/account-selector';
import type { IAccountSelectorProps } from './index-new.vue';
import { QueryRuleOPEnum, IAccountItem } from '@/typings';
import { vendorProperty } from './vendor.plugin';
import { resourceCond } from './resource-cond.plugin';

const useList = (props: IAccountSelectorProps) => {
  const accountSelectorStore = useAccountSelectorStore();
  const list = ref<IAccountItem[]>([]);
  let loading: Ref<boolean>;
  const { businessAccountList, resourceAccountList, businessAccountLoading, resourceAccountLoading } =
    storeToRefs(accountSelectorStore);
  const { getBusinessAccountList, getResourceAccountList } = accountSelectorStore;

  const getList = async (bizId: number) => {
    if (bizId) {
      loading = businessAccountLoading;
      await getBusinessAccountList({
        bizId,
        account_type: 'resource',
      });
      list.value = businessAccountList.value;
    } else {
      loading = resourceAccountLoading;
      await getResourceAccountList({
        filter: {
          op: 'and',
          rules: [
            {
              field: 'type',
              op: QueryRuleOPEnum.EQ,
              value: 'resource',
            },
            ...resourceCond,
          ],
        },
      });
      list.value = resourceAccountList.value;
    }
  };

  watch(
    () => props.bizId,
    (bizId, oldBizId) => {
      if (Number(bizId) !== Number(oldBizId)) {
        list.value = [];
        getList(bizId);
      }
    },
    { immediate: true },
  );

  return {
    list,
    loading,
  };
};

const dataCommon = {
  useList,
  vendorProperty,
};

export type FactoryType = typeof dataCommon;

export default dataCommon;
