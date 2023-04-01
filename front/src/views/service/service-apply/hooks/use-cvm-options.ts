import { ref, watch } from 'vue';
import { VendorEnum } from '@/common/constant';
import type { Cond } from './use-condtion';
import type { ICvmFormData } from './use-cvm-form-data';

export default (cond: Cond, formData: ICvmFormData) => {
  const sysDiskTypes = ref([]);

  const dataDiskTypes = ref([]);

  const billingModes = ref([]);

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

  watch(() => cond.vendor, (vendor) => {
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

    sysDiskTypes.value = sysDiskTypeValues[vendor];

    dataDiskTypes.value = dataDiskTypeValues[vendor];

    billingModes.value = billingModeValues[vendor] || [];
  });

  watch(() => cond.bizId, () => {
    sysDiskTypes.value = [];
    dataDiskTypes.value = [];
  });

  return {
    sysDiskTypes,
    dataDiskTypes,
    billingModes,
    purchaseDurationUnits,
  };
};
