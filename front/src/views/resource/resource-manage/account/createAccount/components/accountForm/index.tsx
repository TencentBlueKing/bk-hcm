import { Button, Card, Form, Input, Radio } from 'bkui-vue';
import { PropType, defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import { VendorEnum } from '@/common/constant';
import tcloudVendor from '@/assets/image/vendor-tcloud.png';
import awsVendor from '@/assets/image/vendor-aws.png';
import azureVendor from '@/assets/image/vendor-azure.png';
import gcpVendor from '@/assets/image/vendor-gcp.png';
import huaweiVendor from '@/assets/image/vendor-huawei.png';
import { Success, InfoLine, TextFile } from 'bkui-vue/lib/icon';
import http from '@/http';
import successIcon from '@/assets/image/success.png';
import failedIcon from '@/assets/image/failure.png';
import MemberSelect from '@/components/MemberSelect';

const { FormItem } = Form;
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

interface IExtensionItem {
  label: string,
  value: string,
}
enum ValidateStatus {
  YES,
  NO,
  UNKOWN,
};
interface IExtension {
  input: Record<string, IExtensionItem>,    // 输入
  output1: Record<string, IExtensionItem>, // 需要显眼的输出
  output2: Record<string, IExtensionItem>, // 不需要显眼的输出
  validatedStatus: ValidateStatus,        // 是否校验通过
  validateFailedReason?: string;          // 不通过的理由
}

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
  },
  setup(props) {
    const formModel = reactive({
      site: 'china' as 'china'|'international', // 站点
      vendor: VendorEnum.TCLOUD, // 云厂商
      name: '', // 账号别名
      managers: [] as Array<string>, // 责任人
      type: 'resource', // 账号类型，当前产品形态固定为 resource，资源账号
      memo: '', // 备注
      extension: {}, // 不同云的secretKey\id
    });
    const apiFormInstance = ref(null);
    const isValidateLoading = ref(false);
    // 腾讯云
    const tcloudExtension: IExtension = reactive({
      output1: {
        cloud_main_account_id: {
          value: '',
          label: '云主账户ID',
        },
      },
      output2: {
        cloud_sub_account_id: {
          value: '',
          label: '云子账户ID',
        },
      },
      input: {
        cloud_secret_id: {
          value: '',
          label: '云加密ID',
        },
        cloud_secret_key: {
          value: '',
          label: '云密钥',
        },
      },
      validatedStatus: ValidateStatus.UNKOWN,
    });
    // 亚马逊云
    const awsExtension: IExtension = reactive({
      output1: {
        cloud_account_id: {
          value: '',
          label: '云账号ID',
        },
      },
      output2: {
        cloud_iam_username: {
          value: '',
          label: '云IAM用户名',
        },
      },
      input: {
        cloud_secret_id: {
          value: '',
          label: '云加密ID',
        },
        cloud_secret_key: {
          value: '',
          label: '云密钥',
        },
      },
      validatedStatus: ValidateStatus.UNKOWN,
    });
    // 华为云
    const huaweiExtension: IExtension = reactive({
      output1: {
        cloud_sub_account_id: {
          value: '',
          label: '云子账户ID',
        },
      },
      output2: {
        cloud_sub_account_name: {
          value: '',
          label: '云子账户名称',
        },
        cloud_iam_user_id: {
          value: '',
          label: '云加密ID',
        },
        cloud_iam_username: {
          value: '',
          label: '云IAM用户名称',
        },
      },
      input: {
        cloud_secret_id: {
          value: '',
          label: '云加密ID',
        },
        cloud_secret_key: {
          value: '',
          label: '云密钥',
        },
      },
      validatedStatus: ValidateStatus.UNKOWN,
    });
    // 谷歌云
    const gcpExtension: IExtension = reactive({
      output1: {
        cloud_project_id: {
          label: '云项目ID',
          value: '',
        },
        cloud_project_name: {
          label: '云项目名称',
          value: '',
        },
      },
      output2: {
        cloud_service_account_id: {
          label: '云服务账户ID',
          value: '',
        },
        cloud_service_account_name: {
          label: '云服务账户名称',
          value: '',
        },
        cloud_service_secret_id: {
          label: '云服务密钥ID',
          value: '',
        },
      },
      input: {
        cloud_service_secret_key: {
          label: '云服务密钥',
          value: '',
        },
      },
      validatedStatus: ValidateStatus.UNKOWN,
    });
    // 微软云
    const azureExtension: IExtension = reactive({
      output1: {
        cloud_subscription_id: {
          value: '',
          label: '云订阅ID',
        },
      },
      output2: {
        cloud_subscription_name: {
          label: '云订阅名称',
          value: '',
        },
        cloud_application_name: {
          label: '云应用名称',
          value: '',
        },
      },
      input: {
        cloud_tenant_id: {
          value: '',
          label: '云租户ID',
        },
        cloud_application_id: {
          value: '',
          label: '云应用ID',
        },
        cloud_client_secret_key: {
          value: '',
          label: '云客户端密钥',
        },
      },
      validatedStatus: ValidateStatus.UNKOWN,
    });
    // 当前选中的云厂商对应的 extension
    const curExtension = ref<IExtension>(tcloudExtension);

    watch(
      () => formModel.vendor,
      (vendor) => {
        switch (vendor) {
          case VendorEnum.TCLOUD: {
            curExtension.value = tcloudExtension;
            break;
          }
          case VendorEnum.AWS: {
            curExtension.value = awsExtension;
            break;
          }
          case VendorEnum.HUAWEI: {
            curExtension.value = huaweiExtension;
            break;
          }
          case VendorEnum.GCP: {
            curExtension.value = gcpExtension;
            break;
          }
          case VendorEnum.AZURE: {
            curExtension.value = azureExtension;
            break;
          }
        };
      },
    );

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
          console.log(666, formModel);
        }
      },
      {
        deep: true,
      },
    );

    watch(
      () => formModel,
      (model) => {
        props.changeSubmitData(model);
      },
      {
        deep: true,
      },
    );

    const handleValidate = async () => {
      isValidateLoading.value = true;
      const payload = Object.entries(curExtension.value.input).reduce((prev, [key, { value }]) => {
        prev[key] = value;
        return prev;
      }, {});
      try {
        const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${formModel.vendor}/accounts/secret`, payload);
        if (res.data) {
          Object.entries(res.data).forEach(([key, val]) => {
            if (curExtension.value.output1[key]) curExtension.value.output1[key].value = val as string;
            if (curExtension.value.output2[key]) curExtension.value.output2[key].value = val as string;
          });
        }
        curExtension.value.validatedStatus = ValidateStatus.YES;
      } catch (err: any) {
        curExtension.value.validateFailedReason = err.message;
        curExtension.value.validatedStatus = ValidateStatus.NO;
      } finally {
        isValidateLoading.value = false;
      }
    };

    return () => (
      <div class={'account-form'}>
        <Card
          class={'account-form-card'}
          showHeader={false}
        >
          <p class={'account-form-card-title'}>
            账号归属
          </p>
          <div class={'account-form-card-content'}>
            <Form
              formType='vertical'
            >
              <FormItem
                label='厂商选择'
                required
              >
                <div class={'account-vendor-selector'}>
                  {
                    VENDORS_INFO.map(({ vendor, name, icon }) => (
                      <div
                        class={`account-vendor-option ${vendor === formModel.vendor ? 'account-vendor-option-active' : ''}`}
                        onClick={() => formModel.vendor = vendor}
                        >
                        <img src={icon} alt={name} class={'account-vendor-option-icon'}/>
                        <p class={'account-vendor-option-text'}>
                        { name }
                        </p>
                        {
                          formModel.vendor === vendor
                            ? <Success fill='#3A84FF' class={'active-icon'}/>
                            : null
                        }
                      </div>
                    ))
                  }
                </div>
              </FormItem>
              <FormItem
                required
                label='站点种类'
              >
                <Radio
                  label={'china'}
                  v-model={formModel.site}
                >
                  中国站
                </Radio>
                <Radio
                  label={'international'}
                  v-model={formModel.site}
                >
                  国际站
                </Radio>
              </FormItem>
            </Form>
          </div>
        </Card>

        <Card
          class={'account-form-card'}
          showHeader={false}
        >
          <>
                <div class={'api-secret-header'}>
                  <p class={'account-form-card-title'}>API 密钥</p>
                  <InfoLine fill='#979BA5' />
                  <p class={'header-text'}>
                    同一个主账号下,只允许接入一次。如后续对API密钥更新,必须是隶属于同一主账号D。
                  </p>
                  <TextFile fill='#3A84FF' />
                  <Button theme='primary' text class={'header-btn'}>
                    接入指引
                  </Button>
                </div>
                <div class={'account-form-card-content'}>
                  <Form
                    ref={apiFormInstance}
                    formType='vertical'
                    class={'account-form-card-content-grid'}>
                    <div>
                      {
                        Object.entries(curExtension.value.input).map(([property, { label }]) => (
                          <FormItem label={label} property={property} required>
                            <Input v-model={curExtension.value.input[property].value} type={
                             property === 'cloud_service_secret_key' && formModel.vendor === VendorEnum.GCP
                               ? 'textarea'
                               : 'text'
                            }
                            rows={8}
                            />
                          </FormItem>
                        ))
                      }
                    </div>
                    <div>
                      {Object.entries(curExtension.value.output1).map(([property, { label, value }]) => (
                          <FormItem label={label} property={property} required>
                            <Input v-model={value} disabled placeholder={' '}/>
                          </FormItem>
                      ))}
                    </div>
                  </Form>
                </div>
                <div class={'validate-btn-block'}>
                  <Button theme='primary' class={'account-validate-btn'} onClick={handleValidate} loading={isValidateLoading.value}>
                    账号校验
                  </Button>
                  {
                    curExtension.value.validatedStatus === ValidateStatus.YES
                      ? (
                      <>
                        <img src={successIcon} alt="success" class={'validate-icon'}></img>
                        <span> 校验成功 </span>
                      </>
                      )
                      : null
                  }
                  {
                    curExtension.value.validatedStatus === ValidateStatus.NO
                      ? (
                      <>
                        <img src={failedIcon} alt="success" class={'validate-icon'}></img>
                        <span> 校验失败 {curExtension.value.validateFailedReason}</span>
                      </>
                      )
                      : null
                  }
                </div>
              </>
        </Card>

        <Card
          class={'account-form-card'}
          showHeader={false}
        >
          <p class={'account-form-card-title'}>
            其他信息
          </p>
          <div class={'account-form-card-content'}>
            <Form
              formType='vertical'
              model={formModel}
              auto-check
            >
              {/* eslint-disable-next-line @typescript-eslint/no-unused-vars */}
              {Object.entries(curExtension.value.output2).map(([property, { label, value }]) => (
                  <FormItem label={label} required>
                    <Input v-model={value} disabled placeholder={' '}/>
                  </FormItem>
              ))}
              <FormItem label='账号别名' class={'api-secret-selector'} required property='name'>
                <Input v-model={formModel.name}/>
              </FormItem>
              <FormItem label='责任人' class={'api-secret-selector'} required property='managers'>
                <MemberSelect v-model={formModel.managers}/>
              </FormItem>
              <FormItem label='备注'>
                <Input type={'textarea'} v-model={formModel.memo}/>
              </FormItem>
            </Form>
          </div>
        </Card>
      </div>
    );
  },
});
