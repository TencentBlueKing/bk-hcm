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
      name: 'test',
      cloudName: '',
      scretId: '',
      account: '',
      user: ['poloohuang'],
      remark: '测试描述',
      business: 'huawei',
    };
    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const cloudType = reactive([
      { key: '华为云', value: 'huawei' },
      { key: '腾讯云', value: 'tencent' },
      { key: '亚马逊', value: 'aws' },
    ]);

    // const members = ['poloohuang'];
    // const department = [6544];

    const check = (val: any): boolean => {
      return  /^[a-z][a-z-z0-9_-]*$/.test(val);
    };

    const formRules = {
      name: [{ trigger: 'blur', message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾业务与项目至少填一个', validator: check }],
    };

    // 更新信息方法
    const updateFormData = () => {

    };

    const handleEditStatus = (val: boolean, key: string) => {
      formBaseInfo.forEach((e) => {
        e.data = e.data.map((item) => {
          if (item.property === key) {
            item.isEdit = val;
          }
          return item;
        });
      });
    };

    const handleblur = async (val: boolean, key: string) => {
      handleEditStatus(val, key);     // 未通过检验前状态为编辑态
      await formRef.value?.validate();
      handleEditStatus(false, key);   // 通过检验则把状态改为不可编辑态
      if (projectModel[key] !== initProjectModel[key]) {
        console.log('projectModel', projectModel);
        updateFormData();    // 更新数据
      }
    };

    const formBaseInfo = reactive([
      {
        name: '基本信息:',
        data: [
          {
            label: t('云厂商:'),
            required: false,
            property: 'cloudType',
            isEdit: false,
            component: () => <span>{t('腾讯云')}</span>,
          },
          {
            label: t('账号类别:'),
            required: false,
            property: 'accountType',
            isEdit: false,
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
            isEdit: false,
            component() {
              // eslint-disable-next-line max-len
              return (<RenderDetailEdit v-model={projectModel.name} fromKey={this.property} isEdit={this.isEdit} onBlur={handleblur}/>);
            },
          },
          {
            label: t('主账号:'),
            required: false,
            property: 'account',
            component: () => <span>23445</span>,
          },
          {
            label: t('子账号:'),
            required: false,
            property: 'account',
            component: () => <span>23445</span>,
          },
          {
            label: t('负责人:'),
            required: false,
            property: 'user',
            isEdit: false,
            component() {
              return (<RenderDetailEdit v-model={projectModel.user} fromKey={this.property} fromType="member" isEdit={this.isEdit} onBlur={handleblur}/>);
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
            isEdit: false,
            component() {
              // eslint-disable-next-line max-len
              return (<RenderDetailEdit v-model={projectModel.remark} fromKey={this.property} fromType="textarea" isEdit={this.isEdit} onBlur={handleblur}/>);
            },
          },
        ],
      },
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
            property: 'business',
            isEdit: false,
            selectData: cloudType,
            component() {
              // eslint-disable-next-line max-len
              return (<RenderDetailEdit v-model={projectModel.business} fromKey={this.property}
                selectData={this.selectData} fromType="select" isEdit={this.isEdit} onBlur={handleblur}/>);
            },
          },
        ],
      },

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

    ]);

    // const test = () => {
    //   console.log('1111333');
    // };

    // console.log('formBaseInfo', formBaseInfo);

    return () => (
        <div class="w1000 detail-warp">
            {/* 基本信息 */}
            {formBaseInfo.map((baseItem, index) => (
                <div>
                    <div class="font-bold pb10">
                      {baseItem.name}
                      {index === 2
                        ? <span>
                            <i class={'icon hcm-icon bkhcm-icon-invisible1 pl15 account-edit-icon'}/>
                            <i class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'}/>
                          </span> : ''}
                    </div>
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
        </div>
    );
  },
});
