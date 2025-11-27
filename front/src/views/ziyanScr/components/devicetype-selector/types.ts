interface CvmDeviceTypeReqParams {
  require_type?: number | number[];
  region?: string | string[];
  zone?: string | string[];
  device_group?: string[];
  device_size?: string;
  cpu?: number | string;
  mem?: number | string;
  disk?: number | string;
  enable_capacity?: boolean | string;
  enable_apply?: boolean | string;
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
  cpu_amount: number; // cpu核数
  device_group: string; // 机型族
  ram_amount: number; // 内容容量
}
export type CvmDeviceTypeList = Array<CvmDeviceType>;

// 物理机
export interface IdcpmDeviceType {
  id: number;
  device_type: string;
  cpu: number;
  mem: number;
  raid: string;
  network: string;
  remark: string;
  label: object;
}
export type IdcpmDeviceTypeList = Array<IdcpmDeviceType>;

export type OptionsType = { cvm: CvmDeviceTypeList; idcpm: IdcpmDeviceTypeList };

export type DeviceType = CvmDeviceType | IdcpmDeviceType;
type DeviceTypeList = CvmDeviceTypeList | IdcpmDeviceTypeList;
export type SelectionType = DeviceType | DeviceTypeList;
