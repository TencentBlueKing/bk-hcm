import {
  defineComponent,
  ref,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
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
    const {
      t,
    } = useI18n();

    // 状态
    const tableData = ref([]);
    const columns: any[] = [{ label: '23' }];
    const steps = [
      {
        component: () => <>
          <span>{ t('您已选择 {count} 个云硬盘，进行卸载操作，请确认', { count: 5 }) }：</span>
          <bk-table
            class="mt20"
            row-hover="auto"
            columns={columns}
            data={tableData.value}
          />
          <h3 class="g-resource-tips">
            { t('win实例：强烈建议您在卸载之前，对该硬盘执行脱机操作') }<br />
            { t('linux实例：建议您在卸载之前，确保该硬盘的所有分区处于非加载状态 (umounted)。部分linux操作系统可能不支持硬盘热拔插') }<br />
          </h3>
        </>,
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
    return <>
      <step-dialog
        title={this.t('卸载云硬盘')}
        isShow={this.isShow}
        steps={this.steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});
