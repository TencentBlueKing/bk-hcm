import { VendorEnum } from '@/common/constant';
import http from '@/http';
import { reactive, ref, watch } from 'vue';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export type IProp = {
  vendor: VendorEnum;
  [key: string]: any;
};
export interface IExtensionItem {
  label: string;
  value: string;
  placeholder?: string;
}
export enum ValidateStatus {
  YES,
  NO,
  UNKOWN,
}
export interface IExtension {
  input: Record<string, IExtensionItem>; // 输入
  output1: Record<string, IExtensionItem>; // 需要显眼的输出
  output2: Record<string, IExtensionItem>; // 不需要显眼的输出
  validatedStatus: ValidateStatus; // 是否校验通过
  validateFailedReason?: string; // 不通过的理由
  formType?: string;
  selectType?: any[]; // 是选择类型的字段
  selectParams?: any;
}
export const useSecretExtension = (props: IProp, isValidate?: boolean) => {
  const { entry = 'cloud' } = props;
  // 腾讯云
  const tcloudExtension: IExtension = reactive({
    output1: {
      cloud_sub_account_id: {
        value: '',
        label: '云子账户ID',
        placeholder: '密钥校验成功后自动填充',
      },
    },
    output2: {
      cloud_main_account_id: {
        value: '',
        label: '云主账户ID',
        placeholder: '密钥校验成功后自动填充',
      },
    },
    input: {
      cloud_secret_id: {
        value: '',
        label: '云密钥ID',
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
        placeholder: '密钥校验成功后自动填充',
      },
    },
    output2: {
      cloud_iam_username: {
        value: '',
        label: '云IAM用户名',
        placeholder: '密钥校验成功后自动填充',
      },
    },
    input: {
      cloud_secret_id: {
        value: '',
        label: '云密钥ID',
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
        placeholder: '密钥校验成功后自动填充',
      },
    },
    output2: {
      cloud_sub_account_name: {
        value: '',
        label: '云子账户名称',
        placeholder: '密钥校验成功后自动填充',
      },
      cloud_iam_user_id: {
        value: '',
        label: '云密钥ID',
        placeholder: '密钥校验成功后自动填充',
      },
      cloud_iam_username: {
        value: '',
        label: '云IAM用户名称',
        placeholder: '密钥校验成功后自动填充',
      },
    },
    input: {
      cloud_secret_id: {
        value: '',
        label: '云密钥ID',
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
    formType: entry === 'cloud' ? 'input' : 'select',
    selectType: ['cloud_project_id'],
    selectParams: {
      cloud_project_id: {
        idKey: 'cloud_project_id',
        displayKey: 'cloud_project_id',
        list: [],
        follow: 'cloud_project_name',
      },
    },
    output1: {
      cloud_project_id: {
        label: '云项目ID',
        value: '',
        placeholder: '密钥校验成功后自动填充',
      },
      cloud_project_name: {
        label: '云项目名称',
        value: '',
        placeholder: '密钥校验成功后自动填充',
      },
    },
    output2: {
      cloud_service_account_id: {
        label: '云服务账户ID',
        value: '',
        placeholder: '密钥校验成功后自动填充',
      },
      cloud_service_account_name: {
        label: '云服务账户名称',
        value: '',
        placeholder: '密钥校验成功后自动填充',
      },
      cloud_service_secret_id: {
        label: '云服务密钥ID',
        value: '',
        placeholder: '密钥校验成功后自动填充',
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
        placeholder: '密钥校验成功后自动填充',
      },
    },
    output2: {
      cloud_subscription_name: {
        label: '云订阅名称',
        value: '',
        placeholder: '密钥校验成功后自动填充',
      },
      cloud_application_name: {
        label: '云应用名称',
        value: '',
        placeholder: '密钥校验成功后自动填充',
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
  // 腾讯云
  const zenlayerOrKaopuExtension: IExtension = reactive({
    output1: {},
    output2: {},
    input: {
      cloud_account_id: {
        value: '',
        label: '云账号ID',
      },
    },
    // 无需校验
    validatedStatus: ValidateStatus.YES,
  });
  // 当前选中的云厂商对应的 extension
  const curExtension = ref<IExtension>(tcloudExtension);
  const isValidateLoading = ref(false);
  const isValidateDiasbled = ref(true);
  // 接口需要的 payload
  const extensionPayload = ref({});

  watch(
    () => props.vendor,
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
        case VendorEnum.ZENLAYER:
        case VendorEnum.KAOPU:
          curExtension.value = zenlayerOrKaopuExtension;
          break;
      }
    },
    {
      immediate: true,
    },
  );

  watch(
    () => curExtension.value,
    () => {
      isValidateDiasbled.value = Object.entries(curExtension.value.input).reduce(
        (prev, [_key, { value }]) => prev || !value,
        false,
      );
      extensionPayload.value = [
        ...Object.entries(curExtension.value.input),
        ...Object.entries(curExtension.value.output1),
        ...Object.entries(curExtension.value.output2),
      ].reduce((prev, [key, { value }]) => {
        prev[key] = value;
        return prev;
      }, {});
    },
    {
      deep: true,
    },
  );

  const handleValidate = async (callback: Function = undefined, type: String = 'accounts') => {
    isValidateLoading.value = true;
    const payload = extensionPayload.value;
    // props.changeExtension(payload);
    if (callback) callback?.(payload);
    try {
      const server = type === 'accounts' ? 'cloud' : 'account';
      const res = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/${server}/vendors/${props.vendor}/${type}/secret`, {
        ...payload,
        disable_check: isValidate || false,
      });

      const dataIsArray = res.data instanceof Array;
      const data = dataIsArray ? res.data[0] : res.data;
      if (data) {
        if (dataIsArray) {
          const { selectType = [], selectParams = {} } = curExtension.value;
          selectType.forEach((key) => {
            selectParams[key].list = res.data;
          });
        }
        Object.entries(data).forEach(([key, val]) => {
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

  return {
    curExtension,
    tcloudExtension,
    awsExtension,
    azureExtension,
    gcpExtension,
    huaweiExtension,
    handleValidate,
    isValidateLoading,
    isValidateDiasbled,
    extensionPayload,
  };
};
