import { QueryRuleOPEnum } from "@/typings";

export const searchData = [
  {
    name: '申请ID',
  },
  {
    name: '来源',
    id: 'source',
  },
  {
    name: '序列号',
    id: 'sn',
  },
  {
    name: '申请类型',
    id: 'type',
  },
  {
    name: '申请状态',
    id: 'status',
  },
  {
    name: '申请人',
    id: 'applicant',
  },
  {
    name: '申请内容',
    id: 'content',
  },
  {
    name: '交付详情',
    id: 'delivery_detail',
  },
  {
    name: '备注',
    id: 'memo',
  },
  {
    name: '创建者',
    id: 'creator',
  },
  {
    name: '更新者',
    id: 'reviser',
  },
  {
    name: '创建时间',
    id: 'created_at',
  },
  {
    name: '更新时间',
    id: 'updated_at',
  },
];

export const APPLY_TYPES = [
  {
    label: '全部',
    name: 'all',
    rule: [
    ],
  },
  {
    label: '云主机',
    name: 'cloudMachines',
    rule: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: [
          'create_cvm',
        ],
      }
    ],
  },
  {
    label: '账号',
    name: 'account',
    rule: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: [
          'add_account',
        ],
      }
    ],
  },
  {
    label: '硬盘',
    name: 'disk',
    rule: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: [
          'create_disk',
        ],
      }
    ],
  },
  {
    label: 'VPC',
    name: 'vpc',
    rule: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: [
          'create_disk',
        ],
      }
    ],
  },
  {
    label: '安全组',
    name: 'securityGroup',
    rule: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: [
          "create_security_group",
          "update_security_group",
          "delete_security_group",
          "associate_security_group",
          "disassociate_security_group",
          "create_security_group_rule",
          "update_security_group_rule",
          "delete_security_group_rule"
        ],
      }
    ],
  },
  {
    label: '负载均衡',
    name: 'loadBalancer',
    rule: [
      {
        field: 'type',
        op: QueryRuleOPEnum.IN,
        value: [
          'create_load_balancer',
        ],
      }
    ],
  },
];
