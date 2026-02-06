<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import usePage from '@/hooks/use-page';
import { useCvmDeviceStore, type ICvmDeviceItem } from '@/store/cvm/device';
import { RequirementType } from '@/store/config/requirement';
import { transformSimpleCondition } from '@/utils/search';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import { QueryRuleOPEnum, QueryRuleOPEnumLegacy } from '@/typings';
import { getColumnName } from '@/model/utils';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import apiService from '@/api/scrApi';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneSelector from '@/views/ziyanScr/hostApplication/components/ZoneSelector';
import DevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/index.vue';
import { VendorEnum } from '@/common/constant';
import { type ICondition } from '../../typings';
import { deviceGroupsMap } from '../../constants';

interface Props {
  requireType?: RequirementType;
  bizId?: number;
  initialCondition?: ICondition;
}

const props = withDefaults(defineProps<Props>(), {});

const emit = defineEmits<{
  apply: [data: ICvmDeviceItem];
}>();

const { pagination, pageParams, handlePageChange, handlePageSizeChange, handleSort } = usePage(false);

const cvmDeviceStore = useCvmDeviceStore();

const deviceGroups = computed(() => deviceGroupsMap[props.requireType] || deviceGroupsMap.default);
const defaultSearchValues: () => ICondition = () => ({
  // 小额绿通和春保资源池使用常规项目查询，春保资源池暂已下线
  require_type: [RequirementType.GreenChannel, RequirementType.SpringResPool].includes(props.requireType)
    ? RequirementType.Regular
    : props.requireType,
  region: [],
  zone: [],
  device_families: [deviceGroups.value[0]],
  ...props.initialCondition,
});

const searchValues = ref(defaultSearchValues());

const searchFields: ModelPropertySearch[] = [
  {
    id: 'require_type',
    name: '需求类型',
    type: 'number',
    op: QueryRuleOPEnum.EQ,
  },
  {
    id: 'region',
    name: '地域',
    type: 'array',
    meta: {
      search: {
        filterRules(value: string[]) {
          return {
            field: 'dc.region',
            op: QueryRuleOPEnum.IN,
            value,
          };
        },
      },
    },
  },
  {
    id: 'zone',
    name: '可用区',
    type: 'array',
    meta: {
      search: {
        filterRules(value: string[]) {
          return {
            field: 'dc.zone',
            op: QueryRuleOPEnum.IN,
            value,
          };
        },
      },
    },
  },
  {
    id: 'device_families',
    name: '实例族',
    type: 'array',
    meta: {
      search: {
        filterRules(value: string[]) {
          return {
            field: 'device_family',
            op: QueryRuleOPEnum.IN,
            value,
          };
        },
      },
    },
  },
  {
    id: 'device_type',
    name: '机型',
    type: 'array',
    meta: {
      search: {
        filterRules(value: string[]) {
          return {
            field: 'dc.device_type',
            op: QueryRuleOPEnum.IN,
            value,
          };
        },
      },
    },
  },
  {
    id: 'cpu',
    name: 'CPU',
    type: 'number',
    op: QueryRuleOPEnumLegacy.EQ,
    meta: {
      search: {
        filterRules(value: number) {
          return {
            field: 'cpu_core',
            op: QueryRuleOPEnum.EQ,
            value,
          };
        },
      },
    },
  },
  {
    id: 'mem',
    name: '内存',
    type: 'number',
    op: QueryRuleOPEnumLegacy.EQ,
    meta: {
      search: {
        filterRules(value: number) {
          return {
            field: 'memory',
            op: QueryRuleOPEnum.EQ,
            value,
          };
        },
      },
    },
  },
];

const columns: ModelPropertyColumn[] = [
  {
    id: 'region',
    name: '地域',
    type: 'region',
    render: ({ row }: any) => getRegionCn(row.region),
  },
  {
    id: 'zone',
    name: '可用区',
    type: 'string',
    render: ({ row }: any) => getZoneCn(row.zone),
  },
  {
    id: 'device_type_class',
    name: '机型类型',
    type: 'enum',
    option: {
      SpecialType: '专用机型',
      CommonType: '通用机型',
    },
  },
  {
    id: 'device_type',
    name: '机型',
    type: 'string',
  },
  {
    id: 'cpu_core',
    name: 'CPU',
    type: 'number',
    unit: '核',
    width: 110,
    align: 'right',
  },
  {
    id: 'memory',
    name: '内存',
    type: 'number',
    unit: 'GB',
    width: 110,
    align: 'right',
  },
];

const cvmDevicetypeParams = computed(() => {
  const { region, zone, device_families } = searchValues.value;
  return {
    vendor: VendorEnum.ZIYAN,
    region,
    zone,
    device_family: device_families,
    disable: false,
  };
});
const isGreenChannel = computed(() => props.requireType === RequirementType.GreenChannel);

const deviceList = ref<ICvmDeviceItem[]>([]);
const cpuList = ref<number[]>([]);
const memList = ref<number[]>([]);

watch(pageParams, () => {
  getList();
});

const getSearchAndFields = () => {
  const fields = [...searchFields];
  const cpuItem = fields.find((item) => item.id === 'cpu');

  if (isGreenChannel.value && !searchValues.value.cpu) {
    cpuItem.op = QueryRuleOPEnumLegacy.LESS_OR_EQUAL;
  } else {
    cpuItem.op = QueryRuleOPEnumLegacy.EQ;
  }

  return fields;
};

const getSearchParams = () => {
  const params = { ...searchValues.value };
  if (isGreenChannel.value && !params.cpu) {
    params.cpu = cpuList.value.at(-1);
  }

  return params;
};

const getList = async () => {
  const fields = getSearchAndFields();
  const params = getSearchParams();
  const { list, count } = await cvmDeviceStore.getDeviceList({
    filter: transformSimpleCondition(params, fields, false),
    page: pageParams.value,
  });

  deviceList.value = list;
  pagination.count = count;
};

const getOptions = async () => {
  const { cpu, mem } = await apiService.getRestrict();
  cpuList.value = !isGreenChannel.value ? cpu : cpu.filter((v: number) => v <= 16);
  memList.value = mem;
};

onMounted(async () => {
  await getOptions();
  getList();
});

const handleDeviceGroupChange = () => {
  searchValues.value.device_type = [];
  if (isGreenChannel.value) {
    searchValues.value.device_families = ['标准型'];
  }
};
const handleAreaChange = () => {
  searchValues.value.zone = [];
};
const handleDevicetypeChange = () => {
  searchValues.value.cpu = undefined;
  searchValues.value.mem = undefined;
};
const handleDeviceRestrictChange = () => {
  searchValues.value.device_type = [];
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
    <grid-container layout="vertical" :column="3" :content-min-width="'1fr'" :gap="[16, 24]">
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
          :clearable="false"
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
          :disabled="searchValues.cpu > 0 || searchValues.mem > 0"
          resource-type="cvm"
          multiple
          @change="handleDevicetypeChange"
        />
      </grid-item-form-element>
      <grid-item-form-element label="CPU（核）">
        <bk-select
          v-model="searchValues.cpu"
          :disabled="searchValues.device_type?.length > 0"
          clearable
          filterable
          @change="handleDeviceRestrictChange"
        >
          <bk-option v-for="(item, index) in cpuList" :key="index" :value="item" :label="item" />
        </bk-select>
      </grid-item-form-element>
      <grid-item-form-element label="内存（G）">
        <bk-select
          v-model="searchValues.mem"
          :disabled="searchValues.device_type?.length > 0"
          clearable
          filterable
          @change="handleDeviceRestrictChange"
        >
          <bk-option v-for="(item, index) in memList" :key="index" :value="item" :label="item" />
        </bk-select>
      </grid-item-form-element>
      <grid-item :span="3" class="row-action">
        <bk-button theme="primary" @click="handleSearch">查询</bk-button>
        <bk-button @click="handleReset">重置</bk-button>
      </grid-item>
    </grid-container>
  </div>
  <div class="alert-content">
    <div class="result-title">可申请机型</div>
  </div>
  <div class="list" v-bkloading="{ loading: cvmDeviceStore.deviceListLoading }">
    <bk-table
      row-hover="auto"
      :data="deviceList"
      :pagination="pagination"
      :max-height="'calc(100vh - 410px)'"
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
        :label="getColumnName(column)"
        :sort="column.sort"
        :align="column.align"
        :width="column.width"
        :render="column.render"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
      <bk-table-column :label="'操作'" :width="110">
        <template #default="{ row }: { row: ICvmDeviceItem }">
          <!-- 滚服项目、小额绿通不支持一键申领专用机型 -->
          <bk-button
            theme="primary"
            text
            :disabled="
              [RequirementType.GreenChannel, RequirementType.RollServer].includes(requireType) &&
              row.device_type_class === 'SpecialType'
            "
            v-bk-tooltips="{
              content: '滚服项目、小额绿通不支持一键申领专用机型',
              disabled: !(
                [RequirementType.GreenChannel, RequirementType.RollServer].includes(requireType) &&
                row.device_type_class === 'SpecialType'
              ),
            }"
            @click="emit('apply', row)"
          >
            一键申领
          </bk-button>
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

  .result-title {
    font-weight: 700;
    font-size: 14px;
    color: #4d4f56;
  }

  .result-alert {
    margin-left: auto;
  }
}
</style>
