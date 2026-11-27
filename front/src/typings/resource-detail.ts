import type { IEip } from './business';
import type { VendorEnum } from '@/common/constant';
import type { ISecurityGroupItem } from '@/store/security-group';

interface ResourceIdentity {
  id: string;
  account_id: string;
  bk_biz_id: number;
  cloud_id: string;
  name: string;
  region: string;
  created_at: string;
}

type VendorByName = {
  tcloud: VendorEnum.TCLOUD;
  aws: VendorEnum.AWS;
  azure: VendorEnum.AZURE;
  gcp: VendorEnum.GCP;
  huawei: VendorEnum.HUAWEI;
  other: VendorEnum.OTHER;
};

type ByVendor<Base, Extensions> = {
  [V in keyof Extensions & keyof VendorByName]: Base & { vendor: VendorByName[V]; extension?: Extensions[V] | null };
}[keyof Extensions & keyof VendorByName];

interface CvmBase extends ResourceIdentity {
  ip?: string;
  status: string;
  recycle_status: string;
  cloud_image_id: string;
  cloud_subnet_ids: string[];
  cloud_vpc_ids: string[];
  cloud_created_time: string;
  cloud_launched_time: string;
  cloud_expired_time: string;
}

interface TCloudCvmExtension {
  disable_api_termination?: boolean;
  cpu?: number;
  memory?: number;
  instance_charge_type?: string;
  internet_accessible?: { internet_charge_type?: string };
  cloud_security_group_ids?: string[];
}

export type Cvm = ByVendor<
  CvmBase,
  {
    tcloud: TCloudCvmExtension;
    aws: { private_dns_name?: string; dns_name?: string; cpu_options?: { core_count?: number } };
    azure: { resource_group_name?: string; zones?: string[]; cloud_network_interface_ids?: string[] };
    gcp: { deletion_protection?: boolean; can_ip_forward?: boolean; cloud_network_interface_ids?: string[] };
    huawei: { flavor?: { vcpus?: string; ram?: string } };
    other: never;
  }
>;

interface DiskBase extends ResourceIdentity {
  instance_id?: string;
  is_system_disk: boolean;
  status: string;
  recycle_status: string;
  zone: string;
  disk_type: string;
  disk_size: number;
}

interface DiskEncryption {
  encrypted?: boolean;
}

export interface AwsDiskAttachment {
  instance_id?: string;
  device_name?: string;
  delete_on_termination?: boolean;
}

export type Disk = ByVendor<
  DiskBase,
  {
    tcloud: DiskEncryption & {
      disk_charge_type: string;
      deadline_time?: string;
      delete_with_instance?: boolean;
      backup_disk?: boolean;
      attached?: boolean;
      instance_id?: string;
    };
    aws: DiskEncryption & { attachment?: AwsDiskAttachment[] };
    azure: DiskEncryption & { resource_group_name: string; sku_name?: string; zones?: string[] };
    gcp: DiskEncryption & { self_link: string; source_image?: string; description?: string };
    huawei: DiskEncryption & { charge_type?: string; expire_time?: string };
  }
>;

export type AttachedDisk = Disk & { cvm_id: string };

interface NetworkInterfaceBase extends ResourceIdentity {
  instance_id: string;
  vpc_id: string;
  subnet_id: string;
  cloud_vpc_id: string;
  cloud_subnet_id: string;
  private_ipv4: string[] | null;
  private_ipv6: string[] | null;
  public_ipv4: string[] | null;
  public_ipv6: string[] | null;
  virtualNetworkSubnetId?: string;
}

interface AzureIpConfiguration {
  name?: string;
  properties?: {
    private_ip_address?: string;
    private_ip_address_version?: string;
    primary?: boolean;
    public_ip_address?: { properties?: { ip_address?: string } };
  };
}

export type NetworkInterface = ByVendor<
  NetworkInterfaceBase,
  {
    azure: {
      type?: string;
      resource_group_name?: string;
      mac_address?: string;
      enable_accelerated_networking?: boolean;
      enable_ip_forwarding?: boolean;
      cloud_gateway_load_balancer_id?: string;
      cloud_security_group_id?: string;
      dns_settings?: { dns_servers?: string[]; applied_dns_servers?: string[] };
      ip_configurations?: AzureIpConfiguration[];
    };
    gcp: {
      can_ip_forward?: boolean;
      access_configs?: { name?: string; nat_ip?: string; network_tier?: string; type?: string }[];
    };
    huawei: {
      port_state?: string;
      mac_addr?: string;
      cloud_security_group_ids?: string[];
      virtual_ip_list?: { ip?: string; elasticity_ip?: string }[];
    };
  }
>;

interface VpcCidr {
  type: string;
  cidr: string;
}

export type Vpc = ByVendor<
  ResourceIdentity,
  {
    tcloud: {
      cidr: (VpcCidr & { category: string })[];
      is_default: boolean;
      dns_server_set: string[];
      domain_name?: string;
      enable_multicast: boolean;
    };
    aws: {
      cidr: (VpcCidr & { address_pool: string; state: string })[];
      state: string;
      enable_dns_hostnames: boolean;
      enable_dns_support: boolean;
      instance_tenancy: string;
      is_default: boolean;
    };
    azure: { cidr: VpcCidr[]; resource_group_name: string; dns_servers: string[] };
    gcp: {
      auto_create_subnetworks: boolean;
      routing_mode: string;
      enable_ula_internal_ipv6: boolean;
      mtu: number;
      cidr?: never;
    };
    huawei: { cidr: VpcCidr[]; status: string };
  }
>;

export type Subnet = ByVendor<
  ResourceIdentity & { ipv4_nums?: number },
  {
    tcloud: { is_default: boolean; cloud_network_acl_id?: string };
    aws: {
      state: string;
      map_public_ip_on_launch: boolean;
      is_default: boolean;
      assign_ipv6_address_on_creation: boolean;
      hostname_type: string;
    };
    azure: { nat_gateway?: string; cloud_security_group_id?: string };
    gcp: {
      stack_type: string;
      ipv6_access_type: string;
      gateway_address: string;
      private_ip_google_access: boolean;
      enable_flow_logs: boolean;
    };
    huawei: { status: string; dhcp_enable: boolean; gateway_ip: string; dns_list: string[]; ntp_addresses: string[] };
  }
>;

interface EipBase extends Omit<IEip, 'vendor'> {
  instance_type?: string;
  address_type?: string;
}

export type Eip = ByVendor<
  EipBase,
  {
    tcloud: { internet_service_provider?: string; bandwidth?: number; internet_charge_type?: string };
    aws: { domain?: string; public_ipv4_pool?: string; private_ip_address?: string; network_interface_id?: string };
    azure: { resource_group_name?: string; location?: string; sku?: string; sku_tier?: string; fqdn?: string };
    gcp: { address_type?: string; network_tier?: string };
    huawei: {
      enterprise_project_id?: string;
      type?: string;
      bandwidth_name?: string;
      bandwidth_id?: string;
      bandwidth_size?: number;
      bandwidth_share_type?: string;
    };
  }
>;

export type SecurityGroup = ByVendor<
  Omit<ISecurityGroupItem, 'vendor'> & {
    cvm_count?: number;
    network_interface_count?: number;
    subnet_count?: number;
  },
  {
    tcloud: { cloud_project_id?: string };
    aws: { vpc_id: string; cloud_vpc_id?: string; cloud_owner_id?: string };
    azure: { resource_group_name: string; etag?: string; flush_connection?: boolean; resource_guid?: string };
    huawei: { cloud_project_id: string; cloud_enterprise_project_id: string };
  }
>;

export type RouteTable = ByVendor<
  ResourceIdentity,
  {
    tcloud: { main: boolean };
    aws: { main: boolean };
    azure: { resource_group_name: string; cloud_subscription_id: string };
    huawei: { default: boolean; tenant_id: string };
    gcp: never;
  }
>;

interface TCloudImageExtension {
  region: string;
  image_source: string;
  image_size: number;
}

export type Image = ByVendor<
  ResourceIdentity,
  {
    tcloud: TCloudImageExtension;
    aws: { region: string };
    azure: { region: string };
    gcp: { region: string };
    huawei: { region: string };
  }
>;

export interface ResourceDetailMap {
  cvms: Cvm;
  disks: Disk;
  eips: Eip;
  images: Image;
  network_interfaces: NetworkInterface;
  route_tables: RouteTable;
  security_groups: SecurityGroup;
  subnets: Subnet;
  vpcs: Vpc;
}

// Other endpoints still use the generic resource store.
export type ResourceResponse<T extends string> = (ResourceDetailMap & Record<string, any>)[T];

export interface ResourceListItem {
  spec?: object;
  attachment?: object;
  revision?: object;
}

export interface ResourceDetailLabels {
  vendorName: string;
  bk_biz_id_label: number | string;
  bk_biz_id_name: string;
}

export type VendorResource<T, V extends string> = V extends keyof VendorByName
  ? Extract<T, { vendor: VendorByName[V] }>
  : T;
