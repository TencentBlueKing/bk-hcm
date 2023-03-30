<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  h,
} from 'vue';
import {
  Button,
  InfoBox
} from 'bkui-vue';
import {
  useResourceStore,
} from '@/store/resource';
import useDelete from '../../hooks/use-delete';
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useSelection from '../../hooks/use-selection';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

// use hooks
const resourceStore = useResourceStore();

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'eips');

const columns = useColumns('eips');

const {
  selections,
  handleSelectionChange,
} = useSelection();

const {
  handleShowDelete,
  DeleteDialog,
} = useDelete(
  columns,
  selections,
  'eips',
  '删除 EIP',
  true,
);

const renderColumns = [
  ...columns,
  {
    label: '操作',
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          disabled: data.cvm_id || data.bk_biz_id !== -1,
          onClick() {
            InfoBox({
              title: '请确认是否删除',
              subTitle: `将删除【${data.id}】`,
              theme: 'danger',
              headerAlign: 'center',
              footerAlign: 'center',
              contentAlign: 'center',
              onConfirm() {
                resourceStore
                  .deleteBatch(
                    'eips',
                    {
                      ids: [data.id],
                    },
                  );
              },
            });
          },
        },
        [
          '删除',
        ],
      );
    },
  }
]
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section>
      <slot></slot>
      <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="selections.length <= 0"
        @click="handleShowDelete(selections.map(selection => selection.id))"
      >
        删除
      </bk-button>
    </section>

    <bk-table
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="renderColumns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="handleSelectionChange"
    />
  </bk-loading>
  <delete-dialog />
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
</style>
