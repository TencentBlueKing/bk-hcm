import { parse, parseCIDR, IPv4, isValid } from 'ipaddr.js';
import { SecurityRule } from './add-rule';
import { VendorEnum } from '@/common/constant';

export const securityRuleValidators = (data: SecurityRule, vendor: VendorEnum) => {
  return {
    protocalAndPort: [
      {
        trigger: 'change',
        message: '协议和端口均不能为空',
        validator: () => {
          if (['cloud_service_id', 'cloud_service_group_id'].includes(data.protocol)) return true;
          return (!!data.port || vendor === VendorEnum.HUAWEI) && !!data.protocol;
        },
      },
      {
        trigger: 'blur',
        message: '请填写合法的端口号, 注意需要在 1-65535 之间, 若需使用逗号时请注意使用英文逗号,',
        validator: () => {
          if (['cloud_service_id', 'cloud_service_group_id'].includes(data.protocol)) return true;
          return vendor === VendorEnum.HUAWEI || isPortAvailable(data.port);
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
            const ipType = validateIpCidr(data[val]);
            if (ipType === IpType.invalid) return false;
            if ([IpType.ipv4, IpType.ipv4_cidr].includes(ipType) && val !== 'ipv4_cidr') return false;
            if ([IpType.ipv6, IpType.ipv6_cidr].includes(ipType) && val !== 'ipv6_cidr') return false;
          }
          return true;
        },
      },
      {
        trigger: 'blur',
        message: '填写格式不正确。所有IPv4地址：0.0.0.0/0，所有IPv6地址：0::0/0或::/0',
        validator: (val: string) => {
          return !['0.0.0.0', '0::0', '::'].includes(data[val]);
        },
      },
      {
        trigger: 'blur',
        message: '填写对应合法的 IP CIDR (必须带子网掩码), 注意区分 IPV4 与 IPV6',
        validator: (val: string) => {
          if (vendor === VendorEnum.AWS) {
            return [IpType.cidr, IpType.ipv4_cidr, IpType.ipv6_cidr].includes(validateIpCidr(data[val]));
          }
          return true;
        },
      },
      {
        trigger: 'blur',
        message: '请填写合法的 IP',
        validator: (val: string) => {
          if (['remote_ip_prefix', 'source_address_prefix'].includes(val)) {
            return validateIpCidr(data[val]) !== IpType.invalid;
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
            return validateIpCidr(data[val]) !== IpType.invalid;
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
        message: '请填写合法的端口号, 注意需要在 1-65535 之间, 若需使用逗号时请注意使用英文逗号,',
        validator: (val: string | number) => {
          return data.protocol === '*' || isPortAvailable(val);
        },
      },
    ],
    source_port_range: [
      {
        trigger: 'blur',
        message: '请填写合法的端口号, 注意需要在 1-65535 之间, 若需使用逗号时请注意使用英文逗号,',
        validator: isPortAvailable,
      },
    ],
    // memo: [
    //   {
    //     trigger: 'change',
    //     message:
    //       'Invalid rule description. Valid descriptions are strings less
    //        than 256 characters from the following set: a-zA-Z0-9. _-:/()#,@[]+=&;{}!$*',
    //     pattern: /^[a-zA-Z0-9. _\-:/()#,@[\]+=&;{}!$*]{0,256}$/,
    //   },
    // ],
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
    return !nums.some((num) => +num < 0 || +num > 65535);
  }
  if (/-/g.test(port)) {
    const nums = port.split(/-/);
    return !nums.some((num) => +num < 0 || +num > 65535) && +nums[0] < +nums[1];
  }
  return +port >= 0 && +port <= 65535;
};

/**
 * 检查是否合法的 IP CIDR
 * @param ip IP CIDR
 * @returns ip 类型，不合法则返回 'invalid'
 */
export const validateIpCidr = (ip: string): IpType => {
  ip = ip?.trim();
  if (isValid(ip)) {
    const type = parse(ip).kind();
    if (type === IpType.ipv4 && IPv4.isValidFourPartDecimal(ip)) return IpType.ipv4;
    if (type === IpType.ipv6) return IpType.ipv6;
    return IpType.invalid;
  }
  try {
    const [host, _mask] = parseCIDR(ip);
    const host_type = host.kind();
    if (host_type === IpType.ipv4) return IpType.ipv4_cidr;
    if (host_type === IpType.ipv6) return IpType.ipv6_cidr;
  } catch (err) {
    return IpType.invalid;
  }
  return IpType.cidr;
};

export enum IpType {
  invalid = 'invalid',
  ipv4 = 'ipv4',
  ipv6 = 'ipv6',
  cidr = 'cidr',
  ipv4_cidr = 'ipv4_cidr',
  ipv6_cidr = 'ipv6_cidr',
}
