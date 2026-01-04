import isIP from 'validator/es/lib/isIP';
import { AddressDescription } from '@/typings';
import { IAuthSign } from '@/common/auth-service';
import { ExportOptions } from '@/vendor/Export2Excel';

const getAuthSignByBusinessId = (
  businessId: number,
  rscAuthSymbol: symbol,
  bizAuthSymbol: symbol,
): IAuthSign | IAuthSign[] => {
  if (businessId) return { type: bizAuthSymbol, relation: [businessId] };
  return { type: rscAuthSymbol };
};

/**
 * 获取实例的ip地址
 * @param inst 实例
 * @returns 实例的ip地址
 */
const getInstVip = (inst: any) => {
  const {
    private_ipv4_addresses,
    private_ipv6_addresses,
    public_ipv4_addresses,
    public_ipv6_addresses,
    private_ip_address,
    public_ip_address,
  } = inst ?? {};
  if (private_ipv4_addresses || private_ipv6_addresses || public_ipv4_addresses || public_ipv6_addresses) {
    if (public_ipv4_addresses.length > 0) return public_ipv4_addresses.join(',');
    if (public_ipv6_addresses.length > 0) return public_ipv6_addresses.join(',');
    if (private_ipv4_addresses.length > 0) return private_ipv4_addresses.join(',');
    if (private_ipv6_addresses.length > 0) return private_ipv6_addresses.join(',');
  }
  if (private_ip_address || public_ip_address) {
    if (private_ip_address.length > 0) return private_ip_address.join(',');
    if (public_ip_address.length > 0) return public_ip_address.join(',');
  }

  return '--';
};

const getPrivateIPs = (data: any) => {
  return [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--';
};
const getPublicIPs = (data: any) => {
  return [...(data.public_ipv4_addresses || []), ...(data.public_ipv6_addresses || [])].join(',') || '--';
};

/**
 * 清洗请求载荷，去除空值
 * @param payload 请求载荷
 * @returns 返回新的请求载荷
 */
const cleanPayload = (payload: any) => {
  const newPayload = {};
  Object.keys(payload).forEach((key) => {
    if (Object.prototype.hasOwnProperty.call(payload, key)) {
      const value = payload[key];
      if (value !== '' && !(Array.isArray(value) && value.length === 0)) {
        newPayload[key] = value;
      }
    }
  });
  return newPayload;
};

/** 导出列配置 */
interface ExportColumn {
  /** 列标题 */
  label?: string;
  /** 字段名，支持嵌套路径如 'a.b.c' */
  field?: string;
  /** 列类型，如 'selection'、'index' 等，有 type 的列会被过滤 */
  type?: string;
  /** 单元格格式化函数 */
  formatter?: (cellData: Record<string, any>) => string | number;
  /** 导出专用格式化函数，接收整行数据 */
  exportFormatter?: (rowData: Record<string, any>) => string | number;
}

/**
 * 导出表格数据为 Excel
 * @param list 表格数据
 * @param columns 表格列配置
 * @param filename 文件名，自动添加时间戳
 */
const exportTableToExcel = (
  list: Record<string, any>[],
  columns: ExportColumn[],
  filename: string,
  opts?: ExportOptions,
): Promise<void> => {
  return new Promise((resolve, reject) => {
    import('@/vendor/Export2Excel')
      .then(async (excel) => {
        const header = columns.map((col) => col.label).filter((label): label is string => !!label);
        const newColumns = columns.filter((item) => !item.type);

        function getNestedProperty(obj: Record<string, any>, path: string): any {
          return path.split('.').reduce((acc, part) => acc?.[part], obj);
        }

        const data = list.map((item, rowIndex) =>
          newColumns.map((col, colIndex) => {
            try {
              if (col.formatter && col.field) {
                return col.formatter({ [col.field]: item[col.field] });
              }

              if (col.exportFormatter) {
                return col.exportFormatter(item);
              }

              if (col.field?.includes('.')) {
                return getNestedProperty(item, col.field);
              }
              return col.field ? item?.[col.field] : '';
            } catch (cellError) {
              console.warn(`exportTableToExcel: 处理单元格数据失败 [行${rowIndex}, 列${colIndex}]`, cellError);
              return ''; // 单元格错误时返回空字符串，不中断整个导出
            }
          }),
        );

        await excel.export_json_to_excel({
          header,
          data,
          filename: `${filename}${getDate('yyyyMMddhhmmss')}`,
          ...opts,
        });
        resolve();
      })
      .catch((err) => {
        reject(new Error(`exportTableToExcel: 导出失败 - ${err.message || err}`));
      });
  });
};
const getDate = (fmt: string, n?: number) => {
  let d;
  if (n) {
    let nd = Date.parse(new Date());
    nd = nd + n * 86400000;
    d = new Date(nd);
  } else {
    d = new Date();
  }
  const o = {
    'M+': d.getMonth() + 1, // 月份
    'd+': d.getDate(), // 日
    'h+': d.getHours(), // 小时
    'm+': d.getMinutes(), // 分
    's+': d.getSeconds(), // 秒
    'q+': Math.floor((d.getMonth() + 3) / 3), // 季度
    S: d.getMilliseconds(), // 毫秒
  };

  if (/(y+)/.test(fmt)) {
    fmt = fmt.replace(RegExp.$1, `${d.getFullYear()}`.substr(4 - RegExp.$1.length));
  }
  Object.keys(o).forEach((k) => {
    if (new RegExp(`(${k})`).test(fmt)) {
      fmt = fmt.replace(RegExp.$1, RegExp.$1.length === 1 ? o[k] : `00${o[k]}`.substr(`${o[k]}`.length));
    }
  });
  return fmt;
};

// 拼接 接口 路径
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
const getEntirePath = (tailPath: string, interfacePrefix = '/api/v1/woa/') => {
  return `${BK_HCM_AJAX_URL_PREFIX + interfacePrefix + tailPath}`;
};

const getDisplayText = (value: any, placeholder = '--') => {
  if (value === null || value === undefined || value === '') {
    return placeholder;
  }
  if (Array.isArray(value) && !value.length) {
    return placeholder;
  }
  return value;
};

/**
 * 按内置分隔符切割IP文本
 * @param raw 原始文本
 * @returns 切割后的列表
 */
const splitIP = (raw: string): string[] => {
  const list: string[] = [];
  raw
    .trim()
    .split(/\n|;|；|,|，|\|/)
    .forEach((text) => {
      const ip = text.trim();
      ip.length && list.push(ip);
    });
  return list;
};

/**
 * 从文本中解析出IP地址
 * @param text IP文本
 * @returns IPv4与IPv6地址列表
 */
const parseIP = (text: string) => {
  const list = splitIP(text);
  const IPv4List: string[] = [];
  const IPv6List: string[] = [];

  list.forEach((text) => {
    if (isIP(text, 4)) {
      IPv4List.push(text);
    } else if (isIP(text, 6)) {
      IPv6List.push(text);
    }
  });

  return {
    IPv4List,
    IPv6List,
  };
};

// 将值进行btoa编码
const encodeValueByBtoa = (v: any) => btoa(encodeURIComponent(JSON.stringify(v)));
// 获取atob解码后的值
const decodeValueByAtob = (v: string) => JSON.parse(decodeURIComponent(atob(v)));

/**
 * 从文本（单个IP、CIDR 网段、连续地址段）中解析出IP地址和备注
 * @param text 单个IP、CIDR 网段、连续地址段的IP文本
 * @returns IP地址列表
 */
const analysisIP = (text: string): AddressDescription[] => {
  const list: AddressDescription[] = [];
  // 通过换行符来分割字符串
  const lines = text.split('\n');
  // 判断每一行的情况（单个IP、CIDR 网段、连续地址段）
  lines.forEach((text) => {
    // 剔除备注
    const parts = text.split(/\s+/);
    const description = parts.length >= 2 ? parts.slice(1).join(' ') : '';
    if (isSingleIP(parts[0]) || isCIDR(parts[0]) || isRange(parts[0])) {
      // 1. 单个IP    2. CIDR 网段     // 3. 连续地址段
      list.push({ address: parts[0], description });
    }
  });
  return list;
};

const isIpsValid = (text: string) => {
  // 全部行数
  const lines = text.split('\n').filter((element) => element !== '');
  if (lines.length > analysisIP(text).length) {
    return false;
  }
  return true;
};
// 判断是否为单个IP
const isSingleIP = (ip: string) => {
  return isIP(ip, 4) || isIP(ip, 6);
};
// 判断是否为CIDR网段
const isCIDR = (cidr: string) => {
  const parts = cidr.split('/');
  if (parts.length !== 2) {
    return false;
  }
  const [ip, prefix] = parts;
  if (isIP(ip, 4)) {
    const prefixNum = parseInt(prefix, 10);
    return prefixNum >= 0 && prefixNum <= 32;
  }
  if (isIP(ip, 6)) {
    const prefixNum = parseInt(prefix, 10);
    return prefixNum >= 0 && prefixNum <= 128;
  }
  return false;
};
// 判断是否为连续地址段
const isRange = (range: string) => {
  const parts = range.split('-');
  if (parts.length !== 2) {
    return false;
  }
  const [startIP, endIP] = parts;
  return (isIP(startIP, 4) && isIP(endIP, 4)) || (isIP(startIP, 6) && isIP(endIP, 6));
};

/**
 * 从文本（单个端口、多个离散端口、连续端口、所有端口）中解析出协议端口和备注
 * @param text 单个端口、多个离散端口、连续端口、所有端口的协议端口文本
 * @returns 协议端口列表
 */
const analysisPort = (port: string) => {
  // 判断是否为合法端口
  function isPortNumber(port: string) {
    // 使用正则表达式检查字符串是否只包含数字
    const isNumeric = /^\d+$/.test(port);
    if (!isNumeric) {
      return false;
    }
    const portNumber = parseInt(port, 10);
    return !isNaN(portNumber) && portNumber > 0 && portNumber <= 65535;
  }
  // 判断是否为多个离散端口方案
  function isDispersedPort(port: string) {
    if (!port.includes(',')) return false;
    const ports = port.split(',');
    return ports.every(isPortNumber);
  }
  // 判断是否为连续端口方案
  function isContinuityPort(port: string) {
    const rangeParts = port.split('-');
    if (rangeParts.length !== 2) {
      // 端口范围应该只有两个部分
      return false;
    }
    return rangeParts.every(isPortNumber);
  }

  const list: AddressDescription[] = [];
  const protocolArray = ['tcp', 'TCP', 'UDP', 'udp'];
  const protocolSpecial = ['ICMP', 'icmp', 'GRE', 'gre'];
  // 通过换行符来分割字符串
  const lines = port.split('\n');
  lines.forEach((text) => {
    // 剔除备注
    const parts = text.split(/\s+/);
    const description = parts.length >= 2 ? parts.slice(1).join(' ') : '';
    const portArr = parts[0].trim().split(':');
    if (portArr.length === 2) {
      const [protocol, port] = portArr;
      if (protocolArray.includes(protocol)) {
        if (isPortNumber(port) || ['all', 'ALL'].includes(port) || isDispersedPort(port) || isContinuityPort(port)) {
          // 1. 单个端口   // 2. 多个离散端口  // 3. 连续端口  // 4. 所有端口
          list.push({
            address: parts[0],
            description,
          });
        }
      }
    } else {
      const [protocol] = portArr;
      if (protocolSpecial.includes(protocol)) {
        list.push({
          address: parts[0],
          description,
        });
      }
    }
  });
  return list;
};
const isPortValid = (text: string) => {
  // 全部行数
  const lines = text.split('\n').filter((element) => element !== '');

  if (lines.length > analysisPort(text).length) {
    return false;
  }
  return true;
};

const formatDisplayNumber = (value: string | number, fractionDigits = 2) => {
  if (typeof value === 'string' && isNaN(Number(value))) return value;
  if (value === undefined || value === null) return '--';
  return Number(Number(value).toFixed(fractionDigits));
};

export {
  getAuthSignByBusinessId,
  getInstVip,
  getPrivateIPs,
  getPublicIPs,
  exportTableToExcel,
  getEntirePath,
  cleanPayload,
  getDate,
  getDisplayText,
  splitIP,
  parseIP,
  encodeValueByBtoa,
  decodeValueByAtob,
  analysisIP,
  analysisPort,
  isIpsValid,
  isPortValid,
  formatDisplayNumber,
};
