<script setup lang="ts">
import { computed, inject, reactive, Ref, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { cloneDeep } from 'lodash';
import routerAction from '@/router/utils/action';
import { Message } from 'bkui-vue';
import ModalFooter from '@/components/modal/modal-footer.vue';
import { VendorEnum, ResourceTypeEnum } from '@/common/constant';
import { RsDeviceType } from '@/views/load-balancer/constants';
import RsIpGroup from '../children/rs-ip-group.vue';
import { useLoadBalancerRsStore } from '@/store/load-balancer/rs';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';

interface IProps {
  selections: any[];
  vendor: VendorEnum;
  type: RsDeviceType;
}

const model = defineModel<boolean>();

const props = defineProps<IProps>();

const { t } = useI18n();

const info: Partial<
  Record<
    RsDeviceType,
    {
      title: string;
      confirm: string;
      checkText: string;
    }
  >
> = {
  [RsDeviceType.ADJUST]: {
    title: t('批量调整 RS 权重'),
    confirm: t('确认并提交'),
    checkText: t('我已确认所选的RS IP正确，并确认批量调整权重'),
  },
  [RsDeviceType.UNBIND]: {
    title: t('批量解绑 RS'),
    confirm: t('确认并解绑'),
    checkText: t('我已确认所选的RS IP正确，并确认批量为其解绑RS'),
  },
};

const loadBalancerRsStore = useLoadBalancerRsStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');

const formRef = ref(null);
const list = ref(cloneDeep(props.selections));
const confirmCheck = ref(false);
const formData = reactive<{ new_weight: number }>({ new_weight: '' as unknown as number });

const IPLength = computed(() => list.value.length);
const rsLength = computed(() =>
  list.value.reduce((prev, cur) => {
    return prev + cur.targets.length;
  }, 0),
);

const handleConfirm = async () => {
  await formRef.value?.validate();
  if (props.type === RsDeviceType.ADJUST) return handleBatchAdjust();
  return handleBatchUnbind();
};
const handleBatchUnbind = async () => {
  const res = await loadBalancerRsStore.batchUnbind(
    {
      target_ids: list.value.reduce((prev, cur) => {
        cur.targets.forEach((item: { id: string }) => prev.push(item.id));
        return prev;
      }, []),
      account_id: list.value[0].targets[0].account_id,
    },
    currentGlobalBusinessId.value,
  );
  handleRes(res, '删除成功');
};
const handleBatchAdjust = async () => {
  const res = await loadBalancerRsStore.batchUpdateWeight(
    {
      target_ids: list.value.reduce((prev, cur) => {
        cur.targets.forEach((item: { id: string }) => prev.push(item.id));
        return prev;
      }, []),
      account_id: list.value[0].targets[0].account_id,
      new_weight: formData.new_weight,
    },
    currentGlobalBusinessId.value,
  );
  handleRes(res, '更新成功');
};
const handleRes = (res: any, message: string) => {
  if (res.data?.task_management_id) {
    routerAction.open({
      name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
      query: { bizs: currentGlobalBusinessId.value },
      params: { resourceType: ResourceTypeEnum.CLB, id: res.data.task_management_id },
    });
  } else {
    Message({ theme: 'success', message: t(message) });
  }
  handleClosed();
};
const handleDelete = (rowKey: string) => {
  const index = list.value.findIndex((item) => item.rowKey === rowKey);
  list.value.splice(index, 1);
};
const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" :title="info[type].title" width="80vw" class="batch-rs-operation-dialog">
    <div v-if="type === RsDeviceType.ADJUST">
      <bk-alert
        theme="info"
        :title="t('此操作为将多个监听器中已绑定RS的权重，批量调整为同一个权重，请确认下面表格中所选的监听器是否正确')"
      />
      <bk-form ref="formRef" class="form" form-type="vertical" :model="formData">
        <bk-form-item label="RS 权重批量调整为：" property="new_weight" required>
          <bk-input v-model.number="formData.new_weight" placeholder="请输入" class="rs-weight-input" />
        </bk-form-item>
      </bk-form>
    </div>
    <div class="count">
      已选
      <span class="ip-count">{{ IPLength }}</span>
      个IP, 共
      <span class="rs-count">{{ rsLength }}</span>
      个RS
    </div>
    <rs-ip-group :rs-list="list" :vendor="vendor" :type="type" class="rs-list" @delete="handleDelete" />
    <template #footer>
      <div class="confirm-check">
        <bk-checkbox v-model="confirmCheck">
          {{ info[type].checkText }}
        </bk-checkbox>
      </div>
      <modal-footer
        :confirm-text="info[type].confirm"
        :loading="loadBalancerRsStore.batchUpdateWeightLoading || loadBalancerRsStore.batchUnbindLoading"
        :disabled="!confirmCheck || !rsLength"
        @confirm="handleConfirm"
        @closed="handleClosed"
      />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.batch-rs-operation-dialog {
  .confirm-check {
    float: left;
    line-height: 32px;
  }

  .form {
    margin: 24px 0;

    .rs-weight-input {
      width: 400px;
    }
  }

  .rs-list {
    height: 400px;
    overflow-y: auto;
  }

  .count {
    font-weight: 700;
    margin-bottom: 16px;

    .ip-count {
      color: #3a84ff;
    }

    .rs-count {
      color: #f59500;
    }
  }
}
</style>
