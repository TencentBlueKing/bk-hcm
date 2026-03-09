<script setup lang="ts">
import { watch } from 'vue';
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useFilterFromRoute from '@/views/resource-manage/hooks/use-filter-from-route';
import { ResourceTypeEnum } from '@/common/resource-constant';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';

const { columns, settings } = useColumns('image');

const { searchValue, filter, searchQs } = useFilterFromRoute(ResourceTypeEnum.IMAGE);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryList(
  { filter: filter.value },
  'images',
);

// 字段列表
const fieldList: string[] = columns.map((item) => item.field);
let dataList: any = datas;
// 接口缺失字段填充默认值
watch(datas, (list) => {
  dataList = list.map((item) => {
    fieldList.forEach((field) => {
      if (!Object.hasOwnProperty.call(item, field)) {
        item[field] = '--';
      }
    });
    return item;
  });
});
</script>

<template>
  <bk-loading :loading="isLoading">
    <resource-search-select
      class="search"
      v-model="searchValue"
      :resource-type="ResourceTypeEnum.IMAGE"
      @change="(condition) => searchQs.set(condition)"
    />
    <bk-table
      :settings="settings"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      :data="dataList"
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />
  </bk-loading>
</template>

<style lang="scss" scoped>
.search {
  margin-left: auto;
}
</style>
