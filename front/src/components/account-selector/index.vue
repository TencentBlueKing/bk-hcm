<script lang="ts" setup>
import { computed, ref, defineExpose, PropType, useAttrs, watch, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useAccountStore } from '@/store';

import type {
  // PlainObject,
  FilterType,
} from '@/typings/resource';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { filterAccountList } from '@pluginHandler/account-selector';
import { QueryRuleOPEnum } from '@/typings';

const props = defineProps({
  bizId: Number as PropType<number>,
  filter: {
    type: Object as PropType<FilterType>,
    default() {
      return { op: 'and', rules: [] };
    },
  },
  type: String as PropType<string>,
  mustBiz: Boolean as PropType<boolean>,
  disabled: Boolean,
  optionDisabled: {
    type: Function as PropType<(accountItem?: any) => boolean>,
    default: () => false,
  },
});
const emit = defineEmits(['input', 'change']);

const attrs = useAttrs();

const accountStore = useAccountStore();
const accountList = ref([]);
const loading = ref(null);
const accountPage = ref(0);
const { whereAmI, isResourcePage } = useWhereAmI();
const route = useRoute();

const selectedValue = computed({
  get() {
    return '';
  },
  set(val) {
    emit('input', val);
  },
});

const getAccoutList = async (bizs?: number) => {
  if (props.mustBiz && !props.bizId && !bizs) {
    return;
  }

  if (loading.value === true) {
    return;
  }

  loading.value = true;
  const data = {
    filter: props.filter,
    page: {
      start: accountPage.value * 100,
      limit: 100,
    },
  };
  if (props.type) {
    if (isResourcePage) {
      data.filter.rules.push({ field: 'type', op: QueryRuleOPEnum.EQ, value: 'resource' });
    } else {
      data.params = { account_type: props.type };
    }
  }
  const res = await accountStore.getAccountList(data, props.bizId || bizs, isResourcePage);

  accountPage.value += 1;

  if (!isResourcePage) {
    accountList.value.push(...(res?.data || []));
  } else {
    accountList.value.push(...(res?.data?.details || []));
  }
  // 负载均衡、目标组、证书托管、参数模板这四个暂时只腾讯云支持
  if (
    (isResourcePage && route.query.type === 'certs') ||
    route.query.scene === 'template' ||
    (route.meta.isFilterAccount as boolean)
  ) {
    accountList.value = filterAccountList(accountList.value);
  }
  loading.value = false;
};

onMounted(() => {
  getAccoutList();
});

watch(
  () => whereAmI.value,
  (whereAmI) => {
    if (whereAmI === Senarios.business) {
      accountList.value = [];
      getAccoutList();
    }
  },
);

watch(
  () => accountStore.bizs,
  (bizs) => {
    getAccoutList(bizs);
  },
);

const handleChange = (val: string) => {
  const data = accountList.value.find((item) => item.id === val);
  emit('change', data);
};

defineExpose({
  accountList,
});
</script>

<template>
  <bk-select
    v-model="selectedValue"
    filterable
    @scroll-end="whereAmI === Senarios.business ? undefined : getAccoutList"
    :loading="loading"
    @change="handleChange"
    v-bind="attrs"
    :disabled="props.disabled"
  >
    <bk-option
      v-for="(item, index) in accountList"
      :disabled="optionDisabled(item)"
      :key="index"
      :value="item.id"
      :label="item.name"
    />
  </bk-select>
</template>
