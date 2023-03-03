import {
  Table,
  Loading
} from 'bkui-vue';
import {
  defineComponent,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import useQueryList  from '../../../hooks/use-query-list';
import useColumns from '../../../hooks/use-columns';
import useSelection from '../../../hooks/use-selection';
import {
  useResourceStore
} from '@/store/resource';

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
    vender: {
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
            field: 'vender',
            op: 'eq',
            value: props.vender,
          }],
        },
      },
      'disks'
    );

    const {
      selections,
      handleSelectionChange,
    } = useSelection();

    const columns = useColumns('drive');

    const resourceStore = useResourceStore();

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      resourceStore.attachDisk(
        props.vender,
        selections.value.map(selection => ({
          disk_id: selection.disk_id,
          cvm_id: props.id,
          caching_type: selection.caching_type
        }))
      ).then(() => {
        handleClose();
      })
    };

    return {
      datas,
      pagination,
      isLoading,
      columns,
      handlePageChange,
      handlePageSizeChange,
      handleSort,
      handleSelectionChange,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    const steps = [
      {
        component: () =>
          <Loading loading={this.isLoading}>
            <Table
              class="mt20"
              row-hover="auto"
              remote-pagination
              pagination={this.pagination}
              columns={this.columns}
              data={this.datas}
              onPageLimitChange={this.handlePageSizeChange}
              onPageValueChange={this.handlePageChange}
              onColumnSort={this.handleSort}
              onSelectionChange={this.handleSelectionChange}
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
