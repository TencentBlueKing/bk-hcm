import { computed, defineComponent, nextTick, ref, watch, watchEffect } from 'vue';
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
import { Extension_Name_Map } from './constants';
import { useRoute, useRouter } from 'vue-router';
import BusinessSelector from '@/components/business-selector/index.vue';
import { PluginHandlerMailbox } from '@/plugin-handler/create-account-mail-suffix';
import EmailInput from './create-section-email';
import { MENU_SERVICE_TICKET_DETAILS } from '@/constants/menu-symbol';
const { FormItem } = Form;
export default defineComponent({
  setup() {
    const router = useRouter();
    const route = useRoute();
    const userStore = useUserStore();
    const billStore = useBillStore();

    const { suffixText, emailRules, isMailValid } = PluginHandlerMailbox;

    const formInstance = ref();
    const { formModel } = useFormModel({
      name: '', // 名字
      vendor: VendorEnum.GCP, // 云厂商
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
      let tip = '账号名称只能包括英文字母、数字和中划线-，并且仅能以英文字母开头，长度为6-30个字符';
      switch (formModel.vendor) {
        case VendorEnum.AWS:
        case VendorEnum.HUAWEI:
        case VendorEnum.AZURE:
          tip = '账号名称只能包括英文字母、数字和下划线_，并且仅能以英文字母开头，长度为6-30个字符';
          break;
      }
      return tip;
    });
    const rules = computed(() => {
      return {
        name: [
          {
            trigger: 'change',
            message: nameTips.value,
            validator: (val: string) => {
              const vendorList = [VendorEnum.AWS, VendorEnum.AZURE, VendorEnum.HUAWEI];
              const regex = vendorList.includes(formModel.vendor as VendorEnum)
                ? /^[a-zA-Z][a-zA-Z0-9_]{5,29}$/
                : /^[a-zA-Z][a-zA-Z0-9-]{5,29}$/;
              const isValid = regex.test(val);

              emailInputRef.value?.changeNameValid(isValid);
              return isValid;
            },
          },
        ],
        email: emailRules,
      };
    });

    watchEffect(() => {
      const currentUser = userStore.username;
      formModel.managers = [currentUser];
      formModel.bak_managers = [currentUser];
    });

    const handleVendorChange = (vendor: VendorEnum) => {
      formModel.vendor = vendor;
      nextTick(() => formInstance.value?.clearValidate());
    };

    const isLoading = ref(false);
    const handleSubmit = async () => {
      try {
        isLoading.value = true;
        await formInstance.value.validate();
        const { data } = await billStore.create_main_account({
          ...formModel,
          email: `${formModel.email}${suffixText}`,
          business_type: formModel.site,
          extension: {
            [Extension_Name_Map[formModel.vendor]]: formModel.name,
          },
        });
        Message({
          message: '创建成功',
          theme: 'success',
        });
        router.push({
          name: MENU_SERVICE_TICKET_DETAILS,
          query: {
            ...route.query,
            id: data.id,
          },
        });
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
    const emailInputRef = ref();
    const changeEmail = (value: string) => {
      formModel.email = value;
    };

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
                <Form formType='vertical' model={formModel} ref={formInstance} auto-check={true} rules={rules.value}>
                  <CommonCard title={() => '基础信息'} class={'info-card'}>
                    <div class={'account-form-card-content'}>
                      <FormItem label='云厂商' required property='vendor'>
                        <div class={'account-vendor-selector'}>
                          {MAIN_ACCOUNT_VENDORS.map(({ vendor, name, icon }) =>
                            [VendorEnum.GCP, VendorEnum.HUAWEI, VendorEnum.AWS].includes(vendor) ? (
                              <div
                                class={`account-vendor-option ${
                                  vendor === formModel.vendor ? 'account-vendor-option-active' : ''
                                }`}
                                onClick={() => handleVendorChange(vendor)}>
                                <img src={icon} alt={name} class={'account-vendor-option-icon'} />
                                <p class={'account-vendor-option-text'}>{name}</p>
                                {formModel.vendor === vendor ? <Success fill='#3A84FF' class={'active-icon'} /> : null}
                              </div>
                            ) : (
                              <div
                                class={`account-vendor-option disabled-option`}
                                v-bk-tooltips={{
                                  content: '该云厂商的账号创建暂不支持',
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
                    </div>
                  </CommonCard>
                  <CommonCard title={() => '账号信息'} class={'info-card'}>
                    <div class={'account-form-card-content'}>
                      <FormItem label='账号名称' required property='name' description={nameTips.value}>
                        <Input v-model={formModel.name} placeholder='请输入账号名称'></Input>
                      </FormItem>
                      <FormItem label='账号邮箱' required property='email'>
                        <EmailInput
                          ref={emailInputRef}
                          suffixText={suffixText}
                          isMailValid={isMailValid.value}
                          formModel={formModel}
                          onChangeEmail={changeEmail}
                        />
                      </FormItem>

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
                        <Input type='textarea' rows={5} maxlength={200} v-model={formModel.memo} />
                      </FormItem>
                    </div>
                  </CommonCard>
                  <Button theme='primary' class={'mr8 ml24 mw88'} onClick={handleSubmit} loading={isLoading.value}>
                    提交
                  </Button>
                  <Button class='mw88' disabled={isLoading.value} onClick={() => router.back()}>
                    取消
                  </Button>
                </Form>
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
      </div>
    );
  },
});
