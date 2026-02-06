<script setup lang="ts">
import { computed, h, onBeforeMount, reactive, ref, useTemplateRef, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useZiyanScrStore } from '@/store';
import usePlanStore from '@/store/usePlanStore';
import { GLOBAL_BIZS_KEY, INSTANCE_CHARGE_MAP, VendorEnum } from '@/common/constant';
import { ModelPropertyDisplay } from '@/model/typings';
import isNil from 'lodash/isNil';
import type { IApplyOrderItem } from '@/typings/ziyanScr';
import type { IDemandVerification } from '@/typings/plan';
import { MENU_SERVICE_HOST_APPLICATION, MENU_BUSINESS_TICKET_MANAGEMENT } from '@/constants/menu-symbol';
import { RequirementType } from '@/store/config/requirement';

// TODO: 这些翻译项，后续都要通过display组件进行优化
import { getDiskTypesName, getImageName } from '@/views/ziyanScr/cvm-produce/component/property-display/transform';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';

import { Form, Message } from 'bkui-vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import GridDisplay from './grid-display.vue';
import CollapsePanelGrid from './collapse-panel-grid.vue';
import Panel from '@/components/panel';
import NetworkInfoCollapsePanel from '../network-info-collapse-panel/index.vue';

import DeviceTypeCvmSelector from '@/components/device-type-selector/cvm-apply/cvm-apply.vue';
import type { ICvmDeviceTypeFormData } from '@/components/device-type-selector/typings';

const route = useRoute();
const router = useRouter();
const ziyanScrStore = useZiyanScrStore();
const planStore = usePlanStore();

const businessId = computed(() => Number(route.query.bk_biz_id));
const suborderId = computed(() => route.query.suborder_id as string);

const formRef = useTemplateRef<typeof Form>('adjust-form');
const networkInfoPanelRef = useTemplateRef<typeof NetworkInfoCollapsePanel>('network-info-panel');

const formModel = reactive({ zones: [], res_assign: undefined, device_type: '', replicas: 0, vpc: '', subnet: '' });

const details = ref<IApplyOrderItem>();
const unProductNum = computed(() => (!details.value ? 0 : details.value.origin_num - details.value.product_num));

const chargeType = computed(() => details.value?.spec.charge_type);
const chargeMonths = computed(() => details.value?.spec.charge_months);

const getDetails = async () => {
  const list = await ziyanScrStore.getApplyOrderList({
    bk_biz_id: [businessId.value],
    suborder_id: [suborderId.value],
    get_product: true,
  });
  // suborder_id请求回来的只会是一个单据数据
  [details.value] = list;
};
onBeforeMount(async () => {
  await getDetails();
  // 初始化表单
  const { origin_num, product_num } = details.value || {};
  const { zones, res_assign, device_type, vpc, subnet } = details.value?.spec || {};
  Object.assign(formModel, {
    zones,
    res_assign,
    device_type,
    replicas: origin_num - product_num,
    vpc,
    subnet,
  });
  setNetworkInfoDisabled(formModel.zones);
});

const RESOURCE_TYPE_NAME_MAP: Record<string, string> = {
  QCLOUDCVM: '腾讯云_CVM',
  IDCPM: 'IDC物理机',
  QCLOUDDVM: '腾讯云_DockerVM',
  IDCDVM: 'IDC_DockerVM',
};

const formRules = {
  subnet: [
    {
      validator: (value: string) => (formModel.vpc ? !!value : true),
      message: '选择 VPC 后必须选择子网',
      trigger: 'change',
    },
  ],
  // // 对剩余数量进行校验
  // replicas: [
  //   {
  //     validator: (value: number) => value > 0,
  //     message: '剩余数量必须大于0',
  //     trigger: 'change',
  //   },
  // ],
};

const baseInfoFields: ModelPropertyDisplay[] = [
  { id: 'suborder_id', name: '子单据ID', type: 'string' },
  { id: 'bk_biz_id', name: '业务', type: 'business' },
  { id: 'bk_username', name: '提单人', type: 'user' },
  { id: 'expect_time', name: '期望交付时间', type: 'datetime' },
];
const originDemandFields: ModelPropertyDisplay[] = [
  { id: 'require_type', name: '项目类型', type: 'req-type' },
  { id: 'resource_type', name: '资源类型', type: 'enum', option: RESOURCE_TYPE_NAME_MAP },
  { id: 'spec.device_type', name: '机型', type: 'string' },
  { id: 'spec.charge_type', name: '计费模式', type: 'enum', option: INSTANCE_CHARGE_MAP },
  { id: 'spec.region', name: '地域', type: 'region' },
  {
    id: 'spec.zones',
    name: '园区',
    type: 'array',
    meta: {
      display: {
        format: (val) => {
          if (isNil(val)) {
            return '--';
          }
          if (val === 'all') {
            return '全部可用区';
          }
          return getZoneCn(val);
        },
      },
    },
  },
  {
    id: 'spec.image_id',
    name: '镜像',
    type: 'string',
    meta: {
      display: {
        format: (val) => {
          return isNil(val) ? '--' : `${getImageName(val)}`;
        },
      },
    },
  },
  { id: 'spec.vpc', name: '所属VPC', type: 'string' },
  {
    id: 'spec.disk_type',
    name: '数据盘类型',
    type: 'string',
    render: (details) => {
      const cell = details.spec.disk_type;
      return isNil(cell) ? '--' : `${getDiskTypesName(cell)}(${details.spec.disk_size}G)`;
    },
  },
  { id: 'spec.subnet', name: '所属子网', type: 'string' },
];
const productionFields: ModelPropertyDisplay[] = [
  { id: 'origin_num', name: '需求总数', type: 'number' },
  { id: 'product_num', name: '已生产数', type: 'number' },
  {
    id: 'un_product_num',
    name: '未生产数',
    type: 'number',
    render: () => h('span', { class: 'text-warning' }, unProductNum.value),
  },
];

const setNetworkInfoDisabled = (zones: string[]) => {
  networkInfoDisabled.value = zones?.length !== 1 || zones?.[0] === 'all';
};

const selectedDeviceType = ref<ICvmDeviceTypeFormData['deviceTypeList'][number]>();
// 初始的机型核数
const initialCpuCore = ref(0);
// 剩余生产数量的最大值，初始为 unProductNum，切换机型后动态计算
const maxReplicas = ref<number | null>(null);
const computedMaxReplicas = computed(() => maxReplicas.value ?? unProductNum.value);

const handleDeviceTypeChange = (
  data: Partial<ICvmDeviceTypeFormData>,
  from: 'confirm' | 'auto',
  changed?: { zones: boolean },
) => {
  if (changed?.zones) {
    formModel.subnet = '';
    setNetworkInfoDisabled(data.zones);
  }

  if (!selectedDeviceType.value) {
    selectedDeviceType.value = data?.deviceTypeList?.[0];
    initialCpuCore.value = selectedDeviceType.value?.cpu_core || 0;
  }

  const newDeviceType = data?.deviceTypeList?.[0];

  // 如果机型发生变化且有核数信息，重新计算剩余生产数量
  if (newDeviceType && initialCpuCore.value) {
    const newCpuAmount = newDeviceType.cpu_core;
    // 计算新的剩余生产数量：(修改前核数 × 原剩余生产数量) ÷ 修改后核数，向下取整
    const newReplicas = Math.floor((initialCpuCore.value * unProductNum.value) / newCpuAmount);
    // 更新最大值
    maxReplicas.value = newReplicas;
    // 确保新数量不超过最大值的限制
    formModel.replicas = newReplicas;
  }

  selectedDeviceType.value = newDeviceType;
};

const isNeedVerify = computed(() => {
  return (
    ![RequirementType.RollServer, RequirementType.GreenChannel, RequirementType.SpringResPool].includes(
      details.value?.require_type,
    ) && !verifyResult.value
  );
});
const isVerifyLoading = ref(false);
const verifyResult = ref<IDemandVerification>();
const networkInfoDisabled = ref(false);

const formValidate = async () => {
  try {
    await formRef.value?.validate();
  } catch (error) {
    networkInfoPanelRef.value?.handleToggle(true);
    return Promise.reject(error);
  }
};

const handleVerify = async () => {
  await formValidate();
  const { zones, device_type, vpc, subnet, replicas, res_assign } = formModel;
  const spec = Object.assign({}, details.value.spec, { zones, device_type, vpc, subnet, res_assign });
  isVerifyLoading.value = true;
  try {
    const res = await planStore.verify_resource_demand({
      bk_biz_id: businessId.value,
      require_type: details.value.require_type,
      suborders: [{ ...details.value, replicas, spec }],
    });
    [verifyResult.value] = res.data?.verifications ?? [];
  } catch (error) {
    verifyResult.value = null;
    return Promise.reject(error);
  } finally {
    isVerifyLoading.value = false;
  }
};

watch(formModel, () => {
  verifyResult.value = null;
});

watch(networkInfoDisabled, (disabled) => {
  if (disabled) {
    formModel.vpc = '';
    formModel.subnet = '';
  }
});

const isReplicasInvalid = computed(() => formModel.replicas <= 0);

const handleSubmit = async () => {
  await formValidate();
  const { suborder_id, bk_username } = details.value;
  const { zones, device_type, vpc, subnet, replicas, res_assign } = formModel;
  const spec = Object.assign({}, details.value.spec, { zones, device_type, vpc, subnet, res_assign });
  const res = await ziyanScrStore.modifyApplyOrder({ suborder_id, bk_username, replicas, spec });
  if (res.code === 0) {
    Message({ theme: 'success', message: '提交成功' });
    handleBack();
  }
};
const handleBack = () => {
  const globalBusinessId = route.query[GLOBAL_BIZS_KEY];
  if (globalBusinessId) {
    router.replace({
      name: MENU_BUSINESS_TICKET_MANAGEMENT,
      query: { [GLOBAL_BIZS_KEY]: globalBusinessId, type: 'host_apply' },
    });
  } else {
    router.replace({ name: MENU_SERVICE_HOST_APPLICATION });
  }
};
</script>

<template>
  <detail-header>修改申请</detail-header>
  <div class="container">
    <panel class="panel" title="基本信息">
      <grid-display :fields="baseInfoFields" :details="details" />
    </panel>
    <collapse-panel-grid class="panel" title="原始需求" :fields="originDemandFields" :details="details" />
    <panel class="panel" title="生产情况">
      <grid-display :fields="productionFields" :details="details" />
    </panel>
    <bk-form ref="adjust-form" :model="formModel" form-type="vertical" :rules="formRules">
      <panel title="未生产需求调整">
        <div class="adjust-form-content">
          <bk-form-item required property="device_type">
            <device-type-cvm-selector
              v-model="formModel.device_type"
              v-model:zones="formModel.zones"
              v-model:res-assign-type="formModel.res_assign"
              v-model:charge-type="chargeType"
              v-model:charge-months="chargeMonths"
              :biz-id="businessId"
              :vendor="VendorEnum.ZIYAN"
              :require-type="details?.require_type"
              :region="details?.spec.region"
              :instance-id="details?.spec.inherit_instance_id"
              :edit-mode="true"
              @change="handleDeviceTypeChange"
            />
          </bk-form-item>

          <bk-form-item label="剩余生产数量" required property="replicas">
            <bk-input
              class="form-control"
              v-model.number="formModel.replicas"
              type="number"
              :min="0"
              :max="computedMaxReplicas"
            />
            <div v-if="computedMaxReplicas === 0" class="replicas-warning">核数对应的机器不足为1，请更换机型</div>
            <div v-else-if="formModel.replicas === 0" class="replicas-warning">剩余数量必须大于0</div>
          </bk-form-item>
          <div class="tips">
            <div v-if="selectedDeviceType">
              所需CPU总核心数为 {{ selectedDeviceType?.cpu_core * formModel.replicas }} 核 ({{
                `${selectedDeviceType?.cpu_core}*${formModel.replicas}`
              }})
            </div>
            <div>
              <span class="text-danger">注意：</span>
              已生产 {{ details?.product_num }}，剩余生产数量为
              <span class="text-danger">{{ formModel.replicas }}</span>
              ，将共计生产
              <span class="text-danger">{{ details?.product_num + formModel.replicas }}</span>
              后（原单据需求数为
              <span class="text-danger">{{ details?.origin_num }}</span>
              ），该单据会自动结单，不可以再重试修改
            </div>
          </div>
        </div>
        <network-info-collapse-panel
          ref="network-info-panel"
          v-model:vpc="formModel.vpc"
          v-model:subnet="formModel.subnet"
          vpc-property="vpc"
          subnet-property="subnet"
          :region="details?.spec.region"
          :zone="formModel.zones?.[0]"
          :disabled-vpc="networkInfoDisabled"
          :disabled-subnet="networkInfoDisabled"
        >
          <template #tips>
            <bk-alert
              title="一般需求不需要指定 VPC 和子网，如为 BCS、ODP 等 TKE 场景母机，请提前与平台支持方确认 VPC、子网信息。"
            />
          </template>
        </network-info-collapse-panel>
      </panel>
    </bk-form>
    <panel v-if="verifyResult?.verify_result === 'FAILED'" title="需求预检">
      <bk-alert theme="danger" :title="verifyResult.reason" />
    </panel>
    <div class="footer">
      <bk-button v-if="isNeedVerify" theme="primary" :loading="isVerifyLoading" @click="handleVerify">
        需求校验
      </bk-button>
      <bk-button
        v-else
        theme="primary"
        :loading="ziyanScrStore.modifyApplyOrderLoading"
        :disabled="verifyResult?.verify_result === 'FAILED' || isReplicasInvalid"
        @click="handleSubmit"
      >
        提交修改
      </bk-button>
      <bk-button :disabled="isVerifyLoading || ziyanScrStore.modifyApplyOrderLoading" @click="handleBack">
        取消
      </bk-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.container {
  margin-top: 52px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;

  .adjust-form-content {
    padding: 0 24px;
  }

  .tips {
    margin-top: 8px;
    line-height: normal;
    font-size: 12px;
  }

  .replicas-warning {
    margin-top: 4px;
    font-size: 12px;
    color: #ea3636;
    line-height: normal;
  }

  :deep(.device-type-selector) {
    .device-type-info {
      width: 35%;
    }
  }

  .form-control {
    width: 35%;
  }

  .footer {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}
</style>
