import { QueryRuleOPEnum } from '@/typings';

export const reverseObj = (originalMap: Object) => {
  Object.fromEntries(Object.entries(originalMap).map(([key, value]) => [value, key]));
};

// 单据类型映射 (英中)
export const APPLICATION_TYPE_MAP: Record<string, string> = {
  add_account: '资源接入账号',
  create_cvm: '创建虚拟机',
  create_vpc: '创建VPC',
  create_disk: '创建云盘',
  create_main_account: '创建二级账号',
  update_main_account: '修改二级账号',
  create_load_balancer: '创建负载均衡',
  create_security_group: '创建安全组',
  update_security_group: '更新安全组',
  delete_security_group: '删除安全组',
  associate_security_group: '安全组关联资源',
  disassociate_security_group: '安全组资源解关联',
  create_security_group_rule: '创建安全组规则',
  update_security_group_rule: '更新安全组规则',
  delete_security_group_rule: '删除安全组规则',
};

// 单据类型映射 (中英)
export const APPLICATION_TYPE_MAP_CN = reverseObj(APPLICATION_TYPE_MAP);

// 单据申请状态映射
export const APPLICATION_STATUS_MAP: Record<string, string> = {
  pending: '待审批',
  pass: '审批通过',
  rejected: '审批拒绝',
  cancelled: '审批撤销',
  delivering: '审批中',
  completed: '交付成功',
  deliver_partial: '部分成功',
  deliver_error: '交付失败',
};

// 二级账号管理单据类型
export const ACCOUNT_TYPES = ['create_main_account', 'update_main_account'];

export const searchData = [
  { name: '单号', id: 'sn' },
  {
    name: '申请类型',
    id: 'type',
    children: Object.keys(APPLICATION_TYPE_MAP).map((key) => {
      return { name: APPLICATION_TYPE_MAP[key], id: key };
    }),
  },
  {
    name: '申请状态',
    id: 'status',
    children: Object.keys(APPLICATION_STATUS_MAP).map((key) => {
      return { name: APPLICATION_STATUS_MAP[key], id: key };
    }),
  },
  { name: '申请人', id: 'applicant' },
];

export const APPLY_TYPES = [
  {
    label: '全部',
    name: 'all',
    rules: [],
  },
  {
    label: '云主机',
    name: 'cloudMachines',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['create_cvm'],
      },
    ],
  },
  {
    label: '账号',
    name: 'account',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['add_account', 'create_main_account', 'update_main_account'],
      },
    ],
  },
  {
    label: '硬盘',
    name: 'disk',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['create_disk'],
      },
    ],
  },
  {
    label: 'VPC',
    name: 'vpc',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['create_disk'],
      },
    ],
  },
  {
    label: '安全组',
    name: 'securityGroup',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: [
          'create_security_group',
          'update_security_group',
          'delete_security_group',
          'associate_security_group',
          'disassociate_security_group',
          'create_security_group_rule',
          'update_security_group_rule',
          'delete_security_group_rule',
        ],
      },
    ],
  },
  {
    label: '负载均衡',
    name: 'load_balancer',
    rules: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: ['create_load_balancer'],
      },
    ],
  },
];
