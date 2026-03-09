<script setup lang="ts">
import { computed, nextTick, reactive, ref, useTemplateRef, watch, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '@/store';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import usePlanDeviceType from '@/views/ziyanScr/hostApplication/plan/usePlanDeviceType';
import { useItDeviceType } from './use-it-device-type';
import { VendorEnum } from '@/common/constant';
import { CvmDeviceType } from '@/views/ziyanScr/components/devicetype-selector/types';
import { ICvmDeviceDetailItem } from '@/typings/ziyanScr';
import { ICvmDataDisk } from '@/views/ziyanScr/components/cvm-data-disk/typings';
import { ICvmSystemDisk } from '@/views/ziyanScr/components/cvm-system-disk/typings';
import http from '@/http';

import { Message } from 'bkui-vue';
import { HelpFill } from 'bkui-vue/lib/icon';
import Panel from '@/components/panel';
import AccountSelector from '@/components/account-selector/index-new.vue';
import AreaSelector from '@/views/ziyanScr/hostApplication/components/AreaSelector';
import ZoneTagSelector from '@/components/zone-tag-selector/index.vue';
import PlanLinkAlert from '@/views/ziyanScr/hostApplication/plan/plan-link-alert.vue';
import InheritPackageFormItem, {
  type RollingServerHost,
} from '@/views/ziyanScr/rolling-server/inherit-package-form-item/index.vue';
import ChargeMonthsSelector from './children/charge-months-selector.vue';
import ChargeTypeSelector from './children/charge-type-selector.vue';
import NetworkInfoCollapsePanel from '@/views/ziyanScr/hostApplication/components/network-info-collapse-panel/index.vue';
import CvmDevicetypeSelector from '@/views/ziyanScr/components/devicetype-selector/cvm-devicetype-selector.vue';
import FormCvmImageSelector from '@/views/ziyanScr/components/ostype-selector/form-cvm-image-selector.vue';
import CvmSystemDisk from '@/views/ziyanScr/components/cvm-system-disk/form.vue';
import CvmDataDisk from '@/views/ziyanScr/components/cvm-data-disk/form.vue';
import CvmMaxCapacity from '@/views/ziyanScr/components/cvm-max-capacity/index.vue';
import ModalFooter from '@/components/modal/modal-footer.vue';

const model = defineModel<boolean>();
const props = defineProps<{ cvmDeviceDetail: ICvmDeviceDetailItem }>();
const emit = defineEmits<(e: 'confirm-success' | 'closed') => void>();

const userStore = useUserStore();
const { t } = useI18n();
const { cvmChargeTypes } = useCvmChargeType();

const getDefaultModel = () => {
  // 如果是快速生产，则填充需求类型、云地域、可用区、机型等信息
  const { require_type: requireType, region, zone, device_type: deviceType } = props.cvmDeviceDetail || {};
  return {
    bk_biz_id: 931,
    bk_module_id: 29309,
    require_type: requireType || 1,
    spec: {
      region: region || '',
      zone: zone || '',
      inherit_instance_id: '',
      charge_type: cvmChargeTypes.PREPAID,
      charge_months: 36,
      vpc: '',
      subnet: '',
      device_type: deviceType || '',
      image_id: '',
      system_disk: { disk_type: undefined, disk_size: 0, disk_num: 1 } as ICvmSystemDisk,
      data_disk: [] as ICvmDataDisk[],
    },
    bk_asset_id: '',
    replicas: 1,
    remark: '',
  };
};

const formModel = reactive(getDefaultModel());

const isRollingServer = computed(() => formModel.require_type === 6);
const isSpringPool = computed(() => formModel.require_type === 8);
const isRollingServerLike = computed(() => isRollingServer.value || isSpringPool.value);
const isSpecialRequirement = computed(() => [6, 7].includes(formModel.require_type));
watch(isSpecialRequirement, (val) => {
  if (!val) {
    // 如果不是特殊需求，需要清空继承套餐相关参数
    formModel.bk_asset_id = '';
    formModel.spec.inherit_instance_id = '';
    return;
  }
});
// 云地域变更时，置空可用区、镜像、机型
const handleRegionChange = () => {
  Object.assign(formModel.spec, { zone: '', image_id: '', device_type: '' });
};
// 可用区变更时，置空vpc、子网、机型
const handleZoneChange = () => {
  Object.assign(formModel.spec, { vpc: '', subnet: '', device_type: '' });
};
// 计费模式变更时，处理购买时长默认值
watch(
  () => formModel.spec.charge_type,
  (chargeType: string) => {
    if (chargeType === cvmChargeTypes.POSTPAID_BY_HOUR) {
      formModel.spec.charge_months = undefined;
    } else {
      // 这里需要将calculateChargeMonthsState放到下一个tick中执行，避免计算时用的还是旧的计费模式值
      nextTick(() => {
        const { chargeMonths } = cvmDevicetypeSelectorRef.value?.calculateChargeMonthsState() || {};
        formModel.spec.charge_months = chargeMonths;
      });
    }
  },
);

// 预测
const isPlanAlertShow = computed(() => {
  return (
    !isSpecialRequirement.value && formModel.spec.zone && !hasPlanedDeviceType.value && !isPlanedDeviceTypeLoading.value
  );
});
const cvmDevicetypeSelectorRef = useTemplateRef<typeof CvmDevicetypeSelector>('cvm-devicetype-selector');
const selectedChargeType = computed(() => formModel.spec.charge_type);
const {
  isPlanedDeviceTypeLoading,
  availableDeviceTypeSet,
  computedAvailableDeviceTypeSet,
  hasPlanedDeviceType,
  getPlanedDeviceType,
} = usePlanDeviceType(cvmDevicetypeSelectorRef, selectedChargeType);
// 获取有效预测范围内的机型
watch(
  [() => formModel.require_type, () => formModel.spec.region, () => formModel.spec.zone],
  async ([require_type, region, zone]) => {
    if (!require_type || !region || !zone || isSpecialRequirement.value) return;
    // 业务为：资源运营服务
    await getPlanedDeviceType(931, require_type, region, zone);
    if (availableDeviceTypeSet.prepaid.size === 0) {
      formModel.spec.charge_type = cvmChargeTypes.POSTPAID_BY_HOUR;
    }
  },
  { deep: true },
);

// 滚服
let rollingServerHost: RollingServerHost = null;
const handleInheritPackageValidateSuccess = (host: RollingServerHost) => {
  const { instance_charge_type: chargeType, charge_months: chargeMonths, bk_cloud_inst_id: bkCloudInstId } = host;
  Object.assign(formModel.spec, {
    charge_type: chargeType,
    charge_months: chargeType === cvmChargeTypes.PREPAID ? chargeMonths : undefined,
    inherit_instance_id: bkCloudInstId,
  });
  // 机型族与上次数据不一致时需要清除机型选择
  if (rollingServerHost && host.device_group !== rollingServerHost.device_group && formModel.spec.device_type) {
    formModel.spec.device_type = '';
  }
  rollingServerHost = host;
};
const handleInheritPackageValidateFailed = () => {
  // 恢复默认值
  Object.assign(formModel.spec, { charge_type: cvmChargeTypes.PREPAID, charge_months: 36 });
  formModel.spec.device_type && (formModel.spec.device_type = '');
  rollingServerHost = null;
};

const chargeMonthsDisabledState = ref(null);
const handleDeviceTypeChange = (result: {
  deviceType: CvmDeviceType;
  chargeMonths: number;
  chargeMonthsDisabledState: { disabled: boolean; content: string };
}) => {
  formModel.spec.charge_months = result?.chargeMonths;
  chargeMonthsDisabledState.value = result?.chargeMonthsDisabledState;
};

const currentSpecDeviceType = computed(() => formModel.spec.device_type);
const { currentCloudInstanceConfig, ziyanAccountId, isItDeviceType } = useItDeviceType(
  false,
  currentSpecDeviceType,
  () => {
    const { region, zone, charge_type: chargeType } = formModel.spec;
    return { region, zone, chargeType };
  },
);

// 需求数量
const cvmMaxCapacityQueryParams = computed(() => {
  const { require_type } = formModel;
  const { region, zone, charge_type, vpc, subnet, device_type } = formModel.spec;
  return { require_type, region, zone, charge_type, vpc, subnet, device_type };
});

const isSubmitDisabled = computed(() => !isSpecialRequirement.value && !hasPlanedDeviceType.value);
const isSubmitLoading = ref(false);
const formRef = ref();
const formRules = {
  replicas: [
    {
      validator: (value: number) => !(isRollingServerLike.value && value > 100),
      message: t('注意：因云接口限制，单次的机器数最大值为100，超过后请手动克隆为多条配置'),
      trigger: 'blur',
    },
  ],
  'spec.subnet': [
    {
      validator: (value: string) => (formModel.spec.vpc ? !!value : true),
      message: t('选择 VPC 后必须选择子网'),
      trigger: 'change',
    },
  ],
  'spec.system_disk': [
    {
      validator: (value: ICvmSystemDisk) => !!value.disk_type,
      message: '请选择系统盘类型',
      trigger: 'change',
      required: true,
    },
  ],
  'spec.data_disk': [
    {
      validator: (value: ICvmDataDisk[]) => {
        if (value.length === 0) return true;
        return value.every((item) => item.disk_type && item.disk_size && item.disk_num);
      },
      message: '数据盘信息不能为空',
      trigger: 'change',
    },
  ],
};
const handleConfirm = async () => {
  await formRef.value.validate();
  isSubmitLoading.value = true;
  try {
    const params = { ...formModel, bk_username: userStore.username };
    await http.post('/api/v1/woa/cvm/create/apply/order', params);
    Message({ theme: 'success', message: t('提交成功') });
    emit('confirm-success');
    model.value = false;
  } finally {
    isSubmitLoading.value = false;
  }
};
watchEffect(() => {
  if (model.value) {
    Object.assign(formModel, getDefaultModel());
    nextTick(() => {
      formRef.value.clearValidate();
    });
  } else {
    emit('closed');
  }
});
</script>

<template>
  <bk-dialog v-model:is-show="model" :title="t('CVM生产')" width="1500" show-mask class="create-cvm-dialog">
    <bk-form ref="formRef" label-width="150" :model="formModel" :rules="formRules">
      <!-- 基本信息 -->
      <panel :title="t('基本信息')" class="mb12">
        <div class="form-controls-row">
          <bk-form-item :label="t('云账号')">
            <account-selector
              :model-value="ziyanAccountId"
              :vendor="VendorEnum.ZIYAN"
              disabled
              class="form-controls-item small"
            />
          </bk-form-item>
          <bk-form-item :label="t('业务')" required>{{ t('资源运营服务') }}</bk-form-item>
          <bk-form-item :label="t('模块')" required>{{ t('SA云化池') }}</bk-form-item>
        </div>
        <bk-form-item :label="t('需求类型')" required property="require_type">
          <hcm-form-req-type
            v-model="formModel.require_type"
            appearance="card"
            :filter="(list: any) => list.filter((item: any) => item.require_type !== 8)"
          />
        </bk-form-item>
        <bk-form-item :label="t('云地域')" required property="spec.region">
          <!-- CVM生产这里只生产CVM，直接指定类型即可，无需通过props传入resourceType -->
          <area-selector
            class="form-controls-item"
            v-model="formModel.spec.region"
            :params="{ resourceType: 'QCLOUDCVM' }"
            :popover-options="{ boundary: 'parent' }"
            @change="handleRegionChange"
          />
        </bk-form-item>
        <bk-form-item :label="t('可用区')" required property="spec.zone">
          <zone-tag-selector
            v-model="formModel.spec.zone"
            :key="formModel.spec.region"
            :vendor="VendorEnum.ZIYAN"
            :region="formModel.spec.region"
            :empty-text="t('请先选择云地域')"
            resource-type="QCLOUDCVM"
            auto-expand="selected"
            :separate-campus="true"
            @change="handleZoneChange"
          />
        </bk-form-item>
        <!-- 预测指引 -->
        <bk-form-item v-if="isPlanAlertShow" class="plan-link-alert"><plan-link-alert :bk-biz-id="931" /></bk-form-item>
        <!-- 滚服项目 - 继承套餐 -->
        <inherit-package-form-item
          v-if="isRollingServer"
          v-model="formModel.bk_asset_id"
          :region="formModel.spec.region"
          @validate-success="handleInheritPackageValidateSuccess"
          @validate-failed="handleInheritPackageValidateFailed"
        />
        <bk-form-item :label="t('计费模式')" required property="spec.charge_type">
          <charge-type-selector
            class="form-controls-item"
            v-model="formModel.spec.charge_type"
            :require-type="formModel.require_type"
            :zone="formModel.spec.zone"
            :available-device-type-set="availableDeviceTypeSet"
            :disabled="isRollingServer"
            :tooltips-option="{ content: t('继承原有套餐，计费模式不可选'), disabled: !isRollingServer }"
          />
        </bk-form-item>
        <bk-form-item
          v-if="formModel.spec.charge_type === cvmChargeTypes.PREPAID"
          :label="t('购买时长')"
          required
          property="spec.charge_months"
        >
          <charge-months-selector
            class="form-controls-item"
            v-model="formModel.spec.charge_months"
            :require-type="formModel.require_type"
            :is-gpu-device-type="cvmDevicetypeSelectorRef?.isGpuDeviceType"
            :disabled="chargeMonthsDisabledState?.disabled"
            v-bk-tooltips="{
              content: chargeMonthsDisabledState?.content,
              disabled: !chargeMonthsDisabledState?.disabled,
            }"
          />
        </bk-form-item>
      </panel>
      <!-- 网络信息 -->
      <network-info-collapse-panel
        class="mb12"
        v-model:vpc="formModel.spec.vpc"
        v-model:subnet="formModel.spec.subnet"
        :region="formModel.spec.region"
        :zone="formModel.spec.zone"
        :disabled="formModel.spec.zone === 'cvm_separate_campus'"
        :disabled-vpc="!formModel.spec.region"
        :disabled-subnet="!formModel.spec.vpc"
        v-bk-tooltips="{
          content: t('可用区为分Campus时无法指定子网'),
          disabled: formModel.spec.zone !== 'cvm_separate_campus',
        }"
      >
        <template #tips>
          <bk-alert
            :title="
              t('一般需求不需要指定 VPC 和子网，如为 BCS、ODP 等 TKE 场景母机，请提前与平台支持方确认 VPC、子网信息。')
            "
          />
        </template>
      </network-info-collapse-panel>
      <!-- 实例配置 -->
      <panel :title="t('实例配置')">
        <bk-form-item :label="t('机型')" required property="spec.device_type">
          <cvm-devicetype-selector
            ref="cvm-devicetype-selector"
            v-model="formModel.spec.device_type"
            selector-class="form-controls-item"
            :region="formModel.spec.region"
            :zone="formModel.spec.zone"
            :require-type="formModel.require_type"
            :charge-type="formModel.spec.charge_type"
            :computed-available-device-type-set="computedAvailableDeviceTypeSet"
            :rolling-server-host="rollingServerHost"
            :disabled="!formModel.spec.zone"
            :is-loading="isPlanedDeviceTypeLoading"
            :placeholder="!formModel.spec.zone ? t('请选择可用区') : t('请选择机型')"
            :popover-options="{ boundary: 'parent' }"
            @change="handleDeviceTypeChange"
          />
        </bk-form-item>
        <bk-form-item :label="t('镜像')" required property="spec.image_id">
          <form-cvm-image-selector
            class="form-controls-item"
            v-model="formModel.spec.image_id"
            :region="[formModel.spec.region]"
            :disabled="!formModel.spec.region"
            :popover-options="{ boundary: 'parent' }"
          />
        </bk-form-item>
        <bk-form-item required property="spec.system_disk">
          <template #label>
            系统盘
            <i
              class="hcm-icon bkhcm-icon-prompt text-gray cursor ml4"
              v-bk-tooltips="{ content: '系统盘大小范围为50G-1000G' }"
            />
          </template>
          <cvm-system-disk
            v-model="formModel.spec.system_disk"
            :is-it-device-type="isItDeviceType"
            :current-cloud-instance-config="currentCloudInstanceConfig"
          />
        </bk-form-item>
        <bk-form-item property="spec.data_disk">
          <template #label>
            数据盘
            <i
              class="hcm-icon bkhcm-icon-prompt text-gray cursor ml4"
              v-bk-tooltips="{ content: '数据盘大小范围为20G-32000G，且为10的倍数' }"
            />
          </template>
          <cvm-data-disk
            v-model="formModel.spec.data_disk"
            :is-it-device-type="isItDeviceType"
            :current-cloud-instance-config="currentCloudInstanceConfig"
          />
        </bk-form-item>
        <bk-form-item :label="t('需求数量')" required property="replicas">
          <div class="form-controls-item flex-row align-items-center">
            <hcm-form-number v-model="formModel.replicas" :min="1" :max="1000" />
            <help-fill
              class="ml4 cursor"
              v-bk-tooltips="{
                content: `${t('如果需求数量超过最大可申请量，请提单后联系管理员')} forestchen, dommyzhang`,
              }"
            />
          </div>
          <cvm-max-capacity :params="cvmMaxCapacityQueryParams" />
        </bk-form-item>
        <bk-form-item :label="t('备注')" property="remark">
          <bk-input
            class="form-controls-item"
            v-model="formModel.remark"
            type="textarea"
            show-word-limit
            :resize="false"
            :maxlength="128"
          />
        </bk-form-item>
      </panel>
    </bk-form>
    <template #footer>
      <modal-footer
        :loading="isSubmitLoading"
        :disabled="isSubmitDisabled"
        @confirm="handleConfirm"
        @closed="model = false"
      />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.create-cvm-dialog {
  :deep(.bk-modal-content) {
    background-color: #f5f7fa;
  }

  .form-controls-row {
    display: flex;
    align-items: center;
    gap: 24px;
  }

  .form-controls-item,
  :deep(.form-controls-item),
  :deep(.cvm-vpc-selector),
  :deep(.cvm-subnet-selector) {
    width: 600px;

    &.small {
      width: 300px;
    }
  }

  .plan-link-alert {
    margin-top: -10px;
  }

  .disk-type {
    width: 400px;
  }
}
</style>
