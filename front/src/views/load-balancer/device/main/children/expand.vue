<template>
  <div>
    <template v-if="rsList.length">
      <div class="rs-tips" v-if="count > 0 && !moreData">
        已选
        <span class="count">{{ count }}</span>
        个RS，可
        <bk-button text theme="primary" @click="handleClearSelection">{{ t('一键清空') }}</bk-button>
        <bk-button text theme="primary" @click="handleSelectAll">{{ t('全选所有IP') }}</bk-button>
      </div>
      <div v-if="moreData" class="rs-warning">
        <info-line class="warning" />
        <span class="mr10 ml3">
          {{ t(`当前操作的RS数量超过${max}个，会导致批量变更时间较长，请减少数量后再操作`) }}
        </span>
        <bk-button text theme="primary" @click="handleClearSelection">{{ t('一键清空') }}</bk-button>
      </div>
      <bk-collapse use-block-theme class="rs-expand" v-model="active">
        <bk-collapse-panel class="rule-panel" v-for="(item, index) in rsList" :key="item.inst_id" :name="item.inst_id">
          <template #header>
            <div class="header" :class="{ 'is-selected': isExpand(item.inst_id) }">
              <bk-checkbox
                v-if="hasSelection"
                :checked="checkStatus[index]?.checked"
                :indeterminate="checkStatus[index]?.indeterminate"
                class="mr10 checked"
                @change="(val: boolean, event: any) => handleHeadChange(val, item.inst_id, event)"
              />
              <div>
                <AngleUpFill v-if="isExpand(item.inst_id)"></AngleUpFill>
                <RightShape v-else></RightShape>
              </div>
              <div class="info mr20 ml10">
                <a
                  :class="[
                    'ip',
                    {
                      actived: allActiveIP.has(item.inst_id),
                    },
                  ]"
                  @click="() => handleIPClick(item.inst_id)"
                >
                  {{ item.ip }}
                </a>
                (
                <span class="name">{{ t(item?.inst_name ?? '') }}</span>
                )
              </div>
              <div class="rs-num mr20">{{ t('RS数量：') }} {{ item.targets.length }}</div>
              <div class="region mr20">{{ t('可用区：') }} {{ regionStore.getZoneName(item.zone, vendor) }}</div>
              <div class="vpc">{{ t('所属vpc：') }} {{ item.cloud_vpc_ids.join(',') }}</div>
              <bk-button
                text
                class="single-delete-btn"
                @click.stop="handleSingleDelete(item.inst_id)"
                v-if="type !== RsDeviceType.INFO"
              >
                <i class="hcm-icon bkhcm-icon-minus-circle-shape"></i>
              </bk-button>
            </div>
          </template>

          <template #content>
            <bk-table
              row-key="id"
              selection-key="id"
              row-hover="auto"
              :data="item.targets"
              show-overflow-tooltip
              ref="tableRef"
              :key="item.id"
              @selection-change="({ checked }) => handleSelectionChange(item.inst_id, checked)"
            >
              <bk-table-column v-if="hasSelection" :width="40" :min-width="40" type="selection" fixed="left" />
              <bk-table-column
                v-for="(column, columnIndex) in dataListColumns"
                :key="columnIndex"
                :prop="column.id"
                :label="column.name"
                :sort="column.sort"
                :width="column.width"
                :fixed="column.fixed"
                :render="column.render"
                :filter="column.filter"
              >
                <template #default="{ row }">
                  <display-value
                    :property="column"
                    :value="row[column.id]"
                    :display="column?.meta?.display"
                    v-bind="getDisplayCompProps(column, row)"
                  />
                </template>
              </bk-table-column>
            </bk-table>
          </template>
        </bk-collapse-panel>
      </bk-collapse>
    </template>
    <template v-else>
      <bk-exception description="没有数据" scene="part" type="empty" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, ref, h, watch, ComputedRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { Button } from 'bkui-vue';
import { DisplayFieldType, DisplayFieldFactory } from '@/views/load-balancer/children/display/field-factory';
import { ModelPropertyColumn } from '@/model/typings';
import { AngleUpFill, RightShape } from 'bkui-vue/lib/icon';
import { useRegionsStore } from '@/store/useRegionsStore';
import { VendorEnum, GLOBAL_BIZS_KEY } from '@/common/constant';
import { RsDeviceType } from '@/views/load-balancer/constants';
import { cloneDeep } from 'lodash';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TARGET_GROUP_DETAILS } from '@/constants/menu-symbol';

const props = defineProps<{ rsList: any[]; vendor: VendorEnum; type: RsDeviceType }>();

const emit = defineEmits(['delete']);

const currentGlobalBusinessId = inject<ComputedRef<number>>('currentGlobalBusinessId');

const displayFieldIds = [
  'port',
  'weight',
  'inst_type',
  'target_group_name',
  'rule_url',
  'lb_domain',
  'lbl_port',
  'lbl_name',
  'lb_vips',
  'lb_region',
];
const displayProperties = DisplayFieldFactory.createModel(DisplayFieldType.DeviceRs).getProperties();
const displayConfig: Record<string, Partial<ModelPropertyColumn>> = {
  target_group_name: {
    render: ({ row, cell }) => {
      const handleClick = async () => {
        routerAction.open({
          name: MENU_BUSINESS_TARGET_GROUP_DETAILS,
          params: {
            id: row.target_group_id,
          },
          query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value, type: 'list', vendor: props.vendor },
        });
      };
      return h(Button, { theme: 'primary', text: true, onClick: handleClick }, cell);
    },
  },
};
const dataListColumns = displayFieldIds.map((id) => {
  const property = displayProperties.find((field) => field.id === id);
  return { ...property, ...displayConfig[id] };
});

const { t } = useI18n();
const regionStore = useRegionsStore();

const allActiveIP = new Set();
const active = ref<string[]>([props.rsList[0]?.inst_id]);
const count = ref<number>(0);
const max = ref<number>(5000);
const tableRef = ref(null);
const checkStatus = ref([]);

const hasSelection = computed(() => props.type === RsDeviceType.INFO);
const moreData = computed(() => count.value > max.value);
const selections = computed(() => {
  if (count.value === 0 || moreData.value) return [];
  const res = [];
  const list = cloneDeep(props.rsList);
  list.map((item, index) => {
    if (checkStatus.value[index]?.checked) {
      item.targets = tableRef.value[index].getSelection();
      res.push(item);
    }
  });
  return res;
});

const isExpand = (id: string) => active.value.includes(id);
const getLength = (index: number) => {
  const nowChecked = tableRef.value[index].getSelection();
  const { targets } = props.rsList[index];
  const { length } = nowChecked;
  const { length: targetLength } = targets;
  return {
    length,
    targetLength,
  };
};
const handleIPClick = (id: string) => {
  allActiveIP.add(id);
  routerAction.open({
    name: 'hostBusinessDetail',
    query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value, id, type: props.vendor },
  });
};
const handleHeadChange = (val: boolean, id: string, event: any) => {
  if (!event) return;
  const index = props.rsList.findIndex((item) => item.inst_id === id);
  const { targetLength, length } = getLength(index);
  const nowCount = count.value;
  const leftLength = targetLength - length;
  checkStatus.value[index].checked = val;
  checkStatus.value[index].indeterminate = false;
  tableRef.value[index].toggleAllSelection(val);
  count.value = val ? nowCount + leftLength : Math.max(0, nowCount - length);
};
const handleSelectionChange = (id: string, isCheck: boolean) => {
  const nowCount = count.value;
  let checked = false;
  let indeterminate = false;
  const index = props.rsList.findIndex((item) => item.inst_id === id);
  const { length, targetLength } = getLength(index);
  if (length > 0) {
    checked = true;
    if (length !== targetLength) indeterminate = true;
  }
  checkStatus.value[index].checked = checked;
  checkStatus.value[index].indeterminate = indeterminate;
  count.value = isCheck ? nowCount + 1 : Math.max(0, nowCount - 1);
};
const handleSelectAll = () => props.rsList.forEach(({ inst_id }) => handleHeadChange(true, inst_id, true));
const handleClearSelection = () => props.rsList.forEach(({ inst_id }) => handleHeadChange(false, inst_id, true));
const handleSingleDelete = (id: string) => {
  emit('delete', id);
};
const getDisplayCompProps = (column: ModelPropertyColumn, row: any) => {
  const { id } = column;
  if (id === 'region') {
    return { vendor: row.vendor };
  }
  return {};
};

watch(
  () => props.rsList,
  (list) => {
    checkStatus.value = list.map(() => ({
      checked: false,
      indeterminate: false,
    }));
    active.value = [list[0]?.inst_id];
    count.value = 0;
  },
  {
    deep: true,
  },
);

defineExpose({ selections });
</script>

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

    .info {
      width: 250px;

      .name {
        display: inline-block;
        max-width: 140px;
        text-overflow: ellipsis;
        white-space: nowrap;
        overflow: hidden;
        line-height: 100%;
      }
    }
    .rs-num {
      width: 80px;
    }
    .region {
      width: 150px;
    }

    &.is-selected {
      background: #f0f5ff !important;
    }

    .single-delete-btn {
      color: #c4c6cc;
      margin-left: auto;
    }
  }

  div {
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
  }

  .ip {
    color: #3a84ff;
    cursor: pointer;
    display: inline-block;

    &.actived {
      color: #8334f4;
      text-decoration: underline;
    }
  }
  .name {
    color: #979ba5;
  }

  :deep(.bk-table) {
    .bk-table-head {
      .bk-checkbox {
        display: none;
      }
    }
    .bk-table-body {
      max-height: 420px;
    }
  }

  :deep(.bk-collapse-item) {
    margin-bottom: 0 !important;

    .bk-collapse-content {
      padding-left: 0;
      padding-right: 0;
    }

    &:nth-child(even) {
      .header {
        background: #fafbfd;
      }
    }
  }
}
</style>
