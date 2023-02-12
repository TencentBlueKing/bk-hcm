// 资源类型
export const RESOURCE_TYPES = [
  {
    name: '主机',
    type: 'host',
  },
  {
    name: 'VPC',
    type: 'vpc',
  },
  {
    name: '子网',
    type: 'subnet',
  },
  {
    name: '安全组/GCP防火墙',
    type: 'security',
  },
  {
    name: '云硬盘',
    type: 'drive',
  },
  {
    name: '网络接口',
    type: 'network-interface',
  },
  {
    name: '弹性 IP',
    type: 'ip',
  },
  {
    name: '路由表',
    type: 'routing',
  },
  {
    name: '镜像',
    type: 'image',
  },
];

// 云厂商
export const VENDORS = [
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
];
