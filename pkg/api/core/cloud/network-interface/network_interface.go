/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package networkinterface

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// NetworkInterface defines network interface info.
type NetworkInterface[T NetworkInterfaceExtension] struct {
	BaseNetworkInterface `json:",inline"`
	Extension            *T `json:"extension"`
}

// GetID ...
func (ni NetworkInterface[T]) GetID() string {
	return ni.ID
}

// GetCloudID ...
func (ni NetworkInterface[T]) GetCloudID() string {
	return ni.CloudID
}

// NetworkInterfaceAssociate defines network interface associate info.
type NetworkInterfaceAssociate struct {
	BaseNetworkInterface `json:",inline"`
	CvmID                string `json:"cvm_id"`
	RelCreator           string `json:"rel_creator"`
	RelCreatedAt         string `json:"rel_created_at"`
}

// BaseNetworkInterface define network interface.
type BaseNetworkInterface struct {
	ID             string        `json:"id"`
	Vendor         enumor.Vendor `json:"vendor"`
	Name           string        `json:"name"`
	AccountID      string        `json:"account_id"`
	Region         string        `json:"region"`
	Zone           string        `json:"zone"`
	CloudID        string        `json:"cloud_id"`
	VpcID          string        `json:"vpc_id"`
	CloudVpcID     string        `json:"cloud_vpc_id"`
	SubnetID       string        `json:"subnet_id"`
	CloudSubnetID  string        `json:"cloud_subnet_id"`
	PrivateIPv4    []string      `json:"private_ipv4"`
	PrivateIPv6    []string      `json:"private_ipv6"`
	PublicIPv4     []string      `json:"public_ipv4"`
	PublicIPv6     []string      `json:"public_ipv6"`
	BkBizID        int64         `json:"bk_biz_id"`
	InstanceID     string        `json:"instance_id"`
	*core.Revision `json:",inline"`
}

// NetworkInterfaceExtension defines network interface extensional info.
type NetworkInterfaceExtension interface {
	AzureNIExtension | HuaWeiNIExtension | GcpNIExtension
}

// AzureNIExtension defines azure network interface extensional info.
type AzureNIExtension struct {
	// Type 网络接口类型
	Type *string `json:"type,omitempty"`
	// ResourceGroupName 资源组名称
	ResourceGroupName string `json:"resource_group_name,omitempty"`
	// MacAddress Mac地址
	MacAddress *string `json:"mac_address,omitempty"`
	// EnableAcceleratedNetworking 是否加速网络
	EnableAcceleratedNetworking *bool `json:"enable_accelerated_networking,omitempty"`
	// EnableIPForwarding 是否允许IP转发
	EnableIPForwarding *bool `json:"enable_ip_forwarding,omitempty"`
	// DNSSettings DNS设置
	DNSSettings *InterfaceDNSSettings `json:"dns_settings,omitempty"`
	// CloudGatewayLoadBalancerID 网关负载均衡器ID
	CloudGatewayLoadBalancerID *string `json:"cloud_gateway_load_balancer_id,omitempty"`
	// CloudSecurityGroupID 网络安全组ID
	CloudSecurityGroupID *string `json:"cloud_security_group_id,omitempty"`
	SecurityGroupID      *string `json:"security_group_id,omitempty"`
	// IPConfigurations IP配置列表
	IPConfigurations []*InterfaceIPConfiguration `json:"ip_configurations,omitempty"`
	// CloudVirtualMachineID 虚拟机
	CloudVirtualMachineID *string `json:"cloud_virtual_machine_id,omitempty"`
}

// InterfaceDNSSettings - DNS settings of a network interface.
type InterfaceDNSSettings struct {
	// List of DNS servers IP addresses. Use 'AzureProvidedDNS' to switch to azure provided DNS resolution.
	//'AzureProvidedDNS' value cannot be combined with other IPs, it must be the only value in dnsServers
	// collection.
	DNSServers []*string `json:"dns_servers,omitempty"`

	// READ-ONLY; If the VM that uses this NIC is part of an Availability Set,
	//then this list will have the union of all DNS servers
	// from all NICs that are part of the Availability Set. This property is what is
	// configured on each of those VMs.
	AppliedDNSServers []*string `json:"applied_dns_servers,omitempty" azure:"ro"`
}

// InterfaceIPConfiguration - IPConfiguration in a network interface.
type InterfaceIPConfiguration struct {
	// Resource ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	// Network interface IP configuration properties.
	Properties *InterfaceIPConfigurationPropertiesFormat `json:"properties,omitempty"`

	// Resource type.
	Type *string `json:"type,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`
}

// InterfaceIPConfigurationPropertiesFormat - Properties of IP configuration.
type InterfaceIPConfigurationPropertiesFormat struct {
	// The reference to gateway load balancer frontend IP.
	CloudGatewayLoadBalancerID *string `json:"cloud_gateway_load_balancer_id,omitempty"`

	// Whether this is a primary customer address on the network interface.
	Primary *bool `json:"primary,omitempty"`

	// Private IP address of the IP configuration.
	PrivateIPAddress *string `json:"private_ip_address,omitempty"`

	// Whether the specific IP configuration is IPv4 or IPv6. Default is IPv4.
	PrivateIPAddressVersion *IPVersion `json:"private_ip_address_version,omitempty"`

	// The private IP address allocation method.
	PrivateIPAllocationMethod *IPAllocationMethod `json:"private_ip_allocation_method,omitempty"`

	// Public IP address bound to the IP configuration.
	PublicIPAddress *PublicIPAddress `json:"public_ip_address,omitempty"`

	// CloudSubnetID bound to the IP configuration.
	CloudSubnetID *string `json:"cloud_subnet_id,omitempty"`
}

// IPVersion - IP address version.
type IPVersion string

// IPAllocationMethod - IP address allocation method.
type IPAllocationMethod string

// ProvisioningState - The current provisioning state.
type ProvisioningState string

// PublicIPAddress - Public IP address resource.
type PublicIPAddress struct {
	// Resource ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// Resource location.
	Location *string `json:"location,omitempty"`

	// Public IP address properties.
	Properties *PublicIPAddressPropertiesFormat `json:"properties,omitempty"`

	// Resource tags.
	Tags map[string]*string `json:"tags,omitempty"`

	// A list of availability zones denoting the IP allocated for the resource needs to come from.
	Zones []*string `json:"zones,omitempty"`

	// READ-ONLY; Resource name.
	Name *string `json:"name,omitempty" azure:"ro"`

	// READ-ONLY; Resource type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// PublicIPAddressPropertiesFormat - Public IP address properties.
type PublicIPAddressPropertiesFormat struct {
	// The IP address associated with the public IP address resource.
	IPAddress *string `json:"ip_address,omitempty"`

	// The public IP address version.
	PublicIPAddressVersion *IPVersion `json:"public_ip_address_version,omitempty"`

	// The public IP address allocation method.
	PublicIPAllocationMethod *IPAllocationMethod `json:"public_ip_allocation_method,omitempty"`

	// The Public IP Prefix this Public IP Address should be allocated from.
	CloudPublicIPPrefixID *string `json:"cloud_public_ip_prefix_id,omitempty"`

	// READ-ONLY; The resource GUID property of the public IP address resource.
	ResourceGUID *string `json:"resource_guid,omitempty" azure:"ro"`
}

// NatGatewaySKUName - ImportMode of Nat Gateway SKU.
type NatGatewaySKUName string

// NatGateway - Nat Gateway resource.
type NatGateway struct {
	// Resource ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// Resource location.
	Location *string `json:"location,omitempty"`

	// Nat Gateway properties.
	Properties *NatGatewayPropertiesFormat `json:"properties,omitempty"`

	// The nat gateway SKU.
	SKU *NatGatewaySKU `json:"sku,omitempty"`

	// Resource tags.
	Tags map[string]*string `json:"tags,omitempty"`

	// A list of availability zones denoting the zone in which Nat Gateway should be deployed.
	Zones []*string `json:"zones,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`

	// READ-ONLY; Resource name.
	Name *string `json:"name,omitempty" azure:"ro"`

	// READ-ONLY; Resource type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// NatGatewayPropertiesFormat - Nat Gateway properties.
type NatGatewayPropertiesFormat struct {
	// The idle timeout of the nat gateway.
	IdleTimeoutInMinutes *int32 `json:"idle_timeout_in_minutes,omitempty"`

	// An array of public ip addresses associated with the nat gateway resource.
	CloudPublicIPAddressesID []*string `json:"cloud_public_ip_addresses_id,omitempty"`

	// An array of public ip prefixes associated with the nat gateway resource.
	CloudPublicIPPrefixesID []*string `json:"cloud_public_ip_prefixes_id,omitempty"`

	// READ-ONLY; The provisioning state of the NAT gateway resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`

	// READ-ONLY; The resource GUID property of the NAT gateway resource.
	ResourceGUID *string `json:"resource_guid,omitempty" azure:"ro"`

	// READ-ONLY; An array of references to the subnets using this nat gateway resource.
	CloudSubnetIDs []*string `json:"cloud_subnet_ids,omitempty" azure:"ro"`
}

// NatGatewaySKU - SKU of nat gateway.
type NatGatewaySKU struct {
	// Name of Nat Gateway SKU.
	Name *NatGatewaySKUName `json:"name,omitempty"`
}

// SubnetPropertiesFormat - Properties of the subnet.
type SubnetPropertiesFormat struct {
	// The address prefix for the subnet.
	AddressPrefix *string `json:"address_prefix,omitempty"`

	// List of address prefixes for the subnet.
	AddressPrefixes []*string `json:"address_prefixes,omitempty"`

	// Application gateway IP configurations of virtual network resource.
	ApplicationGatewayIPConfigurations []*ApplicationGatewayIPConfiguration `json:"application_gateway_ip_configurations,omitempty"`

	// An array of references to the delegations on the subnet.
	Delegations []*Delegation `json:"delegations,omitempty"`

	// CloudIPAllocationsID Array of IpAllocation which reference this subnet.
	CloudIPAllocationIDs []*string `json:"cloud_ip_allocation_ids,omitempty"`

	// CloudNatGatewayID Nat gateway associated with this subnet.
	CloudNatGatewayID *string `json:"cloud_nat_gateway_id,omitempty"`

	// The reference to the CloudSecurityGroupID resource.
	CloudSecurityGroupID *string `json:"cloud_security_group_id,omitempty"`

	// Enable or Disable apply network policies on private end point in the subnet.
	PrivateEndpointNetworkPolicies *VirtualNetworkPrivateEndpointNetworkPolicies `json:"private_endpoint_network_policies,omitempty"`

	// Enable or Disable apply network policies on private link service in the subnet.
	PrivateLinkServiceNetworkPolicies *VirtualNetworkPrivateLinkServiceNetworkPolicies `json:"private_link_service_network_policies,omitempty"`

	// The reference to the RouteTable resource.
	RouteTable *RouteTable `json:"route_table,omitempty"`

	// An array of service endpoint policies.
	ServiceEndpointPolicies []*ServiceEndpointPolicy `json:"service_endpoint_policies,omitempty"`

	// An array of service endpoints.
	ServiceEndpoints []*ServiceEndpointPropertiesFormat `json:"service_endpoints,omitempty"`

	// READ-ONLY; Array of IP configuration profiles which reference this subnet.
	IPConfigurationProfiles []*IPConfigurationProfile `json:"ip_configuration_profiles,omitempty" azure:"ro"`

	// READ-ONLY; An array of references to the network interface IP configurations using subnet.
	IPConfigurations []*IPConfiguration `json:"ip_configurations,omitempty" azure:"ro"`

	// READ-ONLY; An array of references to private endpoints.
	PrivateEndpoints []*PrivateEndpoint `json:"private_endpoints,omitempty" azure:"ro"`

	// READ-ONLY; The provisioning state of the subnet resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`

	// READ-ONLY; A read-only string identifying the intention of use for this subnet based on
	// delegations and other user-defined
	// properties.
	Purpose *string `json:"purpose,omitempty" azure:"ro"`

	// READ-ONLY; An array of references to the external resources using subnet.
	ResourceNavigationLinks []*ResourceNavigationLink `json:"resource_navigation_links,omitempty" azure:"ro"`

	// READ-ONLY; An array of references to services injecting into this subnet.
	ServiceAssociationLinks []*ServiceAssociationLink `json:"service_association_links,omitempty" azure:"ro"`
}

// ResourceNavigationLink resource.
type ResourceNavigationLink struct {
	// Resource ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// Name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	// Resource navigation link properties format.
	Properties *ResourceNavigationLinkFormat `json:"properties,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`

	// READ-ONLY; Resource type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// ResourceNavigationLinkFormat - Properties of ResourceNavigationLink.
type ResourceNavigationLinkFormat struct {
	// Link to the external resource.
	Link *string `json:"link,omitempty"`

	// Resource type of the linked resource.
	LinkedResourceType *string `json:"linked_resource_type,omitempty"`

	// READ-ONLY; The provisioning state of the resource navigation link resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// ServiceAssociationLink resource.
type ServiceAssociationLink struct {
	// Resource ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// Name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	// Resource navigation link properties format.
	Properties *ServiceAssociationLinkPropertiesFormat `json:"properties,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`

	// READ-ONLY; Resource type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// ServiceAssociationLinkPropertiesFormat - Properties of ServiceAssociationLink.
type ServiceAssociationLinkPropertiesFormat struct {
	// If true, the resource can be deleted.
	AllowDelete *bool `json:"allow_delete,omitempty"`

	// Link to the external resource.
	Link *string `json:"link,omitempty"`

	// Resource type of the linked resource.
	LinkedResourceType *string `json:"linked_resource_type,omitempty"`

	// A list of locations.
	Locations []*string `json:"locations,omitempty"`

	// READ-ONLY; The provisioning state of the service association link resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// ApplicationGatewayIPConfiguration - IP configuration of an application gateway.
// Currently, 1 public and 1 private IP configuration
// is allowed.
type ApplicationGatewayIPConfiguration struct {
	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// Name of the IP configuration that is unique within an Application Gateway.
	Name *string `json:"name,omitempty"`

	// Properties of the application gateway IP configuration.
	Properties *ApplicationGatewayIPConfigurationPropertiesFormat `json:"properties,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`

	// READ-ONLY; Type of the resource.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// ApplicationGatewayIPConfigurationPropertiesFormat - Properties of IP configuration of an application gateway.
type ApplicationGatewayIPConfigurationPropertiesFormat struct {
	// Reference to the subnet resource. A subnet from where application gateway gets its private address.
	CloudSubnetID *string `json:"cloud_subnet_id,omitempty"`

	// READ-ONLY; The provisioning state of the application gateway IP configuration resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// Delegation - UsageBizInfos the service to which the subnet is delegated.
type Delegation struct {
	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// The name of the resource that is unique within a subnet. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	// Properties of the subnet.
	Properties *ServiceDelegationPropertiesFormat `json:"properties,omitempty"`

	// Resource type.
	Type *string `json:"type,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`
}

// ServiceDelegationPropertiesFormat - Properties of a service delegation.
type ServiceDelegationPropertiesFormat struct {
	// The name of the service to whom the subnet should be delegated (e.g. Microsoft.Sql/servers).
	ServiceName *string `json:"service_name,omitempty"`

	// READ-ONLY; The actions permitted to the service upon delegation.
	Actions []*string `json:"actions,omitempty" azure:"ro"`

	// READ-ONLY; The provisioning state of the service delegation resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// VirtualNetworkPrivateEndpointNetworkPolicies - Enable or Disable apply network policies on private end point
// in the subnet.
type VirtualNetworkPrivateEndpointNetworkPolicies string

// VirtualNetworkPrivateLinkServiceNetworkPolicies - Enable or Disable apply network policies on private link service
// in the subnet.
type VirtualNetworkPrivateLinkServiceNetworkPolicies string

// ServiceEndpointPolicy - Service End point policy resource.
type ServiceEndpointPolicy struct {
	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// Resource location.
	Location *string `json:"location,omitempty"`

	// Properties of the service end point policy.
	Properties *ServiceEndpointPolicyPropertiesFormat `json:"properties,omitempty"`

	// Resource tags.
	Tags map[string]*string `json:"tags,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`

	// READ-ONLY; Kind of service endpoint policy. This is metadata used for the Azure portal experience.
	Kind *string `json:"kind,omitempty" azure:"ro"`

	// READ-ONLY; Resource name.
	Name *string `json:"name,omitempty" azure:"ro"`

	// READ-ONLY; Resource type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// ServiceEndpointPropertiesFormat - The service endpoint properties.
type ServiceEndpointPropertiesFormat struct {
	// A list of locations.
	Locations []*string `json:"locations,omitempty"`

	// The type of the endpoint service.
	Service *string `json:"service,omitempty"`

	// READ-ONLY; The provisioning state of the service endpoint resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// IPConfigurationProfile - IP configuration profile child resource.
type IPConfigurationProfile struct {
	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// The name of the resource. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	// Properties of the IP configuration profile.
	Properties *IPConfigurationProfilePropertiesFormat `json:"properties,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`

	// READ-ONLY; Sub Resource type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// IPConfiguration - IP configuration.
type IPConfiguration struct {
	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	// Properties of the IP configuration.
	Properties *IPConfigurationPropertiesFormat `json:"properties,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`
}

// IPConfigurationPropertiesFormat - Properties of IP configuration.
type IPConfigurationPropertiesFormat struct {
	// The private IP address of the IP configuration.
	PrivateIPAddress *string `json:"private_ip_address,omitempty"`

	// The private IP address allocation method.
	PrivateIPAllocationMethod *IPAllocationMethod `json:"private_ip_allocation_method,omitempty"`

	// The reference to the public IP resource.
	PublicIPAddress *PublicIPAddress `json:"public_ip_address,omitempty"`

	// The reference to the subnet resource.
	CloudSubnetID *string `json:"cloud_subnet_id,omitempty"`

	// READ-ONLY; The provisioning state of the IP configuration resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// PrivateEndpoint - Private endpoint resource.
type PrivateEndpoint struct {
	// The extended location of the load balancer.
	ExtendedLocation *ExtendedLocation `json:"extended_location,omitempty"`

	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// Resource location.
	Location *string `json:"location,omitempty"`

	// Resource tags.
	Tags map[string]*string `json:"tags,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`

	// READ-ONLY; Resource name.
	Name *string `json:"name,omitempty" azure:"ro"`

	// READ-ONLY; Resource type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// ExtendedLocation complex type.
type ExtendedLocation struct {
	// The name of the extended location.
	Name *string `json:"name,omitempty"`

	// The type of the extended location.
	Type *ExtendedLocationTypes `json:"type,omitempty"`
}

// IPConfigurationProfilePropertiesFormat - IP configuration profile properties.
type IPConfigurationProfilePropertiesFormat struct {
	// The reference to the subnet resource to create a container network interface ip configuration.
	CloudSubnetID *string `json:"cloud_subnet_id,omitempty"`

	// READ-ONLY; The provisioning state of the IP configuration profile resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// ExtendedLocationTypes - The supported ExtendedLocation types. Currently only EdgeZone is supported
// in Microsoft.Network
// resources.
type ExtendedLocationTypes string

// ServiceEndpointPolicyPropertiesFormat - Service Endpoint Policy resource.
type ServiceEndpointPolicyPropertiesFormat struct {
	// A collection of contextual service endpoint policy.
	ContextualServiceEndpointPolicies []*string `json:"contextual_service_endpoint_policies,omitempty"`

	// The alias indicating if the policy belongs to a service
	ServiceAlias *string `json:"service_alias,omitempty"`

	// A collection of service endpoint policy definitions of the service endpoint policy.
	ServiceEndpointPolicyDefinitions []*ServiceEndpointPolicyDefinition `json:"service_endpoint_policy_definitions,omitempty"`

	// READ-ONLY; The provisioning state of the service endpoint policy resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`

	// READ-ONLY; The resource GUID property of the service endpoint policy resource.
	ResourceGUID *string `json:"resource_guid,omitempty" azure:"ro"`

	// READ-ONLY; A collection of references to subnets.
	CloudSubnetIDs []*string `json:"cloud_subnet_ids,omitempty" azure:"ro"`
}

// ServiceEndpointPolicyDefinition - Service Endpoint policy definitions.
type ServiceEndpointPolicyDefinition struct {
	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	// Properties of the service endpoint policy definition.
	Properties *ServiceEndpointPolicyDefinitionPropertiesFormat `json:"properties,omitempty"`

	// The type of the resource.
	Type *string `json:"type,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`
}

// ServiceEndpointPolicyDefinitionPropertiesFormat - Service Endpoint policy definition resource.
type ServiceEndpointPolicyDefinitionPropertiesFormat struct {
	// A description for this rule. Restricted to 140 chars.
	Description *string `json:"description,omitempty"`

	// Service endpoint name.
	Service *string `json:"service,omitempty"`

	// A list of service resources.
	ServiceResources []*string `json:"service_resources,omitempty"`

	// READ-ONLY; The provisioning state of the service endpoint policy definition resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// RouteTable - Route table resource.
type RouteTable struct {
	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// Resource location.
	Location *string `json:"location,omitempty"`

	// Properties of the route table.
	Properties *RouteTablePropertiesFormat `json:"properties,omitempty"`

	// Resource tags.
	Tags map[string]*string `json:"tags,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`

	// READ-ONLY; Resource name.
	Name *string `json:"name,omitempty" azure:"ro"`

	// READ-ONLY; Resource type.
	Type *string `json:"type,omitempty" azure:"ro"`
}

// RouteTablePropertiesFormat - Route Table resource.
type RouteTablePropertiesFormat struct {
	// Whether to disable the routes learned by BGP on that route table. True means disable.
	DisableBgpRoutePropagation *bool `json:"disable_bgp_route_propagation,omitempty"`

	// Collection of routes contained within a route table.
	Routes []*Route `json:"routes,omitempty"`

	// READ-ONLY; The provisioning state of the route table resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`

	// READ-ONLY; The resource GUID property of the route table.
	ResourceGUID *string `json:"resource_guid,omitempty" azure:"ro"`

	// READ-ONLY; A collection of references to subnets.
	CloudSubnetIDs []*string `json:"cloud_subnet_ids,omitempty" azure:"ro"`
}

// Route resource.
type Route struct {
	// CloudID Cloud ID.
	CloudID *string `json:"cloud_id,omitempty"`

	// The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`

	// Properties of the route.
	Properties *RoutePropertiesFormat `json:"properties,omitempty"`

	// The type of the resource.
	Type *string `json:"type,omitempty"`

	// READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty" azure:"ro"`
}

// RoutePropertiesFormat - Route resource.
type RoutePropertiesFormat struct {
	// REQUIRED; The type of Azure hop the packet should be sent to.
	NextHopType *RouteNextHopType `json:"next_hop_type,omitempty"`

	// The destination CIDR to which the route applies.
	AddressPrefix *string `json:"address_prefix,omitempty"`

	// A value indicating whether this route overrides overlapping BGP routes regardless of LPM.
	HasBgpOverride *bool `json:"has_bgp_override,omitempty"`

	// The IP address packets should be forwarded to. Next hop values are only allowed in routes
	// where the next hop type is VirtualAppliance.
	NextHopIPAddress *string `json:"next_hop_ip_address,omitempty"`

	// READ-ONLY; The provisioning state of the route resource.
	ProvisioningState *ProvisioningState `json:"provisioning_state,omitempty" azure:"ro"`
}

// RouteNextHopType - The type of Azure hop the packet should be sent to.
type RouteNextHopType string

// AzureNI defines azure network interface.
type AzureNI NetworkInterface[AzureNIExtension]

// HuaWeiNI defines huawei network interface.
type HuaWeiNI NetworkInterface[HuaWeiNIExtension]

// HuaWeiNIExtension defines huawei network interface extensional info.
type HuaWeiNIExtension struct {
	// FixedIps 网卡私网IP信息列表
	FixedIps []ServerInterfaceFixedIp `json:"fixed_ips,omitempty"`
	// MacAddr 网卡Mac地址信息
	MacAddr *string `json:"mac_addr,omitempty"`
	// NetId 网卡端口所属网络ID
	NetId *string `json:"net_id,omitempty"`
	// PortId 网卡端口ID
	PortId *string `json:"port_id,omitempty"`
	// PortState 网卡端口状态
	PortState *string `json:"port_state,omitempty"`
	// DeleteOnTermination 卸载网卡时，是否删除网卡
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty"`
	// DriverMode 从guest os中，网卡的驱动类型。可选值为virtio和hinic，默认为virtio
	DriverMode *string `json:"driver_mode,omitempty"`
	// MinRate 网卡带宽下限
	MinRate *int32 `json:"min_rate,omitempty"`
	// MultiqueueNum 网卡多队列个数
	MultiqueueNum *int32 `json:"multiqueue_num,omitempty"`
	// PciAddress 弹性网卡在Linux GuestOS里的BDF号
	PciAddress *string `json:"pci_address,omitempty"`
	// IpV6 IpV6地址
	IpV6 *string `json:"ipv6,omitempty"`
	// VirtualIPList 虚拟IP地址数组
	VirtualIPList []NetVirtualIP `json:"virtual_ip_list,omitempty"`
	// Addresses 云服务器对应的网络地址信息
	Addresses *EipNetwork `json:"addresses,omitempty"`
	// CloudSecurityGroupIDs 云服务器所属安全组ID
	CloudSecurityGroupIDs []string `json:"cloud_security_group_ids,omitempty"`
}

// EipNetwork 华为云主机绑定的弹性IP
type EipNetwork struct {
	// IPVersion IP地址类型，值为4或6(4：IP地址类型是IPv4 6：IP地址类型是IPv6)
	IPVersion int32 `json:"ip_version,omitempty"`
	// PublicIPAddress IP地址
	PublicIPAddress string `json:"public_ip_address,omitempty"`
	// PublicIPV6Address IPV6地址
	PublicIPV6Address string `json:"public_ipv6_address,omitempty"`
	// BandwidthID 带宽ID
	BandwidthID string `json:"bandwidth_id,omitempty"`
	// BandwidthSize 带宽大小
	BandwidthSize int32 `json:"bandwidth_size,omitempty"`
	// BandwidthType 带宽类型，示例:5_bgp(全动态BGP)
	BandwidthType string `json:"bandwidth_type,omitempty"`
}

// NetVirtualIP 网络接口的虚拟IP
type NetVirtualIP struct {
	// IP 虚拟IP
	IP string `json:"ip,omitempty"`
	// ElasticityIP 弹性公网IP
	ElasticityIP string `json:"elasticity_ip,omitempty"`
}

// ServerInterfaceFixedIp ...
type ServerInterfaceFixedIp struct {
	// IpAddress 网卡私网IP信息。
	IpAddress *string `json:"ip_address,omitempty"`
	// SubnetId 网卡私网IP对应子网信息。
	SubnetId *string `json:"subnet_id,omitempty"`
}

// GcpNIExtension defines gcp network interface extensional info.
type GcpNIExtension struct {
	CanIpForward      bool            `json:"can_ip_forward,omitempty"`
	Status            string          `json:"status,omitempty"`
	StackType         string          `json:"stack_type,omitempty"`
	AccessConfigs     []*AccessConfig `json:"access_configs,omitempty"`
	Ipv6AccessConfigs []*AccessConfig `json:"ipv6_access_configs,omitempty"`
	VpcSelfLink       string          `json:"vpc_self_link,omitempty"`
	SubnetSelfLink    string          `json:"subnet_self_link,omitempty"`
}

// AccessConfig An access configuration attached to an instance's
// network interface. Only one access config per instance is supported.
type AccessConfig struct {
	// Name: The name of this access configuration. The default and
	// recommended name is External NAT, but you can use any arbitrary
	// string, such as My external IP or Network Access.
	Name string `json:"name,omitempty"`

	// NatIP: An external IP address associated with this instance. Specify
	// an unused static external IP address available to the project or
	// leave this field undefined to use an IP from a shared ephemeral IP
	// address pool. If you specify a static external IP address, it must
	// live in the same region as the zone of the instance.
	NatIP string `json:"nat_ip,omitempty"`

	// NetworkTier: This signifies the networking tier used for configuring
	// this access configuration and can only take the following values:
	// PREMIUM, STANDARD. If an AccessConfig is specified without a valid
	// external IP address, an ephemeral IP will be created with this
	// networkTier. If an AccessConfig with a valid external IP address is
	// specified, it must match that of the networkTier associated with the
	// Address resource owning that IP.
	//
	// Possible values:
	//   "FIXED_STANDARD" - Public internet quality with fixed bandwidth.
	//   "PREMIUM" - High quality, Google-grade network tier, support for
	// all networking products.
	//   "STANDARD" - Public internet quality, only limited support for
	// other networking products.
	//   "STANDARD_OVERRIDES_FIXED_STANDARD" - (Output only) Temporary tier
	// for FIXED_STANDARD when fixed standard tier is expired or not
	// configured.
	NetworkTier string `json:"network_tier,omitempty"`

	// Type: The type of configuration. The default and only option is
	// ONE_TO_ONE_NAT.
	//
	// Possible values:
	//   "DIRECT_IPV6"
	//   "ONE_TO_ONE_NAT" (default)
	Type string `json:"type,omitempty"`
}
