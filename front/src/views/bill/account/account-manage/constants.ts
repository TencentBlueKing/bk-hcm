import { convertToIdNameMap } from "./util";

export enum AccountLevelEnum {
  FirstLevel = 'firstAccount',
  SecondLevel = 'secondaryAccount',
}

export const tabs = [
  {
    key: AccountLevelEnum.FirstLevel,
    label: '一级账号',
    component: '',
  },
  {
    key: AccountLevelEnum.SecondLevel,
    label: '二级账号',
  },
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

// 云厂商
export const BILL_VENDORS = [
  {
    id: 'tcloud',
    name: '腾讯云',
  },
  {
    id: 'aws',
    name: '亚马逊云',
  },
  {
    id: 'azure',
    name: '微软云',
  },
  {
    id: 'gcp',
    name: '谷歌云',
  },
  {
    id: 'huawei',
    name: '华为云',
  },
  {
    id: 'zenlayer',
    name: 'zenlayer',
  },
  {
    id: 'kaopu',
    name: '靠谱云',
  },
];

export const BILL_VENDORS_MAP = convertToIdNameMap(BILL_VENDORS);

export const searchData = [
  {
    name: '一级帐号名称',
    id: 'name',
  },
  {
    name: '一级帐号ID',
    id: 'cloud_id',
  },
  {
    name: '云厂商',
    id: 'vendor',
    children: BILL_VENDORS,
  },
  {
    name: '帐号邮箱',
    id: 'email',
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

export const BILL_SITE_TYPES = [
    {
        name: '中国站',
        id: 'china',
      },
      {
        name: '国际站',
        id: 'international',
      },
];

export const BILL_SITE_TYPES_MAP = convertToIdNameMap(BILL_SITE_TYPES);

export const secondarySearchData = [
  {
    name: '二级帐号名称',
    id: 'name',
  },
  {
    name: '二级帐号ID',
    id: 'cloud_id',
  },
  {
    name: '所属一级帐号名称',
    id: 'parent_account_name',
  },
  {
    name: '云厂商',
    id: 'vendor',
  },
  {
    name: '站点类型',
    id: 'site',
    children: BILL_SITE_TYPES,
  },
  {
    name: '帐号邮箱',
    id: 'accountEmail',
  },
];
