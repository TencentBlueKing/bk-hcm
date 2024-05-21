import { defineComponent, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import StepDialog from '@/components/step-dialog/step-dialog';
import useColumns from '../../../hooks/use-columns';
import { useResourceStore } from '@/store/resource';

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
    data: {
      type: Object,
    },
  },

  emits: ['update:isShow', 'success'],

  setup(props, { emit }) {
    const { t } = useI18n();

    // 状态
    const { columns } = useColumns('drive', true);
    const resourceStore = useResourceStore();
    const isLoading = ref(false);
    const renderColumns = [
      {
        label: 'ID',
        field: 'id',
      },
      ...columns,
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      isLoading.value = true;
      resourceStore
        .detachDisk({
          disk_id: props.data.id,
          cvm_id: props.data.instance_id,
        })
        .then(() => {
          emit('success');
          handleClose();
        })
        .catch((err: any) => {
          console.error(err.message || err);
        })
        .finally(() => {
          isLoading.value = false;
        });
    };

    return {
      isLoading,
      renderColumns,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    const steps = [
      {
        isConfirmLoading: this.isLoading,
        component: () => (
          <>
            <span>{this.t('您已选择 {count} 个云硬盘，进行卸载操作，请确认', { count: 1 })}：</span>
            <bk-table
              class='mt20'
              row-hover='auto'
              columns={this.renderColumns}
              data={[this.data]}
              show-overflow-tooltip
            />
            <h3 class='g-resource-tips mt20'>
              {this.t('win实例：强烈建议您在卸载之前，对该硬盘执行脱机操作')}
              <br />
              {this.t(
                'linux实例：建议您在卸载之前，确保该硬盘的所有分区处于非加载状态 (umounted)。部分linux操作系统可能不支持硬盘热拔插',
              )}
              <br />
            </h3>
          </>
        ),
      },
    ];

    return (
      <>
        <step-dialog
          title={this.t('卸载云硬盘')}
          isShow={this.isShow}
          steps={steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
