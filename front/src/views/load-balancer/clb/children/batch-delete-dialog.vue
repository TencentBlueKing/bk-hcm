<script setup lang="ts">
import { computed, h, inject, Ref, ref } from 'vue';
import { ILoadBalancerWithDeleteProtectionItem, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { ThemeEnum } from 'bkui-vue/lib/shared';
import { DisplayFieldFactory, DisplayFieldType } from '../../children/display/field-factory';
import { ModelPropertyColumn } from '@/model/typings';
import { ISearchSelectValue, SortType } from '@/typings';
import { LB_TYPE_NAME, LoadBalancerType } from '../../constants';
import { ConditionKeyType, SearchConditionFactory } from '../../children/search/condition-factory';
import usePage from '@/hooks/use-page';
import { cloneDeep } from 'lodash';
import { getInstVip, parseIP } from '@/utils';
import { getLocalFilterFnBySearchSelect } from '@/utils/search';

import { Message, Tag } from 'bkui-vue';
import Search from '../../children/search/search.vue';
import DataList from '../../children/display/data-list.vue';
import ModalFooter from '@/components/modal/modal-footer.vue';

interface IProps {
  selections: ILoadBalancerWithDeleteProtectionItem[];
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits<{ 'confirm-success': [] }>();

const loadBalancerClbStore = useLoadBalancerClbStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const displayFieldProperties = DisplayFieldFactory.createModel(DisplayFieldType.CLB).getProperties();
const displayFieldIds = [
  'name',
  'cloud_id',
  'domain',
  'lb_vip',
  'region',
  'lb_type',
  'listener_count',
  'delete_protect',
];
const displayFieldConfig: Record<string, Partial<ModelPropertyColumn>> = {
  lb_vip: {
    render: ({ row }) => getInstVip(row),
  },
  lb_type: {
    render: ({ cell }) => {
      return h(
        Tag,
        { radius: '11px', class: ['lb-type-tag', cell === LoadBalancerType.OPEN ? 'is-open' : 'is-internal'] },
        LB_TYPE_NAME[cell as LoadBalancerType],
      );
    },
  },
  delete_protect: {
    render: ({ cell }) => {
      return h(Tag, { theme: cell ? 'success' : undefined }, cell ? '开启' : '关闭');
    },
  },
};
const datalistColumns = displayFieldIds.map((id) => {
  const property = displayFieldProperties.find((item) => item.id === id);
  return { ...property, ...displayFieldConfig[id] };
});

const conditionProperties = SearchConditionFactory.createModel(ConditionKeyType.CLB).getProperties();
const conditionIds = ['name', 'cloud_id', 'domain', 'lb_vip', 'lb_type'];
const searchFields = conditionIds.map((id) => conditionProperties.find((item) => item.id === id));

const list = ref(cloneDeep(props.selections));
const { pagination } = usePage(false);

const canDeletePredicate = (item: ILoadBalancerWithDeleteProtectionItem) => {
  return item.listener_count === 0 && !item.delete_protect;
};

const active = ref(props.selections.every(canDeletePredicate));
const localSearchFilter = ref<(item: ILoadBalancerWithDeleteProtectionItem) => boolean>(() => true);
const sort = ref<SortType>();
const handleSearch = (searchValue: ISearchSelectValue) => {
  localSearchFilter.value = getLocalFilterFnBySearchSelect(searchValue, [
    {
      field: 'lb_vip',
      checker: (_key, values, item) => {
        const ipv4 = [...item.private_ipv4_addresses, ...item.public_ipv4_addresses];
        const ipv6 = [...item.private_ipv6_addresses, ...item.public_ipv6_addresses];

        const ipv4Set = new Set<string>();
        const ipv6Set = new Set<string>();
        values.forEach((item) => {
          const { IPv4List, IPv6List } = parseIP(item);
          IPv4List.forEach((item) => ipv4Set.add(item));
          IPv6List.forEach((item) => ipv6Set.add(item));
        });

        const hasIPv4Intersection = ipv4.some((ip) => ipv4Set.has(ip));
        const hasIPv6Intersection = ipv6.some((ip) => ipv6Set.has(ip));

        return hasIPv4Intersection || hasIPv6Intersection;
      },
    },
  ]);
};
const handleSort = (sortType: SortType) => {
  sort.value = sortType;
};

const displayList = computed(() => {
  // 第一步：根据active的值进行过滤
  let result = active.value
    ? list.value.filter(canDeletePredicate)
    : list.value.filter((item) => item.listener_count !== 0 || item.delete_protect);

  // 第二步：根据搜索条件过滤
  const filterResult = result.filter(localSearchFilter.value);
  result = result.filter(localSearchFilter.value);

  // 第三步：如果有排序，则进行排序
  // TODO: 排序可以做成通用的
  if (!sort.value || sort.value.type === 'null') {
    return filterResult;
  }

  const { type, column } = sort.value;
  const { field } = column;

  result = [...result].sort((a, b) => {
    const aVal = a[field];
    const bVal = b[field];

    // 处理字符串比较
    if (typeof aVal === 'string' && typeof bVal === 'string') {
      return type === 'asc' ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
    }

    // 处理布尔值比较
    if (typeof aVal === 'boolean' && typeof bVal === 'boolean') {
      const aNum = aVal ? 1 : 0;
      const bNum = bVal ? 1 : 0;
      return type === 'asc' ? aNum - bNum : bNum - aNum;
    }

    // 处理数字比较
    if (typeof aVal === 'number' && typeof bVal === 'number') {
      return type === 'asc' ? aVal - bVal : bVal - aVal;
    }

    // 默认比较（包括混合类型）
    return type === 'asc' ? String(aVal).localeCompare(String(bVal)) : String(bVal).localeCompare(String(aVal));
  });

  return result;
});
const hasListenerLoadBalancerCount = computed(() => list.value.filter((item) => item.listener_count !== 0).length);
const hasDeleteProtectCount = computed(() => list.value.filter((item) => item.delete_protect).length);
const canDelete = computed(() => list.value.some(canDeletePredicate));

const handleSingleDelete = (row: ILoadBalancerWithDeleteProtectionItem) => {
  const idx = list.value.findIndex((item) => item.id === row.id);
  if (idx > -1) {
    list.value.splice(idx, 1);
  }
};

const handleConfirm = async () => {
  await loadBalancerClbStore.batchDeleteLoadBalancer(
    { ids: list.value.filter(canDeletePredicate).map((item) => item.id) },
    currentGlobalBusinessId.value,
  );
  Message({ theme: 'success', message: '删除成功' });
  handleClosed();
  emit('confirm-success');
};

const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" title="批量删除负载均衡" width="60vw" class="batch-delete-clb-dialog">
    <div class="mb12">
      已选择
      <span class="text-primary">{{ list.length }}</span>
      个负载均衡，其中
      <span class="text-danger">{{ hasListenerLoadBalancerCount }}</span>
      个存在监听器、
      <span class="text-danger">{{ hasDeleteProtectCount }}</span>
      个负载均衡开启了删除保护，不可删除。
    </div>
    <div class="toolbar">
      <bk-radio-group v-model="active">
        <bk-radio-button :label="true">可删除</bk-radio-button>
        <bk-radio-button :label="false">不可删除</bk-radio-button>
      </bk-radio-group>
      <search class="search" :fields="searchFields" @search="handleSearch" />
    </div>
    <data-list
      class="data-list"
      :columns="datalistColumns"
      :list="displayList"
      :enable-query="false"
      :pagination="pagination"
      :remote-pagination="false"
      :max-height="500"
      @column-sort="handleSort"
    >
      <template #action v-if="active">
        <bk-table-column label="">
          <template #default="{ row }">
            <bk-button text class="single-delete-btn" @click="handleSingleDelete(row)">
              <i class="hcm-icon bkhcm-icon-minus-circle-shape"></i>
            </bk-button>
          </template>
        </bk-table-column>
      </template>
    </data-list>
    <template #footer>
      <modal-footer
        confirm-text="删除"
        :confirm-button-theme="ThemeEnum.DANGER"
        :loading="loadBalancerClbStore.batchDeleteLoadBalancerLoading"
        :disabled="!canDelete"
        @confirm="handleConfirm"
        @closed="handleClosed"
      />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.batch-delete-clb-dialog {
  .toolbar {
    margin-bottom: 12px;
    display: flex;
    align-items: center;

    .search {
      width: 300px;
      margin-left: auto;
    }
  }

  .data-list {
    :deep(.lb-type-tag) {
      &.is-open {
        background-color: #d8edd9;
      }

      &.is-internal {
        background-color: #fff2c9;
      }
    }

    .single-delete-btn {
      color: #c4c6cc;
    }
  }
}
</style>
