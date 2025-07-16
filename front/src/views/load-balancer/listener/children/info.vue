<script setup lang="ts">
import { computed, h, inject, reactive, Ref, ref, useTemplateRef } from 'vue';
import { IListenerDetails, IListenerItem, useLoadBalancerListenerStore } from '@/store/load-balancer/listener';
import { ModelPropertyDisplay } from '@/model/typings';
import { BindingStatus, LAYER_7_LISTENER_PROTOCOL, ListenerProtocol, SESSION_TYPE_NAME } from '../../constants';
import { DisplayFieldType, DisplayFieldFactory } from '../../children/display/field-factory';

import { Form, Message } from 'bkui-vue';
import Panel from '@/components/panel';
import GridDetails from '../../children/display/grid-details.vue';

const props = defineProps<{ rowData: IListenerItem; details: IListenerDetails }>();
const emit = defineEmits<{ 'update-success': [id: string] }>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const isLayer7 = computed(() => LAYER_7_LISTENER_PROTOCOL.includes(props.rowData.protocol));
const isHttps = computed(() => ListenerProtocol.HTTPS === props.rowData.protocol);

const columnProperties = DisplayFieldFactory.createModel(DisplayFieldType.LISTENER).getProperties();

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
    const property = columnProperties.find((item) => item.id === id) as ModelPropertyDisplay;
    return { ...property, ...baseInfoFieldsConfig[id] };
  });
});

const certInfoIds = ['certificate.ssl_mode', 'certificate.ca_cloud_id', 'certificate.cert_cloud_ids'];
const certInfoFields = certInfoIds.map((id) => {
  const property = columnProperties.find((item) => item.id === id) as ModelPropertyDisplay;
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
  const property = columnProperties.find((item) => item.id === id) as ModelPropertyDisplay;
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
  const property = columnProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return { ...property, ...healthCheckFieldsConfig[id] };
});

const isSessionKeeping = computed(() => props.details?.session_expire !== 0);
const isHealthCheck = computed(() => props.details?.health_check?.health_switch === 1);

const formRef = useTemplateRef<typeof Form>('edit-form');
const formModel = reactive({ name: '' });
const rules = {
  name: [
    {
      validator: (value: string) => /^[\u4e00-\u9fa5A-Za-z0-9\-._:]{1,60}$/.test(value),
      message: '不能超过60个字符，只能使用中文、英文、数字、下划线、分隔符“-”、小数点、冒号',
      trigger: 'change',
    },
  ],
};

const isEditDialogShow = ref(false);
const handleEdit = () => {
  isEditDialogShow.value = true;
  formModel.name = props.details.name;
};

const handleConfirmUpdate = async () => {
  await formRef.value.validate();
  const { id, account_id } = props.details;
  await loadBalancerListenerStore.updateListener(
    { id, account_id, name: formModel.name },
    currentGlobalBusinessId.value,
  );
  Message({ theme: 'success', message: '提交成功' });
  isEditDialogShow.value = false;
  emit('update-success', id);
};
</script>

<template>
  <div class="base-info-container">
    <bk-button
      class="fix-button"
      theme="primary"
      outline
      :disabled="props.rowData.binding_status === BindingStatus.BINDING"
      @click="handleEdit"
    >
      编辑
    </bk-button>
    <panel title="基本信息" no-shadow>
      <grid-details
        :fields="baseInfoFields"
        :details="details"
        :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
      />
    </panel>
    <panel v-if="isHttps" title="证书信息" no-shadow>
      <grid-details
        :fields="certInfoFields"
        :details="details"
        :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
      />
    </panel>
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
        :details="details"
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
        </div>
      </template>
      <grid-details
        v-if="isHealthCheck"
        :fields="healthCheckFields"
        :details="details"
        :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
      />
    </panel>

    <bk-dialog
      :is-show="isEditDialogShow"
      title="编辑监听器"
      :width="640"
      :is-loading="loadBalancerListenerStore.updateListenerLoading"
      @confirm="handleConfirmUpdate"
      @closed="isEditDialogShow = false"
    >
      <bk-form ref="edit-form" :model="formModel" form-type="vertical" :rules="rules">
        <bk-form-item label="监听器名称" required property="name">
          <bk-input v-model="formModel.name" />
          <div class="form-control-tips">
            不能超过60个字符，只能使用中文、英文、数字、下划线、分隔符“-”、小数点、冒号
          </div>
        </bk-form-item>
      </bk-form>
    </bk-dialog>
  </div>
</template>

<style scoped lang="scss">
.base-info-container {
  display: flex;
  flex-direction: column;
  gap: 32px;

  .fix-button {
    position: absolute;
    right: 0;
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
