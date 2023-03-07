import {
  Table,
  Loading,
  Radio,
  Message,
} from 'bkui-vue';
import {
  defineComponent,
  h,
  ref,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import useQueryList  from '../../../hooks/use-query-list';
import useColumns from '../../../hooks/use-columns';
import {
  useResourceStore
} from '@/store/resource';

// 主机选硬盘挂载
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
    vendor: {
      type: String,
    },
    id: {
      type: String
    }
  },

  emits: ['update:isShow'],

  setup(props, { emit }) {
    const {
      t,
    } = useI18n();

    const {
      datas,
      pagination,
      isLoading,
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
            value: props.vendor,
          }],
        },
      },
      'disks'
    );

    const columns = useColumns('drive', true);

    const resourceStore = useResourceStore();

    const selection = ref<any>({});

    const isConfirmLoading = ref(false);

    const renderColumns = [
      {
        label: 'ID',
        field: 'id',
        render({ data }: any) {
          return h(
            Radio,
            {
              'model-value': selection.value.id,
              label: data.id,
              key: data.id,
              onChange() {
                selection.value = data;
              },
            }
          );
        },
      },
      ...columns
    ]

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      isConfirmLoading.value = true;
      resourceStore.attachDisk(
        props.vendor,
        {
          disk_id: selection.value.id,
          cvm_id: props.id,
          caching_type: selection.value.caching_type
        }
      ).then(() => {
        handleClose();
      }).catch((err: any) => {
        Message({
          theme: 'error',
          message: err.message || err
        })
      }).finally(() => {
        isConfirmLoading.value = false;
      })
    };

    return {
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
        component: () =>
          <Loading loading={this.isLoading}>
            <Table
              class="mt20"
              row-hover="auto"
              remote-pagination
              pagination={this.pagination}
              columns={this.renderColumns}
              data={this.datas}
              onPageLimitChange={this.handlePageSizeChange}
              onPageValueChange={this.handlePageChange}
              onColumnSort={this.handleSort}
            />
          </Loading>
      },
    ];

    return <>
      <step-dialog
        title={this.t('挂载云硬盘')}
        isShow={this.isShow}
        steps={steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});
