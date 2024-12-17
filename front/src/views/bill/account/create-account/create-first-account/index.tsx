import { defineComponent, reactive, ref, watchEffect } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Button, Form, Input, Message, Select } from 'bkui-vue';
import { BILL_VENDORS_INFO } from '../constants';
import { InfoLine, Success } from 'bkui-vue/lib/icon';
import { VendorEnum, AccountVerifyEnum } from '@/common/constant';
import MemberSelect from '@/components/MemberSelect';
import { useUserStore } from '@/store';
import { useRouter } from 'vue-router';
import useBillStore from '@/store/useBillStore';
import successIcon from '@/assets/image/corret-fill.png';
import failedIcon from '@/assets/image/delete-fill.png';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
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
      site: 'china', // 站点
      dept_id: -1, // 组织架构ID
      memo: '', // 备忘录
      extension: {}, // 扩展字段对象
      accountType: AccountVerifyEnum.ROOT,
    });

    watchEffect(() => {
      const currentUser = userStore.username;
      formModel.managers = [currentUser];
      formModel.bak_managers = [currentUser];
    });

    const { curExtension, isValidateDiasbled, handleValidate, isValidateLoading } = useSecretExtension(formModel);

    const handleSubmit = async () => {
      await formRef.value.validate();
      formModel.extension = Object.entries({
        ...curExtension.value.input,
        ...curExtension.value.output1,
        ...curExtension.value.output2,
      }).reduce((prev, [key, { value }]) => {
        prev[key] = value;
        return prev;
      }, {});
      await billStore.root_accounts_add({
        ...formModel,
        email: `${formModel.email}@tencent.com`,
      });
      Message({
        message: '一级账号录入成功',
        theme: 'success',
      });
      router.go(-1);
    };
    const handleChange = (keyVal: any, property: string, follow: string) => {
      curExtension.value.output1[property].value = keyVal;
      const data = curExtension.value.selectParams[property].list.filter((item: any) => item[property] === keyVal);
      curExtension.value.output1[follow].value = data[0][follow];
    };

    return () => (
      <div class={'create-first-account-wrapper'}>
        <DetailHeader>
          <span class={'header-title'}>录入一级账号</span>
        </DetailHeader>

        <CommonCard title={() => '基础信息'} class={'info-card'}>
          <div class={'account-form-card-content'}>
            <Form formType='vertical' model={formModel}>
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
            </Form>
          </div>
        </CommonCard>

        <CommonCard title={() => '账号信息'} class={'info-card'}>
          <div class={'account-form-card-content'}>
            <Form
              formType='vertical'
              model={formModel}
              auto-check
              ref={formRef}
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
              <Form formType='vertical' class={'account-form-card-content-grid'}>
                <div>
                  {Object.entries(curExtension.value.input).map(([property, { label }]) => (
                    <FormItem label={label} property={property} required>
                      <Input
                        v-model={curExtension.value.input[property].value}
                        type={
                          property === 'cloud_service_secret_key' && formModel.vendor === VendorEnum.GCP
                            ? 'textarea'
                            : 'text'
                        }
                        rows={8}
                        resize={!(formModel.vendor === VendorEnum.GCP)}
                      />
                    </FormItem>
                  ))}
                </div>
                <div class={'account-form-card-content-grid-right'}>
                  {Object.entries(curExtension.value.output1).map(([property, { label, value, placeholder }]) => (
                    <FormItem label={label} property={property}>
                      {curExtension.value?.selectType?.includes(property) ? (
                        <Select
                          v-model={value}
                          placeholder={placeholder}
                          list={curExtension.value.selectParams[property].list}
                          idKey={curExtension.value.selectParams[property].idKey}
                          displayKey={curExtension.value.selectParams[property].displayKey}
                          clearable={false}
                          onChange={(val) =>
                            handleChange(val, property, curExtension.value.selectParams[property].follow)
                          }
                        />
                      ) : (
                        <Input v-model={value} readonly placeholder={placeholder} />
                      )}
                    </FormItem>
                  ))}
                </div>
              </Form>
            </div>
            {![VendorEnum.KAOPU, VendorEnum.ZENLAYER].includes(formModel.vendor) && (
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
            )}
          </>
        </CommonCard>

        <Button
          theme='primary'
          class={'mr8 ml24 mw88'}
          disabled={curExtension.value.validatedStatus !== ValidateStatus.YES}
          v-bk-tooltips={{
            disabled: !(curExtension.value.validatedStatus !== ValidateStatus.YES),
            content: 'API密钥校验通过才能提交',
          }}
          onClick={() => {
            handleSubmit();
          }}>
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
    );
  },
});
