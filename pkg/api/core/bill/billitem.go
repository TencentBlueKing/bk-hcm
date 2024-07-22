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
	UsageEndTime              *string          `json:"usage_end_time"`
	UsagePricingUnit          *string          `json:"usage_pricing_unit"`
	UsageStartTime            *string          `json:"usage_start_time"`
	UsageUnit                 *string          `json:"usage_unit"`
	Zone                      *string          `json:"zone"`
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
}

// AwsRawBillItem Aws 原始账单结构
type AwsRawBillItem struct {
	BillBillType                                             string `json:"bill_bill_type"`
	BillBillingEntity                                        string `json:"bill_billing_entity"`
	BillBillingPeriodEndDate                                 string `json:"bill_billing_period_end_date"`
	BillBillingPeriodStartDate                               string `json:"bill_billing_period_start_date"`
	BillInvoiceId                                            string `json:"bill_invoice_id"`
	BillInvoicingEntity                                      string `json:"bill_invoicing_entity"`
	BillPayerAccountId                                       string `json:"bill_payer_account_id"`
	DiscountEdpDiscount                                      string `json:"discount_edp_discount"`
	DiscountTotalDiscount                                    string `json:"discount_total_discount"`
	DiscountPrivateRateDiscount                              string `json:"discount_private_rate_discount"`
	IdentityLineItemId                                       string `json:"identity_line_item_id"`
	IdentityTimeInterval                                     string `json:"identity_time_interval"`
	LineItemAvailabilityZone                                 string `json:"line_item_availability_zone"`
	LineItemBlendedCost                                      string `json:"line_item_blended_cost"`
	LineItemBlendedRate                                      string `json:"line_item_blended_rate"`
	LineItemCurrencyCode                                     string `json:"line_item_currency_code"`
	LineItemLegalEntity                                      string `json:"line_item_legal_entity"`
	LineItemLineItemDescription                              string `json:"line_item_line_item_description"`
	LineItemLineItemType                                     string `json:"line_item_line_item_type"`
	LineItemNetUnblendedCost                                 string `json:"line_item_net_unblended_cost"`
	LineItemNetUnblendedRate                                 string `json:"line_item_net_unblended_rate"`
	LineItemNormalizationFactor                              string `json:"line_item_normalization_factor"`
	LineItemNormalizedUsageAmount                            string `json:"line_item_normalized_usage_amount"`
	LineItemOperation                                        string `json:"line_item_operation"`
	LineItemProductCode                                      string `json:"line_item_product_code"`
	LineItemResourceId                                       string `json:"line_item_resource_id"`
	LineItemTaxType                                          string `json:"line_item_tax_type"`
	LineItemUnblendedCost                                    string `json:"line_item_unblended_cost"`
	LineItemUnblendedRate                                    string `json:"line_item_unblended_rate"`
	LineItemUsageAccountId                                   string `json:"line_item_usage_account_id"`
	LineItemUsageAmount                                      string `json:"line_item_usage_amount"`
	LineItemUsageEndDate                                     string `json:"line_item_usage_end_date"`
	LineItemUsageStartDate                                   string `json:"line_item_usage_start_date"`
	LineItemUsageType                                        string `json:"line_item_usage_type"`
	Month                                                    string `json:"month"`
	PricingCurrency                                          string `json:"pricing_currency"`
	PricingPublicOnDemandCost                                string `json:"pricing_public_on_demand_cost"`
	PricingPublicOnDemandRate                                string `json:"pricing_public_on_demand_rate"`
	PricingRateCode                                          string `json:"pricing_rate_code"`
	PricingRateId                                            string `json:"pricing_rate_id"`
	PricingTerm                                              string `json:"pricing_term"`
	PricingUnit                                              string `json:"pricing_unit"`
	ProductAccessType                                        string `json:"product_access_type"`
	ProductAccountAssistance                                 string `json:"product_account_assistance"`
	ProductAlarmType                                         string `json:"product_alarm_type"`
	ProductAlpha3Countrycode                                 string `json:"product_alpha3countrycode"`
	ProductArchitecturalReview                               string `json:"product_architectural_review"`
	ProductArchitectureSupport                               string `json:"product_architecture_support"`
	ProductAttachmentType                                    string `json:"product_attachment_type"`
	ProductAvailability                                      string `json:"product_availability"`
	ProductAvailabilityZone                                  string `json:"product_availability_zone"`
	ProductBackupservice                                     string `json:"product_backupservice"`
	ProductBestPractices                                     string `json:"product_best_practices"`
	ProductCacheEngine                                       string `json:"product_cache_engine"`
	ProductCapacity                                          string `json:"product_capacity"`
	ProductCapacitystatus                                    string `json:"product_capacitystatus"`
	ProductCaseSeverityresponseTimes                         string `json:"product_case_severityresponse_times"`
	ProductCategory                                          string `json:"product_category"`
	ProductCiType                                            string `json:"product_ci_type"`
	ProductClassicnetworkingsupport                          string `json:"product_classicnetworkingsupport"`
	ProductClientLocation                                    string `json:"product_client_location"`
	ProductClockSpeed                                        string `json:"product_clock_speed"`
	ProductComponent                                         string `json:"product_component"`
	ProductComputeFamily                                     string `json:"product_compute_family"`
	ProductComputeType                                       string `json:"product_compute_type"`
	ProductConnectionType                                    string `json:"product_connection_type"`
	ProductContentType                                       string `json:"product_content_type"`
	ProductContinent                                         string `json:"product_continent"`
	ProductCountry                                           string `json:"product_country"`
	ProductCurrentGeneration                                 string `json:"product_current_generation"`
	ProductCustomerServiceAndCommunities                     string `json:"product_customer_service_and_communities"`
	ProductDatabaseEngine                                    string `json:"product_database_engine"`
	ProductDatatransferout                                   string `json:"product_datatransferout"`
	ProductDedicatedEbsThroughput                            string `json:"product_dedicated_ebs_throughput"`
	ProductDeploymentOption                                  string `json:"product_deployment_option"`
	ProductDescription                                       string `json:"product_description"`
	ProductDirectConnectLocation                             string `json:"product_direct_connect_location"`
	ProductDominantnondominant                               string `json:"product_dominantnondominant"`
	ProductDurability                                        string `json:"product_durability"`
	ProductEcu                                               string `json:"product_ecu"`
	ProductEndpointType                                      string `json:"product_endpoint_type"`
	ProductEngineCode                                        string `json:"product_engine_code"`
	ProductEngineMajorVersion                                string `json:"product_engine_major_version"`
	ProductEnhancedNetworkingSupported                       string `json:"product_enhanced_networking_supported"`
	ProductExtendedSupportPricingYear                        string `json:"product_extended_support_pricing_year"`
	ProductFeeDescription                                    string `json:"product_fee_description"`
	ProductFreeQueryTypes                                    string `json:"product_free_query_types"`
	ProductFreeTier                                          string `json:"product_free_tier"`
	ProductFromLocation                                      string `json:"product_from_location"`
	ProductFromLocationType                                  string `json:"product_from_location_type"`
	ProductFromRegionCode                                    string `json:"product_from_region_code"`
	ProductGeoregioncode                                     string `json:"product_georegioncode"`
	ProductGpu                                               string `json:"product_gpu"`
	ProductGpuMemory                                         string `json:"product_gpu_memory"`
	ProductGroup                                             string `json:"product_group"`
	ProductGroupDescription                                  string `json:"product_group_description"`
	ProductIncludedServices                                  string `json:"product_included_services"`
	ProductInsightstype                                      string `json:"product_insightstype"`
	ProductInstance                                          string `json:"product_instance"`
	ProductInstanceFamily                                    string `json:"product_instance_family"`
	ProductInstanceName                                      string `json:"product_instance_name"`
	ProductInstanceType                                      string `json:"product_instance_type"`
	ProductInstanceTypeFamily                                string `json:"product_instance_type_family"`
	ProductIntelAvx2Available                                string `json:"product_intel_avx2_available"`
	ProductIntelAvxAvailable                                 string `json:"product_intel_avx_available"`
	ProductIntelTurboAvailable                               string `json:"product_intel_turbo_available"`
	ProductLaunchSupport                                     string `json:"product_launch_support"`
	ProductLicenseModel                                      string `json:"product_license_model"`
	ProductLocation                                          string `json:"product_location"`
	ProductLocationType                                      string `json:"product_location_type"`
	ProductLogsDestination                                   string `json:"product_logs_destination"`
	ProductMailboxStorage                                    string `json:"product_mailbox_storage"`
	ProductMarketoption                                      string `json:"product_marketoption"`
	ProductMaxIopsBurstPerformance                           string `json:"product_max_iops_burst_performance"`
	ProductMaxIopsvolume                                     string `json:"product_max_iopsvolume"`
	ProductMaxThroughputvolume                               string `json:"product_max_throughputvolume"`
	ProductMaxVolumeSize                                     string `json:"product_max_volume_size"`
	ProductMemory                                            string `json:"product_memory"`
	ProductMemoryGib                                         string `json:"product_memory_gib"`
	ProductMessageDeliveryFrequency                          string `json:"product_message_delivery_frequency"`
	ProductMessageDeliveryOrder                              string `json:"product_message_delivery_order"`
	ProductMinVolumeSize                                     string `json:"product_min_volume_size"`
	ProductNetworkPerformance                                string `json:"product_network_performance"`
	ProductNormalizationSizeFactor                           string `json:"product_normalization_size_factor"`
	ProductOperatingSystem                                   string `json:"product_operating_system"`
	ProductOperation                                         string `json:"product_operation"`
	ProductOperationsSupport                                 string `json:"product_operations_support"`
	ProductOrigin                                            string `json:"product_origin"`
	ProductPhysicalCpu                                       string `json:"product_physical_cpu"`
	ProductPhysicalGpu                                       string `json:"product_physical_gpu"`
	ProductPhysicalProcessor                                 string `json:"product_physical_processor"`
	ProductPlatoclassificationtype                           string `json:"product_platoclassificationtype"`
	ProductPlatoinstancename                                 string `json:"product_platoinstancename"`
	ProductPlatoinstancetype                                 string `json:"product_platoinstancetype"`
	ProductPlatopricingtype                                  string `json:"product_platopricingtype"`
	ProductPortSpeed                                         string `json:"product_port_speed"`
	ProductPreInstalledSw                                    string `json:"product_pre_installed_sw"`
	ProductPricingUnit                                       string `json:"product_pricing_unit"`
	ProductProactiveGuidance                                 string `json:"product_proactive_guidance"`
	ProductProcessorArchitecture                             string `json:"product_processor_architecture"`
	ProductProcessorFeatures                                 string `json:"product_processor_features"`
	ProductProductFamily                                     string `json:"product_product_family"`
	ProductProductName                                       string `json:"product_product_name"`
	ProductProgrammaticCaseManagement                        string `json:"product_programmatic_case_management"`
	ProductProvisioned                                       string `json:"product_provisioned"`
	ProductQueueType                                         string `json:"product_queue_type"`
	ProductRecipient                                         string `json:"product_recipient"`
	ProductRegion                                            string `json:"product_region"`
	ProductRegionCode                                        string `json:"product_region_code"`
	ProductRequestDescription                                string `json:"product_request_description"`
	ProductRequestType                                       string `json:"product_request_type"`
	ProductResourceEndpoint                                  string `json:"product_resource_endpoint"`
	ProductResourcePriceGroup                                string `json:"product_resource_price_group"`
	ProductResourceType                                      string `json:"product_resource_type"`
	ProductRoutingTarget                                     string `json:"product_routing_target"`
	ProductRoutingType                                       string `json:"product_routing_type"`
	ProductServicecode                                       string `json:"product_servicecode"`
	ProductServicename                                       string `json:"product_servicename"`
	ProductSku                                               string `json:"product_sku"`
	ProductStorage                                           string `json:"product_storage"`
	ProductStorageClass                                      string `json:"product_storage_class"`
	ProductStorageFamily                                     string `json:"product_storage_family"`
	ProductStorageMedia                                      string `json:"product_storage_media"`
	ProductStorageType                                       string `json:"product_storage_type"`
	ProductTechnicalSupport                                  string `json:"product_technical_support"`
	ProductTenancy                                           string `json:"product_tenancy"`
	ProductThirdpartySoftwareSupport                         string `json:"product_thirdparty_software_support"`
	ProductTickettype                                        string `json:"product_tickettype"`
	ProductTiertype                                          string `json:"product_tiertype"`
	ProductToLocation                                        string `json:"product_to_location"`
	ProductToLocationType                                    string `json:"product_to_location_type"`
	ProductToRegionCode                                      string `json:"product_to_region_code"`
	ProductTrafficDirection                                  string `json:"product_traffic_direction"`
	ProductTraining                                          string `json:"product_training"`
	ProductTransferType                                      string `json:"product_transfer_type"`
	ProductUsagetype                                         string `json:"product_usagetype"`
	ProductVaulttype                                         string `json:"product_vaulttype"`
	ProductVcpu                                              string `json:"product_vcpu"`
	ProductVersion                                           string `json:"product_version"`
	ProductVirtualInterfaceType                              string `json:"product_virtual_interface_type"`
	ProductVolumeApiName                                     string `json:"product_volume_api_name"`
	ProductVolumeType                                        string `json:"product_volume_type"`
	ProductVpcnetworkingsupport                              string `json:"product_vpcnetworkingsupport"`
	ProductWhoCanOpenCases                                   string `json:"product_who_can_open_cases"`
	ReservationAmortizedUpfrontCostForUsage                  string `json:"reservation_amortized_upfront_cost_for_usage"`
	ReservationAmortizedUpfrontFeeForBillingPeriod           string `json:"reservation_amortized_upfront_fee_for_billing_period"`
	ReservationEffectiveCost                                 string `json:"reservation_effective_cost"`
	ReservationEndTime                                       string `json:"reservation_end_time"`
	ReservationModificationStatus                            string `json:"reservation_modification_status"`
	ReservationNetAmortizedUpfrontCostForUsage               string `json:"reservation_net_amortized_upfront_cost_for_usage"`
	ReservationNetAmortizedUpfrontFeeForBillingPeriod        string `json:"reservation_net_amortized_upfront_fee_for_billing_period"`
	ReservationNetEffectiveCost                              string `json:"reservation_net_effective_cost"`
	ReservationNetRecurringFeeForUsage                       string `json:"reservation_net_recurring_fee_for_usage"`
	ReservationNetUnusedAmortizedUpfrontFeeForBillingPeriod  string `json:"reservation_net_unused_amortized_upfront_fee_for_billing_period"`
	ReservationNetUnusedRecurringFee                         string `json:"reservation_net_unused_recurring_fee"`
	ReservationNetUpfrontValue                               string `json:"reservation_net_upfront_value"`
	ReservationNormalizedUnitsPerReservation                 string `json:"reservation_normalized_units_per_reservation"`
	ReservationNumberOfReservations                          string `json:"reservation_number_of_reservations"`
	ReservationRecurringFeeForUsage                          string `json:"reservation_recurring_fee_for_usage"`
	ReservationStartTime                                     string `json:"reservation_start_time"`
	ReservationSubscriptionId                                string `json:"reservation_subscription_id"`
	ReservationTotalReservedNormalizedUnits                  string `json:"reservation_total_reserved_normalized_units"`
	ReservationTotalReservedUnits                            string `json:"reservation_total_reserved_units"`
	ReservationUnitsPerReservation                           string `json:"reservation_units_per_reservation"`
	ReservationUnusedAmortizedUpfrontFeeForBillingPeriod     string `json:"reservation_unused_amortized_upfront_fee_for_billing_period"`
	ReservationUnusedNormalizedUnitQuantity                  string `json:"reservation_unused_normalized_unit_quantity"`
	ReservationUnusedQuantity                                string `json:"reservation_unused_quantity"`
	ReservationUnusedRecurringFee                            string `json:"reservation_unused_recurring_fee"`
	ReservationUpfrontValue                                  string `json:"reservation_upfront_value"`
	SavingsPlanAmortizedUpfrontCommitmentForBillingPeriod    string `json:"savings_plan_amortized_upfront_commitment_for_billing_period"`
	SavingsPlanNetAmortizedUpfrontCommitmentForBillingPeriod string `json:"savings_plan_net_amortized_upfront_commitment_for_billing_period"`
	SavingsPlanNetRecurringCommitmentForBillingPeriod        string `json:"savings_plan_net_recurring_commitment_for_billing_period"`
	SavingsPlanNetSavingsPlanEffectiveCost                   string `json:"savings_plan_net_savings_plan_effective_cost"`
	SavingsPlanRecurringCommitmentForBillingPeriod           string `json:"savings_plan_recurring_commitment_for_billing_period"`
	SavingsPlanSavingsPlanARN                                string `json:"savings_plan_savings_plan_a_r_n"`
	SavingsPlanSavingsPlanEffectiveCost                      string `json:"savings_plan_savings_plan_effective_cost"`
	SavingsPlanSavingsPlanRate                               string `json:"savings_plan_savings_plan_rate"`
	SavingsPlanTotalCommitmentToDate                         string `json:"savings_plan_total_commitment_to_date"`
	SavingsPlanUsedCommitment                                string `json:"savings_plan_used_commitment"`
	Year                                                     string `json:"year"`
}
