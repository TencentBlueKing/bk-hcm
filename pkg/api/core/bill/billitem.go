/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package bill

import (
	rawjson "encoding/json"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
	"github.com/shopspring/decimal"
)

// BaseBillItem 存储分账后的明细
type BaseBillItem struct {
	ID             string              `json:"id,omitempty"`
	RootAccountID  string              `json:"root_account_id"`
	MainAccountID  string              `json:"main_account_id"`
	Vendor         enumor.Vendor       `json:"vendor" validate:"required"`
	ProductID      int64               `json:"product_id" validate:"omitempty"`
	BkBizID        int64               `json:"bk_biz_id" validate:"omitempty"`
	BillYear       int                 `json:"bill_year" validate:"required"`
	BillMonth      int                 `json:"bill_month" validate:"required"`
	BillDay        int                 `json:"bill_day" validate:"required"`
	VersionID      int                 `json:"version_id" validate:"required"`
	Currency       enumor.CurrencyCode `json:"currency" validate:"required"`
	Cost           decimal.Decimal     `json:"cost" validate:"required"`
	HcProductCode  string              `json:"hc_product_code,omitempty"`
	HcProductName  string              `json:"hc_product_name,omitempty"`
	ResAmount      decimal.Decimal     `json:"res_amount,omitempty"`
	ResAmountUnit  string              `json:"res_amount_unit,omitempty"`
	*core.Revision `json:",inline"`
}

// BillItem ...
type BillItem[E BillItemExtension] struct {
	*BaseBillItem `json:",inline"`
	Extension     *E `json:"extension,omitempty"`
}

// BillItemRaw ...
type BillItemRaw struct {
	*BaseBillItem `json:",inline"`
	Extension     rawjson.RawMessage `json:"extension,omitempty"`
}

// TCloudBillItem ...
type TCloudBillItem = BillItem[TCloudBillItemExtension]

// HuaweiBillItem ...
type HuaweiBillItem = BillItem[HuaweiBillItemExtension]

// AzureBillItem ...
type AzureBillItem = BillItem[AzureBillItemExtension]

// AwsBillItem ...
type AwsBillItem = BillItem[AwsBillItemExtension]

// GcpBillItem ...
type GcpBillItem = BillItem[GcpBillItemExtension]

// KaopuBillItem ...
type KaopuBillItem = BillItem[KaopuBillItemExtension]

// ZenlayerBillItem ...
type ZenlayerBillItem = BillItem[ZenlayerBillItemExtension]

// BillItemExtension 账单详情
type BillItemExtension interface {
	TCloudBillItemExtension |
		HuaweiBillItemExtension |
		AwsBillItemExtension |
		AzureBillItemExtension |
		GcpBillItemExtension |
		KaopuBillItemExtension |
		ZenlayerBillItemExtension |
		rawjson.RawMessage
}

// TCloudBillItemExtension ...
type TCloudBillItemExtension struct {
}

// AwsBillItemExtension ...
type AwsBillItemExtension struct {
	*AwsRawBillItem `json:",inline"`
}

// HuaweiBillItemExtension ...
type HuaweiBillItemExtension struct {
	*model.ResFeeRecordV2 `json:",inline"`
}

// GcpRawBillItem bill item from big query
type GcpRawBillItem struct {
	BillingAccountID          string           `json:"billing_account_id"`
	Cost                      *decimal.Decimal `json:"cost"`
	CostType                  *string          `json:"cost_type"`
	Country                   *string          `json:"country"`
	CreditsAmount             *string          `json:"credits_amount"`
	Currency                  *string          `json:"currency"`
	CurrencyConversionRate    *decimal.Decimal `json:"currency_conversion_rate"`
	Location                  *string          `json:"location"`
	Month                     *string          `json:"month"`
	ProjectID                 *string          `json:"project_id"`
	ProjectName               *string          `json:"project_name"`
	ProjectNumber             *string          `json:"project_number"`
	Region                    *string          `json:"region"`
	ResourceGlobalName        *string          `json:"resource_global_name"`
	ResourceName              *string          `json:"resource_name"`
	ServiceDescription        *string          `json:"service_description"`
	ServiceID                 *string          `json:"service_id"`
	SkuDescription            *string          `json:"sku_description"`
	SkuID                     *string          `json:"sku_id"`
	TotalCost                 *decimal.Decimal `json:"total_cost"`
	ReturnCost                *decimal.Decimal `json:"return_cost"`
	UsageAmount               *decimal.Decimal `json:"usage_amount"`
	UsageAmountInPricingUnits *decimal.Decimal `json:"usage_amount_in_pricing_units"`
	UsageEndTime              *string          `json:"usage_end_time,omitempty"`
	UsagePricingUnit          *string          `json:"usage_pricing_unit"`
	UsageStartTime            *string          `json:"usage_start_time,omitempty"`
	UsageUnit                 *string          `json:"usage_unit"`
	Zone                      *string          `json:"zone"`
	CreditInfos               []GcpCredit      `json:"credit_infos,omitempty"`
}

// GcpCredit gcp credit info
type GcpCredit struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Name     string           `json:"name"`
	FullName string           `json:"full_name"`
	Amount   *decimal.Decimal `json:"amount"`
}

func (c *GcpCredit) String() string {
	return fmt.Sprintf("[%s] cost: %s %s/%s/%s ", c.Amount.String(), c.Type, c.ID, c.Name, c.FullName)
}

// GcpBillItemExtension ...
type GcpBillItemExtension struct {
	*GcpRawBillItem `json:",inline"`
}

// AzureBillItemExtension ...
type AzureBillItemExtension struct {
}

// KaopuBillItemExtension ...
type KaopuBillItemExtension struct {
}

// ZenlayerBillItemExtension ...
type ZenlayerBillItemExtension struct {
	*ZenlayerRawBillItem `json:",inline"`
}

// ZenlayerRawBillItem bill item from zenlayer
type ZenlayerRawBillItem struct {
	BillID         *string          `json:"bill_id"`         // 账单ID
	ZenlayerOrder  *string          `json:"zenlayer_order"`  // Zenlayer订单编号
	CID            *string          `json:"cid"`             // CID
	GroupID        *string          `json:"group_id"`        // GROUP ID
	Currency       *string          `json:"currency"`        // 币种
	City           *string          `json:"city"`            // 城市
	PayContent     *string          `json:"pay_content"`     // 付费内容
	Type           *string          `json:"type"`            // 类型
	AcceptanceNum  *decimal.Decimal `json:"acceptance_num"`  // 验收数量
	PayNum         *decimal.Decimal `json:"pay_num"`         // 付费数量
	UnitPriceUSD   *decimal.Decimal `json:"unit_price_usd"`  // 单价USD
	TotalPayable   *decimal.Decimal `json:"total_payable"`   // 应付USD
	BillingPeriod  *string          `json:"billing_period"`  // 账期
	ContractPeriod *string          `json:"contract_period"` // 合约周期
	Remarks        *string          `json:"remarks"`         // 备注
	BusinessGroup  *string          `json:"business_group"`  // 业务组
	CPU            *string          `json:"cpu"`             // CPU
	Disk           *string          `json:"disk"`            // 硬盘
	Memory         *string          `json:"memory"`          // 内存
}

// AwsRawBillItem Aws 原始账单结构
type AwsRawBillItem struct {
	BillBillType                                             string `json:"bill_bill_type,omitempty"`
	BillBillingEntity                                        string `json:"bill_billing_entity,omitempty"`
	BillBillingPeriodEndDate                                 string `json:"bill_billing_period_end_date,omitempty"`
	BillBillingPeriodStartDate                               string `json:"bill_billing_period_start_date,omitempty"`
	BillInvoiceId                                            string `json:"bill_invoice_id,omitempty"`
	BillInvoicingEntity                                      string `json:"bill_invoicing_entity,omitempty"`
	BillPayerAccountId                                       string `json:"bill_payer_account_id,omitempty"`
	DiscountEdpDiscount                                      string `json:"discount_edp_discount,omitempty"`
	DiscountTotalDiscount                                    string `json:"discount_total_discount,omitempty"`
	DiscountPrivateRateDiscount                              string `json:"discount_private_rate_discount,omitempty"`
	IdentityLineItemId                                       string `json:"identity_line_item_id,omitempty"`
	IdentityTimeInterval                                     string `json:"identity_time_interval,omitempty"`
	LineItemAvailabilityZone                                 string `json:"line_item_availability_zone,omitempty"`
	LineItemBlendedCost                                      string `json:"line_item_blended_cost,omitempty"`
	LineItemBlendedRate                                      string `json:"line_item_blended_rate,omitempty"`
	LineItemCurrencyCode                                     string `json:"line_item_currency_code,omitempty"`
	LineItemLegalEntity                                      string `json:"line_item_legal_entity,omitempty"`
	LineItemLineItemDescription                              string `json:"line_item_line_item_description,omitempty"`
	LineItemLineItemType                                     string `json:"line_item_line_item_type,omitempty"`
	LineItemNetUnblendedCost                                 string `json:"line_item_net_unblended_cost,omitempty"`
	LineItemNetUnblendedRate                                 string `json:"line_item_net_unblended_rate,omitempty"`
	LineItemNormalizationFactor                              string `json:"line_item_normalization_factor,omitempty"`
	LineItemNormalizedUsageAmount                            string `json:"line_item_normalized_usage_amount,omitempty"`
	LineItemOperation                                        string `json:"line_item_operation,omitempty"`
	LineItemProductCode                                      string `json:"line_item_product_code,omitempty"`
	LineItemResourceId                                       string `json:"line_item_resource_id,omitempty"`
	LineItemTaxType                                          string `json:"line_item_tax_type,omitempty"`
	LineItemUnblendedCost                                    string `json:"line_item_unblended_cost,omitempty"`
	LineItemUnblendedRate                                    string `json:"line_item_unblended_rate,omitempty"`
	LineItemUsageAccountId                                   string `json:"line_item_usage_account_id,omitempty"`
	LineItemUsageAmount                                      string `json:"line_item_usage_amount,omitempty"`
	LineItemUsageEndDate                                     string `json:"line_item_usage_end_date,omitempty"`
	LineItemUsageStartDate                                   string `json:"line_item_usage_start_date,omitempty"`
	LineItemUsageType                                        string `json:"line_item_usage_type,omitempty"`
	Month                                                    string `json:"month,omitempty"`
	PricingCurrency                                          string `json:"pricing_currency,omitempty"`
	PricingPublicOnDemandCost                                string `json:"pricing_public_on_demand_cost,omitempty"`
	PricingPublicOnDemandRate                                string `json:"pricing_public_on_demand_rate,omitempty"`
	PricingRateCode                                          string `json:"pricing_rate_code,omitempty"`
	PricingRateId                                            string `json:"pricing_rate_id,omitempty"`
	PricingTerm                                              string `json:"pricing_term,omitempty"`
	PricingUnit                                              string `json:"pricing_unit,omitempty"`
	ProductAccessType                                        string `json:"product_access_type,omitempty"`
	ProductAccountAssistance                                 string `json:"product_account_assistance,omitempty"`
	ProductAlarmType                                         string `json:"product_alarm_type,omitempty"`
	ProductAlpha3Countrycode                                 string `json:"product_alpha3countrycode,omitempty"`
	ProductArchitecturalReview                               string `json:"product_architectural_review,omitempty"`
	ProductArchitectureSupport                               string `json:"product_architecture_support,omitempty"`
	ProductAttachmentType                                    string `json:"product_attachment_type,omitempty"`
	ProductAvailability                                      string `json:"product_availability,omitempty"`
	ProductAvailabilityZone                                  string `json:"product_availability_zone,omitempty"`
	ProductBackupservice                                     string `json:"product_backupservice,omitempty"`
	ProductBestPractices                                     string `json:"product_best_practices,omitempty"`
	ProductCacheEngine                                       string `json:"product_cache_engine,omitempty"`
	ProductCapacity                                          string `json:"product_capacity,omitempty"`
	ProductCapacitystatus                                    string `json:"product_capacitystatus,omitempty"`
	ProductCaseSeverityresponseTimes                         string `json:"product_case_severityresponse_times,omitempty"`
	ProductCategory                                          string `json:"product_category,omitempty"`
	ProductCiType                                            string `json:"product_ci_type,omitempty"`
	ProductClassicnetworkingsupport                          string `json:"product_classicnetworkingsupport,omitempty"`
	ProductClientLocation                                    string `json:"product_client_location,omitempty"`
	ProductClockSpeed                                        string `json:"product_clock_speed,omitempty"`
	ProductComponent                                         string `json:"product_component,omitempty"`
	ProductComputeFamily                                     string `json:"product_compute_family,omitempty"`
	ProductComputeType                                       string `json:"product_compute_type,omitempty"`
	ProductConnectionType                                    string `json:"product_connection_type,omitempty"`
	ProductContentType                                       string `json:"product_content_type,omitempty"`
	ProductContinent                                         string `json:"product_continent,omitempty"`
	ProductCountry                                           string `json:"product_country,omitempty"`
	ProductCurrentGeneration                                 string `json:"product_current_generation,omitempty"`
	ProductCustomerServiceAndCommunities                     string `json:"product_customer_service_and_communities,omitempty"`
	ProductDatabaseEngine                                    string `json:"product_database_engine,omitempty"`
	ProductDatatransferout                                   string `json:"product_datatransferout,omitempty"`
	ProductDedicatedEbsThroughput                            string `json:"product_dedicated_ebs_throughput,omitempty"`
	ProductDeploymentOption                                  string `json:"product_deployment_option,omitempty"`
	ProductDescription                                       string `json:"product_description,omitempty"`
	ProductDirectConnectLocation                             string `json:"product_direct_connect_location,omitempty"`
	ProductDominantnondominant                               string `json:"product_dominantnondominant,omitempty"`
	ProductDurability                                        string `json:"product_durability,omitempty"`
	ProductEcu                                               string `json:"product_ecu,omitempty"`
	ProductEndpointType                                      string `json:"product_endpoint_type,omitempty"`
	ProductEngineCode                                        string `json:"product_engine_code,omitempty"`
	ProductEngineMajorVersion                                string `json:"product_engine_major_version,omitempty"`
	ProductEnhancedNetworkingSupported                       string `json:"product_enhanced_networking_supported,omitempty"`
	ProductExtendedSupportPricingYear                        string `json:"product_extended_support_pricing_year,omitempty"`
	ProductFeeDescription                                    string `json:"product_fee_description,omitempty"`
	ProductFreeQueryTypes                                    string `json:"product_free_query_types,omitempty"`
	ProductFreeTier                                          string `json:"product_free_tier,omitempty"`
	ProductFromLocation                                      string `json:"product_from_location,omitempty"`
	ProductFromLocationType                                  string `json:"product_from_location_type,omitempty"`
	ProductFromRegionCode                                    string `json:"product_from_region_code,omitempty"`
	ProductGeoregioncode                                     string `json:"product_georegioncode,omitempty"`
	ProductGpu                                               string `json:"product_gpu,omitempty"`
	ProductGpuMemory                                         string `json:"product_gpu_memory,omitempty"`
	ProductGroup                                             string `json:"product_group,omitempty"`
	ProductGroupDescription                                  string `json:"product_group_description,omitempty"`
	ProductIncludedServices                                  string `json:"product_included_services,omitempty"`
	ProductInsightstype                                      string `json:"product_insightstype,omitempty"`
	ProductInstance                                          string `json:"product_instance,omitempty"`
	ProductInstanceFamily                                    string `json:"product_instance_family,omitempty"`
	ProductInstanceName                                      string `json:"product_instance_name,omitempty"`
	ProductInstanceType                                      string `json:"product_instance_type,omitempty"`
	ProductInstanceTypeFamily                                string `json:"product_instance_type_family,omitempty"`
	ProductIntelAvx2Available                                string `json:"product_intel_avx2_available,omitempty"`
	ProductIntelAvxAvailable                                 string `json:"product_intel_avx_available,omitempty"`
	ProductIntelTurboAvailable                               string `json:"product_intel_turbo_available,omitempty"`
	ProductLaunchSupport                                     string `json:"product_launch_support,omitempty"`
	ProductLicenseModel                                      string `json:"product_license_model,omitempty"`
	ProductLocation                                          string `json:"product_location,omitempty"`
	ProductLocationType                                      string `json:"product_location_type,omitempty"`
	ProductLogsDestination                                   string `json:"product_logs_destination,omitempty"`
	ProductMailboxStorage                                    string `json:"product_mailbox_storage,omitempty"`
	ProductMarketoption                                      string `json:"product_marketoption,omitempty"`
	ProductMaxIopsBurstPerformance                           string `json:"product_max_iops_burst_performance,omitempty"`
	ProductMaxIopsvolume                                     string `json:"product_max_iopsvolume,omitempty"`
	ProductMaxThroughputvolume                               string `json:"product_max_throughputvolume,omitempty"`
	ProductMaxVolumeSize                                     string `json:"product_max_volume_size,omitempty"`
	ProductMemory                                            string `json:"product_memory,omitempty"`
	ProductMemoryGib                                         string `json:"product_memory_gib,omitempty"`
	ProductMessageDeliveryFrequency                          string `json:"product_message_delivery_frequency,omitempty"`
	ProductMessageDeliveryOrder                              string `json:"product_message_delivery_order,omitempty"`
	ProductMinVolumeSize                                     string `json:"product_min_volume_size,omitempty"`
	ProductNetworkPerformance                                string `json:"product_network_performance,omitempty"`
	ProductNormalizationSizeFactor                           string `json:"product_normalization_size_factor,omitempty"`
	ProductOperatingSystem                                   string `json:"product_operating_system,omitempty"`
	ProductOperation                                         string `json:"product_operation,omitempty"`
	ProductOperationsSupport                                 string `json:"product_operations_support,omitempty"`
	ProductOrigin                                            string `json:"product_origin,omitempty"`
	ProductPhysicalCpu                                       string `json:"product_physical_cpu,omitempty"`
	ProductPhysicalGpu                                       string `json:"product_physical_gpu,omitempty"`
	ProductPhysicalProcessor                                 string `json:"product_physical_processor,omitempty"`
	ProductPlatoclassificationtype                           string `json:"product_platoclassificationtype,omitempty"`
	ProductPlatoinstancename                                 string `json:"product_platoinstancename,omitempty"`
	ProductPlatoinstancetype                                 string `json:"product_platoinstancetype,omitempty"`
	ProductPlatopricingtype                                  string `json:"product_platopricingtype,omitempty"`
	ProductPortSpeed                                         string `json:"product_port_speed,omitempty"`
	ProductPreInstalledSw                                    string `json:"product_pre_installed_sw,omitempty"`
	ProductPricingUnit                                       string `json:"product_pricing_unit,omitempty"`
	ProductProactiveGuidance                                 string `json:"product_proactive_guidance,omitempty"`
	ProductProcessorArchitecture                             string `json:"product_processor_architecture,omitempty"`
	ProductProcessorFeatures                                 string `json:"product_processor_features,omitempty"`
	ProductProductFamily                                     string `json:"product_product_family,omitempty"`
	ProductProductName                                       string `json:"product_product_name,omitempty"`
	ProductProgrammaticCaseManagement                        string `json:"product_programmatic_case_management,omitempty"`
	ProductProvisioned                                       string `json:"product_provisioned,omitempty"`
	ProductQueueType                                         string `json:"product_queue_type,omitempty"`
	ProductRecipient                                         string `json:"product_recipient,omitempty"`
	ProductRegion                                            string `json:"product_region,omitempty"`
	ProductRegionCode                                        string `json:"product_region_code,omitempty"`
	ProductRequestDescription                                string `json:"product_request_description,omitempty"`
	ProductRequestType                                       string `json:"product_request_type,omitempty"`
	ProductResourceEndpoint                                  string `json:"product_resource_endpoint,omitempty"`
	ProductResourcePriceGroup                                string `json:"product_resource_price_group,omitempty"`
	ProductResourceType                                      string `json:"product_resource_type,omitempty"`
	ProductRoutingTarget                                     string `json:"product_routing_target,omitempty"`
	ProductRoutingType                                       string `json:"product_routing_type,omitempty"`
	ProductServicecode                                       string `json:"product_servicecode,omitempty"`
	ProductServicename                                       string `json:"product_servicename,omitempty"`
	ProductSku                                               string `json:"product_sku,omitempty"`
	ProductStorage                                           string `json:"product_storage,omitempty"`
	ProductStorageClass                                      string `json:"product_storage_class,omitempty"`
	ProductStorageFamily                                     string `json:"product_storage_family,omitempty"`
	ProductStorageMedia                                      string `json:"product_storage_media,omitempty"`
	ProductStorageType                                       string `json:"product_storage_type,omitempty"`
	ProductTechnicalSupport                                  string `json:"product_technical_support,omitempty"`
	ProductTenancy                                           string `json:"product_tenancy,omitempty"`
	ProductThirdpartySoftwareSupport                         string `json:"product_thirdparty_software_support,omitempty"`
	ProductTickettype                                        string `json:"product_tickettype,omitempty"`
	ProductTiertype                                          string `json:"product_tiertype,omitempty"`
	ProductToLocation                                        string `json:"product_to_location,omitempty"`
	ProductToLocationType                                    string `json:"product_to_location_type,omitempty"`
	ProductToRegionCode                                      string `json:"product_to_region_code,omitempty"`
	ProductTrafficDirection                                  string `json:"product_traffic_direction,omitempty"`
	ProductTraining                                          string `json:"product_training,omitempty"`
	ProductTransferType                                      string `json:"product_transfer_type,omitempty"`
	ProductUsagetype                                         string `json:"product_usagetype,omitempty"`
	ProductVaulttype                                         string `json:"product_vaulttype,omitempty"`
	ProductVcpu                                              string `json:"product_vcpu,omitempty"`
	ProductVersion                                           string `json:"product_version,omitempty"`
	ProductVirtualInterfaceType                              string `json:"product_virtual_interface_type,omitempty"`
	ProductVolumeApiName                                     string `json:"product_volume_api_name,omitempty"`
	ProductVolumeType                                        string `json:"product_volume_type,omitempty"`
	ProductVpcnetworkingsupport                              string `json:"product_vpcnetworkingsupport,omitempty"`
	ProductWhoCanOpenCases                                   string `json:"product_who_can_open_cases,omitempty"`
	ReservationAmortizedUpfrontCostForUsage                  string `json:"reservation_amortized_upfront_cost_for_usage,omitempty"`
	ReservationAmortizedUpfrontFeeForBillingPeriod           string `json:"reservation_amortized_upfront_fee_for_billing_period,omitempty"`
	ReservationEffectiveCost                                 string `json:"reservation_effective_cost,omitempty"`
	ReservationEndTime                                       string `json:"reservation_end_time,omitempty"`
	ReservationModificationStatus                            string `json:"reservation_modification_status,omitempty"`
	ReservationNetAmortizedUpfrontCostForUsage               string `json:"reservation_net_amortized_upfront_cost_for_usage,omitempty"`
	ReservationNetAmortizedUpfrontFeeForBillingPeriod        string `json:"reservation_net_amortized_upfront_fee_for_billing_period,omitempty"`
	ReservationNetEffectiveCost                              string `json:"reservation_net_effective_cost,omitempty"`
	ReservationNetRecurringFeeForUsage                       string `json:"reservation_net_recurring_fee_for_usage,omitempty"`
	ReservationNetUnusedAmortizedUpfrontFeeForBillingPeriod  string `json:"reservation_net_unused_amortized_upfront_fee_for_billing_period,omitempty"`
	ReservationNetUnusedRecurringFee                         string `json:"reservation_net_unused_recurring_fee,omitempty"`
	ReservationNetUpfrontValue                               string `json:"reservation_net_upfront_value,omitempty"`
	ReservationNormalizedUnitsPerReservation                 string `json:"reservation_normalized_units_per_reservation,omitempty"`
	ReservationNumberOfReservations                          string `json:"reservation_number_of_reservations,omitempty"`
	ReservationRecurringFeeForUsage                          string `json:"reservation_recurring_fee_for_usage,omitempty"`
	ReservationStartTime                                     string `json:"reservation_start_time,omitempty"`
	ReservationSubscriptionId                                string `json:"reservation_subscription_id,omitempty"`
	ReservationTotalReservedNormalizedUnits                  string `json:"reservation_total_reserved_normalized_units,omitempty"`
	ReservationTotalReservedUnits                            string `json:"reservation_total_reserved_units,omitempty"`
	ReservationUnitsPerReservation                           string `json:"reservation_units_per_reservation,omitempty"`
	ReservationUnusedAmortizedUpfrontFeeForBillingPeriod     string `json:"reservation_unused_amortized_upfront_fee_for_billing_period,omitempty"`
	ReservationUnusedNormalizedUnitQuantity                  string `json:"reservation_unused_normalized_unit_quantity,omitempty"`
	ReservationUnusedQuantity                                string `json:"reservation_unused_quantity,omitempty"`
	ReservationUnusedRecurringFee                            string `json:"reservation_unused_recurring_fee,omitempty"`
	ReservationUpfrontValue                                  string `json:"reservation_upfront_value,omitempty"`
	SavingsPlanAmortizedUpfrontCommitmentForBillingPeriod    string `json:"savings_plan_amortized_upfront_commitment_for_billing_period,omitempty"`
	SavingsPlanNetAmortizedUpfrontCommitmentForBillingPeriod string `json:"savings_plan_net_amortized_upfront_commitment_for_billing_period,omitempty"`
	SavingsPlanNetRecurringCommitmentForBillingPeriod        string `json:"savings_plan_net_recurring_commitment_for_billing_period,omitempty"`
	SavingsPlanNetSavingsPlanEffectiveCost                   string `json:"savings_plan_net_savings_plan_effective_cost,omitempty"`
	SavingsPlanRecurringCommitmentForBillingPeriod           string `json:"savings_plan_recurring_commitment_for_billing_period,omitempty"`
	SavingsPlanSavingsPlanARN                                string `json:"savings_plan_savings_plan_a_r_n,omitempty"`
	SavingsPlanSavingsPlanEffectiveCost                      string `json:"savings_plan_savings_plan_effective_cost,omitempty"`
	SavingsPlanSavingsPlanRate                               string `json:"savings_plan_savings_plan_rate,omitempty"`
	SavingsPlanTotalCommitmentToDate                         string `json:"savings_plan_total_commitment_to_date,omitempty"`
	SavingsPlanUsedCommitment                                string `json:"savings_plan_used_commitment,omitempty"`
	Year                                                     string `json:"year,omitempty"`
}
