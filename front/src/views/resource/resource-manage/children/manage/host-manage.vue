<script setup lang="ts">
import type {
  // PlainObject,
  DoublePlainObject,
  FilterType,
} from '@/typings/resource';
import { Button, Dropdown, Message, Checkbox, bkTooltips } from 'bkui-vue';
import { PropType, h, ref, withDirectives } from 'vue';
import { useI18n } from 'vue-i18n';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import useFilterHost from '@/views/resource/resource-manage/hooks/use-filter-host';
import { useResourceStore } from '@/store';
import HostOperations, { HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from '../../common/table/HostOperations';
import BusinessSelector from '@/components/business-selector/index.vue';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import Confirm, { confirmInstance } from '@/components/confirm';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';

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

const isLoadingCloudAreas = ref(false);
const cloudAreaPage = ref(0);
const cloudAreas = ref([]);
const { whereAmI, isResourcePage, isBusinessPage } = useWhereAmI();

const { searchValue, filter } = useFilterHost(props);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'cvms',
);
// 主机列表分页支持500条
Object.assign(pagination.value, { 'limit-list': [10, 20, 50, 100, 500] });

const { selections, handleSelectionChange, resetSelections } = useSelection();

const currentOperateCvm = ref(null);
const { columns, generateColumnsSettings } = useColumns('cvms');
const isDialogShow = ref(false);
const isDialogBtnLoading = ref(false);
const selectedBizId = ref(0);
const resourceStore = useResourceStore();

const operationDropdownList = [
  { label: '开机', type: 'start' },
  { label: '关机', type: 'stop' },
  { label: '重启', type: 'reboot' },
  { label: '回收', type: 'recycle', hidden: isBusinessPage },
];
const currentOperateRowIndex = ref(-1);
// 操作的相关信息
const cvmInfo = ref({
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
                isResourcePage && handleSingleDistribution(data);
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

const getCloudAreas = () => {
  if (isLoadingCloudAreas.value) return;
  isLoadingCloudAreas.value = true;
  resourceStore
    .getCloudAreas({
      page: {
        start: cloudAreaPage.value,
        limit: 100,
      },
    })
    .then((res: any) => {
      cloudAreaPage.value += 1;
      cloudAreas.value.push(...(res?.data?.info || []));
    })
    .finally(() => {
      isLoadingCloudAreas.value = false;
    });
};

// 主机相关操作 - 分配业务
const handleSingleDistribution = (cvm: any) => {
  isDialogShow.value = true;
  currentOperateCvm.value = cvm;
};
const handleSingleDistributionConfirm = async () => {
  isDialogBtnLoading.value = true;
  try {
    await resourceStore.assignBusiness('cvms', {
      cvm_ids: [currentOperateCvm.value.id],
      bk_biz_id: selectedBizId.value,
    });
    Message({ message: t('操作成功'), theme: 'success' });
    triggerApi();
  } finally {
    isDialogShow.value = false;
    isDialogBtnLoading.value = false;
  }
};

getCloudAreas();
</script>

<template>
  <bk-loading :loading="isLoading" opacity="1">
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <BatchDistribution
        :selections="selections"
        :type="DResourceType.cvms"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <HostOperations
        :selections="selections"
        :on-finished="(type: 'confirm' | 'cancel' = 'confirm') => {
        if(type === 'confirm') triggerApi();
        resetSelections();
      }"
      ></HostOperations>

      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <resource-search-select v-model="searchValue" :resource-type="ResourceTypeEnum.CVM" value-behavior="need-key" />
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

    <bk-dialog
      :is-show="isDialogShow"
      title="主机分配"
      :theme="'primary'"
      quick-close
      @closed="() => (isDialogShow = false)"
      @confirm="handleSingleDistributionConfirm"
      :is-loading="isDialogBtnLoading"
    >
      <p class="selected-host-info">当前操作主机为：{{ currentOperateCvm.name }}</p>
      <p class="mb6">请选择所需分配的目标业务</p>
      <business-selector v-model="selectedBizId" :authed="true" class="mb32" :auto-select="true"></business-selector>
    </bk-dialog>
  </bk-loading>
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
.mb32 {
  margin-bottom: 32px;
}
.distribution-cls {
  display: flex;
  align-items: center;
}
.mr10 {
  margin-right: 10px;
}
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
.selected-host-info {
  margin-bottom: 16px;
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
