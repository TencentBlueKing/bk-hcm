import { Alert, Button, Card, Dialog, Form, Input, Loading, Radio, Select, Table } from 'bkui-vue';
import { PropType, defineComponent, onMounted, reactive, ref, watch, watchEffect } from 'vue';
import './index.scss';
import { VendorEnum } from '@/common/constant';
import tcloudVendor from '@/assets/image/vendor-tcloud.svg';
import awsVendor from '@/assets/image/vendor-aws.svg';
import azureVendor from '@/assets/image/vendor-azure.svg';
import gcpVendor from '@/assets/image/vendor-gcp.svg';
import huaweiVendor from '@/assets/image/vendor-huawei.svg';
import { Success, InfoLine, TextFile } from 'bkui-vue/lib/icon';
import http from '@/http';
import successIcon from '@/assets/image/corret-fill.png';
import failedIcon from '@/assets/image/delete-fill.png';
import MemberSelect from '@/components/MemberSelect';
import { useAccountStore, useUserStore } from '@/store';
import { ValidateStatus, useSecretExtension } from './useSecretExtension';

const { FormItem } = Form;
const { Option } = Select;
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const VENDORS_INFO = [
  {
    vendor: VendorEnum.TCLOUD,
    name: '腾讯云',
    icon: tcloudVendor,
  },
  {
    vendor: VendorEnum.AWS,
    name: '亚马逊云',
    icon: awsVendor,
  },
  {
    vendor: VendorEnum.AZURE,
    name: '微软云',
    icon: azureVendor,
  },
  {
    vendor: VendorEnum.GCP,
    name: '谷歌云',
    icon: gcpVendor,
  },
  {
    vendor: VendorEnum.HUAWEI,
    name: '华为云',
    icon: huaweiVendor,
  },
];

export default defineComponent({
  props: {
    changeEnableNextStep: {
      type: Function as PropType<(val: boolean) => void>,
      required: true,
    },
    changeSubmitData: {
      type: Function as PropType<(val: Record<string, string | Object>) => void>,
      required: true,
    },
    changeValidateForm: {
      type: Function as PropType<(callback: () => Promise<void>) => void>,
      required: true,
    },
    changeExtension: {
      type: Function as PropType<(extension: Record<string, string>) => void>,
      required: true,
    },
  },
  setup(props) {
    const userStore = useUserStore();
    const formModel = reactive({
      site: 'international' as 'china' | 'international', // 站点
      vendor: VendorEnum.TCLOUD, // 云厂商
      name: '', // 账号别名
      managers: [], // 责任人
      type: 'resource', // 账号类型，当前产品形态固定为 resource，资源账号
      memo: '', // 备注
      extension: {}, // 不同云的secretKey\id
      bk_biz_ids: [], // 业务ID
    });
    const infoFormInstance = ref(null);
    const businessList = ref([]);
    const accountStore = useAccountStore();
    const isAuthDialogShow = ref(false);
    const isAuthTableLoading = ref(false);
    const authTableData = ref([]);

    watchEffect(() => {
      formModel.managers = [userStore.username];
    });

    const { curExtension, tcloudExtension, handleValidate, isValidateLoading, isValidateDiasbled } =
      useSecretExtension(formModel);
    watch(
      () => curExtension.value.validatedStatus,
      (val) => {
        props.changeEnableNextStep(val === ValidateStatus.YES);
        if (val === ValidateStatus.YES) {
          formModel.extension = Object.entries({
            ...curExtension.value.input,
            ...curExtension.value.output1,
            ...curExtension.value.output2,
          }).reduce((prev, [key, { value }]) => {
            prev[key] = value;
            return prev;
          }, {});
        }
      },
      {
        deep: true,
      },
    );

    watch(
      () => formModel.vendor,
      () => {
        if ([VendorEnum.AZURE, VendorEnum.GCP, VendorEnum.HUAWEI].includes(formModel.vendor)) {
          formModel.site = 'international';
        }
      },
    );

    watch(
      () => formModel,
      (model) => {
        props.changeSubmitData({
          ...model,
          bk_biz_ids: [model.bk_biz_ids],
        });
        props.changeValidateForm(() => infoFormInstance.value.validate());
      },
      {
        deep: true,
      },
    );

    watch(
      () => isAuthDialogShow.value,
      async (isShow) => {
        if (!isShow) return;
        isAuthTableLoading.value = true;
        const res = await http.post(
          `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/tcloud/accounts/auth_policies/list`,
          {
            cloud_secret_id: tcloudExtension.input.cloud_secret_id.value,
            cloud_secret_key: tcloudExtension.input.cloud_secret_key.value,
            uin: +tcloudExtension.output1.cloud_sub_account_id.value,
          },
        );
        authTableData.value = res.data?.[0]?.Policy;
        isAuthTableLoading.value = false;
      },
    );
    onMounted(async () => {
      const res = await accountStore.getBizList();
      businessList.value = res?.data || [];
    });

    return () => (
      <div class={'account-form'}>
        <Card class={'account-form-card'} showHeader={false}>
          <p class={'account-form-card-title'}>账号归属</p>
          <div class={'account-form-card-content'}>
            <Form formType='vertical'>
              <FormItem label='厂商选择' required>
                <div class={'account-vendor-selector'}>
                  {VENDORS_INFO.map(({ vendor, name, icon }) => (
                    <div
                      class={`account-vendor-option ${
                        vendor === formModel.vendor ? 'account-vendor-option-active' : ''
                      }`}
                      onClick={() => (formModel.vendor = vendor)}>
                      <img src={icon} alt={name} class={'account-vendor-option-icon'} />
                      <p class={'account-vendor-option-text'}>{name}</p>
                      {formModel.vendor === vendor ? <Success fill='#3A84FF' class={'active-icon'} /> : null}
                    </div>
                  ))}
                </div>
              </FormItem>
              <FormItem required label='站点种类'>
                {![VendorEnum.AZURE, VendorEnum.GCP].includes(formModel.vendor) ? (
                  <Radio
                    label={'china'}
                    v-model={formModel.site}
                    disabled={[VendorEnum.HUAWEI, VendorEnum.AWS].includes(formModel.vendor)}>
                    中国站
                  </Radio>
                ) : null}
                <Radio label={'international'} v-model={formModel.site}>
                  国际站
                </Radio>
              </FormItem>
            </Form>
          </div>
        </Card>

        <Card class={'account-form-card'} showHeader={false}>
          <>
            <div class={'api-secret-header'}>
              <p class={'account-form-card-title'}>API 密钥</p>
              <InfoLine fill='#979BA5' />
              <p class={'header-text'}>同一个主账号下,只允许接入一次。如后续对API密钥更新,必须是隶属于同一主账号。</p>
              <div class='btn-guide-wrap'>
                <TextFile fill='#3A84FF' />
                <Button theme='primary' text class={'header-btn'}>
                  接入指引
                </Button>
              </div>
            </div>
            <div class={'account-form-card-content'}>
              <Form formType='vertical' class={'account-form-card-content-grid'}>
                <div>
                  {Object.entries(curExtension.value.input).map(([property, { label }]) => (
                    <FormItem label={label} property={property} required>
                      <Input
                        v-model={curExtension.value.input[property].value}
                        type={
                          property === 'cloud_service_secret_key' && formModel.vendor === VendorEnum.GCP
                            ? 'textarea'
                            : 'text'
                        }
                        rows={8}
                        resize={!(formModel.vendor === VendorEnum.GCP)}
                      />
                    </FormItem>
                  ))}
                </div>
                <div class={'account-form-card-content-grid-right'}>
                  {formModel.vendor === VendorEnum.TCLOUD && tcloudExtension.validatedStatus === ValidateStatus.YES ? (
                    <Button
                      text
                      theme='primary'
                      class={'api-form-btn'}
                      onClick={() => {
                        isAuthDialogShow.value = true;
                      }}>
                      <TextFile fill='#3A84FF' />
                      查看账号权限
                    </Button>
                  ) : null}
                  {Object.entries(curExtension.value.output1).map(([property, { label, value, placeholder }]) => (
                    <FormItem label={label} property={property}>
                      <Input v-model={value} readonly placeholder={placeholder} />
                    </FormItem>
                  ))}
                </div>
              </Form>
            </div>
            <div class={'validate-btn-block'}>
              <Button
                theme='primary'
                outline={curExtension.value.validatedStatus === ValidateStatus.YES}
                class={'account-validate-btn'}
                onClick={() => handleValidate((payload: Record<string, string>) => props.changeExtension(payload))}
                disabled={isValidateDiasbled.value}
                loading={isValidateLoading.value}>
                账号校验
              </Button>
              {curExtension.value.validatedStatus === ValidateStatus.YES ? (
                <>
                  <img src={successIcon} alt='success' class={'validate-icon'}></img>
                  <span> 校验成功 </span>
                </>
              ) : null}
              {curExtension.value.validatedStatus === ValidateStatus.NO ? (
                <>
                  <img src={failedIcon} alt='success' class={'validate-icon'}></img>
                  <span> 校验失败 {curExtension.value.validateFailedReason}</span>
                </>
              ) : null}
            </div>
          </>
        </Card>

        <Card class={'account-form-card'} showHeader={false}>
          <p class={'account-form-card-title'}>其他信息</p>
          <div class={'account-form-card-content'}>
            <Form
              formType='vertical'
              model={formModel}
              auto-check
              ref={infoFormInstance}
              rules={{
                name: [
                  {
                    trigger: 'blur',
                    message:
                      '必须以小写字母开头, 后面可跟小写字母、数字、连字符 - 或 下划线 _ , 但不能以连字符 - 或下划线 _ 结尾。名称长度不少于 3 个字符，且不多于 64 个字符',
                    validator: (val: any): boolean => {
                      // return /^[a-z][a-z0-9-]{1,31}[a-z0-9]$/.test(val);
                      return /^[a-z][a-z0-9-_]{1,62}[a-z0-9]$/.test(val);
                    },
                  },
                ],
              }}>
              {/* eslint-disable-next-line @typescript-eslint/no-unused-vars */}
              {Object.entries(curExtension.value.output2).map(([, { label, value, placeholder }]) => (
                <FormItem label={label}>
                  <Input v-model={value} readonly placeholder={placeholder} />
                </FormItem>
              ))}
              <FormItem
                label='账号别名'
                class={'api-secret-selector'}
                required
                property='name'
                description={
                  '必须以小写字母开头，后面可跟小写字母、数字、连字符 - 或 下划线 _ ，但不能以连字符 - 或下划线 _ 结尾。\n名称长度不少于 3 个字符，且不多于 64 个字符。'
                }>
                <Input v-model={formModel.name} />
              </FormItem>
              <FormItem label='责任人' class={'api-secret-selector'} required property='managers'>
                <MemberSelect
                  v-model={formModel.managers}
                  defaultUserlist={[
                    {
                      username: userStore.username,
                      display_name: userStore.username,
                    },
                  ]}
                />
              </FormItem>
              <FormItem label='使用业务' property='bk_biz_ids' required>
                <Select filterable placeholder='请选择使用业务' v-model={formModel.bk_biz_ids}>
                  {businessList.value.map(({ id, name }) => (
                    <Option key={id} value={id} label={name}>
                      {name}
                    </Option>
                  ))}
                </Select>
              </FormItem>
              <FormItem label='备注'>
                <Input type={'textarea'} v-model={formModel.memo} maxlength={255} resize={false} />
              </FormItem>
            </Form>
          </div>
        </Card>

        <Dialog
          isShow={isAuthDialogShow.value}
          onClosed={() => (isAuthDialogShow.value = false)}
          dialogType='show'
          theme='primary'
          title='账号权限详情'
          width={900}>
          <Alert theme='info' class={'mb16'}>
            该账号在云上拥有的权限组列表如下，如需调整权限请到
            <Button
              theme='primary'
              text
              onClick={() => {
                isAuthDialogShow.value = false;
                window.open('https://console.cloud.tencent.com/cam/overview', '_blank', 'noopener,noreferrer');
              }}>
              云控制台
            </Button>
            调整
          </Alert>
          <Loading loading={isAuthTableLoading.value}>
            <Table
              columns={[
                {
                  label: '权限组名称',
                  field: 'PolicyName',
                  width: 200,
                },
                {
                  label: '描述',
                  field: 'PolicyDescription',
                },
              ]}
              data={authTableData.value}
            />
          </Loading>
        </Dialog>
      </div>
    );
  },
});
