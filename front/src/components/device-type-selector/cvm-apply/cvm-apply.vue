<script setup lang="ts">
import { computed, provide, reactive, ref, watch, toRefs, watchEffect } from 'vue';
import isEqual from 'lodash/isEqual';
import { useFormItem } from 'bkui-vue/lib/shared';
import { VendorEnum } from '@/common/constant';
import { RequirementType } from '@/store/config/requirement';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import CreateButton from './children/create-button.vue';
import DetailsInfo from './children/details-info.vue';
import DeviceTypeDialog from './children/device-type-dialog.vue';
import type { ICvmDeviceTypeFormData } from '../typings';
import { useDeviceTypePlan } from './children/use-device-type-plan';
import { useChargeTypeDefault } from './children/use-charge-type-default';

const deviceType = defineModel<ICvmDeviceTypeFormData['deviceTypes'][number]>();

const zones = defineModel<ICvmDeviceTypeFormData['zones']>('zones');

const chargeType = defineModel<ICvmDeviceTypeFormData['chargeType']>('chargeType');

const chargeMonths = defineModel<ICvmDeviceTypeFormData['chargeMonths']>('chargeMonths');

const resAssignType = defineModel<ICvmDeviceTypeFormData['resAssignType']>('resAssignType');

const props = defineProps<{
  bizId: string | number;
  vendor: VendorEnum;
  requireType: RequirementType;
  region: string;
  // 是否为编辑模式，添加需求还是修改需求重试
  editMode: boolean;
  assetId?: string;
  instanceId?: string;
  disabled?: boolean;
  // 是否处于编辑态
  isEditing?: boolean;
}>();

const emit = defineEmits<{
  change: [value: Partial<ICvmDeviceTypeFormData>, from: 'confirm' | 'auto', changed?: { zones: boolean }];
  'update-selected': [value: ICvmDeviceTypeFormData['deviceTypeList']];
}>();

const { cvmChargeTypes } = useCvmChargeType();

// 机型与预测统一的处理逻辑与数据来源
const { availableDeviceTypeMap, loading: chargeTypeDeviceTypeListLoading } = useDeviceTypePlan(toRefs(props));

const deviceTypeList = ref<ICvmDeviceTypeFormData['deviceTypeList']>();

// 滚服继承的机型固资号
const inheritAssetId = ref<ICvmDeviceTypeFormData['inheritAssetId']>(props.assetId);
// 滚服继承的机型实例ID
const inheritInstanceId = ref<ICvmDeviceTypeFormData['inheritInstanceId']>(props.instanceId);

const formItem = useFormItem();

const isRollingServer = computed(() => props.requireType === RequirementType.RollServer);
const isGreenChannel = computed(() => props.requireType === RequirementType.GreenChannel);
const isSpringPool = computed(() => props.requireType === RequirementType.SpringResPool);
const isRollingServerOrGreenChannel = computed(() => isRollingServer.value || isGreenChannel.value);
const isGreenChannelOrSpringPool = computed(() => isGreenChannel.value || isSpringPool.value);

const isInfoMode = ref(props.isEditing || props.editMode);

const defaultData = ref<Partial<ICvmDeviceTypeFormData>>({});
const originalDefaultData = ref<Partial<ICvmDeviceTypeFormData>>({});

const dialogState = reactive({
  isShow: false,
  isHidden: true,
});

// 选中的单个机型，当前仅支持单选
const selectedDeviceType = computed(() => {
  return deviceTypeList.value?.[0];
});

// 计费模式默认值
const { isDefaultFourYears, isGpuDeviceType } = useChargeTypeDefault({
  selectedDeviceType,
  requireType: props.requireType,
});

// 这里初始化值与dialog中目的不同，这里是在详情态时，场景是通过一键申领这种填充默认值的（可能没有计费模式），确保在不点开编辑时正确初始化相关的值
watchEffect(() => {
  // 需要看预测的情况下，如有预测内的预测，则包年包月可用，否则默认选中按量计费
  if (!isRollingServerOrGreenChannel.value && availableDeviceTypeMap.value.get(cvmChargeTypes.PREPAID)?.size === 0) {
    chargeType.value = cvmChargeTypes.POSTPAID_BY_HOUR;
  }

  if (isDefaultFourYears.value) {
    chargeMonths.value = 48;
  }
  if (isGpuDeviceType.value) {
    chargeMonths.value = 72;
  }
});

const updateDefaultData = () => {
  defaultData.value.deviceTypes = deviceType.value ? [deviceType.value] : undefined;
  defaultData.value.zones = zones.value ?? ['all'];
  defaultData.value.chargeType = chargeType.value;
  defaultData.value.chargeMonths = chargeMonths.value;
  defaultData.value.resAssignType = resAssignType.value;
  // 在初始化为编辑模式时，此值为undefined
  defaultData.value.deviceTypeList = deviceTypeList.value;
  defaultData.value.inheritAssetId = inheritAssetId.value;
  // 详情态编辑时使用props传入的实例ID
  defaultData.value.inheritInstanceId = props.instanceId ?? inheritInstanceId.value;

  // 编辑模式记录原始数据，由于编辑模式完整数据需要异步查询，数据是动态变化的，这里通过是否有初始化值判断，来初始化原始值
  if (props.editMode) {
    if (!originalDefaultData.value.deviceTypes?.length) {
      originalDefaultData.value.deviceTypes = defaultData.value.deviceTypes;
    }
    if (!originalDefaultData.value.deviceTypeList?.length) {
      originalDefaultData.value.deviceTypeList = defaultData.value.deviceTypeList;
    }
    if (!originalDefaultData.value.zones?.length) {
      originalDefaultData.value.zones = defaultData.value.zones;
    }
    if (originalDefaultData.value.resAssignType === undefined) {
      originalDefaultData.value.resAssignType = defaultData.value.resAssignType;
    }
  }
};

const handleAdd = () => {
  defaultData.value = {};
  dialogState.isHidden = false;
  dialogState.isShow = true;
};

const handleEdit = () => {
  updateDefaultData();
  dialogState.isHidden = false;
  dialogState.isShow = true;
};

const handleConfirm = (data: ICvmDeviceTypeFormData) => {
  const changed = { zones: !isEqual(zones.value, data.zones) };
  // 机型目前仅支持单选，为了减少消费数据时的适配，这里只返回单个值
  deviceType.value = data.deviceTypes?.[0];
  zones.value = data.zones;
  chargeType.value = data.chargeType;
  chargeMonths.value = data.chargeMonths;
  resAssignType.value = data.resAssignType;
  deviceTypeList.value = data.deviceTypeList;
  inheritAssetId.value = data.inheritAssetId;
  inheritInstanceId.value = data.inheritInstanceId;
  isInfoMode.value = true;
  emit('change', data, 'confirm', changed);
  formItem?.validate('change');
};

const handleUpdateSelected = (value: ICvmDeviceTypeFormData['deviceTypeList']) => {
  // 存在机型，但通过机型查询不到数据的时候，认为是错误数据，清理掉
  if (deviceType.value && !value?.length) {
    deviceType.value = undefined;
  }
  deviceTypeList.value = value;
  emit('update-selected', value);
};

// 详情依赖的数据变化时，更新详情初始化数据
watch(
  () => [
    deviceType.value,
    deviceTypeList.value,
    zones.value,
    chargeType.value,
    chargeMonths.value,
    resAssignType.value,
    () => props.instanceId,
  ],
  () => {
    if (isInfoMode.value) {
      updateDefaultData();
    }
  },
  { immediate: true },
);

// 初始化后外部依赖的数据有变化，需要通过change对外提供
watch(
  () => [deviceType.value, deviceTypeList.value],
  ([deviceType, deviceTypeList]) => {
    emit(
      'change',
      {
        deviceTypes: [deviceType] as ICvmDeviceTypeFormData['deviceTypes'],
        deviceTypeList: deviceTypeList as ICvmDeviceTypeFormData['deviceTypeList'],
      },
      'auto',
    );
  },
);

watch(
  () => props.region,
  (newRegion, oldRegion) => {
    // 新建模式下，地域变化时，切回到添加态
    if (!props.editMode && oldRegion !== undefined && newRegion !== oldRegion) {
      isInfoMode.value = false;
    }
  },
);

provide('isRollingServer', isRollingServer);
provide('isGreenChannel', isGreenChannel);
provide('isSpringPool', isSpringPool);
provide('isRollingServerOrGreenChannel', isRollingServerOrGreenChannel);
provide('isGreenChannelOrSpringPool', isGreenChannelOrSpringPool);
provide('editMode', props.editMode);
provide('isInfoMode', isInfoMode);
</script>

<template>
  <div class="device-type-selector">
    <div class="trigger" @click="handleAdd" v-if="!isInfoMode">
      <slot name="button" v-bind="{ disabled }">
        <create-button :disabled="disabled" />
      </slot>
    </div>
    <details-info
      :require-type="requireType"
      :region="region"
      :data="defaultData"
      :original-data="originalDefaultData"
      @update-selected="handleUpdateSelected"
      @edit="handleEdit"
      v-else
    />
    <template v-if="!dialogState.isHidden">
      <device-type-dialog
        v-model:is-show="dialogState.isShow"
        :biz-id="Number(bizId)"
        :vendor="vendor"
        :require-type="requireType"
        :region="region"
        :charge-type-device-type-map="availableDeviceTypeMap"
        :charge-type-device-type-loading="chargeTypeDeviceTypeListLoading"
        :default-data="defaultData"
        @hidden="dialogState.isHidden = true"
        @confirm="handleConfirm"
      />
    </template>
  </div>
</template>

<style scoped lang="scss">
.device-type-selector {
  .trigger {
    display: inline-flex;
  }
}
</style>
<style lang="scss">
.device-type-dialog {
  .form-label {
    font-size: 12px;
  }

  .required {
    &::after {
      display: inline-block;
      content: '*' !important;
      width: 14px;
      text-align: center;
      color: #ea3636;
    }
  }

  .bottom-dashed {
    border-bottom: 1px dashed #979ba5;
  }
}
</style>
