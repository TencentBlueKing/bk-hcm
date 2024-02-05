<template>
  <Loading :loading="isLoading">
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <BatchDistribution
        :selections="selections"
        :type="DResourceType.clbs"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <Button class="mw88">批量删除</Button>
      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <bk-search-select class="w500" clearable :conditions="[]" :data="clbsSearchData" v-model="searchValue" />
        <slot name="recycleHistory"></slot>
      </div>
    </section>
    <Table
      class="has-selection"
      :columns="columns"
      :data="datas"
      :settings="settings"
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

    <Dialog
      :is-show="isDialogShow"
      :title="t('负载均衡分配')"
      :theme="'primary'"
      quick-close
      @closed="() => (isDialogShow = false)"
      @confirm="handleConfirm"
      :is-loading="isDialogBtnLoading"
    >
      <p class="selected-host-count-tip">
        已选择
        <span class="selected-host-count">{{ selections.length }}</span>
        台负载均衡，可选择所需分配的目标业务
      </p>
      <p class="mb6">目标业务</p>
      <BusinessSelector v-model="selectedBizId" :authed="true" class="mb32" :auto-select="true" />
    </Dialog>
  </Loading>
</template>

<script setup lang="ts">
import { PropType, ref, computed } from 'vue';
import { Loading, Table, Button, Dialog } from 'bkui-vue';
import { BatchDistribution, DResourceType } from '../dialog/batch-distribution';
import BusinessSelector from '@/components/business-selector/index.vue';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import type { DoublePlainObject, FilterType } from '@/typings/resource';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import { useResourceStore } from '@/store';
import { useI18n } from 'vue-i18n';

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

const { t } = useI18n();
const { whereAmI } = useWhereAmI();
const { searchData, searchValue, filter } = useFilter(props);
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'clbs',
);
const { selections, handleSelectionChange, resetSelections } = useSelection();
const { columns, settings } = useColumns('clbs');
const resourceStore = useResourceStore();

const selectedBizId = ref(0);
const isDialogShow = ref(false);
const isDialogBtnLoading = ref(false);

const clbsSearchData = computed(() => [
  {
    name: '负载均衡ID',
    id: 'cloud_id',
  },
  ...searchData.value,
]);

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

const handleConfirm = async () => {
  isDialogBtnLoading.value = true;
  await resourceStore.assignBusiness('clbs', {
    clb_ids: selections.value?.map((v) => v.id) || [],
    bk_biz_id: selectedBizId.value,
  });
  triggerApi();
  isDialogBtnLoading.value = false;
  isDialogShow.value = false;
};
</script>

<style lang="scss" scoped>
.mr15 {
  margin-right: 15px;
}

.search-selector-container {
  margin-left: auto;
}
</style>
