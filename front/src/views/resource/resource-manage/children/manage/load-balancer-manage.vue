<template>
   <bk-loading :loading="isLoading">
  <section>
    <div
      class="flex-row operate-warp justify-content-between align-items-center mb20"
    >
      <div>
        <bk-button theme="primary">
          {{ t('购买') }}
        </bk-button>
        <bk-button style="margin-left: 10px">
          {{ t('分配') }}
        </bk-button>
        <bk-button style="margin-left: 10px">
          {{ t('批量删除') }}
        </bk-button>
      </div>

      <div
        class="flex-row input-warp justify-content-between align-items-center"
      >
        <bk-search-select
          class="w500 ml10 mr15"
          clearable
          :conditions="[]"
          :data="loadSearchData"
          v-model="searchValue"
        />
      </div>
    </div>
  </section>
 
      <bk-table
        class="mt20"
        row-hover="auto"
        :columns="distribColumns"
        :data="datas"
        :settings="tableSettings"
        :pagination="pagination"
        remote-pagination
        show-overflow-tooltip
        row-key="id"
      />

      <bk-dialog 
        width="820" :title="t('主机分配')" 
        theme="primary"
        quick-close
        @confirm="handleDistributionConfirm">
        <section class="distribution-cls">
          目标业务
          <bk-select 
          class="ml20" 
          filterable>
            <bk-option 
            
            />
          </bk-select>
        </section>
        <bk-table
          class="mt20"
          row-hover="auto"
          :columns="distribColumns"
          :data="selections"
          show-overflow-tooltip
        />
      </bk-dialog>

      <bk-dialog
        :is-show="isDialogShow"
        title="主机分配"
        :theme="'primary'"
        quick-close
      >
        <p class="selected-host-count-tip">
          已选择
          <span class="selected-host-count">{{ selections.length }}</span>
          台主机，可选择所需分配的目标业务
        </p>
        <p class="mb6">目标业务</p>
        <business-selector
          class="mb32"
        >
        </business-selector>
      </bk-dialog>
    </bk-loading>
</template>

<script setup lang="ts">
import bkUi from 'bkui-vue';
import type { FilterType } from '@/typings/resource';
import {
  h,
  PropType,
  reactive,
  watch,
  toRefs,
  defineComponent,
  onMounted,
  ref,
  computed,
} from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import { Button, InfoBox, Message, Tag } from 'bkui-vue';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});
const { t } = useI18n();
const { searchData, searchValue, filter } = useFilter(props);
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange } =
  useQueryList({ filter: filter.value }, 'cvms');
const isShowDistribution = ref(false);
const handleDistributionConfirm = () => {
  isShowDistribution.value = true;
};

const loadSearchData = computed(() => {
  return [
    {
      name: '名称',
      field: 'name',
    },
    {
      name: '负载均衡域名',
      field: 'type',
    },
    {
      name: '负载均衡VIP',
      field: 'vendor',
    },
    {
      name: '网络类型',
      field: 'site',
    },
    {
      name: '监听数量',
      field: 'managers',
    },
    {
      name: 'IP版本',
      field: 'creator',
    },
  ];
});
const tableColumns = [];
const tableSettings = ref({
  fields: [
    {
      label: '负载均衡域名称',
      field: 'name',
    },
    {
      label: '云厂商',
      field: 'vendor',
    },
    {
      label: '地域',
      field: 'region',
    },
    {
      label: '可用区域',
      field: 're',
    },
    {
      label: '负载均衡域名',
      field: 'domain',
    },
    {
      label: '负载均衡VIP',
      field: 'VIP',
    },
    {
      label: '网络类型',
      field: 'network',
    },
    {
      label: '监听器数量',
      field: 'count',
    },
    {
      label: '状态',
      field: 'state',
    },
    {
      label: '分配状态',
      field: 'fpstate',
    },
    {
      label: '所属网络',
      field: 'bk_biz_id2',
    },
    {
      label: 'IP版本',
      field: 'IP',
    },
  ],
  checked: ['name', 'domain', 'VIP', 'network', 'count', 'fpstate', 'IP'],
});

const distribColumns = [
  {
    type: 'selection',
    width: '100',
    onlyShowOnList: true,
  },
  {
    label: '负载均衡域名称',
    field: 'name',
  },
  {
    label: '云厂商',
    field: 'vendor',
  },
  {
    label: '地域',
    field: 'region',
  },
  {
    label: '可用区域',
    field: 're',
  },
  {
    label: '负载均衡域名',
    field: 'domain',
  },
  {
    label: '负载均衡VIP',
    field: 'VIP',
  },
  {
    label: '网络类型',
    field: 'network',
  },
  {
    label: '监听器数量',
    field: 'count',
  },
  {
    label: '状态',
    field: 'state',
  },
  {
      label: '分配状态',
      field: 'fpstate',
  },
  {
    label: '所属网络',
    field: 'bk_biz_id2',
  },
  {
    label: 'IP版本',
    field: 'IP',
  },
  {
    label: '操作',
    field: 'operate',
    render({ data }: any) {
      return h('span', {}, [
        h(
          Button,
          {
            class: 'ml10',
            text: true,
            theme: 'primary',
            onClick() {},
          },
          [t('编辑')],
        ),
        h(
          Button,
          {
            class: 'ml10',
            text: true,
            theme: 'primary',
            onClick() {},
          },
          [t('删除')],
        ),
      ]);
    },
  },
];
</script>

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
.mb32 {
  margin-bottom: 32px;
}
.distribution-cls {
  display: flex;
  align-items: center;
}
.mr15 {
  margin-right: 15px;
}
.search-selector-container {
  margin-left: auto;
}
.operations-container {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  // cursor: pointer;
  &:hover {
    background: #f0f1f5;
  }
}
</style>
