export const CLOUD_TYPE = [
  {
    id: 'tcloud',
    name: '腾讯云',
  },
  {
    id: 'aws',
    name: '亚马逊',
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
];

export const BUSINESS_TYPE = [
  {
    id: 1,
    name: 'cmdb',
  },
  {
    id: 2,
    name: 'lesscode',
  },
  {
    id: 3,
    name: '开发者中心',
  },
];

export const ACCOUNT_TYPE = [
  // {
  //   label: '资源账号',
  //   value: 'resource',
  // },
  {
    label: '登记账号',
    value: 'registration',
  },
  {
    label: '安全审计账号',
    value: 'security_audit',
  },
];

export const ACCOUNT_TYPE_ENUM = {
  RESOURCE: 'resource',
  REGISTRATION: 'registration',
  SECURITY_AUDIT: 'security_audit',
};

export const SITE_TYPE = [
  {
    label: '中国站',
    value: 'china',
  },
  {
    label: '国际站',
    value: 'international',
  },
];

export const DESC_ACCOUNT = {
  tcloud: {
    vendor: '<p>腾讯云支持中国站 <a target="_blank" href="http://cloud.tencent.com">cloud.tencent.com</a></p>',
    accountInfo:
      '主账号ID，子账号ID如何查看？登陆控制台，鼠标停留在右上角的账号图标上。第一行个人账号@后的数字为主账号ID。第二行账号ID后的数字为子账号ID',
    apiSecret:
      '在腾讯云控制台-访问管理-访问密钥-API密钥管理中查看 <a target="_blank" href="https://console.cloud.tencent.com/cam/capi">https://console.cloud.tencent.com/cam/capi</a>',
  },
  aws: {
    vendor:
      '亚马逊云支持中国站(<a target="_blank" href="http://amazonaws.cn">amazonaws.cn</a>)和国际站(<a target="_blank" href="http://aws.amazon.com">aws.amazon.com</a>)',
    accountInfo: '账号ID和IAM用户如何查看？进入AWS控制台，点击右上角用户名，可见“账户ID”和“IAM用户名称”',
    apiSecret:
      '进入AWS控制台，点击右上角用户名，点击“安全凭证”（<a target="_blank" href="https://us-east-1.console.aws.amazon.com/iam/home?region=ap-northeast-1#/security_credentials">https://us-east-1.console.aws.amazon.com/iam/home?region=ap-northeast-1#/security_credentials</a>）在“访问密钥”栏下可查看访问密钥ID。',
  },
  azure: {
    vendor:
      '微软云支持中国站(<a target="_blank" href="http://azure.cn">azure.cn</a>)和国际站(<a target="_blank" href="http://azure.microsoft.com">azure.microsoft.com</a>)',
    accountInfo:
      '租户ID，在Azure portal里的“Azure Active Directory”里查看，链接：<a target="_blank" href="https://portal.azure.com/#view/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/~/Overview">portal.azure.com/#view/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/~/Overview</a>' +
      '订阅ID和订阅名称，在“订阅”里查看，链接：<a target="_blank" href="https://portal.azure.com/#view/Microsoft_Azure_Billing/SubscriptionsBlade">https://portal.azure.com/#view/Microsoft_Azure_Billing/SubscriptionsBlade</a>',
    apiSecret:
      '在“应用注册”里（链接：<a target="_blank" href="https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade">https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade</a> ）获取应用程序名称和应用程序（客户端）ID。',
  },
  gcp: {
    vendor: '谷歌云国际站(console.cloud.google.com)',
    accountInfo:
      '项目ID和项目名称进入GCP控制台->IAM和管理->设置，或直接点击链接：' +
      '<a target="_blank" href="https://console.cloud.google.com/iam-admin/settings?orgonly=true&supportedpurview=organizationId">https://console.cloud.google.com/iam-admin/settings?orgonly=true&supportedpurview=organizationId</a>可找到项目名称和项目ID。',
    apiSecret:
      '进入GCP控制台=>IAM和管理=>服务账号，或直接点击链接：' +
      '<a target="_blank" href="https://console.cloud.google.com/iam-admin/serviceaccounts?orgonly=true&supportedpurview=organizationId">https://console.cloud.google.com/iam-admin/serviceaccounts?orgonly=true&supportedpurview=organizationId</a>' +
      '可找到服务账号名称和密钥ID',
  },
  huawei: {
    vendor:
      '华为支持中国站(<a target="_blank" href="http://huaweicloud.com">huaweicloud.com</a>)和国际站(<a target="_blank" href="http://huaweicloud.com/intl/">huaweicloud.com/intl/</a>)',
    accountInfo:
      '主账号ID进入华为云控制台，点击右上方账号浮窗下的基本信息，在账号中心点击我的主账号，可找到主账号名。' +
      '<a target="_blank" href="https://account-intl.huaweicloud.com/usercenter/?region=ap-southeast-1&locale=zh-cn#/accountindex/associatedAccount">https://account-intl.huaweicloud.com/usercenter/?region=ap-southeast-1&locale=zh-cn#/accountindex/associatedAccount</a>' +
      '<p>IAM用户名、IAM用户ID、账号名和账号ID。进入华为云控制台，点击右上方账号浮窗下的“我的凭证”，在“API凭证”</p>' +
      '<a target="_blank" href="https://console-intl.huaweicloud.com/iam/?region=ap-southeast-1&locale=zh-cn#/mine/apiCredential">https://console-intl.huaweicloud.com/iam/?region=ap-southeast-1&locale=zh-cn#/mine/apiCredential</a>',
    apiSecret: '在“我的凭证”里，点击“访问密钥”，点击“新增访问密钥”。下载后打开对应csv文件获取。',
  },
};
