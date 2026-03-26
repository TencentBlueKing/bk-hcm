<script setup lang="ts">
import { h, ref, watchEffect } from 'vue';
import { Button } from 'bkui-vue';
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import {
  useRollingServerBillsStore,
  type IRollingServerBillItem,
  type IFineDetailsItem,
} from '@/store/rolling-server-bills';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import routerAction from '@/router/utils/action';
import billsDetailsViewProperties from '@/model/rolling-server/bills-details.view';
import { transformSimpleCondition } from '@/utils/search';
import { getColumnName } from '@/model/utils';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import BusinessValue from '@/components/display-value/business-value.vue';
import { MENU_BUSINESS_TICKET_MANAGEMENT } from '@/constants/menu-symbol';

const model = defineModel<boolean>();
const props = defineProps<{ dataRow: IRollingServerBillItem }>();
const rollingServerBillsStore = useRollingServerBillsStore();
const { pagination, pageParams, handlePageChange, handlePageSizeChange, handleSort } = usePage(false);

const detailsList = ref([]);

const columnConfig: Record<string, PropertyColumnConfig> = {
  suborder_id: {
    render: ({ row }: { row?: IFineDetailsItem }) =>
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            const searchQs = useSearchQs();
            routerAction.open({
              name: MENU_BUSINESS_TICKET_MANAGEMENT,
              query: {
                [GLOBAL_BIZS_KEY]: row.bk_biz_id,
                type: 'host_apply',
                filter: searchQs.build({ order_id: [row.order_id], bk_username: [] }),
              },
            });
          },
        },
        row.suborder_id,
      ),
  },
  delivered_core: { align: 'right' },
  returned_core: { align: 'right' },
  not_returned_core: {
    align: 'right',
    render: ({ row }: { row?: IFineDetailsItem }) => {
      const not_returned_core = row.delivered_core - row.returned_core;
      const exempted = row.exempted_returned_core || 0;
      if (exempted > 0) {
        return h('span', {}, [not_returned_core, `(减免后${not_returned_core - exempted})`]);
      }
      return not_returned_core;
    },
  },
};

const columns: ModelPropertyColumn[] = [];
for (const [fieldId, config] of Object.entries(columnConfig)) {
  columns.push({
    ...billsDetailsViewProperties.find((prop) => prop.id === fieldId),
    ...config,
  });
}

watchEffect(async () => {
  const condition = {
    bk_biz_id: props.dataRow.bk_biz_id,
    year: props.dataRow.year,
    month: props.dataRow.month,
    day: props.dataRow.day,
  };
  const { list, count } = await rollingServerBillsStore.getBillFineDetailsList({
    filter: transformSimpleCondition(condition, billsDetailsViewProperties),
    page: pageParams.value,
  });

  detailsList.value = list;

  pagination.count = count;
});
</script>

<template>
  <bk-dialog dialog-type="show" title="关联单据详情" width="960" v-model:is-show="model">
    <grid-container :column="2" :fixed="true" :gap="36" label-align="left" label-width="auto" content-min-width="auto">
      <grid-item label="业务"><business-value :value="dataRow.bk_biz_id" /></grid-item>
      <grid-item label="核算日期">{{ `${dataRow.year}-${dataRow.month}-${dataRow.day}` }}</grid-item>
    </grid-container>
    <bk-table
      v-bkloading="{ loading: rollingServerBillsStore.billFineDetailsListLoading, size: 'small' }"
      row-hover="auto"
      remote-pagination
      show-overflow-tooltip
      :max-height="500"
      :min-height="240"
      :data="detailsList"
      :pagination="pagination"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    >
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="getColumnName(column)"
        :sort="column.sort"
        :align="column.align"
        :render="column.render"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
    </bk-table>
  </bk-dialog>
</template>
