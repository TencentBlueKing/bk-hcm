<script lang="ts" setup>
import { reactive, watch, ref, inject } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAccountStore, useResourceStore } from '@/store';
import {
  BusinessFormFilter,
  QueryFilterType,
  QueryRuleOPEnum,
} from '@/typings';
import { CLOUD_TYPE } from '@/constants';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';

const props = defineProps({
  hidden: {
    type: Array,
    default() {
      return [];
    },
  },
  type: {
    type: String,
    default() {
      return '';
    },
  },
});

const { t } = useI18n();
const accountStore = useAccountStore();
const resourceStore = useResourceStore();
const emit = defineEmits(['change']);
const accountList = ref([]);
const cloudRegionsList = ref([]);
const accountLoading = ref(false);
const cloudRegionsLoading = ref(false);
const cloudAreaPage = ref(0);
const { isResourcePage } = useWhereAmI();

const securityType: any = inject('securityType');
const resourceAccountStore = useResourceAccountStore();

const state = reactive<{ filter: BusinessFormFilter }>({
  filter: {
    vendor: '',
    account_id: '',
    region: '',
  },
});

const filter = ref<QueryFilterType>({
  op: 'and',
  rules: [],
});

watch(
  () => state.filter,
  (value) => {
    emit('change', value);
  },
  { deep: true },
);

watch(
  () => state.filter.vendor,
  (val) => {
    state.filter.region = '';
    cloudRegionsList.value = [];
    switch (val) {
      case VendorEnum.TCLOUD:
        filter.value.rules = [
          {
            field: 'vendor',
            op: QueryRuleOPEnum.EQ,
            value: val,
          },
          {
            field: 'status',
            op: QueryRuleOPEnum.EQ,
            value: 'AVAILABLE',
          },
        ];
        break;
      case VendorEnum.AWS:
        filter.value.rules = [
          {
            field: 'vendor',
            op: QueryRuleOPEnum.EQ,
            value: val,
          },
          {
            field: 'status',
            op: QueryRuleOPEnum.EQ,
            value: 'opt-in-not-required',
          },
        ];
        break;
      case VendorEnum.GCP:
        filter.value.rules = [
          {
            field: 'vendor',
            op: QueryRuleOPEnum.EQ,
            value: val,
          },
          {
            field: 'status',
            op: QueryRuleOPEnum.EQ,
            value: 'UP',
          },
        ];
        break;
      case VendorEnum.HUAWEI:
        filter.value.rules = [
          {
            field: 'type',
            op: QueryRuleOPEnum.EQ,
            value: 'public',
          },
          {
            field: 'service',
            op: QueryRuleOPEnum.EQ,
            value: 'vpc',
          },
        ];
        break;
      case VendorEnum.AZURE:
        filter.value.rules = [
          {
            field: 'type',
            op: QueryRuleOPEnum.EQ,
            value: 'Region',
          },
        ];
        break;
    }
    getCloudRegionList();
  },
);

watch(
  () => state.filter.account_id,
  (val) => {
    if (val) {
      state.filter.vendor = accountList.value.find((e: any) => {
        return e.id === val;
      }).vendor;
    }
  },
);

const getAccountList = async () => {
  // const rulesData = [];
  // if (state.filter.vendor) {
  //   rulesData.push({ field: 'vendor', op: 'eq', value: state.filter.vendor });
  // }
  try {
    accountLoading.value = true;
    const payload = isResourcePage
      ? {
        page: {
          count: false,
          limit: 100,
          start: 0,
        },
        filter: { op: 'and', rules: [] },
      }
      : {
        params: {
          account_type: 'resource',
        },
      };
    const res = await accountStore.getAccountList(payload, accountStore.bizs);
    if (resourceAccountStore.resourceAccount?.id) {
      accountList.value = res.data?.details
        .filter(({ id }: {id: string}) => id === resourceAccountStore.resourceAccount.id);
      // 自动填充当前账号
      state.filter.account_id = accountList.value?.[0].id;
      return;
    }
    accountList.value = isResourcePage ? res?.data?.details : res?.data;
    if (props.type === 'security') {
      // 安全组需要区分
      if (securityType.value && securityType.value === 'gcp') {
        accountList.value = accountList.value.filter(e => e.vendor === 'gcp');
      } else {
        accountList.value = accountList.value.filter(e => e.vendor !== 'gcp');
      }
    }
  } catch (error) {
    console.log(error);
  } finally {
    accountLoading.value = false;
  }
};

const getCloudRegionList = () => {
  if (cloudRegionsLoading.value || !state.filter.vendor) return;
  cloudRegionsLoading.value = true;
  resourceStore
    .getCloudRegion(state.filter.vendor, {
      filter: filter.value,
      page: {
        count: false,
        start: cloudAreaPage.value,
        limit: 100,
      },
    })
    .then((res: any) => {
      cloudAreaPage.value += 1;
      cloudRegionsList.value.push(...(res?.data?.details || []));
    })
    .finally(() => {
      cloudRegionsLoading.value = false;
    });
};

const formRef = ref(null);
const validate = () => {
  return formRef.value.validate();
};
defineExpose([validate]);

getAccountList();
</script>
<template>
  <bk-form class="pt20 bussine-form" label-width="150" :model="state.filter" ref="formRef">
    <bk-form-item
      :label="t('云账号')"
      class="item-warp"
      required
      property="account_id"
    >
      <bk-select
        class="item-warp-component"
        :loading="accountLoading"
        v-model="state.filter.account_id"
      >
        <bk-option
          v-for="(item, index) in accountList"
          :key="index"
          :value="item.id"
          :label="item.name"
        />
      </bk-select>
    </bk-form-item>
    <bk-form-item
      :label="t('云厂商')"
      class="item-warp"
      required
      property="vendor"
    >
      <bk-select
        disabled
        class="item-warp-component"
        v-model="state.filter.vendor"
      >
        <bk-option
          v-for="(item, index) in CLOUD_TYPE"
          :key="index"
          :value="item.id"
          :label="item.name"
        />
      </bk-select>
    </bk-form-item>
    <bk-form-item
      :label="t('云地域')"
      class="item-warp"
      v-if="!props.hidden.includes('region')"
      required
      property="region"
    >
      <bk-select
        class="item-warp-component"
        filterable
        :disabled="!state.filter.vendor"
        :loading="cloudRegionsLoading"
        v-model="state.filter.region"
      >
        <bk-option
          v-for="(item, index) in cloudRegionsList"
          :key="index"
          :value="
            state.filter.vendor === 'azure'
              ? item.name
              : item.region_id || item.id
          "
          :label="item.locales_zh_cn || item.region_name || item.region_id || item.name"
        />
      </bk-select>
    </bk-form-item>
  </bk-form>
</template>
<style lang="scss" scoped>
.bussine-form {
  padding-right: 20px;
}
</style>
