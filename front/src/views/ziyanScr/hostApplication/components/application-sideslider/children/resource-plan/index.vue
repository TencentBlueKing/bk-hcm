<script setup lang="ts">
import { ref, computed, watch, onMounted, h } from 'vue';
import dayjs from 'dayjs';
import isoWeek from 'dayjs/plugin/isoWeek';
import { Tag } from 'bkui-vue';
import { AngleDoubleDownLine, AngleDoubleUpLine } from 'bkui-vue/lib/icon';
import { useConfigRequirementStore, type IRequirementObsProject } from '@/store/config/requirement';
import { useResourcePlanStore, IResourcesDemandItem, ResourcesDemandStatus } from '@/store/resource-plan';
import { RequirementType } from '@/store/config/requirement';
import usePage from '@/hooks/use-page';
import { ModelPropertyColumn } from '@/model/typings';
import routerAction from '@/router/utils/action';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { type ICondition } from '../../typings';
import { deviceGroups } from '../../constants';

const props = withDefaults(defineProps<Props>(), {});

const emit = defineEmits<{
  apply: [data: IResourcesDemandItem];
}>();

dayjs.extend(isoWeek);

interface Props {
  requireType: RequirementType;
  bizId?: number;
  initialCondition?: ICondition;
}
const { pagination, pageParams, handlePageChange, handlePageSizeChange, handleSort } = usePage(false);

const resourcePlanStore = useResourcePlanStore();

const { getRequirementObsProject } = useConfigRequirementStore();
const requirementObsProjectMap = ref<IRequirementObsProject>({});

const defaultSearchValues: () => ICondition = () => ({
  region: [],
  zone: [],
  device_families: [],
  ...props.initialCondition,
});

const searchValues = ref(defaultSearchValues());

const isAlertFolded = ref(false);
const isAlertFoldable = [RequirementType.Regular, RequirementType.Spring].includes(props.requireType);

const columns = computed(() => {
  const baseCol: ModelPropertyColumn[] = [
    {
      id: 'device_family',
      name: '实例族',
      type: 'string',
    },
    {
      id: 'device_type',
      name: '预测通配机型',
      type: 'string',
      width: 150,
    },
    {
      id: 'total_cpu_core',
      name: 'CPU总核数',
      type: 'number',
      align: 'right',
    },
    {
      id: 'total_memory',
      name: '内存总量',
      type: 'number',
      align: 'right',
    },
    {
      id: 'region_name',
      name: '地域',
      type: 'string',
    },
    {
      id: 'plan_type',
      name: '预测类型',
      type: 'string',
      render: ({ row }: { row?: IResourcesDemandItem }) =>
        h(Tag, { theme: row.plan_type === '预测内' ? 'success' : 'warning' }, row.plan_type),
    },
    {
      id: 'zone_name',
      name: '可用区',
      type: 'string',
    },
  ];
  if (props.requireType === RequirementType.ShortRental) {
    baseCol.push({
      id: 'return_plan_time',
      name: '短租退回日期',
      type: 'string',
    });
  }
  return baseCol;
});

const cvmDevicetypeParams = computed(() => {
  const { region, zone, device_families } = searchValues.value;
  return { region_ids: region, zone_ids: zone, device_group: device_families, enable_capacity: true };
});

const demandList = ref<IResourcesDemandItem[]>([]);

watch(pageParams, () => {
  getList();
});

const getList = async () => {
  const params = {
    page: pageParams.value,
    obs_projects: [requirementObsProjectMap.value[props.requireType]],
    region_ids: searchValues.value.region,
    zone_ids: searchValues.value.zone,
    device_families: searchValues.value.device_families,
    device_types: searchValues.value.device_type,
    // 查询当月有效的预测
    expect_time_range: {
      // 当前时间所在月份的第1天往前加1周
      start: dayjs().startOf('month').subtract(1, 'week').startOf('day').format('YYYY-MM-DD'),
      // 当前时间月份的最后1天往后加1周
      end: dayjs().endOf('month').add(1, 'week').endOf('day').format('YYYY-MM-DD'),
    },
    statuses: [ResourcesDemandStatus.CAN_APPLY],
  };
  const { list, count } = await resourcePlanStore.getDemandList(params);

  demandList.value = list;

  pagination.count = count;
};

onMounted(async () => {
  requirementObsProjectMap.value = await getRequirementObsProject();
  getList();
});

const handleDeviceGroupChange = () => {
  searchValues.value.device_type = [];
};
const handleAreaChange = () => {
  searchValues.value.zone = [];
};

const handleSearch = () => {
  if (pagination.current !== 1) {
    pagination.current = 1;
    return;
  }
  getList();
};
const handleReset = () => {
  searchValues.value = defaultSearchValues();
  handleSearch();
};
</script>

<template>
  <div class="search">
    <grid-container layout="vertical" :column="2" :content-min-width="'1fr'" :gap="[16, 24]">
      <grid-item-form-element label="地域">
        <area-selector
          ref="areaSelector"
          v-model="searchValues.region"
          multiple
          clearable
          filterable
          :params="{ resourceType: 'QCLOUDCVM' }"
          @change="handleAreaChange"
        />
      </grid-item-form-element>
      <grid-item-form-element label="可用区">
        <zone-selector
          v-model="searchValues.zone"
          :separate-campus="false"
          multiple
          :params="{
            resourceType: 'QCLOUDCVM',
            region: searchValues.region,
          }"
        />
      </grid-item-form-element>
      <grid-item-form-element label="实例族">
        <bk-select
          v-model="searchValues.device_families"
          multiple
          clearable
          collapse-tags
          @change="handleDeviceGroupChange"
        >
          <bk-option v-for="(item, index) in deviceGroups" :key="index" :value="item" :label="item" />
        </bk-select>
      </grid-item-form-element>
      <grid-item-form-element label="机型">
        <devicetype-selector
          v-model="searchValues.device_type"
          :params="cvmDevicetypeParams"
          resource-type="cvm"
          multiple
        />
      </grid-item-form-element>
      <grid-item :span="2" class="row-action">
        <bk-button theme="primary" @click="handleSearch">查询</bk-button>
        <bk-button @click="handleReset">重置</bk-button>
      </grid-item>
    </grid-container>
  </div>
  <div :class="['alert-content', { foldable: isAlertFoldable, folded: isAlertFolded }]">
    <div class="result-title" v-show="demandList.length">有预测可申请机型</div>
    <bk-alert class="result-alert" theme="warning" v-if="isAlertFoldable">
      <span class="result-alert-title" @click="isAlertFolded = false">
        无预测机型申领建议
        <angle-double-down-line class="icon-down" v-show="isAlertFolded" />
      </span>
      <div class="result-alert-content" v-show="!isAlertFolded">
        <ul>
          <li>
            1. 可增加预测，如需本周申领，到货日期选择本周。
            <bk-button
              theme="primary"
              text
              @click="routerAction.open({ path: '/business/resource-plan', query: { [GLOBAL_BIZS_KEY]: bizId } })"
            >
              去增加预测
            </bk-button>
          </li>
          <li>2. 切换到“滚服项目”或“小额绿通”的需求类型申请，资源限量供给，额度有限</li>
          <li>
            2. 如需查看库存情况，请
            <bk-button
              theme="primary"
              text
              @click="routerAction.open({ path: '/business/hostInventory', query: { [GLOBAL_BIZS_KEY]: bizId } })"
            >
              跳转至主机库存
            </bk-button>
            查看
          </li>
        </ul>
        <span class="button-up" @click="isAlertFolded = true">
          收起
          <angle-double-up-line />
        </span>
      </div>
    </bk-alert>
    <bk-alert class="result-alert normal" theme="info" v-else>
      如找不到想要的机型，则可以切换到常规项目，或报备预测
    </bk-alert>
  </div>
  <div class="list" v-bkloading="{ loading: resourcePlanStore.demandListLoading }">
    <bk-table
      row-hover="auto"
      :data="demandList"
      :pagination="pagination"
      :max-height="'calc(100vh - 480px)'"
      remote-pagination
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      row-key="id"
    >
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :sort="column.sort"
        :align="column.align"
        :width="column.width"
        :render="column.render"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
      <bk-table-column :label="'操作'">
        <template #default="{ row }">
          <bk-button theme="primary" text @click="emit('apply', row)">一键申领</bk-button>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<style lang="scss" scoped>
.search {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #dcdee5;

  .row-action {
    padding: 4px 0;

    :deep(.item-content) {
      gap: 10px;
    }

    .bk-button {
      min-width: 86px;
    }
  }
}

.list {
  margin-top: 16px;
}

.alert-content {
  display: flex;
  align-items: center;
  gap: 8px;

  &.foldable {
    flex-direction: column-reverse;
    align-items: normal;

    &.folded {
      flex-direction: row;
      align-items: center;
      justify-content: space-between;

      .result-alert {
        border-radius: 16px;
        margin-left: auto;

        .result-alert-title {
          font-weight: 400;
          cursor: pointer;

          .icon-down {
            color: #c4c6cc;
            margin: 0 4px;
          }
        }

        :deep(.bk-alert-wraper) {
          padding: 2px 4px;
        }

        &:hover {
          box-shadow: 0 2px 4px 0 #0000001a;

          .icon-down {
            color: #3a84ff;
          }
        }
      }
    }
  }

  .result-title {
    font-weight: 700;
    font-size: 14px;
    color: #4d4f56;
  }

  .result-alert {
    border-radius: 8px;

    .result-alert-title {
      display: flex;
      align-items: center;
      gap: 4px;
      font-weight: 700;
    }

    .button-up {
      position: absolute;
      right: 10px;
      bottom: 10px;
      display: flex;
      align-items: center;
      gap: 4px;
      color: #3a84ff;
      cursor: pointer;
    }

    &.normal {
      background: none;
      border: none;
      color: #4d4f56;

      :deep(.bk-alert-icon-info) {
        color: #979ba5;
        font-size: 14px;
      }
    }
  }
}
</style>
