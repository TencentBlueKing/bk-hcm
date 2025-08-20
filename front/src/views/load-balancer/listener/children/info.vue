<script setup lang="ts">
import { computed, h, inject, reactive, Ref } from 'vue';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import { IListenerDetails, IListenerItem, useLoadBalancerListenerStore } from '@/store/load-balancer/listener';
import { ModelPropertyDisplay } from '@/model/typings';
import { BindingStatusType, LAYER_7_LISTENER_PROTOCOL, ListenerProtocol, SESSION_TYPE_NAME } from '../../constants';
import { DisplayFieldType, DisplayFieldFactory } from '../../children/display/field-factory';
import { cloneDeep } from 'lodash';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TARGET_GROUP_DETAILS } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { IAuthSign } from '@/common/auth-service';

import Panel from '@/components/panel';
import GridDetails from '../../children/display/grid-details.vue';
import EditListenerSideslider from '../add.vue';

const props = defineProps<{
  listenerRowData: IListenerItem;
  listenerDetails: IListenerDetails;
  loadBalancerDetails: ILoadBalancerDetails;
}>();
const emit = defineEmits<{ 'update-success': [id: string] }>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();

const clbOperationAuthSign = inject<Ref<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');

const isLayer7 = computed(() => LAYER_7_LISTENER_PROTOCOL.includes(props.listenerRowData.protocol));
const isHttps = computed(() => ListenerProtocol.HTTPS === props.listenerRowData.protocol);

const displayProperties = DisplayFieldFactory.createModel(DisplayFieldType.LISTENER).getProperties();

const baseInfoFieldsConfig: Record<string, Partial<ModelPropertyDisplay & { copy?: boolean }>> = {
  cloud_id: { copy: true },
  protocol_and_port: {
    render: (data: IListenerDetails) => {
      const { protocol, port, end_port } = data ?? {};
      return end_port ? `${protocol}:${port}-${end_port}` : `${protocol}:${port}`;
    },
  },
};
const baseInfoFields = computed(() => {
  const fieldIds = ['name', 'cloud_id', 'protocol_and_port', 'scheduler', 'created_at'];
  if (isLayer7.value) {
    fieldIds.splice(2, 0, 'domain_num', 'url_num');
  }
  return fieldIds.map((id) => {
    const property = displayProperties.find((item) => item.id === id) as ModelPropertyDisplay;
    return { ...property, ...baseInfoFieldsConfig[id] };
  });
});

const certInfoIds = ['certificate.ssl_mode', 'certificate.ca_cloud_id', 'certificate.cert_cloud_ids'];
const certInfoFields = certInfoIds.map((id) => {
  const property = displayProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return property;
});

const sessionIds = ['session_expire_time'];
const sessionFieldsConfig: Record<string, Partial<ModelPropertyDisplay>> = {
  session_expire_time: {
    render: (data: IListenerDetails) => {
      const { session_type, session_expire } = data ?? {};
      return `${SESSION_TYPE_NAME[session_type]}${session_expire} 秒`;
    },
  },
};
const sessionFields = sessionIds.map((id) => {
  const property = displayProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return { ...property, ...sessionFieldsConfig[id] };
});

const healthCheckIds = [
  'health_check.source_ip_type',
  'health_check.check_type',
  'health_check.check_port',
  'health_check.check_scheme',
];
const healthCheckFieldsConfig: Record<string, Partial<ModelPropertyDisplay>> = {
  'health_check.check_scheme': {
    render: (data) => {
      const { time_out, interval_time, un_health_num, health_num } = data?.health_check ?? {};
      return h('div', [
        h('div', `响应超时(${time_out}秒);`),
        h('div', `检查间隔(${interval_time}秒);`),
        h('div', `不健康阈值(${un_health_num}次);`),
        h('div', `健康阈值(${health_num}次);`),
      ]);
    },
  },
};
const healthCheckFields = healthCheckIds.map((id) => {
  const property = displayProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return { ...property, ...healthCheckFieldsConfig[id] };
});

const isSessionKeeping = computed(() => props.listenerDetails?.session_expire !== 0);
const isHealthCheck = computed(() => props.listenerDetails?.health_check?.health_switch === 1);

const editListenerSidesliderState = reactive({ isShow: false, isHidden: true, isEdit: false, initialModel: null });
const handleEditListener = async () => {
  Object.assign(editListenerSidesliderState, {
    isShow: true,
    isHidden: false,
    isEdit: true,
    initialModel: cloneDeep(props.listenerDetails),
  });
};
const handleAddSidesliderConfirmSuccess = (id?: string) => {
  emit('update-success', id);
};
const handleAddSidesliderHidden = () => {
  Object.assign(editListenerSidesliderState, { isShow: false, isHidden: true, isEdit: false, initialModel: null });
};

const openTargetGroupHealthCheckPage = () => {
  const { target_group_id, bk_biz_id, vendor } = props.listenerDetails;
  routerAction.open({
    name: MENU_BUSINESS_TARGET_GROUP_DETAILS,
    params: { id: target_group_id },
    query: {
      [GLOBAL_BIZS_KEY]: bk_biz_id,
      vendor,
      type: 'health',
    },
  });
};
</script>

<template>
  <div class="base-info-container">
    <hcm-auth class="fix-button" :sign="clbOperationAuthSign" v-slot="{ noPerm }">
      <bk-button
        theme="primary"
        outline
        :disabled="noPerm || props.listenerRowData.binding_status === BindingStatusType.BINDING"
        @click="handleEditListener"
      >
        编辑
      </bk-button>
    </hcm-auth>
    <panel title="基本信息" no-shadow>
      <grid-details
        :fields="baseInfoFields"
        :details="listenerDetails"
        :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
      />
    </panel>
    <panel v-if="isHttps" title="证书信息" no-shadow>
      <grid-details
        :fields="certInfoFields"
        :details="listenerDetails"
        :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
      />
    </panel>
    <!-- 七层无会话保持和健康检查 -->
    <template v-if="!isLayer7">
      <panel no-shadow>
        <template #title>
          <div class="panel-header">
            <span class="panel-title">会话保持</span>
            <bk-tag class="ml4" :theme="isSessionKeeping ? 'success' : 'default'">
              {{ isSessionKeeping ? '已开启' : '未开启' }}
            </bk-tag>
          </div>
        </template>
        <grid-details
          v-if="isSessionKeeping"
          :fields="sessionFields"
          :details="listenerDetails"
          :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
        />
      </panel>
      <panel no-shadow>
        <template #title>
          <div class="panel-header">
            <span class="panel-title">健康检查</span>
            <bk-tag class="ml4" :theme="isHealthCheck ? 'success' : 'default'">
              {{ isHealthCheck ? '已开启' : '未开启' }}
            </bk-tag>
            <bk-button
              class="ml4"
              theme="primary"
              text
              v-bk-tooltips="{ content: '跳转至目标组编辑健康检查', placement: 'top-start' }"
              @click="openTargetGroupHealthCheckPage"
            >
              <i class="hcm-icon bkhcm-icon-edit"></i>
            </bk-button>
          </div>
        </template>
        <grid-details
          v-if="isHealthCheck"
          :fields="healthCheckFields"
          :details="listenerDetails"
          :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
        />
      </panel>
    </template>

    <template v-if="!editListenerSidesliderState.isHidden">
      <edit-listener-sideslider
        v-model="editListenerSidesliderState.isShow"
        :load-balancer-details="loadBalancerDetails"
        :is-edit="editListenerSidesliderState.isEdit"
        :initial-model="editListenerSidesliderState.initialModel"
        @confirm-success="handleAddSidesliderConfirmSuccess"
        @hidden="handleAddSidesliderHidden"
      />
    </template>
  </div>
</template>

<style scoped lang="scss">
.base-info-container {
  display: flex;
  flex-direction: column;
  gap: 32px;

  .fix-button {
    position: absolute;
    right: 40px;
  }

  .panel-header {
    margin-bottom: 16px;

    .panel-title {
      font-size: 14px;
      color: #313238;
      font-weight: 700;
    }
  }
}

.form-control-tips {
  height: 20px;
  line-height: 20px;
  font-size: 12px;
  color: #979ba5;
}
</style>
