import './expansion-drive.scss';
import { defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
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

  setup(props, { emit }) {
    const { t } = useI18n();

    // 状态
    const steps = [
      {
        title: '选择目标云硬盘',
        component: () => (
          <>
            <bk-table data={[]} outer-border={false} dark-header show-overflow-tooltip>
              <bk-table-column type='selection' width='60' />
              <bk-table-column label='ID' prop='id' />
              <bk-table-column label='云硬盘名' prop='id' />
              <bk-table-column label='配置' prop='id' />
              <bk-table-column label='计费模式' prop='id' />
            </bk-table>
          </>
        ),
      },
      {
        title: '调整容量',
        component: () => (
          <>
            <bk-table data={[]} outer-border={false} dark-header key='size' show-overflow-tooltip>
              <bk-table-column label='云硬盘名称' prop='id' />
              <bk-table-column label='云硬盘ID' prop='id' />
              <bk-table-column label='计费模式' prop='id' />
              <bk-table-column label='配置' prop='id' />
            </bk-table>
            <section class='expansion-info'>
              <span class='expansion-name'>当前容量</span>
              <span class='expansion-value'>50 GB</span>
            </section>
            <section class='expansion-info'>
              <span class='expansion-name'>目标容量</span>
              <bk-input type='number' class='expansion-value mr5'></bk-input>
              GB
            </section>
            <section class='expansion-info'>
              <span class='expansion-name'>费用</span>
              <span class='expansion-value'>100</span>
            </section>
          </>
        ),
      },
      {
        title: '扩容分区及文件系统',
        component: () => (
          <>
            <bk-alert
              theme='info'
              title='完成扩容操作后，请登录实例确认是否已完成自动扩展文件系统，否则需要手动扩文件系统及分区'
            />
            <bk-alert theme='info mt20' title='当前操作需要实例在关机状态下进行，为了避免数据丢失，请仔细确认' />
            <section class='expansion-info mt20'>
              <span class='expansion-name'>强制关机</span>
              <bk-checkbox value='value'>同意强制关机</bk-checkbox>
            </section>
          </>
        ),
      },
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      handleClose();
    };

    return {
      steps,
      t,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    return (
      <>
        <step-dialog
          title={this.t('云硬盘扩容')}
          isShow={this.isShow}
          steps={this.steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
