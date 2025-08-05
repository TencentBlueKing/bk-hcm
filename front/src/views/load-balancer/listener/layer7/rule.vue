<script setup lang="ts">
import { inject, onMounted, reactive, Ref, ref } from 'vue';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import {
  IListenerDomainInfoItem,
  IListenerDomainsListResponseData,
  IListenerItem,
  useLoadBalancerListenerStore,
} from '@/store/load-balancer/listener';

import { Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import RuleCollapsePanel from './rule-collapse-panel.vue';
import AddDomainDialog from './add-domain-dialog.vue';
import { IAuthSign } from '@/common/auth-service';

const props = defineProps<{
  listenerRowData: IListenerItem;
  loadBalancerDetails: ILoadBalancerDetails;
}>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();
const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');
const clbOperationAuthSign = inject<Ref<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');

const domainInfo = ref<IListenerDomainsListResponseData>();

onMounted(async () => {
  const { id, vendor } = props.listenerRowData;
  const res = await loadBalancerListenerStore.getDomainListByListenerId(vendor, id, currentGlobalBusinessId.value);
  const domainList = res.domain_list.map<IListenerDomainInfoItem>((item) => {
    return { ...item, key: item.domain };
  });
  domainInfo.value = { ...res, domain_list: domainList };
});

// domain操作
const addDomainDialogState = reactive({ isShow: false, isHidden: true, isEdit: false, initialModel: null });
const handleAddDomain = () => {
  Object.assign(addDomainDialogState, { isShow: true, isHidden: false });
};
const handleEditDomain = (initialModel: IListenerDomainInfoItem) => {
  Object.assign(addDomainDialogState, { isShow: true, isHidden: false, isEdit: true, initialModel });
};
const handleDomainFormConfirmSuccess = async (isEdit: boolean, newDomainInfo: IListenerDomainInfoItem) => {
  if (!isEdit) {
    domainInfo.value.domain_list.unshift({
      ...newDomainInfo,
      key: newDomainInfo.domain,
      displayConfig: { isNew: true },
    });
  } else {
    domainInfo.value.domain_list.forEach((item) => {
      const { originDomain } = newDomainInfo.displayConfig ?? {};
      if (item.domain === originDomain) {
        const newDomain = newDomainInfo.domain;
        Object.assign(item, { domain: newDomain, displayConfig: newDomainInfo.displayConfig });

        // 如果当变更的域名是默认域名，则需要重新记录默认域名
        if (domainInfo.value.default_domain === originDomain) {
          domainInfo.value.default_domain = newDomain;
        }
      }
    });
  }
  handleDomainFormHidden();
};
const handleDomainFormHidden = () => {
  Object.assign(addDomainDialogState, { isShow: false, isHidden: true, isEdit: false, initialModel: null });
};
const setDefaultDomainHandler = async (domain: string) => {
  await loadBalancerListenerStore.updateDomain(
    props.listenerRowData.id,
    { domain, default_server: true },
    currentGlobalBusinessId.value,
  );
  Message({ theme: 'success', message: '设置成功' });
  domainInfo.value.default_domain = domain;
};
const currentActiveDomain = ref('');
const handleRemoveDomain = async (domain: string) => {
  currentActiveDomain.value = domain;
  try {
    const { id, vendor } = props.listenerRowData;
    await loadBalancerListenerStore.batchDeleteDomain(vendor, id, { domains: [domain] }, currentGlobalBusinessId.value);
    Message({ theme: 'success', message: '删除成功' });
    const idx = domainInfo.value.domain_list.findIndex((item) => item.domain === domain);
    domainInfo.value.domain_list.splice(idx, 1);
  } finally {
    currentActiveDomain.value = '';
  }
};

// rule
const handleRuleNumChange = (domain: string, count: number) => {
  domainInfo.value.domain_list.forEach((item) => {
    if (item.domain === domain) {
      item.url_count = count;
    }
  });
};
</script>

<template>
  <div class="rule-container">
    <div class="tools">
      <hcm-auth :sign="clbOperationAuthSign" v-slot="{ noPerm }">
        <bk-button theme="primary" outline :disabled="noPerm" class="button" @click="handleAddDomain">
          <plus class="f26" />
          新增域名
        </bk-button>
      </hcm-auth>
      <span>域名数量：{{ domainInfo?.domain_list.length }}</span>
      <span>默认域名：{{ domainInfo?.default_domain }}</span>
    </div>
    <div class="panel-list">
      <rule-collapse-panel
        v-for="item in domainInfo?.domain_list"
        :key="item.key"
        :is-new="item.displayConfig?.isNew"
        :is-expand="item.displayConfig?.isExpand"
        :is-default="item.domain === domainInfo.default_domain"
        :domain-info="item"
        :listener-row-data="listenerRowData"
        :load-balancer-details="loadBalancerDetails"
        :set-default-domain-handler="setDefaultDomainHandler"
        :active="item.domain === currentActiveDomain"
        :clb-operation-auth-sign="clbOperationAuthSign"
        @edit-domain="handleEditDomain"
        @remove-domain="handleRemoveDomain"
        @rule-num-change="handleRuleNumChange"
      />
    </div>
    <template v-if="!addDomainDialogState.isHidden">
      <add-domain-dialog
        v-model="addDomainDialogState.isShow"
        :listener-row-data="listenerRowData"
        :load-balancer-details="loadBalancerDetails"
        :is-edit="addDomainDialogState.isEdit"
        :initial-model="addDomainDialogState.initialModel"
        @confirm-success="handleDomainFormConfirmSuccess"
        @hidden="handleDomainFormHidden"
      />
    </template>
  </div>
</template>

<style scoped lang="scss">
.rule-container {
  .tools {
    position: relative;
    display: flex;
    align-items: center;
    height: 36px;
    gap: 24px;
    margin-bottom: 16px;
    font-size: 12px;
    color: #313238;

    .button {
      min-width: 88px;
    }
  }

  .panel-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-height: calc(100% - 52px);
    overflow-y: auto;
  }

  .f26 {
    font-size: 26px;
  }
}
</style>
