import { Table, Loading, Radio } from 'bkui-vue';
import { defineComponent, h, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import useQueryList from '../../../hooks/use-query-list';
import useColumns from '../../../hooks/use-columns';
import { useResourceStore } from '@/store/resource';

// 绑定eip
export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    isShow: {
      type: Boolean,
    },
    detail: {
      type: Object,
    },
  },

  emits: ['update:isShow', 'success-assign'],

  setup(props, { emit }) {
    const { t } = useI18n();

    const type = ['tcloud', 'aws'].includes(props.detail.vendor) ? 'cvms' : 'network_interfaces';
    const columnType = ['tcloud', 'aws'].includes(props.detail.vendor) ? 'cvms' : 'networkInterface';

    const rules = [
      {
        field: 'vendor',
        op: 'eq',
        value: props.detail.vendor,
      },
      {
        field: 'account_id',
        op: 'eq',
        value: props.detail.account_id,
      },
      {
        field: 'region',
        op: 'eq',
        value: props.detail.region,
      },
    ];

    if (type === 'network_interfaces') {
      rules.push(
        ...[
          {
            field: 'public_ipv4',
            op: 'json_length',
            value: 0,
          },
          {
            field: 'public_ipv6',
            op: 'json_length',
            value: 0,
          },
        ],
      );
    } else {
      rules.push(
        ...[
          {
            field: 'public_ipv6_addresses',
            op: 'json_length',
            value: 0,
          },
          {
            field: 'public_ipv4_addresses',
            op: 'json_length',
            value: 0,
          },
          {
            field: 'recycle_status',
            op: 'neq',
            value: 'recycling',
          },
          {
            field: 'status',
            op: 'in',
            value: [
              'RUNNING',
              'STOPPED',
              'STARTING',
              'STOPPING',
              'REBOOTING',
              'REBOOT',
              'HARD_REBOOT',
              'ACTIVE',
              'SHUTOFF',
              'running',
              'stopping',
              'stopped',
              'Succeeded',
              'PowerState/running',
            ],
          },
        ],
      );
    }

    const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryList(
      {
        filter: {
          op: 'and',
          rules,
        },
      },
      type,
      null,
      type === 'cvms' ? 'list' : 'getUnbindEipNetworkList',
      type === 'cvms' ? {} : { is_associate: true },
    );

    const { columns } = useColumns(columnType, true);
    const resourceStore = useResourceStore();
    const selection = ref<any>({});
    const isConfirmLoading = ref(false);
    const renderColumns = [
      {
        label: 'ID',
        field: 'id',
        render({ data }: any) {
          return h(Radio, {
            'model-value': selection.value.id,
            label: data.id,
            key: data.id,
            onChange() {
              selection.value = data;
            },
          });
        },
      },
      ...columns,
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      isConfirmLoading.value = true;
      const postData =
        type === 'cvms'
          ? {
              eip_id: props.detail.id,
              cvm_id: selection.value.id,
            }
          : {
              eip_id: props.detail.id,
              network_interface_id: selection.value.id,
              cvm_id: selection.value.cvm_id,
            };
      resourceStore
        .associateEip(postData)
        .then(() => {
          handleClose();
          emit('success-assign');
        })
        .finally(() => {
          isConfirmLoading.value = false;
        });
    };

    return {
      type,
      datas,
      pagination,
      isLoading,
      renderColumns,
      isConfirmLoading,
      handlePageChange,
      handlePageSizeChange,
      handleSort,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    const steps = [
      {
        isConfirmLoading: this.isConfirmLoading,
        component: () => (
          <Loading loading={this.isLoading}>
            <Table
              class='mt20'
              row-hover='auto'
              remote-pagination
              pagination={this.pagination}
              columns={this.renderColumns}
              data={this.datas}
              onPageLimitChange={this.handlePageSizeChange}
              onPageValueChange={this.handlePageChange}
              onColumnSort={this.handleSort}
            />
          </Loading>
        ),
      },
    ];

    return (
      <>
        <step-dialog
          title={this.type === 'cvms' ? '绑定EIP' : '绑定网络接口'}
          isShow={this.isShow}
          steps={steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
