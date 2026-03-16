<script setup lang="ts">
import { h } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { Button } from 'bkui-vue';
import type { IPermissionPolicyItem, IRelatedAccount } from '../../typings';

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: IPermissionPolicyItem[];
  pagination: PaginationType;
  loading?: boolean;
}

const props = withDefaults(defineProps<IDataListProps>(), {
  loading: false,
});

// 定义事件
const emit = defineEmits<{
  'view-details': [row: IPermissionPolicyItem];
  'apply-to-account': [row: IPermissionPolicyItem];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);

// 查看详情
const handleViewDetails = (row: IPermissionPolicyItem) => {
  emit('view-details', row);
};

// 应用到二级账号
const handleApplyToAccount = (row: IPermissionPolicyItem) => {
  emit('apply-to-account', row);
};

// 跳转二级账号详情（新开标签页）
const handleGoToAccount = (account: IRelatedAccount) => {
  // TODO: 替换为真实路由，跳转到三级账号页面
  const url = `${window.location.origin}/#/cloud-account-manage/secondary-account/${account.account_id}`;
  window.open(url, '_blank');
};

// 判断是否为需要自定义渲染的列（排除 related_account_count，它在 template 中单独处理）
const getColumnRender = (column: ModelPropertyColumn) => {
  if (column.id === 'name') {
    return ({ row }: { row: IPermissionPolicyItem }) =>
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
  return null;
};

// 判断列是否为关联二级账号数
const isRelatedAccountColumn = (column: ModelPropertyColumn) => column.id === 'related_account_count';
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
          <!-- 关联二级账号数 - hover 弹出账号列表 -->
          <template v-if="isRelatedAccountColumn(column)">
            <bk-popover
              v-if="row.related_accounts?.length"
              theme="light"
              trigger="hover"
              placement="auto"
              :popover-delay="[200, 150]"
              :max-height="240"
              :arrow="true"
              ext-cls="related-account-popover"
            >
              <span class="related-count-link">{{ row.related_account_count ?? 0 }}</span>
              <template #content>
                <div class="related-account-list">
                  <div
                    v-for="account in row.related_accounts"
                    :key="account.account_id"
                    class="related-account-item"
                    @click="handleGoToAccount(account)"
                  >
                    <span class="account-id">{{ account.account_id }}</span>
                    <i class="hcm-icon bkhcm-icon-jump-fill account-link-icon" />
                  </div>
                </div>
              </template>
            </bk-popover>
            <span v-else class="related-count-zero">{{ row.related_account_count ?? 0 }}</span>
          </template>
          <!-- 其他自定义渲染列 -->
          <template v-else-if="getColumnRender(column)">
            <component :is="() => getColumnRender(column)({ row })" />
          </template>
          <!-- 默认渲染 -->
          <template v-else>
            <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
          </template>
        </template>
      </bk-table-column>
      <bk-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <bk-button theme="primary" text @click="handleApplyToAccount(row)">应用到二级账号</bk-button>
        </template>
      </bk-table-column>
    </bk-table>
  </bk-loading>
</template>

<style lang="scss" scoped>
.related-count-link {
  color: #3a84ff;
  cursor: pointer;
}

.related-count-zero {
  color: #63656e;
}
</style>

<style lang="scss">
.related-account-popover {
  padding: 8px !important;

  .related-account-list {
    max-height: 220px;
    overflow-y: auto;

    // padding: 4px 0;

    .related-account-item {
      display: flex;
      align-items: center;
      padding: 0 12px;
      height: 36px;
      line-height: 20px;
      cursor: pointer;
      transition: background-color 0.15s;

      &:hover {
        background-color: #f0f1f5;
      }

      .account-id {
        font-size: 12px;
        color: #4d4f56;
        margin-right: 8px;
      }

      .account-link-icon {
        font-size: 16px;
        color: #3a84ff;
        flex-shrink: 0;
        font-weight: 400 !important;
      }
    }
  }
}
</style>
