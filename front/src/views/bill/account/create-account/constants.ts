import { VendorEnum } from '@/common/constant';
import tcloudVendor from '@/assets/image/vendor-tcloud.svg';
import awsVendor from '@/assets/image/vendor-aws.svg';
import azureVendor from '@/assets/image/vendor-azure.svg';
import gcpVendor from '@/assets/image/vendor-gcp.svg';
import huaweiVendor from '@/assets/image/vendor-huawei.svg';
import zenlayerVendor from '@/assets/image/zenlayer.png';
import kaopuVendor from '@/assets/image/kaopu.png';
import disabledTcloudVendor from '@/assets/image/disabled-tcloud.png';

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

export const BILL_VENDORS_INFO = [
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
  {
    vendor: VendorEnum.AWS,
    name: '亚马逊云',
    icon: awsVendor,
  },
  {
    vendor: VendorEnum.ZENLAYER,
    name: 'zenlayer',
    icon: zenlayerVendor,
  },
  {
    vendor: VendorEnum.KAOPU,
    name: '靠谱云',
    icon: kaopuVendor,
  },
];

export const MAIN_ACCOUNT_VENDORS = [
  ...BILL_VENDORS_INFO,
  {
    vendor: VendorEnum.TCLOUD,
    name: '腾讯云',
    icon: disabledTcloudVendor,
  },
];
