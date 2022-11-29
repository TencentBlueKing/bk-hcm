import { Form, Input, Select, Button } from 'bkui-vue';
import { reactive, defineComponent } from 'vue';
import { useRoute } from 'vue-router';
import { ProjectModel } from '@/typings';
import { useI18n } from 'vue-i18n';
import MemberSelect from '@/components/MemberSelect';
import OrganizationSelect from '@/components/OrganizationSelect';
const { FormItem } = Form;
const { Option } = Select;
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const { t } = useI18n();
    const route = useRoute();

    const { id } = route.query;
    if (id) {
      route.meta.breadcrumb = ['云管', '账户', '编辑账户']; // 未完善 需添加进状态管理区
    }

    const initProjectModel: ProjectModel = {
      resourceName: '',
      name: '',
      cloudName: '',
      scretId: '',
      account: '',
      remark: '',
      user: [],
      business: '',
    };
    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const cloudType = [
      { key: '华为云', value: 'huawei' },
    ];

    const members = ['poloohuang'];
    const department = [6544];

    const check = (): boolean => {
      return false;
    };

    const submit = () => {
      console.log(1111);
    };

    const formRules = {
      name: [{ trigger: 'blur', message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾业务与项目至少填一个', validator: check }],
    };

    const formItems = [
      {
        label: t('名称'),
        required: true,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('云厂商'),
        required: true,
        property: 'name',
        component: () => <Select class="w450" modelValue={projectModel.cloudName}>
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
        property: 'account',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.account} />,
      },
      {
        label: t('子账号ID'),
        required: false,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('子账号名称'),
        required: false,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('密钥ID'),
        required: true,
        property: 'scretId',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.scretId} />,
      },
      {
        label: 'SecretKey',
        required: true,
        property: 'secretKey',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('责任人'),
        required: false,
        property: 'name',
        content: () => (
          <MemberSelect class="w450" v-model={members}/>
        ),
      },
      {
        label: t('组织架构'),
        required: false,
        property: 'name',
        content: () => (
          <OrganizationSelect class="w450" v-model={department} />
        ),
      },
      {
        label: t('使用业务'),
        required: false,
        property: 'name',
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
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} type="textarea" maxlength={100} showWordLimit rows={2} />,
      },
      {
        required: false,
        property: 'name',
        component: () => <Button class="w90" theme="primary" onClick={submit}>{t('确认')}</Button>,
      },
    ];

    return () => (
      <Form model={projectModel} labelWidth={100} rules={formRules}>
      {formItems.map(item => (
        <FormItem label={item.label} required={item.required} property={item.property}>
          {item.component ? item.component() : item.content()}
        </FormItem>
      ))}
    </Form>
    );
  },
});
