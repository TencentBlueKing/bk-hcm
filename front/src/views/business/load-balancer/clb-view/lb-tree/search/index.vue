<script setup lang="ts">
import { ref } from 'vue';
import { SearchSelect } from 'bkui-vue';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';

import { VendorEnum, VendorMap } from '@/common/constant';
import { useI18n } from 'vue-i18n';
import { useRegionsStore } from '@/store/useRegionsStore';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { LB_NETWORK_TYPE_MAP } from '@/constants';

const emit = defineEmits<(e: 'search', rules: RulesItem[], searchValue: string) => void>();

const { t } = useI18n();
const regionsStore = useRegionsStore();

const searchValue = ref([]);
const searchData: Array<ISearchItem> = [
  {
    id: 'vendor',
    name: t('云厂商'),
    children: [{ id: VendorEnum.TCLOUD, name: VendorMap[VendorEnum.TCLOUD] }],
  },
  { id: 'name', name: t('负载均衡名称') },
  { id: 'lb_vip', name: t('负载均衡VIP') },
  { id: 'domain', name: t('负载均衡域名') },
  {
    id: 'lb_type',
    name: t('负载均衡网络类型'),
    children: Object.keys(LB_NETWORK_TYPE_MAP).map((lb_type) => ({
      id: lb_type,
      name: LB_NETWORK_TYPE_MAP[lb_type as keyof typeof LB_NETWORK_TYPE_MAP],
    })),
  },
  {
    id: 'ip_version',
    name: t('IP版本'),
    children: [
      { id: 'ipv4', name: 'IPv4' },
      { id: 'ipv6', name: 'IPv6' },
      { id: 'ipv6_dual_stack', name: 'IPv6DualStack' },
      { id: 'ipv6_nat64', name: 'IPv6Nat64' },
    ],
  },
];

const handleSelectKey = () => {
  searchValue.value = [];
};

const handleChange = () => {
  const getOp = (field: string) => {
    if (['name', 'domain'].includes(field)) return QueryRuleOPEnum.CIS;
    if (['lb_vip'].includes(field)) return QueryRuleOPEnum.JSON_OVERLAPS;
    return QueryRuleOPEnum.EQ;
  };

  if (searchValue.value.length === 0) {
    emit('search', [], '');
  } else {
    // 单条件查询
    const { id, values } = searchValue.value[0];
    let value = values[0].id;
    let rules: RulesItem[] = [{ field: id, op: getOp(id), value }];

    switch (id) {
      case 'lb_vip':
        value = [value];
        rules = [
          {
            op: QueryRuleOPEnum.OR,
            rules: [
              { field: 'private_ipv4_addresses', op: getOp(id), value },
              { field: 'private_ipv6_addresses', op: getOp(id), value },
              { field: 'public_ipv4_addresses', op: getOp(id), value },
              { field: 'public_ipv6_addresses', op: getOp(id), value },
            ],
          },
        ];
        break;
      case 'region':
        value = regionsStore.getRegionNameEN(value) || value;
        rules = [{ field: id, op: getOp(id), value }];
        break;
    }

    emit('search', rules, value);
  }
};
</script>

<template>
  <div class="search-wrapper">
    <SearchSelect
      v-model="searchValue"
      :data="searchData"
      unique-select
      value-behavior="need-key"
      @select-key="handleSelectKey"
      @update:model-value="handleChange"
    />
  </div>
</template>

<style scoped lang="scss">
.search-wrapper {
  height: 56px;
  padding: 12px 16px;
}
</style>
