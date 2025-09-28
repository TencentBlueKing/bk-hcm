package eip

import (
	"hcm/pkg/rest"
)

// EipExtRetrieveResp 返回单个 eip 详情
type EipExtRetrieveResp[T EipExtensionResult] struct {
	rest.BaseResp `json:",inline"`
	Data          *EipExtResult[T] `json:"data"`
}

// EipExtResult ...
type EipExtResult[T EipExtensionResult] struct {
	ID        string  `json:"id,omitempty"`
	AccountID string  `json:"account_id,omitempty"`
	Vendor    string  `json:"vendor,omitempty"`
	Name      *string `json:"name,omitempty"`
	CloudID   string  `json:"cloud_id,omitempty"`
	BkBizID   int64   `json:"bk_biz_id,omitempty"`
	Region    string  `json:"region,omitempty"`
	// InstanceID db并不返回该字段
	InstanceID    *string `json:"instance_id,omitempty"`
	InstanceType  string  `json:"instance_type,omitempty"`
	Status        string  `json:"status,omitempty"`
	RecycleStatus string  `json:"recycle_status"`
	PublicIp      string  `json:"public_ip,omitempty"`
	PrivateIp     string  `json:"private_ip,omitempty"`
	Extension     *T      `json:"extension,omitempty"`
	Creator       string  `json:"creator,omitempty"`
	Reviser       string  `json:"reviser,omitempty"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
}

// GetID ...
func (eip *EipExtResult[T]) GetID() string {
	return eip.ID
}

// GetCloudID ...
func (eip *EipExtResult[T]) GetCloudID() string {
	return eip.CloudID
}

// EipExtensionResult ...
type EipExtensionResult interface {
	TCloudEipExtensionResult |
		AwsEipExtensionResult |
		GcpEipExtensionResult |
		AzureEipExtensionResult |
		HuaWeiEipExtensionResult
}

// EipListResp ...
type EipListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *EipListResult `json:"data"`
}

// EipListResult ...
type EipListResult struct {
	Count   *uint64      `json:"count,omitempty"`
	Details []*EipResult `json:"details"`
}

// EipResult ...
type EipResult struct {
	ID            string  `json:"id,omitempty"`
	Vendor        string  `json:"vendor,omitempty"`
	AccountID     string  `json:"account_id,omitempty"`
	Name          *string `json:"name,omitempty"`
	CloudID       string  `json:"cloud_id,omitempty"`
	BkBizID       int64   `json:"bk_biz_id,omitempty"`
	Region        string  `json:"region,omitempty"`
	InstanceID    *string `json:"instance_id,omitempty"`
	InstanceType  string  `json:"instance_type,omitempty"`
	Status        string  `json:"status,omitempty"`
	RecycleStatus string  `json:"recycle_status,omitempty"`
	PublicIp      string  `json:"public_ip,omitempty"`
	PrivateIp     string  `json:"private_ip,omitempty"`
	Creator       string  `json:"creator,omitempty"`
	Reviser       string  `json:"reviser,omitempty"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
}

// EipExtListResult ...
type EipExtListResult[T EipExtensionResult] struct {
	Count   *uint64            `json:"count,omitempty"`
	Details []*EipExtResult[T] `json:"details"`
}

// EipExtListResp ...
type EipExtListResp[T EipExtensionResult] struct {
	rest.BaseResp `json:",inline"`
	Data          *EipExtListResult[T] `json:"data"`
}
