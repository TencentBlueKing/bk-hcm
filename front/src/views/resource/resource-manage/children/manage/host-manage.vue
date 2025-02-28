<script setup lang="ts">
import type {
  // PlainObject,
  DoublePlainObject,
  FilterType,
} from '@/typings/resource';
import { Button, Dropdown, Message, Checkbox, bkTooltips } from 'bkui-vue';
import { PropType, h, reactive, ref, withDirectives } from 'vue';
import { useI18n } from 'vue-i18n';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import useFilterHost from '@/views/resource/resource-manage/hooks/use-filter-host';
import { useHostStore, useResourceStore } from '@/store';
import HostOperations, { HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from '../../common/table/HostOperations';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import Confirm, { confirmInstance } from '@/components/confirm';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';
// 主机分配
import BatchAssign from './assign-host/dialog/batch-assign.vue';
import SingleAssign from './assign-host/dialog/single-assign.vue';
import type { ICvmItem } from '@/store';

const { DropdownMenu, DropdownItem } = Dropdown;

const { t } = useI18n();

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  whereAmI: {
    type: String,
  },
});

const { whereAmI, isResourcePage, isBusinessPage } = useWhereAmI();

const { searchValue, filter } = useFilterHost(props);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'cvms',
);
// 主机列表分页支持500条
Object.assign(pagination.value, { 'limit-list': [10, 20, 50, 100, 500] });

const { selections, handleSelectionChange, resetSelections } = useSelection();

const { columns, generateColumnsSettings } = useColumns('cvms');
const resourceStore = useResourceStore();

const operationDropdownList = [
  { label: '开机', type: 'start' },
  { label: '关机', type: 'stop' },
  { label: '重启', type: 'reboot' },
  { label: '回收', type: 'recycle', hidden: isBusinessPage },
];
const currentOperateRowIndex = ref(-1);
// 操作的相关信息
const cvmInfo = ref<Record<string, any>>({
  start: { op: '开机', loading: false, status: HOST_RUNNING_STATUS },
  stop: {
    op: '关机',
    loading: false,
    status: HOST_SHUTDOWN_STATUS,
  },
  reboot: { op: '重启', loading: false, status: HOST_SHUTDOWN_STATUS },
  recycle: { op: '回收', loading: false, status: HOST_SHUTDOWN_STATUS },
});
const getBkToolTipsOption = (data: any) => {
  if (isResourcePage) {
    return {
      content: '该主机仅可在业务下操作',
      disabled: !(isResourcePage && data.bk_biz_id !== -1),
    };
  }
  if (isBusinessPage) {
    return {
      content: `当前主机处于 ${CLOUD_HOST_STATUS[data.status]} 状态`,
      disabled: !(isBusinessPage && cvmInfo.value.stop.status.includes(data.status)),
    };
  }
  return {
    disabled: true,
  };
};
const tableColumns = [
  ...columns,
  {
    label: '操作',
    width: 120,
    showOverflowTooltip: false,
    render: ({ data, index }: { data: any; index: number }) => {
      return h('div', { class: 'operation-column' }, [
        withDirectives(
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              class: 'mr10',
              onClick: () => {
                // isResourcePage && 主机分配
                isResourcePage && showSingleAssignHost(data);
                // isBusinessPage && 主机回收
                isBusinessPage && handleCvmOperate('回收', 'recycle', data);
              },
              // TODO: 权限
              disabled:
                (isResourcePage && data.bk_biz_id !== -1) ||
                (isBusinessPage && cvmInfo.value.stop.status.includes(data.status)),
            },
            isResourcePage ? '分配' : '回收',
          ),
          [[bkTooltips, getBkToolTipsOption(data)]],
        ),
        withDirectives(
          h(
            Dropdown,
            {
              trigger: 'click',
              popoverOptions: {
                renderType: 'shown',
                onAfterShow: () => (currentOperateRowIndex.value = index),
                onAfterHidden: () => (currentOperateRowIndex.value = -1),
              },
              // TODO: 权限
              disabled: isResourcePage && data.bk_biz_id !== -1,
            },
            {
              default: () =>
                h(
                  'div',
                  {
                    class: [
                      `more-action${currentOperateRowIndex.value === index ? ' current-operate-row' : ''}`,
                      isResourcePage && data.bk_biz_id !== -1 ? 'disabled' : '',
                    ],
                  },
                  h('i', { class: 'hcm-icon bkhcm-icon-more-fill' }),
                ),
              content: () =>
                h(
                  DropdownMenu,
                  null,
                  operationDropdownList
                    .filter((action) => !action.hidden)
                    .map(({ label, type }) => {
                      return withDirectives(
                        h(
                          DropdownItem,
                          {
                            key: type,
                            onClick: () => handleCvmOperate(label, type, data),
                            extCls: `more-action-item${
                              cvmInfo.value[type].status.includes(data.status) ? ' disabled' : ''
                            }`,
                          },
                          label,
                        ),
                        [
                          [
                            bkTooltips,
                            {
                              content: `当前主机处于 ${CLOUD_HOST_STATUS[data.status]} 状态`,
                              disabled: !cvmInfo.value[type].status.includes(data.status),
                            },
                          ],
                        ],
                      );
                    }),
                ),
            },
          ),
          [[bkTooltips, { content: '该主机仅可在业务下操作', disabled: !(isResourcePage && data.bk_biz_id !== -1) }]],
        ),
      ]);
    },
  },
];

const tableSettings = generateColumnsSettings(tableColumns);

// 回收参数「云硬盘/EIP 随主机回收」
const isRecycleDiskWithCvm = ref(false);
const isRecycleEipWithCvm = ref(false);
// 重置回收参数
const resetRecycleSingleCvmParams = () => {
  isRecycleDiskWithCvm.value = false;
  isRecycleEipWithCvm.value = false;
};
// 主机相关操作 - 单个操作
const handleCvmOperate = async (label: string, type: string, data: any) => {
  // 判断当前主机是否可以执行对应操作
  if (cvmInfo.value[type].status.includes(data.status)) return;
  resetRecycleSingleCvmParams();
  let infoboxContent;
  if (type === 'recycle') {
    // 请求 cvm 所关联的资源(硬盘, eip)个数
    const {
      data: [target],
    } = await resourceStore.getRelResByCvmIds({ ids: [data.id] });
    const { disk_count, eip_count, eip } = target;
    infoboxContent = h('div', { style: { textAlign: 'justify' } }, [
      h('div', { style: { marginBottom: '10px' } }, [
        `当前操作主机为：${data.name}`,
        h('br'),
        `共关联 ${disk_count - 1} 个数据盘，${eip_count} 个弹性 IP${eip ? '('.concat(eip.join(','), ')') : ''}`,
      ]),
      h('div', null, [
        h(
          Checkbox,
          {
            checked: isRecycleDiskWithCvm.value,
            onChange: (checked: boolean) => (isRecycleDiskWithCvm.value = checked),
          },
          '云硬盘随主机回收',
        ),
        h(
          Checkbox,
          {
            checked: isRecycleEipWithCvm.value,
            onChange: (checked: boolean) => (isRecycleEipWithCvm.value = checked),
          },
          '弹性 IP 随主机回收',
        ),
      ]),
    ]);
  } else {
    infoboxContent = `当前操作主机为：${data.name}`;
  }
  Confirm(`确定${label}`, infoboxContent, async () => {
    confirmInstance.hide();
    isLoading.value = true;
    try {
      if (type === 'recycle') {
        await resourceStore.recycledCvmsData({
          infos: [{ id: data.id, with_disk: isRecycleDiskWithCvm.value, with_eip: isRecycleEipWithCvm.value }],
        });
      } else {
        await resourceStore.cvmOperate(type, { ids: [data.id] });
      }
      Message({ message: t('操作成功'), theme: 'success' });
      triggerApi();
    } finally {
      isLoading.value = false;
    }
  });
};

const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (whereAmI.value === Senarios.business) return true;
  if (row.id) {
    return row.bk_biz_id === -1;
  }
};

// 主机分配（批量）
const hostStore = useHostStore();
const batchAssignHostOptions = reactive({
  isShow: false,
  isHidden: true,
  previewList: [],
});
const showBatchAssignHost = async (cvms: ICvmItem[]) => {
  try {
    batchAssignHostOptions.isShow = true;
    batchAssignHostOptions.isHidden = false;
    // 获取预览数据
    batchAssignHostOptions.previewList = await hostStore.getAssignPreviewList(cvms);
  } catch (error) {
    console.error(error);
    batchAssignHostOptions.isShow = false;
    batchAssignHostOptions.isHidden = true;
  }
};

// 主机分配（单个）
const singleAssignHostOptions = reactive({
  isHidden: true,
  cvm: null,
});
const showSingleAssignHost = (cvm: ICvmItem) => {
  singleAssignHostOptions.isHidden = false;
  singleAssignHostOptions.cvm = { ...cvm };
};
</script>

<template>
  <bk-loading :loading="isLoading" opacity="1">
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <bk-button class="ml8 mr8" :disabled="!selections.length" @click="showBatchAssignHost(selections)">
        {{ t('批量分配') }}
      </bk-button>
      <HostOperations
        :selections="selections"
        :on-finished="(type: 'confirm' | 'cancel' = 'confirm') => {
        if(type === 'confirm') triggerApi();
        resetSelections();
      }"
      ></HostOperations>

      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <resource-search-select v-model="searchValue" :resource-type="ResourceTypeEnum.CVM" />
        <slot name="recycleHistory"></slot>
      </div>
    </section>

    <bk-table
      row-hover="auto"
      :columns="tableColumns"
      :data="datas"
      :settings="tableSettings"
      :pagination="pagination"
      remote-pagination
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
      @column-sort="handleSort"
      row-key="id"
    />
  </bk-loading>

  <!-- 批量分配主机 -->
  <template v-if="!batchAssignHostOptions.isHidden">
    <batch-assign
      v-model="batchAssignHostOptions.isShow"
      :preview-list="batchAssignHostOptions.previewList"
      :reload-table="triggerApi"
      @hidden="batchAssignHostOptions.isHidden = true"
    />
  </template>

  <!-- 单个分配主机 -->
  <template v-if="!singleAssignHostOptions.isHidden">
    <single-assign
      :cvm="singleAssignHostOptions.cvm"
      :reload-table="triggerApi"
      @hidden="singleAssignHostOptions.isHidden = true"
    />
  </template>
</template>

<style lang="scss" scoped>
.search-selector-container {
  margin-left: auto;
}
:deep(.operation-column) {
  height: 100%;
  display: flex;
  align-items: center;

  .more-action {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    cursor: pointer;

    & > i {
      position: absolute;
    }

    &:hover {
      background-color: #f0f1f5;
    }

    &.current-operate-row {
      background-color: #f0f1f5;
    }

    &.disabled {
      background-color: #fff;
      color: #dcdee5;
      cursor: not-allowed;
    }
  }
}
</style>

<style lang="scss">
.more-action-item {
  &.disabled {
    color: #dcdee5;
    cursor: not-allowed;
  }
}
</style>
