import { defineComponent, reactive, ref } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Button, Form, Input } from 'bkui-vue';
import { BILL_VENDORS_INFO } from '../constants';
import { InfoLine, Success } from 'bkui-vue/lib/icon';
import { VendorEnum } from '@/common/constant';
import MemberSelect from '@/components/MemberSelect';
import { useUserStore } from '@/store';
import { useRouter } from 'vue-router';
import useBillStore from '@/store/useBillStore';
import successIcon from '@/assets/image/corret-fill.png';
import failedIcon from '@/assets/image/delete-fill.png';
import {
  ValidateStatus,
  useSecretExtension,
} from '@/views/resource/resource-manage/account/createAccount/components/accountForm/useSecretExtension';

const { FormItem } = Form;

export default defineComponent({
  setup() {
    const userStore = useUserStore();
    const router = useRouter();
    const billStore = useBillStore();
    const formRef = ref();

    const formModel = reactive({
      name: '', // 名字
      vendor: VendorEnum.AZURE, // 云厂商
      email: '', // 邮箱
      managers: [], // 负责人数组
      bak_managers: [], // 备份负责人数组
      site: '', // 站点
      dept_id: '', // 组织架构ID
      memo: '', // 备忘录
      extension: {}, // 扩展字段对象
    });

    // const resetFormModel = () => {
    //   formModel.name = '';
    //   formModel.vendor = VendorEnum.GCP;
    //   formModel.email = '';
    //   formModel.managers = [];
    //   formModel.bak_managers = [];
    //   formModel.site = '';
    //   formModel.dept_id = '';
    //   formModel.memo = '';
    //   formModel.extension = {};
    // };

    const { curExtension, isValidateDiasbled, handleValidate, isValidateLoading, extensionPayload } =
      useSecretExtension(formModel);

    const handleSubmit = async () => {
      await formRef.value.validate();
      await billStore.root_accounts_add({
        ...formModel,
        extension: extensionPayload.value,
      });
    };

    return () => (
      <div class={'create-first-account-wrapper'}>
        <DetailHeader>
          <span class={'header-title'}>录入一级账号</span>
        </DetailHeader>

        <CommonCard title={() => '基础信息'} class={'info-card'}>
          <div class={'account-form-card-content'}>
            <Form formType='vertical' model={formModel} ref={formRef}>
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
            </Form>
          </div>
        </CommonCard>

        <CommonCard title={() => '账号信息'} class={'info-card'}>
          <div class={'account-form-card-content'}>
            <Form formType='vertical' model={formModel} auto-check>
              <FormItem label='帐号名称' required property='name'>
                <Input v-model={formModel.name} placeholder='请输入账号名称'></Input>
              </FormItem>
              <FormItem label='帐号邮箱' required property='email'>
                <Input v-model={formModel.email} suffix='@tencent.com'></Input>
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
                  <MemberSelect v-model={formModel.bak_managers} />
                </FormItem>
              </div>
              {/* <FormItem label='所属组织架构' required property='dept_id'>
                <OrganizationSelect />
              </FormItem> */}
              <FormItem label='备注' property='memo'>
                <Input type='textarea' rows={5} maxlength={100} v-model={formModel.memo} />
              </FormItem>
            </Form>
          </div>
        </CommonCard>

        <CommonCard
          title={() => (
            <div class={'api-secret-header'}>
              <p class={'account-form-card-title'}>API 密钥</p>
              <InfoLine fill='#979BA5' />
              <p class={'header-text'}>同一个主账号下,只允许接入一次。如后续对API密钥更新,必须是隶属于同一主账号。</p>
            </div>
          )}
          class={'info-card'}>
          <>
            <div class={'account-form-card-content'}>
              <Form labelWidth={130} ref={formRef} formType='vertical'>
                {Object.entries(curExtension.value.input).map(([property, { label }]) => (
                  <FormItem label={label} property={property}>
                    <Input
                      v-model={curExtension.value.input[property].value}
                      type={
                        property === 'cloud_service_secret_key' && formModel.vendor === VendorEnum.GCP
                          ? 'textarea'
                          : 'text'
                      }
                      rows={8}
                    />
                  </FormItem>
                ))}
                {[curExtension.value.output1, curExtension.value.output2].map((output) =>
                  Object.entries(output).map(([property, { label, placeholder }]) => (
                    <FormItem label={label} property={property}>
                      <Input v-model={output[property].value} readonly placeholder={placeholder} />
                    </FormItem>
                  )),
                )}
              </Form>
            </div>
            <div class={'validate-btn-block'}>
              <Button
                theme='primary'
                outline={curExtension.value.validatedStatus === ValidateStatus.YES}
                class={'account-validate-btn'}
                onClick={() => handleValidate()}
                disabled={isValidateDiasbled.value}
                loading={isValidateLoading.value}>
                账号校验
              </Button>
              {curExtension.value.validatedStatus === ValidateStatus.YES ? (
                <>
                  <img src={successIcon} alt='success' class={'validate-icon'}></img>
                  <span> 校验成功 </span>
                </>
              ) : null}
              {curExtension.value.validatedStatus === ValidateStatus.NO ? (
                <>
                  <img src={failedIcon} alt='success' class={'validate-icon'}></img>
                  <span> 校验失败 {curExtension.value.validateFailedReason}</span>
                </>
              ) : null}
            </div>
          </>
        </CommonCard>

        <Button
          theme='primary'
          class={'mr8 ml24'}
          onClick={() => {
            handleSubmit();
          }}>
          提交
        </Button>
        <Button
          onClick={() => {
            router.back();
          }}>
          取消
        </Button>
      </div>
    );
  },
});
