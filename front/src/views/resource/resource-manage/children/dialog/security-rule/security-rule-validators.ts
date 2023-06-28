import { isValid, parse, parseCIDR } from 'ipaddr.js';
import { SecurityRule } from './add-rule';

export const securityRuleValidators = (data: SecurityRule) => {
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
    sourceAddress: [
      {
        trigger: 'blur',
        message: '源地址类型与内容均不能为空',
        validator: (val: string) => {
          return !!val && !!data[val];
        },
      },
      {
        trigger: 'blur',
        message: '请填写对应合法的 IP, 注意区分 IPV4 与 IPV6',
        validator: (val: string) => {
          if (['ipv6_cidr', 'ipv4_cidr', 'source_address_prefix'].includes(val)) {
            const ip = data[val].trim();
            if (isValid(ip)) {
              if (['source_address_prefix'].includes(val)) return true;
              const ipType = parse(ip).kind();
              return (ipType === 'ipv4' && val === 'ipv4_cidr') || (ipType === 'ipv6' && val === 'ipv6_cidr');
            }
            try {
              parseCIDR(ip);
            } catch (err) {
              return false;
            }
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
        message: '请填写对应合法的 IP, 注意区分 IPV4 与 IPV6',
        validator: (val: string) => {
          if (['destination_address_prefix'].includes(val)) {
            const ip = data[val].trim();
            if (isValid(ip)) return true;
            try {
              parseCIDR(ip);
            } catch (err) {
              return false;
            }
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
  const isPortValid = /^(ALL|0|[1-9]\d*|(\d+,\d+)+|(\d+-\d+)+)$/.test(port);
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
