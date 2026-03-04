<script setup lang="ts">
import { h } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { Button } from 'bkui-vue';
import type { ISecondaryAccountItem } from '@/store/cloud-account';
import BusinessValue from '@/components/display-value/business-value.vue';

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

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);

// 格式化邮箱展示（脱敏处理）
const formatEmail = (email: string) => {
  if (!email) return '--';
  const atIndex = email.indexOf('@');
  if (atIndex <= 3) return email;
  const prefix = email.substring(0, 3);
  const suffix = email.substring(atIndex);
  return `${prefix}***${suffix}`;
};

// 格式化数组展示
const formatArray = (arr: any[]) => {
  if (!arr || !arr.length) return '--';
  return arr.join(', ');
};

// 查看详情 - 触发事件
const handleViewDetails = (row: ISecondaryAccountItem) => {
  emit('view-details', row);
};

// 编辑账号 - 触发事件
const handleEditAccount = (row: ISecondaryAccountItem) => {
  emit('edit-account', row);
};

// 自定义渲染列
const getColumnRender = (column: ModelPropertyColumn) => {
  // 名称列 - 点击打开详情侧栏
  if (column.id === 'name') {
    return ({ row }: { row: ISecondaryAccountItem }) =>
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
  // 邮箱列 - 脱敏处理
  if (column.id === 'email') {
    return ({ row }: { row: ISecondaryAccountItem }) => formatEmail(row.email);
  }
  // 负责人、安全负责人列 - 数组展示
  if (column.id === 'managers' || column.id === 'security_managers') {
    return ({ row }: { row: ISecondaryAccountItem }) => formatArray(row[column.id]);
  }
  // 使用业务列 - 使用 BusinessValue 组件，tag 模式展示
  if (column.id === 'usage_biz_ids') {
    return ({ row }: { row: ISecondaryAccountItem }) =>
      h(BusinessValue, {
        value: row.usage_biz_ids,
        display: { appearance: 'tag' },
      });
  }
  // 三级账号数、密钥数 - 普通标签样式
  if (column.id === 'sub_account_count' || column.id === 'account_secret_count') {
    return ({ row }: { row: ISecondaryAccountItem }) => {
      const value = row[column.id as keyof ISecondaryAccountItem] ?? 0;
      return h('span', { style: { color: '#3A84FF' } }, value);
    };
  }
  // 资源纳管状态列 - 标签样式
  if (column.id === 'sync_status') {
    return ({ row }: { row: ISecondaryAccountItem }) => {
      const statusMap: Record<string, { class: string; text: string }> = {
        sync_success: { class: 'status-tag status-tag-success', text: '同步成功' },
        sync_failed: { class: 'status-tag status-tag-failed', text: '同步失败' },
        not_sync: { class: 'status-tag status-tag-not-sync', text: '未同步' },
        syncing: { class: 'status-tag status-tag-syncing', text: '同步中' },
        managed: { class: 'status-tag status-tag-managed', text: '已纳管' },
        unmanaged: { class: 'status-tag status-tag-unmanaged', text: '未纳管' },
      };
      const status = statusMap[row.sync_status] || { class: '', text: row.sync_status || '--' };
      return h('span', { class: status.class }, status.text);
    };
  }
  return null;
};
</script>

<template>
  <bk-loading :loading="loading">
    <bk-table
      row-hover="auto"
      :data="list"
      :pagination="pagination"
      :max-height="`calc(100vh - 400px)`"
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
          <template v-if="getColumnRender(column)">
            <component :is="() => getColumnRender(column)({ row })" />
          </template>
          <template v-else>
            <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
          </template>
        </template>
      </bk-table-column>
      <bk-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <bk-button theme="primary" text @click="handleEditAccount(row)">编辑</bk-button>
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
