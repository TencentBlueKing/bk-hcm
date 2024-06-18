import { defineComponent, ref } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Alert, Button, Form, Input, Message, ResizeLayout } from 'bkui-vue';
import { VendorEnum } from '@/common/constant';
import { BILL_VENDORS_INFO } from '../constants';
import { Success } from 'bkui-vue/lib/icon';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import MemberSelect from '@/components/MemberSelect';
import { useUserStore } from '@/store';
import BusinessSelector from '@/components/business-selector/index.vue';
import useFormModel from '@/hooks/useFormModel';
import useBillStore from '@/store/useBillStore';
import { Extension_Name_Map } from './constants';
import { useRouter } from 'vue-router';

const { FormItem } = Form;

export default defineComponent({
  setup() {
    const userStore = useUserStore();
    const formInstance = ref();
    const isLoading = ref(false);
    const billStore = useBillStore();
    const router = useRouter();
    const { formModel } = useFormModel({
      name: '', // 名字
      vendor: VendorEnum.AZURE, // 云厂商
      email: '', // 邮箱
      managers: [userStore.username], // 负责人数组
      bak_managers: [userStore.username], // 备份负责人数组
      site: 'china', // 站点
      // dept_id: '', // 组织架构ID
      memo: '', // 备忘录
      bk_biz_id: '', // 业务
      // extension: {}, // 扩展字段对象
    });

    const handleSubmit = async () => {
      try {
        isLoading.value = true;
        await formInstance.value.validate();
        await billStore.create_main_account({
          ...formModel,
          business_type: formModel.site,
          extension: {
            [Extension_Name_Map[formModel.vendor]]: formModel.name,
          },
        });
        Message({
          message: '创建成功',
          theme: 'success',
        });
        router.go(-1);
      } catch (err) {
        // console.log(err);
      } finally {
        isLoading.value = false;
      }
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
                <CommonCard title={() => '基础信息'} class={'info-card'}>
                  <div class={'account-form-card-content'}>
                    <Form formType='vertical' model={formModel} ref={formInstance} auto-check={true}>
                      <FormItem label='云厂商' required property='vendor'>
                        <div class={'account-vendor-selector'}>
                          {BILL_VENDORS_INFO.map(({ vendor, name, icon }) => (
                            <div
                              class={`account-vendor-option ${
                                vendor === formModel.vendor ? 'account-vendor-option-active' : ''
                              }`}
                              onClick={() => (formModel.vendor = vendor)}>
                              <img src={icon} alt={name} class={'account-vendor-option-icon'} />
                              <p class={'account-vendor-option-text'}>{name}</p>
                              {formModel.vendor === vendor ? <Success fill='#3A84FF' class={'active-icon'} /> : null}
                            </div>
                          ))}
                        </div>
                      </FormItem>
                      <FormItem label='站点类型' required property='site'>
                        <BkRadioGroup v-model={formModel.site}>
                          <BkRadioButton label='china'>中国站</BkRadioButton>
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
                      rules={{
                        name: [
                          {
                            trigger: 'change',
                            message: '账号名称只能包括小写字母和数字，并且仅能以小写字母开头，长度为6-20个字符',
                            validator: (val: string) => {
                              return /^[a-z][a-z0-9]{5,19}$/.test(val);
                            },
                          },
                        ],
                      }}>
                      <FormItem
                        label='帐号名称'
                        required
                        property='name'
                        description='账号名称只能包括小写字母和数字，并且仅能以小写字母开头，长度为6-20个字符'>
                        <Input v-model={formModel.name} placeholder='请输入账号名称'></Input>
                      </FormItem>
                      <FormItem label='帐号邮箱' required property='email'>
                        <Input v-model={formModel.email} suffix='@tencent.com'></Input>
                        <p class={'email-tip'}>
                          <i class={'hcm-icon bkhcm-icon-alert email-tip-icon'}></i>
                          请确保邮箱已按指引配置，否则后续帐号将无法创建
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
                      <FormItem label='运营产品' required property=''>
                        <BusinessSelector authed autoSelect v-model={formModel.bk_biz_id} />
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
                <Button theme='primary' class={'mr8 ml24'} onClick={() => handleSubmit()}>
                  提交
                </Button>
                <Button>取消</Button>
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
