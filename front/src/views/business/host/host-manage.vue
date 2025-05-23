<script setup lang="ts">
import type { DoublePlainObject, FilterType } from '@/typings';
import { PropType, ref } from 'vue';
import useTableSelection from '@/hooks/useTableSelection';
import businessHostManagePlugin from '@pluginHandler/business-host-manage';
import useFilterHost from '@/views/resource/resource-manage/hooks/use-filter-host';
import { useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { ResourceTypeEnum } from '@/common/resource-constant';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';
import { ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { parseIP } from '@/utils';

const { useColumns, useTableListQuery, HostOperations } = businessHostManagePlugin;

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  whereAmI: {
    type: String,
  },
});

const isLoadingCloudAreas = ref(false);
const cloudAreaPage = ref(0);
const cloudAreas = ref([]);
const { whereAmI, isResourcePage } = useWhereAmI();

const { searchValue, filter } = useFilterHost(props);
const validateValues: ValidateValuesFunc = async (item, values) => {
  if (!item) return '请选择条件';
  // IP值为单选，这里可以简单处理（即便是多IP搜索，粘贴上去也是一个值）
  if (['private_ip', 'public_ip'].includes(item.id)) {
    const { IPv4List, IPv6List } = parseIP(values[0].id);
    return Boolean(IPv4List.length || IPv6List.length) ? true : 'IP格式有误';
  }
  return true;
};

const { selections, handleSelectionChange, resetSelections } = useTableSelection();

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } =
  useTableListQuery({ filter: filter.value }, 'cvms', () => {
    resetSelections();
  });
// 主机列表分页支持500条
Object.assign(pagination.value, { 'limit-list': [10, 20, 50, 100, 500] });

const hostOperationRef = ref(null);
const tableRef = ref(null);
const { columns, generateColumnsSettings } = useColumns({
  extra: {
    isLoading,
    triggerApi,
    getHostOperationRef: () => hostOperationRef,
    getTableRef: () => tableRef,
  },
});
const resourceStore = useResourceStore();

const tableSettings = generateColumnsSettings(columns);

const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (whereAmI.value === Senarios.business) return true;
  if (row.id) {
    return row.bk_biz_id === -1;
  }
};

const getCloudAreas = () => {
  if (isLoadingCloudAreas.value) return;
  isLoadingCloudAreas.value = true;
  resourceStore
    .getCloudAreas({
      page: {
        start: cloudAreaPage.value,
        limit: 100,
      },
    })
    .then((res: any) => {
      cloudAreaPage.value += 1;
      cloudAreas.value.push(...(res?.data?.info || []));
    })
    .finally(() => {
      isLoadingCloudAreas.value = false;
    });
};

getCloudAreas();
</script>

<template>
  <bk-loading :loading="isLoading" opacity="1">
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <HostOperations
        ref="hostOperationRef"
        :selections="selections"
        :on-finished="(type: 'confirm' | 'cancel' = 'confirm') => {
        if(type === 'confirm') triggerApi();
        resetSelections();
      }"
      ></HostOperations>

      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <resource-search-select
          v-model="searchValue"
          :resource-type="ResourceTypeEnum.CVM"
          :validate-values="validateValues"
        />
        <slot name="recycleHistory"></slot>
      </div>
    </section>

    <bk-table
      ref="tableRef"
      row-hover="auto"
      :columns="columns"
      :data="datas"
      :settings="tableSettings"
      :pagination="pagination"
      remote-pagination
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
      @column-sort="handleSort"
      row-key="id"
    />
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.mt20 {
  margin-top: 20px;
}
.mb32 {
  margin-bottom: 32px;
}
.distribution-cls {
  display: flex;
  align-items: center;
}
.mr10 {
  margin-right: 10px;
}
.search-selector-container {
  margin-left: auto;
}
:deep(.operation-column) {
  height: 100%;
  display: flex;
  align-items: center;

  .more-action {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    cursor: pointer;

    & > i {
      position: absolute;
    }

    &:hover {
      background-color: #f0f1f5;
    }

    &.current-operate-row {
      background-color: #f0f1f5;
    }

    &.disabled {
      background-color: #fff;
      color: #dcdee5;
      cursor: not-allowed;
    }
  }
}
.selected-host-info {
  margin-bottom: 16px;
}
</style>

<style lang="scss">
.more-action-item {
  &.disabled {
    color: #dcdee5;
    cursor: not-allowed;
  }
}
</style>
