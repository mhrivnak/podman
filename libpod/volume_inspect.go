package libpod

import (
	"github.com/containers/podman/v3/libpod/define"
	pluginapi "github.com/docker/go-plugins-helpers/volume"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Inspect provides detailed information about the configuration of the given
// volume.
func (v *Volume) Inspect() (*define.InspectVolumeData, error) {
	if !v.valid {
		return nil, define.ErrVolumeRemoved
	}

	v.lock.Lock()
	defer v.lock.Unlock()

	if err := v.update(); err != nil {
		return nil, err
	}

	data := new(define.InspectVolumeData)

	data.Mountpoint = v.config.MountPoint
	if v.UsesVolumeDriver() {
		logrus.Debugf("Querying volume plugin %s for status", v.config.Driver)
		data.Mountpoint = v.state.MountPoint

		if v.plugin == nil {
			return nil, errors.Wrapf(define.ErrMissingPlugin, "volume %s uses volume plugin %s but it is not available, cannot inspect", v.Name(), v.config.Driver)
		}

		// Retrieve status for the volume.
		// Need to query the volume driver.
		req := new(pluginapi.GetRequest)
		req.Name = v.Name()
		resp, err := v.plugin.GetVolume(req)
		if err != nil {
			return nil, errors.Wrapf(err, "error retrieving volume %s information from plugin %s", v.Name(), v.Driver())
		}
		if resp != nil {
			data.Status = resp.Status
		}
	}

	data.Name = v.config.Name
	data.Driver = v.config.Driver
	data.CreatedAt = v.config.CreatedTime
	data.Labels = make(map[string]string)
	for k, v := range v.config.Labels {
		data.Labels[k] = v
	}
	data.Scope = v.Scope()
	data.Options = make(map[string]string)
	for k, v := range v.config.Options {
		data.Options[k] = v
	}
	data.UID = v.uid()
	data.GID = v.gid()
	data.Anonymous = v.config.IsAnon
	data.MountCount = v.state.MountCount
	data.NeedsCopyUp = v.state.NeedsCopyUp
	data.NeedsChown = v.state.NeedsChown

	return data, nil
}
