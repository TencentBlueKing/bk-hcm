import { Form, Input, Select, Button, Radio, Message } from 'bkui-vue';
import { reactive, defineComponent, ref, watch, onMounted, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { ProjectModel, FormItems } from '@/typings';
import { CLOUD_TYPE, ACCOUNT_TYPE, BUSINESS_TYPE, SITE_TYPE, DESC_ACCOUNT } from '@/constants';
import { VendorEnum } from '@/common/constant';
import { useI18n } from 'vue-i18n';
import MemberSelect from '@/components/MemberSelect';
import { useAccountStore } from '@/store';
import './index.scss';
const { FormItem } = Form;
const { Option } = Select;
const { Group, Group: RadioGroup, Button: RadioButton } = Radio;
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const { t } = useI18n();
    const accountStore = useAccountStore();
    const router = useRouter();

    const initProjectModel: ProjectModel = {
      id: 0,
      type: 'registration', // 账号类型
      name: '', // 名称
      vendor: VendorEnum.TCLOUD, // 云厂商
      managers: [], // 责任人
      bizIds: [], // 使用业务
      memo: '', // 备注
      mainAccount: '', // 主账号
      subAccount: '', // 子账号
      subAccountName: '', // 子账号名称
      secretId: '', // 密钥id
      secretKey: '', // 密钥key
      accountId: '',
      accountName: '', // 账号名称
      iamUsername: '',
      iamUserId: '',
      site: 'china',
      projectName: '', // 项目名称
      projectId: '', // 项目ID
      tenantId: '', // 租户ID
      subScriptionId: '', // 订阅ID
      subScriptionName: '', // 订阅名称
      applicationId: '', // 应用程序ID
      applicationName: '', // 应用程序名称
    };

    onMounted(async () => {
      /* 获取业务列表接口 */
      getBusinessList();
    });
    const formRef = ref<InstanceType<typeof Form>>(null);
    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const optionalRequired: string[] = [
      'secretId',
      'secretKey',
      'accountName',
      'accountId',
      'applicationId',
      'applicationName',
      'bizIds',
    ];
    const requiredData: string[] = ['secretId', 'secretKey', 'bizIds'];
    const cloudType = reactive(CLOUD_TYPE);
    const submitLoading = ref(false);
    const isChangeVendor = ref(false);

    const businessList = reactive({
      list: BUSINESS_TYPE,
    }); // 业务列表
    const getBusinessList = async () => {
      try {
        const res = await accountStore.getBizList();
        businessList.list = res?.data || BUSINESS_TYPE;
      } catch (error: any) {
        Message({ theme: 'error', message: error?.message || '系统异常' });
      }
    };

    const check = (val: any): boolean => {
      return /^[a-z][a-z-z0-9_-]*$/.test(val);
    };

    // 提交操作
    const submit = async () => {
      await formRef.value?.validate();
      submitLoading.value = true;
      const typeNamePrefix = {
        registration: 'hcm',
        security_audit: 'sec',
      };
      const vendorAccountMains = {
        tcloud: projectModel.mainAccount,
        aws: projectModel.accountId,
      };
      const vendorAccountSubs = {
        tcloud: projectModel.subAccount,
        aws: projectModel.iamUsername,
      };
      try {
        const params = {
          vendor: projectModel.vendor,
          type: projectModel.type,
          // 最新修改名称不在页面输入且只支持tcloud，名称使用默认规则拼接
          name:
            projectModel.name ||
            [
              typeNamePrefix[projectModel.type],
              vendorAccountMains[projectModel.vendor],
              vendorAccountSubs[projectModel.vendor],
            ].join('-'),
          managers: projectModel.managers,
          memo: projectModel.memo,
          site: projectModel.site,
          bk_biz_ids: Array.isArray(projectModel.bizIds) ? [-1] : [projectModel.bizIds],
          extension: {},
        };
        switch (projectModel.vendor) {
          case 'tcloud':
            params.extension = {
              cloud_main_account_id: projectModel.mainAccount,
              cloud_sub_account_id: projectModel.subAccount,
              cloud_secret_id: projectModel.secretId,
              cloud_secret_key: projectModel.secretKey,
            };
            break;
          case 'aws':
            params.extension = {
              cloud_account_id: projectModel.accountId,
              cloud_iam_username: projectModel.iamUsername,
              cloud_secret_id: projectModel.secretId,
              cloud_secret_key: projectModel.secretKey,
            };
            break;
          case 'huawei':
            params.extension = {
              cloud_main_account_name: projectModel.mainAccount,
              cloud_sub_account_id: projectModel.subAccount,
              cloud_sub_account_name: projectModel.subAccountName,
              cloud_secret_id: projectModel.secretId,
              cloud_secret_key: projectModel.secretKey,
              cloud_iam_username: projectModel.iamUsername,
              cloud_iam_user_id: projectModel.iamUserId,
            };
            break;
          case 'gcp':
            params.extension = {
              cloud_project_id: projectModel.projectId,
              cloud_project_name: projectModel.projectName,
              cloud_service_account_id: projectModel.accountId,
              cloud_service_account_name: projectModel.accountName,
              cloud_service_secret_id: projectModel.secretId,
              cloud_service_secret_key: projectModel.secretKey,
            };
            break;
          case 'azure':
            params.extension = {
              cloud_tenant_id: projectModel.tenantId,
              cloud_subscription_id: projectModel.subScriptionId,
              cloud_subscription_name: projectModel.subScriptionName,
              cloud_application_id: projectModel.applicationId,
              cloud_application_name: projectModel.applicationName,
              cloud_client_secret_id: projectModel.secretId,
              cloud_client_secret_key: projectModel.secretKey,
            };
            break;
          default:
            break;
        }

        // 只有安全审计账号才需要密钥相关参数
        if (projectModel.type !== 'security_audit') {
          Reflect.deleteProperty(params.extension, 'cloud_secret_id');
          Reflect.deleteProperty(params.extension, 'cloud_secret_key');
        }

        // 安全审计账号类型aws中国站不调用test
        if (
          !(
            projectModel.type === 'security_audit' &&
            projectModel.vendor === VendorEnum.AWS &&
            projectModel.site === 'china'
          )
        ) {
          await accountStore.testAccountConnection({
            vendor: params.vendor,
            type: projectModel.type,
            extension: params.extension,
          });
        }

        await accountStore.applyAccount(params);
        Message({
          message: t('提交申请成功'),
          theme: 'success',
        });
        // router.go(-1);
        router.push({
          path: '/service/my-apply', // 返回审批列表
        });
      } catch (error: any) {
        console.error(error);
      } finally {
        submitLoading.value = false;
      }
    };

    onMounted(() => {
      changeCloud(projectModel.vendor);
    });

    const changeCloud = (val: string) => {
      isChangeVendor.value = true;
      nextTick(() => {
        formRef.value?.clearValidate(); // 切换清除表单检验
      });
      const startIndex = formList.findIndex((e) => e.property === 'vendor');
      const endIndex = formList.findIndex((e) => e.property === 'managers');
      let insertFormData: any = [];
      switch (val) {
        case 'huawei':
          insertFormData = [
            {
              required: true,
              property: 'site',
              component: () => (
                <Group v-model={projectModel.site}>
                  {SITE_TYPE.map((e) => (
                    <Radio disabled={e.value === 'china'} label={e.value}>
                      {t(e.label)}
                    </Radio>
                  ))}
                </Group>
              ),
            },
            {
              label: t('主账号名'),
              formName: t('账号信息'),
              noBorBottom: true,
              required: true,
              property: 'mainAccount',
              component: () => (
                <Input class='w450' placeholder={t('请输入主账号')} v-model_trim={projectModel.mainAccount} />
              ),
            },
            {
              label: t('账号ID'),
              noBorBottom: true,
              required: true,
              property: 'subAccount',
              component: () => (
                <Input class='w450' placeholder={t('请输入子账号ID')} v-model_trim={projectModel.subAccount} />
              ),
            },
            {
              label: t('账号名称'),
              noBorBottom: true,
              required: true,
              property: 'subAccountName',
              component: () => (
                <Input class='w450' placeholder={t('请输入子账号名称')} v-model_trim={projectModel.subAccountName} />
              ),
            },
            {
              label: t('IAM用户ID'),
              noBorBottom: true,
              required: true,
              property: 'iamUserId',
              component: () => (
                <Input class='w450' placeholder={t('请输入IAM用户ID')} v-model_trim={projectModel.iamUserId} />
              ),
            },
            {
              label: t('IAM用户名称'),
              required: true,
              property: 'iamUsername',
              component: () => (
                <Input class='w450' placeholder={t('请输入IAM用户名称')} v-model_trim={projectModel.iamUsername} />
              ),
            },
            // {
            //   label: t('SecretId/密钥ID'),
            //   formName: t('API 密钥'),
            //   noBorBottom: true,
            //   required: projectModel.type !== 'registration',
            //   property: 'secretId',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.secretId} />
            //   ),
            // },
            // {
            //   label: 'SecretKey',
            //   required: projectModel.type !== 'registration',
            //   property: 'secretKey',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />
            //   ),
            // },
          ];
          projectModel.site = 'international';
          break;
        case 'aws':
          insertFormData = [
            {
              required: true,
              property: 'site',
              component: () => (
                <Group v-model={projectModel.site}>
                  {SITE_TYPE.map((e) => (
                    <Radio label={e.value}>{t(e.label)}</Radio>
                  ))}
                </Group>
              ),
            },
            {
              label: t('账号ID'),
              formName: t('账号信息'),
              noBorBottom: true,
              required: true,
              property: 'accountId',
              component: () => (
                <Input class='w450' placeholder={t('请输入账号ID')} v-model_trim={projectModel.accountId} />
              ),
            },
            {
              label: t('IAM用户名称'),
              required: true,
              property: 'iamUsername',
              component: () => (
                <Input class='w450' placeholder={t('请输入IAM用户名称')} v-model_trim={projectModel.iamUsername} />
              ),
            },
            {
              label: t('SecretId/密钥ID'),
              formName: t('API 密钥'),
              noBorBottom: true,
              required: projectModel.type !== 'registration',
              hidden: projectModel.type === 'registration',
              property: 'secretId',
              component: () => (
                <Input class='w450' placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.secretId} />
              ),
            },
            {
              label: 'SecretKey',
              required: projectModel.type !== 'registration',
              hidden: projectModel.type === 'registration',
              property: 'secretKey',
              component: () => (
                <Input class='w450' placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />
              ),
            },
          ];
          projectModel.site = 'international';
          break;
        case 'gcp':
          insertFormData = [
            {
              required: true,
              property: 'site',
              component: () => (
                <Group v-model={projectModel.site}>
                  {SITE_TYPE.map((e) => (
                    <Radio disabled={e.value === 'china'} label={e.value}>
                      {t(e.label)}
                    </Radio>
                  ))}
                </Group>
              ),
            },
            {
              label: t('项目 ID'),
              formName: t('账号信息'),
              noBorBottom: true,
              required: true,
              property: 'projectId',
              component: () => (
                <Input class='w450' placeholder={t('请输入项目 ID')} v-model_trim={projectModel.projectId} />
              ),
            },
            {
              label: t('项目名称'),
              required: true,
              property: 'projectName',
              component: () => (
                <Input class='w450' placeholder={t('请输入项目名称')} v-model_trim={projectModel.projectName} />
              ),
            },
            // {
            //   label: t('服务账号ID'),
            //   formName: t('API 密钥'),
            //   noBorBottom: true,
            //   required: projectModel.type !== 'registration',
            //   property: 'accountId',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入服务账号ID')} v-model={projectModel.accountId} />
            //   ),
            // },
            // {
            //   label: t('服务账号名称'),
            //   noBorBottom: true,
            //   required: projectModel.type !== 'registration',
            //   property: 'accountName',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入服务账号名称')} v-model={projectModel.accountName} />
            //   ),
            // },
            // {
            //   label: '服务账号密钥ID',
            //   noBorBottom: true,
            //   required: projectModel.type !== 'registration',
            //   property: 'secretId',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入服务账号密钥ID')} v-model={projectModel.secretId} />
            //   ),
            // },
            // {
            //   label: '服务账号密钥',
            //   required: projectModel.type !== 'registration',
            //   property: 'secretKey',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入服务账号密钥')} v-model={projectModel.secretKey} />
            //   ),
            // },
          ];
          projectModel.site = 'international';
          break;
        case 'azure':
          insertFormData = [
            {
              required: true,
              property: 'site',
              component: () => (
                <Group v-model={projectModel.site}>
                  {SITE_TYPE.map((e) => (
                    <Radio disabled={e.value === 'china'} label={e.value}>
                      {t(e.label)}
                    </Radio>
                  ))}
                </Group>
              ),
            },
            {
              label: t('租户 ID'),
              formName: t('账号信息'),
              noBorBottom: true,
              required: true,
              property: 'tenantId',
              component: () => (
                <Input class='w450' placeholder={t('请输入租户 ID')} v-model_trim={projectModel.tenantId} />
              ),
            },
            {
              label: t('订阅 ID'),
              required: true,
              noBorBottom: true,
              property: 'subScriptionId',
              component: () => (
                <Input class='w450' placeholder={t('请输入订阅 ID')} v-model_trim={projectModel.subScriptionId} />
              ),
            },
            {
              label: t('订阅名称'),
              required: true,
              property: 'subScriptionName',
              component: () => (
                <Input class='w450' placeholder={t('请输入订阅名称')} v-model_trim={projectModel.subScriptionName} />
              ),
            },
            // {
            //   label: t('应用(客户端) ID'),
            //   formName: t('API 密钥'),
            //   noBorBottom: true,
            //   required: projectModel.type !== 'registration',
            //   property: 'applicationId',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入应用程序(客户端) ID')} v-model={projectModel.applicationId} />
            //   ),
            // },
            // {
            //   label: t('应用程序名称'),
            //   noBorBottom: true,
            //   required: projectModel.type !== 'registration',
            //   property: 'applicationName',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入应用程序名称')} v-model={projectModel.applicationName} />
            //   ),
            // },
            // {
            //   label: t('客户端密钥ID'),
            //   noBorBottom: true,
            //   required: projectModel.type !== 'registration',
            //   property: 'secretId',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入客户端密钥ID')} v-model={projectModel.secretId} />
            //   ),
            // },
            // {
            //   label: t('客户端密钥'),
            //   required: projectModel.type !== 'registration',
            //   property: 'secretKey',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入客户端密钥')} v-model={projectModel.secretKey} />
            //   ),
            // },
          ];
          projectModel.site = 'international';
          break;
        case 'tcloud':
          insertFormData = [
            {
              required: true,
              property: 'site',
              component: () => (
                <Group v-model={projectModel.site}>
                  {SITE_TYPE.map((e) => (
                    <Radio label={e.value}>{t(e.label)}</Radio>
                  ))}
                </Group>
              ),
            },
            {
              label: t('主账号ID'),
              formName: t('账号信息'),
              noBorBottom: true,
              required: true,
              property: 'mainAccount',
              rules: [{ pattern: /^\d+$/, message: '必须为数值', trigger: 'change' }],
              component: () => (
                <Input class='w450' placeholder={t('请输入主账号')} v-model_trim={projectModel.mainAccount} />
              ),
            },
            {
              label: t('子账号ID'),
              required: true,
              property: 'subAccount',
              rules: [{ pattern: /^\d+$/, message: '必须为数值', trigger: 'change' }],
              component: () => (
                <Input class='w450' placeholder={t('请输入子账号')} v-model_trim={projectModel.subAccount} />
              ),
            },
            // {
            //   label: 'SecretId',
            //   formName: t('API 密钥'),
            //   noBorBottom: true,
            //   required: projectModel.type !== 'registration',
            //   property: 'secretId',
            //   component: () => <Input class='w450' placeholder={t('请输入SecretId')} v-model={projectModel.secretId} />,
            // },
            // {
            //   label: 'SecretKey',
            //   required: projectModel.type !== 'registration',
            //   property: 'secretKey',
            //   component: () => (
            //     <Input class='w450' placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />
            //   ),
            // },
          ];
          break;
        default:
          insertFormData = [
            {
              label: t('主账号名'),
              formName: t('账号信息'),
              noBorBottom: true,
              required: true,
              property: 'mainAccount',
              component: () => (
                <Input class='w450' placeholder={t('请输入主账号')} v-model_trim={projectModel.mainAccount} />
              ),
            },
            {
              label: t('子账号ID'),
              noBorBottom: true,
              required: true,
              property: 'subAccount',
              component: () => (
                <Input class='w450' placeholder={t('请输入子账号ID')} v-model_trim={projectModel.subAccount} />
              ),
            },
            {
              label: t('子账号名称'),
              required: true,
              property: 'subAccountName',
              component: () => (
                <Input class='w450' placeholder={t('请输入子账号名称')} v-model_trim={projectModel.subAccountName} />
              ),
            },
            {
              label: t('SecretId'),
              formName: t('API 密钥'),
              noBorBottom: true,
              required: true,
              property: 'secretId',
              component: () => (
                <Input class='w450' placeholder={t('请输入SecretId')} v-model_trim={projectModel.secretId} />
              ),
            },
            {
              label: 'SecretKey',
              required: true,
              property: 'secretKey',
              component: () => (
                <Input class='w450' placeholder={t('请输入SecretKey')} v-model_trim={projectModel.secretKey} />
              ),
            },
          ];
          break;
      }
      const interceLength = endIndex - startIndex - 1; // 需要删除的长度
      formList.splice(startIndex + 1, interceLength, ...insertFormData);
    };

    // 表单检验
    const formRules = {
      name: [
        {
          trigger: 'blur',
          message: '名称必须以小写字母开头，后面最多可跟 32个小写字母、数字或连字符，但不能以连字符结尾',
          validator: check,
        },
      ],
    };

    watch(
      () => projectModel.type,
      (val, oldValue) => {
        formRef.value?.clearValidate(); // 切换清除表单检验
        if (val === 'registration') {
          // 登记账号
          formList?.forEach((e) => {
            if (optionalRequired.includes(e.property)) {
              e.required = false;
            }
            if (projectModel.vendor === 'aws' && ['secretId', 'secretKey'].includes(e.property)) {
              e.hidden = true;
            }
          });
        } else if (val === 'resource') {
          // 资源账号
          formList?.forEach((e) => {
            if (e.label && requiredData.includes(e.property)) {
              // 资源账号必填项
              e.required = true;
            }
          });
        } else {
          formList?.forEach((e) => {
            if (e.label && (e.property === 'memo' || e.property === 'bizIds')) {
              // 备注、使用业务不需必填
              e.required = false;
            }
            if (projectModel.vendor === 'aws' && ['secretId', 'secretKey'].includes(e.property)) {
              e.hidden = false;
            }
          });
        }

        // 安全审计账号暂只支持aws
        if (val === 'security_audit') {
          projectModel.vendor = VendorEnum.AWS;
        }

        // 触发一次云厂商变更，因展示字段需要更新
        if (oldValue !== undefined && val !== oldValue) {
          changeCloud(projectModel.vendor);
        }
      },
      { immediate: true },
    );

    const formList = reactive<FormItems[]>([
      {
        label: t('账号类型'),
        formName: t('账号用途'),
        required: false,
        component: () => (
          <Group v-model={projectModel.type}>
            {ACCOUNT_TYPE.map((e) => (
              <Radio label={e.value}>{t(e.label)}</Radio>
            ))}
          </Group>
        ),
      },
      {
        label: t('云厂商'),
        formName: t('云厂商'),
        required: true,
        property: 'vendor',
        component: () => (
          <RadioGroup v-model={projectModel.vendor}>
            {cloudType.map((item) => (
              <RadioButton
                onChange={changeCloud}
                label={item.id}
                disabled={
                  !['tcloud', 'aws'].includes(item.id) ||
                  (projectModel.type === 'security_audit' && item.id === 'tcloud')
                }>
                {item.name}
              </RadioButton>
            ))}
          </RadioGroup>
        ),
      },
      {
        label: t('主账号'),
        formName: t('账号信息'),
        noBorBottom: true,
        required: true,
        property: 'mainAccount',
        component: () => <Input class='w450' placeholder={t('请输入主账号')} v-model={projectModel.mainAccount} />,
      },
      {
        label: t('子账号ID'),
        noBorBottom: true,
        required: true,
        property: 'subAccount',
        component: () => <Input class='w450' placeholder={t('请输入子账号ID')} v-model={projectModel.subAccount} />,
      },
      {
        label: t('子账号名称'),
        required: true,
        property: 'subAccountName',
        component: () => (
          <Input class='w450' placeholder={t('请输入子账号名称')} v-model={projectModel.subAccountName} />
        ),
      },
      {
        label: t('SecretId/密钥ID'),
        formName: t('API 密钥'),
        noBorBottom: true,
        required: true,
        property: 'secretId',
        component: () => (
          <Input class='w450' placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.secretId} />
        ),
      },
      {
        label: 'SecretKey',
        required: true,
        property: 'secretKey',
        component: () => <Input class='w450' placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />,
      },
      {
        label: t('责任人'),
        formName: t('账号归属'),
        noBorBottom: true,
        required: true,
        property: 'managers',
        content: () => (
          <section>
            <MemberSelect class='w450' v-model={projectModel.managers} />
          </section>
        ),
      },
      {
        label: t('使用业务'),
        noBorBottom: true,
        required: true,
        property: 'bizIds',
        component: () => (
          <Select
            filterable
            collapse-tags
            multipleMode='tag'
            placeholder={t('请选择使用业务')}
            class='w450'
            v-model={projectModel.bizIds}>
            {businessList.list.map((item) => (
              <Option key={item.id} value={item.id} label={item.name}>
                {item.name}
              </Option>
            ))}
          </Select>
        ),
      },
      {
        label: t('备注'),
        required: false,
        property: 'memo',
        component: () => (
          <Input
            class='w450'
            placeholder={t('请输入备注')}
            v-model={projectModel.memo}
            type='textarea'
            maxlength={100}
            showWordLimit
            rows={2}
          />
        ),
      },
      {
        required: false,
        type: 'button',
        component: () => (
          <Button theme='primary' loading={submitLoading.value} onClick={submit}>
            {t('提交审批')}
          </Button>
        ),
      },
    ]);

    return () => (
      <div class='form-container flex-row justify-content-between'>
        <Form class='form-warp' model={projectModel} labelWidth={140} rules={formRules} ref={formRef}>
          {formList
            .filter((item) => item.hidden !== true)
            .map((item) => (
              <>
                {item.formName && <div class='mt10 mb10'>{item.formName}</div>}
                <div
                  class={{
                    'form-item-warp': true,
                    'no-border-top': !item.formName,
                    'no-border-bottom': item.noBorBottom || (item.property === 'vendor' && isChangeVendor.value),
                    'no-border': item.type === 'button',
                  }}>
                  <FormItem
                    class='account-form-item'
                    label={item.label}
                    required={item.required}
                    property={item.property}
                    description={item.description}
                    rules={item.rules}>
                    {item.component ? item.component() : item.content()}
                  </FormItem>
                </div>
              </>
            ))}
        </Form>
        <div class='desc-container flex-1'>
          <div class='desc-item'>
            <div class='title mb10'>账号类型</div>
            <div class='desc'>
              <p>资源账号：用于从云上同步、更新、操作、购买资源的账号，需要API密钥。</p>
              <p>登记账号：云上的普通登录用户，用于被安全审计的账号对象。</p>
              <p>安全审计账号：用于对云上资源进行安全审计的账号，需要API密钥，权限比资源账号低</p>
            </div>
          </div>
          <div class='desc-item'>
            <div class='title mb10 mt10'>云厂商</div>
            <div class='desc'>
              <p v-html={DESC_ACCOUNT[projectModel.vendor]?.vendor}></p>
            </div>
          </div>
          <div class='desc-item'>
            <div class='title mb10 mt10'>账号名称</div>
            <div class='desc'>
              <p>用于标识账号的用途，增强可读性。</p>
            </div>
          </div>
          <div class='desc-item'>
            <div class='title mb10 mt10'>账号信息</div>
            <div class='desc'>
              <p v-html={DESC_ACCOUNT[projectModel.vendor]?.accountInfo}></p>
            </div>
          </div>
          <div class='desc-item'>
            <div class='title mb10 mt10'>API密钥</div>
            <div class='desc'>
              <p v-html={DESC_ACCOUNT[projectModel.vendor]?.apiSecret}></p>
            </div>
          </div>
          <div class='desc-item'>
            <div class='title mb10 mt10'>账号归属</div>
            <div class='desc'>
              <p>使用业务：该账号所属的业务。账号绑定到业务，则该业务具有该账号的资源管理权限。</p>
              <p>责任人：该账号的负责人，请填写2个以上的负责人，最多支持5个负责人。</p>
            </div>
          </div>
        </div>
      </div>
    );
  },
});
