import { VendorEnum } from "@/common/constant";

export const IP_RANGES: any = {
  [VendorEnum.TCLOUD]: [
    {
      ip: '10.0.0.0',
      mask: 8,
      minMask: 12,
      maxMask: 28,
    },
    {
      ip: '172.16.0.0',
      mask: 12,
      minMask: 12,
      maxMask: 28,
    },
    {
      ip: '192.168.0.0',
      mask: 16,
      minMask: 16,
      maxMask: 28,
    },
  ],
  // AWS\AZURE\GCP相同
  [VendorEnum.AWS]: [
    {
      ip: '10.0.0.0',
      mask: 8,
      minMask: 16,
      maxMask: 28,
    },
    {
      ip: '172.16.0.0',
      mask: 12,
      minMask: 12,
      maxMask: 28,
    },
    {
      ip: '192.168.0.0',
      mask: 16,
      minMask: 16,
      maxMask: 28,
    },
  ],
  [VendorEnum.AZURE]: [
    {
      ip: '10.0.0.0',
      mask: 8,
      minMask: 16,
      maxMask: 28,
    },
    {
      ip: '172.16.0.0',
      mask: 12,
      minMask: 16,
      maxMask: 28,
    },
    {
      ip: '192.168.0.0',
      mask: 16,
      minMask: 16,
      maxMask: 28,
    },
  ],
  [VendorEnum.GCP]: [
    {
      ip: '10.0.0.0',
      mask: 8,
      minMask: 16,
      maxMask: 28,
    },
    {
      ip: '172.16.0.0',
      mask: 12,
      minMask: 16,
      maxMask: 28,
    },
    {
      ip: '192.168.0.0',
      mask: 16,
      minMask: 16,
      maxMask: 28,
    },
  ],
  [VendorEnum.HUAWEI]: [
    {
      ip: '10.0.0.0',
      mask: 8,
      minMask: 8,
      maxMask: 28,
    },
    {
      ip: '172.16.0.0',
      mask: 12,
      minMask: 12,
      maxMask: 28,
    },
    {
      ip: '192.168.0.0',
      mask: 16,
      minMask: 16,
      maxMask: 28,
    },
  ],
}