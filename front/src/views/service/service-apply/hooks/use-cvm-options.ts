/* eslint-disable no-useless-escape */
import { ref, watch } from 'vue';
import { VendorEnum } from '@/common/constant';
import type { Cond } from './use-condtion';
import type { ICvmFormData } from './use-cvm-form-data';

export default (cond: Cond, formData: ICvmFormData) => {
  const sysDiskTypes = ref([]);

  const dataDiskTypes = ref([]);

  const billingModes = ref([]);

  const internetChargeTypes = ref([]);

  const purchaseDurationUnits = [
    {
      id: 'm',
      name: '月',
    },
    {
      id: 'y',
      name: '年',
    },
  ];

  const sysDiskTypeValues = {
    [VendorEnum.TCLOUD]: [
      {
        id: 'CLOUD_PREMIUM',
        name: '高性能云硬盘',
      },
      {
        id: 'CLOUD_BSSD',
        name: '通用型SSD云硬盘',
      },
      {
        id: 'CLOUD_SSD',
        name: 'SSD云硬盘',
      },
    ],
    [VendorEnum.AWS]: [
      {
        id: 'gp3',
        name: '通用型SSD卷(gp3)',
      },
      {
        id: 'gp2',
        name: '通用型SSD卷(gp2)',
      },
      {
        id: 'io1',
        name: '预置IOPS SSD卷(io1)',
      },
      {
        id: 'io2',
        name: '预置IOPS SSD卷(io2)',
      },
      {
        id: 'st1',
        name: '吞吐量优化型HDD卷(st1)',
      },
      {
        id: 'sc1',
        name: 'Cold HDD卷(sc1)',
      },
      {
        id: 'standard',
        name: '上一代磁介质卷(standard)',
      },
    ],
    [VendorEnum.AZURE]: [
      {
        id: 'Premium_LRS',
        name: '高级SSD',
      },
      {
        id: 'StandardSSD_LRS',
        name: '标准SSD',
      },
      {
        id: 'Standard_LRS',
        name: '标准HDD',
      },
    ],
    [VendorEnum.GCP]: [
      {
        id: 'pd-standard',
        name: '标准永久性磁盘',
      },
      {
        id: 'pd-balanced',
        name: '均衡永久性磁盘',
      },
      {
        id: 'pd-ssd',
        name: '性能(SSD)永久性磁盘',
      },
      {
        id: 'pd-extreme',
        name: '极端永久性磁盘',
      },
    ],
    [VendorEnum.HUAWEI]: [
      {
        id: 'SATA',
        name: '普通IO云硬盘',
      },
      {
        id: 'SAS',
        name: '高IO云硬盘',
      },
      {
        id: 'GPSSD',
        name: '通用型SSD云硬盘',
      },
      {
        id: 'SSD',
        name: '超高IO云硬盘',
      },
      {
        id: 'ESSD',
        name: '极速IO云硬盘',
      },
    ],
  };

  const dataDiskTypeValues = {
    [VendorEnum.TCLOUD]: [
      {
        id: 'CLOUD_PREMIUM',
        name: '高性能云硬盘',
      },
      {
        id: 'CLOUD_BSSD',
        name: '通用型SSD云硬盘',
      },
      {
        id: 'CLOUD_SSD',
        name: 'SSD云硬盘',
      },
      {
        id: 'CLOUD_HSSD',
        name: '增强型SSD云硬盘',
      },
    ],
    [VendorEnum.AWS]: [
      {
        id: 'gp3',
        name: '通用型SSD卷(gp3)',
      },
      {
        id: 'gp2',
        name: '通用型SSD卷(gp2)',
      },
      {
        id: 'io1',
        name: '预置IOPS SSD卷(io1)',
      },
      {
        id: 'io2',
        name: '预置IOPS SSD卷(io2)',
      },
      {
        id: 'st1',
        name: '吞吐量优化型HDD卷(st1)',
      },
      {
        id: 'sc1',
        name: 'Cold HDD卷(sc1)',
      },
      {
        id: 'standard',
        name: '上一代磁介质卷(standard)',
      },
    ],
    [VendorEnum.AZURE]: [
      {
        id: 'Premium_LRS',
        name: '高级SSD',
      },
      {
        id: 'StandardSSD_LRS',
        name: '标准SSD',
      },
      {
        id: 'Standard_LRS',
        name: '标准HDD',
      },
    ],
    [VendorEnum.GCP]: [
      {
        id: 'pd-standard',
        name: '标准永久性磁盘',
      },
      {
        id: 'pd-balanced',
        name: '均衡永久性磁盘',
      },
      {
        id: 'pd-ssd',
        name: '性能(SSD)永久性磁盘',
      },
      {
        id: 'pd-extreme',
        name: '极端永久性磁盘',
      },
    ],
    [VendorEnum.HUAWEI]: [
      {
        id: 'SATA',
        name: '普通IO云硬盘',
      },
      {
        id: 'SAS',
        name: '高IO云硬盘',
      },
      {
        id: 'GPSSD',
        name: '通用型SSD云硬盘',
      },
      {
        id: 'SSD',
        name: '超高IO云硬盘',
      },
      {
        id: 'ESSD',
        name: '极速IO云硬盘',
      },
    ],
  };

  const billingModeValues = {
    [VendorEnum.TCLOUD]: [
      {
        id: 'PREPAID',
        name: '包年包月',
      },
      {
        id: 'POSTPAID_BY_HOUR',
        name: '按量计费',
      },
      {
        id: 'SPOTPAID',
        name: '竞价实例',
      },
    ],
    [VendorEnum.HUAWEI]: [
      {
        id: 'prePaid',
        name: '包年包月',
      },
      {
        id: 'postPaid',
        name: '按量计费',
      },
    ],
  };

  const internetChargeTypeValues = {
    [VendorEnum.TCLOUD]: [
      {
        id: 'BANDWIDTH_PREPAID',
        name: '包月带宽',
      },
      {
        id: 'BANDWIDTH_POSTPAID_BY_HOUR',
        name: '按小时带宽',
      },
      {
        id: 'TRAFFIC_POSTPAID_BY_HOUR',
        name: '按流量计费',
      },
      {
        id: 'BANDWIDTH_PACKAGE',
        name: '共享带宽包',
      },
    ],
  };

  watch(
    () => cond.vendor,
    (vendor) => {
      sysDiskTypes.value = sysDiskTypeValues[vendor] || [];

      dataDiskTypes.value = dataDiskTypeValues[vendor] || [];

      billingModes.value = billingModeValues[vendor] || [];

      internetChargeTypes.value = internetChargeTypeValues[vendor] || [];
    },
    {
      immediate: true,
    },
  );

  // watch(() => cond.bizId, () => {
  //   sysDiskTypes.value = [];
  //   dataDiskTypes.value = [];
  // });

  watch(
    () => formData.instance_type,
    (type) => {
      if (cond.vendor === VendorEnum.AZURE) {
        const disablePremiumTypes = [
          'Standard_A1_v2',
          'Standard_A2_v2',
          'Standard_A4_v2',
          'Standard_A8_v2',
          'Standard_A2m_v2',
          'Standard_A4m_v2',
          'Standard_A8m_v2',
          'Standard_D1_v2',
          'Standard_D2_v2',
          'Standard_D3_v2',
          'Standard_D4_v2',
          'Standard_D5_v2',
          'Standard_D2_v3',
          'Standard_D4_v3',
          'Standard_D8_v3',
          'Standard_D16_v3',
          'Standard_D32_v3',
          'Standard_D48_v3',
          'Standard_D64_v3',
          'Standard_D2_v4',
          'Standard_D4_v4',
          'Standard_D8_v4',
          'Standard_D16_v4',
          'Standard_D32_v4',
          'Standard_D48_v4',
          'Standard_D64_v4',
          'Standard_D2a_v4',
          'Standard_D4a_v4',
          'Standard_D8a_v4',
          'Standard_D16a_v4',
          'Standard_D32a_v4',
          'Standard_D48a_v4',
          'Standard_D64a_v4',
          'Standard_D96a_v4',
          'Standard_D2d_v4',
          'Standard_D4d_v4',
          'Standard_D8d_v4',
          'Standard_D16d_v4',
          'Standard_D32d_v4',
          'Standard_D64d_v4',
          'Standard_D2_v5',
          'Standard_D4_v5',
          'Standard_D8_v5',
          'Standard_D16_v5',
          'Standard_D32_v5',
          'Standard_D48_v5',
          'Standard_D64_v5',
          'Standard_D96_v5',
          'Standard_D2d_v5',
          'Standard_D4d_v5',
          'Standard_D8d_v5',
          'Standard_D16d_v5',
          'Standard_D32d_v5',
          'Standard_D48_v5',
          'Standard_D64_v5',
          'Standard_D96_v5',
          'Standard_D11_v2',
          'Standard_D12_v2',
          'Standard_D13_v2',
          'Standard_D14_v2',
          'Standard_D15_v2',
          'Standard_E2_v3',
          'Standard_E4_v3',
          'Standard_E8_v3',
          'Standard_E16_v3',
          'Standard_E20_v3',
          'Standard_E32_v3',
          'Standard_E48_v3',
          'Standard_E64_v3',
          'Standard_E64i_v3',
          'Standard_E2a_v4',
          'Standard_E4a_v4',
          'Standard_E8a_v4',
          'Standard_E16a_v4',
          'Standard_E20a_v4',
          'Standard_E32a_v4',
          'Standard_E48a_v4',
          'Standard_E64a_v4',
          'Standard_E96a_v4',
          'Standard_E2d_v4',
          'Standard_E4d_v4',
          'Standard_E8d_v4',
          'Standard_E16d_v4',
          'Standard_E20d_v4',
          'Standard_E32d_v4',
          'Standard_E48d_v4',
          'Standard_E64d_v4',
          'Standard_E2_v4',
          'Standard_E4_v4',
          'Standard_E8_v4',
          'Standard_E16_v4',
          'Standard_E20_v4',
          'Standard_E32_v4',
          'Standard_E48_v4',
          'Standard_E64_v4',
          'Standard_E2_v5',
          'Standard_E4_v5',
          'Standard_E8_v5',
          'Standard_E16_v5',
          'Standard_E20_v5',
          'Standard_E32_v5',
          'Standard_E48_v5',
          'Standard_E64_v5',
          'Standard_E96_v5',
          'Standard_E104i_v5',
          'Standard_E2d_v5',
          'Standard_E4d_v5',
          'Standard_E8d_v5',
          'Standard_E16d_v5',
          'Standard_E20d_v5',
          'Standard_E32d_v5',
          'Standard_E48d_v5',
          'Standard_E64d_v5',
          'Standard_E96d_v5',
          'Standard_E104id_v5',
          'Standard_NC6',
          'Standard_NC12',
          'Standard_NC24',
          'Standard_NC24r',
          'Standard_NV12',
          'Standard_NV6',
          'Standard_NV24',
          'Standard_F1',
          'Standard_F2',
          'Standard_F4',
          'Standard_F8',
          'Standard_F16',
          'A0Basic_A0',
          'A1Basic_A1',
          'A2Basic_A2',
          'A3Basic_A3',
          'A4Basic_A4',
          'Standard_A0',
          'Standard_A1',
          'Standard_A2',
          'Standard_A3',
          'Standard_A4',
          'Standard_A5',
          'Standard_A6',
          'Standard_A7',
          'Standard_A8',
          'Standard_A9',
          'Standard_A10',
          'Standard_A11',
          'Standard_D1',
          'Standard_D2',
          'Standard_D3',
          'Standard_D4',
          'Standard_D11',
          'Standard_D12',
          'Standard_D13',
          'Standard_D14',
          'Standard_G1',
          'Standard_G2',
          'Standard_G3',
          'Standard_G4',
          'Standard_G5',
          'Standard_NV6',
          'Standard_NV12',
          'Standard_NV24',
          'Standard_NC6',
          'Standard_NC12',
          'Standard_NC24',
          'Standard_NC24r',
        ];

        sysDiskTypes.value = (sysDiskTypeValues[cond.vendor] || []).slice();
        dataDiskTypes.value = (dataDiskTypeValues[cond.vendor] || []).slice();

        if (disablePremiumTypes.includes(type)) {
          const sysIndex = sysDiskTypes.value.findIndex((item) => item.id === 'Premium_LRS');
          const dataIndex = dataDiskTypes.value.findIndex((item) => item.id === 'Premium_LRS');
          sysDiskTypes.value.splice(sysIndex, 1);
          dataDiskTypes.value.splice(dataIndex, 1);
        }
      }
    },
  );

  watch(
    () => formData.instance_charge_type,
    (chargeType: string) => {
      if (cond.vendor === VendorEnum.TCLOUD) {
        if (chargeType === 'PREPAID') {
          internetChargeTypes.value = internetChargeTypeValues[VendorEnum.TCLOUD].filter(
            (item) => item.id !== 'BANDWIDTH_POSTPAID_BY_HOUR',
          );
        } else {
          internetChargeTypes.value = internetChargeTypeValues[VendorEnum.TCLOUD].filter(
            (item) => item.id !== 'BANDWIDTH_PREPAID',
          );
        }
      }
    },
    { immediate: true },
  );

  return {
    sysDiskTypes,
    dataDiskTypes,
    billingModes,
    purchaseDurationUnits,
    internetChargeTypes,
  };
};
