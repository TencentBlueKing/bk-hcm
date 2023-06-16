import { VendorEnum } from '@/common/constant';
import { ref, watch } from 'vue';
import type { Cond } from './use-condtion';
import type { IDiskFormData } from './use-disk-form-data';

// eslint-disable-next-line
export default (cond: Cond, formData: IDiskFormData) => {
  const diskTypes = ref([]);

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

  watch(
    () => cond.vendor,
    (vendor) => {
      const diskTypeValues = {
        [VendorEnum.TCLOUD]: [
          // {
          //   id: 'CLOUD_BASIC',
          //   name: '普通云硬盘',
          // },
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
          // {
          //   id: 'CLOUD_HSSD',
          //   name: '增强型SSD云硬盘',
          // },
          // {
          //   id: 'CLOUD_TSSD',
          //   name: '极速型SSD云硬盘',
          // },
        ],
        [VendorEnum.AWS]: [
          {
            id: 'gp3',
            name: 'General Purpose SSD(通用型SSD gp3)',
          },
          {
            id: 'gp2',
            name: 'General Purpose SSD(通用型SSD gp2)',
          },
          {
            id: 'io1',
            name: 'Provisioned IOPS SSD(预置 IOPS SSD io1',
          },
          {
            id: 'io2',
            name: 'Provisioned IOPS SSD(预置 IOPS SSD io2)',
          },
          {
            id: 'st1',
            name: 'Throughput Optimized HDD(吞吐优化型 HDD)',
          },
          {
            id: 'sc1',
            name: 'Cold HDD(Cold HDD)',
          },
          {
            id: 'standard',
            name: 'Magnetic(磁介质)',
          },
        ],
        [VendorEnum.AZURE]: [
          {
            id: 'Standard_LRS',
            name: '标准 HDD(本地冗余存储)',
          },
          {
            id: 'Premium_LRS',
            name: '高级 SSD(本地冗余存储)',
          },
          {
            id: 'StandardSSD_LRS',
            name: '标准 SSD(本地冗余存储)',
          },
          {
            id: 'UltraSSD_LRS',
            name: '超级 SSD(本地冗余存储)',
          },
        ],
        [VendorEnum.GCP]: [
          {
            id: 'pd-standard',
            name: '标准永久性磁盘',
          },
          {
            id: 'pd-balanced',
            name: '平衡的永久性磁盘',
          },
          {
            id: 'pd-ssd',
            name: 'SSD 永久性磁盘',
          },
          {
            id: 'pd-extreme',
            name: '极端永久性磁盘',
          },
        ],
        [VendorEnum.HUAWEI]: [
          {
            id: 'SATA',
            name: '普通IO云硬盘(已售罄)',
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
          // {
          //   id: 'GPSSD2',
          //   name: '通用型SSD V2云硬盘',
          // },
          // {
          //   id: 'ESSD2',
          //   name: '极速型SSD',
          // },
        ],
      };

      const billingModeValues = {
        [VendorEnum.TCLOUD]: [
          {
            id: 'PREPAID',
            name: '包年/包月',
          },
          {
            id: 'POSTPAID_BY_HOUR',
            name: '按需计费',
          },
        ],
        [VendorEnum.HUAWEI]: [
          {
            id: 'prePaid',
            name: '包年/包月',
          },
          {
            id: 'postPaid',
            name: '按量计费',
          },
        ],
      };

      diskTypes.value = diskTypeValues[vendor];

      billingModes.value = billingModeValues[vendor] || [];
    },
  );

  watch(
    () => cond.bizId,
    () => {
      diskTypes.value = [];
    },
  );

  return {
    diskTypes,
    billingModes,
    purchaseDurationUnits,
  };
};
