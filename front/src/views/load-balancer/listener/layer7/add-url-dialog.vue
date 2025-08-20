<script setup lang="ts">
import { computed, inject, nextTick, reactive, Ref, useTemplateRef, watch, watchEffect } from 'vue';
import { DisplayFieldFactory, DisplayFieldType } from '../../children/display/field-factory';
import { ModelPropertyDisplay } from '@/model/typings';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import {
  IListenerDetails,
  IListenerItem,
  IListenerRuleItem,
  IListenerRuleModel,
  useLoadBalancerListenerStore,
} from '@/store/load-balancer/listener';
import { SCHEDULER_LIST, SCHEDULER_NAME } from '../../constants';

import { Form, Message } from 'bkui-vue';
import GridDetails from '../../children/display/grid-details.vue';
import ModalFooter from '@/components/modal/modal-footer.vue';
import TargetGroupSelector from '@/views/business/load-balancer/clb-view/components/TargetGroupSelector';

interface IProps {
  listenerRowData: IListenerItem;
  loadBalancerDetails: ILoadBalancerDetails;
  isEdit?: boolean;
  domain: string;
  initialModel?: IListenerRuleModel;
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits<{ 'confirm-success': [isEdit: boolean, rule: Partial<IListenerRuleItem>] }>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();
const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const title = computed(() => (props.isEdit ? '编辑URL路径' : '新增URL路径'));

const displayProperties = [
  ...DisplayFieldFactory.createModel(DisplayFieldType.LISTENER).getProperties(),
  ...DisplayFieldFactory.createModel(DisplayFieldType.Rule).getProperties(),
];
const fieldIds = ['name', 'protocol_and_port', 'domain'];
const fieldConfig: Record<string, Partial<ModelPropertyDisplay>> = {
  protocol_and_port: {
    render: (data: IListenerDetails) => {
      const { protocol, port, end_port } = data ?? {};
      return end_port ? `${protocol}:${port}-${end_port}` : `${protocol}:${port}`;
    },
  },
};
const displayFields = fieldIds.map((id) => {
  const property = displayProperties.find((item) => item.id === id) as ModelPropertyDisplay;
  return { ...property, ...fieldConfig[id] };
});

const formRef = useTemplateRef<typeof Form>('form');
const rules = {
  url: [
    {
      validator: (value: string) => /^\/.{0,199}$/.test(value),
      message: '必须以斜杠(/)开头，长度不能超过 200',
      trigger: 'change',
    },
  ],
};
const formModel = reactive<IListenerRuleModel>({ url: '', scheduler: undefined, target_group_id: '' });
watchEffect(() => {
  if (props.initialModel) {
    Object.assign(formModel, props.initialModel);
  }
});

const targetGroupSelectorRef = useTemplateRef<typeof TargetGroupSelector>('target-group-selector');
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
  props.isEdit ? loadBalancerListenerStore.updateRuleLoading : loadBalancerListenerStore.createRulesLoading,
);
const handleConfirm = async () => {
  await formRef.value.validate();
  const { id: listenerId, vendor } = props.listenerRowData;
  const { url, target_group_id, scheduler } = formModel;
  if (props.isEdit) {
    await loadBalancerListenerStore.updateRule(
      vendor,
      listenerId,
      props.initialModel.id,
      { url, target_group_id, scheduler },
      currentGlobalBusinessId.value,
    );
    Message({ theme: 'success', message: '编辑成功' });
    emit('confirm-success', props.isEdit, { ...formModel, id: props.initialModel.id });
  } else {
    const res = await loadBalancerListenerStore.createRules(
      vendor,
      listenerId,
      { domains: [props.domain], url, target_group_id, scheduler },
      currentGlobalBusinessId.value,
    );
    Message({ theme: 'success', message: '新增成功' });
    emit('confirm-success', props.isEdit, { ...formModel, id: res.success_ids[0] });
  }
};

const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" :title="title" class="url-form-dialog">
    <grid-details
      :fields="displayFields"
      :details="{ ...listenerRowData, domain }"
      :is-loading="loadBalancerListenerStore.listenerDetailsLoading"
    />
    <bk-form ref="form" form-type="vertical" :model="formModel" :rules="rules">
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
      <bk-form-item v-if="!isEdit" label="目标组" required property="target_group_id">
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
    </bk-form>
    <template #footer>
      <modal-footer :loading="isConfirmLoading" @confirm="handleConfirm" @closed="handleClosed" />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss"></style>
