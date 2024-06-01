<script lang="ts" setup>
import { computed } from 'vue';
import DetailDiffAccount from './children/detail-diff-account.vue';
import DetailDiffCVM from './children/detail-diff-cvm.vue';
import DetailDiffVPC from './children/detail-diff-vpc.vue';
import DetailDiffGcpFirewallRule from './children/detail-diff-gpc-firewall-rule.vue';
import BusinessName from './children/business-name';
import useDetail from './use-detail';
import { CloudType } from '@/typings/account';
import { timeFormatter } from '@/common/util';
import { AUDIT_SOURCE_MAP, AUDIT_ACTION_MAP } from './constants';

const props = defineProps<{
  id: number;
  bizId: number;
  type: string;
}>();

const { details, isLoading } = useDetail(props);

const isJsonDisplay = computed(() =>
  ['security_group', 'eip', 'disk', 'route_table', 'image', 'network_interface', 'subnet'].includes(
    details.value.res_type,
  ),
);

const diffCompMap = {
  account: DetailDiffAccount,
  cvm: DetailDiffCVM,
  vpc: DetailDiffVPC,
  gcp_firewall_rule: DetailDiffGcpFirewallRule,
};
</script>

<template>
  <bk-loading :loading="isLoading">
    <div class="details-list">
      <div class="details-item">
        <div class="item-label">资源类型</div>
        <div class="item-content">{{ details.res_type }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">云厂商</div>
        <div class="item-content">{{ CloudType[details.vendor] }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">云账号</div>
        <div class="item-content">{{ details.account_id }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">业务</div>
        <div class="item-content">
          <business-name :id="details.bk_biz_id" :empty-text="'未分配'"></business-name>
        </div>
      </div>
      <div class="details-item">
        <div class="item-label">实例ID</div>
        <div class="item-content">{{ details.res_id }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">实例名称</div>
        <div class="item-content">{{ details.res_name }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">云资源ID</div>
        <div class="item-content">{{ details.cloud_res_id || '--' }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">动作</div>
        <div class="item-content">{{ AUDIT_ACTION_MAP[details.action] }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">操作者</div>
        <div class="item-content">{{ details.operator }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">操作时间</div>
        <div class="item-content">{{ timeFormatter(details.created_at) }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">请求ID</div>
        <div class="item-content">{{ details.rid }}</div>
      </div>
      <div class="details-item">
        <div class="item-label">来源</div>
        <div class="item-content">{{ AUDIT_SOURCE_MAP[details.source] }}</div>
      </div>
    </div>
    <div class="details-json" v-if="isJsonDisplay">
      <pre><code>{{ details?.detail?.data }}</code></pre>
    </div>
    <div class="details-table" v-else>
      <component
        :is="diffCompMap[details.res_type]"
        :action="details?.action"
        :detail="details?.detail"
        :audit-type="props.type"
        :business-list="businessList"
      ></component>
    </div>
  </bk-loading>
</template>

<style lang="scss" scoped>
.details-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  padding: 24px;
  .details-item {
    display: flex;
    .item-label {
      width: 90px;
      flex: none;
      &::after {
        content: ': ';
      }
    }
  }
}

.details-json {
  margin: 0 24px;
  background: #455070;
  border-radius: 2px;
  color: #bfc6e0;
  font-size: 12px;
  overflow: auto;
}

.details-table {
  padding: 0 24px;
}
</style>
