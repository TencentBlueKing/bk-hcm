<script setup lang="ts">
import { h, inject, onMounted, reactive, Ref, ref, watch } from 'vue';
import { DisplayFieldFactory, DisplayFieldType } from '../../children/display/field-factory';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import {
  IListenerDomainInfoItem,
  IListenerItem,
  IListenerRuleItem,
  useLoadBalancerListenerStore,
} from '@/store/load-balancer/listener';
import { useLoadBalancerTargetGroupStore } from '@/store/load-balancer/target-group';
import usePage from '@/hooks/use-page';
import { transformSimpleCondition } from '@/utils/search';
import { ISearchCondition } from '@/typings';
import { ModelPropertyDisplay } from '@/model/typings';
import { ConditionKeyType, SearchConditionFactory } from '../../children/search/condition-factory';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import { MENU_BUSINESS_TARGET_GROUP_DETAILS } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { merge } from 'lodash';
import routerAction from '@/router/utils/action';
import { IAuthSign } from '@/common/auth-service';

import { Button, Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import DataList from '../../children/display/data-list.vue';
import BindingStatus from '../children/binding-status.vue';
import AddUrlDialog from './add-url-dialog.vue';
import RsPreviewDialog from './rs-preview-dialog.vue';

interface IProps {
  isNew?: boolean;
  isDefault: boolean;
  domainInfo: IListenerDomainInfoItem;
  listenerRowData: IListenerItem;
  loadBalancerDetails: ILoadBalancerDetails;
  setDefaultDomainHandler: (domain: string) => Promise<void>;
  active: boolean;
  clbOperationAuthSign: IAuthSign | IAuthSign[];
}

const props = defineProps<IProps>();
const emit = defineEmits<{
  'edit-domain': [initialModel: IListenerDomainInfoItem];
  'remove-domain': [domain: string];
  'rule-num-change': [domain: string, num: number];
}>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();
const loadBalancerTargetGroupStore = useLoadBalancerTargetGroupStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const isExpand = ref(props.domainInfo.displayConfig?.isExpand ?? false);

const jumpToTargetGroupDetails = (id: string) => {
  routerAction.open({
    name: MENU_BUSINESS_TARGET_GROUP_DETAILS,
    params: { id },
    query: {
      [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value,
      type: 'detail',
      vendor: props.listenerRowData.vendor,
    },
  });
};

const fieldIds = ['url', 'scheduler', 'target_group_id', 'rs_num', 'binding_status'];
const displayProperties = DisplayFieldFactory.createModel(DisplayFieldType.Rule).getProperties();
const fieldConfig: Record<string, Partial<ModelPropertyDisplay>> = {
  binding_status: {
    render: ({ row, cell }) => {
      return h(BindingStatus, { value: cell, protocol: row.protocol });
    },
  },
  target_group_id: {
    render: ({ cell }) => {
      return h('div', { class: 'target-group-cell' }, [
        cell,
        h(
          Button,
          { theme: 'primary', text: true, class: 'link', onClick: () => jumpToTargetGroupDetails(cell) },
          h('i', { class: 'hcm-icon bkhcm-icon-jump-fill' }),
        ),
      ]);
    },
  },
};
const displayFields = fieldIds.map((id) => {
  const property = displayProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return { ...property, ...fieldConfig[id] };
});
const conditionProperties = SearchConditionFactory.createModel(ConditionKeyType.URL).getProperties();
const rowClassCallback = (row: IListenerRuleItem) => {
  const { displayConfig } = row ?? {};
  if (displayConfig?.isNew) return 'is-new';
};

const urlList = ref<IListenerRuleItem[]>([]);
const condition: ISearchCondition = { domain: props.domainInfo.domain };

const { pagination, getPageParams } = usePage(false);
const sort = ref('created_at');
const order = ref('DESC');

const setRsNum = async (list: IListenerRuleItem[] = []) => {
  if (!list.length) return;

  const targetsWeightStatList = await loadBalancerTargetGroupStore.getTargetsWeightStat(
    list.map((item) => item.target_group_id),
    currentGlobalBusinessId.value,
  );

  urlList.value.forEach((item) => {
    const targetsWeightStat = targetsWeightStatList.find((i) => i.target_group_id === item.target_group_id);
    Object.assign(item, { rs_num: targetsWeightStat.rs_weight_non_zero_num + targetsWeightStat.rs_weight_zero_num });
  });
};

const setBindingStatus = async (list: IListenerRuleItem[] = []) => {
  if (!list.length) return;

  const requestIds = list.filter((item) => item.binding_status !== 'success').map((i) => i.id);
  if (!requestIds.length) return;

  const { id, vendor } = props.listenerRowData;
  const statusList = await loadBalancerListenerStore.getRulesBindingStatusList(
    vendor,
    id,
    { rule_ids: requestIds },
    currentGlobalBusinessId.value,
  );

  urlList.value.forEach((item) => {
    const status = statusList.find((i) => i.rule_id === item.id);
    if (status) {
      Object.assign(item, { binding_status: status.binding_status });
    }
  });

  if (statusList.every((item) => item.binding_status === 'success')) {
    taskPoll.pause();
  }
};

const loading = ref(false);
onMounted(async () => {
  const { id, vendor } = props.listenerRowData;
  try {
    const { list, count } = await loadBalancerListenerStore.getRuleListByListenerId(
      vendor,
      id,
      {
        filter: transformSimpleCondition(condition, conditionProperties),
        page: getPageParams(pagination, { sort: sort.value, order: order.value }),
      },
      currentGlobalBusinessId.value,
    );

    // 为了方便轮询，这里统一将binding_status设置为'binding'态
    urlList.value = list.map((item) => ({ ...item, binding_status: 'binding' }));
    pagination.count = count;

    setRsNum(urlList.value);
    setBindingStatus(urlList.value);
  } finally {
    loading.value = false;
  }
});

const isScrollLoading = ref(false);
const handleScrollBottom = async () => {
  if (isScrollLoading.value) return;
  if (pagination.count <= urlList.value.length) return;

  pagination.current += 1;
  isScrollLoading.value = true;
  try {
    const { id, vendor } = props.listenerRowData;
    const { list, count } = await loadBalancerListenerStore.getRuleListByListenerId(
      vendor,
      id,
      {
        filter: transformSimpleCondition(condition, conditionProperties),
        page: getPageParams(pagination, { sort: sort.value, order: order.value }),
      },
      currentGlobalBusinessId.value,
    );

    urlList.value.push(...list.map((item) => ({ ...item, binding_status: 'binding' })));
    pagination.count = count;
  } finally {
    isScrollLoading.value = false;
  }
};

const taskPoll = useTimeoutPoll(
  () => {
    setBindingStatus(urlList.value);
  },
  30000,
  { max: 10 },
);

watch(isExpand, (val) => {
  if (val && urlList.value.some((item) => item.binding_status !== 'success')) {
    taskPoll.resume();
  } else {
    taskPoll.pause();
  }
});

const setDefaultDomainLoading = ref(false);
const handleSetDefaultDomain = async () => {
  setDefaultDomainLoading.value = true;
  try {
    await props.setDefaultDomainHandler(props.domainInfo.domain);
  } finally {
    setDefaultDomainLoading.value = false;
  }
};

const handleEditDomain = () => {
  emit('edit-domain', merge({}, props.domainInfo, { config: { isExpand: isExpand.value } }));
};
const handleRemoveDomain = () => {
  emit('remove-domain', props.domainInfo.domain);
};

// url
const addUrlDialogState = reactive({ isShow: false, isHidden: true, isEdit: false, initialModel: null });
const handleAddUrl = () => {
  Object.assign(addUrlDialogState, { isShow: true, isHidden: false });
};
const handleEditUrl = (row: IListenerRuleItem) => {
  Object.assign(addUrlDialogState, { isShow: true, isHidden: false, isEdit: true, initialModel: row });
};
const handleUrlFormConfirmSuccess = (isEdit: boolean, rule: Partial<IListenerRuleItem>) => {
  if (!isEdit) {
    urlList.value.unshift({
      ...rule,
      rs_num: 0,
      binding_status: 'binding',
      displayConfig: { isNew: true },
    } as IListenerRuleItem);
    emit('rule-num-change', props.domainInfo.domain, urlList.value.length);
  } else {
    urlList.value.forEach((item) => {
      if (item.id === rule.id) {
        Object.assign(item, { ...rule });
      }
    });
  }
  handleUrlFormHidden();
};
const handleRemoveUrl = async (ruleId: string) => {
  const { id, vendor } = props.listenerRowData;
  await loadBalancerListenerStore.batchDeleteRule(vendor, id, { rule_ids: [ruleId] }, currentGlobalBusinessId.value);
  Message({ theme: 'success', message: '删除成功' });
  const idx = urlList.value.findIndex((item) => item.id === ruleId);
  urlList.value.splice(idx, 1);
  emit('rule-num-change', props.domainInfo.domain, urlList.value.length);
};
const handleUrlFormHidden = () => {
  Object.assign(addUrlDialogState, { isShow: false, isHidden: true, isEdit: false, initialModel: null });
};

// rs-preview
const rsPreviewDialogState = reactive({ isShow: false, isHidden: true, targetGroupId: '' });
const handleShowRsPreview = async (row: IListenerRuleItem) => {
  Object.assign(rsPreviewDialogState, { isShow: true, isHidden: false, targetGroupId: row.target_group_id });
};
const handleRsPreviewDialogHidden = () => {
  Object.assign(rsPreviewDialogState, { isShow: false, isHidden: true, targetGroupId: '' });
};
</script>

<template>
  <!-- !：域名相关操作，请求成功后直接更新响应式数据，不重新请求接口 -->
  <!-- !：URL相关操作，请求成功后直接更新响应式数据，如果涉及新增，先操作响应式数据，再开启轮询查RS数量、同步状态  -->
  <bk-collapse class="rule-collapse-panel" :class="{ 'is-new': isNew }">
    <bk-collapse-panel v-model="isExpand" icon="right-shape" alone>
      <template #default>
        <span class="text-light">{{ domainInfo.domain }}</span>
        <bk-tag v-if="isNew" theme="success" type="filled" size="small">new</bk-tag>
        <bk-loading v-if="setDefaultDomainLoading" size="mini" mode="spin" theme="primary" loading></bk-loading>
        <template v-else>
          <bk-tag v-if="isDefault" class="default-tag">默认</bk-tag>
          <hcm-auth v-else :sign="clbOperationAuthSign" v-slot="{ noPerm }">
            <bk-button
              theme="primary"
              text
              :disabled="noPerm"
              @click.stop="handleSetDefaultDomain"
              class="set-default-btn"
            >
              设为默认
            </bk-button>
          </hcm-auth>
        </template>
        <span class="ml-auto text-light">URL数量：{{ domainInfo.url_count }}</span>
        <hcm-auth :sign="clbOperationAuthSign" @click.stop v-slot="{ noPerm }">
          <bk-button text :disabled="noPerm" @click="handleEditDomain">
            <i class="hcm-icon bkhcm-icon-bianji operation"></i>
          </bk-button>
        </hcm-auth>
        <hcm-auth :sign="clbOperationAuthSign" @click.stop v-slot="{ noPerm }">
          <bk-pop-confirm content="确认删除该域名？" trigger="click" @confirm="handleRemoveDomain">
            <bk-button
              text
              :loading="active && loadBalancerListenerStore.batchDeleteDomainLoading"
              :disabled="noPerm || isDefault"
              v-bk-tooltips="{ content: '默认域名不可删除', disabled: !isDefault }"
            >
              <i class="hcm-icon bkhcm-icon-delete operation"></i>
            </bk-button>
          </bk-pop-confirm>
        </hcm-auth>
      </template>
      <template #content>
        <data-list
          class="data-list"
          v-bkloading="{ loading }"
          :columns="displayFields"
          :list="urlList"
          :enable-query="false"
          :max-height="360"
          :row-class="rowClassCallback"
          :scroll-loading="isScrollLoading"
          @scroll-bottom="handleScrollBottom"
        >
          <template #action>
            <bk-table-column label="操作" :show-overflow-tooltip="false" :width="180" fixed="right">
              <template #default="{ row }">
                <div class="action-cell">
                  <bk-button theme="primary" text @click="handleShowRsPreview(row)">预览RS信息</bk-button>
                  <hcm-auth :sign="clbOperationAuthSign" v-slot="{ noPerm }">
                    <bk-button theme="primary" text :disabled="noPerm" @click="handleEditUrl(row)">编辑</bk-button>
                  </hcm-auth>
                  <hcm-auth :sign="clbOperationAuthSign" v-slot="{ noPerm }">
                    <bk-pop-confirm content="确认删除该URL路径？" trigger="click" @confirm="handleRemoveUrl(row.id)">
                      <bk-button
                        theme="primary"
                        text
                        :loading="loadBalancerListenerStore.batchDeleteRuleLoading"
                        :disabled="noPerm"
                      >
                        删除
                      </bk-button>
                    </bk-pop-confirm>
                  </hcm-auth>
                </div>
              </template>
            </bk-table-column>
          </template>
        </data-list>
        <div class="fixed-bottom">
          <hcm-auth :sign="clbOperationAuthSign" v-slot="{ noPerm }">
            <bk-button theme="primary" text :disabled="noPerm" @click="handleAddUrl">
              <plus class="f26" />
              新增URL路径
            </bk-button>
          </hcm-auth>
        </div>
      </template>
    </bk-collapse-panel>
  </bk-collapse>

  <template v-if="!addUrlDialogState.isHidden">
    <add-url-dialog
      v-model="addUrlDialogState.isShow"
      :listener-row-data="listenerRowData"
      :load-balancer-details="loadBalancerDetails"
      :is-edit="addUrlDialogState.isEdit"
      :domain="props.domainInfo.domain"
      :initial-model="addUrlDialogState.initialModel"
      @confirm-success="handleUrlFormConfirmSuccess"
      @hidden="handleUrlFormHidden"
    />
  </template>

  <template v-if="!rsPreviewDialogState.isHidden">
    <rs-preview-dialog
      v-model="rsPreviewDialogState.isShow"
      :target-group-id="rsPreviewDialogState.targetGroupId"
      @hidden="handleRsPreviewDialogHidden"
    />
  </template>
</template>

<style scoped lang="scss">
.rule-collapse-panel {
  :deep(.bk-collapse-item) {
    border: 1px solid #dcdee5;
    border-radius: 2px;

    .bk-collapse-header {
      height: 36px;
      line-height: 36px;
      display: flex;
      align-items: center;
      background: #f0f1f5;

      .bk-collapse-title {
        flex: 1;
        margin-left: 20px;
        display: flex;
        align-items: center;
        gap: 16px;
        font-size: 12px;
      }

      .bk-collapse-icon svg {
        font-size: 12px;
      }

      .set-default-btn {
        opacity: 0;
      }

      .default-tag {
        background: #dae9fd;
        color: #4193e5;
      }

      &:hover {
        .set-default-btn {
          opacity: 1;
        }
      }
    }

    .bk-collapse-content {
      padding: 0;

      .fixed-bottom {
        display: flex;
        align-items: center;
        height: 36px;
      }
    }
  }

  &.is-new {
    :deep(.bk-collapse-item) {
      .bk-collapse-header {
        background: #ebfaf0;
      }
    }
  }

  .data-list {
    .action-cell {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    :deep(.is-new td) {
      background: #ebfaf0 !important;
    }

    :deep(.target-group-cell) {
      display: inline-flex;
      align-items: center;
      gap: 4px;

      .link {
        opacity: 0;
      }
    }

    :deep(tr) {
      &:hover {
        .link {
          opacity: 1;
        }
      }
    }
  }

  .ml-auto {
    margin-left: auto;
  }

  .f26 {
    font-size: 26px;
  }
}
</style>
