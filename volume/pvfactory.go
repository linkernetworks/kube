package volumes

import (
	"bitbucket.org/linkernetworks/aurora/src/entity"
	v1 "k8s.io/api/core/v1"
)

type PVFactory interface {
	NewPV(name string, capacity string, accessMode string, provider entity.VolumeProvider) (*v1.PersistentVolume, error)
}
