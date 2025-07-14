<template>
  <div class="target-group-container">
    <div class="flex-row align-items-center">
      <span>目标组：</span>
      <bk-loading v-if="isListenerLoading" size="mini" mode="spin" theme="primary" loading></bk-loading>
      <template v-else>
        <template v-if="listener?.target_group_id">
          <span class="link-text-btn" @click="gotoTargetGroupDetail">
            {{ listener.target_group_name }}
          </span>
          <bk-loading v-if="isTargetGroupBinding" size="mini" mode="spin" theme="primary" loading></bk-loading>
          <copy-to-clipboard class="copy-btn ml4" :content="listener.target_group_name" />
        </template>
        <span v-else>--</span>
      </template>
    </div>
    <rs-config-table only-show :rs-list="rsList" :loading="isRsListLoading" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import routerAction from '@/router/utils/action';
import { useAccountStore, useBusinessStore } from '@/store';
import { LBRouteName, ListenerPanelEnum } from '@/constants';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { QueryRuleOPEnum } from '@/typings';

import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import RsConfigTable from '../../../group-view/components/RsConfigTable';

const props = defineProps<{ id: string; type: string }>();
const accountStore = useAccountStore();
const businessStore = useBusinessStore();

const listener = ref();
const isListenerLoading = ref(false);
const getListenerDetail = async (id: string) => {
  isListenerLoading.value = true;
  try {
    const res = await businessStore.detail('listeners', id);
    listener.value = res.data;
  } finally {
    isListenerLoading.value = false;
  }
};

const isTargetGroupBinding = ref(false);
const getTargetGroupBindingStatus = async (listenerId: string, loadBalancerId: string) => {
  // 判断目标组是否在绑定中
  const listRes = await businessStore.list(
    {
      filter: { op: QueryRuleOPEnum.AND, rules: [{ field: 'id', op: QueryRuleOPEnum.EQ, value: listenerId }] },
      page: { count: false, start: 0, limit: 1 },
    },
    `load_balancers/${loadBalancerId}/listeners`,
  );
  isTargetGroupBinding.value = listRes.data.details[0].binding_status === 'binding';
};

const rsList = ref([]);
const isRsListLoading = ref(false);
const getRsList = async (targetGroupId: string) => {
  isRsListLoading.value = true;
  try {
    const res = await businessStore.getTargetGroupDetail(targetGroupId);
    res.data.target_list = res.data.target_list.map((item: any) => {
      item.region = item.zone.slice(0, item.zone.lastIndexOf('-'));
      return item;
    });
    rsList.value = res.data.target_list;
  } finally {
    isRsListLoading.value = false;
  }
};

watch(
  [() => props.id, () => props.type],
  async ([id, type]) => {
    if (id && type === ListenerPanelEnum.TARGET_GROUP) {
      await getListenerDetail(props.id);
      await getTargetGroupBindingStatus(props.id, listener.value.lb_id);
      await getRsList(listener.value.target_group_id);
    }
  },
  { immediate: true },
);

const gotoTargetGroupDetail = () => {
  routerAction.open({
    name: LBRouteName.tg,
    query: {
      [GLOBAL_BIZS_KEY]: accountStore.bizs,
      type: 'detail',
      vendor: listener.value.vendor,
    },
    params: {
      id: listener.value.target_group_id,
    },
  });
};
</script>
