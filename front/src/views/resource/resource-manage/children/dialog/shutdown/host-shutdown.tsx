import { defineComponent, ref } from 'vue';
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
    const tableData = ref([]);
    const columns: any[] = [{ label: '23' }];
    const steps = [
      {
        component: () => (
          <>
            <span>{t('您已选择 {count} 台实例，进行关机操作前，请确认', { count: 6 })}：</span>
            <bk-table class='mt20' row-hover='auto' columns={columns} data={tableData.value} show-overflow-tooltip />
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
      handleClose,
      handleConfirm,
    };
  },

  render() {
    return (
      <>
        <step-dialog
          title={this.title}
          isShow={this.isShow}
          steps={this.steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
