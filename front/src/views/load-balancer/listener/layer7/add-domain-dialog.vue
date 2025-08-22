<script setup lang="ts">
import { computed, h, inject, nextTick, reactive, Ref, useTemplateRef, watch, watchEffect } from 'vue';
import { DisplayFieldFactory, DisplayFieldType } from '../../children/display/field-factory';
import { ModelPropertyDisplay } from '@/model/typings';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import {
  IListenerDomainInfoItem,
  IListenerItem,
  IListenerRuleModel,
  useLoadBalancerListenerStore,
} from '@/store/load-balancer/listener';
import {
  ListenerProtocol,
  SCHEDULER_LIST,
  SCHEDULER_NAME,
  SSL_MODE_NAME,
  SSLMode,
  DOMAIN_REGEX,
} from '../../constants';

import { Form, Message, Tag } from 'bkui-vue';
import GridDetails from '../../children/display/grid-details.vue';
import ModalFooter from '@/components/modal/modal-footer.vue';
import TargetGroupSelector from '@/views/business/load-balancer/clb-view/components/TargetGroupSelector';
import CertSelector from '@/views/business/load-balancer/clb-view/components/CertSelector';

interface IProps {
  listenerRowData: IListenerItem;
  loadBalancerDetails: ILoadBalancerDetails;
  isEdit?: boolean;
  // TODO-CLB：这里可能存在回填问题，旧版负载均衡也没有回填certificate相关信息。需要找产品确认一下。
  initialModel?: IListenerDomainInfoItem;
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits<{
  'confirm-success': [isEdit: boolean, newDomainInfo: IListenerDomainInfoItem];
}>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();
const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const title = computed(() => (props.isEdit ? '编辑域名' : '新增域名'));

const displayProperties = DisplayFieldFactory.createModel(DisplayFieldType.LISTENER).getProperties();
const fieldIds = ['name', 'protocol_and_port', 'sni_switch'];
const fieldConfig: Record<string, Partial<ModelPropertyDisplay>> = {
  protocol_and_port: {
    render: (data: IListenerItem) => {
      const { protocol, port, end_port } = data ?? {};
      return end_port ? `${protocol}:${port}-${end_port}` : `${protocol}:${port}`;
    },
  },
  sni_switch: {
    render: (data: IListenerItem) => {
      const isOpen = data.sni_switch === 1;
      return h(Tag, { theme: isOpen ? 'success' : '' }, isOpen ? '已开启' : '未开启');
    },
  },
};
const displayFields = fieldIds.map((id) => {
  const property = displayProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return { ...property, ...fieldConfig[id] };
});

const formRef = useTemplateRef<typeof Form>('form');
const rules = {
  domain: [
    {
      validator: (value: string) => DOMAIN_REGEX.test(value),
      message: '域名不符合规范',
      trigger: 'change',
    },
  ],
  url: [
    {
      validator: (value: string) => /^\/.{0,199}$/.test(value),
      message: '必须以斜杠(/)开头，长度不能超过 200',
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
const formModel = reactive<IListenerRuleModel>({
  domain: '',
  url: '',
  scheduler: undefined,
  target_group_id: '',
  certificate: { ssl_mode: SSLMode.UNIDIRECTIONAL, ca_cloud_id: '', cert_cloud_ids: [] },
});
let originDomain: string;
watchEffect(() => {
  if (props.initialModel) {
    const { domain } = props.initialModel;
    Object.assign(formModel, { domain });
    originDomain = domain;
  }
});

const targetGroupSelectorRef = useTemplateRef<typeof TargetGroupSelector>('target-group-selector');
const svrCertSelectorRef = useTemplateRef<typeof CertSelector>('svr-cert-selector');
watch(
  model,
  (val) => {
    if (!val || props.isEdit) return;
    nextTick(() => {
      targetGroupSelectorRef.value?.handleRefresh();
    });
  },
  { immediate: true },
);

const isConfirmLoading = computed(() =>
  props.isEdit ? loadBalancerListenerStore.updateDomainLoading : loadBalancerListenerStore.createRulesLoading,
);
const handleConfirm = async () => {
  await formRef.value.validate();
  const { id, vendor, protocol, sni_switch } = props.listenerRowData;
  const { domain, url, target_group_id, scheduler, certificate } = formModel;
  const isHttpsAndSniOpen = ListenerProtocol.HTTPS === protocol && sni_switch === 1;
  if (props.isEdit) {
    await loadBalancerListenerStore.updateDomain(
      id,
      {
        domain: originDomain,
        new_domain: domain,
        certificate: isHttpsAndSniOpen ? certificate : undefined,
      },
      currentGlobalBusinessId.value,
    );
    Message({ theme: 'success', message: '编辑成功' });
  } else {
    await loadBalancerListenerStore.createRules(
      vendor,
      id,
      {
        domains: [domain],
        url,
        target_group_id,
        scheduler,
        certificate: isHttpsAndSniOpen ? certificate : undefined,
      },
      currentGlobalBusinessId.value,
    );
    Message({ theme: 'success', message: '新增成功' });
  }
  // 前端交互
  const response = props.isEdit
    ? {
        domain,
        url_count: props.initialModel.url_count,
        displayConfig: { ...props.initialModel.displayConfig, originDomain },
      }
    : { domain, url_count: 1 };
  emit('confirm-success', props.isEdit, response);
};

const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" :title="title" class="domain-form-dialog">
    <grid-details
      :fields="displayFields"
      :details="listenerRowData"
      :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
    />
    <bk-form ref="form" form-type="vertical" :model="formModel" :rules="rules">
      <bk-form-item label="域名" required property="domain">
        <bk-input v-model="formModel.domain" />
      </bk-form-item>
      <template v-if="!isEdit">
        <bk-form-item label="URL路径" required property="url">
          <bk-input v-model="formModel.url" />
        </bk-form-item>
        <bk-form-item label="均衡方式" required property="scheduler">
          <bk-select v-model="formModel.scheduler">
            <template v-for="scheduler in SCHEDULER_LIST" :key="scheduler">
              <bk-option :id="scheduler" :name="SCHEDULER_NAME[scheduler]" />
            </template>
          </bk-select>
        </bk-form-item>
        <bk-form-item label="目标组" required property="target_group_id">
          <!-- TODO-CLB: 使用roll-request进行改造 -->
          <target-group-selector
            ref="target-group-selector"
            v-model="formModel.target_group_id"
            :account-id="loadBalancerDetails.account_id"
            :cloud-vpc-id="loadBalancerDetails.cloud_vpc_id"
            :region="loadBalancerDetails.region"
            :protocol="listenerRowData.protocol"
            :is-cors-v2="loadBalancerDetails.extension.snat_pro"
          />
        </bk-form-item>
      </template>
      <template v-if="ListenerProtocol.HTTPS === listenerRowData.protocol && listenerRowData.sni_switch === 1">
        <bk-form-item label="SSL解析方式" required property="certificate.ssl_mode">
          <bk-radio-group v-model="formModel.certificate.ssl_mode">
            <bk-radio :label="SSLMode.UNIDIRECTIONAL">
              {{ SSL_MODE_NAME[SSLMode.UNIDIRECTIONAL] }}
              <bk-tag theme="info" class="ml4">推荐</bk-tag>
            </bk-radio>
            <bk-radio :label="SSLMode.MUTUAL" class="ml24">{{ SSL_MODE_NAME[SSLMode.MUTUAL] }}</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item label="服务器证书" required property="certificate.cert_cloud_ids">
          <cert-selector
            ref="svr-cert-selector"
            v-model="formModel.certificate.cert_cloud_ids"
            type="SVR"
            :account-id="loadBalancerDetails.account_id"
          />
        </bk-form-item>
        <bk-form-item
          v-if="formModel.certificate.ssl_mode === SSLMode.MUTUAL"
          label="CA证书"
          required
          property="certificate.ca_cloud_id"
        >
          <cert-selector
            v-model="formModel.certificate.ca_cloud_id"
            type="CA"
            :account-id="loadBalancerDetails.account_id"
          />
        </bk-form-item>
      </template>
    </bk-form>
    <template #footer>
      <modal-footer :loading="isConfirmLoading" @confirm="handleConfirm" @closed="handleClosed" />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss"></style>
