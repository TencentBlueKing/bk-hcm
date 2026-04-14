<script setup lang="ts">
import { h, ref, watch } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { Button } from 'bkui-vue';
import type { ISubAccountItem } from '@/store/cloud-account';
import { useAccountStore } from '@/store/account';
import { AUTH_UPDATE_SUB_ACCOUNT, AUTH_DELETE_SUB_ACCOUNT } from '@/constants/auth-symbols';

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

const formatEmail = (email: string) => {
  if (!email) return '--';
  const atIndex = email.indexOf('@');
  if (atIndex <= 3) return email;
  const prefix = email.substring(0, 3);
  const suffix = email.substring(atIndex);
  return `${prefix}***${suffix}`;
};

const formatPhone = (phone: string) => {
  if (!phone) return '--';
  if (phone.length < 7) return phone;
  return `${phone.substring(0, 3)}****${phone.substring(phone.length - 4)}`;
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

const isCurRowSelectEnable = (_row: any) => true;

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

const getColumnRender = (column: ModelPropertyColumn) => {
  if (column.id === 'name') {
    return ({ row }: { row: ISubAccountItem }) =>
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick: () => handleViewDetails(row),
        },
        () => row.name || '--',
      );
  }
  if (column.id === 'email') {
    return ({ row }: { row: ISubAccountItem }) => formatEmail(row.email);
  }
  if (column.id === 'phone_num') {
    return ({ row }: { row: ISubAccountItem }) => formatPhone(row.phone_num);
  }
  return null;
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
          <template v-if="getColumnRender(column)">
            <component :is="() => getColumnRender(column)({ row })" />
          </template>
          <template v-else>
            <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
          </template>
        </template>
      </bk-table-column>
      <bk-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <hcm-auth :sign="{ type: AUTH_UPDATE_SUB_ACCOUNT, relation: [accountStore.bizs] }" v-slot="{ noPerm }">
            <bk-button
              theme="primary"
              text
              :disabled="noPerm || row.operable === false"
              @click="handleEditAccount(row)"
            >
              编辑
            </bk-button>
          </hcm-auth>
          <hcm-auth :sign="{ type: AUTH_DELETE_SUB_ACCOUNT, relation: [accountStore.bizs] }" v-slot="{ noPerm }">
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
