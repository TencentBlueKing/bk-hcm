import { Form } from 'bkui-vue';
// import { Form, Input, Select, Button } from 'bkui-vue';
import { reactive, defineComponent, ref } from 'vue';
import { ProjectModel } from '@/typings';
import { useI18n } from 'vue-i18n';
// import MemberSelect from '@/components/MemberSelect';
// import OrganizationSelect from '@/components/OrganizationSelect';
import RenderDetailEdit from '@/components/RenderDetailEdit';
import './account-detail.scss';
const { FormItem } = Form;
// const { Option } = Select;
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const { t } = useI18n();
    const formRef = ref<InstanceType<typeof Form>>(null);

    const initProjectModel: ProjectModel = {
      resourceName: '',
      name: '中文',
      cloudName: '',
      scretId: '',
      account: '',
    };
    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const isEdit = ref(true);
    console.log('isEdit', isEdit.value);

    // const cloudType = [
    //   { key: '华为云', value: 'huawei' },
    // ];

    // const members = ['poloohuang'];
    // const department = [6544];

    const check = (val: any): boolean => {
      console.log(11111, val, /^[a-z][a-z-z0-9_-]*$/.test(val));
      return  /^[a-z][a-z-z0-9_-]*$/.test(val);
    };

    const formRules = {
      name: [{ trigger: 'blur', message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾业务与项目至少填一个', validator: check }],
    };

    const handleblur = async () => {
      console.log('check222', 111122323);
      await formRef.value?.validate();
      console.log('过');
      isEdit.value = false;
    };

    const formBaseInfo = reactive([
      {
        name: '基本信息:',
        data: [
          {
            label: t('云厂商:'),
            required: false,
            property: 'cloudType',
            component: () => <span>{t('腾讯云')}</span>,
          },
          {
            label: t('账号类别:'),
            required: false,
            property: 'accountType',
            component: () => <span>{t('资源账号')}</span>,
          },
          {
            label: 'ID:',
            required: false,
            property: 'id',
            component: () => <span>qcloud-for-lol</span>,
          },
          {
            label: t('名称:'),
            required: false,
            property: 'name',
            component: () => {
              return (<RenderDetailEdit v-model={projectModel.name} fromKey="name" isEdit={isEdit.value} onBlur={handleblur}/>);
            },
          },
          {
            label: t('主账号:'),
            required: false,
            property: 'account',
            component: () => <span>23445</span>,
          },
          {
            label: t('负责人:'),
            required: false,
            property: 'user',
            component: () => {
              return (<RenderDetailEdit v-model={projectModel.name} fromKey="name" isEdit={isEdit.value} onBlur={handleblur}/>);
            },
          },
          {
            label: t('余额:'),
            required: false,
            property: 'money',
            component: () => <span>1234</span>,
          },
          {
            label: t('创建人:'),
            required: false,
            property: 'creator',
            component: () => <span>dommy</span>,
          },
          {
            label: t('创建时间:'),
            required: false,
            property: 'create-time',
            component: () => <span>2022-09-03 13：09</span>,
          },
          {
            label: t('修改人:'),
            required: false,
            property: 'update',
            component: () => <span>kelsey</span>,
          },
          {
            label: t('修改时间:'),
            required: false,
            property: 'update-time',
            component: () => <span>2022-09-03 13：09</span>,
          },
          {
            label: t('备注:'),
            required: false,
            property: 'remark',
            component: () => {
              return (
                  <span>
                      <span>1223</span>
                      <i class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'}/>
                  </span>
              );
            },
          },
        ],
      },
    ]);

    const formBusinessInfo = [
      {
        name: t('业务信息'),
        data: [
          {
            label: t('组织架构:'),
            required: false,
            property: 'name',
            component: () => {
              return (
                  <span>
                      <span>IEG互动娱乐事业群/技术运营部</span>
                      <i class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'}/>
                  </span>
              );
            },
          },
          {
            label: t('使用业务:'),
            required: false,
            property: 'name',
            component: () => {
              return (
                  <span>
                      <span>资源运营服务(12)，蓝鲸配置平台(23)</span>
                      <i class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'}/>
                  </span>
              );
            },
          },
        ],
      },
    ];

    const formSecretInfo = [
      {
        name: t('密钥信息'),
        data: [
          {
            label: 'Secret ID',
            required: false,
            property: 'name',
            component: () => <span>11111</span>,
          },
          {
            label: 'Secret Key',
            required: false,
            property: 'name',
            component: () => <span>11111</span>,
          },
        ],
      },
    ];

    // const test = () => {
    //   console.log('1111333');
    // };

    console.log('formBaseInfo', formBaseInfo);

    return () => (
        <div class="w1000 detail-warp">
            {/* 基本信息 */}
            {formBaseInfo.map(baseItem => (
                <div>
                    <div class="font-bold pb10">{baseItem.name}</div>
                    <Form model={projectModel} labelWidth={100} rules={formRules} ref={formRef}>
                        <div class="flex-row align-items-center flex-wrap">
                            {baseItem.data.map(formItem => (
                                <FormItem class="formItem-cls" label={formItem.label} required={formItem.required} property={formItem.property}>
                                    {formItem.component()}
                                </FormItem>
                            ))
                        }
                        </div>
                    </Form>
                </div>
            ))
            }

            {/* 业务信息 */}
            {formBusinessInfo.map(businessItem => (
                <div>
                    <div class="font-bold pb10">{businessItem.name}</div>
                    <Form model={projectModel} labelWidth={100} rules={formRules}>
                        {businessItem.data.map(formItem => (
                            <FormItem class="formItem-cls" label={formItem.label} required={formItem.required} property={formItem.property}>
                                {formItem.component()}
                            </FormItem>
                        ))
                    }
                    </Form>
                </div>
            ))
            }

            {/* 密钥信息 */}
            {formSecretInfo.map(secretItem => (
                <div>
                    <div class="font-bold pb10">
                        {secretItem.name}
                        <i class={'icon hcm-icon bkhcm-icon-invisible1 pl15 account-edit-icon'}/>
                        <i class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'}/>
                    </div>
                    <Form model={projectModel} labelWidth={100} rules={formRules}>
                        {secretItem.data.map(formItem => (
                            <FormItem class="formItem-cls" label={formItem.label} required={formItem.required} property={formItem.property}>
                                {formItem.component()}
                            </FormItem>
                        ))
                    }
                    </Form>
                </div>
            ))
            }
        </div>
    );
  },
});
