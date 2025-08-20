<script setup lang="ts">
import { computed, inject, onMounted, Ref, ref } from 'vue';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import { IListenerDetails, IListenerItem, useLoadBalancerListenerStore } from '@/store/load-balancer/listener';
import { LAYER_7_LISTENER_PROTOCOL } from '../constants';
import { DisplayFieldFactory, DisplayFieldType } from '../children/display/field-factory';
import { ModelPropertyDisplay } from '@/model/typings';

import Info from './children/info.vue';
import Rule from './layer7/rule.vue';
import TargetGroup from './layer4/target-group.vue';
import GridDetails from '../children/display/grid-details.vue';

const model = defineModel<boolean>();
const props = defineProps<{ rowData: IListenerItem; loadBalancerDetails: ILoadBalancerDetails }>();
const emit = defineEmits<{ 'update-success': [id: string] }>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();
const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const isLayer7 = computed(() => LAYER_7_LISTENER_PROTOCOL.includes(props.rowData.protocol));

const panels = computed(() => {
  const info = { name: 'info', label: '基本信息', component: Info };
  if (isLayer7.value) {
    return [{ name: 'rule', label: '规则信息', component: Rule }, info];
  }
  return [{ name: 'rule', label: '目标组', component: TargetGroup }, info];
});
const active = ref(panels.value[0].name);

const details = ref<IListenerDetails>();
const getListenerDetails = async () => {
  details.value = await loadBalancerListenerStore.getListenerDetails(props.rowData.id, currentGlobalBusinessId.value);
};

const displayProperties = DisplayFieldFactory.createModel(DisplayFieldType.LISTENER).getProperties();
const fieldIds = ['name', 'cloud_id', 'protocol_and_port', 'scheduler'];
const fieldConfig: Record<string, Partial<ModelPropertyDisplay>> = {
  name: { meta: { display: { showOverflowTooltip: true } } },
  protocol_and_port: {
    render: (data: IListenerItem) => {
      const { protocol, port, end_port } = data ?? {};
      return end_port ? `${protocol}:${port}-${end_port}` : `${protocol}:${port}`;
    },
  },
};
const displayFields = fieldIds.map((id) => {
  const property = displayProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return { ...property, ...fieldConfig[id] };
});

const handleUpdateSuccess = (id: string) => {
  getListenerDetails();
  emit('update-success', id);
};

onMounted(() => {
  getListenerDetails();
});
</script>

<template>
  <bk-sideslider v-model:is-show="model" :width="960" class="listener-details-sideslider">
    <template #header>
      监听器详情
      <span class="name">{{ details?.name ?? rowData.name }}</span>
    </template>
    <grid-details
      class="overview-container"
      :fields="displayFields"
      :details="details"
      :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
      layout="vertical"
      :column="5"
      :gap="[0, 24]"
      :content-min-width="120"
      :content-max-width="180"
    />
    <bk-tab v-model:active="active" type="card-grid">
      <bk-tab-panel v-for="item in panels" :key="item.name" :label="item.label" :name="item.name">
        <component
          :is="item.component"
          :listener-row-data="rowData"
          :listener-details="details"
          :load-balancer-details="loadBalancerDetails"
          @update-success="handleUpdateSuccess"
        />
      </bk-tab-panel>
    </bk-tab>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.listener-details-sideslider {
  .name {
    font-size: 14px;
    color: #4d4f56;

    &::before {
      content: '-';
      margin: 0 4px;
    }
  }

  .overview-container {
    padding: 20px 40px;
    background: #f5f7fa;

    :deep(.item-label) {
      color: #979ba5;
    }
  }

  :deep(.bk-tab-header) {
    padding: 0 40px;
    background: #f5f7fa;
  }

  :deep(.bk-tab-content) {
    padding: 24px 40px;
    box-shadow: none;
  }
}
</style>
