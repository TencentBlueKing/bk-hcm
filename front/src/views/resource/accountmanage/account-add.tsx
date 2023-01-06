import { Form, Input, Select, Button, Radio, Message } from 'bkui-vue';
import { reactive, defineComponent, ref, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ProjectModel, FormItems } from '@/typings';
import { CLOUD_TYPE, ACCOUNT_TYPE, BUSINESS_TYPE, SITE_TYPE } from '@/constants';
import { useI18n } from 'vue-i18n';
import MemberSelect from '@/components/MemberSelect';
import OrganizationSelect from '@/components/OrganizationSelect';
import { useAccountStore, useUserStore } from '@/store';
const { FormItem } = Form;
const { Option } = Select;
const { Group } = Radio;
export default defineComponent({
  name: 'AccountManageAdd',
  setup() {
    const { t } = useI18n();
    const accountStore = useAccountStore();
    const useUser = useUserStore();
    const router = useRouter();

    const initProjectModel: ProjectModel = {
      id: 0,
      type: 'resource',   // 账号类型
      name: '', // 名称
      vendor: '', // 云厂商
      managers: [useUser.username], // 责任人
      departmentId: [],   // 组织架构
      bizIds: '',   // 使用业务
      memo: '',     // 备注
      mainAccount: '',    // 主账号
      subAccount: '',    // 子账号
      subAccountName: '',    // 子账号名称
      secretId: '',    // 密钥id
      secretKey: '',  // 密钥key
      accountId: '',
      iamUsername: '',
      site: 'china',
    };

    onMounted(async () => {
      console.log(122133333);
      /* 获取业务列表接口 */
      getBusinessList();
    });
    const formRef = ref<InstanceType<typeof Form>>(null);
    const noUser = ref<Boolean>(false);
    const noOrganize = ref<Boolean>(false);
    const projectModel = reactive<ProjectModel>({
      ...initProjectModel,
    });

    const optionalRequired = ['secretId', 'secretKey'];
    const cloudType = reactive(CLOUD_TYPE);
    const isTestConnection = ref(false);

    const businessList = reactive({
      list: BUSINESS_TYPE,
    });    // 业务列表
    const getBusinessList = async () => {
      try {
        const res = await accountStore.getBizList();
        console.log(res);
        businessList.list = res?.data || BUSINESS_TYPE;
      } catch (error: any) {
        Message({ theme: 'error', message: error?.message || '系统异常' });
      }
    };


    const check = (val: any): boolean => {
      return  /^[a-z][a-z-z0-9_-]*$/.test(val);
    };

    // 提交操作
    const submit = async () => {
      noOrganize.value = !projectModel.departmentId.length;
      await formRef.value?.validate();
      try {
        const params = {
          vendor: projectModel.vendor,
          spec: {
            type: projectModel.type,
            name: projectModel.name,
            managers: projectModel.managers,
            memo: projectModel.memo,
            department_id: Number(projectModel.departmentId.join(',')),
            site: projectModel.site,
          },
          attachment: {
            bk_biz_ids: projectModel.bizIds.length === businessList.list.length
              ? -1 : projectModel.bizIds,
          },
          extension: {},
        };
        switch (projectModel.vendor) {
          case 'tcloud':
            params.extension = {
              cloud_main_account: projectModel.mainAccount,
              cloud_sub_account: projectModel.subAccount,
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
          default:
            break;
        }
        if (isTestConnection.value) {
          await accountStore.addAccount(params);
          Message({
            message: t('新增成功'),
            theme: 'success',
          });
          router.go(-1);  // 返回列表
        } else {
          await accountStore.testAccountConnection({ vendor: params.vendor, extension: params.extension });
          Message({
            message: t('验证成功'),
            theme: 'success',
          });
          isTestConnection.value = true;
        }
      } catch (error: any) {
        console.log(error);
        Message({ theme: 'error', message: error?.message || '系统异常' });
      }
    };

    const changeCloud = (val: string) => {
      formRef.value?.clearValidate(); // 切换清除表单检验
      const startIndex = formList.findIndex(e => e.property === 'vendor');
      const endIndex = formList.findIndex(e => e.property === 'managers');
      let insertFormData: any = [];
      switch (val) {
        case 'huawei':
          insertFormData = [
            {
              label: t('主账号名'),
              required: true,
              property: 'account',
              component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.account} />,
            },
            {
              label: t('子账号ID'),
              required: true,
              property: 'subAccount',
              component: () => <Input class="w450" placeholder={t('请输入子账号ID')} v-model={projectModel.subAccount} />,
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
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.secretId} />,
            },
            {
              label: 'SecretKey',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />,
            },
          ];
          break;
        case 'aws':
          insertFormData = [
            {
              required: true,
              property: 'site',
              component: () => <Group v-model={projectModel.site}>
                {SITE_TYPE.map(e => (
                  <Radio label={e.value}>{t(e.label)}</Radio>
                ))}
              </Group>,
            },
            {
              label: t('账号ID'),
              required: true,
              property: 'accountId',
              component: () => <Input class="w450" placeholder={t('请输入账号ID')} v-model={projectModel.accountId} />,
            },
            {
              label: t('IAM用户名称'),
              required: true,
              property: 'iamUsername',
              component: () => <Input class="w450" placeholder={t('请输入IAM用户名称')} v-model={projectModel.iamUsername} />,
            },
            {
              label: t('SecretId/密钥ID'),
              required: projectModel.type === 'resource',
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.secretId} />,
            },
            {
              label: 'SecretKey',
              required: projectModel.type === 'resource',
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.secretKey} />,
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
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入服务账号ID')} v-model={projectModel.account} />,
            },
            {
              label: t('服务账号名称'),
              required: projectModel.type === 'resource',
              property: 'secretId',
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
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入订阅名称')} v-model={projectModel.account} />,
            },
            {
              label: t('应用程序(客户端) ID'),
              required: projectModel.type === 'resource',
              property: 'secretId',
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
        case 'tcloud':
          insertFormData = [
            {
              required: true,
              property: 'site',
              component: () => <Group v-model={projectModel.site}>
                {SITE_TYPE.map(e => (
                  <Radio label={e.value}>{t(e.label)}</Radio>
                ))}
              </Group>,
            },
            {
              label: t('主账号ID'),
              required: true,
              property: 'mainAccount',
              component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.mainAccount} />,
            },
            {
              label: t('子账号ID'),
              required: projectModel.type === 'resource',
              property: 'subAccount',
              component: () => <Input class="w450" placeholder={t('请输入子账号')} v-model={projectModel.subAccount} />,
            },
            {
              label: 'SecretId',
              required: projectModel.type === 'resource',
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入SecretId')} v-model={projectModel.secretId} />,
            },
            {
              label: 'SecretKey',
              required: projectModel.type === 'resource',
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />,
            },
          ];
          break;
        default:
          insertFormData = [
            {
              label: t('主账号名'),
              required: true,
              property: 'mainAccount',
              component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.mainAccount} />,
            },
            {
              label: t('子账号ID'),
              required: true,
              property: 'subAccount',
              component: () => <Input class="w450" placeholder={t('请输入子账号ID')} v-model={projectModel.subAccount} />,
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
              property: 'secretId',
              component: () => <Input class="w450" placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.secretId} />,
            },
            {
              label: 'SecretKey',
              required: true,
              property: 'secretKey',
              component: () => <Input class="w450" placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />,
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

    watch(
      () => projectModel.departmentId,
      (val) => {
        noOrganize.value = !val.length;
      },
    );


    const formList = reactive<FormItems[]>([
      {
        required: false,
        component: () => <Group v-model={projectModel.type}>
          {ACCOUNT_TYPE.map(e => (
            <Radio label={e.value}>{t(e.label)}</Radio>
          ))}
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
        property: 'vendor',
        component: () => <Select class="w450" placeholder={t('请选择云厂商')} v-model={projectModel.vendor} onChange={changeCloud}>
          {cloudType.map(item => (
              <Option
                key={item.id}
                value={item.id}
                label={item.name}
              >
                {item.name}
              </Option>
          ))
        }</Select>,
      },
      {
        label: t('主账号'),
        required: true,
        property: 'mainAccount',
        component: () => <Input class="w450" placeholder={t('请输入主账号')} v-model={projectModel.mainAccount} />,
      },
      {
        label: t('子账号ID'),
        required: true,
        property: 'subAccount',
        component: () => <Input class="w450" placeholder={t('请输入子账号ID')} v-model={projectModel.subAccount} />,
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
        property: 'secretId',
        component: () => <Input class="w450" placeholder={t('请输入SecretId/密钥ID')} v-model={projectModel.secretId} />,
      },
      {
        label: 'SecretKey',
        required: true,
        property: 'secretKey',
        component: () => <Input class="w450" placeholder={t('请输入SecretKey')} v-model={projectModel.secretKey} />,
      },
      {
        label: t('责任人'),
        required: true,
        property: 'managers',
        content: () => (
          <section>
            <MemberSelect class="w450" v-model={projectModel.managers}/>
            {noUser.value ? <span class="form-error-tip">责任人不能为空</span> : ''}
          </section>
        ),
      },
      {
        label: t('组织架构'),
        required: true,
        property: 'departmentId',
        content: () => (
          <section>
            <OrganizationSelect class="w450" v-model={projectModel.departmentId} />
            {noOrganize.value ? <span class="form-error-tip">组织架构不能为空</span> : ''}
          </section>
        ),
      },
      {
        label: t('使用业务'),
        required: true,
        property: 'bizIds',
        component: () => <Select multiple show-select-all collapse-tags multipleMode='tag'
        placeholder={t('请选择使用业务')} class="w450" v-model={projectModel.bizIds}>
          {businessList.list.map(item => (
              <Option
                key={item.id}
                value={item.id}
                label={item.name}
              >
                {item.name}
              </Option>
          ))
        }</Select>,
      },
      {
        label: t('备注'),
        required: false,
        property: 'memo',
        component: () => <Input class="w450" placeholder={t('请输入备注')} v-model={projectModel.memo} type="textarea" maxlength={100} showWordLimit rows={2} />,
      },
      {
        required: false,
        component: () => <Button theme="primary" onClick={submit}>{t(isTestConnection.value ? t('确认') : t('账号验证'))}</Button>,
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
