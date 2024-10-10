import { cloneDeep } from 'lodash-es';
import { TcloudSecurityGroupRule, TcloudSourceTypeArr, TcloudTemplatePort, TcloudTemplatePortArr } from '.';
import { TcloudSourceAddressType } from './SourceAddress';

export const tcloudHandler = (data: TcloudSecurityGroupRule & { sourceAddress: TcloudSourceAddressType }) => {
  // 协议选择参数模板端口/端口组
  if (TcloudTemplatePortArr.includes(data.protocol as TcloudTemplatePort)) {
    delete data.protocol;
    delete data.port;
  }
  // 仅保留选中的源地址类型
  TcloudSourceTypeArr.forEach((type) => data.sourceAddress !== type && delete data[type]);
};

export const tcloudPreHandler = (data: TcloudSecurityGroupRule & { sourceAddress: TcloudSourceAddressType }) => {
  const res = cloneDeep(data);
  // 源地址类型
  TcloudSourceTypeArr.forEach((type) => res[type] && (res.sourceAddress = type));

  // 协议为参数模板时，给协议、端口赋值
  TcloudTemplatePortArr.forEach((type) => res[type] && (res.protocol = type));

  return res;
};
