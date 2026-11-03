import { VendorEnum } from '@/common/constant';
import tcloudVendorIcon from '@/assets/image/vendor-tcloud.svg';
import awsVendorIcon from '@/assets/image/vendor-aws.svg';
import azureVendorIcon from '@/assets/image/vendor-azure.svg';
import gcpVendorIcon from '@/assets/image/vendor-gcp.svg';
import huaweiVendorIcon from '@/assets/image/vendor-huawei.svg';
import otherVendorIcon from '@/assets/image/vendor-other.svg';

export const vendorProperty = new Map<VendorEnum, { icon: any }>([
  [
    VendorEnum.TCLOUD,
    {
      icon: tcloudVendorIcon,
    },
  ],
  [
    VendorEnum.AWS,
    {
      icon: awsVendorIcon,
    },
  ],
  [
    VendorEnum.AZURE,
    {
      icon: azureVendorIcon,
    },
  ],
  [
    VendorEnum.GCP,
    {
      icon: gcpVendorIcon,
    },
  ],
  [
    VendorEnum.HUAWEI,
    {
      icon: huaweiVendorIcon,
    },
  ],
  [
    VendorEnum.OTHER,
    {
      icon: otherVendorIcon,
    },
  ],
]);
