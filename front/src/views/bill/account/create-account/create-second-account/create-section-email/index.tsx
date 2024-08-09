import { Alert, Button, Dialog, Form, Input } from 'bkui-vue';
import { computed, defineComponent, onUnmounted, ref, nextTick, watch } from 'vue';
import { Scenes } from '../constants';
import useBillStore from '@/store/useBillStore';
import cssModule from './index.module.scss';
const { FormItem } = Form;
export default defineComponent({
  props: {
    suffixText: String,
    isMailValid: Boolean,
    formModel: Object,
  },
  emits: ['changeEmail'],
  setup(props, { expose, emit }) {
    const billStore = useBillStore();

    // 表单input部分
    const email = ref('');
    watch(
      () => email.value,
      () => {
        emit('changeEmail', email.value);
      },
    );
    const isComplete = ref(false);
    const emailCodeVerfiyResult = ref(null);
    const formCodeRef = ref();
    const countdownNum = ref(60);
    const isCountdown = ref(false);
    const isNameValid = ref(false);
    const isSendBtnDisabled = computed(() => {
      return countdownNum.value > 0;
    });
    const clearValidate = () => {
      nextTick(() => {
        formCodeRef.value.clearValidate();
      });
    };

    // dialog部分
    const isDialogShow = ref(false);
    const dialogForm = ref({
      code: '',
    });

    const handleClose = () => {
      isComplete.value = true;
      empty();
      clearValidate();
    };

    const handleResendButton = () => {
      dialogForm.value.code = '';
      getCode();
      countdown();
    };

    const empty = () => {
      isCountdown.value = true;
      isDialogShow.value = false;
    };

    const handleConfirm = async () => {
      await formCodeRef.value.validate();
      checkingCode(false);
      empty();
      clearValidate();
    };

    const checkingCode = async (isDeleteAfterVerify: boolean) => {
      try {
        const { data } = await billStore.verify_code({
          mail: `${email.value}${props.suffixText}`,
          scene: Scenes.SecondAccountApplication,
          verify_code: dialogForm.value.code,
          delete_after_verify: isDeleteAfterVerify,
        });
        emailCodeVerfiyResult.value = !!data;
      } catch (err) {
        // console.log(err);
      } finally {
        isComplete.value = true;
      }
    };

    const getCode = async () => {
      try {
        await billStore.send_code({
          mail: `${email.value}${props.suffixText}`,
          scene: Scenes.SecondAccountApplication,
          info: {
            vendor: props.formModel.vendor,
            account_name: props.formModel.name,
          },
        });
      } catch (err) {
        // console.log(err);
      }
    };

    // 邮箱验证码
    let timer: string | number | NodeJS.Timeout = null;

    const countdown = () => {
      clearInterval(timer);
      countdownNum.value = 60; // 重置倒计时
      timer = setInterval(() => {
        if (countdownNum.value > 0) {
          countdownNum.value = countdownNum.value - 1;
          if (countdownNum.value === 1) {
            isCountdown.value = false;
          }
        } else {
          clearInterval(timer);
        }
      }, 1000);
    };

    const handleverifi = async () => {
      try {
        getCode();
        dialogForm.value.code = '';
        isDialogShow.value = true;
        countdown();
      } catch (error) {}
    };

    const isfirst = computed(() => {
      return countdownNum.value < 60 && isCountdown.value;
    });

    const changeNameValid = (value: boolean) => {
      isNameValid.value = value;
    };

    onUnmounted(() => {
      clearInterval(timer);
    });

    expose({
      changeNameValid,
      dialogForm,
      emailCodeVerfiyResult,
    });

    return () => (
      <>
        <Input class={cssModule['email-input']} v-model={email.value} suffix={props.suffixText} />
        {isfirst.value ? (
          <>
            <Button theme='primary' style={'width:86px'} disabled={isSendBtnDisabled.value}>
              {countdownNum.value}s
            </Button>
          </>
        ) : (
          <>
            <Button
              theme='primary'
              disabled={!(props.isMailValid && isNameValid.value)}
              onClick={handleverifi}
              v-bk-tooltips={{
                content: !isNameValid.value
                  ? !props.isMailValid
                    ? '请输入账号名称，账号邮箱'
                    : '请输入账号名称'
                  : !props.isMailValid && '请输入邮箱',
                disabled: props.isMailValid && isNameValid.value,
              }}>
              {isComplete.value ? ' 重新校验' : '邮箱校验'}
            </Button>
          </>
        )}

        <p class={cssModule['email-tip']}>
          {(function () {
            const iconClassList = ['hcm-icon'];
            let text = '请确保邮箱已按指引配置，否则后续帐号将无法创建';
            if (emailCodeVerfiyResult.value === null)
              iconClassList.push('bkhcm-icon-alert', cssModule['email-tip-icon']);
            else if (emailCodeVerfiyResult.value === true) {
              iconClassList.push('bkhcm-icon-check-circle-fill', cssModule['email-tip-check']);
              text = '校验通过';
            } else {
              iconClassList.push('bkhcm-icon-close-circle-fill', cssModule['email-tip-close']);
              text = '校验失败';
            }
            return (
              <>
                <i class={iconClassList}></i>
                {text}
              </>
            );
          })()}
        </p>

        <Dialog v-model:is-show={isDialogShow.value} title='邮箱校验' quick-close>
          {{
            default: () => (
              <div>
                <Alert
                  class={cssModule['dialog-alert']}
                  theme='info'
                  closable
                  title='验证码已发送至该邮箱帐号，请在下方输入验证码以进行校验'
                />
                <Form
                  formType='vertical'
                  ref={formCodeRef}
                  model={dialogForm.value}
                  rules={{
                    code: [
                      {
                        required: true,
                        trigger: 'blur',
                        message: '请输入六位数字验证码',
                        validator: (val: string) => {
                          return /^\d{6}$/.test(val);
                        },
                      },
                    ],
                  }}>
                  <FormItem label='验证码输入' required property='code'>
                    <div class='flex-row'>
                      <Input v-model={dialogForm.value.code} placeholder='请输入' />
                      <Button
                        theme='primary'
                        class='ml8'
                        onClick={handleResendButton}
                        disabled={isSendBtnDisabled.value}>
                        {isSendBtnDisabled.value ? `${countdownNum.value}s` : '重新发送'}
                      </Button>
                    </div>
                  </FormItem>
                </Form>
              </div>
            ),
            footer: () => (
              <>
                <Button theme='primary' disabled={!dialogForm.value.code} onClick={handleConfirm}>
                  提交
                </Button>
                <Button class='ml8' onClick={handleClose}>
                  取消
                </Button>
              </>
            ),
          }}
        </Dialog>
      </>
    );
  },
});
