export enum AccountLevelEnum  {
  FirstLevel = 'firstAccount',
  SecondLevel = 'secondaryAccount',
}

export const tabs = [
  {
    key: AccountLevelEnum.FirstLevel,
    label: '一级账号',
    component: ''
  },
  {
    key: AccountLevelEnum.SecondLevel,
    label: '二级账号',
  }
];

export const reviewData = [
  {
      primaryAccountName: '主账户A',
      primaryAccountId: '12345',
      cloudProvider: 'AWS',
      accountEmail: 'exampleA@example.com',
      mainResponsiblePerson: '张三',
      organizationalStructure: '部门A',
      secondaryAccountCount: 5,
      actions: 'Delete',
  },
  {
      primaryAccountName: '主账户B',
      primaryAccountId: '67890',
      cloudProvider: 'Azure',
      accountEmail: 'exampleB@example.com',
      mainResponsiblePerson: '李四',
      organizationalStructure: '部门B',
      secondaryAccountCount: 3,
      actions: 'Update',
  },
  {
      primaryAccountName: '主账户C',
      primaryAccountId: '54321',
      cloudProvider: 'Google Cloud',
      accountEmail: 'exampleC@example.com',
      mainResponsiblePerson: '王五',
      organizationalStructure: '部门C',
      secondaryAccountCount: 7,
      actions: 'View',
  },
];

export const searchData = [
  {
      name: '一级帐号名称',
      id: 'primaryAccountName',
  },
  {
      name: '一级帐号ID',
      id: 'primaryAccountId',
  },
  {
      name: '云厂商',
      id: 'cloudProvider',
  },
  {
      name: '帐号邮箱',
      id: 'accountEmail',
  },
  {
      name: '主负责人',
      id: 'mainResponsiblePerson',
  },
  {
      name: '组织架构',
      id: 'organizationalStructure',
  },
  {
      name: '二级帐号个数',
      id: 'secondaryAccountCount',
  },
  {
      name: '操作',
      id: 'actions',
  },
];

// 二级账号
export const secondaryReviewData = [
  {
      secondaryAccountName: '子账户A',
      secondaryAccountId: '10001',
      parentPrimaryAccount: '主账户A',
      cloudProvider: 'AWS',
      siteType: '电商',
      accountEmail: 'childA@example.com',
      mainResponsiblePerson: '刘一',
      operatingProduct: '产品A',
      actions: 'Delete',
  },
  {
      secondaryAccountName: '子账户B',
      secondaryAccountId: '10002',
      parentPrimaryAccount: '主账户B',
      cloudProvider: 'Azure',
      siteType: '社交',
      accountEmail: 'childB@example.com',
      mainResponsiblePerson: '陈二',
      operatingProduct: '产品B',
      actions: 'Update',
  },
  {
      secondaryAccountName: '子账户C',
      secondaryAccountId: '10003',
      parentPrimaryAccount: '主账户C',
      cloudProvider: 'Google Cloud',
      siteType: '媒体',
      accountEmail: 'childC@example.com',
      mainResponsiblePerson: '张三',
      operatingProduct: '产品C',
      actions: 'View',
  },
];

export const secondarySearchData = [
  {
      name: '二级帐号名称',
      id: 'secondaryAccountName',
  },
  {
      name: '二级帐号ID',
      id: 'secondaryAccountId',
  },
  {
      name: '所属一级帐号',
      id: 'parentPrimaryAccount',
  },
  {
      name: '云厂商',
      id: 'cloudProvider',
  },
  {
      name: '站点类型',
      id: 'siteType',
  },
  {
      name: '帐号邮箱',
      id: 'accountEmail',
  },
  {
      name: '主负责人',
      id: 'mainResponsiblePerson',
  },
  {
      name: '运营产品',
      id: 'operatingProduct',
  },
  {
      name: '操作',
      id: 'actions',
  },
];