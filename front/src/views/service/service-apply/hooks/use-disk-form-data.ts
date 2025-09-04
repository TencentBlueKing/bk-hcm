import http from '@/http';
import { reactive, watch, ref, nextTick } from 'vue';
import { VendorEnum } from '@/common/constant';
import type { Cond } from './use-condtion';
import { Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useWhereAmI } from '@/hooks/useWhereAmI';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export interface IPurchaseDuration {
  count: number;
  unit: 'm' | 'y';
}

export interface IChargePrepaidTcloud {
  period: number;
  renew_flag: 'NOTIFY_AND_AUTO_RENEW' | 'NOTIFY_AND_MANUAL_RENEW';
}
export interface IChargePrepaidHuawei {
  period_num: number;
  period_type: 'month' | 'year';
  is_auto_renew: 'ture' | 'false';
}

export interface IDiskBaseData {
  disk_name: string;
  zone: string;
  disk_type: string;
  disk_size: number;
  disk_count: number;
  disk_charge_type?: 'PREPAID' | 'POSTPAID_BY_HOUR' | 'prePaid' | 'postPaid';
  disk_charge_prepaid?: IChargePrepaidTcloud | IChargePrepaidHuawei;
  memo: string;
}

export interface IDiskFormData extends IDiskBaseData {
  purchase_duration?: IPurchaseDuration;
  auto_renew?: boolean;
}
export interface IDiskSaveData extends IDiskBaseData {
  bk_biz_id: number;
  account_id: string;
  region: string;
  vendor?: string;
  resource_group_name?: string;
}

export default (cond: Cond) => {
  const { t } = useI18n();
  const router = useRouter();

  const vendorDiffFormData = (vendor: string) => {
    const diff = {
      [VendorEnum.TCLOUD]: {
        disk_charge_type: 'PREPAID',
        purchase_duration: {
          count: 1,
          unit: 'm',
        },
        auto_renew: true,
      },
      [VendorEnum.AWS]: {},
      [VendorEnum.AZURE]: {},
      [VendorEnum.GCP]: {},
      [VendorEnum.HUAWEI]: {
        disk_charge_type: 'prePaid',
        purchase_duration: {
          count: 1,
          unit: 'm',
        },
        auto_renew: true,
      },
    };
    return diff[vendor] || {};
  };
  const defaultFormData = (vendor: string) => {
    const base: IDiskFormData = {
      disk_name: '',
      zone: '',
      disk_type: '',
      disk_size: null,
      disk_count: 1,
      memo: '',
    };

    return {
      ...base,
      ...vendorDiffFormData(vendor),
    };
  };

  const formData = reactive<IDiskFormData>(defaultFormData(cond.vendor));
  const formRef = ref(null);

  const resetFormData = () => {
    const keys = [
      'zone',
      'disk_name',
      'disk_type',
      'disk_charge_type',
      'disk_charge_prepaid',
      'purchase_duration',
      'auto_renew',
    ];
    keys.forEach((key) => resetFormItemData(key));
  };

  const resetFormItemData = (key: string) => {
    const defaultData: IDiskFormData = defaultFormData(cond.vendor);
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
    const { purchase_duration, auto_renew, ...saveFormData } = formData;
    const saveData: IDiskSaveData = {
      ...saveFormData,
      bk_biz_id: cond.bizId,
      account_id: cond.cloudAccountId,
      region: cond.region,
      zone: formData?.zone?.[0],
    };

    if (cond.vendor === VendorEnum.TCLOUD) {
      saveData.disk_charge_prepaid =
        saveFormData.disk_charge_type === 'PREPAID'
          ? {
              period: purchase_duration.count * (purchase_duration.unit === 'y' ? 12 : 1),
              renew_flag: auto_renew ? 'NOTIFY_AND_AUTO_RENEW' : 'NOTIFY_AND_MANUAL_RENEW',
            }
          : undefined;
    }

    if (cond.vendor === VendorEnum.HUAWEI) {
      saveData.disk_charge_prepaid =
        saveFormData.disk_charge_type === 'prePaid'
          ? {
              period_num: purchase_duration.count,
              period_type: purchase_duration.unit === 'y' ? 'year' : 'month',
              is_auto_renew: auto_renew ? 'ture' : 'false',
            }
          : undefined;
    }

    if (cond.vendor === VendorEnum.AZURE) {
      saveData.resource_group_name = cond.resourceGroup;
    }

    return saveData;
  };

  const { isResourcePage } = useWhereAmI();
  const submitting = ref(false);
  const handleFormSubmit = async () => {
    await formRef.value.validate();
    const saveData = getSaveData();
    // console.log(saveData, '-----saveData');
    try {
      submitting.value = true;
      const url = isResourcePage
        ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/disks/create`
        : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${cond.vendor}/applications/types/create_disk`;
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
