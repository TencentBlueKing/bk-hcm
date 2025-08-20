<script setup lang="ts">
import { computed, inject, nextTick, reactive, Ref, ref, useTemplateRef, watch, watchEffect } from 'vue';
import { useBusinessStore } from '@/store';
import { IListenerDetails, IListenerModel, useLoadBalancerListenerStore } from '@/store/load-balancer/listener';
import { ILoadBalancerDetails, ILoadBalancerLockStatus, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { ITargetGroupDetails, useLoadBalancerTargetGroupStore } from '@/store/load-balancer/target-group';
import {
  LAYER_7_LISTENER_PROTOCOL,
  LISTENER_PROTOCOL_LIST,
  ListenerProtocol,
  Scheduler,
  SCHEDULER_LIST,
  SCHEDULER_NAME,
  SessionType,
  SSL_MODE_NAME,
  SSLMode,
} from '../constants';
import { goAsyncTaskDetail } from '@/utils';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TARGET_GROUP_DETAILS } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { useSideslider } from '@/hooks/use-sideslider';

import { Form, Message } from 'bkui-vue';
import CertSelector from '@/views/business/load-balancer/clb-view/components/CertSelector';
import TargetGroupSelector from '@/views/business/load-balancer/clb-view/components/TargetGroupSelector';
import RsPreviewTable from './children/rs-preview-table.vue';

interface IProps {
  loadBalancerDetails: ILoadBalancerDetails;
  isEdit?: boolean;
  initialModel?: IListenerDetails;
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits<{ 'confirm-success': [id: string] }>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();
const loadBalancerClbStore = useLoadBalancerClbStore();
const loadBalancerTargetGroupStore = useLoadBalancerTargetGroupStore();
const businessStore = useBusinessStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const title = computed(() => (props.isEdit ? '编辑监听器' : '新增监听器'));

const formRef = useTemplateRef<typeof Form>('form');
const rules = {
  name: [
    {
      validator: (value: string) => /^[\u4e00-\u9fa5A-Za-z0-9\-._:]{1,60}$/.test(value),
      message: '不能超过60个字符，只能使用中文、英文、数字、下划线、分隔符“-”、小数点、冒号',
      trigger: 'change',
    },
  ],
  port: [
    {
      validator: (value: number) => value >= 1 && value <= 65535,
      message: '端口号不符合规范',
      trigger: 'change',
    },
  ],
  domain: [
    {
      validator: (value: string) => /^(?:(?:[a-zA-Z0-9]+-?)+(?:\.[a-zA-Z0-9-]+)+)$/.test(value),
      message: '域名不符合规范',
      trigger: 'change',
    },
  ],
  url: [
    {
      validator: (value: string) => /^\/[\w\-/]*$/.test(value),
      message: 'URL路径不符合规范',
      trigger: 'change',
    },
  ],
  'certificate.cert_cloud_ids': [
    {
      validator: (value: string[]) => value.length <= 2,
      message: '最多选择 2 个证书',
      trigger: 'change',
    },
    {
      validator: (value: string[]) => {
        // 判断证书类型是否重复
        const [cert1, cert2] = svrCertSelectorRef.value.dataList.filter((cert: any) => value.includes(cert.cloud_id));
        return cert1?.encrypt_algorithm !== cert2?.encrypt_algorithm;
      },
      message: '不能选择加密算法相同的证书',
      trigger: 'change',
    },
  ],
};
const formModel = reactive<IListenerModel>({
  id: '',
  account_id: props.loadBalancerDetails.account_id,
  lb_id: props.loadBalancerDetails.id,
  name: '',
  protocol: ListenerProtocol.TCP,
  port: undefined,
  scheduler: undefined,
  session_open: false,
  session_type: SessionType.NORMAL,
  session_expire: 0,
  target_group_id: '',
  domain: '',
  url: '/',
  sni_switch: 0,
  certificate: { ssl_mode: SSLMode.UNIDIRECTIONAL, ca_cloud_id: '', cert_cloud_ids: [] },
});
const bindingTargetGroupName = ref('');
// edit场景下. 记录SNI是否开启, 如果开启, 则编辑的时候不可将其关闭
const isSniOpen = ref(false);
watchEffect(() => {
  if (!props.initialModel) return;
  const { default_domain, target_group_name, session_expire, sni_switch, certificate, extension } = props.initialModel;
  Object.assign(formModel, props.initialModel, {
    domain: default_domain,
    session_open: session_expire !== 0,
    // SNI开启时，证书在域名上；SNI关闭时，域名在监听器上
    certificate: (sni_switch ? certificate : extension.certificate) || {
      ssl_mode: SSLMode.UNIDIRECTIONAL,
      ca_cloud_id: '',
      cert_cloud_ids: [],
    },
  });
  bindingTargetGroupName.value = target_group_name;
  isSniOpen.value = !!sni_switch;
});

const isLayer7 = computed(() => LAYER_7_LISTENER_PROTOCOL.includes(formModel.protocol));
watch([isLayer7, () => formModel.scheduler], ([isLayer7, scheduler]) => {
  // 不支持会话保持的case：七层监听器 | 均衡方式为加权最小连接数
  if (isLayer7 || Scheduler.LEAST_CONN === scheduler) {
    formModel.session_open = false;
  }
});
watch(
  () => formModel.session_open,
  (sessionOpen) => {
    // session_expire传0即为关闭会话保持
    sessionOpen ? (formModel.session_expire = 30) : (formModel.session_expire = 0);
  },
);
watch(
  () => formModel.certificate.ssl_mode,
  (sslMode) => {
    if (!formModel.certificate) return;
    // 如果需要客户端也提供证书, 则需要SSL认证类型为双向认证
    SSLMode.MUTUAL !== sslMode && (formModel.certificate.ca_cloud_id = '');
  },
  { deep: true },
);

const svrCertSelectorRef = useTemplateRef<typeof CertSelector>('svr-cert-selector');
const targetGroupSelectorRef = useTemplateRef<typeof TargetGroupSelector>('target-group-selector');
watch(
  [model, () => formModel.protocol],
  ([isShow]) => {
    if (!isShow || props.isEdit) return;
    formModel.target_group_id = '';
    nextTick(() => {
      targetGroupSelectorRef.value?.handleRefresh();
      formRef.value?.clearValidate();
    });
  },
  { immediate: true },
);

const jumpToTargetGroupDetails = (id: string) => {
  routerAction.open({
    name: MENU_BUSINESS_TARGET_GROUP_DETAILS,
    params: { id },
    query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value, type: 'detail', vendor: props.initialModel.vendor },
  });
};

const targetGroupDetails = ref<ITargetGroupDetails>();

watch(
  () => formModel.target_group_id,
  async (targetGroupId) => {
    if (!targetGroupId) return;
    targetGroupDetails.value = await loadBalancerTargetGroupStore.getTargetGroupDetails(
      targetGroupId,
      currentGlobalBusinessId.value,
    );
  },
  { immediate: true },
);

const loadBalancerLockStatus = ref<ILoadBalancerLockStatus>();
const jumpToAsyncTaskDetails = () => {
  // TODO-CLB: utils替换为hooks
  goAsyncTaskDetail(businessStore.list, loadBalancerLockStatus.value.flow_id, currentGlobalBusinessId.value);
};

const isConfirmLoading = computed(() =>
  props.isEdit ? loadBalancerListenerStore.updateListenerLoading : loadBalancerListenerStore.addListenerLoading,
);
const checkLoadBalancerLockState = async () => {
  const res = await loadBalancerClbStore.getLoadBalancerLockStatus(
    props.loadBalancerDetails.id,
    currentGlobalBusinessId.value,
  );
  loadBalancerLockStatus.value = res;

  if (res.status !== 'success') {
    return Promise.reject();
  }
};
const handleConfirm = async () => {
  await formRef.value.validate();
  await checkLoadBalancerLockState();
  const {
    id,
    account_id,
    name,
    protocol,
    scheduler,
    // session_type,
    session_expire,
    domain,
    url,
    sni_switch,
    certificate,
  } = formModel;
  const isHttps = ListenerProtocol.HTTPS === protocol;
  if (props.isEdit) {
    await loadBalancerListenerStore.updateListener(
      {
        id,
        account_id,
        name,
        sni_switch,
        extension: isHttps ? { certificate } : undefined,
      },
      currentGlobalBusinessId.value,
    );
    // 如果启用了SNI，需要调用规则（域名）更新接口来更新证书信息
    if (sni_switch) {
      await loadBalancerListenerStore.updateDomain(
        id,
        { domain, certificate: isHttps ? certificate : undefined },
        currentGlobalBusinessId.value,
      );
    }
    Message({ theme: 'success', message: '更新成功' });
  } else {
    await loadBalancerListenerStore.addListener(
      {
        ...formModel,
        id: undefined,
        session_open: undefined,
        // session_type: isLayer4 && Scheduler.LEAST_CONN !== scheduler ? session_type : undefined, // 后端要求必填
        session_expire: !isLayer7.value && Scheduler.LEAST_CONN !== scheduler ? session_expire : undefined,
        domain: isLayer7.value ? domain : undefined,
        url: isLayer7.value ? url : undefined,
        sni_switch: isHttps ? sni_switch : undefined,
        certificate: isHttps ? certificate : undefined,
      },
      currentGlobalBusinessId.value,
    );
    Message({ theme: 'success', message: '新增成功' });
  }
  handleCancel();
  emit('confirm-success', props.isEdit ? id : undefined);
};

const handleCancel = () => {
  model.value = false;
};

const { beforeClose } = useSideslider(formModel);
</script>

<template>
  <bk-sideslider
    v-model:is-show="model"
    :title="title"
    :width="640"
    class="listener-form-container"
    :before-close="beforeClose"
  >
    <bk-alert v-if="loadBalancerLockStatus && loadBalancerLockStatus.status !== 'success'" theme="danger" class="mb24">
      <template #title>
        当前负载均衡正在变更中，无法执行此次操作，请稍后再尝试提交。
        <bk-button theme="primary" text @click="jumpToAsyncTaskDetails">查看正在执行的任务</bk-button>
      </template>
    </bk-alert>
    <bk-form ref="form" form-type="vertical" :model="formModel" :rules="rules">
      <bk-form-item label="负载均衡名称">
        <bk-input :model-value="loadBalancerDetails.name" disabled />
      </bk-form-item>
      <bk-form-item label="监听器名称" required property="name">
        <bk-input v-model="formModel.name" />
        <div class="form-control-tips">不能超过60个字符，只能使用中文、英文、数字、下划线、分隔符“-”、小数点、冒号</div>
      </bk-form-item>
      <bk-form-item label="监听协议" required property="protocol">
        <bk-radio-group v-model="formModel.protocol" type="card" :disabled="isEdit">
          <bk-radio-button v-for="protocol in LISTENER_PROTOCOL_LIST" :key="protocol" :label="protocol">
            {{ protocol }}
          </bk-radio-button>
        </bk-radio-group>
      </bk-form-item>
      <bk-form-item label="监听端口" required property="port">
        <bk-input v-model.number="formModel.port" type="number" :disabled="isEdit" />
      </bk-form-item>
      <template v-if="ListenerProtocol.HTTPS === formModel.protocol">
        <div class="flex-row justify-content-between">
          <bk-form-item label="SNI" required property="sni_switch">
            <bk-switcher
              v-model.number="formModel.sni_switch"
              theme="primary"
              :true-value="1"
              :false-value="0"
              :disabled="isSniOpen"
            />
          </bk-form-item>
          <bk-form-item label="SSL解析方式" required property="certificate.ssl_mode">
            <bk-radio-group v-model="formModel.certificate.ssl_mode">
              <bk-radio :label="SSLMode.UNIDIRECTIONAL">
                {{ SSL_MODE_NAME[SSLMode.UNIDIRECTIONAL] }}
                <bk-tag theme="info" class="ml4">推荐</bk-tag>
              </bk-radio>
              <bk-radio :label="SSLMode.MUTUAL" class="ml24">{{ SSL_MODE_NAME[SSLMode.MUTUAL] }}</bk-radio>
            </bk-radio-group>
          </bk-form-item>
        </div>
        <bk-form-item label="服务器证书" required property="certificate.cert_cloud_ids">
          <cert-selector
            ref="svr-cert-selector"
            v-model="formModel.certificate.cert_cloud_ids"
            type="SVR"
            :account-id="formModel.account_id"
          />
        </bk-form-item>
        <bk-form-item
          v-if="formModel.certificate.ssl_mode === SSLMode.MUTUAL"
          label="CA证书"
          required
          property="certificate.ca_cloud_id"
        >
          <cert-selector v-model="formModel.certificate.ca_cloud_id" type="CA" :account-id="formModel.account_id" />
        </bk-form-item>
      </template>
      <template v-if="LAYER_7_LISTENER_PROTOCOL.includes(formModel.protocol) && !isEdit">
        <bk-form-item label="默认域名" required property="domain">
          <bk-input v-model="formModel.domain" />
        </bk-form-item>
        <bk-form-item label="URL路径" required property="url">
          <bk-input v-model="formModel.url" />
        </bk-form-item>
      </template>
      <!-- 新增 -->
      <template v-if="!isEdit">
        <bk-form-item label="均衡方式" required property="scheduler">
          <bk-select v-model="formModel.scheduler">
            <template v-for="scheduler in SCHEDULER_LIST" :key="scheduler">
              <!-- 七层支持IP Hash，四层不支持 -->
              <bk-option
                v-if="isLayer7 || (!isLayer7 && Scheduler.IP_HASH !== scheduler)"
                :id="scheduler"
                :name="SCHEDULER_NAME[scheduler]"
              />
            </template>
          </bk-select>
        </bk-form-item>
        <!-- 四层支持会话保持，七层不支持；均衡方式为加权最小连接数时，不支持配置会话保持 -->
        <template v-if="!isLayer7 && Scheduler.LEAST_CONN !== formModel.scheduler">
          <div class="flex-row">
            <bk-form-item
              label="会话保持"
              required
              property="session_open"
              description="会话保持可使得来自同一 IP 的请求被转发到同一台后端服务器上。参考官方文档https://cloud.tencent.com/document/product/214/6154"
            >
              <bk-switcher v-model="formModel.session_open" theme="primary" />
            </bk-form-item>
            <bk-form-item label="保持时间" required property="session_expire" class="ml40">
              <bk-input
                v-model.number="formModel.session_expire"
                type="number"
                :min="30"
                suffix="秒"
                :disabled="!formModel.session_open"
              />
            </bk-form-item>
          </div>
        </template>
        <bk-form-item label="目标组" required property="target_group_id">
          <!-- TODO-CLB: 使用roll-request进行改造 -->
          <target-group-selector
            ref="target-group-selector"
            v-model="formModel.target_group_id"
            :account-id="formModel.account_id"
            :cloud-vpc-id="loadBalancerDetails.cloud_vpc_id"
            :region="loadBalancerDetails.region"
            :protocol="formModel.protocol"
            :is-cors-v2="loadBalancerDetails.extension.snat_pro"
          />
        </bk-form-item>
      </template>
      <!-- 编辑 -->
      <template v-else>
        <div v-if="!isLayer7" class="mb24">
          <span class="label">已绑定的目标组：</span>
          <bk-button
            v-if="formModel.target_group_id"
            theme="primary"
            text
            @click="jumpToTargetGroupDetails(formModel.target_group_id)"
            class="value"
          >
            {{ bindingTargetGroupName || '未命名' }}
          </bk-button>
          <span v-else class="value">--</span>
        </div>
      </template>
      <bk-form-item v-if="!(isEdit && isLayer7)" label="RS预览">
        <rs-preview-table
          :loading="loadBalancerTargetGroupStore.targetGroupDetailsLoading"
          :list="targetGroupDetails?.target_list"
        />
      </bk-form-item>
    </bk-form>
    <template #footer>
      <bk-button class="mr8" theme="primary" :loading="isConfirmLoading" @click="handleConfirm">提交</bk-button>
      <bk-button :disabled="isConfirmLoading" @click="handleCancel">取消</bk-button>
    </template>
  </bk-sideslider>
</template>

<style scoped lang="scss">
.listener-form-container {
  :deep(.bk-modal-content) {
    padding: 24px 40px;
  }

  .form-control-tips {
    height: 20px;
    line-height: 20px;
    font-size: 12px;
    color: #979ba5;
  }
}
</style>
