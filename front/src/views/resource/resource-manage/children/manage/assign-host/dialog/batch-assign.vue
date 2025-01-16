<script setup lang="ts">
import { computed, Fragment, h, ref, useTemplateRef, watch, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { groupBy } from 'lodash';

import { Button, Message, Tag } from 'bkui-vue';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';

import {
  HOST_RUNNING_STATUS,
  HOST_SHUTDOWN_STATUS,
} from '@/views/resource/resource-manage/common/table/HostOperations';
import { CLOUD_HOST_STATUS } from '@/common/constant';
import {
  type CvmsAssignPreviewItem,
  type CvmBatchAssignOpItem,
  type ICvmsAssignBizsPreviewItem,
  useHostStore,
} from '@/store';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { timeFormatter } from '@/common/util';

import MatchHost from './match-host.vue';
import ManualAssign from './manual-assign.vue';

const props = defineProps<{
  previewList: Array<CvmsAssignPreviewItem>;
  reloadTable: () => void;
}>();
const emit = defineEmits<(e: 'hidden') => void>();
const model = defineModel<boolean>();

const { t } = useI18n();
const hostStore = useHostStore();

const previewOpList = ref<CvmBatchAssignOpItem[]>([]);
const activeIndex = ref([]);

const findCvmIndex = (accountName: string, cloudVpcId: string, cvmId: string) => {
  const outerIdx = previewOpList.value.findIndex(
    ({ account_name, cloud_vpc_id }) => account_name === accountName && cloud_vpc_id === cloudVpcId,
  );
  const innerIdx = previewOpList.value[outerIdx].tableData.findIndex(({ id }) => id === cvmId);

  return [outerIdx, innerIdx];
};

const unConfirmFilterFn = (d: CvmsAssignPreviewItem) =>
  d.match_type !== 'auto' && (d.bk_biz_id === undefined || d.bk_cloud_id === undefined);

// 计算未确认的主机数量
const calculateUnConfirmedCount = (tableData: CvmsAssignPreviewItem[]) => tableData.filter(unConfirmFilterFn).length;

// 计算不同云区域的数量
const calculateBkCloudCount = (tableData: CvmsAssignPreviewItem[]) => {
  const bkCloudIdSet = tableData.reduce((prev, curr) => {
    if (curr.bk_cloud_id !== undefined && !prev.has(curr.bk_cloud_id)) {
      prev.add(curr.bk_cloud_id);
    }
    return prev;
  }, new Set());
  return bkCloudIdSet.size;
};
// 根据条件过滤表格数据
const filterTableData = (onlyShowUnConfirmed: boolean, tableData: CvmsAssignPreviewItem[]) =>
  onlyShowUnConfirmed ? tableData.filter(unConfirmFilterFn) : tableData;

// 计算属性，更新视图数据列表，包括未确认数量、主机数量、云区域数量，并根据条件过滤表格数据
const computedCvmsBatchAssignViewDataList = computed(() =>
  previewOpList.value.map<CvmBatchAssignOpItem>((item) => ({
    ...item,
    // 计算未确认的主机数量
    unConfirmedCount: calculateUnConfirmedCount(item.tableData),
    // 主机总数
    hostCount: item.tableData.length,
    // 不同云区域的数量
    bkCloudCount: calculateBkCloudCount(item.tableData),
    // 根据是否只显示未确认的主机来过滤表格数据
    tableData: filterTableData(item.onlyShowUnConfirmed, item.tableData),
  })),
);
const hasUnConfirmedHost = computed(() => previewOpList.value.some((item) => item.tableData.some(unConfirmFilterFn)));

watchEffect(() => {
  // 根据账户ID和云VPC ID分组数据（一个主机只能属于一个 vpc）
  const groupedData = groupBy(props.previewList, (item) => `${item.account_id}-${item.cloud_vpc_ids[0]}`);
  const list = Object.values(groupedData).map((group) => ({
    account_name: group[0].account_name,
    cloud_vpc_id: group[0].cloud_vpc_ids[0],
    onlyShowUnConfirmed: false,
    tableData: group,
  }));
  // 更新响应式数据列表
  previewOpList.value = list;
});

// 面板交互
const handleDeleteCollapsePanel = (index: number) => {
  previewOpList.value.splice(index, 1);
  // 如果删除的是最后一个面板，则关闭弹窗
  if (previewOpList.value.length === 0) {
    handleClosed();
  }
};
watch(
  previewOpList,
  (list) => {
    // 默认展开存有未确认主机的面板
    activeIndex.value = list.reduce((prev, curr, index) => {
      if (curr.tableData.some(unConfirmFilterFn)) prev.push(index);
      return prev;
    }, []);
  },
  { once: true },
);

// 表格配置
const columns = [
  { id: 'private_ip_address', name: t('内网IP'), type: 'string', width: 150 },
  { id: 'public_ip_address', name: t('公网IP'), type: 'string', width: 150 },
  { id: 'bk_cloud_id', name: t('管控区域'), type: 'cloud-area', width: 150 },
  { id: 'bk_biz_id', name: t('分配的目标业务'), type: 'business', width: 120 },
  {
    id: 'match_type',
    name: t('是否与配置平台关联'),
    type: 'string',
    width: 150,
    render: ({ cell, data }: { cell: ICvmsAssignBizsPreviewItem['match_type']; data: CvmsAssignPreviewItem }) => {
      const isIconShow = ['no_match', 'manual'].includes(cell);

      const tagMap: Record<
        ICvmsAssignBizsPreviewItem['match_type'],
        { text: string; theme: 'success' | 'danger' | 'warning'; clickHandler?: () => void }
      > = {
        no_match: {
          text: t('待关联'),
          theme: 'danger',
          clickHandler: () => {
            isManualAssignShow.value = true;
            currentCvm.value = { ...data };
          },
        },
        manual: {
          text: t('手动关联'),
          theme: 'warning',
          clickHandler: () => {
            isMatchHostShow.value = true;
            currentCvm.value = { ...data };
          },
        },
        auto: { text: t('自动关联'), theme: 'success' },
      };

      const { text, theme, clickHandler } = tagMap[cell];

      return h(Fragment, null, [
        h(Tag, { theme }, text),
        isIconShow
          ? h(
              Button,
              { theme: 'primary', text: true, class: 'ml8', onClick: clickHandler },
              h('i', { class: 'hcm-icon bkhcm-icon-configuration', style: 'font-size: 16px' }),
            )
          : null,
      ]);
    },
  },
  { id: 'region', name: t('地域'), type: 'region', width: 150 },
  { id: 'cloud_vpc_ids', name: t('所属vpc'), type: 'string', width: 150, render: ({ cell }: any) => cell?.join(',') },
  { id: 'name', name: t('主机名称'), type: 'string', width: 150 },
  {
    id: 'status',
    name: t('主机状态'),
    type: 'string',
    width: 120,
    render: ({ cell }: any) => {
      // eslint-disable-next-line no-nested-ternary
      const src = HOST_SHUTDOWN_STATUS.includes(cell)
        ? cell.toLowerCase() === 'stopped'
          ? StatusUnknown
          : StatusAbnormal
        : HOST_RUNNING_STATUS.includes(cell)
        ? StatusNormal
        : StatusUnknown;

      return h('div', { class: 'flex-row align-items-center' }, [
        h('img', { class: 'mr6', src, width: 14, height: 14 }),
        h('span', null, CLOUD_HOST_STATUS[cell] || cell || t('未获取')),
      ]);
    },
  },
  { id: 'machine_type', name: t('实例规格'), type: 'string', width: 120 },
  { id: 'os_name', name: t('操作系统'), type: 'string', width: 200 },
  { id: 'created_at', name: t('创建时间'), type: 'string', width: 180, render: ({ cell }: any) => timeFormatter(cell) },
];

// 关联配置平台主机
const isMatchHostShow = ref(false);
const matchHostDialogRef = useTemplateRef('match-host-dialog');
const currentCvm = ref<CvmsAssignPreviewItem>(null);
const handleBackfill = (cvm: CvmsAssignPreviewItem, bkBizId: number, bkCloudId: number) => {
  const [outerIdx, innerIdx] = findCvmIndex(cvm.account_name, cvm.cloud_vpc_ids[0], cvm.id);
  previewOpList.value[outerIdx].tableData[innerIdx].bk_biz_id = bkBizId;
  previewOpList.value[outerIdx].tableData[innerIdx].bk_cloud_id = bkCloudId;

  // 手动关联且手动分配，关闭弹框并清空form
  if (cvm.match_type === 'manual' && isMatchHostShow.value) {
    matchHostDialogRef.value.handleClosed();
  }
};
const handleManualAssign = () => {
  isManualAssignShow.value = true;
};

// 分配主机
const isManualAssignShow = ref(false);

const handleConfirm = async () => {
  const cvms = previewOpList.value.flatMap((item) =>
    item.tableData.map(({ id: cvm_id, bk_biz_id, bk_cloud_id }) => ({ cvm_id, bk_biz_id, bk_cloud_id })),
  );
  await hostStore.assignCvmsToBiz(cvms);
  Message({ theme: 'success', message: t('批量分配成功') });
  handleClosed();
  props.reloadTable();
};

const handleClosed = () => {
  model.value = false;
  emit('hidden');
};
</script>

<template>
  <bk-dialog
    :is-show="model"
    :title="t('批量分配主机')"
    header-align="center"
    :quick-close="false"
    :show-mask="false"
    fullscreen
    dialog-type="show"
    class="batch-assign-dialog"
    @closed="handleClosed"
  >
    <bk-loading v-if="hostStore.isAssignPreviewLoading" loading>
      <div style="width: 100%; height: 360px" />
    </bk-loading>
    <template v-else>
      <div class="content">
        <bk-collapse v-model="activeIndex" use-card-theme class="collapse-container">
          <bk-collapse-panel v-for="(item, index) in computedCvmsBatchAssignViewDataList" :key="index" :name="index">
            <!-- collapse-header -->
            <span class="collapse-header">
              <span class="info-wrap">
                <span>
                  <span class="info-value" style="font-size: 14px">{{ item.account_name }}</span>
                  <span class="text-desc">（vpc：{{ item.cloud_vpc_id }}）</span>
                </span>
                <span>
                  {{ t('当前有') }}
                  <span class="un-confirm-count">{{ item.unConfirmedCount }}</span>
                  {{ t('个主机未确认') }}
                </span>
                <bk-checkbox
                  style="font-size: 12px"
                  :model-value="item.onlyShowUnConfirmed"
                  @change="(v: boolean) => (previewOpList[index].onlyShowUnConfirmed = v)"
                >
                  {{ t('仅展示需处理项') }}
                </bk-checkbox>
                <span>
                  {{ t('主机数：') }}
                  <span class="info-value">{{ item.hostCount }}</span>
                </span>
                <span>
                  {{ t('管控区域数：') }}
                  <span class="info-value">{{ item.bkCloudCount }}</span>
                </span>
              </span>
              <bk-button class="delete-btn" text @click.stop="handleDeleteCollapsePanel(index)">
                <i class="hcm-icon bkhcm-icon-delete"></i>
              </bk-button>
            </span>
            <!-- collapse-content -->
            <template #content>
              <bk-table
                :data="item.tableData"
                :thead="{ color: '#F0F1F5' }"
                row-key="id"
                row-hover="auto"
                show-overflow-tooltip
                pagination
              >
                <bk-table-column
                  v-for="(column, idx) in columns"
                  :key="idx"
                  :prop="column.id"
                  :label="column.name"
                  :render="column.render"
                  :width="column.width"
                >
                  <template #default="{ row }">
                    <display-value
                      :property="column"
                      :value="row[column.id]"
                      :vendor="row?.vendor"
                      :resource-type="ResourceTypeEnum.CVM"
                    />
                  </template>
                </bk-table-column>
                <bk-table-column :label="t('操作')" fixed="right" width="100">
                  <template #default="{ row }">
                    <bk-button
                      class="button"
                      text
                      :disabled="item.tableData.length === 1"
                      @click="
                        () => {
                          const [outerIdx, innerIdx] = findCvmIndex(item.account_name, item.cloud_vpc_id, row.id);
                          previewOpList[outerIdx].tableData.splice(innerIdx, 1);
                        }
                      "
                    >
                      <i class="hcm-icon bkhcm-icon-minus-circle-shape"></i>
                    </bk-button>
                  </template>
                </bk-table-column>
              </bk-table>
            </template>
          </bk-collapse-panel>
        </bk-collapse>
      </div>

      <div class="footer">
        <bk-button
          class="button"
          theme="primary"
          :loading="hostStore.isAssignCvmsToBizsLoading"
          :disabled="hasUnConfirmedHost"
          v-bk-tooltips="{
            content: t('存在未确认主机，不可提交'),
            disabled: !hasUnConfirmedHost,
            placement: 'top-end',
          }"
          @click="handleConfirm"
        >
          {{ t('提交') }}
        </bk-button>
        <bk-button class="button" :disabled="hostStore.isAssignCvmsToBizsLoading" @click="handleClosed">
          {{ t('取消') }}
        </bk-button>
      </div>
    </template>
  </bk-dialog>

  <!-- 关联配置平台主机 -->
  <match-host
    v-model="isMatchHostShow"
    ref="match-host-dialog"
    action="backfill"
    :cvm="currentCvm"
    @backfill="handleBackfill"
    @manual-assign="handleManualAssign"
  />

  <!-- 分配主机 -->
  <manual-assign
    v-model="isManualAssignShow"
    action="backfill"
    :cvm="currentCvm"
    @backfill="(bkBizId, bkCloudId) => handleBackfill(currentCvm, bkBizId, bkCloudId)"
  />
</template>

<style scoped lang="scss">
.batch-assign-dialog {
  :deep(.bk-modal-wrapper) {
    top: 52px !important;

    .bk-dialog-header {
      padding: 0;
      height: 52px;
      line-height: 52px;
      font-size: 16px;
      color: #4d4f56;
      box-shadow: 0 1px 0 0 #dcdee5, 0 3px 4px 0 #4070cb0f;
    }

    .bk-modal-content {
      background: #f5f7fa;

      .bk-dialog-content {
        margin: 0;
        padding: 24px 40px 52px;

        .footer {
          display: flex;
          align-items: center;
          gap: 8px;
          height: 48px;

          .button {
            min-width: 88px;
          }
        }
      }
    }
  }
}

.collapse-container {
  background: #fff;

  :deep(.bk-collapse-content) {
    padding: 0;
    max-height: 300px;
  }

  .collapse-header {
    display: inline-flex;
    width: calc(100% - 26px);

    .info-wrap {
      display: inline-flex;
      align-items: center;
      gap: 24px;
      font-size: 12px;
      color: #4d4f56;

      .info-value {
        color: #313238;
      }

      .un-confirm-count {
        color: #e71818;
        font-weight: 700;
      }

      :deep(.bk-checkbox) {
        color: #4d4f56;
      }
    }

    .delete-btn {
      margin-left: auto;
      color: #ea3636;
    }
  }

  .bkhcm-icon-minus-circle-shape {
    color: #c4c6cc;
  }
}
</style>
