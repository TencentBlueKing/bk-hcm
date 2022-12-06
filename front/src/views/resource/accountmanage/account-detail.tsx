import { Form, Dialog, Input } from 'bkui-vue';
import { reactive, defineComponent, ref, onMounted } from 'vue';
import { ProjectModel, SecretModel, CloudType, AccountType } from '@/typings';
import { CLOUD_TYPE } from '@/constants';
import { useI18n } from 'vue-i18n';
import { useAccountStore } from '@/store';
// import { useRoute } from 'vue-router';
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
    const accountStore = useAccountStore();
    // const route = useRoute();
    const formDiaRef = ref(null);

    const initProjectModel: ProjectModel = {
      id: 1,
      type: '',   // 账号类型
      name: '', // 名称
      vendor: '', // 云厂商
      account: '',    // 主账号
      subAccountId: '',    // 子账号id
      subAccountName: '',    // 子账号名称
      secretId: '',    // 密钥id
      secretKey: '',  // 密钥key
      managers: ['poloohuang'], // 责任人
      departmentId: [],   // 组织架构
      bizIds: '',   // 使用业务
      memo: '',     // 备注
      price: 0,
      extension: {},   // 特殊信息
    };

    const isTestConnection = ref(false);

    const initSecretModel: SecretModel = {
      secretId: '',
      secretKey: '',
    };

    let projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const secretModel = reactive<SecretModel>({
      ...initSecretModel,
    });

    const cloudType = reactive(CLOUD_TYPE);

    let dialogForm = reactive([]);

    onMounted(async () => {
      // const { id } = route.query;
      // console.log('route.query', route.query);
      // const res = await accountStore.getAccountDetail({ id: Number(id) });
      const res = { data: {
        id: 1,
        name: 'qcloud-account',
        vendor: 'aws',  // 云厂商，枚举值有：tcloud 、aws、azure、gcp、huawei
        type: 'resource',  // resource表示资源账号，register表示登记账号
        managers: ['jiananzhang', 'jamesge'],  // 负责人
        price: 500.01,  // 余额
        price_unit: '$', // 余额单位，可能是美元、人民币等
        created_at: '2022-12-05T10:44:55Z',
        updated_at: '2022-12-05T12:44:55Z',
        creator: 'jamesge',  // 创建者
        reviser: 'jiananzhang', // 更新者
        memo: '测试账号',  // 备注
        department_id: 2,  // 部门ID
        department_full_name: 'IEG互动娱乐事业群/技术运营部/计算资源中心',
        related_bk_biz_ids: [1, 3, 10], // 关联的业务列表，若选择All，则是-1
        extension: { account_id: 1, iam_username: 'poloohuang', secret_id: '**', secret_key: '**' },
      } };
      projectModel = res?.data;
      projectModel.departmentId = res?.data.department_id;
      projectModel.bizIds = res?.data.related_bk_biz_ids;
      renderDialogForm(projectModel);
      renderBaseInfoForm(projectModel);
    });

    // 动态表单
    const renderBaseInfoForm = (data: any) => {
      console.log('data', data.vendor);
      let insertFormData: any = [];
      switch (data.vendor) {
        case 'huawei':

          break;
        case 'tcloud':
          insertFormData = [
            {
              label: t('主账号:'),
              required: false,
              property: 'account',
              component: () => <span>{projectModel.extension.main_account}</span>,
            },
            {
              label: t('子账号:'),
              required: false,
              property: 'account',
              component: () => <span>{projectModel.extension.sub_account}</span>,
            },
          ];
          formBaseInfo[0].data.splice(4, 0, ...insertFormData);
          formBaseInfo.push({
            name: t('密钥信息'),
            data: [
              {
                label: 'Secret ID',
                required: false,
                property: 'secretId',
                component: () => <span>{projectModel.extension.secret_id}</span>,
              },
              {
                label: 'Secret Key',
                required: false,
                property: 'secretKey',
                component: () => <span>{projectModel.extension.secret_key}</span>,
              },
            ],
          });
          break;
        case 'aws':
          insertFormData = [
            {
              label: t('账号ID:'),
              required: false,
              property: 'account',
              component: () => <span>{projectModel.extension.account_id}</span>,
            },
            {
              label: t('IAM用户名称:'),
              required: false,
              property: 'account',
              component: () => <span>{projectModel.extension.iam_username}</span>,
            },
          ];
          formBaseInfo[0].data.splice(4, 0, ...insertFormData);
          formBaseInfo.push({
            name: t('密钥信息'),
            data: [
              {
                label: 'Secret ID',
                required: false,
                property: 'secretId',
                component: () => <span>{projectModel.extension.secret_id}</span>,
              },
              {
                label: 'Secret Key',
                required: false,
                property: 'secretKey',
                component: () => <span>{projectModel.extension.secret_key}</span>,
              },
            ],
          });
          break;
        default:
          break;
      }
    };

    // 弹窗
    const renderDialogForm = (data: any) => {
      switch (data.vendor) {
        case 'huawei':
          dialogForm = [
            {
              label: 'Secret ID',
              required: true,
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretId} />,
            },
            {
              label: 'Secret Key',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretKey} />,
            },
          ];
          break;
        case 'aws':
          dialogForm = [
            // {
            //   label: t('密钥ID'),
            //   required: true,
            //   property: 'secretId',
            //   component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretId} />,
            // },
            {
              label: 'Secret ID',
              required: true,
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretId} />,
            },
            {
              label: 'Secret Key',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretKey} />,
            },
          ];
          break;
        case 'gcp':
          dialogForm = [
            {
              label: 'Secret ID',
              required: true,
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretId} />,
            },
            {
              label: 'Secret Key',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretKey} />,
            },
          ];
          break;
        case 'azure':
          dialogForm = [
            {
              label: t('客户端ID'),
              required: true,
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretId} />,
            },
            {
              label: t('客户端密钥'),
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretKey} />,
            },
          ];
          break;
        case 'tcloud':
          dialogForm = [
            {
              label: 'Secret ID',
              required: true,
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretId} />,
            },
            {
              label: 'Secret Key',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretKey} />,
            },
          ];
          break;
        default:
          break;
      }
    };

    const isShowModifyScretDialog = ref(false);

    // const members = ['poloohuang'];
    // const department = [6544];

    const check = (val: any): boolean => {
      return  /^[a-z][a-z-z0-9_-]*$/.test(val);
    };

    const formRules = {
      name: [
        { trigger: 'blur', message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾业务与项目至少填一个', validator: check },
      ],
    };
    // 更新信息方法
    const updateFormData = async (key: any) => {
      let params: any = { department_id: '', related_bk_biz_ids: '' };
      if (key === 'departmentId') {
        params.department_id = projectModel[key];
      } else if (key === 'bizIds') {
        params.related_bk_biz_ids = projectModel[key];
      } else {
        params = {};
        params[key] = projectModel[key];
      }
      try {
        await accountStore.updateAccount({    // 更新密钥信息
          id: projectModel.id,
          ...params,
        });
      } catch (error) {

      }
    };

    // 显示弹窗
    const handleModifyScret = () => {
      isShowModifyScretDialog.value = true;
    };

    // 弹窗确认
    const onConfirm = async () => {
      await formDiaRef.value?.validate();
      const extension = {
        secret_id: secretModel.secretId,
        secret_key: secretModel.secretKey,
      };
      if (isTestConnection.value) {
        await accountStore.updateAccount({    // 更新密钥信息
          id: projectModel.id,
          extension,
        });
      } else {
        await accountStore.testAccountConnection({    // 测试连接密钥信息
          id: projectModel.id,
          extension,
        });
      }
    };

    // 取消
    const onClosed = () => {
      isShowModifyScretDialog.value = false;
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

    // 处理失焦
    const handleblur = async (val: boolean, key: string) => {
      handleEditStatus(val, key);     // 未通过检验前状态为编辑态
      await formRef.value?.validate();
      if (projectModel[key].length) {
        handleEditStatus(false, key);   // 通过检验则把状态改为不可编辑态
      }
      if (projectModel[key] !== initProjectModel[key]) {
        updateFormData(key);    // 更新数据
      }
    };

    const formBaseInfo = reactive([
      {
        name: '基本信息:',
        data: [
          {
            label: t('云厂商:'),
            required: false,
            property: 'vendor',
            isEdit: false,
            component: () => <span>{CloudType[projectModel.vendor]}</span>,
          },
          {
            label: t('账号类别:'),
            required: false,
            property: 'type',
            isEdit: false,
            component: () => <span>{AccountType[projectModel.type]}</span>,
          },
          {
            label: 'ID:',
            required: false,
            property: 'id',
            component: () => <span>{projectModel.id}</span>,
          },
          {
            label: t('名称:'),
            required: true,
            property: 'name',
            isEdit: false,
            component() {
              // eslint-disable-next-line max-len
              return (<RenderDetailEdit v-model={projectModel.name} fromPlaceholder={t('请输入名称')} fromKey={this.property} isEdit={this.isEdit} onBlur={handleblur}/>);
            },
          },
          {
            label: t('负责人:'),
            required: true,
            property: 'managers',
            isEdit: false,
            component() {
              return (<RenderDetailEdit v-model={projectModel.managers} fromKey={this.property} fromType="member" isEdit={this.isEdit} onBlur={handleblur}/>);
            },
          },
          {
            label: t('余额:'),
            required: false,
            property: 'price',
            component: () => <span>{projectModel.price}{projectModel.price_unit}</span>,
          },
          {
            label: t('创建人:'),
            required: false,
            property: 'creator',
            component: () => <span>{projectModel.creator}</span>,
          },
          {
            label: t('创建时间:'),
            required: false,
            property: 'created_at',
            component: () => <span>{projectModel.created_at}</span>,
          },
          {
            label: t('修改人:'),
            required: false,
            property: 'reviser',
            component: () => <span>{projectModel.reviser}</span>,
          },
          {
            label: t('修改时间:'),
            required: false,
            property: 'updated_at',
            component: () => <span>{projectModel.updated_at}</span>,
          },
          {
            label: t('备注:'),
            required: false,
            property: 'memo',
            isEdit: false,
            component() {
              // eslint-disable-next-line max-len
              return (<RenderDetailEdit v-model={projectModel.memo} fromKey={this.property} fromType="textarea" isEdit={this.isEdit} onBlur={handleblur}/>);
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
            property: 'departmentId',
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
            property: 'bizIds',
            isEdit: false,
            selectData: cloudType,
            component() {
              // eslint-disable-next-line max-len
              return (<RenderDetailEdit v-model={projectModel.bizIds} fromKey={this.property}
                selectData={this.selectData} fromType="select" isEdit={this.isEdit} onBlur={handleblur}/>);
            },
          },
        ],
      },

    ]);

    // const dialogForm = reactive([
    //   {
    //     label: 'Secret ID',
    //     required: true,
    //     property: 'secretId',
    //     component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretId} />,
    //   },
    //   {
    //     label: 'Secret Key',
    //     required: true,
    //     property: 'secretKey',
    //     component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.secretKey} />,
    //   },
    // ]);

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
                            <i class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'} onClick={handleModifyScret}/>
                          </span> : ''}
                    </div>
                    <Form model={projectModel} labelWidth={140} rules={formRules} ref={formRef}>
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

          <Dialog
            isShow={isShowModifyScretDialog.value}
            width={680}
            title={t('密钥信息')}
            onConfirm={onConfirm}
            confirmText={isTestConnection.value ? '确认' : '测试连接'}
            onClosed={onClosed}
          >
            <Form labelWidth={100} model={secretModel} ref={formDiaRef}>
            {dialogForm.map(formItem => (
                <FormItem label={formItem.label} required={formItem.required} property={formItem.property}>
                    {formItem.component()}
                </FormItem>
            ))
            }
            </Form>
          </Dialog>
        </div>
    );
  },
});
