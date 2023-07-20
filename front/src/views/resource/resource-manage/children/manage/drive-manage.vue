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
import { VendorEnum } from '@/common/constant';

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
  whereAmI: {
    type: String,
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
  filter,
} = useFilter(props);

const emit = defineEmits(['auth']);

const isDisabledRecycle = (vendor: VendorEnum, status: string) => {
  let res = true;
  switch (vendor) {
    case VendorEnum.TCLOUD: {
      if (['UNATTACHED'].includes(status)) res = false;
      break;
    }
    case VendorEnum.HUAWEI: {
      if (['available'].includes(status)) res = false;
      break;
    }
    case VendorEnum.AWS: {
      if (!['attaching', 'attached', 'detaching'].includes(status)) res = false;
      break;
    }
    case VendorEnum.GCP: {
      if (!['CREATING', 'DELETING', 'RESTORING'].includes(status)) res = false;
      break;
    }
    case VendorEnum.AZURE: {
      if (['Unattached'].includes(status)) res = false;
      break;
    }
  }
  return res;
};

const unableRecycled = (data: { instance_id: string; vendor: VendorEnum; status: string; }) => {
  return !props.authVerifyData?.permissionAction[props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate']
    || data.instance_id || isDisabledRecycle(data?.vendor, data?.status);
};

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
              disabled: unableRecycled(data),
              onClick() {
                InfoBox({
                  title: '请确认是否删除',
                  subTitle: `将删除【${data.name}】`,
                  // @ts-ignore
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
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
  triggerApi,
} = useQueryList({ filter: filter.value }, 'disks');

const {
  handleShowDelete,
  DeleteDialog,
} = useDelete(
  simpleColumns,
  selections,
  'disks',
  '回收硬盘',
  true,
  'recycle',
  triggerApi,
);

/**
 * 资源下，未绑定 且 未分配 可删除；
 * 业务下，未绑定 可删除；
 */
const isRowSelectEnable = ({ row }) => {
  if (!props.isResourcePage) return !unableRecycled(row);
  if (row.id) {
    return row.bk_biz_id === -1 && !unableRecycled(row);
  }
};

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
      <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="selections.length <= 0"
        @click="handleShowDelete(selections.filter(
          selection => !isDisabledRecycle(selection?.vendor, selection?.status)).map(selection => selection.id)
        )"
      >
        {{ t('回收') }}
      </bk-button>
      <div class="flex-row align-items-center justify-content-arround mlauto">
        <bk-search-select
          class="w500 ml10 mr15"
          clearable
          :conditions="[]"
          :data="selectSearchData"
          v-model="searchValue"
        />
        <slot name="recycleHistory"></slot>
      </div>

    </section>

    <bk-table
      class="mt20"
      row-hover="auto"
      remote-pagination
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      :pagination="pagination"
      :columns="renderColumns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="handleSelectionChange"
      row-key="id"
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
.mr15 {
  margin-right: 15px;
}
.mlauto {
  margin-left: auto;
}
</style>
