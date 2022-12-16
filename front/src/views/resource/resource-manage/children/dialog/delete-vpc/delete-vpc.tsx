import {
  defineComponent,
  ref,
} from 'vue';

import {
  useI18n,
} from 'vue-i18n';
import {
  useResourceStore,
} from '@/store/resource';

import StepDialog from '@/components/step-dialog/step-dialog';

export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
  },

  emits: ['update:isShow'],

  setup(_, { emit }) {
    const {
      t,
    } = useI18n();

    const resourceStore = useResourceStore();

    // 状态
    const isDeleting = ref(false);
    const tableData = ref([]);
    const columns = [
      {
        label: 'ID',
        field: 'id',
      },
      {
        label: '资源 ID',
        field: 'cid',
      },
      {
        label: '名称',
        field: 'name',
      },
      {
        label: '云厂商',
        field: 'vendor',
      },
      {
        label: '业务',
        field: 'id',
      },
      {
        label: '业务拓扑',
        field: 'id',
      },
      {
        label: '云区域',
        field: 'id',
      },
      {
        label: '地域',
        field: 'region',
      },
      {
        label: 'IPv4 CIDR',
        field: 'ipv4_cidr',
      },
      {
        label: 'IPv6 CIDR',
        field: 'ipv6_cidr',
      },
      {
        label: '状态',
        field: 'status',
      },
      {
        label: '默认 VPC',
        field: 'is_default',
      },
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      isDeleting.value = true;
      resourceStore
        .delete('vpc', 123)
        .then(() => {
          handleClose();
        })
        .finally(() => {
          isDeleting.value = false;
        });
    };

    return {
      isDeleting,
      columns,
      tableData,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    const steps = [
      {
        component: () => <>
          <bk-table
            class="mb20"
            row-hover="auto"
            columns={this.columns}
            data={this.tableData}
          />
          <h3 class="g-resource-tips">
            { this.t('请注意该VPC包含一个或多个资源，在释放这些资源前，无法删除VPC') }：<br />
            { this.t('子网：{count} 个', { count: 5 }) }<br />
            { this.t('CVM：{count} 个', { count: 5 }) }
          </h3>
        </>,
        isConfirmLoading: this.isDeleting,
      },
    ];

    return <>
      <step-dialog
        title={this.t('删除 VPC')}
        isShow={this.isShow}
        steps={steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});
