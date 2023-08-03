package region

import "hcm/pkg/criteria/enumor"

// AwsRegion define aws region.
type AwsRegion struct {
	ID         string        `json:"id"`
	Vendor     enumor.Vendor `json:"vendor"`
	AccountID  string        `json:"account_id"`
	RegionID   string        `json:"region_id"`
	RegionName string        `json:"region_name"`
	Status     string        `json:"status"`
	Endpoint   string        `json:"endpoint"`
	Creator    string        `json:"creator"`
	Reviser    string        `json:"reviser"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

// GetID ...
func (region AwsRegion) GetID() string {
	return region.ID
}

// GetCloudID ...
func (region AwsRegion) GetCloudID() string {
	return region.RegionID
}
