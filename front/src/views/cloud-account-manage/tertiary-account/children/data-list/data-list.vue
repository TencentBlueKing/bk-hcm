<script setup lang="ts">
import { h, ref, watch } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { Button } from 'bkui-vue';
import type { ISubAccountItem } from '@/store/cloud-account';
import BusinessValue from '@/components/display-value/business-value.vue';

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: ISubAccountItem[];
  pagination: PaginationType;
  loading?: boolean;
}

const props = withDefaults(defineProps<IDataListProps>(), {
  loading: false,
});

const emit = defineEmits<{
  'view-details': [row: ISubAccountItem];
  'edit-account': [row: ISubAccountItem];
  'delete-account': [row: ISubAccountItem];
  'selection-change': [selection: ISubAccountItem[]];
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

// 格式化手机号（脱敏处理）
const formatPhone = (phone: string) => {
  if (!phone) return '--';
  if (phone.length < 7) return phone;
  return `${phone.substring(0, 3)}****${phone.substring(phone.length - 4)}`;
};

// 格式化数组展示
const formatArray = (arr: any[]) => {
  if (!arr || !arr.length) return '--';
  return arr.join(', ');
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

// 列表数据变化时自动清空选中（与 business/host-manage 的 completeCallback 模式一致）
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

// 自定义渲染列
const getColumnRender = (column: ModelPropertyColumn) => {
  // 名称列 - 点击打开详情侧栏
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
  // 邮箱列 - 脱敏处理
  if (column.id === 'email') {
    return ({ row }: { row: ISubAccountItem }) => formatEmail(row.email);
  }
  // 手机号列 - 脱敏处理
  if (column.id === 'phone_num') {
    return ({ row }: { row: ISubAccountItem }) => formatPhone(row.phone_num);
  }
  // 负责人列 - 数组展示
  if (column.id === 'managers') {
    return ({ row }: { row: ISubAccountItem }) => formatArray(row.managers);
  }
  // 所属业务列 - 使用 BusinessValue 组件
  if (column.id === 'bk_biz_ids') {
    return ({ row }: { row: ISubAccountItem }) =>
      h(BusinessValue, {
        value: row.bk_biz_ids,
        display: { appearance: 'tag' },
      });
  }
  // 密钥数 - 蓝色数字
  if (column.id === 'sub_account_secret_count') {
    return ({ row }: { row: ISubAccountItem }) => {
      const value = row.sub_account_secret_count ?? 0;
      return h('span', { style: { color: '#3A84FF' } }, value);
    };
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
      :max-height="`calc(100vh - 400px)`"
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
          <bk-button theme="primary" text @click="handleEditAccount(row)">编辑</bk-button>
          <bk-button theme="primary" text style="margin-left: 8px" @click="handleDeleteAccount(row)">删除</bk-button>
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
