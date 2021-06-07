/*
Copyright 2021 The Kubernetes Authors.

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

package assettasks

import (
	"fmt"
	"sort"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
)

type copyAssetsTarget struct {
}

func (c copyAssetsTarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (c copyAssetsTarget) ProcessDeletions() bool {
	return false
}

func Copy(imageAssets []*assets.ImageAsset, fileAssets []*assets.FileAsset, cluster *kops.Cluster) error {
	tasks := map[string]fi.Task{}

	for _, imageAsset := range imageAssets {
		if imageAsset.DownloadLocation != imageAsset.CanonicalLocation {
			ctx := &fi.ModelBuilderContext{
				Tasks: tasks,
			}

			copyImageTask := &CopyImage{
				Name:        fi.String(imageAsset.DownloadLocation),
				SourceImage: fi.String(imageAsset.CanonicalLocation),
				TargetImage: fi.String(imageAsset.DownloadLocation),
				Lifecycle:   fi.LifecycleSync,
			}

			if err := ctx.EnsureTask(copyImageTask); err != nil {
				return fmt.Errorf("error adding image-copy task: %v", err)
			}
			tasks = ctx.Tasks
		}
	}

	for _, fileAsset := range fileAssets {

		// test if the asset needs to be copied
		if fileAsset.DownloadURL.String() != fileAsset.CanonicalURL.String() {
			ctx := &fi.ModelBuilderContext{
				Tasks: tasks,
			}

			copyFileTask := &CopyFile{
				Name:       fi.String(fileAsset.CanonicalURL.String()),
				TargetFile: fi.String(fileAsset.DownloadURL.String()),
				SourceFile: fi.String(fileAsset.CanonicalURL.String()),
				SHA:        fi.String(fileAsset.SHAValue),
				Lifecycle:  fi.LifecycleSync,
			}

			if err := ctx.EnsureTask(copyFileTask); err != nil {
				return fmt.Errorf("error adding file-copy task: %v", err)
			}
			tasks = ctx.Tasks
		}
	}

	var options fi.RunTasksOptions
	options.InitDefaults()

	context, err := fi.NewContext(&copyAssetsTarget{}, cluster, nil, nil, nil, nil, true, tasks)
	if err != nil {
		return fmt.Errorf("error building context: %v", err)
	}
	defer context.Close()

	ch := make(chan error, 5)
	for i := 0; i < cap(ch); i++ {
		ch <- nil
	}

	gotError := false
	names := make([]string, 0, len(tasks))
	for name := range tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		task := tasks[name]
		err := <-ch
		if err != nil {
			klog.Warning(err)
			gotError = true
		}
		go func(n string, t fi.Task) {
			err := t.Run(context)
			if err != nil {
				err = fmt.Errorf("%s: %v", n, err)
			}
			ch <- err
		}(name, task)
	}

	for i := 0; i < cap(ch); i++ {
		err = <-ch
		if err != nil {
			klog.Warning(err)
			gotError = true
		}
	}

	close(ch)
	if gotError {
		return fmt.Errorf("not all assets copied successfully")
	}
	return nil
}
