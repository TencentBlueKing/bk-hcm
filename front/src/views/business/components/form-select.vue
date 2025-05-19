<script lang="ts" setup>
import { reactive, watch, ref, inject, nextTick, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAccountStore, useResourceStore } from '@/store';
import { BusinessFormFilter, QueryFilterType, QueryRuleOPEnum, IAccountItem } from '@/typings';
import { CLOUD_TYPE } from '@/constants';
import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import AccountSelector from '@/components/account-selector/index-new.vue';

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
  show: Boolean,
});

const { t } = useI18n();
const accountStore = useAccountStore();
const resourceStore = useResourceStore();
const emit = defineEmits(['change']);
const cloudRegionsList = ref([]);
const accountLoading = ref(false);
const cloudRegionsLoading = ref(false);
const cloudAreaPage = ref(0);
const accountSelector = ref(null);
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

const handleChange = (account: IAccountItem) => getAccountList(account);
const getAccountList = (account: Partial<IAccountItem>) => {
  accountLoading.value = true;
  state.filter.account_id = account?.id ?? '';
  state.filter.vendor = account?.vendor ?? '';
  setOptionDisabled();
};
const optionDisabled = ref<() => boolean>(() => false);
const setOptionDisabled = () => {
  if (props.type === 'security') {
    // 安全组需要区分
    if (securityType.value && securityType.value === 'gcp') {
      optionDisabled.value = (account?: IAccountItem) => account.vendor !== 'gcp';
    } else {
      optionDisabled.value = (account?: IAccountItem) => account.vendor === 'gcp';
    }
  } else {
    optionDisabled.value = () => false;
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
const resetForm = () => {
  state.filter.account_id = '';
  state.filter.vendor = '';
  nextTick(() => formRef.value.clearValidate());
};

watch(
  () => props.show,
  (val) => {
    if (val) {
      // 业务下或资源未选择指定账号情况下为空
      return getAccountList({
        vendor: resourceAccountStore.resourceAccount?.vendor,
        id: resourceAccountStore.resourceAccount?.id,
      });
    }
    return resetForm();
  },
  {
    immediate: true,
  },
);

const isSecurityGroup = computed(() => securityType.value === 'group' && props.type === 'security');

defineExpose([validate]);
</script>
<template>
  <bk-form
    class="pt20 bussine-form"
    :class="{ 'group-bussine-form': isSecurityGroup }"
    label-width="150"
    :form-type="isSecurityGroup ? 'vertical' : 'default'"
    :model="state.filter"
    ref="formRef"
  >
    <bk-form-item :label="t('云账号')" class="item-warp" required property="account_id">
      <AccountSelector
        ref="accountSelector"
        v-model="state.filter.account_id"
        :biz-id="isResourcePage ? undefined : accountStore.bizs"
        :disabled="isResourcePage"
        :option-disabled="optionDisabled"
        :placeholder="isResourcePage ? t('请在左侧选择账号') : undefined"
        @change="handleChange"
      />
    </bk-form-item>
    <bk-form-item
      :label="t('云厂商')"
      class="item-warp"
      required
      property="vendor"
      v-if="!props.hidden.includes('vendor')"
    >
      <bk-select disabled class="item-warp-component" v-model="state.filter.vendor">
        <bk-option v-for="(item, index) in CLOUD_TYPE" :key="index" :value="item.id" :label="item.name" />
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
          :value="state.filter.vendor === 'azure' ? item.name : item.region_id || item.id"
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

.group-bussine-form {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding-right: 0;
}

.item-warp {
  flex-basis: 50%;
}
</style>
