import http from '@/http';
import { reactive, watch, ref, nextTick } from 'vue';
import { VendorEnum } from '@/common/constant';
import type { Cond } from './use-condtion';
import { Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export interface ISubnet {
  name: string;
  ipv4_cidr: string;
  private_ip_google_access?: boolean;
  zone?: string;
  enable_flow_logs?: boolean;
  ipv6_enable?: boolean;
}

export interface IVpcBaseData {
  name: string;
  ipv4_cidr?: string | number[];
  bk_cloud_id: number;
  subnet?: ISubnet;
  routing_mode?: 'REGIONAL' | 'GLOBAL';
  instance_tenancy?: 'default' | 'dedicated';
  memo: string;
}
export interface IVpcFormData extends IVpcBaseData {
  ipv4_cidr: string;
  type?: number;
  ip_source_type?: number;
  bastion_host_enable?: boolean;
  ddos_enable?: boolean;
  firewall_enable?: boolean;
}
export interface IVpcSaveData extends IVpcBaseData {
  bk_biz_id: number;
  account_id: string;
  region: string;
  vendor?: string;
  resource_group_name?: string;
}

export default (cond: Cond) => {
  const { t } = useI18n();
  const router = useRouter();
  const { isResourcePage, whereAmI } = useWhereAmI();

  const vendorDiffFormData = (vendor: string) => {
    const diff = {
      [VendorEnum.TCLOUD]: {
        ip_source_type: 0,
        subnet: {
          name: '',
          ipv4_cidr: '10.0.0.0/8',
          zone: '',
        },
      },
      [VendorEnum.AWS]: {
        type: 0,
        instance_tenancy: 'default',
      },
      [VendorEnum.AZURE]: {
        ip_source_type: 0,
        bastion_host_enable: false,
        ddos_enable: false,
        firewall_enable: false,
        subnet: {
          name: '',
          ipv4_cidr: '10.0.0.0/8',
        },
      },
      [VendorEnum.GCP]: {
        routing_mode: 'REGIONAL',
        subnet: {
          name: '',
          ipv4_cidr: '10.0.0.0/8',
          private_ip_google_access: false,
          enable_flow_logs: false,
        },
      },
      [VendorEnum.HUAWEI]: {
        ip_source_type: 0,
        subnet: {
          name: '',
          gateway_ip: '',
          ipv4_cidr: '10.0.0.0/8',
          ipv6_enable: false,
        },
      },
    };
    return diff[vendor] || {};
  };
  const defaultFormData = (vendor: string) => {
    const base: IVpcFormData = {
      name: '',
      ipv4_cidr: '10.0.0.0/8',
      bk_cloud_id: null,
      subnet: {
        name: '',
        ipv4_cidr: '10.0.0.0/8',
      },
      memo: '',
    };

    return {
      ...base,
      ...vendorDiffFormData(vendor),
    };
  };

  const formData = reactive<IVpcFormData>(defaultFormData(cond.vendor));
  const formRef = ref(null);

  const resetFormData = () => {
    const keys = [
      'ipv4_cidr',
      'subnet',
      'type',
      'instance_tenancy',
      'ip_source_type',
      'bastion_host_enable',
      'ddos_enable',
      'firewall_enable',
      'routing_mode',
    ];
    keys.forEach((key) => resetFormItemData(key));
  };

  const resetFormItemData = (key: string) => {
    const defaultData: IVpcFormData = defaultFormData(cond.vendor);
    formData[key] = defaultData[key];
  };

  watch(cond, (val) => {
    resetFormData();
    Object.assign(formData, val);

    nextTick(() => {
      formRef.value.clearValidate();
    });
  });

  const getSaveData = () => {
    // console.log(formData, '---formData');
    const {
      type,
      subnet,
      ipv4_cidr,
      ip_source_type,
      instance_tenancy,
      bastion_host_enable,
      ddos_enable,
      firewall_enable,
      routing_mode,
      ...saveFormData
    } = formData;
    const saveData: IVpcSaveData = {
      ...saveFormData,
      bk_biz_id: cond.bizId,
      account_id: cond.cloudAccountId,
      region: cond.region,
    };
    saveData.ipv4_cidr = formData.ipv4_cidr;
    if (cond.vendor === VendorEnum.TCLOUD) {
      saveData.subnet = {
        name: subnet.name,
        ipv4_cidr: formData.ipv4_cidr,
        zone: subnet.zone?.[0],
      };
    }

    if (cond.vendor === VendorEnum.AWS) {
      saveData.instance_tenancy = instance_tenancy;
    }

    if (cond.vendor === VendorEnum.HUAWEI) {
      saveData.subnet = {
        name: subnet.name,
        gateway_ip: `${subnet.ipv4_cidr.split('.').slice(0, 3).join('.')}.1`,
        ipv4_cidr: formData.ipv4_cidr,
        ipv6_enable: subnet.ipv6_enable,
      };
    }

    if (cond.vendor === VendorEnum.AZURE) {
      saveData.resource_group_name = cond.resourceGroup;
      saveData.subnet = {
        name: subnet.name,
        ipv4_cidr: formData.ipv4_cidr,
      };
    }

    if (cond.vendor === VendorEnum.GCP) {
      saveData.routing_mode = routing_mode;
      saveData.subnet = {
        name: subnet.name,
        ipv4_cidr: formData.ipv4_cidr,
        private_ip_google_access: subnet.private_ip_google_access,
        enable_flow_logs: subnet.enable_flow_logs,
      };
    }

    if (whereAmI.value === Senarios.resource) {
      delete saveData.bk_biz_id;
      delete saveData.bizId;
    }
    return saveData;
  };

  const submitting = ref(false);
  const handleFormSubmit = async () => {
    await formRef.value.validate();
    const saveData = getSaveData();
    try {
      submitting.value = true;
      const url = isResourcePage
        ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vpcs/create`
        : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${cond.vendor}/applications/types/create_vpc`;
      await http.post(url, saveData);

      Message({
        theme: 'success',
        message: t('提交成功'),
      });

      if (isResourcePage) router.back();
      else {
        router.push({
          path: '/service/my-apply',
        });
      }
    } catch (err) {
      // console.error(err);
    } finally {
      submitting.value = false;
    }
  };

  return {
    formRef,
    formData,
    getSaveData,
    submitting,
    handleFormSubmit,
  };
};
