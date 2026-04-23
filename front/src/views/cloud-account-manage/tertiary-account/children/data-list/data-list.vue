<script setup lang="ts">
import { ref, watch } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import routeAction from '@/router/utils/action';
import { useAccountStore } from '@/store/account';
import { AUTH_BIZ_DELETE_SUB_ACCOUNT, AUTH_BIZ_UPDATE_SUB_ACCOUNT } from '@/constants/auth-symbols';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';
import type { ISubAccountItem } from '@/store/cloud-account-manage/tertiary-account';

const props = withDefaults(defineProps<IDataListProps>(), {
  loading: false,
});

const emit = defineEmits<{
  'view-details': [row: ISubAccountItem];
  'edit-account': [row: ISubAccountItem];
  'delete-account': [row: ISubAccountItem];
  'selection-change': [selection: ISubAccountItem[]];
}>();

const accountStore = useAccountStore();

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: ISubAccountItem[];
  pagination: PaginationType;
  loading?: boolean;
}

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);

const formatPhone = (row: { phone_num: string; country_code: string }) => {
  const { phone_num, country_code } = row;
  if (!phone_num) return '--';
  const codeStr = country_code ? `+${country_code}` : '';
  return `${codeStr}${phone_num}`;
};
const handleViewDetails = (row: ISubAccountItem) => {
  emit('view-details', row);
};

const handleEditAccount = (row: ISubAccountItem) => {
  emit('edit-account', row);
};

const handleDeleteAccount = (row: ISubAccountItem) => {
  emit('delete-account', row);
};

const { selections, handleSelectionChange, resetSelections } = useSelection();

const isCurRowSelectEnable = (row: ISubAccountItem) => row.operable !== false;

const isRowSelectEnable = ({ row, isCheckAll }: ISubAccountItem) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};

const tableRef = ref();

watch(
  () => props.list,
  () => {
    resetSelections();
    tableRef.value?.clearSelection();
  },
);

watch(
  () => selections.value,
  (val) => {
    emit('selection-change', val);
  },
  { deep: true },
);

const handleGoToPage = (row: ISubAccountItem, column: ModelPropertyColumn, type: string) => {
  const { searchField, valueField } = column?.meta?.search?.props;
  routeAction.open({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: { type, filter: `${searchField}=${row?.[valueField]}` },
  });
};
</script>

<template>
  <bk-loading :loading="loading">
    <bk-table
      ref="tableRef"
      row-hover="auto"
      :data="list"
      :pagination="pagination"
      :max-height="`calc(100vh - 500px)`"
      :settings="settings"
      :is-row-select-enable="isRowSelectEnable"
      remote-pagination
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="(selection: any) => handleSelectionChange(selection, isCurRowSelectEnable)"
      @select-all="(selection: any) => handleSelectionChange(selection, isCurRowSelectEnable, true)"
      row-key="id"
    >
      <bk-table-column type="selection" width="50" fixed="left" />
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :sort="column.sort"
        :width="column.width"
        :min-width="column.minWidth"
        :fixed="column.fixed"
        v-bind="column"
      >
        <template #default="{ row }">
          <template v-if="column.id === 'name'">
            <bk-button theme="primary" text @click="handleViewDetails(row)">{{ row.name || '--' }}</bk-button>
          </template>
          <template v-else-if="column.id === 'phone_num'">
            {{ formatPhone(row) }}
          </template>
          <template v-else-if="column.id === 'permission_template_count'">
            <display-value
              :property="column"
              :value="row?.permission_templates?.length ?? 0"
              :display="{
                appearance: 'link-button',
                appearanceProps: { isIcon: true, onClick: () => handleGoToPage(row, column, 'permission-template') },
              }"
            />
          </template>
          <template v-else-if="column.id === 'sub_account_secret_count'">
            <display-value
              :property="column"
              :value="row.sub_account_secret_count ?? 0"
              :display="{
                appearance: 'link-button',
                appearanceProps: { isIcon: true, onClick: () => handleGoToPage(row, column, 'cloud-secret') },
              }"
            />
          </template>
          <template v-else>
            <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
          </template>
        </template>
      </bk-table-column>
      <bk-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <hcm-auth
            v-if="accountStore.bizs"
            :sign="{ type: AUTH_BIZ_UPDATE_SUB_ACCOUNT, relation: [accountStore.bizs] }"
            v-slot="{ noPerm }"
          >
            <bk-button
              theme="primary"
              text
              :disabled="noPerm || row.operable === false"
              @click="handleEditAccount(row)"
            >
              编辑
            </bk-button>
          </hcm-auth>
          <hcm-auth
            v-if="accountStore.bizs"
            :sign="{ type: AUTH_BIZ_DELETE_SUB_ACCOUNT, relation: [accountStore.bizs] }"
            v-slot="{ noPerm }"
          >
            <bk-button
              theme="primary"
              text
              style="margin-left: 8px"
              :disabled="noPerm || row.operable === false"
              @click="handleDeleteAccount(row)"
            >
              删除
            </bk-button>
          </hcm-auth>
        </template>
      </bk-table-column>
    </bk-table>
  </bk-loading>
</template>

<style lang="scss" scoped>
:deep(.status-tag) {
  display: inline-block;
  height: 18px;
  line-height: 18px;
  padding: 0 8px;
  border-radius: 9px;
  font-size: 12px;
}
</style>
