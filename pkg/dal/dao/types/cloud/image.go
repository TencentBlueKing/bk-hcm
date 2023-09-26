package cloud

import "hcm/pkg/dal/table/cloud/image"

// ImageListResult ...
type ImageListResult struct {
	Count   uint64
	Details []*image.ImageModel
}
