import { defineComponent, reactive } from 'vue';
import './index.scss';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import CommonCard from '@/components/CommonCard';
import { Button, Form, Input } from 'bkui-vue';
import { VENDORS_INFO } from '../constants';
import { InfoLine, Success } from 'bkui-vue/lib/icon';
import { VendorEnum } from '@/common/constant';
import MemberSelect from '@/components/MemberSelect';
import { useUserStore } from '@/store';
import OrganizationSelect from '@/components/OrganizationSelect';
import { useRouter } from 'vue-router';

const { FormItem } = Form;

export default defineComponent({
  setup() {
    const userStore = useUserStore();
    const router = useRouter();

    const formModel = reactive({
      name: '', // 名字
      vendor: VendorEnum.GCP, // 云厂商
      email: '', // 邮箱
      managers: [], // 负责人数组
      bak_managers: [], // 备份负责人数组
      site: '', // 站点
      dept_id: '', // 组织架构ID
      memo: '', // 备忘录
      extension: {}, // 扩展字段对象
    });

    const resetFormModel = () => {
      formModel.name = '';
      formModel.vendor = VendorEnum.GCP;
      formModel.email = '';
      formModel.managers = [];
      formModel.bak_managers = [];
      formModel.site = '';
      formModel.dept_id = '';
      formModel.memo = '';
      formModel.extension = {};
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
                  {VENDORS_INFO.map(({ vendor, name, icon }) => (
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
            <Form formType='vertical' model={formModel}>
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
              <FormItem label='所属组织架构' required property='dept_id'>
                <OrganizationSelect/>
              </FormItem>
              <FormItem label='备注' property='memo'>
                  <Input type='textarea' rows={5} maxlength={100} v-model={formModel.memo}/>
              </FormItem>
            </Form>
          </div>
        </CommonCard>

        <CommonCard title={() => (
          <div class={'api-secret-header'}>
          <p class={'account-form-card-title'}>API 密钥</p>
          <InfoLine fill='#979BA5' />
          <p class={'header-text'}>同一个主账号下,只允许接入一次。如后续对API密钥更新,必须是隶属于同一主账号。</p>
        </div>
        )} class={'info-card'}>
          不同云字段不一样
        </CommonCard>

        <Button theme='primary' class={'mr8 ml24'}>
          提交
        </Button>
        <Button onClick={() => {
          router.back();
        }}>取消</Button>
      </div>
    );
  },
});
