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

package validation

import (
	"testing"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
)

func TestAWSValidateExternalCloudConfig(t *testing.T) {
	grid := []struct {
		Input          kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.ClusterSpec{
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
			},
			ExpectedErrors: []string{"Forbidden::spec.externalCloudControllerManager"},
		},
		{
			Input: kops.ClusterSpec{
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
				CloudConfig: &kops.CloudConfiguration{
					AWSEBSCSIDriver: &kops.AWSEBSCSIDriver{
						Enabled: fi.Bool(true),
					},
				},
			},
		},
		{
			Input: kops.ClusterSpec{
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
				KubeControllerManager: &kops.KubeControllerManagerConfig{
					ExternalCloudVolumePlugin: "aws",
				},
			},
		},
	}
	for _, g := range grid {
		errs := awsValidateExternalCloudControllerManager(g.Input)

		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func TestValidateInstanceGroupSpec(t *testing.T) {
	grid := []struct {
		Input          kops.InstanceGroupSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{},
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"sg-1234abcd"},
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"sg-1234abcd", ""},
			},
			ExpectedErrors: []string{"Invalid value::spec.additionalSecurityGroups[1]"},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{" ", ""},
			},
			ExpectedErrors: []string{
				"Invalid value::spec.additionalSecurityGroups[0]",
				"Invalid value::spec.additionalSecurityGroups[1]",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"--invalid"},
			},
			ExpectedErrors: []string{"Invalid value::spec.additionalSecurityGroups[0]"},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "t2.micro",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "t2.invalidType",
			},
			ExpectedErrors: []string{"Invalid value::test-nodes.spec.machineType"},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "m5.large",
				Image:       "k8s-1.9-debian-stretch-amd64-hvm-ebs-2018-03-11",
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "c5.large",
				Image:       "k8s-1.9-debian-stretch-amd64-hvm-ebs-2018-03-11",
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(55),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(380),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(125),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(120),
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				InstanceInterruptionBehavior: fi.String("invalidValue"),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.instanceInterruptionBehavior",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				InstanceInterruptionBehavior: fi.String("terminate"),
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				InstanceInterruptionBehavior: fi.String("hibernate"),
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				InstanceInterruptionBehavior: fi.String("stop"),
			},
			ExpectedErrors: []string{},
		},
	}
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	for _, g := range grid {
		ig := &kops.InstanceGroup{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-nodes",
			},
			Spec: g.Input,
		}
		errs := awsValidateInstanceGroup(ig, cloud)

		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func TestInstanceMetadataOptions(t *testing.T) {
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")

	tests := []struct {
		ig       *kops.InstanceGroup
		expected []string
	}{
		{
			ig: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-ig",
				},
				Spec: kops.InstanceGroupSpec{
					Role: "Node",
					InstanceMetadata: &kops.InstanceMetadataOptions{
						HTTPPutResponseHopLimit: fi.Int64(1),
						HTTPTokens:              fi.String("abc"),
					},
				},
			},
			expected: []string{"Unsupported value::spec.instanceMetadata.httpTokens"},
		},
		{
			ig: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-ig",
				},
				Spec: kops.InstanceGroupSpec{
					Role: "Node",
					InstanceMetadata: &kops.InstanceMetadataOptions{
						HTTPPutResponseHopLimit: fi.Int64(-1),
						HTTPTokens:              fi.String("required"),
					},
				},
			},
			expected: []string{"Invalid value::spec.instanceMetadata.httpPutResponseHopLimit"},
		},
	}

	for _, test := range tests {
		errs := ValidateInstanceGroup(test.ig, cloud)
		testErrors(t, test.ig.ObjectMeta.Name, errs, test.expected)
	}
}
