import { VendorEnum } from "@/common/constant";
import tcloudVendor from '@/assets/image/vendor-tcloud.svg';
import awsVendor from '@/assets/image/vendor-aws.svg';
import azureVendor from '@/assets/image/vendor-azure.svg';
import gcpVendor from '@/assets/image/vendor-gcp.svg';
import huaweiVendor from '@/assets/image/vendor-huawei.svg';

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