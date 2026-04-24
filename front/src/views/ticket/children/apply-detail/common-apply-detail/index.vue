<script setup lang="ts">
import { computed } from 'vue';
import share from 'bkui-vue/lib/icon/share';
import Panel from '@/components/panel';
import { type ModelProperty } from '@/model/typings';
import { APPLICATION_TYPE_MAP, APPLICATION_STATUS_MAP } from '@/views/ticket/constants';
import { VendorMap, SITE_TYPE_MAP, ACCOUNT_TYPES } from '@/common/constant';
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
  { id: 'operation', name: '申请类型', type: 'enum', option: APPLICATION_TYPE_MAP },
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
  { id: 'region', name: '云地域', type: 'region' },
  { id: 'zone', name: '可用区', type: 'string' },
  { id: 'cloud_image_id', name: '镜像', type: 'string' },
  { id: 'bk_cloud_id', name: '所属的蓝鲸云区域', type: 'string' },
  { id: 'memo', name: '备注', type: 'string' },

  // 云主机
  { id: 'name', name: '实例名称', type: 'string' },
  { id: 'instance_type', name: '机型', type: 'string' },
  { id: 'public_ip_assigned', name: '是否自动分配公网IP', type: 'bool', option: { trueText: '是', falseText: '否' } },
  { id: 'bandwidth_package_id', name: '带宽包ID', type: 'string' },
  { id: 'cloud_security_group_ids', name: '安全组', type: 'string' },
  { id: 'cloud_subnet_id', name: '子网', type: 'string' },
  { id: 'cloud_vpc_id', name: 'VPC', type: 'string' },
  { id: 'data_disk', name: '数据盘', type: 'json' },
  { id: 'system_disk', name: '系统盘', type: 'json' },
  { id: 'instance_charge_paid_period', name: '购买时长', type: 'number' },
  { id: 'auto_renew', name: '是否自动续费', type: 'bool', option: { trueText: '是', falseText: '否' } },
  {
    id: 'instance_charge_type',
    name: '计费模式',
    type: 'enum',
    option: {
      PREPAID: '包年包月',
      POSTPAID_BY_HOUR: '按量计费',
    },
  },
  { id: 'required_count', name: '购买数量', type: 'number' },

  // 账号
  { id: 'managers', name: '负责人', type: 'user' },
  {
    id: 'site',
    name: '站点类型',
    type: 'enum',
    option: SITE_TYPE_MAP,
  },
  {
    id: 'type',
    name: '账号类型',
    type: 'enum',
    option: Object.fromEntries(ACCOUNT_TYPES.map((item) => [item.id, item.name])),
  },
  { id: 'usage_biz_ids', name: '使用业务', type: 'business' },
  { id: 'policy_library_id', name: '策略库ID', type: 'string' },
  { id: 'secret_id', name: 'secret_id', type: 'string' },
  { id: 'target_status', name: 'target_status', type: 'string' },
  { id: 'extension', name: '扩展字段', type: 'json' },
  { id: 'req', name: 'req', type: 'json' },

  // 硬盘
  { id: 'disk_count', name: 'disk_count', type: 'number' },
  { id: 'disk_name', name: 'disk_name', type: 'string' },
  { id: 'disk_size', name: 'disk_size', type: 'number' },
  { id: 'disk_type', name: 'disk_type', type: 'string' },

  // VPC
  { id: 'routing_mode', name: 'routing_mode', type: 'string' },
  { id: 'subnet', name: 'subnet', type: 'json' },
  { id: 'instance_tenancy', name: '租期', type: 'enum', option: { default: '默认', dedicated: '专用' } },
  { id: 'ipv4_cidr', name: 'IPv4 CIDR', type: 'string' },
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
        <bk-loading v-if="['pending', 'delivering'].includes(status)" size="mini" mode="spin" theme="primary" loading />
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
      <div class="status-link">
        <bk-link theme="primary" :href="details.ticket_url" target="_blank">
          <div class="link-content">
            ITSM单据
            <share width="12" height="12" />
          </div>
        </bk-link>
      </div>
    </Panel>
    <Panel title="基本信息">
      <GridContainer :column="2" fixed :content-min-width="200" :content-max-width="400" :label-width="240">
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
      <GridContainer :column="2" fixed :content-min-width="200" :content-max-width="400" :label-width="240">
        <GridItem v-for="field in displayParamsFields" :key="field.id" :label="field.name">
          <DisplayValue
            :property="field"
            :value="detailsParams[field.id]"
            :vendor="detailsParams.vendor"
            :display="{ ...field.meta?.display, showOverflowTooltip: true, on: 'info' }"
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

    .status-link {
      margin-left: auto;
      flex-shrink: 0;

      .link-content {
        display: flex;
        align-items: center;
        gap: 4px;
      }
    }
  }
}
</style>
