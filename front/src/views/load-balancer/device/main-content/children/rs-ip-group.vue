<script setup lang="ts">
import { computed, inject, ref, watch, ComputedRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { PrimaryTable, TableColumn, type TableProps, type SelectOptions } from '@blueking/tdesign-ui';
import { ModelPropertyColumn } from '@/model/typings';
import { useRegionsStore } from '@/store/useRegionsStore';
import { VendorEnum, GLOBAL_BIZS_KEY } from '@/common/constant';
import { RsInstType, RsDeviceType } from '@/views/load-balancer/constants';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TARGET_GROUP_DETAILS } from '@/constants/menu-symbol';

const props = defineProps<{ rsList: any[]; vendor: VendorEnum; type: RsDeviceType }>();

const emit = defineEmits(['delete']);

const currentGlobalBusinessId = inject<ComputedRef<number>>('currentGlobalBusinessId');

const { t } = useI18n();
const regionStore = useRegionsStore();

const dataListColumns: ModelPropertyColumn[] = [
  {
    id: 'port',
    name: 'RS端口',
    type: 'number',
    width: 80,
    fixed: 'left',
    cell: 'port',
  },
  {
    id: 'weight',
    name: props.type === RsDeviceType.ADJUST ? 'RS原权重' : 'RS权重',
    type: 'string',
    width: 80,
    fixed: 'left',
  },
  {
    id: 'inst_type',
    name: 'RS类型',
    type: 'enum',
    width: 100,
    option: RsInstType,
    fixed: 'left',
  },
  {
    id: 'target_group_name',
    name: '所属目标组',
    type: 'string',
    width: 150,
    ellipsis: (h, { row }) => row.target_group_name,
  },
  {
    id: 'rule_url',
    name: '所属URL',
    type: 'string',
    width: 120,
  },
  {
    id: 'lb_domain',
    name: '所属监听器域名',
    type: 'string',
    width: 120,
  },
  {
    id: 'lbl_port',
    name: '所属监听器端口',
    type: 'number',
    width: 120,
  },
  {
    id: 'lbl_name',
    name: '所属监听器名称',
    type: 'string',
    width: 120,
    ellipsis: true,
  },
  {
    id: 'lb_vips',
    name: '所属负载均衡VIP',
    type: 'array',
    width: 120,
  },
  {
    id: 'lb_region',
    name: '所属地域',
    type: 'region',
    width: 120,
  },
];

const visitedIpSet = new Set();

const activeGroupKeys = ref<string[]>([]);

const MAX_COUNT = 5000;

const RS_ROW_KEY = 'id';

// 选中RS，key为rsList的rowKey，value为该IP下选中的RS ID数组
const selectedRsMap = ref<Map<string, string[]>>(new Map());

const selectedCount = computed(() => {
  let count = 0;
  for (const arr of selectedRsMap.value.values()) {
    count += arr.length;
  }
  return count;
});

// 所有选中的RS，外部依赖与原rsList结构保持一致
const selections = computed(() => {
  const result: any[] = [];
  for (const [key, value] of selectedRsMap.value) {
    if (!value.length) {
      continue;
    }
    const item = props.rsList.find((item) => item.rowKey === key);
    result.push({
      ...item,
      targets: item.targets.filter((rs: any) => value.includes(rs[RS_ROW_KEY])),
    });
  }
  return result;
});

const hasSelection = computed(() => props.type === RsDeviceType.INFO);

const isExceeded = computed(() => selectedCount.value > MAX_COUNT);

watch(
  () => props.rsList,
  (list) => {
    // 重置选中数据
    selectedRsMap.value.clear();
    list.forEach((item) => {
      selectedRsMap.value.set(item.rowKey, []);
    });

    // 默认展开第一个
    activeGroupKeys.value = [list[0]?.rowKey];
  },
);

const isExpand = (rowKey: string) => activeGroupKeys.value.includes(rowKey);

// 获取IP下所有RS ID
const getRowTargetIds = (rowKey: string) => {
  const targets = props.rsList.find((item) => item.rowKey === rowKey)?.targets || [];
  return targets.map((rs: any) => rs[RS_ROW_KEY]);
};

const getDisplayCompProps = (column: ModelPropertyColumn) => {
  const { type } = column;
  if (type === 'region') {
    return { vendor: props.vendor };
  }
  return {};
};

const getVpc = (ids: string[]) => {
  if (!ids.length) return '--';
  return ids.join(',');
};

// 表格的选中状态变化，value每次为最新选中的RS ID数组
const handleTableSelectChange = (value: TableProps['selectedRowKeys'], ctx: SelectOptions<any>, rowKey: string) => {
  selectedRsMap.value.set(rowKey, value as string[]);
};

const handleIPClick = (instId: string, rowKey: string) => {
  visitedIpSet.add(rowKey);
  routerAction.open({
    name: 'hostBusinessDetail',
    query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value, id: instId, type: props.vendor },
  });
};

const handleSelectAll = () => {
  for (const key of selectedRsMap.value.keys()) {
    selectedRsMap.value.set(key, getRowTargetIds(key));
  }
};

const handleClearSelection = () => {
  for (const key of selectedRsMap.value.keys()) {
    selectedRsMap.value.set(key, []);
  }
};

const handleSingleDelete = (rowKey: string) => {
  emit('delete', rowKey);
};

const handleGroupCheckedChange = (checked: boolean, rowKey: string) => {
  selectedRsMap.value.set(rowKey, checked ? getRowTargetIds(rowKey) : []);
};

const handleViewTargetGroupDetails = (row: any) => {
  routerAction.open({
    name: MENU_BUSINESS_TARGET_GROUP_DETAILS,
    params: {
      id: row.target_group_id,
    },
    query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value, type: 'list', vendor: props.vendor },
  });
};

defineExpose({ selections, isExceeded });
</script>

<template>
  <div>
    <template v-if="rsList.length">
      <div class="rs-tips" v-if="selectedCount > 0 && !isExceeded">
        已选
        <span class="count">{{ selectedCount }}</span>
        个RS，可
        <bk-button text theme="primary" @click="handleClearSelection" class="mr14">{{ t('一键清空') }}</bk-button>
        <bk-button text theme="primary" @click="handleSelectAll">{{ t('全选所有IP') }}</bk-button>
      </div>
      <div v-if="isExceeded" class="rs-warning">
        <info-line class="warning" />
        <span class="mr10 ml3">
          {{ t(`当前操作的RS数量超过${MAX_COUNT}个，会导致批量变更时间较长，请减少数量后再操作`) }}
        </span>
        <bk-button text theme="primary" @click="handleClearSelection">{{ t('一键清空') }}</bk-button>
      </div>
      <bk-collapse use-block-theme class="rs-expand" v-model="activeGroupKeys" accordion>
        <bk-collapse-panel v-for="item in rsList" :key="item.rowKey" :name="item.rowKey">
          <template #header>
            <div class="header" :class="{ 'is-expand': isExpand(item.rowKey) }">
              <bk-checkbox
                v-if="hasSelection"
                :checked="selectedRsMap.get(item.rowKey)?.length === item.targets.length && item.targets.length > 0"
                :indeterminate="
                  selectedRsMap.get(item.rowKey)?.length > 0 &&
                  selectedRsMap.get(item.rowKey)?.length < item.targets.length
                "
                class="mr10 checked"
                @change="(checked: boolean) => handleGroupCheckedChange(checked, item.rowKey)"
              />
              <i class="hcm-icon bkhcm-icon-right-shape arrow-icon" />
              <div class="cvm-info row-fixed-cell" :title="`${item?.ip} ( ${item?.inst_name ?? '--'} ) `">
                <a
                  :class="[
                    'ip',
                    {
                      visited: visitedIpSet.has(item.rowKey),
                    },
                  ]"
                  @click="() => handleIPClick(item.inst_id, item.rowKey)"
                >
                  {{ item?.ip ?? '--' }}
                </a>
                <span class="name">（{{ t(item?.inst_name ?? '--') }}）</span>
              </div>
              <div class="rs-num row-fixed-cell" :title="item?.targets.length">
                {{ t('RS数量：') }} {{ item.targets.length }}
              </div>
              <div class="region row-fixed-cell" :title="regionStore.getZoneName(item.zone, vendor)">
                {{ t('可用区：') }} {{ regionStore.getZoneName(item.zone, vendor) }}
              </div>
              <div class="vpc row-fixed-cell" :title="getVpc(item.cloud_vpc_ids)">
                {{ t('所属vpc：') }} {{ getVpc(item.cloud_vpc_ids) }}
              </div>
              <bk-button
                text
                class="single-delete-btn"
                @click.stop="handleSingleDelete(item.rowKey)"
                v-if="type !== RsDeviceType.INFO"
              >
                <i class="hcm-icon bkhcm-icon-minus-circle-shape"></i>
              </bk-button>
            </div>
          </template>

          <template #content>
            <primary-table
              class="tx-table"
              max-height="420px"
              :row-key="RS_ROW_KEY"
              :scroll="{ type: 'virtual', rowHeight: 42, bufferSize: 10 }"
              :data="item.targets"
              :selected-row-keys="selectedRsMap.get(item.rowKey)"
              @select-change="(value, ctx) => handleTableSelectChange(value, ctx, item.rowKey)"
            >
              <table-column v-if="hasSelection" width="40" col-key="row-select" type="multiple"></table-column>
              <template v-for="column in dataListColumns" :key="column.id">
                <table-column
                  :col-key="column.id"
                  :title="column.name"
                  :sort="column.sort"
                  :width="column.width"
                  :fixed="column.fixed"
                  :ellipsis="column.ellipsis"
                >
                  <template #default="{ row }">
                    <template v-if="column.id === 'target_group_name'">
                      <bk-button text theme="primary" @click="() => handleViewTargetGroupDetails(row)">
                        {{ row.target_group_name }}
                      </bk-button>
                    </template>
                    <template v-else-if="column.id === 'weight'">{{ row.weight }}</template>
                    <display-value
                      v-else
                      :property="column"
                      :value="row[column.id]"
                      :display="column?.meta?.display"
                      v-bind="getDisplayCompProps(column)"
                    />
                  </template>
                </table-column>
              </template>
            </primary-table>
          </template>
        </bk-collapse-panel>
      </bk-collapse>
    </template>
    <template v-else>
      <bk-exception description="没有数据" scene="part" type="empty" />
    </template>
  </div>
</template>

<style scoped lang="scss">
.rs-tips {
  background: #f0f1f5;
  border-radius: 2px;
  line-height: 32px;
  font-size: 12px;
  padding-left: 20px;
  color: #313238;
  margin-bottom: 12px;

  .count {
    font-weight: bold;
  }
}

.rs-warning {
  background: #f9d090;
  border: 1px solid #f9d090;
  border-radius: 2px;
  font-size: 12px;
  color: #4d4f56;
  line-height: 32px;
  padding-left: 9px;
  margin-bottom: 12px;

  .warning {
    color: #f59500;
  }
}

.rs-expand {
  .header {
    display: flex;
    align-items: center;
    font-size: 12px;
    color: #4d4f56;
    line-height: 40px;
    padding: 0 12px;
    cursor: pointer;

    .row-fixed-cell {
      margin-right: 20px;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }

    .arrow-icon {
      color: #979ba5;
      font-size: 14px;
      transition: transform 0.3s ease;
    }

    .cvm-info {
      display: flex;
      align-items: center;
      width: 250px;
      margin-left: 10px;

      .name {
        color: #979ba5;
        display: inline-block;
        max-width: 140px;
        text-overflow: ellipsis;
        white-space: nowrap;
        overflow: hidden;
        line-height: 100%;
      }
    }

    .rs-num {
      width: 110px;
    }

    .region {
      width: 150px;
    }

    &.is-expand {
      background: #f0f5ff !important;

      .arrow-icon {
        transform: rotate(90deg);
      }
    }

    .single-delete-btn {
      color: #c4c6cc;
      margin-left: auto;
    }
  }

  .ip {
    color: #3a84ff;
    cursor: pointer;
    display: inline-block;

    &.visited {
      color: #8334f4;
      text-decoration: underline;
    }
  }

  :deep(.t-table) {
    thead {
      [data-colkey='row-select'] {
        .t-checkbox {
          display: none;
        }
      }
    }

    .weight {
      background: rgb(253 244 232);
      margin: 0 -16px;
      padding: 0 16px;
    }
  }

  :deep(.bk-collapse-item-active) {
    + .bk-collapse-item {
      box-shadow: -1px -8px 20px 0 rgb(0 0 0 / 10%);

      // fix: shadow被固定列遮挡
      z-index: 31;
    }
  }

  :deep(.bk-collapse-item) {
    margin-bottom: 0 !important;

    .bk-collapse-content {
      padding-left: 0;
      padding-right: 0;
      padding-top: 0;
    }

    &:nth-child(even) {
      .header {
        background: #fafbfd;
      }
    }
  }
}
</style>
