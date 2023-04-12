<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  h,
  computed,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import {
  Button,
  InfoBox,
  Message,
} from 'bkui-vue';
import {
  useResourceStore,
} from '@/store/resource';
import useDelete from '../../hooks/use-delete';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  authVerifyData: {
    type: Object as PropType<any>,
  },
});

const {
  t,
} = useI18n();

const columns = useColumns('drive');
const simpleColumns = useColumns('drive', true);
const resourceStore = useResourceStore();

const selectSearchData = computed(() => {
  return [
    ...searchData.value,
    ...[{
      name: '云地域',
      id: 'region',
    }],
  ];
});

const {
  searchData,
  searchValue,
} = useFilter(props);

const emit = defineEmits(['auth']);

const renderColumns = [
  ...columns,
  {
    label: '操作',
    render({ data }: any) {
      return h(h(
        'span',
        {
          onClick() {
            emit('auth', props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate');
          },
        },
        [
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              disabled: !props.authVerifyData?.permissionAction[props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate']
                  || data.instance_id,
              onClick() {
                InfoBox({
                  title: '请确认是否删除',
                  subTitle: `将删除【${data.name}】`,
                  theme: 'danger',
                  headerAlign: 'center',
                  footerAlign: 'center',
                  contentAlign: 'center',
                  onConfirm() {
                    resourceStore
                      .recycled(
                        'disks',
                        {
                          infos: [{ id: data.id }],
                        },
                      ).then(() => {
                        Message({
                          theme: 'success',
                          message: '回收成功',
                        });
                      });
                  },
                });
              },
            },
            [
              t('回收'),
            ],
          )],
      ));
    },
  },
];

const {
  selections,
  handleSelectionChange,
} = useSelection();

const {
  // handleShowDelete,
  DeleteDialog,
} = useDelete(
  simpleColumns,
  selections,
  'disks',
  t('删除硬盘'),
  true,
);

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'disks');
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'">
      <slot>
      </slot>
      <!-- <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="selections.length <= 0"
        @click="handleShowDelete(selections.map(selection => selection.id))"
      >
        {{ t('删除') }}
      </bk-button> -->
      <bk-search-select
        class="w500 ml10"
        clearable
        :data="selectSearchData"
        v-model="searchValue"
      />
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
