<script setup lang="ts">
import { computed, h } from 'vue';
import dayjs from 'dayjs';
import { Button, Tag } from 'bkui-vue';
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import type { IDataListProps } from '../../typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useSearchQs from '@/hooks/use-search-qs';
import { useRollingServerStore } from '@/store/rolling-server';
import routerAction from '@/router/utils/action';
import usageOrderViewProperties from '@/model/rolling-server/usage-order.view';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { timeFormatter } from '@/common/util';
import { RollingServerRecordItem } from '@/store';
import { MENU_BUSINESS_TICKET_MANAGEMENT } from '@/constants/menu-symbol';

withDefaults(defineProps<IDataListProps>(), {});

const emit = defineEmits<{
  'show-returned-records': [id: string];
  'show-not-notice': [row: RollingServerRecordItem];
  'show-order-config': [row: RollingServerRecordItem];
}>();

const rollingServerStore = useRollingServerStore();
const { handlePageChange, handlePageSizeChange, handleSort } = usePage();
const { isBusinessPage } = useWhereAmI();

// 判断单据是否可配置
const canConfigOrder = (row: RollingServerRecordItem) => {
  // 条件1：单据日期在121天内
  const rollDate = dayjs(String(row.roll_date), 'YYYYMMDD');
  const daysDiff = dayjs().diff(rollDate, 'day');
  const isWithin121Days = daysDiff < 121;

  // 条件2：执行率 < 100%（即有未退还的核心数）
  const execRateNum = parseFloat(row.exec_rate?.replace('%', '') || '0');
  const isNotFullReturned = execRateNum < 100;

  return isWithin121Days && isNotFullReturned;
};

const fieldIds = [
  'suborder_id',
  'bk_biz_id',
  'roll_date',
  'created_at',
  'applied_core',
  'delivered_core',
  'returned_core',
  'not_returned_core',
  'exec_rate',
  'creator',
  'not_notice',
];
const columnConfig: Record<string, PropertyColumnConfig> = {
  suborder_id: {
    width: 110,
    render: ({ cell, data }) => {
      const linkVNode = h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            const searchQs = useSearchQs();
            routerAction.open({
              name: MENU_BUSINESS_TICKET_MANAGEMENT,
              query: {
                [GLOBAL_BIZS_KEY]: data.bk_biz_id,
                type: 'host_apply',
                filter: searchQs.build({ order_id: [data.order_id], bk_username: [] }),
              },
            });
          },
        },
        cell,
      );

      if (rollingServerStore.resPollBusinessIds.includes(data.bk_biz_id)) {
        return h('div', { class: 'flex-row justify-content-between' }, [
          linkVNode,
          h(Tag, { theme: 'success' }, '资源池'),
        ]);
      }
      return linkVNode;
    },
  },
  bk_biz_id: { width: 150 },
  roll_date: { width: 110, sort: true, render: ({ cell }) => timeFormatter(String(cell), 'YYYY-MM-DD') },
  created_at: { width: 150, defaultHidden: true },
  applied_core: { width: 130, sort: true, align: 'right' },
  delivered_core: { width: 130, sort: true, align: 'right' },
  returned_core: { width: 110, align: 'right' },
  not_returned_core: {
    width: 130,
    align: 'right',
    render: ({ data }) => {
      const { not_returned_core, exempted_returned_core, not_returned_core_after_exempted } = data;
      // 如果有减免，显示格式：原值(减免后xxx)
      if (exempted_returned_core && exempted_returned_core > 0) {
        return h('span', {}, [not_returned_core, h('span', {}, `(减免后${not_returned_core_after_exempted})`)]);
      }
      // 无减免，只显示原值
      return not_returned_core;
    },
  },
  exec_rate: { width: 100, align: 'right' },
  creator: { width: 100, render: ({ cell }) => (cell === 'itsm_callback' ? '平台' : cell) },
  not_notice: { width: 100 },
};
const columns: ModelPropertyColumn[] = fieldIds.map((id) => ({
  ...usageOrderViewProperties.find((view) => view.id === id),
  ...columnConfig[id],
}));
const renderColumns = computed(() => {
  return isBusinessPage ? columns.filter((column) => column.id !== 'bk_biz_id') : columns;
});

const { settings } = useTableSettings(renderColumns.value);
</script>

<template>
  <div class="rolling-server-usage-data-list">
    <div class="table-tools">
      <ul class="summary">
        <li class="item">
          <div class="label">总交付（CPU核数）：</div>
          <div class="value">{{ summaryInfo?.sum_delivered_core ?? '--' }}</div>
        </li>
        <li class="item">
          <div class="label">总退还（CPU核数）：</div>
          <div class="value">{{ summaryInfo?.sum_returned_applied_core ?? '--' }}</div>
        </li>
      </ul>
    </div>
    <bk-table
      ref="tableRef"
      row-hover="auto"
      :data="list"
      :pagination="pagination"
      :max-height="'calc(100vh - 401px)'"
      :settings="settings"
      remote-pagination
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      row-key="id"
    >
      <bk-table-column
        v-for="(column, index) in renderColumns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :width="column.width"
        :sort="column.sort"
        :align="column.align"
        :render="column.render"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
      <bk-table-column :label="'操作'" fixed="right" width="220">
        <template #default="{ row }">
          <!-- 仅在管理员页面（非业务页面）显示，且满足约束条件 -->
          <bk-button
            v-if="!isBusinessPage && !row.isResPollBusiness && canConfigOrder(row)"
            class="mr8"
            theme="primary"
            text
            @click="emit('show-order-config', row)"
          >
            单据配置
          </bk-button>
          <bk-button v-if="!row.isResPollBusiness" theme="primary" text @click="emit('show-returned-records', row.id)">
            退还记录
          </bk-button>
          <template v-else>--</template>
          <bk-button class="ml8" theme="primary" text :disabled="row.not_notice" @click="emit('show-not-notice', row)">
            {{ row.not_notice ? '确认状态' : '屏蔽通知' }}
          </bk-button>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<style scoped lang="scss">
.rolling-server-usage-data-list {
  padding: 16px 24px;
  background-color: #fff;

  .table-tools {
    margin-bottom: 12px;
    display: flex;
    align-items: center;

    .summary {
      margin-left: auto;
      display: flex;
      align-items: center;
      gap: 40px;

      .item {
        display: inline-flex;

        .value {
          font-weight: 700;
          color: $warning-color;
        }
      }
    }
  }
}
</style>
