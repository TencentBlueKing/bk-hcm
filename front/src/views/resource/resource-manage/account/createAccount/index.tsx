import { Button, Dialog, Steps } from 'bkui-vue';
import { PropType, defineComponent, ref } from 'vue';
import './index.scss';
import AccountForm from './components/accountForm';
import AccountResource from './components/accountResource';
import ResultPage from './components/resultPage';
import { useAccountStore } from '@/store';
import { useCalcTopWithNotice } from '@/views/home/hooks/useCalcTopWithNotice';

export default defineComponent({
  props: {
    isShow: {
      type: Boolean,
      required: true,
    },
    onSubmit: {
      type: Function as PropType<() => void>,
      required: true,
    },
    onCancel: {
      type: Function as PropType<() => void>,
      required: true,
    },
  },
  setup(props) {
    const step = ref(1);
    const enableNextStep = ref(false);
    const changeEnableNextStep = (val: boolean) => {
      enableNextStep.value = val;
    };
    const submitData = ref({});
    const changeSubmitData = (val: Record<string, string | Object>) => {
      submitData.value = val;
    };
    const isSubmitLoading = ref(false);
    const accountStore = useAccountStore();
    const errMsg = ref('');
    const validateForm = ref(async () => {});
    const secretIds = ref({});
    const handleSubmit = async () => {
      isSubmitLoading.value = true;
      try {
        await accountStore.applyAccount(submitData.value);
      } catch (err: any) {
        errMsg.value = err.message;
      } finally {
        isSubmitLoading.value = false;
        step.value += 1;
      }
    };
    const handleNextStep = async () => {
      await validateForm.value();
      step.value += 1;
    };

    const [, isNoticeAlert] = useCalcTopWithNotice(52);

    return () => (
      <Dialog
        fullscreen
        showMask={false}
        isShow={props.isShow}
        onClosed={() => {
          step.value = 1;
          props.onCancel();
        }}
        title='云账号接入'
        class={['create-account-dialog-container', { 'has-notice': isNoticeAlert.value }]}>
        {{
          default: () => (
            <div class={'create-account-dialog-content'}>
              {step.value < 3 ? (
                <Steps
                  curStep={step.value}
                  class={'create-account-dialog-steps'}
                  steps={[
                    {
                      title: '录入账号',
                    },
                    {
                      title: '资源同步',
                    },
                  ]}
                />
              ) : (
                <ResultPage errMsg={errMsg.value} type={errMsg.value.length > 0 ? 'failure' : 'success'} />
              )}
              <AccountForm
                changeEnableNextStep={changeEnableNextStep}
                changeSubmitData={changeSubmitData}
                changeValidateForm={(callback) => (validateForm.value = callback)}
                changeExtension={(extension) => (secretIds.value = extension)}
                style={
                  step.value === 1
                    ? ''
                    : {
                        display: 'none',
                      }
                }
              />
              {step.value === 2 ? (
                <AccountResource secretIds={submitData.value?.extension} vendor={submitData.value?.vendor} />
              ) : null}
            </div>
          ),
          footer: () => (
            <div class={'create-account-dialog-footer'}>
              <>
                {step.value < 3 ? (
                  <>
                    {step.value > 1 ? (
                      <Button class={'mr8'} onClick={() => (step.value -= 1)} loading={isSubmitLoading.value}>
                        上一步
                      </Button>
                    ) : null}
                    {step.value < 2 ? (
                      <Button theme={'primary'} class={'mr8'} disabled={!enableNextStep.value} onClick={handleNextStep}>
                        下一步
                      </Button>
                    ) : (
                      <Button theme={'primary'} class={'mr8'} loading={isSubmitLoading.value} onClick={handleSubmit}>
                        提交
                      </Button>
                    )}
                  </>
                ) : null}
              </>

              {step.value < 3 ? (
                <Button
                  onClick={() => {
                    step.value = 1;
                    props.onCancel();
                  }}
                  loading={isSubmitLoading.value}>
                  取消
                </Button>
              ) : null}
            </div>
          ),
        }}
      </Dialog>
    );
  },
});
