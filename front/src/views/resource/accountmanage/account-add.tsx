import { Form, Input, Select, Button, Radio, Message } from 'bkui-vue';
import { reactive, defineComponent, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { ProjectModel, FormItems } from '@/typings';
import { useI18n } from 'vue-i18n';
import MemberSelect from '@/components/MemberSelect';
import OrganizationSelect from '@/components/OrganizationSelect';
import { useAccountStore } from '@/store';
const { FormItem } = Form;
const { Option } = Select;
const { Group } = Radio;
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const { t } = useI18n();
    const accountStore = useAccountStore();
    const router = useRouter();

    const initProjectModel: ProjectModel = {
      type: 'resource',   // 账号类型
      name: '', // 名称
      cloudName: '', // 云厂商
      account: '',    // 主账号
      subAccountId: '',    // 子账号id
      subAccountName: '',    // 子账号名称
      scretId: '',    // 密钥id
      secretKey: '',  // 密钥key
      user: ['poloohuang'], // 责任人
      organize: [6544],   // 组织架构
      business: '',   // 使用业务
      remark: '',     // 备注
    };
    const formRef = ref<InstanceType<typeof Form>>(null);
    const noUser = ref<Boolean>(false);
    const noOrganize = ref<Boolean>(false);
    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const cloudType = reactive([
      { key: '华为云', value: 'huawei' },
      { key: '亚马逊', value: 'AWS' },
      { key: '谷歌云', value: 'GCP' },
      { key: '微软云', value: 'Azure' },
      { key: '腾讯云', value: 'tx' },
    ]);

    const businessList = reactive([
      { key: '华为云', value: 'huawei' },
      { key: '亚马逊', value: 'AWS' },
      { key: '谷歌云', value: 'GCP' },
      { key: '微软云', value: 'Azure' },
      { key: '腾讯云', value: 'tx' },
    ]);

    const optionalRequired = ['scretId', 'secretKey'];

    const check = (): boolean => {
      return true;
    };

    // 提交操作
    const submit = async () => {
      await formRef.value?.validate();
      try {
        await accountStore.addAccount(projectModel);
        Message({
          message: t('新增成功'),
          theme: 'success',
        });
        router.go(-1);  // 返回列表
      } catch (error) {
        console.log(error);
      }
    };

    // // 修改表单
    // const updateFormItems = (startIndex: number, endIndex: number) => {
    //   console.log(startIndex, endIndex);
    //   const interceLength = endIndex - startIndex - 1;    // 需要删除的长度
    //   formList.splice(startIndex + 1, interceLength);
    // };

    const changeCloud = (val: string) => {
      console.log(1111, val, projectModel);
      const startIndex = formList.findIndex(e => e.property === 'cloudName');
      const endIndex = formList.findIndex(e => e.property === 'user');
      let insertFormData: any = [];
      switch (val) {
        case 'huawei':
          insertFormData = [
            {
              label: t('主账号'),
              required: true,
              property: 'account',
              component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.account} />,
            },
            {
              label: t('子账号ID'),
              required: true,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入子账号ID')} v-model={projectModel.subAccountId} />,
            },
            {
              label: t('子账号名称'),
              required: true,
              property: 'subAccountName',
              component: () => <Input class="w450" placeholder={t('请输入子账号名称')} v-model={projectModel.subAccountName} />,
            },
            {
              label: t('SecretId/密钥ID'),
              required: true,
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.scretId} />,
            },
            {
              label: 'SecretKey',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />,
            },
          ];
          break;
        case 'AWS':
          insertFormData = [
            {
              label: t('账号ID'),
              required: true,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入账号ID')} v-model={projectModel.account} />,
            },
            {
              label: t('账号名称'),
              required: true,
              property: 'subAccountName',
              component: () => <Input class="w450" placeholder={t('请输入账号名称')} v-model={projectModel.account} />,
            },
            {
              label: t('IAM用户名称'),
              required: true,
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入IAM用户名称')} v-model={projectModel.account} />,
            },
            {
              label: t('SecretId/密钥ID'),
              required: projectModel.type === 'resource',
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.account} />,
            },
            {
              label: 'SecretKey',
              required: projectModel.type === 'resource',
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入请输入主账号')} v-model={projectModel.account} />,
            },
          ];
          break;
        case 'GCP':
          insertFormData = [
            {
              label: t('项目 ID'),
              required: true,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入项目 ID')} v-model={projectModel.account} />,
            },
            {
              label: t('项目名称'),
              required: true,
              property: 'subAccountName',
              component: () => <Input class="w450" placeholder={t('请输入项目名称')} v-model={projectModel.account} />,
            },
            {
              label: t('服务账号ID'),
              required: projectModel.type === 'resource',
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入服务账号ID')} v-model={projectModel.account} />,
            },
            {
              label: t('服务账号名称'),
              required: projectModel.type === 'resource',
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入服务账号名称')} v-model={projectModel.account} />,
            },
            {
              label: '服务账号密钥ID',
              required: projectModel.type === 'resource',
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入服务账号密钥ID')} v-model={projectModel.account} />,
            },
            {
              label: '服务账号密钥',
              required: projectModel.type === 'resource',
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入服务账号密钥')} v-model={projectModel.account} />,
            },
          ];
          break;
        case 'Azure':
          insertFormData = [
            {
              label: t('租户 ID'),
              required: true,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入租户 ID')} v-model={projectModel.account} />,
            },
            {
              label: t('订阅 ID'),
              required: true,
              property: 'subAccountName',
              component: () => <Input class="w450" placeholder={t('请输入订阅 ID')} v-model={projectModel.account} />,
            },
            {
              label: t('订阅名称'),
              required: true,
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入订阅名称')} v-model={projectModel.account} />,
            },
            {
              label: t('应用程序(客户端) ID'),
              required: projectModel.type === 'resource',
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入应用程序(客户端) ID')} v-model={projectModel.account} />,
            },
            {
              label: '应用程序名称',
              required: projectModel.type === 'resource',
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入应用程序名称')} v-model={projectModel.account} />,
            },
            {
              label: '客户端密钥ID',
              required: projectModel.type === 'resource',
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入客户端密钥ID')} v-model={projectModel.account} />,
            },
            {
              label: '客户端密钥',
              required: projectModel.type === 'resource',
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入客户端密钥')} v-model={projectModel.account} />,
            },
          ];
          break;
        case 'tx':
          insertFormData = [
            {
              label: t('主账号'),
              required: true,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.account} />,
            },
            {
              label: t('子账号'),
              required: projectModel.type === 'resource',
              property: 'subAccountName',
              component: () => <Input class="w450" placeholder={t('请输入子账号')} v-model={projectModel.account} />,
            },
            {
              label: 'SecretId',
              required: projectModel.type === 'resource',
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入SecretId')} v-model={projectModel.account} />,
            },
          ];
          break;
        default:
          insertFormData = [
            {
              label: t('主账号'),
              required: true,
              property: 'masterAccount',
              component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.account} />,
            },
            {
              label: t('子账号ID'),
              required: true,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入子账号ID')} v-model={projectModel.account} />,
            },
            {
              label: t('子账号名称'),
              required: true,
              property: 'subAccountName',
              component: () => <Input class="w450" placeholder={t('请输入子账号名称')} v-model={projectModel.account} />,
            },
            {
              label: t('SecretId/密钥ID'),
              required: true,
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.account} />,
            },
            {
              label: 'SecretKey',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入SecretKey')} v-model={projectModel.account} />,
            },
          ];
          break;
      }
      const interceLength = endIndex - startIndex - 1;    // 需要删除的长度
      console.log('insertFormData', insertFormData);
      formList.splice(startIndex + 1, interceLength, ...insertFormData);
    };

    // 表单检验
    const formRules = {
      name: [
        { trigger: 'blur', message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾业务与项目至少填一个', validator: check },
      ],
    };

    watch(
      () => projectModel.type,
      (val) => {
        formRef.value?.clearValidate(); // 切换清除表单检验
        if (val === 'register') {  // 登记账号
          formList?.forEach((e) => {
            if (optionalRequired.includes(e.property)) {
              e.required = false;
            }
          });
        } else {
          formList?.forEach((e) => {
            if (e.label && e.property !== 'remark') {   // 备注不需必填
              e.required = true;
            }
          });
        }
      },
      { immediate: true },
    );


    const formList = reactive<FormItems[]>([
      {
        required: false,
        component: () => <Group v-model={projectModel.type}>
          <Radio label='resource'>{t('资源账号')}</Radio>
          <Radio label='register'>{t('登记账号')}</Radio>
        </Group>,
      },
      {
        label: t('名称'),
        required: true,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入名称')} v-model={projectModel.name} />,
      },
      {
        label: t('云厂商'),
        required: true,
        property: 'cloudName',
        component: () => <Select class="w450" placeholder={t('请选择云厂商')} v-model={projectModel.cloudName} onChange={changeCloud}>
          {cloudType.map(item => (
              <Option
                key={item.key}
                value={item.value}
                label={item.key}
              >
                {item.key}
              </Option>
          ))
        }</Select>,
      },
      {
        label: t('主账号'),
        required: true,
        property: 'masterAccount',
        component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.account} />,
      },
      {
        label: t('子账号ID'),
        required: true,
        property: 'subAccountId',
        component: () => <Input class="w450" placeholder={t('请输入子账号ID')} v-model={projectModel.account} />,
      },
      {
        label: t('子账号名称'),
        required: true,
        property: 'subAccountName',
        component: () => <Input class="w450" placeholder={t('请输入子账号名称')} v-model={projectModel.account} />,
      },
      {
        label: t('SecretId/密钥ID'),
        required: true,
        property: 'scretId',
        component: () => <Input class="w450" placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.account} />,
      },
      {
        label: 'SecretKey',
        required: true,
        property: 'secretKey',
        component: () => <Input class="w450" placeholder={t('请输入SecretKey')} v-model={projectModel.account} />,
      },
      {
        label: t('责任人'),
        required: true,
        property: 'user',
        content: () => (
          <section>
            <MemberSelect class="w450" v-model={projectModel.user}/>
            {noUser.value ? <span class="form-error-tip">责任人不能为空</span> : ''}
          </section>
        ),
      },
      {
        label: t('组织架构'),
        required: true,
        property: 'organize',
        content: () => (
          <section>
            <OrganizationSelect class="w450" v-model={projectModel.organize} />
            {noOrganize.value ? <span class="form-error-tip">组织架构不能为空</span> : ''}
          </section>
        ),
      },
      {
        label: t('使用业务'),
        required: true,
        property: 'business',
        component: () => <Select multiple show-select-all collapse-tags multipleMode='tag'
        placeholder={t('请选择使用业务')} class="w450" v-model={projectModel.business}>
          {businessList.map(item => (
              <Option
                key={item.key}
                value={item.value}
                label={item.key}
              >
                {item.key}
              </Option>
          ))
        }</Select>,
      },
      {
        label: t('备注'),
        required: false,
        property: 'remark',
        component: () => <Input class="w450" placeholder={t('请输入备注')} v-model={projectModel.remark} type="textarea" maxlength={100} showWordLimit rows={2} />,
      },
      {
        required: false,
        component: () => <Button class="w90" theme="primary" onClick={submit}>{t('确认')}</Button>,
      },
    ]);

    return () => (
      <Form model={projectModel} labelWidth={140} rules={formRules} ref={formRef}>
      {formList.map(item => (
        <FormItem label={item.label} required={item.required} property={item.property}>
          {item.component ? item.component() : item.content()}
        </FormItem>
      ))}
    </Form>
    );
  },
});
