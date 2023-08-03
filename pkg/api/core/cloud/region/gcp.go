package region

import "hcm/pkg/criteria/enumor"

// GcpRegion define gcp region.
type GcpRegion struct {
	ID         string        `json:"id"`
	Vendor     enumor.Vendor `json:"vendor"`
	RegionID   string        `json:"region_id"`
	RegionName string        `json:"region_name"`
	Status     string        `json:"status"`
	Creator    string        `json:"creator"`
	Reviser    string        `json:"reviser"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

// GetID ...
func (region GcpRegion) GetID() string {
	return region.ID
}

// GetCloudID ...
func (region GcpRegion) GetCloudID() string {
	return region.RegionID
}
