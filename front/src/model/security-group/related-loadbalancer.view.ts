import { ModelProperty } from '@/model/typings';
import accountProperties from '../account/properties';
import loadbalancerProperties from '../load-balancer/properties';

export default [
  ...accountProperties,
  ...loadbalancerProperties,
  {
    id: 'security_group_ids',
    name: '已绑定的安全组',
    type: 'array',
  },
  {
    id: 'security_group_names',
    name: '已绑定的安全组',
    type: 'array',
  },
] as ModelProperty[];
