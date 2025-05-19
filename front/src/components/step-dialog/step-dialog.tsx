import { defineComponent, ref, VNode, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import './step.dialog.scss';

type StepType = {
  status?: string;
  title: string;
  disableNext?: boolean;
  isConfirmLoading?: boolean;
  component: () => VNode;
  footer?: () => VNode;
};

export default defineComponent({
  props: {
    title: {
      type: String,
    },
    business: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
    steps: {
      type: Array as PropType<StepType[]>,
    },
    size: {
      type: String,
      default: 'large',
    },
    loading: {
      type: Boolean,
    },
    dialogWidth: {
      type: String,
      default() {
        return '1000';
      },
    },
    dialogHeight: {
      type: String,
      default() {
        return '720';
      },
    },
    renderType: String as PropType<'if' | 'is'>,
    confirmDisabled: Boolean,
  },

  emits: ['confirm', 'cancel', 'next'],

  setup(_, { emit }) {
    const { t } = useI18n();

    const curStep = ref(1);

    const handleNextStep = () => {
      curStep.value += 1;
      emit('next', curStep.value);
    };

    const handlePreviousStep = () => {
      curStep.value -= 1;
    };

    const handleClose = () => {
      curStep.value = 1;
      emit('cancel');
    };

    const handleConfirm = () => {
      // curStep.value = 1;
      emit('confirm');
    };

    return {
      curStep,
      t,
      handleNextStep,
      handlePreviousStep,
      handleClose,
      handleConfirm,
    };
  },

  render() {
    return (
      <>
        <bk-dialog
          render-directive={this.renderType || undefined}
          class='step-dialog'
          width={this.dialogWidth}
          height={this.dialogHeight}
          theme='primary'
          headerAlign='center'
          size={this.size}
          title={this.title}
          isShow={this.isShow}
          quick-close={false}
          close-icon={false}
          onClosed={this.handleClose}>
          {{
            default: () => {
              return (
                <>
                  {this.steps.length > 1 ? (
                    <bk-steps class='dialog-steps' steps={this.steps} cur-step={this.curStep} />
                  ) : (
                    ''
                  )}
                  {this.steps[this.curStep - 1].component()}
                </>
              );
            },
            footer: () => {
              return (
                <>
                  {this.steps[this.curStep - 1].footer?.()}
                  {this.curStep > 1 ? (
                    <bk-button class='mr10 dialog-button' onClick={this.handlePreviousStep}>
                      {this.t('上一步')}
                    </bk-button>
                  ) : (
                    ''
                  )}
                  {this.curStep < this.steps.length ? (
                    <bk-button
                      class='mr10 dialog-button'
                      theme='primary'
                      disabled={this.steps[this.curStep - 1].disableNext || (this.curStep > 1 ? !this.business : false)}
                      onClick={this.handleNextStep}>
                      {this.t('下一步')}
                    </bk-button>
                  ) : (
                    ''
                  )}
                  {this.curStep >= this.steps.length ? (
                    <bk-button
                      class='mr10 dialog-button'
                      theme='primary'
                      disabled={this.confirmDisabled}
                      loading={this.steps[this.curStep - 1].isConfirmLoading || this.loading}
                      onClick={this.handleConfirm}>
                      {this.t('确认')}
                    </bk-button>
                  ) : (
                    ''
                  )}
                  <bk-button
                    class='dialog-button'
                    onClick={this.handleClose}
                    disabled={this.steps[this.curStep - 1].isConfirmLoading || this.loading}>
                    {this.t('取消')}
                  </bk-button>
                </>
              );
            },
          }}
        </bk-dialog>
      </>
    );
  },
});
