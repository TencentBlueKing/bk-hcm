import { Form, Input, Select, Button, Radio } from 'bkui-vue';
import { reactive, defineComponent } from 'vue';
import { ProjectModel, FormItems } from '@/typings';
import { useI18n } from 'vue-i18n';
import MemberSelect from '@/components/MemberSelect';
import OrganizationSelect from '@/components/OrganizationSelect';
const { FormItem } = Form;
const { Option } = Select;
const { Group } = Radio;
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const { t } = useI18n();

    const initProjectModel: ProjectModel = {
      type: 'resource',
      resourceName: '',
      name: '',
      cloudName: 'huawei',
      scretId: '',
      account: '',
      remark: '',
      user: ['poloohuang'],
      business: '',
    };
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

    const department = [6544];

    const check = (): boolean => {
      return false;
    };

    // 提交操作
    const submit = () => {
      console.log(1111);
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
              required: false,
              property: 'masterAccount',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
            {
              label: t('子账号ID'),
              required: false,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
            {
              label: t('子账号名称'),
              required: false,
              property: 'subAccountName',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
            {
              label: t('SecretId/密钥ID'),
              required: true,
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
            {
              label: 'SecretKey',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
          ];
          break;
        case 'AWS':
          console.log(formList);
          insertFormData = [
            {
              label: t('账号ID'),
              required: false,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
            {
              label: t('账号名称'),
              required: false,
              property: 'subAccountName',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
            {
              label: t('IAM用户名称'),
              required: true,
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
            {
              label: t('SecretId/密钥ID'),
              required: true,
              property: 'scretId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
            {
              label: 'SecretKey',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
            },
          ];
          break;
        default:
          break;
      }
      const interceLength = endIndex - startIndex - 1;    // 需要删除的长度
      console.log('insertFormData', insertFormData);
      formList.splice(startIndex + 1, interceLength, ...insertFormData);
    };

    // 表单检验
    const formRules = {
      name: [{ trigger: 'blur', message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾业务与项目至少填一个', validator: check }],
    };
    const formList = reactive<FormItems[]>([
      {
        component: () => <Group v-model={projectModel.type}>
          <Radio label='resource'>{t('资源账号')}</Radio>
          <Radio label='register'>{t('登记账号')}</Radio>
        </Group>,
      },
      {
        label: t('名称'),
        required: true,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('云厂商'),
        required: true,
        property: 'cloudName',
        component: () => <Select class="w450" modelValue={projectModel.cloudName} onChange={changeCloud}>
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
        required: false,
        property: 'masterAccount',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
      },
      {
        label: t('子账号ID'),
        required: false,
        property: 'subAccountId',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
      },
      {
        label: t('子账号名称'),
        required: false,
        property: 'subAccountName',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
      },
      {
        label: t('密钥ID'),
        required: true,
        property: 'scretId',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
      },
      {
        label: 'SecretKey',
        required: true,
        property: 'secretKey',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
      },
      {
        label: t('责任人'),
        required: false,
        property: 'user',
        content: () => (
          <MemberSelect class="w450" v-model={projectModel.user}/>
        ),
      },
      {
        label: t('组织架构'),
        required: false,
        property: 'account',
        content: () => (
          <OrganizationSelect class="w450" v-model={department} />
        ),
      },
      {
        label: t('使用业务'),
        required: false,
        property: 'account',
        component: () => <Select multiple show-select-all collapse-tags multipleMode='tag' class="w450" modelValue={projectModel.cloudName}>
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
        label: t('描述'),
        required: false,
        property: 'remark',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} type="textarea" maxlength={100} showWordLimit rows={2} />,
      },
      {
        component: () => <Button class="w90" theme="primary" onClick={submit}>{t('确认')}</Button>,
      },
    ]);

    return () => (
      <Form model={projectModel} labelWidth={100} rules={formRules}>
      {formList.map(item => (
        <FormItem label={item.label} required={item.required} property={item.property}>
          {item.component ? item.component() : item.content()}
        </FormItem>
      ))}
    </Form>
    );
  },
});
