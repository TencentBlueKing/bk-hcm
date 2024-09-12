import { Form, Dialog, Input, Message, Button, Alert, Select } from 'bkui-vue';
import { reactive, defineComponent, ref, onMounted, computed, watch } from 'vue';
import { ProjectModel, SecretModel, CloudType, SiteType } from '@/typings';
import { useI18n } from 'vue-i18n';
import { useAccountStore } from '@/store';
import { useRoute } from 'vue-router';
import Loading from '@/components/loading';
import RenderDetailEdit from '@/components/RenderDetailEdit';
import DetailHeader from '../resource-manage/common/header/detail-header';
import './account-detail.scss';
import MemberSelect from '@/components/MemberSelect';
import http from '@/http';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import {
  ValidateStatus,
  useSecretExtension,
} from '../resource-manage/account/createAccount/components/accountForm/useSecretExtension';
import { VendorEnum } from '@/common/constant';
import { timeFormatter } from '@/common/util';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const { FormItem } = Form;
const { Option } = Select;

// const { Option } = Select;
export default defineComponent({
  name: 'AccountManageDetail',
  setup() {
    const { t } = useI18n();
    const formRef = ref<InstanceType<typeof Form>>(null);
    const accountStore = useAccountStore();
    const route = useRoute();
    const formDiaRef = ref(null);
    const requestQueue = ref(['detail', 'bizsList']);
    const isDetail = ref(route.query.isDetail);

    const initProjectModel: ProjectModel = {
      id: 1,
      type: '', // 账号类型
      name: '', // 名称
      vendor: VendorEnum.TCLOUD, // 云厂商
      account: '', // 主账号
      subAccountId: '', // 子账号id
      subAccountName: '', // 子账号名称
      secretId: '', // 密钥id
      secretKey: '', // 密钥key
      managers: [], // 责任人
      bizIds: [], // 使用业务
      memo: '', // 备注
      price: 0,
      extension: {}, // 特殊信息
    };
    const isShowModifyScretDialog = ref(false);
    const isShowModifyAccountDialog = ref(false);
    const isAccountDialogLoading = ref(false);
    const isSecretDialogLoading = ref(false);
    const buttonLoading = ref<boolean>(false);
    const accountFormModel = reactive({
      managers: [],
      memo: '',
      bk_biz_ids: [],
    });
    const accountForm = ref(null);

    const computedManagers = computed(() =>
      accountFormModel.managers.map((name) => ({
        username: name,
        display_name: name,
      })),
    );

    const resourceAccountStore = useResourceAccountStore();

    const initSecretModel: SecretModel = {
      secretId: '',
      secretKey: '',
      subAccountId: '',
      iamUserName: '',
    };

    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const {
      curExtension,
      isValidateLoading,
      handleValidate,
      isValidateDiasbled,
      extensionPayload,
    } = useSecretExtension(projectModel, true);

    const secretModel = reactive<SecretModel>({
      ...initSecretModel,
    });

    const businessList = reactive({
      // 业务列表
      list: [],
    });

    const isOrganizationDetail = ref<Boolean>(true); // 组织架构详情展示
    const getDetail = async () => {
      const { id, accountId } = route.query;
      const res = await accountStore.getAccountDetail(id || accountId);
      projectModel.id = res?.data.id;
      projectModel.vendor = res?.data.vendor;
      projectModel.name = res?.data.name;
      projectModel.type = res?.data.type;
      projectModel.managers = res?.data.managers;
      projectModel.price = res?.data.price;
      projectModel.price_unit = res?.data.price_unit;
      projectModel.site = res?.data.site;
      projectModel.memo = res?.data.memo;
      projectModel.creator = res?.data.creator;
      projectModel.reviser = res?.data.reviser;
      projectModel.created_at = res?.data.created_at;
      projectModel.updated_at = res?.data.updated_at;
      projectModel.extension = res?.data.extension;
      projectModel.bizIds = res?.data?.bk_biz_ids;
      requestQueue.value.shift();
      getBusinessList();
      renderBaseInfoForm(projectModel);
    };

    onMounted(() => {
      getDetail(); // 请求数据
    });

    watch(
      () => route.query.accountId,
      (id, oldId) => {
        if (!oldId && id) return;
        if (id) {
          getDetail();
        }
      },
    );

    const isLoading = computed(() => {
      return requestQueue.value.length > 0;
    });

    // 获取业务列表
    const getBusinessList = async () => {
      const res = await accountStore.getBizList();
      businessList.list = res.data;
      requestQueue.value.shift();
    };

    // 动态表单
    const renderBaseInfoForm = (data: any) => {
      let insertFormData: any = [];
      const siteIndex = formBaseInfo[0].data.findIndex((e) => e.property === 'site');
      const creatorIndex = formBaseInfo[0].data.findIndex((e) => e.property === 'creator');
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
              label: t('账号 ID'),
              required: false,
              property: 'cloud_sub_account_id',
              component: () => <span>{projectModel.extension.cloud_sub_account_id || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(siteIndex + 1, creatorIndex - (siteIndex + 1), ...insertFormData);
          formBaseInfo[2].data = [
            {
              label: t('IAM用户 ID'),
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
          ];
          break;
        case 'tcloud':
          insertFormData = [
            {
              label: t('主账号 ID'),
              required: false,
              property: 'cloud_main_account_id',
              component: () => <span>{projectModel.extension.cloud_main_account_id || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(siteIndex + 1, creatorIndex - (siteIndex + 1), ...insertFormData);
          formBaseInfo[2].data = [
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
              label: t('子账号 ID'),
              required: false,
              property: 'cloud_sub_account_id',
              component: () => <span>{projectModel.extension.cloud_sub_account_id}</span>,
            },
          ];
          break;
        case 'aws':
          insertFormData = [
            {
              label: t('账号 ID'),
              required: false,
              property: 'cloud_account_id',
              component: () => <span>{projectModel.extension.cloud_account_id || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(siteIndex + 1, creatorIndex - (siteIndex + 1), ...insertFormData);
          formBaseInfo[2].data = [
            {
              label: t('IAM用户名称'),
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
          ];
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
              label: t('订阅名称'),
              required: false,
              property: 'cloud_subscription_name',
              component: () => <span>{projectModel.extension.cloud_subscription_name || '--'}</span>,
            },
          ];
          formBaseInfo[0].data.splice(siteIndex + 1, creatorIndex - (siteIndex + 1), ...insertFormData);
          formBaseInfo[2].data = [
            {
              label: t('应用(客户端)ID'),
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
              label: t('客户端密钥 ID'),
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
          ];
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
          formBaseInfo[0].data.splice(siteIndex + 1, creatorIndex - (siteIndex + 1), ...insertFormData);
          formBaseInfo[2].data = [
            {
              label: t('服务账号 ID'),
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
          ];
          break;
        default:
          break;
      }
    };

    const check = (val: any): boolean => {
      console.log('check', check);
      return /^[a-z][a-z-z0-9_-]*$/.test(val);
    };

    const formRules = {
      name: [
        {
          trigger: 'blur',
          message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾',
          validator: check,
        },
      ],
    };
    // 更新信息方法
    const updateFormData = async (key: any) => {
      let params: any = {};
      if (key === 'bizIds') {
        // 若选择全部业务，则参数是-1
        // params.bk_biz_ids = projectModel[key].length === businessList.list.length
        //   ? [-1] : projectModel[key];
        params.bk_biz_ids = projectModel[key] ? [projectModel[key]] : [-1];
      } else {
        params = {};
        params[key] = projectModel[key];
      }
      try {
        await accountStore.updateAccount({
          // 更新密钥信息
          id: projectModel.id,
          ...params,
        });
        Message({
          message: t('更新成功'),
          theme: 'success',
        });
      } catch (error) {
        console.log(error);
      } finally {
        isOrganizationDetail.value = true; // 改为详情展示态
        getDetail(); // 请求数据
      }
    };

    // 显示弹窗
    const handleModifyScret = () => {
      secretModel.secretId =
        projectModel.extension.cloud_secret_id ||
        projectModel.extension.cloud_client_secret_id ||
        projectModel.extension.cloud_service_secret_id ||
        '';
      secretModel.secretKey = '';
      secretModel.subAccountId = projectModel.extension.cloud_sub_account_id || '';
      secretModel.iamUserName = projectModel.extension.cloud_iam_username || '';
      secretModel.iamUserId = projectModel.extension.cloud_iam_user_id || '';
      secretModel.accountId = projectModel.extension.cloud_service_account_id || '';
      secretModel.accountName = projectModel.extension.cloud_service_account_name || '';
      secretModel.applicationId = projectModel.extension.cloud_application_id || '';
      secretModel.applicationName = projectModel.extension.cloud_application_name || '';
      isShowModifyScretDialog.value = true;
    };

    // 弹窗确认
    const onConfirm = async () => {
      await formDiaRef.value?.validate();
      buttonLoading.value = true;
      try {
        const extension = extensionPayload.value;
        await accountStore.updateTestAccount({
          // 测试连接密钥信息
          id: projectModel.id,
          extension: {
            cloud_sub_account_id: curExtension.value.output1.cloud_sub_account_id?.value,
            ...extension,
          },
        });
        await accountStore.updateAccount({
          // 更新密钥信息
          id: projectModel.id,
          extension: {
            cloud_sub_account_id: curExtension.value.output1.cloud_sub_account_id?.value,
            ...extension,
          },
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
      if (!projectModel.managers.length) {
        Message({
          message: t('请选择负责人'),
          theme: 'error',
        });
        return;
      }
      handleEditStatus(val, key); // 未通过检验前状态为编辑态
      await formRef.value?.validate();
      if (projectModel[key].length) {
        handleEditStatus(false, key); // 通过检验则把状态改为不可编辑态
      }
      if (projectModel[key] !== initProjectModel[key]) {
        await updateFormData(key); // 更新数据
        setInterval(() => {
          window.location.reload();
        }, 300);
      }
    };

    const handleModifyAccount = () => {
      isShowModifyAccountDialog.value = true;
      // 数据回显
      Object.assign(accountFormModel, {
        managers: projectModel.managers,
        memo: projectModel.memo,
        bk_biz_ids: projectModel.bizIds,
      });
    };

    const handleModifyAccountSubmit = async () => {
      await accountForm.value.validate();
      isAccountDialogLoading.value = true;
      // Select单选下，不返回数组，需要进行转换
      await http.patch(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/accounts/${resourceAccountStore.resourceAccount?.id}`, {
        managers: accountFormModel.managers,
        memo: accountFormModel.memo,
        bk_biz_ids: Array.isArray(accountFormModel.bk_biz_ids)
          ? accountFormModel.bk_biz_ids
          : [accountFormModel.bk_biz_ids],
      });
      isAccountDialogLoading.value = false;
      isShowModifyAccountDialog.value = false;
      getDetail();
    };

    // const handleBizChange = async () => {
    //   handleEditStatus(true, 'bizIds');     // 未通过检验前状态为编辑态
    //   await formRef.value?.validate();
    //   handleEditStatus(false, 'bizIds');   // 通过检验则把状态改为不可编辑态
    //   updateFormData('bizIds');    // 更新数据
    // };

    const formBaseInfo = reactive([
      {
        name: t('基本信息'),
        data: [
          {
            label: 'ID',
            required: false,
            property: 'id',
            component: () => <span>{projectModel.id}</span>,
          },
          {
            label: t('名称'),
            property: 'name',
            isEdit: true,
            component() {
              // eslint-disable-next-line max-len
              return (
                <RenderDetailEdit
                  v-model={projectModel.name}
                  fromPlaceholder={t('请输入名称')}
                  fromKey={this.property}
                  hideEdit={!!isDetail.value}
                  isEdit={this.isEdit}
                  onBlur={handleblur}
                />
              );
            },
          },
          {
            label: t('云厂商'),
            required: false,
            property: 'vendor',
            isEdit: false,
            component: () => <span>{CloudType[projectModel.vendor]}</span>,
          },
          {
            label: t('站点类型'),
            required: false,
            property: 'site',
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
            component: () => <span>{timeFormatter(projectModel.created_at)}</span>,
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
            component: () => <span>{timeFormatter(projectModel.updated_at)}</span>,
          },

          // {
          //   label: t('账号类别'),
          //   required: false,
          //   property: 'type',
          //   isEdit: false,
          //   component: () => <span>{AccountType[projectModel.type]}</span>,
          // },
          // {
          //   label: t('余额'),
          //   required: false,
          //   property: 'price',
          //   component: () => (
          //     <span>
          //       {projectModel?.price || '--'}
          //       {projectModel.price_unit}
          //     </span>
          //   ),
          // },
        ],
      },
      {
        name: '账号归属',
        data: [
          {
            label: t('负责人'),
            property: 'managers',
            isEdit: false,
            component() {
              return (
                <RenderDetailEdit
                  v-model={projectModel.managers}
                  fromKey={this.property}
                  fromType='member'
                  hideEdit={true}
                  isEdit={this.isEdit}
                  onBlur={handleblur}
                />
              );
            },
          },
          {
            label: t('备注'),
            required: false,
            property: 'memo',
            isEdit: false,
            component() {
              // eslint-disable-next-line max-len
              return (
                <RenderDetailEdit
                  v-model={projectModel.memo}
                  fromKey={this.property}
                  fromType='textarea'
                  hideEdit={true}
                  isEdit={this.isEdit}
                  onBlur={handleblur}
                />
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
              // onBlur={handleblur}
              // onChange={handleBizChange}
              return (
                <RenderDetailEdit
                  v-model={projectModel.bizIds}
                  fromKey={this.property}
                  hideEdit={true}
                  selectData={businessList.list}
                  fromType='select'
                  isEdit={this.isEdit}
                />
              );
              // <span>{SiteType[projectModel.bizIds]}</span>
            },
          },
        ],
      },
      {
        name: '密钥信息',
        data: [],
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

    return () =>
      isLoading.value ? (
        <Loading />
      ) : (
        <div class='detail-wrap'>
          {!route.path.includes('resource/resource/account/detail') && (
            <>
              <DetailHeader>
                <span class='header-title-prefix'>账号详情</span>
                <span class='header-title-content'>&nbsp;- ID {projectModel.id}</span>
              </DetailHeader>
              <div class='h16'></div>
            </>
          )}
          {/* 基本信息 */}
          {formBaseInfo.map((baseItem, index) => (
            <div class={index < formBaseInfo.length - 1 ? 'mb32' : 'mb16'}>
              <div class='font-bold pb8'>
                {baseItem.name}
                {index > 0 ? (
                  <span
                    class={'account-detail-edit-icon-font'}
                    onClick={index === 2 ? handleModifyScret : handleModifyAccount}>
                    {/* <i class={'icon hcm-icon bkhcm-icon-invisible1 pl15 account-edit-icon'}/> */}
                    <i class={'hcm-icon bkhcm-icon-bianji account-edit-icon mr6'} />
                    编辑
                  </span>
                ) : (
                  ''
                )}
              </div>
              <Form model={projectModel} labelWidth={140} rules={formRules} ref={formRef}>
                <div class={index === 2 ? 'flex-row align-items-center flex-wrap' : null}>
                  {baseItem.data.map((formItem) => (
                    <FormItem
                      class='formItem-cls info-value'
                      label={`${formItem.label} ：`}
                      required={formItem.required}
                      property={formItem.property}>
                      {formItem.component()}
                    </FormItem>
                  ))}
                </div>
              </Form>
            </div>
          ))}

          <Dialog
            v-model:isShow={isShowModifyScretDialog.value}
            width={680}
            title={'编辑API密钥'}
            onClosed={onClosed}
            onConfirm={onConfirm}
            isLoading={isSecretDialogLoading.value}
            theme='primary'>
            {{
              default: () => (
                <>
                  <Alert class={'mb12'} theme='info' title='更新的API密钥必须属于同一个主账号ID' />
                  <Form labelWidth={130} model={secretModel} ref={formDiaRef} formType='vertical'>
                    {Object.entries(curExtension.value.input).map(([property, { label }]) => (
                      <FormItem label={label} property={property}>
                        <Input
                          v-model={curExtension.value.input[property].value}
                          type={
                            property === 'cloud_service_secret_key' && projectModel.vendor === VendorEnum.GCP
                              ? 'textarea'
                              : 'text'
                          }
                          rows={8}
                        />
                      </FormItem>
                    ))}
                    {[curExtension.value.output1, curExtension.value.output2].map((output) =>
                      Object.entries(output).map(([property, { label, placeholder }]) => (
                        <FormItem label={label} property={property}>
                          <Input v-model={output[property].value} readonly placeholder={placeholder} />
                        </FormItem>
                      )),
                    )}
                  </Form>
                </>
              ),
              footer: () => (
                <div class={'validate-btn-container'}>
                  <Button
                    outline={curExtension.value.validatedStatus === ValidateStatus.YES}
                    theme='primary'
                    class={'validate-btn'}
                    loading={isValidateLoading.value}
                    onClick={() => handleValidate()}
                    disabled={isValidateDiasbled.value}>
                    账号校验
                  </Button>
                  <Button
                    theme='primary'
                    disabled={isValidateDiasbled.value || curExtension.value.validatedStatus !== ValidateStatus.YES}
                    loading={buttonLoading.value}
                    onClick={onConfirm}>
                    {t('确认')}
                  </Button>
                  <Button class='ml10' onClick={onClosed}>
                    {t('取消')}
                  </Button>
                </div>
              ),
            }}
          </Dialog>

          <Dialog
            isShow={isShowModifyAccountDialog.value}
            width={680}
            title={'编辑账号'}
            isLoading={isAccountDialogLoading.value}
            onConfirm={handleModifyAccountSubmit}
            onClosed={() => (isShowModifyAccountDialog.value = false)}
            theme='primary'>
            <Form v-model={accountFormModel} formType='vertical' model={accountFormModel} ref={accountForm}>
              <FormItem label='责任人' class={'api-secret-selector'} required property='managers'>
                <MemberSelect v-model={accountFormModel.managers} defaultUserlist={computedManagers.value} />
              </FormItem>
              <FormItem label='业务' class={'api-secret-selector'} required property='bk_biz_ids'>
                <Select filterable placeholder='请选择使用业务' v-model={accountFormModel.bk_biz_ids}>
                  {businessList.list.map(({ id, name }) => (
                    <Option key={id} value={id} label={name}>
                      {name}
                    </Option>
                  ))}
                </Select>
              </FormItem>
              <FormItem label='备注'>
                <Input type={'textarea'} v-model={accountFormModel.memo} maxlength={256} resize={false} />
              </FormItem>
            </Form>
          </Dialog>
        </div>
      );
  },
});
