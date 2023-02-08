import { Form, Dialog, Input, Message, Button } from 'bkui-vue';
import { reactive, defineComponent, ref, onMounted, computed } from 'vue';
import { ProjectModel, SecretModel, CloudType, AccountType, SiteType } from '@/typings';
import { useI18n } from 'vue-i18n';
import { useAccountStore } from '@/store';
import { useRoute } from 'vue-router';
import Loading from '@/components/loading';
import OrganizationSelect from '@/components/OrganizationSelect';
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
    const route = useRoute();
    const formDiaRef = ref(null);
    const requestQueue = ref(['detail', 'departments', 'bizsList']);

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
      managers: [], // 责任人
      departmentId: [],   // 组织架构
      bizIds: [],   // 使用业务
      memo: '',     // 备注
      price: 0,
      extension: {},   // 特殊信息
    };

    const departmentFullName = ref('IEG互动娱乐事业群/技术运营部/计算资源中心');
    const departmentData = ref([]);
    const isShowModifyScretDialog = ref(false);
    const buttonLoading = ref<boolean>(false);

    const initSecretModel: SecretModel = {
      secretId: '',
      secretKey: '',
      subAccountId: '',
      iamUserName: '',
    };

    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const secretModel = reactive<SecretModel>({
      ...initSecretModel,
    });

    const businessList = reactive({ // 业务列表
      list: [],
    });

    const isOrganizationDetail = ref<Boolean>(true);      // 组织架构详情展示

    const dialogForm = reactive({ list: [] });

    const getDetail = async () => {
      const { id } = route.query;
      const res = await accountStore.getAccountDetail(id);
      projectModel.id = res?.data.id;
      projectModel.vendor = res?.data.vendor;
      projectModel.name = res?.data.name;
      projectModel.type = res?.data.type;
      projectModel.managers = res?.data.managers;
      projectModel.price = res?.data.price;
      projectModel.price_unit = res?.data.price_unit;
      projectModel.site = res?.data.site;
      projectModel.memo = res?.data.memo;
      projectModel.departmentId = res?.data.department_ids; // 1
      projectModel.creator = res?.data.creator;
      projectModel.reviser = res?.data.reviser;
      projectModel.created_at = res?.data.created_at;
      projectModel.updated_at = res?.data.updated_at;
      projectModel.extension = res?.data.extension;
      projectModel.bizIds = res?.data?.bk_biz_ids;
      requestQueue.value.shift();
      getBusinessList();
      getDepartmentInfo(res?.data.department_ids);
      renderDialogForm(projectModel);
      renderBaseInfoForm(projectModel);
    };

    onMounted(() => {
      getDetail();    // 请求数据
    });

    const isLoading = computed(() => {
      return requestQueue.value.length > 0;
    });

    // 获取业务列表
    const getBusinessList = async () => {
      const res = await accountStore.getBizList();
      businessList.list = res.data;
      requestQueue.value.shift();
    };

    // 获取部门信息
    const getDepartmentInfo = async (id: number) => {
      const res = await accountStore.getDepartmentInfo(id);
      departmentFullName.value = res?.data?.full_name;
      departmentData.value = res?.data?.ancestors.map((e: any) => e.id);
      console.log('departmentData.value', departmentData.value);
      requestQueue.value.shift();
    };

    // 动态表单
    const renderBaseInfoForm = (data: any) => {
      let insertFormData: any = [];
      switch (data.vendor) {
        case 'huawei':
          insertFormData = [
            {
              label: t('主账号名'),
              required: false,
              property: 'cloud_main_account_name',
              component: () => <span>{projectModel.extension.cloud_main_account_name || '--'}</span>,
            },
            {
              label: t('账号名'),
              required: false,
              property: 'cloud_sub_account_name',
              component: () => <span>{projectModel.extension.cloud_sub_account_name || '--'}</span>,
            },
            {
              label: t('账号ID'),
              required: false,
              property: 'cloud_sub_account_id',
              component: () => <span>{projectModel.extension.cloud_sub_account_id || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(4, 0, ...insertFormData);
          formBaseInfo.push({
            name: t('密钥信息'),
            data: [
              {
                label: t('IAM用户ID'),
                required: false,
                property: 'cloud_iam_user_id',
                component: () => <span>{projectModel.extension.cloud_iam_user_id || '--'}</span>,
              },
              {
                label: t('IAM用户名'),
                required: false,
                property: 'cloud_iam_username',
                component: () => <span>{projectModel.extension.cloud_iam_username || '--'}</span>,
              },
              {
                label: 'Secret ID',
                required: false,
                property: 'cloud_secret_id',
                component: () => <span>{projectModel.extension.cloud_secret_id}</span>,
              },
              {
                label: 'Secret Key',
                required: false,
                property: 'cloud_secret_key',
                component: () => <span>********</span>,
              },
            ],
          });
          break;
          break;
        case 'tcloud':
          insertFormData = [
            {
              label: t('主账号ID'),
              required: false,
              property: 'cloud_main_account_id',
              component: () => <span>{projectModel.extension.cloud_main_account_id || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(4, 0, ...insertFormData);
          formBaseInfo.push({
            name: t('密钥信息'),
            data: [
              {
                label: 'Secret ID',
                required: false,
                property: 'cloud_secret_id',
                component: () => <span>{projectModel.extension.cloud_secret_id}</span>,
              },
              {
                label: 'Secret Key',
                required: false,
                property: 'cloud_secret_key',
                component: () => <span>********</span>,
              },
              {
                label: t('子账号ID'),
                required: false,
                property: 'cloud_sub_account_id',
                component: () => <span>{projectModel.extension.cloud_sub_account_id}</span>,
              },
            ],
          });
          break;
        case 'aws':
          insertFormData = [
            {
              label: t('账号ID:'),
              required: false,
              property: 'cloud_account_id',
              component: () => <span>{projectModel.extension.cloud_account_id || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(4, 0, ...insertFormData);
          formBaseInfo.push({
            name: t('密钥信息'),
            data: [
              {
                label: t('IAM用户名称:'),
                required: false,
                property: 'cloud_iam_username',
                component: () => <span>{projectModel.extension.cloud_iam_username || '--'}</span>,
              },
              {
                label: 'Secret ID',
                required: false,
                property: 'secretId',
                component: () => <span>{projectModel.extension.cloud_secret_id}</span>,
              },
              {
                label: 'Secret Key',
                required: false,
                property: 'cloud_secret_key',
                component: () => <span>********</span>,
              },
            ],
          });
          break;
        case 'azure':
          insertFormData = [
            {
              label: t('租户 ID'),
              required: false,
              property: 'cloud_tenant_id',
              component: () => <span>{projectModel.extension.cloud_tenant_id || '--'}</span>,
            },
            {
              label: t('订阅 ID'),
              required: false,
              property: 'cloud_subscription_id',
              component: () => <span>{projectModel.extension.cloud_subscription_id || '--'}</span>,
            },
            {
              label: t('订阅 名称'),
              required: false,
              property: 'cloud_subscription_name',
              component: () => <span>{projectModel.extension.cloud_subscription_name || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(4, 0, ...insertFormData);
          formBaseInfo.push({
            name: t('密钥信息'),
            data: [
              {
                label: t('应用程序(客户端) ID'),
                required: false,
                property: 'cloud_application_id',
                component: () => <span>{projectModel.extension.cloud_application_id || '--'}</span>,
              },
              {
                label: t('应用程序名称'),
                required: false,
                property: 'cloud_application_name',
                component: () => <span>{projectModel.extension.cloud_application_name || '--'}</span>,
              },
              {
                label: t('客户端密钥ID'),
                required: false,
                property: 'cloud_client_secret_id',
                component: () => <span>{projectModel.extension.cloud_client_secret_id || '--'}</span>,
              },
              {
                label: t('客户端密钥'),
                required: false,
                property: 'cloud_client_secret_key',
                component: () => <span>********</span>,
              },
            ],
          });
          break;
        case 'gcp':
          insertFormData = [
            {
              label: t('项目 ID'),
              required: false,
              property: 'cloud_project_id',
              component: () => <span>{projectModel.extension.cloud_project_id || '--'}</span>,
            },
            {
              label: t('项目名称'),
              required: false,
              property: 'cloud_project_name',
              component: () => <span>{projectModel.extension.cloud_project_name || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(4, 0, ...insertFormData);
          formBaseInfo.push({
            name: t('密钥信息'),
            data: [
              {
                label: t('服务账号ID'),
                required: false,
                property: 'cloud_service_account_id',
                component: () => <span>{projectModel.extension.cloud_service_account_id}</span>,
              },
              {
                label: t('服务账号名称'),
                required: false,
                property: 'cloud_service_account_name',
                component: () => <span>{projectModel.extension.cloud_service_account_name}</span>,
              },
              {
                label: 'Secret ID',
                required: false,
                property: 'secretId',
                component: () => <span>{projectModel.extension.cloud_service_secret_id}</span>,
              },
              {
                label: 'Secret Key',
                required: false,
                property: 'cloud_secret_key',
                component: () => <span>********</span>,
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
          dialogForm.list = [
            {
              label: t('IAM用户ID'),
              required: true,
              property: 'iamUserId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.iamUserId} />,
            },
            {
              label: 'IAM用户名',
              required: true,
              property: 'iamUserName',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.iamUserName} />,
            },
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
          dialogForm.list = [
            {
              label: t('IAM用户名'),
              required: false,
              property: 'iamUserName',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.iamUserName} />,
            },
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
          dialogForm.list = [
            {
              label: t('服务账号ID'),
              required: true,
              property: 'accountId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.accountId} />,
            },
            {
              label: t('服务账号名称'),
              required: true,
              property: 'accountName',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.accountName} />,
            },
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
          dialogForm.list = [
            {
              label: t('应用程序(客户端) ID'),
              required: true,
              property: 'applicationId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.applicationId} />,
            },
            {
              label: t('应用程序名称'),
              required: true,
              property: 'applicationName',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.applicationName} />,
            },
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
          dialogForm.list = [
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
            {
              label: t('子账号ID'),
              required: true,
              property: 'subAccountId',
              component: () => <Input class="w450" placeholder={t('请输入')} v-model={secretModel.subAccountId} />,
            },
          ];
          break;
        default:
          break;
      }
    };

    const check = (val: any): boolean => {
      return  /^[a-z][a-z-z0-9_-]*$/.test(val);
    };

    const formRules = {
      name: [
        { trigger: 'blur', message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾', validator: check },
      ],
    };
    // 更新信息方法
    const updateFormData = async (key: any, val?: any) => {
      let params: any = {};
      if (key === 'departmentId') {
        params.department_ids = projectModel[key];
        isOrganizationDetail.value = true;  // 改为详情展示态
      } else if (key === 'bizIds') {
        // 若选择全部业务，则参数是-1
        params.bk_biz_ids = projectModel[key].length === businessList.list.length
          ? [-1] : projectModel[key];
      } else {
        params = {};
        params[key] = projectModel[key];
      }
      try {
        await accountStore.updateAccount({    // 更新密钥信息
          id: projectModel.id,
          ...params,
        });
        Message({
          message: t('更新成功'),
          theme: 'success',
        });
        if (key === 'departmentId') {
          departmentFullName.value = val;
        }
      } catch (error) {
        console.log(error);
      } finally {
        isOrganizationDetail.value = true;  // 改为详情展示态
      }
    };

    // 显示弹窗
    const handleModifyScret = () => {
      secretModel.secretId = '';
      secretModel.secretKey = '';
      secretModel.subAccountId = '';
      secretModel.iamUserName = '';
      secretModel.iamUserId = '';
      secretModel.accountId = '';
      secretModel.accountName = '';
      secretModel.applicationId = '';
      secretModel.applicationName = '';
      isShowModifyScretDialog.value = true;
    };

    // 弹窗确认
    const onConfirm = async () => {
      await formDiaRef.value?.validate();
      buttonLoading.value = true;
      const extension: any = {
        cloud_secret_id: secretModel.secretId,
        cloud_secret_key: secretModel.secretKey,
      };
      // 后期拓展
      switch (projectModel.vendor) {
        case 'huawei':
          extension.cloud_iam_username = secretModel.iamUserName;
          extension.cloud_iam_user_id = secretModel.iamUserId;
          break;
        case 'tcloud':
          extension.cloud_sub_account_id = secretModel.subAccountId;
          break;
        case 'aws':
          extension.cloud_iam_username = secretModel.iamUserName;
          break;
        case 'azure':
          extension.cloud_application_id = secretModel.applicationId;
          extension.cloud_application_name = secretModel.applicationName;
          extension.cloud_client_secret_id = secretModel.secretId;
          extension.cloud_client_secret_key = secretModel.secretKey;
          delete extension.cloud_secret_id;
          delete extension.cloud_secret_key;
          break;
        case 'gcp':
          extension.cloud_service_account_id = secretModel.accountId;   // 服务账号ID
          extension.cloud_service_account_name = secretModel.accountName;   // 服务账号ID
          extension.cloud_service_secret_id = secretModel.secretId;
          extension.cloud_service_secret_key = secretModel.secretKey;
          delete extension.cloud_secret_id;
          delete extension.cloud_secret_key;
          break;
        default:
          break;
      }
      try {
        await accountStore.updateTestAccount({    // 测试连接密钥信息
          id: projectModel.id,
          extension,
        });
        await accountStore.updateAccount({    // 更新密钥信息
          id: projectModel.id,
          extension,
        });
        Message({
          message: t('更新密钥信息成功'),
          theme: 'success',
        });
        projectModel.extension = extension;
        onClosed();
      } catch (error) {
        console.log(error);
      } finally {
        buttonLoading.value = false;
      }
    };

    // 取消
    const onClosed = () => {
      isShowModifyScretDialog.value = false;
    };

    const handleEditStatus = (val: boolean, key: string) => {
      console.log(val, key);
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

    // 处理组织架构选择
    const handleOrganChange = (val: any) => {
      updateFormData('departmentId', val);    // 更新数据
    };

    // 组织架构编辑
    const handleEdit = () => {
      isOrganizationDetail.value = false;
    };

    const formBaseInfo = reactive([
      {
        name: t('基本信息'),
        data: [
          {
            label: t('云厂商'),
            required: false,
            property: 'vendor',
            isEdit: false,
            component: () => <span>{CloudType[projectModel.vendor]}</span>,
          },
          {
            label: t('账号类别'),
            required: false,
            property: 'type',
            isEdit: false,
            component: () => <span>{AccountType[projectModel.type]}</span>,
          },
          {
            label: 'ID',
            required: false,
            property: 'id',
            component: () => <span>{projectModel.id}</span>,
          },
          {
            label: t('名称'),
            required: true,
            property: 'name',
            isEdit: false,
            component() {
              // eslint-disable-next-line max-len
              return (<RenderDetailEdit v-model={projectModel.name} fromPlaceholder={t('请输入名称')} fromKey={this.property} isEdit={this.isEdit} onBlur={handleblur}/>);
            },
          },
          {
            label: t('负责人'),
            required: true,
            property: 'managers',
            isEdit: false,
            component() {
              return (<RenderDetailEdit v-model={projectModel.managers} fromKey={this.property} fromType="member" isEdit={this.isEdit} onBlur={handleblur}/>);
            },
          },
          {
            label: t('余额'),
            required: false,
            property: 'price',
            component: () => <span>{projectModel?.price || '--'}{projectModel.price_unit}</span>,
          },
          {
            label: t('站点类型'),
            required: false,
            property: 'price',
            component: () => <span>{SiteType[projectModel.site]}</span>,
          },
          {
            label: t('创建人'),
            required: false,
            property: 'creator',
            component: () => <span>{projectModel.creator}</span>,
          },
          {
            label: t('创建时间'),
            required: false,
            property: 'created_at',
            component: () => <span>{projectModel.created_at}</span>,
          },
          {
            label: t('修改人'),
            required: false,
            property: 'reviser',
            component: () => <span>{projectModel.reviser}</span>,
          },
          {
            label: t('修改时间'),
            required: false,
            property: 'updated_at',
            component: () => <span>{projectModel.updated_at}</span>,
          },
          {
            label: t('备注'),
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
            label: t('组织架构'),
            required: false,
            property: 'departmentId',
            isEdit: true,
            component() {
              return (
                isOrganizationDetail.value ? (<div class="flex-row align-items-center">
                  <span>{departmentFullName.value}</span>
                  <i onClick={handleEdit} class={'icon hcm-icon bkhcm-icon-edit pl15 account-edit-icon'}/>
                </div>)
                  : (<OrganizationSelect v-model={projectModel.departmentId} onChange={handleOrganChange}/>)
              );
            },
          },
          {
            label: t('使用业务'),
            required: false,
            property: 'bizIds',
            isEdit: false,
            component() {
              // eslint-disable-next-line max-len
              return (<RenderDetailEdit v-model={projectModel.bizIds} fromKey={this.property}
                selectData={businessList.list} fromType="select" isEdit={this.isEdit} onBlur={handleblur}/>);
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
      isLoading.value ? (<Loading/>) : (
        <div class="w1400 detail-warp">
            {/* 基本信息 */}
            {formBaseInfo.map((baseItem, index) => (
                <div>
                    <div class="font-bold pb10">
                      {baseItem.name}
                      {index === 2
                        ? <span>
                            {/* <i class={'icon hcm-icon bkhcm-icon-invisible1 pl15 account-edit-icon'}/> */}
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
            v-model:isShow={isShowModifyScretDialog.value}
            width={680}
            title={t('密钥信息')}
            dialogType={'show'}
            onClosed={onClosed}
          >
            <Form labelWidth={100} model={secretModel} ref={formDiaRef}>
            {dialogForm.list.map(formItem => (
                <FormItem label={formItem.label} required={formItem.required} property={formItem.property}>
                    {formItem.component()}
                </FormItem>
            ))
            }
            </Form>
            <div class="button-warp">
              <Button theme="primary" loading={buttonLoading.value} onClick={onConfirm}>{t('确认')}</Button>
              <Button class="ml10" onClick={onClosed}>{t('取消')}</Button>
            </div>
          </Dialog>
        </div>
      )
    );
  },
});
