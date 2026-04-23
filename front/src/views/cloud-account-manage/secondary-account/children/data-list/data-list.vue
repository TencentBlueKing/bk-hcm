<script setup lang="ts">
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { type ISecondaryAccountItem } from '@/store/cloud-account-manage/secondary-account';
import { AUTH_BIZ_UPDATE_SECONDARY_ACCOUNT } from '@/constants/auth-symbols';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';
import routeAction from '@/router/utils/action';

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: ISecondaryAccountItem[];
  pagination: PaginationType;
  loading?: boolean;
}

const props = withDefaults(defineProps<IDataListProps>(), {
  loading: false,
});

// 定义事件
const emit = defineEmits<{
  'view-details': [row: ISecondaryAccountItem];
  'edit-account': [row: ISecondaryAccountItem];
}>();

const { getBizsId } = useWhereAmI();
const { settings } = useTableSettings(props.columns);
const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

// 查看详情 - 触发事件
const handleViewDetails = (row: ISecondaryAccountItem) => {
  emit('view-details', row);
};

// 编辑账号 - 触发事件
const handleEditAccount = (row: ISecondaryAccountItem) => {
  emit('edit-account', row);
};

const handleGoToPage = (row: ISecondaryAccountItem, column: ModelPropertyColumn, type: string) => {
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
      row-key="id"
    >
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

          <template v-else-if="column.id === 'sub_account_count'">
            <display-value
              :property="column"
              :value="row?.sub_account_count ?? 0"
              :display="{
                appearance: 'link-button',
                appearanceProps: { isIcon: true, onClick: () => handleGoToPage(row, column, 'tertiary-account') },
              }"
            />
          </template>

          <template v-else-if="column.id === 'account_secret_count'">
            <display-value
              :property="column"
              :value="row?.account_secret_count ?? 0"
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
      <bk-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <hcm-auth
            v-if="getBizsId()"
            :sign="{ type: AUTH_BIZ_UPDATE_SECONDARY_ACCOUNT, relation: [getBizsId()] }"
            v-slot="{ noPerm }"
          >
            <bk-button theme="primary" text :disabled="noPerm" @click="handleEditAccount(row)">编辑</bk-button>
          </hcm-auth>
        </template>
      </bk-table-column>
    </bk-table>
  </bk-loading>
</template>

<style lang="scss" scoped>
// 状态标签基础样式
:deep(.status-tag) {
  display: inline-block;
  height: 18px;
  line-height: 18px;
  padding: 0 8px;
  border-radius: 9px;
  font-size: 12px;
}

// 同步成功 / 已纳管 - 绿色
:deep(.status-tag-success),
:deep(.status-tag-managed) {
  color: #2dcb56;
  background-color: #daf6e5;
}

// 同步失败 - 红色
:deep(.status-tag-failed) {
  color: #ea3636;
  background-color: #fdd;
}

// 未同步 / 未纳管 - 灰色
:deep(.status-tag-not-sync),
:deep(.status-tag-unmanaged) {
  color: #4d4f56;
  background-color: #f0f1f5;
}

// 同步中 - 蓝色
:deep(.status-tag-syncing) {
  color: #3a84ff;
  background-color: #e1ecff;
}

// 数量标签 - 蓝色
:deep(.count-tag) {
  display: inline-block;
  height: 18px;
  line-height: 18px;
  padding: 0 8px;
  border-radius: 9px;
  font-size: 12px;
  color: #3a84ff;
  background-color: #e1ecff;
}
</style>
