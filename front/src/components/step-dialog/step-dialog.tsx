import {
  defineComponent,
  ref,
  VNode,
  PropType,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import './step.dialog.scss';

type StepType = {
  status?: string;
  title: string;
  component: () => VNode;
};

export default defineComponent({
  props: {
    title: {
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
  },

  emits: ['confirm', 'cancel'],

  setup(_, { emit }) {
    const {
      t,
    } = useI18n();

    const curStep = ref(1);

    const handleNextStep = () => {
      curStep.value += 1;
    };

    const handlePreviousStep = () => {
      curStep.value -= 1;
    };

    const handleClose = () => {
      curStep.value = 1;
      emit('cancel');
    };

    const handleConfirm = () => {
      curStep.value = 1;
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
    return <>
      <bk-dialog
        theme="primary"
        headerAlign="center"
        size={this.size}
        title={this.title}
        isShow={this.isShow}
        onClosed={this.handleClose}
      >
        {{
          default: () => {
            return <>
              {
                this.steps.length > 1
                  ? <bk-steps
                      class="dialog-steps"
                      steps={this.steps}
                      cur-step={this.curStep}
                    />
                  : ''
              }
              {
                this.steps[this.curStep - 1].component()
              }
            </>;
          },
          footer: () => {
            return <>
              {
                this.curStep < this.steps.length
                  ? <bk-button
                      class="mr10 dialog-button"
                      theme="primary"
                      onClick={this.handleNextStep}
                    >{this.t('下一步')}</bk-button>
                  : ''
              }
              {
                this.curStep > 1
                  ? <bk-button
                      class="mr10 dialog-button"
                      onClick={this.handlePreviousStep}
                    >{this.t('上一步')}</bk-button>
                  : ''
              }
              {
                this.curStep >= this.steps.length
                  ? <bk-button
                      class="mr10 dialog-button"
                      theme="primary"
                      onClick={this.handleConfirm}
                    >{this.t('确认')}</bk-button>
                  : ''
              }
              <bk-button
                class="dialog-button"
                onClick={this.handleClose}
              >{this.t('取消')}</bk-button>
            </>;
          },
        }}
      </bk-dialog>
    </>;
  },
});
