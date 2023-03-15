<script lang="ts" setup>
import {
  ref,
  h,
  watch,
} from 'vue';
import {
  Button,
  Message,
} from 'bkui-vue';
import {
  useResourceStore,
} from '@/store/resource';
import useQueryList from '../../../hooks/use-query-list'

const props = defineProps({
  data: {
    type: Object,
  }
});

const resourceStore = useResourceStore();

// 状态
const showAdjustNetwork = ref(false);
const showChangeIP = ref(false);
const showUnbind = ref(false);
const showBind = ref(false);
const unbindData = ref();
const selection = ref();

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
    render(data: any) {
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
          Button,
          {
            text: true,
            theme: 'primary',
            onClick() {
              handleToggleShowUnbind(data);
            },
          },
          [
            '解绑',
          ],
        ),
      ]
    },
  },
]);

// 当前主机下的eip资源
const {
  datas,
  isLoading,
} = useQueryList(
  {},
  'eip',
  () => {
    return Promise.all([resourceStore.getEipListByCvmId(props.data.vendor, props.data.id)])
  }
);

// 当前 vender 下的eip资源
const {
  datas: eipList,
  pagination,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(
  {
    filter: {
      op: 'and',
      rules: [{
        field: 'vendor',
        op: 'eq',
        value: props.data.vendor,
      }],
    },
  },
  'eips'
);

const handleToggleShowAdjustNetwork = () => {
  showAdjustNetwork.value = !showAdjustNetwork.value
}

const handleConfirmAdjustNetwork = () => {
  handleToggleShowAdjustNetwork()
}

const handleToggleShowChangeIP = () => {
  showChangeIP.value = !showChangeIP.value
}

const handleConfirmChangeIP = () => {
  handleToggleShowChangeIP()
}

const handleToggleShowUnbind = (data?: any) => {
  unbindData.value = data
  showUnbind.value = !showUnbind.value
}

const handleConfirmUnbind = () => {
  resourceStore.disassociateEip(
    {
      eip_id: unbindData.value.id
    }
  ).then(() => {
    handleToggleShowUnbind()
  }).catch((err: any) => {
    Message({
      theme: 'error',
      message: err.message || err
    })
  })
}

const handleToggleShowBind = () => {
  showBind.value = !showBind.value
}

const handleConfirmBind = () => {
  resourceStore.associateEip(
    {
      eip_id: selection.value.id,
      cvm_id: props.data.id,
      network_interface_id: props.data.network_interface_id,
    }
  ).then(() => {
    handleToggleShowBind()
  }).catch((err: any) => {
    Message({
      theme: 'error',
      message: err.message || err
    })
  })
}

watch(
  () => props.data,
  () => {
    if (props.data.vendor === 'tcloud') {
      columns.value.splice(4, 0 , ...[
        {
          label: '计费模式',
          field: 'internet_charge_type',
        },
        {
          label: '带宽上限',
          field: 'bandwidth',
        }
      ])
    }
  },
  {
    deep: true,
    immediate: true
  }
)
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <!-- <bk-button
      class="mt20"
      theme="primary"
      @click="handleToggleShowBind"
    >
      绑定
    </bk-button> -->
    <bk-table
      class="mt20"
      row-hover="auto"
      :columns="columns"
      :data="datas"
    />
  </bk-loading>
  <bk-dialog
    :is-show="showAdjustNetwork"
    title="调整网络"
    theme="primary"
    quick-close
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
  </bk-dialog>

  <bk-dialog
    :is-show="showChangeIP"
    title="更换IP"
    theme="primary"
    quick-close
    @closed="handleToggleShowChangeIP"
    @confirm="handleConfirmChangeIP"
  >
    确定更换实例（xxx）上的IP（119.28.100.27）？更换后原IP可能无法找回
  </bk-dialog>

  <bk-dialog
    :is-show="showUnbind"
    title="解绑弹性IP"
    theme="primary"
    quick-close
    @closed="handleToggleShowUnbind"
    @confirm="handleConfirmUnbind"
  >
    <span class="adjust-title">主机（{{ data.id }}）要解除绑定的弹性IP：</span>
    <section class="adjust-info">
      <span class="adjust-name">网络接口ID</span>
      <span class="adjust-value">{{ unbindData.network_interface_id }}</span>
    </section>
    <section class="adjust-info">
      <span class="adjust-name">内部IP</span>
      <span class="adjust-value">{{ unbindData.public_ip }}</span>
    </section>
    <section class="adjust-info">
      <span class="adjust-name">公网IP地址</span>
      <span class="adjust-value">{{ unbindData.public_ip }}</span>
    </section>
  </bk-dialog>

  <bk-dialog
    :is-show="showBind"
    width="620"
    title="绑定弹性IP"
    theme="primary"
    quick-close
    @closed="handleToggleShowBind"
    @confirm="handleConfirmBind"
  >
    <span class="adjust-title">主机（{{ data.id }}）要绑定的弹性IP：</span>
    <bk-radio-group
      value="xxx"
    >
      <bk-radio label="主网卡(192.168.0.169)" />
      <bk-radio label="扩展网卡(192.168.0.169)" />
    </bk-radio-group>
    <bk-table
      :data="eipList"
      :outer-border="false"
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    >
      <bk-table-column
        label="弹性公网IP"
        prop="public_ip"
      />
      <bk-table-column
        label="类型"
        prop="address_type"
      />
      <bk-table-column
        label="带宽大小"
        prop="ip"
      />
      <bk-table-column
        label="带宽类型"
        prop="ip"
      />
    </bk-table>
  </bk-dialog>
</template>

<style lang="scss" scoped>
  .adjust-title {
    display: inline-block;
    margin-bottom: 20px;
  }
  .adjust-info {
    margin-bottom: 20px;
    .adjust-name {
      display: inline-block;
      width: 120px;
      color: #979BA5;
    }
  }
</style>
