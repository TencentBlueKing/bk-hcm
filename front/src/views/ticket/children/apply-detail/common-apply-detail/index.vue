<script setup lang="ts">
import { computed } from 'vue';
import Panel from '@/components/panel';
import { type ModelProperty } from '@/model/typings';
import { APPLICATION_TYPE_MAP, APPLICATION_STATUS_MAP } from '@/views/ticket/constants';
import { VendorMap } from '@/common/constant';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import DisplayValue from '@/components/display-value/index.vue';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import { type IApplicationDetail } from '../index';

const props = defineProps<{
  details: IApplicationDetail;
  loading: boolean;
  cancelLoading: boolean;
  onCancel: () => void;
}>();

const baseFields: ModelProperty[] = [
  { id: 'type', name: '申请类型', type: 'enum', option: APPLICATION_TYPE_MAP },
  { id: 'creator', name: '申请人', type: 'user' },
  { id: 'memo', name: '申请单备注', type: 'string' },
  { id: 'created_at', name: '申请时间', type: 'datetime' },
  { id: 'updated_at', name: '更新时间', type: 'datetime' },
];

// 所有类型的字段集合
const paramsFields: ModelProperty[] = [
  { id: 'account_id', name: '账号', type: 'string' },
  { id: 'vendor', name: '云厂商', type: 'enum', option: VendorMap },
  { id: 'bk_biz_id', name: '业务名称', type: 'business' },
  { id: 'action', name: 'action', type: 'string' },
  { id: 'region', name: '云地域', type: 'region' },
  { id: 'zone', name: '可用区', type: 'string' },
  { id: 'cloud_image_id', name: 'cloud_image_id', type: 'string' },
  { id: 'bk_cloud_id', name: 'bk_cloud_id', type: 'string' },
  // 云主机
  { id: 'auto_renew', name: 'auto_renew', type: 'string' },
  { id: 'bandwidth_package_id', name: 'bandwidth_package_id', type: 'string' },
  { id: 'cloud_security_group_ids', name: 'cloud_security_group_ids', type: 'string' },
  { id: 'cloud_subnet_id', name: 'cloud_subnet_id', type: 'string' },
  { id: 'cloud_vpc_id', name: 'cloud_vpc_id', type: 'string' },
  { id: 'data_disk', name: 'data_disk', type: 'json' },
  { id: 'system_disk', name: 'system_disk', type: 'json' },
  { id: 'instance_charge_paid_period', name: 'instance_charge_paid_period', type: 'string' },
  {
    id: 'instance_charge_type',
    name: 'instance_charge_type',
    type: 'enum',
    option: {
      PREPAID: '包年包月',
      POSTPAID_BY_HOUR: '按量计费',
    },
  },
  { id: 'instance_type', name: 'instance_type', type: 'string' },
  { id: 'name', name: 'name', type: 'string' },
  { id: 'required_count', name: 'required_count', type: 'number' },
  // 账号
  { id: 'req', name: 'req', type: 'json' },
  // 硬盘
  { id: 'disk_count', name: 'disk_count', type: 'number' },
  { id: 'disk_name', name: 'disk_name', type: 'string' },
  { id: 'disk_size', name: 'disk_size', type: 'number' },
  { id: 'disk_type', name: 'disk_type', type: 'string' },

  // VPC
  { id: 'routing_mode', name: 'routing_mode', type: 'string' },
  { id: 'subnet', name: 'subnet', type: 'json' },
  { id: 'instance_tenancy', name: 'instance_tenancy', type: 'string' },
  { id: 'ipv4_cidr', name: 'ipv4_cidr', type: 'string' },

  { id: 'memo', name: 'memo', type: 'string' },
];

const detailsParams = computed(() => {
  try {
    const params = JSON.parse(props.details.content);
    return params;
  } catch (error) {
    console.error(error);
    return {};
  }
});

const displayParamsFields = computed(() => {
  const paramsKeys = Object.keys(detailsParams.value);
  return paramsFields.filter((field) => paramsKeys.includes(field.id));
});

const status = computed(() => props.details?.status ?? '');
const deliveryDetail = computed(() => props.details?.delivery_detail ?? '');
const isNotEmptyDeliveryDetail = computed(() => deliveryDetail.value && deliveryDetail.value.trim() !== '{}');
</script>

<template>
  <div class="common-apply-detail-container">
    <Panel class="status-panel">
      <div class="status-icon">
        <bk-loading
          v-if="['pending', 'delivering'].includes(status)"
          style="transform: scale(0.5)"
          mode="spin"
          theme="primary"
          loading
        />
        <i v-else-if="['rejected'].includes(status)" class="hcm-icon bkhcm-icon-38moxingshibai-01" />
        <i v-else-if="['pass', 'completed'].includes(status)" class="hcm-icon bkhcm-icon-7chenggong-01" />
        <i v-else-if="['deliver_error'].includes(status)" class="hcm-icon bkhcm-icon-close-circle-fill"></i>
        <img v-else :src="StatusUnknown" :style="{ width: '22px' }" />
      </div>
      <div class="status-name">{{ APPLICATION_STATUS_MAP[status] }}</div>
      <div
        v-if="isNotEmptyDeliveryDetail"
        :class="[
          'status-message',
          { error: status === 'deliver_error', success: status === 'pass' || status === 'completed' },
        ]"
      >
        <bk-overflow-title type="tips" resizeable class="message-text">
          {{ deliveryDetail }}
        </bk-overflow-title>
        <CopyToClipboard :content="deliveryDetail" />
      </div>
    </Panel>
    <Panel title="基本信息">
      <GridContainer fixed :column="2" :content-min-width="300" :label-width="250">
        <GridItem v-for="field in baseFields" :key="field.id" :label="field.name">
          <DisplayValue
            :property="field"
            :value="details[field.id]"
            :display="{ ...field.meta?.display, on: 'info' }"
          />
        </GridItem>
      </GridContainer>
    </Panel>
    <Panel title="参数信息">
      <GridContainer fixed :column="2" :content-min-width="300" :label-width="250">
        <GridItem v-for="field in displayParamsFields" :key="field.id" :label="field.name">
          <DisplayValue
            :property="field"
            :value="detailsParams[field.id]"
            :vendor="detailsParams.vendor"
            :display="{ ...field.meta?.display, on: 'info' }"
          />
        </GridItem>
      </GridContainer>
    </Panel>
  </div>
</template>

<style scoped lang="scss">
.common-apply-detail-container {
  display: flex;
  flex-direction: column;
  gap: 12px;

  .status-panel {
    display: flex;
    align-items: center;
    gap: 12px;

    .status-icon {
      .hcm-icon {
        font-size: 21px;
        color: #3a84ff;
      }

      .bkhcm-icon-7chenggong-01 {
        color: #2dcb56;
      }

      .bkhcm-icon-38moxingshibai-01,
      .bkhcm-icon-close-circle-fill {
        color: #cc4053;
      }
    }

    .status-name {
      flex-shrink: 0;
      margin-left: 8px;
      color: #313238;
    }

    .status-message {
      display: flex;
      align-items: center;
      gap: 8px;
      flex: 0 1 auto;
      min-width: 0;
      max-width: 1280px;
      .message-text {
        flex: 0 1 auto;
        min-width: 0;
      }
      &.error {
        color: $danger-color;
      }
      &.success {
        color: $success-color;
      }
    }
  }
}
</style>
