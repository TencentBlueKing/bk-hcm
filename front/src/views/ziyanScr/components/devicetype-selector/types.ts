interface CvmDeviceTypeReqParams {
  vendor?: string;
  region?: string | string[];
  zone?: string | string[];
  device_family?: string[]; // 原 device_group
  core_type?: string; // 核心类型，枚举值：小核心、中核心、大核心
  cpu?: number | string;
  mem?: number | string;
  disk?: number | string;
  disable?: boolean; // 原 enable_apply，改为 disable，默认值 false
  technical_class?: string;
}

export interface IProps {
  resourceType: 'cvm' | 'idcpm';
  params: CvmDeviceTypeReqParams;
  multiple?: boolean;
  disabled?: boolean;
  isLoading?: boolean;
  optionDisabled?: (option: DeviceType) => boolean;
  optionDisabledTipsContent?: (option: DeviceType) => string;
  placeholder?: string;
  sort?: (a: DeviceType, b: DeviceType) => number;
  editable?: boolean;
}

// 云主机
export interface CvmDeviceType {
  device_type: string; // 机型
  device_type_class: 'SpecialType' | 'CommonType'; // 通/专用机型，SpecialType专用，CommonType通用
  cpu_core: number; // cpu核数（原 cpu_amount）
  device_family: string; // 机型族（原 device_group）
  memory: number; // 内存容量（原 ram_amount）
  core_type: string; // 核心类型，枚举值：小核心、中核心、大核心
  technical_class: string; // 技术分类
  device_class?: string; // 机型分类
}
export type CvmDeviceTypeList = Array<CvmDeviceType>;

// 物理机
export interface IdcpmDeviceType {
  id: string; // 改为字符串类型
  device_type: string;
  cpu_core: number; // 原 cpu
  memory: number; // 原 mem
  raid: string;
}
export type IdcpmDeviceTypeList = Array<IdcpmDeviceType>;

export type OptionsType = { cvm: CvmDeviceTypeList; idcpm: IdcpmDeviceTypeList };

export type DeviceType = CvmDeviceType | IdcpmDeviceType;
type DeviceTypeList = CvmDeviceTypeList | IdcpmDeviceTypeList;
export type SelectionType = DeviceType | DeviceTypeList;
