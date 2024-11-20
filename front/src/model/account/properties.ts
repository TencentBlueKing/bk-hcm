import { VendorMap } from '@/common/constant';
import { ModelProperty } from '@/model/typings';

export default [
  {
    id: 'account_id',
    name: '云账号',
    type: 'account',
  },
  {
    id: 'vendor',
    name: '云厂商',
    type: 'enum',
    option: VendorMap,
  },
] as ModelProperty[];
