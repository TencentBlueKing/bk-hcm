import { parse, parseCIDR, IPv4, isValid } from 'ipaddr.js';
import { SecurityRule } from './add-rule';
import { VendorEnum } from '@/common/constant';
const { isValidFourPartDecimal } = IPv4;

export const securityRuleValidators = (data: SecurityRule, vendor: VendorEnum) => {
  return {
    protocalAndPort: [
      {
        trigger: 'change',
        message: '协议和端口均不能为空',
        validator: () => {
          return !!data.port && !!data.protocol;
        },
      },
      {
        trigger: 'blur',
        message: '请填写合法的端口号, 注意需要在 0-65535 之间, 若需使用逗号时请注意使用英文逗号,',
        validator: () => {
          return isPortAvailable(data.port);
        },
      },
    ],
    priority: [
      {
        trigger: 'change',
        message: '必须是 1-100的整数',
        validator: (val: string | number) => {
          if ([VendorEnum.HUAWEI].includes(vendor)) {
            return Number.isInteger(+val) && +val <= 100 && +val >= 0;
          }
          return true;
        },
      },
      {
        trigger: 'change',
        message: '取值范围为100-4096',
        validator: (val: string | number) => {
          if ([VendorEnum.AZURE].includes(vendor)) {
            return Number.isInteger(+val) && +val <= 4096 && +val >= 100;
          }
          return true;
        },
      },
    ],
    sourceAddress: [
      {
        trigger: 'blur',
        message: '源地址类型与内容均不能为空',
        validator: (val: string) => {
          return !!val && !!data[val];
        },
      },
      {
        trigger: 'change',
        message: 'ICMPV6 不支持 IPV4',
        validator: (val: string) => !(data.protocol === 'icmpv6' && val === 'ipv4_cidr'),
      },
      {
        trigger: 'blur',
        message: '请填写对应合法的 IP, 注意区分 IPV4 与 IPV6',
        validator: (val: string) => {
          if (['ipv6_cidr', 'ipv4_cidr'].includes(val)) {
            return isValidIpCidr(data[val], true);
          }
          return true;
        },
      },
      {
        trigger: 'blur',
        message: '仅支持 CIDR',
        validator: (val: string) => {
          if (vendor === VendorEnum.AWS) {
            const ip = data[val].trim();
            try {
              parseCIDR(ip);
            } catch (err) {
              return false;
            }
          }
          return true;
        },
      },
      {
        trigger: 'blur',
        message: '请填写合法的 IP',
        validator: (val: string) => {
          if (['remote_ip_prefix', 'source_address_prefix'].includes(val)) {
            return isValidIpCidr(data[val], false);
          }
          return true;
        },
      },
    ],
    targetAddress: [
      {
        trigger: 'blur',
        message: '目标类型与内容均不能为空',
        validator: (val: string) => {
          return !!val && !!data[val];
        },
      },
      {
        trigger: 'blur',
        message: '请填写合法的 IP',
        validator: (val: string) => {
          if (['destination_address_prefix'].includes(val)) {
            return isValidIpCidr(data[val], false);
          }
          return true;
        },
      },
    ],
    destination_port_range: [
      {
        trigger: 'change',
        message: '目标协议端口不能为空',
        validator: (val: string) => {
          return data.protocol === '*' || (!!data.protocol && !!val);
        },
      },
      {
        trigger: 'blur',
        message: '请填写合法的端口号, 注意需要在 0-65535 之间, 若需使用逗号时请注意使用英文逗号,',
        validator: (val: string | number) => {
          return data.protocol === '*' || isPortAvailable(val);
        },
      },
    ],
    source_port_range: [
      {
        trigger: 'blur',
        message: '请填写合法的端口号, 注意需要在 0-65535 之间, 若需使用逗号时请注意使用英文逗号,',
        validator: isPortAvailable,
      },
    ],
  };
};

/**
 * 检查端口号是否合法
 * @param val 端口号、端口范围、多个端口
 * @returns boolean
 */
export const isPortAvailable = (val: string | number) => {
  const port = String(val).trim();
  const isPortValid = /^(ALL|0|-1|[1-9]\d*|(\d+,\d+)+|(\d+-\d+)+)$/.test(port);
  if (!isPortValid) return false;
  if (/^ALL$/.test(port) || +port === 0) return true;
  if (/,/g.test(port)) {
    const nums = port.split(/,/);
    return !nums.some(num => +num < 0 || +num > 65535);
  }
  if (/-/g.test(port)) {
    const nums = port.split(/-/);
    return !nums.some(num => +num < 0 || +num > 65535);
  }
  return +port >= 0 && +port <= 65535;
};

/**
 * 检查是否合法的 IP CIDR
 * @param val IP CIDR
 * @param hasVersion 是否区分IPV4与IPV6
 * @returns boolean
 */
export const isValidIpCidr = (val: string, hasVersion: boolean) => {
  const ip = val.trim();
  if (isValid(ip)) {
    const ipType = parse(ip).kind();
    if (hasVersion) {
      return (ipType === 'ipv4' && val === 'ipv4_cidr' && isValidFourPartDecimal(ip)) || (ipType === 'ipv6' && val === 'ipv6_cidr');
    }
    return (ipType === 'ipv4' && isValidFourPartDecimal(ip)) || ipType === 'ipv6';
  }
  try {
    parseCIDR(ip);
  } catch (err) {
    return false;
  }
};
