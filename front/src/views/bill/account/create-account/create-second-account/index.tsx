import { computed, defineComponent, ref, watch, watchEffect, onUnmounted, nextTick } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Alert, Button, Form, Input, Message, ResizeLayout } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';
import { MAIN_ACCOUNT_VENDORS } from '../constants';
import { Success } from 'bkui-vue/lib/icon';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import MemberSelect from '@/components/MemberSelect';
import { useUserStore } from '@/store';
import useFormModel from '@/hooks/useFormModel';
import useBillStore from '@/store/useBillStore';
import { Extension_Name_Map, Scenes } from './constants';
import { useRoute, useRouter } from 'vue-router';
import BusinessSelector from '@/components/business-selector/index.vue';
import { PluginHandlerMailbox } from '@pluginHandler/create-account';
const { FormItem } = Form;
export default defineComponent({
  setup() {
    const { isuffix, emailRules, isMailRules } = PluginHandlerMailbox;
    const userStore = useUserStore();
    const formModelRef = ref();
    const formInstance = ref();
    const isLoading = ref(false);
    const formCodeRef = ref();
    const isDialogShow = ref(false);
    const isComplete = ref(false);
    const isNameRules = ref(false);
    const VerificationResults = ref('Not verified');
    const dialogForm = ref({
      code: '',
    });
    const countdownNum = ref(60);
    const isSendBtnDisabled = computed(() => {
      return countdownNum.value > 0;
    });
    const iscountdown = ref(false);
    const billStore = useBillStore();
    const router = useRouter();
    const route = useRoute();
    const { formModel } = useFormModel({
      name: '', // 名字
      vendor: VendorEnum.AZURE, // 云厂商
      email: '', // 邮箱
      managers: [], // 负责人数组
      bak_managers: [], // 备份负责人数组
      site: 'international', // 站点
      // dept_id: '', // 组织架构ID
      memo: '', // 备忘录
      op_product_id: '',
      bk_biz_id: -1, // 业务
      // extension: {}, // 扩展字段对象
    });
    const nameTips = computed(() => {
      let tip = '账号名称只能包括英文字母、数字和中划线-，并且仅能以英文字母开头，长度为6-20个字符';
      switch (formModel.vendor) {
        case VendorEnum.AWS:
        case VendorEnum.HUAWEI:
        case VendorEnum.AZURE:
          tip = '账号名称只能包括英文字母、数字和下划线_，并且仅能以英文字母开头，长度为6-20个字符';
          break;
      }
      return tip;
    });

    watchEffect(() => {
      const currentUser = userStore.username;
      formModel.managers = [currentUser];
      formModel.bak_managers = [currentUser];
    });

    const handleSubmit = async () => {
      try {
        isLoading.value = true;
        await formInstance.value.validate();
        const { data } = await billStore.create_main_account({
          ...formModel,
          email: `${formModel.email}${isuffix}`,
          business_type: formModel.site,
          VerifyCode: dialogForm.value.code,
          extension: {
            [Extension_Name_Map[formModel.vendor]]: formModel.name,
          },
        });
        Message({
          message: '创建成功',
          theme: 'success',
        });
        router.push({
          path: '/service/my-apply/detail',
          query: {
            ...route.query,
            id: data.id,
          },
        });
      } catch (err) {
        // console.log(err);
      } finally {
        isLoading.value = false;
      }
    };

    watch(
      () => formModel.vendor,
      () => {
        if ([VendorEnum.AZURE, VendorEnum.GCP, VendorEnum.HUAWEI].includes(formModel.vendor)) {
          formModel.site = 'international';
        }
      },
    );
    let timer: string | number | NodeJS.Timeout = null;

    // 邮箱验证码
    const countdown = () => {
      if (timer) {
        clearInterval(timer);
      }
      countdownNum.value = 60; // 重置倒计时
      timer = setInterval(() => {
        if (countdownNum.value > 0) {
          countdownNum.value = countdownNum.value - 1;
          if (countdownNum.value === 1) {
            iscountdown.value = false;
          }
        } else {
          clearInterval(timer);
        }
      }, 1000);
    };
    const handleClose = () => {
      isComplete.value = true;
      empty();
      clearValidate();
    };
    const resendButton = () => {
      dialogForm.value.code = '';
      getCode();
      countdown();
    };
    const empty = () => {
      iscountdown.value = true;
      isDialogShow.value = false;
    };
    const handleConfirm = async () => {
      await formCodeRef.value.validate();
      checkingCode(false);
      empty();
      clearValidate();
    };
    const clearValidate = () => {
      nextTick(() => {
        formCodeRef.value.clearValidate();
      });
    };
    const checkingCode = async (isDeleteAfterVerify: boolean) => {
      try {
        const { data } = await billStore.verify_code({
          mail: `${formModel.email}${isuffix}`,
          scene: Scenes.SecondaryAccount,
          verify_code: dialogForm.value.code,
          delete_after_verify: isDeleteAfterVerify,
        });
        VerificationResults.value = data ? 'Pass verified' : 'Fail verified';
      } catch (err) {
        // console.log(err);
      } finally {
        isComplete.value = true;
      }
    };
    const getCode = async () => {
      try {
        await billStore.send_code({
          mail: `${formModel.email}${isuffix}`,
          scene: Scenes.SecondaryAccount,
          info: {
            vendor: formModel.vendor,
            account_name: formModel.name,
          },
        });
      } catch (err) {
        // console.log(err);
      }
    };
    const verification = async () => {
      getCode();
      dialogForm.value.code = '';
      isDialogShow.value = true;
      countdown();
    };
    onUnmounted(() => {
      if (timer) {
        clearInterval(timer);
      }
    });
    return () => (
      <div class={'create-second-account-wrapper'}>
        <DetailHeader class={'header'}>
          <span class={'header-title'}>创建二级账号</span>
        </DetailHeader>
        <ResizeLayout placement='right' initialDivide={'25%'} collapsible>
          {{
            main: () => (
              <div class={'left-container'}>
                <Alert
                  theme='warning'
                  closable
                  title='申请帐号需要进行邮箱配置，请确认邮箱配置完成再提交申请，否则将导致帐号申请失败！'
                  class={'ml24 mr-24'}
                />
                <CommonCard title={() => '基础信息'} class={'info-card'}>
                  <div class={'account-form-card-content'}>
                    <Form formType='vertical' model={formModel} ref={formInstance} auto-check={true}>
                      <FormItem label='云厂商' required property='vendor'>
                        <div class={'account-vendor-selector'}>
                          {MAIN_ACCOUNT_VENDORS.map(({ vendor, name, icon }) =>
                            vendor !== VendorEnum.TCLOUD ? (
                              <div
                                class={`account-vendor-option ${
                                  vendor === formModel.vendor ? 'account-vendor-option-active' : ''
                                }`}
                                onClick={() => (formModel.vendor = vendor)}>
                                <img src={icon} alt={name} class={'account-vendor-option-icon'} />
                                <p class={'account-vendor-option-text'}>{name}</p>
                                {formModel.vendor === vendor ? <Success fill='#3A84FF' class={'active-icon'} /> : null}
                              </div>
                            ) : (
                              <div
                                class={`account-vendor-option disabled-option`}
                                v-bk-tooltips={{
                                  content: (
                                    <span>
                                      腾讯云账号需要到云梯申请，请参考{' '}
                                      <Button text theme='primary'>
                                        腾讯云账号申请指引
                                      </Button>
                                    </span>
                                  ),
                                }}>
                                <img src={icon} alt={name} class={'account-vendor-option-icon'} />
                                <p>{name}</p>
                              </div>
                            ),
                          )}
                        </div>
                      </FormItem>
                      <FormItem label='站点类型' required property='site'>
                        <BkRadioGroup v-model={formModel.site}>
                          <BkRadioButton
                            label='china'
                            v-bk-tooltips={{
                              disabled: ![VendorEnum.AZURE, VendorEnum.GCP, VendorEnum.HUAWEI].includes(
                                formModel.vendor,
                              ),
                              content: '当前微软云、谷歌云、华为云的站点类型默认限制为“国际站”，“中国站”不可选 ',
                            }}
                            disabled={[VendorEnum.AZURE, VendorEnum.GCP, VendorEnum.HUAWEI].includes(formModel.vendor)}>
                            中国站
                          </BkRadioButton>
                          <BkRadioButton label='international'>国际站</BkRadioButton>
                        </BkRadioGroup>
                      </FormItem>
                      {/* <FormItem label='站点地址' required property=''>
                        <Input />
                      </FormItem> */}
                    </Form>
                  </div>
                </CommonCard>
                <CommonCard title={() => '账号信息'} class={'info-card'}>
                  <div class={'account-form-card-content'}>
                    <Form
                      formType='vertical'
                      model={formModel}
                      ref={formModelRef}
                      rules={{
                        name: [
                          {
                            trigger: 'change',
                            message: nameTips.value,
                            validator: (val: string) => {
                              const vendorList = [VendorEnum.AWS, VendorEnum.AZURE, VendorEnum.HUAWEI];
                              const regex = vendorList.includes(formModel.vendor as VendorEnum)
                                ? /^[a-zA-Z][a-zA-Z0-9_]{5,19}$/
                                : /^[a-zA-Z][a-zA-Z0-9-]{5,19}$/;
                              const isValid = regex.test(val);
                              isNameRules.value = isValid;
                              return isValid;
                            },
                          },
                        ],
                        email: emailRules,
                      }}>
                      <FormItem label='账号名称' required property='name' description={nameTips.value}>
                        <Input v-model={formModel.name} placeholder='请输入账号名称'></Input>
                      </FormItem>
                      <FormItem label='账号邮箱' required property='email'>
                        <Input class={'email-input'} v-model={formModel.email} suffix={isuffix}></Input>
                        {countdownNum.value < 60 && iscountdown.value ? (
                          <>
                            <Button theme='primary' style={'width:86px'} disabled={isSendBtnDisabled.value}>
                              {countdownNum.value}s
                            </Button>
                          </>
                        ) : (
                          <>
                            <Button
                              theme='primary'
                              disabled={!isMailRules.value || !isMailRules.value}
                              onClick={verification}
                              v-bk-tooltips={{
                                content: !isNameRules.value
                                  ? !isMailRules.value
                                    ? '请输入账号名称，账号邮箱'
                                    : '请输入账号名称'
                                  : !isMailRules.value && '请输入邮箱',
                                disabled: isMailRules.value && isNameRules.value,
                              }}>
                              {isComplete.value ? ' 重新校验' : '邮箱校验'}
                            </Button>
                          </>
                        )}
                        <p class={'email-tip'}>
                          {VerificationResults.value === 'Not verified' ? (
                            <>
                              <i class={'hcm-icon bkhcm-icon-alert email-tip-icon'}></i>
                              请确保邮箱已按指引配置，否则后续帐号将无法创建
                            </>
                          ) : VerificationResults.value === 'Pass verified' ? (
                            <>
                              <i class={'hcm-icon bkhcm-icon-check-circle-fill email-tip-check'}></i>
                              校验通过
                            </>
                          ) : (
                            <>
                              <i class={'hcm-icon bkhcm-icon-close-circle-fill email-tip-close'}></i>
                              校验失败
                            </>
                          )}
                        </p>
                      </FormItem>
                      {/* <FormItem label='成本评估' required property=''>
                        <div class={'evaluation-wrapper'}>
                          <Input type='number' min={1} class={'mr8'} />
                          <Select filterable={false}>
                            <Option id={'人民币'} name={'人民币'} key={'人民币'}></Option>
                            <Option id={'美元'} name={'美元'} key={'美元'}></Option>
                          </Select>
                          <p class={'evaluation-suffix'}> / 月</p>
                        </div>
                        <Alert theme='danger' class={'mt4'}>
                          {{
                            title: () => (
                              // 缺图标
                              <p>
                                当前金额超过 1,000 美元，需要上传凭证作为审批依据
                                <Button theme='primary' text class={'ml8'}>
                                  上传
                                </Button>
                              </p>
                            ),
                          }}
                        </Alert>
                      </FormItem> */}
                      {/* <FormItem label='业务' required property=''>
                        <BusinessSelector authed autoSelect v-model={formModel.bk_biz_id} />
                      </FormItem> */}

                      <FormItem label='业务' required property=''>
                        <BusinessSelector v-model={formModel.op_product_id} />
                      </FormItem>
                      <div class={'account-manager-wrapper'}>
                        <FormItem label='主负责人' required property='managers' class={'account-manager'}>
                          <MemberSelect
                            v-model={formModel.managers}
                            defaultUserlist={[
                              {
                                username: userStore.username,
                                display_name: userStore.username,
                              },
                            ]}
                          />
                        </FormItem>
                        <FormItem label='备份负责人' required property='bak_managers' class={'ml24 account-manager'}>
                          <MemberSelect
                            v-model={formModel.bak_managers}
                            defaultUserlist={[
                              {
                                username: userStore.username,
                                display_name: userStore.username,
                              },
                            ]}
                          />
                        </FormItem>
                      </div>
                      {/* <FormItem label='所属组织架构' required property='dept_id'>
                        <OrganizationSelect />
                      </FormItem> */}
                      <FormItem label='账号用途' property='memo' required>
                        <Input type='textarea' rows={5} maxlength={100} v-model={formModel.memo} />
                      </FormItem>
                    </Form>
                  </div>
                </CommonCard>
                <Button
                  theme='primary'
                  class={'mr8 ml24 mw88'}
                  onClick={() => handleSubmit()}
                  loading={isLoading.value}>
                  提交
                </Button>
                <Button
                  class='mw88'
                  onClick={() => {
                    router.back();
                  }}>
                  取消
                </Button>
              </div>
            ),
            aside: () => (
              <div class={'right-container'}>
                <div class={'header'}>
                  <p class={'title'}>申请指引</p>
                  <Button text theme='primary' class={'link'}>
                    查看更多
                  </Button>
                </div>
                <div class={'info-block'}>
                  <p class={'info-title'}>申请邮箱</p>
                  <p>此邮箱是用于接收云账号注册信息的专用邮箱，需要开通外部邮件接收权限并配置邮件转发列表</p>
                </div>
              </div>
            ),
          }}
        </ResizeLayout>
        <bk-dialog v-model:is-show={isDialogShow.value} title='邮箱校验' quick-close>
          {{
            default: () => (
              <div>
                <bk-alert
                  class={'dialog-alert'}
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
                        class={'dialog-button'}
                        onClick={resendButton}
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
                <Button theme='primary' disabled={!dialogForm.value.code} onClick={() => handleConfirm()}>
                  提交
                </Button>
                <Button class='dialog-cancel' onClick={() => handleClose()}>
                  取消
                </Button>
              </>
            ),
          }}
        </bk-dialog>
      </div>
    );
  },
});
