/*
Copyright 2019 The Kubernetes Authors.

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

package mockec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"
)

type vpcInfo struct {
	main       ec2.Vpc
	attributes ec2.DescribeVpcAttributeOutput
}

func (m *MockEC2) FindVpc(id string) *ec2.Vpc {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	vpc := m.Vpcs[id]
	if vpc == nil {
		return nil
	}

	copy := vpc.main
	copy.Tags = m.getTags(ec2.ResourceTypeVpc, *vpc.main.VpcId)

	return &copy
}

func (m *MockEC2) CreateVpcRequest(*ec2.CreateVpcInput) (*request.Request, *ec2.CreateVpcOutput) {
	panic("Not implemented")
}

func (m *MockEC2) CreateVpcWithContext(aws.Context, *ec2.CreateVpcInput, ...request.Option) (*ec2.CreateVpcOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) CreateVpcWithId(request *ec2.CreateVpcInput, id string) (*ec2.CreateVpcOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tags := tagSpecificationsToTags(request.TagSpecifications, ec2.ResourceTypeVpc)

	vpc := &vpcInfo{
		main: ec2.Vpc{
			VpcId:     s(id),
			CidrBlock: request.CidrBlock,
			IsDefault: aws.Bool(false),
			Tags:      tags,
		},
		attributes: ec2.DescribeVpcAttributeOutput{
			EnableDnsHostnames: &ec2.AttributeBooleanValue{Value: aws.Bool(true)},
			EnableDnsSupport:   &ec2.AttributeBooleanValue{Value: aws.Bool(true)},
		},
	}

	if m.Vpcs == nil {
		m.Vpcs = make(map[string]*vpcInfo)
	}
	m.Vpcs[id] = vpc

	m.addTags(id, tags...)

	response := &ec2.CreateVpcOutput{
		Vpc: &vpc.main,
	}
	return response, nil
}

func (m *MockEC2) CreateVpc(request *ec2.CreateVpcInput) (*ec2.CreateVpcOutput, error) {
	klog.Infof("CreateVpc: %v", request)

	if request.DryRun != nil {
		klog.Fatalf("DryRun")
	}

	id := m.allocateId("vpc")

	return m.CreateVpcWithId(request, id)
}

func (m *MockEC2) DescribeVpcsRequest(*ec2.DescribeVpcsInput) (*request.Request, *ec2.DescribeVpcsOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeVpcsWithContext(aws.Context, *ec2.DescribeVpcsInput, ...request.Option) (*ec2.DescribeVpcsOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeVpcs(request *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeVpcs: %v", request)

	if len(request.VpcIds) != 0 {
		request.Filters = append(request.Filters, &ec2.Filter{Name: s("vpc-id"), Values: request.VpcIds})
	}

	var vpcs []*ec2.Vpc

	for k, vpc := range m.Vpcs {
		allFiltersMatch := true
		for _, filter := range request.Filters {
			match := false
			switch *filter.Name {
			case "vpc-id":
				for _, v := range filter.Values {
					if k == *v {
						match = true
					}
				}
			default:
				match = m.hasTag(ec2.ResourceTypeVpc, *vpc.main.VpcId, filter)
			}

			if !match {
				allFiltersMatch = false
				break
			}
		}

		if !allFiltersMatch {
			continue
		}

		copy := vpc.main
		copy.Tags = m.getTags(ec2.ResourceTypeVpc, *vpc.main.VpcId)
		vpcs = append(vpcs, &copy)
	}

	response := &ec2.DescribeVpcsOutput{
		Vpcs: vpcs,
	}

	return response, nil
}

func (m *MockEC2) DescribeVpcAttributeRequest(*ec2.DescribeVpcAttributeInput) (*request.Request, *ec2.DescribeVpcAttributeOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeVpcAttributeWithContext(aws.Context, *ec2.DescribeVpcAttributeInput, ...request.Option) (*ec2.DescribeVpcAttributeOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DescribeVpcAttribute(request *ec2.DescribeVpcAttributeInput) (*ec2.DescribeVpcAttributeOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DescribeVpcAttribute: %v", request)

	vpc := m.Vpcs[*request.VpcId]
	if vpc == nil {
		return nil, fmt.Errorf("not found")
	}

	response := &ec2.DescribeVpcAttributeOutput{
		VpcId: vpc.main.VpcId,

		EnableDnsHostnames: vpc.attributes.EnableDnsHostnames,
		EnableDnsSupport:   vpc.attributes.EnableDnsSupport,
	}

	return response, nil
}

func (m *MockEC2) ModifyVpcAttribute(request *ec2.ModifyVpcAttributeInput) (*ec2.ModifyVpcAttributeOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("ModifyVpcAttribute: %v", request)

	vpc := m.Vpcs[*request.VpcId]
	if vpc == nil {
		return nil, fmt.Errorf("not found")
	}

	if request.EnableDnsHostnames != nil {
		vpc.attributes.EnableDnsHostnames = request.EnableDnsHostnames
	}

	if request.EnableDnsSupport != nil {
		vpc.attributes.EnableDnsSupport = request.EnableDnsSupport
	}

	response := &ec2.ModifyVpcAttributeOutput{}

	return response, nil
}

func (m *MockEC2) ModifyVpcAttributeWithContext(aws.Context, *ec2.ModifyVpcAttributeInput, ...request.Option) (*ec2.ModifyVpcAttributeOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) ModifyVpcAttributeRequest(*ec2.ModifyVpcAttributeInput) (*request.Request, *ec2.ModifyVpcAttributeOutput) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteVpc(request *ec2.DeleteVpcInput) (*ec2.DeleteVpcOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DeleteVpc: %v", request)

	id := aws.StringValue(request.VpcId)
	o := m.Vpcs[id]
	if o == nil {
		return nil, fmt.Errorf("VPC %q not found", id)
	}
	delete(m.Vpcs, id)

	return &ec2.DeleteVpcOutput{}, nil
}

func (m *MockEC2) DeleteVpcWithContext(aws.Context, *ec2.DeleteVpcInput, ...request.Option) (*ec2.DeleteVpcOutput, error) {
	panic("Not implemented")
}

func (m *MockEC2) DeleteVpcRequest(*ec2.DeleteVpcInput) (*request.Request, *ec2.DeleteVpcOutput) {
	panic("Not implemented")
}

func (m *MockEC2) AssociateVpcCidrBlock(request *ec2.AssociateVpcCidrBlockInput) (*ec2.AssociateVpcCidrBlockOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("AssociateVpcCidrBlock: %v", request)

	id := aws.StringValue(request.VpcId)
	vpc, ok := m.Vpcs[id]
	if !ok {
		return nil, fmt.Errorf("VPC %q not found", id)
	}
	var ipv4association *ec2.VpcCidrBlockAssociation
	var ipv6association *ec2.VpcIpv6CidrBlockAssociation
	if aws.BoolValue(request.AmazonProvidedIpv6CidrBlock) {
		ipv6association = &ec2.VpcIpv6CidrBlockAssociation{
			Ipv6Pool:      aws.String("Amazon"),
			Ipv6CidrBlock: aws.String("2001:db8::/56"),
			AssociationId: aws.String(fmt.Sprintf("%v-%v", id, len(vpc.main.Ipv6CidrBlockAssociationSet))),
			Ipv6CidrBlockState: &ec2.VpcCidrBlockState{
				State: aws.String(ec2.VpcCidrBlockStateCodeAssociated),
			},
		}
		vpc.main.Ipv6CidrBlockAssociationSet = append(vpc.main.Ipv6CidrBlockAssociationSet, ipv6association)
	} else {
		ipv4association = &ec2.VpcCidrBlockAssociation{
			CidrBlock:     request.CidrBlock,
			AssociationId: aws.String(fmt.Sprintf("%v-%v", id, len(vpc.main.CidrBlockAssociationSet))),
			CidrBlockState: &ec2.VpcCidrBlockState{
				State: aws.String(ec2.VpcCidrBlockStateCodeAssociated),
			},
		}
		vpc.main.CidrBlockAssociationSet = append(vpc.main.CidrBlockAssociationSet, ipv4association)
	}

	return &ec2.AssociateVpcCidrBlockOutput{
		CidrBlockAssociation:     ipv4association,
		Ipv6CidrBlockAssociation: ipv6association,
		VpcId:                    request.VpcId,
	}, nil
}

func (m *MockEC2) DisassociateVpcCidrBlock(request *ec2.DisassociateVpcCidrBlockInput) (*ec2.DisassociateVpcCidrBlockOutput, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	klog.Infof("DisassociateVpcCidrBlock: %v", request)

	id := aws.StringValue(request.AssociationId)
	var association *ec2.VpcCidrBlockAssociation
	var vpcID *string
	for _, vpc := range m.Vpcs {
		for _, a := range vpc.main.CidrBlockAssociationSet {
			if aws.StringValue(a.AssociationId) == id {
				a.CidrBlockState.State = aws.String(ec2.VpcCidrBlockStateCodeDisassociated)
				association = a
				vpcID = vpc.main.VpcId
				break
			}
		}
	}
	if association == nil {
		return nil, fmt.Errorf("VPC association %q not found", id)
	}

	return &ec2.DisassociateVpcCidrBlockOutput{
		CidrBlockAssociation: association,
		VpcId:                vpcID,
	}, nil
}
