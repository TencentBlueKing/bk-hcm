import { VendorEnum } from '@/common/constant';
import tcloudVendorIcon from '@/assets/image/vendor-tcloud.svg';
import awsVendorIcon from '@/assets/image/vendor-aws.svg';
import azureVendorIcon from '@/assets/image/vendor-azure.svg';
import gcpVendorIcon from '@/assets/image/vendor-gcp.svg';
import huaweiVendorIcon from '@/assets/image/vendor-huawei.svg';

export const vendorProperty: { [K in VendorEnum]?: { icon: any; style: any } } = {
  [VendorEnum.TCLOUD]: {
    icon: tcloudVendorIcon,
    style: { backgroundColor: '#DAE9FD', color: '#4193E5' },
  },
  [VendorEnum.AWS]: {
    icon: awsVendorIcon,
    style: { backgroundColor: '#FFF2C9', color: '#E68D00' },
  },
  [VendorEnum.AZURE]: {
    icon: azureVendorIcon,
    style: { backgroundColor: '#D8F4F5', color: '#45A0A5' },
  },
  [VendorEnum.GCP]: {
    icon: gcpVendorIcon,
    style: { backgroundColor: '#DAF5C8', color: '#3FAA3B' },
  },
  [VendorEnum.HUAWEI]: {
    icon: huaweiVendorIcon,
    style: { backgroundColor: '#FFDDDD', color: '#EA4646' },
  },
};
