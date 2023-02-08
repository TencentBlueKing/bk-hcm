<script lang="ts" setup>
import {
  h,
  watch,
  ref,
} from 'vue';

import {
  useRouter,
} from 'vue-router';

import {
  Button,
} from 'bkui-vue';

import useQueryList from '../../../hooks/use-query-list';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const router = useRouter();

const columns = ref([
  {
    label: 'ID',
    field: 'id',
  },
  {
    label: '资源 ID',
    field: 'cloud_id',
    render({ cell }: { cell: string }) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            router.push({
              name: 'resourceDetail',
              params: {
                type: 'subnet',
              },
              query: {
                id: cell,
              },
            });
          },
        },
        [
          cell,
        ],
      );
    },
  },
  {
    label: '名称',
    field: 'name',
  },
]);

watch(
  () => props.detail,
  () => {
    switch (props.detail.vendor) {
      case 'tcloud':
        columns.value.push(...[
          {
            label: '可用区',
            field: 'id',
          },
          {
            label: 'ID',
            field: 'id',
          },
          {
            label: 'ID',
            field: 'id',
          },
          {
            label: 'ID',
            field: 'id',
          },
          {
            label: 'ID',
            field: 'id',
          },
          {
            label: 'ID',
            field: 'id',
          },
        ]);
        break;
      case 'aws':

        break;
      case 'azure':
      case 'huawei':
        break;
    }
  },
  {
    deep: true,
    immediate: true,
  },
);

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
} = useQueryList(
  {
    filter: {
      op: 'and',
      rules: [],
    },
  },
  'subnets',
);
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <bk-table
      class="mt20"
      row-hover="auto"
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
    />
  </bk-loading>
</template>
