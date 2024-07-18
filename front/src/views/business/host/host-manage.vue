<script setup lang="ts">
import type { DoublePlainObject, FilterType } from '@/typings';

import { PropType, ref, computed } from 'vue';
import useTableSelection from '@/hooks/useTableSelection';
import businessHostManagePlugin from '@pluginHandler/business-host-manage';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import { useResourceStore } from '@/store';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

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

const { searchData, searchValue, filter } = useFilter(props);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } =
  useTableListQuery({ filter: filter.value });

const { selections, handleSelectionChange, resetSelections } = useTableSelection();

const { columns, generateColumnsSettings } = useColumns({
  extra: {
    isLoading,
    triggerApi,
  },
});
const resourceStore = useResourceStore();

const hostSearchData = computed(() => {
  return [
    {
      name: '主机ID',
      id: 'cloud_id',
    },
    ...searchData.value,
    ...[
      {
        name: '管控区域',
        id: 'bk_cloud_id',
      },
      {
        name: '操作系统',
        id: 'os_name',
      },
      {
        name: '云地域',
        id: 'region',
      },
      {
        name: '公网IP',
        id: 'public_ipv4_addresses',
      },
      {
        name: '内网IP',
        id: 'private_ipv4_addresses',
      },
      {
        name: '所属VPC',
        id: 'cloud_vpc_ids',
      },
    ],
  ];
});

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
        :selections="selections"
        :on-finished="(type: 'confirm' | 'cancel' = 'confirm') => {
        if(type === 'confirm') triggerApi();
        resetSelections();
      }"
      ></HostOperations>

      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <bk-search-select class="w500" clearable :conditions="[]" :data="hostSearchData" v-model="searchValue" />
        <slot name="recycleHistory"></slot>
      </div>
    </section>

    <bk-table
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
