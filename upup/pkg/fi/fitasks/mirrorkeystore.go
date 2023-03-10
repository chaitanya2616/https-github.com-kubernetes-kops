/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fitasks

import (
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// +kops:fitask
type MirrorKeystore struct {
	Name      *string
	Lifecycle fi.Lifecycle

	MirrorPath vfs.Path
}

var _ fi.CloudupHasDependencies = &MirrorKeystore{}

// GetDependencies returns the dependencies for a MirrorKeystore task - it must run after all secrets have been run
func (e *MirrorKeystore) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Secret); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

// Find implements fi.Task::Find
func (e *MirrorKeystore) Find(c *fi.CloudupContext) (*MirrorKeystore, error) {
	if vfsKeystore, ok := c.T.Keystore.(*fi.VFSCAStore); ok {
		if vfsKeystore.VFSPath().Path() == e.MirrorPath.Path() {
			return e, nil
		}
	}

	// TODO: implement Find so that we aren't always mirroring
	klog.V(2).Infof("MirrorKeystore::Find not implemented; always copying (inefficient)")
	return nil, nil
}

// Run implements fi.Task::Run
func (e *MirrorKeystore) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

// CheckChanges implements fi.Task::CheckChanges
func (s *MirrorKeystore) CheckChanges(a, e, changes *MirrorKeystore) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

// Render implements fi.Task::Render
func (_ *MirrorKeystore) Render(c *fi.CloudupContext, a, e, changes *MirrorKeystore) error {
	keystore := c.T.Keystore

	return keystore.MirrorTo(e.MirrorPath)
}
