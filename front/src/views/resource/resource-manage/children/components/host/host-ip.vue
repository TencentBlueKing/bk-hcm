<script lang="ts" setup>
import { ref, h, watch, inject, computed, withDirectives } from 'vue';
import { bkTooltips, Button, Message } from 'bkui-vue';
import { useResourceStore } from '@/store/resource';
import useQueryList from '../../../hooks/use-query-list';
import bus from '@/common/bus';

const props = defineProps({
  data: {
    type: Object,
  },
  isBindBusiness: {
    type: [Boolean, String],
  },
});

const resourceStore = useResourceStore();
const isResourcePage: any = inject('isResourcePage');
const authVerifyData: any = inject('authVerifyData');

const actionName = computed(() => {
  // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});

// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

// 状态
const showChangeIP = ref(false);
const showUnbind = ref(false);
const showBind = ref(false);
const unbindData = ref();
const isBinding = ref(false);
const inUnbinding = ref(false);
const networklist = ref([]);
const bindData = ref<any>({});
const needNetwork = ref(false);

const columns = ref([
  {
    label: 'ID',
    field: 'id',
  },
  {
    label: '名称',
    field: 'name',
  },
  {
    label: 'IP地址',
    field: 'public_ip',
  },
  {
    label: 'IP类型',
    field: 'address_type',
  },
  {
    label: '备注',
    field: 'memo',
  },
  {
    label: '操作',
    render({ data }: any) {
      return [
        // h(
        //   Button,
        //   {
        //     text: true,
        //     theme: 'primary',
        //     class: 'mr10',
        //     onClick() {
        //       handleToggleShowAdjustNetwork();
        //     },
        //   },
        //   [
        //     '调整带宽',
        //   ],
        // ),
        // h(
        //   Button,
        //   {
        //     text: true,
        //     theme: 'primary',
        //     class: 'mr10',
        //     onClick() {
        //       handleToggleShowChangeIP();
        //     },
        //   },
        //   [
        //     '更换IP',
        //   ],
        // ),
        h(
          'span',
          {
            onClick() {
              showAuthDialog(actionName.value);
            },
          },
          [
            withDirectives(
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  disabled:
                    !authVerifyData.value?.permissionAction[actionName.value] ||
                    (isResourcePage.value && props.data?.bk_biz_id !== -1),
                  onClick() {
                    handleToggleShowUnbind(data);
                  },
                },
                ['解绑'],
              ),
              [[bkTooltips, generateTooltipsOptions()]],
            ),
          ],
        ),
      ];
    },
  },
]);

// 当前主机下的eip资源
const {
  datas,
  isLoading,
  triggerApi: triggerCvmEipApi,
} = useQueryList({}, 'eip', () => {
  return Promise.all([resourceStore.getEipListByCvmId(props.data.vendor, props.data.id)]);
});

const rules = [
  {
    field: 'vendor',
    op: 'eq',
    value: props.data.vendor,
  },
  {
    field: 'region',
    op: 'eq',
    value: props.data.region,
  },
  {
    field: 'account_id',
    op: 'eq',
    value: props.data.account_id,
  },
];

if (props.data.vendor === 'azure') {
  rules.push({
    field: 'extension.resource_group_name',
    op: 'json_eq',
    value: props.data.resource_group_name,
  });
}

// 当前 vendor 下的eip资源
const {
  datas: eipList,
  pagination,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
  triggerApi: triggerEipApi,
} = useQueryList(
  {
    filter: {
      op: 'and',
      rules,
    },
  },
  'eips',
  null,
  'getUnbindCvmEips',
);

const updateList = () => {
  return Promise.all([triggerCvmEipApi(), triggerEipApi(), needNetwork.value ? getNetWorkList() : Promise.resolve()]);
};

const handleToggleShowChangeIP = () => {
  showChangeIP.value = !showChangeIP.value;
};

const handleConfirmChangeIP = () => {
  handleToggleShowChangeIP();
};

const handleToggleShowUnbind = (data?: any) => {
  unbindData.value = data;
  if (!showUnbind.value) {
    resourceStore.detail('eips', unbindData.value.id).then(({ data }: any) => {
      unbindData.value.instance_id = data.instance_id;
    });
    showUnbind.value = true;
  } else {
    showUnbind.value = false;
  }
};

const handleConfirmUnbind = () => {
  const postData: any = {
    eip_id: unbindData.value.id,
  };
  if (['gcp', 'azure'].includes(unbindData.value.vendor)) {
    postData.network_interface_id = unbindData.value.instance_id;
  }
  inUnbinding.value = true;
  resourceStore
    .disassociateEip(postData)
    .then(() => {
      handleToggleShowUnbind();
      return updateList();
    })
    .finally(() => {
      inUnbinding.value = false;
    });
};

const handleToggleShowBind = (value: boolean) => {
  bindData.value = {};
  showBind.value = value;
};

const handleConfirmBind = () => {
  if (!bindData.value.eip_id) {
    Message({
      theme: 'error',
      message: '请先选择 EIP',
    });
    return;
  }
  const postData: any = {
    eip_id: bindData.value.eip_id,
    cvm_id: props.data.id,
  };
  if (needNetwork.value) {
    if (!bindData.value.network_interface_id) {
      Message({
        theme: 'error',
        message: '请先选择网络接口',
      });
      return;
    }
    postData.network_interface_id = bindData.value.network_interface_id;
  }
  isBinding.value = true;
  resourceStore
    .associateEip(postData)
    .then(() => {
      handleToggleShowBind(false);
      return updateList();
    })
    .finally(() => {
      isBinding.value = false;
    });
};

const getNetWorkList = async () => {
  resourceStore.getNetworkList(props.data.vendor, props.data.id).then((res: any) => {
    networklist.value = res.data;
  });
};

const generateTooltipsOptions = () => {
  if (!authVerifyData.value?.permissionAction?.[actionName.value])
    return {
      content: '当前用户无权限操作该按钮',
      disabled: authVerifyData.value?.permissionAction?.[actionName.value],
    };
  if (isResourcePage.value && props.data?.bk_biz_id !== -1)
    return {
      content: '该主机已分配到业务，仅可在业务下操作',
      disabled: props.data.bk_biz_id === -1,
    };

  return {
    disabled: true,
  };
};

watch(
  () => props.data,
  () => {
    if (props.data.vendor === 'tcloud') {
      columns.value.splice(
        4,
        0,
        ...[
          {
            label: '计费模式',
            field: 'internet_charge_type',
          },
          {
            label: '带宽上限',
            field: 'bandwidth',
          },
        ],
      );
    }
    needNetwork.value = !['tcloud', 'aws'].includes(props.data.vendor);
    if (needNetwork.value) {
      getNetWorkList();
    }
  },
  {
    deep: true,
    immediate: true,
  },
);
</script>

<template>
  <bk-loading :loading="isLoading">
    <span @click="showAuthDialog(actionName)">
      <bk-button
        class="btn"
        theme="primary"
        :disabled="isBindBusiness || !authVerifyData?.permissionAction[actionName]"
        @click="handleToggleShowBind(true)"
      >
        绑定
      </bk-button>
    </span>
    <bk-table class="mt16" row-hover="auto" :columns="columns" :data="datas" show-overflow-tooltip />
  </bk-loading>
  <!-- <bk-dialog
    :is-show="showAdjustNetwork"
    title="调整网络"
    theme="primary"
    quick-close
    dialog-type="show"
    @closed="handleToggleShowAdjustNetwork"
    @confirm="handleConfirmAdjustNetwork"
  >
    <section class="adjust-info">
      <span class="adjust-name">实例名/ID</span>
      <span class="adjust-value">xxx</span>
    </section>
    <section class="adjust-info">
      <span class="adjust-name">IP地址</span>
      <span class="adjust-value">xxx</span>
    </section>
    <section class="adjust-info">
      <span class="adjust-name">网络计费模式</span>
      <span class="adjust-value">共享带宽包</span>
    </section>
    <section class="adjust-info">
      <span class="adjust-name">带宽上限</span>
      <span class="adjust-value">6000Mbps</span>
    </section>
    <section class="adjust-info">
      <span class="adjust-name">目标带宽上限</span>
      <span class="adjust-value">
        <bk-radio-group
          value="xxx"
        >
          <bk-radio label="无上限" />
          <bk-radio label="有上限" />
        </bk-radio-group>
      </span>
    </section>
  </bk-dialog> -->

  <bk-dialog
    :is-show="showChangeIP"
    title="更换IP"
    theme="primary"
    quick-close
    @closed="handleToggleShowChangeIP"
    @confirm="handleConfirmChangeIP"
  >
    确定更换实例（xxx）上的IP？更换后原IP可能无法找回
  </bk-dialog>

  <bk-dialog
    title="解绑弹性IP"
    theme="primary"
    quick-close
    :is-show="showUnbind"
    :is-loading="inUnbinding"
    @closed="handleToggleShowUnbind"
    @confirm="handleConfirmUnbind"
  >
    <span class="adjust-title">主机（{{ data.id }}）要解除绑定的弹性IP：</span>
    <section class="adjust-info">
      <span class="adjust-name">EIP地址公网地址</span>
      <span class="adjust-value">{{ unbindData.public_ip }}</span>
    </section>
    <section class="adjust-info">
      <span class="adjust-name">已绑定的实例</span>
      <span class="adjust-value">{{ unbindData.cvm_id }}</span>
    </section>
  </bk-dialog>

  <bk-dialog
    width="620"
    theme="primary"
    quick-close
    :title="`主机（${data.id}）绑定弹性IP`"
    :is-show="showBind"
    @closed="handleToggleShowBind(false)"
  >
    <template v-if="needNetwork">
      <span class="bind-title">选择网络接口</span>
      <bk-select v-model="bindData.network_interface_id" class="mb20">
        <bk-option
          v-for="(item, index) in networklist"
          :key="index"
          :value="item.id"
          :label="item.name"
          :disabled="(item.public_ipv4 && item.public_ipv4.length) || (item.public_ipv6 && item.public_ipv6.length)"
        />
      </bk-select>
    </template>
    <span class="bind-title">选择 EIP</span>
    <bk-table
      :data="eipList"
      :outer-border="false"
      class="mb20"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    >
      <bk-table-column label="ID" prop="public_ip">
        <!-- eslint-disable-next-line vue/no-template-shadow -->
        <template #default="{ data }">
          <bk-radio
            :model-value="bindData.eip_id"
            :label="data?.id"
            :key="data?.id"
            @change="bindData.eip_id = data?.id"
          ></bk-radio>
        </template>
      </bk-table-column>
      <bk-table-column label="名称" prop="name" />
      <bk-table-column label="弹性公网IP" prop="public_ip" />
    </bk-table>
    <template #footer>
      <bk-button theme="primary" :loading="isBinding" @click="handleConfirmBind">确定</bk-button>
      <bk-button class="bk-dialog-cancel" :disabled="isBinding" @click="handleToggleShowBind(false)">取消</bk-button>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.btn {
  min-width: 88px;
}
.adjust-title {
  display: inline-block;
  margin-bottom: 20px;
}
.adjust-info {
  margin-bottom: 20px;
  .adjust-name {
    display: inline-block;
    width: 120px;
    color: #979ba5;
  }
}
.bind-title {
  display: inline-block;
  margin: 10px 0;
}
</style>
