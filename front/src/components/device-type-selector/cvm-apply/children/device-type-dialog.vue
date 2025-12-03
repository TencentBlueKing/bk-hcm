<script setup lang="tsx">
import { computed, ref, shallowReactive, onMounted, watch, watchEffect, inject, Ref, provide } from 'vue';
import { Loading } from 'bkui-vue';
import { PrimaryTable, type TableProps } from '@blueking/tdesign-ui';
import { VendorEnum } from '@/common/constant';
import apiService from '@/api/scrApi';
import { useCvmDeviceStore, type ICvmDevicetypeItem, type IRollingServerCvm } from '@/store/cvm/device';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import { transformSimpleCondition } from '@/utils/search';
import { RequirementType } from '@/store/config/requirement';
import { QueryRuleOPEnumLegacy } from '@/typings';
import ChargeType from './charge-type.vue';
import AssetMatch from './asset-match.vue';
import Inventory from './inventory.vue';
import { useZoneFactory } from './use-zone-factory';
import type { AvailableDeviceTypeMap } from './use-device-type-plan';
import { useChargeTypeDefault } from './use-charge-type-default';
import { RES_ASSIGN_TYPE } from '../../constants';
import type { ICvmDeviceTypeFormData } from '../../typings';

const isShow = defineModel<boolean>('isShow');

const props = defineProps<{
  bizId: number;
  vendor: VendorEnum;
  requireType: RequirementType;
  region: string;
  chargeTypeDeviceTypeMap: AvailableDeviceTypeMap;
  chargeTypeDeviceTypeLoading: boolean;
  defaultData?: Partial<ICvmDeviceTypeFormData>;
}>();

const emit = defineEmits<{
  confirm: [applyData: ICvmDeviceTypeFormData];
}>();

const DEVICE_ROW_KEY = 'device_type';
const ZONE_ALL = 'all';

const isRollingServer = inject<Ref<boolean>>('isRollingServer');
const isGreenChannel = inject<Ref<boolean>>('isGreenChannel');
const isRollingServerOrGreenChannel = inject<Ref<boolean>>('isRollingServerOrGreenChannel');
const isGreenChannelOrSpringPool = inject<Ref<boolean>>('isGreenChannelOrSpringPool');

const editMode = inject('editMode');

const cvmDeviceStore = useCvmDeviceStore();

const { cvmChargeTypes } = useCvmChargeType();

// 勾选的可用区
const zoneChecked = ref<string[]>(props.defaultData?.zones?.slice() ?? [ZONE_ALL]);
// 筛选的可用区
const zoneSelected = ref<string[]>([ZONE_ALL]);

// 与申请接口相关的数据
const applyData = shallowReactive<Omit<ICvmDeviceTypeFormData, 'deviceTypes' | 'deviceTypeList' | 'zones'>>({
  chargeType: props.defaultData?.chargeType ?? cvmChargeTypes.PREPAID,
  chargeMonths: props.defaultData?.chargeMonths ?? 36,
  resAssignType: props.defaultData?.resAssignType ?? undefined,
  inheritAssetId: props.defaultData?.inheritAssetId ?? undefined,
  inheritInstanceId: props.defaultData?.inheritInstanceId ?? undefined,
});

// 表格选中的行rowKey
const selectedRowKeys = ref<string[]>(props.defaultData?.deviceTypes ?? []);

const availableDeviceTypeMap = computed(() => props.chargeTypeDeviceTypeMap);
const chargeTypeDeviceTypeListLoading = computed(() => props.chargeTypeDeviceTypeLoading);

// 选中的机型列表
const selectedDeviceTypeList = computed(() => {
  // 当条件变更使用接口重新获取数据后，原本选中的数据可能不在当前列表中，过滤掉不存在的之后保留之前选中的数据
  return selectedRowKeys.value
    .map((key) => deviceTypeList.value.find((item) => item.device_type === key))
    .filter(Boolean);
});
// 选中的单个机型，当前仅支持单选
const selectedDeviceType = computed(() => {
  return selectedDeviceTypeList.value?.[0];
});

// 可用区hook，获取可用区列表
const useZone = useZoneFactory(props.vendor);
const { list: zoneList } = useZone({
  vendor: props.vendor,
  resourceType: 'QCLOUDCVM',
  region: props.region,
});

// 搜索条件的选项
const option = shallowReactive({
  deviceTypeList: [],
  cpuList: [],
  memList: [],
  deviceGroups: ['全部', '标准型', '高IO型', '大数据型', '计算型'],
});

// 搜索条件的值
const condition = shallowReactive({
  deviceGroup: '全部',
  deviceType: [],
  cpu: [],
  mem: [],
  isAvailable: true,
});

// 接口获得的机型列表
const deviceTypeList = ref<ICvmDevicetypeItem[]>([]);

// 滚服继承套餐的机器
const rollingServerCvm = ref<IRollingServerCvm>();

const inventoryState = shallowReactive({
  activeRowKey: null,
  visibleRowKey: null,
});

// 展示的机型列表，根据搜索条件过滤
const displayDeviceTypeList = computed(() => {
  const newList = deviceTypeList.value
    .map((item) => {
      let available = true;
      let maxLimit = undefined;

      if (isRollingServer.value) {
        available =
          'SpecialType' !== item.device_type_class &&
          item.device_group === rollingServerCvm.value?.device_group &&
          !item.device_type.toLowerCase().startsWith('da');
      } else if (isGreenChannel.value) {
        // 小额绿通禁用了available，在接口默认条件中已经过滤掉了，这里直接返回true
        available = true;
      } else {
        // 当前计费模式对应的有效机型
        const chargeTypeAvailableDeviceTypeMap = availableDeviceTypeMap.value.get(applyData.chargeType);

        // 机型接口与机型预测接口的数据不一定对应，也即这里可能得到空
        const availableDeviceType = chargeTypeAvailableDeviceTypeMap?.get(item.device_type);

        available = !!availableDeviceType;
        maxLimit = Math.floor(
          item.cpu_amount > 0 && availableDeviceType ? availableDeviceType?.remain_core / item.cpu_amount : 0,
        );
      }

      return {
        ...item,
        available,
        maxLimit,
      };
    })
    .filter((item) => {
      let isMatch = true;
      if (condition.isAvailable && !chargeTypeDeviceTypeListLoading.value) {
        isMatch = item.available;
      }
      if (isMatch && condition.deviceType.length) {
        isMatch = condition.deviceType.includes(item.device_type);
      }
      if (isMatch && condition.cpu.length) {
        isMatch = condition.cpu.includes(item.cpu_amount);
      }
      if (isMatch && condition.mem.length) {
        isMatch = condition.mem.includes(item.ram_amount);
      }
      return isMatch;
    });
  return newList;
});

const { isDefaultFourYears, isGpuDeviceType } = useChargeTypeDefault({
  selectedDeviceType,
  requireType: props.requireType,
});

watchEffect(() => {
  // 需要看预测的情况下，如有预测内的预测，则包年包月可用，否则默认选中按量计费
  if (!isRollingServerOrGreenChannel.value && availableDeviceTypeMap.value.get(cvmChargeTypes.PREPAID)?.size === 0) {
    applyData.chargeType = cvmChargeTypes.POSTPAID_BY_HOUR;
  }
  if (isDefaultFourYears.value) {
    applyData.chargeMonths = 48;
  }
  if (isGpuDeviceType.value) {
    applyData.chargeMonths = 72;
  }
});

// 机型表格t-table的列配置
const displayColumns = computed(() => {
  let columns: TableProps['columns'] = [
    {
      colKey: 'row-select',
      type: 'single',
      width: 30,
      checkProps: ({ row }) => {
        return { disabled: !row.available };
      },
    },
    {
      colKey: 'device_type',
      title: '机型',
      width: 220,
      cell: (h, { row }) => (
        <div class='device-type'>
          {<span>{row.device_type}</span>}
          <bk-tag theme={row.device_type_class === 'SpecialType' ? 'danger' : 'success'} size='small'>
            {row.device_type_class === 'SpecialType' ? '专用机型' : '通用机型'}
          </bk-tag>
        </div>
      ),
    },
    {
      colKey: 'device_group',
      title: '机型族',
      width: 90,
    },
    {
      colKey: 'cpu_amount',
      title: 'CPU(核)',
      width: 90,
      align: 'right',
      cell: (h, { row }) => `${row.cpu_amount}`,
      sorter: true,
    },
    {
      colKey: 'ram_amount',
      title: '内存(GB)',
      width: 110,
      align: 'right',
      cell: (h, { row }) => `${row.ram_amount}`,
      sorter: true,
    },
    {
      colKey: 'available',
      title: '是否预测',
      width: 80,
      cell: (h, { row }) => (
        <>
          {chargeTypeDeviceTypeListLoading.value && <Loading theme='primary' mode='spin' size='mini' />}
          {!chargeTypeDeviceTypeListLoading.value && <span>{row.available ? '是' : '否'}</span>}
        </>
      ),
    },
    {
      colKey: 'maxLimit',
      title: '可申请数',
      width: 110,
      sorter: true,
      cell: (h, { row }) => (
        <>
          {chargeTypeDeviceTypeListLoading.value && <Loading theme='primary' mode='spin' size='mini' />}
          {!chargeTypeDeviceTypeListLoading.value && (
            <Inventory
              maxLimit={row.maxLimit}
              region={props.region}
              zone={zoneSelected.value?.[0]}
              zoneList={zoneList.value.map((item) => item.id)}
              deviceType={row.device_type}
              chargeType={applyData.chargeType}
              isActive={row[DEVICE_ROW_KEY] === inventoryState.activeRowKey}
              onActive={(deviceType) => (inventoryState.visibleRowKey = deviceType)}
            />
          )}
        </>
      ),
    },
  ];

  if (isRollingServerOrGreenChannel.value) {
    columns = columns.filter((col) => !['available', 'maxLimit'].includes(col.colKey));
  }

  return columns;
});

const pagination: TableProps['pagination'] = shallowReactive({
  current: 1,
  pageSize: 10,
  total: 0,
  size: 'small',
});

// 选中全部并且只有少于或等于一个可用区
const isSingleCheckedAllZone = computed(() => {
  return zoneList.value.length <= 1 && zoneChecked.value?.[0] === ZONE_ALL;
});

const resAssignTypeDisabled = computed(() => {
  return (zoneChecked.value.length === 1 && zoneChecked.value[0] !== ZONE_ALL) || zoneList.value.length <= 1;
});

const confirmDisabled = computed(() => {
  return (
    selectedRowKeys.value.length === 0 ||
    (!resAssignTypeDisabled.value && [undefined, 0].includes(applyData.resAssignType))
  );
});

// 本地搜索等触发列表变化后重置页码
watch(displayDeviceTypeList, () => {
  pagination.current = 1;
  pagination.total = displayDeviceTypeList.value.length;
});

const getOptions = async () => {
  const { cpu, mem } = await apiService.getRestrict();
  option.cpuList = !isGreenChannel.value ? cpu : cpu.filter((v: number) => v <= 16);
  option.memList = mem;
};

const getDeviceTypeList = async () => {
  const baseCondition: Record<string, any> = {
    require_type: isGreenChannelOrSpringPool.value ? RequirementType.Regular : props.requireType,
    region: props.region,
    zone: zoneSelected.value.includes(ZONE_ALL) ? undefined : zoneSelected.value,
    'label.device_group': condition.deviceGroup === '全部' ? undefined : condition.deviceGroup,
  };

  // 小额绿通特殊处理，只展示标准型，且CPU不超过16核
  if (isGreenChannel.value) {
    baseCondition['label.device_group'] = '标准型';
    baseCondition.cpu = 16;
  }

  const params = transformSimpleCondition(
    baseCondition,
    [
      { id: 'require_type', name: 'require_type', type: 'req-type', op: QueryRuleOPEnumLegacy.EQ },
      { id: 'region', name: 'region', type: 'region', op: QueryRuleOPEnumLegacy.EQ },
      { id: 'zone', name: 'zone', type: 'list' },
      { id: 'label.device_group', name: 'label.device_group', op: QueryRuleOPEnumLegacy.IN, type: 'string' },
      { id: 'cpu', name: 'cpu', op: QueryRuleOPEnumLegacy.LESS_OR_EQUAL, type: 'number' },
    ],
    true,
  );

  const { list, count } = await cvmDeviceStore.getDeviceTypeFullList({ filter: params });
  pagination.total = count;
  deviceTypeList.value = list;

  // 机型下拉列表
  option.deviceTypeList = list.slice().sort((a, b) => (a.device_type > b.device_type ? 1 : -1));
};

const onSelectChange: TableProps['onSelectChange'] = (value) => {
  selectedRowKeys.value = value as string[];
};

const onPageChange: TableProps['onPageChange'] = (pageInfo) => {
  pagination.current = pageInfo.current;
  pagination.pageSize = pageInfo.pageSize;
};

const onRowMouseenter: TableProps['onRowMouseenter'] = (context) => {
  inventoryState.activeRowKey = context.row[DEVICE_ROW_KEY];
};

const onRowMouseleave: TableProps['onRowMouseleave'] = () => {
  inventoryState.activeRowKey = null;
};

const sortChange: TableProps['onSortChange'] = (sortInfo) => {
  if (sortInfo && !Array.isArray(sortInfo)) {
    const { sortBy, descending } = sortInfo;

    // 排序时使用原始的接口数据，因展示的displayDeviceTypeList是一个基于过滤条件生成的无set计算属性值
    // 这样做是为了保持统一的规则即通过控制计算属性依赖的变量来控制最终展示的值
    deviceTypeList.value.sort((a, b) => {
      const displayDeviceTypeA = displayDeviceTypeList.value.find(
        (item) => item.device_type === a.device_type,
      ) as ICvmDevicetypeItem;
      const displayDeviceTypeB = displayDeviceTypeList.value.find(
        (item) => item.device_type === b.device_type,
      ) as ICvmDevicetypeItem;

      // 必须使用?.，因为有可能当前排序的行在displayDeviceTypeList中被过滤掉了
      const aSortValue = a[sortBy] ?? displayDeviceTypeA?.[sortBy];
      const bSortValue = b[sortBy] ?? displayDeviceTypeB?.[sortBy];
      if (descending) {
        return bSortValue - aSortValue;
      }
      return aSortValue - bSortValue;
    });
  } else {
    // 重置排序
    getDeviceTypeList();
  }
};

const closeDialog = () => {
  isShow.value = false;
};

onMounted(() => {
  getDeviceTypeList();
  getOptions();
});

const zoneCheckBeforeChange = (nextValue: boolean) => {
  // 至少要保留一个选中
  if (zoneChecked.value.length === 1 && !nextValue) {
    return false;
  }
  return true;
};

// 计费模式变更后清空选中
const handleChargeTypeChange = () => {
  selectedRowKeys.value = [];
};

const handleZoneCheckChange = (checked: boolean, zone: string) => {
  const index = zoneChecked.value.indexOf(zone);
  if (checked) {
    if (zone === ZONE_ALL) {
      zoneChecked.value = [ZONE_ALL];
    } else {
      const allIndex = zoneChecked.value.indexOf(ZONE_ALL);
      if (allIndex !== -1) {
        zoneChecked.value.splice(allIndex, 1);
      }
      zoneChecked.value.push(zone);
    }
  } else if (index !== -1) {
    zoneChecked.value.splice(index, 1);
  }

  if (resAssignTypeDisabled.value) {
    applyData.resAssignType = undefined;
  }
};

const handleZoneSelect = (zone: string) => {
  zoneSelected.value = [zone];
  selectedRowKeys.value = [];
  getDeviceTypeList();
};

const handleDeviceGroupChange = () => {
  selectedRowKeys.value = [];
  getDeviceTypeList();
};

const handleDeviceTypeChange = () => {
  // 小额绿通禁用了仅展示可用机型
  if (!isGreenChannel.value) {
    // 机型筛选时，自动取消仅展示可用机型的勾选
    condition.isAvailable = false;
  }
};

const handleClearSelection = () => {
  selectedRowKeys.value = [];
};

const handleRemoveSelectedItem = (deviceType: string) => {
  const index = selectedRowKeys.value.indexOf(deviceType);
  if (index !== -1) {
    selectedRowKeys.value.splice(index, 1);
  }
};

const handleAssetMatchSuccess = (cvm: IRollingServerCvm) => {
  const { instance_charge_type: chargeType, charge_months: chargeMonths, bk_cloud_inst_id } = cvm;
  applyData.chargeType = chargeType;
  applyData.chargeMonths = chargeType === cvmChargeTypes.PREPAID ? chargeMonths : undefined;
  applyData.inheritInstanceId = bk_cloud_inst_id;

  // 机型族与上次数据不一致时需要清除机型选择
  if (
    rollingServerCvm.value &&
    cvm.device_group !== rollingServerCvm.value.device_group &&
    selectedRowKeys.value.length > 0
  ) {
    selectedRowKeys.value = [];
  }
  rollingServerCvm.value = cvm;
};
const handleAssetMatchFail = () => {
  // 恢复默认值
  applyData.chargeType = cvmChargeTypes.PREPAID;
  applyData.chargeMonths = 36;
  if (selectedRowKeys.value.length > 0) {
    selectedRowKeys.value = [];
  }
  rollingServerCvm.value = null;
};

const handleConfirm = () => {
  emit('confirm', {
    ...applyData,
    zones: zoneChecked.value,
    deviceTypes: selectedRowKeys.value,
    deviceTypeList: selectedDeviceTypeList.value,
    resAssignType: isSingleCheckedAllZone.value ? RES_ASSIGN_TYPE[1].value : applyData.resAssignType,
  });
  closeDialog();
};

provide('requireType', props.requireType);
</script>

<template>
  <bk-dialog
    class="device-type-dialog"
    width="1360"
    v-model:is-show="isShow"
    :render-directive="'if'"
    :close-icon="false"
  >
    <div class="dialog-container">
      <div class="zone">
        <div class="title required">可用区选择</div>
        <div class="content zone-list">
          <div
            :class="['zone-item', { selected: zoneSelected.includes(ZONE_ALL) }]"
            @click.self="handleZoneSelect(ZONE_ALL)"
          >
            <bk-checkbox
              size="small"
              :model-value="ZONE_ALL"
              :true-label="zoneChecked.includes(ZONE_ALL) ? ZONE_ALL : true"
              :checked="zoneChecked.includes(ZONE_ALL)"
              :immediate-emit-change="false"
              :before-change="zoneCheckBeforeChange"
              @change="(checked: boolean) => handleZoneCheckChange(checked, ZONE_ALL)"
            />
            <div class="zone-name" @click.self="handleZoneSelect(ZONE_ALL)">全部</div>
          </div>
          <div
            v-for="(zone, index) in zoneList"
            :key="index"
            :class="['zone-item', { selected: zoneSelected.includes(zone.id) }]"
            @click.self="handleZoneSelect(zone.id)"
          >
            <bk-checkbox
              size="small"
              :model-value="zone.id"
              :true-label="zoneChecked.includes(zone.id) ? zone.id : true"
              :checked="zoneChecked.includes(zone.id)"
              :immediate-emit-change="false"
              :before-change="zoneCheckBeforeChange"
              @change="(checked: boolean) => handleZoneCheckChange(checked, zone.id)"
            />
            <div class="zone-name" @click.self="handleZoneSelect(zone.id)">{{ zone.name }}</div>
          </div>
        </div>
      </div>
      <div class="device-type">
        <div class="title required">机型选择</div>
        <div class="asset-match-container" v-if="isRollingServer">
          <asset-match
            :biz-id="bizId"
            :region="region"
            :inherit-instance-id="applyData.inheritInstanceId"
            v-model="applyData.inheritAssetId"
            @check-success="handleAssetMatchSuccess"
            @check-fail="handleAssetMatchFail"
          />
        </div>
        <div class="content device-type-list">
          <div class="condition-row">
            <div class="device-group">
              <div class="form-label">机型族</div>
              <bk-radio-group type="capsule" v-model="condition.deviceGroup" @change="handleDeviceGroupChange">
                <bk-radio-button
                  v-for="group in option.deviceGroups"
                  :key="group"
                  :label="group"
                  :disabled="isGreenChannel && !['全部', '标准型'].includes(group)"
                />
              </bk-radio-group>
            </div>
            <div class="available-only">
              <bk-checkbox size="small" :disabled="isGreenChannel" v-model="condition.isAvailable">
                仅展示可用机型
              </bk-checkbox>
            </div>
          </div>
          <div class="condition-row">
            <bk-select
              class="device-type-select"
              v-model="condition.deviceType"
              :auto-height="false"
              prefix="机型"
              filterable
              multiple
              @change="handleDeviceTypeChange"
            >
              <bk-option
                v-for="(item, index) in option.deviceTypeList"
                :id="item.device_type"
                :key="index"
                :name="item.device_type"
              />
            </bk-select>
            <bk-select class="cup-select" v-model="condition.cpu" :auto-height="false" prefix="CPU" filterable multiple>
              <bk-option v-for="(item, index) in option.cpuList" :id="item" :key="index" :name="item" />
            </bk-select>
            <bk-select
              class="mem-select"
              v-model="condition.mem"
              :auto-height="false"
              prefix="内存"
              filterable
              multiple
            >
              <bk-option v-for="(item, index) in option.memList" :id="item" :key="index" :name="item" />
            </bk-select>
          </div>

          <div v-bkloading="{ loading: cvmDeviceStore.deviceTypeFullListLoading }">
            <primary-table
              class="device-type-table"
              table-layout="fixed"
              :height="480 - (isRollingServer ? 50 : 0)"
              :hover="true"
              :hide-sort-tips="true"
              :row-key="DEVICE_ROW_KEY"
              :data="displayDeviceTypeList"
              :columns="displayColumns"
              :row-class-name="
                (context: any) => {
                  return {
                    'row-active': context.row[DEVICE_ROW_KEY] === inventoryState.activeRowKey || context.row[DEVICE_ROW_KEY] === inventoryState.visibleRowKey,
                  };
                }
              "
              :pagination="pagination"
              :selected-row-keys="selectedRowKeys"
              @page-change="onPageChange"
              @select-change="onSelectChange"
              @sort-change="sortChange"
              @row-mouseenter="onRowMouseenter"
              @row-mouseleave="onRowMouseleave"
            >
              <template #empty>
                <div style="display: flex; align-items: center; justify-content: center; height: 100px">
                  <span v-if="deviceTypeList.length === 0">机型列表数据为空</span>
                  <span v-else-if="displayDeviceTypeList.length === 0">
                    无可用的机型，如需查看全部机型，请切换筛选条件
                  </span>
                </div>
              </template>
            </primary-table>
          </div>
        </div>
      </div>
      <div class="result-preview">
        <div class="title">结果预览</div>
        <div class="preview-content">
          <div class="content selected-container">
            <div class="toolbar">
              已选机型
              <div class="operation">
                <div class="clear" @click="handleClearSelection">
                  <i class="hcm-icon bkhcm-icon-cc-clear"></i>
                  清空
                </div>
              </div>
            </div>
            <div class="selected-list">
              <div v-for="item in selectedDeviceTypeList" :key="item.device_type" class="selected-item">
                <div class="device-type-name">
                  <bk-overflow-title type="tips">
                    {{ item.device_type }}({{ item.device_group }}, {{ item.cpu_amount }}核{{ item.ram_amount }}GB)
                  </bk-overflow-title>
                </div>
                <i
                  class="hcm-icon bkhcm-icon-close icon-remove"
                  @click="handleRemoveSelectedItem(item.device_type)"
                ></i>
              </div>
            </div>
          </div>
          <div class="charge-type-container" v-if="!editMode">
            <charge-type
              v-model:charge-type="applyData.chargeType"
              v-model:charge-months="applyData.chargeMonths"
              :available-device-type-map="availableDeviceTypeMap"
              :selected-device-type-list="selectedDeviceTypeList"
              :is-default-four-years="isDefaultFourYears"
              :is-gpu-device-type="isGpuDeviceType"
              :is-charge-type-loading="chargeTypeDeviceTypeListLoading"
              @change="handleChargeTypeChange"
            />
          </div>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <div class="assign-type">
          <div :class="['form-label', { required: !resAssignTypeDisabled }]">资源分布方式</div>
          <bk-radio-group
            v-model="applyData.resAssignType"
            size="small"
            :with-validate="false"
            :disabled="resAssignTypeDisabled"
            v-bk-tooltips="{
              content: isSingleCheckedAllZone ? '只有一个可用区时，不支持勾选' : '选择一个可用区时，不支持勾选',
              disabled: !resAssignTypeDisabled,
            }"
          >
            <bk-radio :label="RES_ASSIGN_TYPE[1].value">
              <div
                class="bottom-dashed"
                v-bk-tooltips="{ content: '主机在资源充足的可用区中优先生产', disabled: resAssignTypeDisabled }"
              >
                {{ RES_ASSIGN_TYPE[1].label }}
              </div>
            </bk-radio>
            <bk-radio :label="RES_ASSIGN_TYPE[2].value">
              <div
                class="bottom-dashed"
                v-bk-tooltips="{
                  content: '主机默认在所选可用区（数量>=2）内平均分布。当资源不足时，则任一可用区分布比例上限为50%',
                  disabled: resAssignTypeDisabled,
                }"
              >
                {{ RES_ASSIGN_TYPE[2].label }}
              </div>
            </bk-radio>
          </bk-radio-group>
        </div>
        <bk-button
          theme="primary"
          @click="handleConfirm"
          :disabled="confirmDisabled"
          v-bk-tooltips="{ content: '请选择机型或资源分布方式', disabled: !confirmDisabled }"
        >
          确定
        </bk-button>
        <bk-button @click="closeDialog">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
// stylelint-disable selector-pseudo-class-no-unknown
.device-type-dialog {
  :deep(.bk-dialog-content) {
    padding: 0;
    margin: 0;
  }

  :deep(.bk-dialog-header) {
    display: none;
  }
}

.dialog-container {
  display: flex;
  width: 100%;
  height: 700px;

  .zone {
    flex: none;
    width: 280px;
    background: #fff;
    border-right: 1px solid #dcdee5;
  }

  .device-type {
    flex: none;
    width: 780px;
    background: #fff;

    .asset-match-container {
      padding: 0 24px 10px;
      box-shadow: 0 2px 4px 0 #00000014;
      margin-bottom: 12px;
    }
  }

  .result-preview {
    flex: none;
    width: 300px;
    background: #f5f7fa;
    border-left: 1px solid #dcdee5;
  }

  .title {
    display: flex;
    align-items: center;
    height: 56px;
    padding: 0 24px;
    font-size: 16px;
    color: #313238;
  }

  .content {
    padding: 0 24px;
  }
}

.dialog-footer {
  display: flex;
  align-items: center;
  gap: 8px;

  .assign-type {
    display: flex;
    align-items: center;
    flex: 1;
    gap: 16px;
    color: #313238;

    .form-label {
      &::after {
        display: inline-block;
        content: ' ';
        width: 14px;
      }
    }
  }
}

.zone-list {
  height: calc(100% - 56px);
  overflow: auto;

  .zone-item {
    display: flex;
    align-items: center;
    gap: 6px;
    height: 36px;
    padding: 0 16px;
    margin: 8px 0;
    background: #f5f7fa;
    border: 1px solid transparent;
    border-radius: 2px;
    font-size: 12px;
    cursor: pointer;

    .zone-name {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    &.selected {
      background: #e1ecff;
      border: 1px solid #3a84ff;
    }
  }
}

.device-type-list {
  .condition-row {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 16px;

    .device-group {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .available-only {
      margin-left: auto;

      .bk-checkbox-small {
        font-size: 12px;
      }
    }

    .device-type-select {
      flex: 2;
    }

    .cup-select {
      flex: 1;
    }

    .mem-select {
      flex: 1;
    }
  }
}

:deep(.device-type-table) {
  .device-type {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .bk-loading-size-mini {
    transform: scale(0.75);
  }

  .max-limit {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 4px;

    .details-icon {
      display: none;
      font-size: 14px;
      color: #3a84ff;
      cursor: pointer;
    }
  }

  .row-active {
    background-color: #fafbfd;

    .max-limit {
      .details-icon {
        display: block;
      }
    }
  }
}

.preview-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: calc(100% - 56px);

  .selected-container {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    overflow: auto;

    .toolbar {
      display: flex;
      align-items: center;
      gap: 16px;
      font-size: 12px;
      padding: 4px 0;
      margin-bottom: 6px;
      position: sticky;
      top: 0;
      z-index: 1;
      background: #f5f7fa;

      .operation {
        margin-left: auto;
        display: flex;
        align-items: center;
        gap: 4px;
        color: #979ba5;

        .clear {
          display: flex;
          align-items: center;
          gap: 4px;
          cursor: pointer;

          &:hover {
            color: #63656e;
          }
        }
      }
    }

    .selected-list {
      .selected-item {
        position: relative;
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        height: 32px;
        padding: 0 12px;
        margin: 4px 0;
        background: #fff;
        border-radius: 2px;
        box-shadow: 0 2px 4px 0 #1919290d;

        .device-type-name {
          flex: 1;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .icon-remove {
          display: none;
          position: absolute;
          right: 8px;
          top: 50%;
          transform: translateY(-50%);
          padding: 4px;
          color: #979ba5;
          font-size: 14px;
          cursor: pointer;

          &:hover {
            color: #63656e;
          }
        }

        &:hover {
          padding-right: 30px;
          box-shadow: 0 2px 4px 0 #0000001a, 0 2px 4px 0 #1919290d;

          .icon-remove {
            display: block;
          }
        }
      }
    }
  }

  .charge-type-container {
    margin-top: auto;
    padding: 12px 24px;
    border-top: 1px solid #dcdee5;
    background: #f0f1f5;
  }
}
</style>
