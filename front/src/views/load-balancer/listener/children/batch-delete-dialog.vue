<script setup lang="ts">
import { computed, h, inject, Ref, ref } from 'vue';
import { IListenerItem, useLoadBalancerListenerStore } from '@/store/load-balancer/listener';
import { ThemeEnum } from 'bkui-vue/lib/shared';
import { DisplayFieldFactory, DisplayFieldType } from '../../children/display/field-factory';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import { cloneDeep } from 'lodash';

import { Message } from 'bkui-vue';
import DataList from '../../children/display/data-list.vue';
import ModalFooter from '@/components/modal/modal-footer.vue';

interface IProps {
  selections: IListenerItem[];
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits<{ 'confirm-success': [] }>();

const loadBalancerListenerStore = useLoadBalancerListenerStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const displayFieldProperties = DisplayFieldFactory.createModel(DisplayFieldType.LISTENER).getProperties();
const commonDisplayFieldIds = ['name', 'protocol', 'lb_vip', 'lb_cloud_id', 'port', 'scheduler'];
const extensionDisplayFieldIds = ['domain_num', 'url_num', 'rs_num'];
const extensionDisplayFieldIds2 = ['rs_weight_stat'];
const displayFieldsIds = computed(() =>
  active.value
    ? [...commonDisplayFieldIds, ...extensionDisplayFieldIds]
    : [...commonDisplayFieldIds, ...extensionDisplayFieldIds2],
);
const displayFieldConfig: Record<string, Partial<ModelPropertyColumn>> = {
  rs_weight_stat: {
    render: ({ row }) => {
      const { rs_num, non_zero_weight_count } = row as IListenerItem;
      return h('div', { class: 'rs-weight-stat' }, [
        h('span', { class: 'non-zero-weight-count' }, non_zero_weight_count),
        h('span', { class: 'ml4 mr4' }, '/'),
        h('span', rs_num),
      ]);
    },
  },
};
const datalistColumns = computed(() => {
  return displayFieldsIds.value.map((id) => {
    const property = displayFieldProperties.find((item) => item.id === id);
    return { ...property, ...displayFieldConfig[id] };
  });
});

const list = ref(cloneDeep(props.selections));
const { pagination } = usePage(false);

const canDeletePredicate = (item: IListenerItem) => item.zero_weight_count === item.rs_num;

const active = ref(props.selections.every(canDeletePredicate));

const displayList = computed(() => {
  if (active.value) {
    return list.value.filter(canDeletePredicate);
  }
  return list.value.filter((item) => item.zero_weight_count !== item.rs_num);
});
const hasNonZeroWeightCount = computed(() => list.value.filter((item) => item.non_zero_weight_count !== 0).length);
const canDeleteCount = computed(() => list.value.filter(canDeletePredicate).length);

const handleSingleDelete = (row: IListenerItem) => {
  const idx = list.value.findIndex((item) => item.id === row.id);
  if (idx > -1) {
    list.value.splice(idx, 1);
  }
};

const handleConfirm = async () => {
  await loadBalancerListenerStore.batchDeleteListener(
    { ids: list.value.filter(canDeletePredicate).map((item) => item.id) },
    currentGlobalBusinessId.value,
  );
  Message({ theme: 'success', message: '删除成功' });
  handleClosed();
  emit('confirm-success');
};

const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" title="批量删除监听器" width="80vw" class="batch-delete-listener-dialog">
    <bk-alert
      class="mb16"
      theme="danger"
      title="权重不为0的监听器不允许删除，监听器删除后不可恢复，请确认好所需删除的监听器，谨慎操作"
    />
    <div class="mb16">
      已选择
      <span class="text-primary font-bold">{{ list.length }}</span>
      个监听器，其中可删除
      <span class="text-success font-bold">{{ canDeleteCount }}</span>
      个，不可删除
      <span class="text-danger font-bold">{{ hasNonZeroWeightCount }}</span>
      个。
    </div>
    <div class="toolbar">
      <bk-radio-group v-model="active">
        <bk-radio-button :label="true">可删除</bk-radio-button>
        <bk-radio-button :label="false">不可删除</bk-radio-button>
      </bk-radio-group>
    </div>
    <data-list
      class="data-list"
      :columns="datalistColumns"
      :list="displayList"
      :enable-query="false"
      :pagination="pagination"
      :remote-pagination="false"
      :max-height="500"
    >
      <template #action v-if="active">
        <bk-table-column label="">
          <template #default="{ row }">
            <bk-button text class="single-delete-btn" @click="handleSingleDelete(row)">
              <i class="hcm-icon bkhcm-icon-minus-circle-shape"></i>
            </bk-button>
          </template>
        </bk-table-column>
      </template>
    </data-list>
    <template #footer>
      <modal-footer
        confirm-text="删除"
        :confirm-button-theme="ThemeEnum.DANGER"
        :loading="loadBalancerListenerStore.batchDeleteListenerLoading"
        :disabled="canDeleteCount === 0"
        @confirm="handleConfirm"
        @closed="handleClosed"
      />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.batch-delete-listener-dialog {
  .toolbar {
    margin-bottom: 12px;
    display: flex;
    align-items: center;
  }

  .data-list {
    :deep(.lb-type-tag) {
      &.is-open {
        background-color: #d8edd9;
      }

      &.is-internal {
        background-color: #fff2c9;
      }
    }

    .single-delete-btn {
      color: #c4c6cc;
    }

    :deep(.rs-weight-stat) {
      display: inline-flex;
      align-items: center;
      line-height: normal;

      .non-zero-weight-count {
        padding: 1px 4px;
        background: #ffebeb;
        color: #e71818;
      }
    }
  }
}
</style>
