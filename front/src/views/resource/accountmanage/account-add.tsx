import { Form, Input, Select, Button } from 'bkui-vue';
import { reactive, defineComponent } from 'vue';
import { ProjectModel } from '@/typings';
import { useI18n } from 'vue-i18n';
// import MemberSelect from '@/components/MemberSelect';
const { FormItem } = Form;
const { Option } = Select;
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const { t } = useI18n();

    const initProjectModel: ProjectModel = {
      resourceName: '',
      name: '',
      cloudName: '',
    };
    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const cloudType = [
      { key: '华为云', value: 'huawei' },
    ];


    const submit = () => {
      console.log(1111);
    };

    const formItems = [
      {
        label: t('名称'),
        required: false,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('云厂商'),
        required: false,
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
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
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
        required: false,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: 'SecretKey',
        required: false,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('责任人'),
        required: false,
        property: 'name',
        // content: () => (
        //   <MemberSelect
        //     disabled={!currentProject.value?.permissions?.project_permission_manage}
        //     v-model={permissionModel.value[Role.ADMIN][PermissionType.USER].members}
        //   />
        // ),
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('组织架构'),
        required: false,
        property: 'name',
        component: () => <Input class="w450" placeholder={t('请输入')} v-model={projectModel.name} />,
      },
      {
        label: t('使用业务'),
        required: false,
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
      <Form model={projectModel} labelWidth={100}>
      {formItems.map(item => (
        <FormItem label={item.label} required={item.required} property={item.property}>
          {item.component()}
        </FormItem>
      ))}
    </Form>
    );
  },
});
