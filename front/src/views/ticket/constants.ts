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
  create_sub_account: '新增三级账号',
  update_sub_account: '修改三级账号',
  delete_sub_account: '删除三级账号',
  create_sub_account_secret: '新增三级账号密钥',
  delete_sub_account_secret: '删除三级账号密钥',
  update_sub_account_secret: '修改三级账号密钥状态',
  apply_permission_policy_library_create: '策略库应用到模板',
  apply_permission_policy_library_update: '策略库更新到模板',
  create_permission_template: '创建权限模板',
  update_permission_template: '修改权限模板',
  delete_permission_template: '删除权限模板',
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
    id: 'operation',
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
  { name: '申请人', id: 'applicant', async: true },
];
