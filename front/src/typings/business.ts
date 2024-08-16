export interface BusinessFormFilter {
  vendor: string;
  account_id: string | number;
  region: string | number;
}

export enum EipStatus {
  BIND = 'BIND',
  ACTIVE = 'ACTIVE',
  ELB = 'ELB',
  VPN = 'VPN',
  IN_USE = 'IN_USE',
  UNBIND = 'UNBIND',
  BIND_ERROR = 'BIND_ERROR',
  DOWN = 'DOWN',
  RESERVED = 'RESERVED',
  ERROR = 'ERROR',
}

export interface IEip {
  account_id: string;
  bk_biz_id: number;
  cloud_id: string;
  created_at: string;
  creator: string;
  id: string;
  instance_id: string;
  name: string;
  public_ip: string;
  region: string;
  reviser: string;
  status: EipStatus;
  updated_at: string;
  vendor: string;
  cvm_id?: string;
}
export interface AddressDescription {
  address: string;
  description: string;
}
