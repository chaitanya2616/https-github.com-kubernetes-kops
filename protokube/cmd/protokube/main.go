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

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/wellknownports"
	gossiputils "k8s.io/kops/protokube/pkg/gossip"
	gossipdns "k8s.io/kops/protokube/pkg/gossip/dns"
	_ "k8s.io/kops/protokube/pkg/gossip/memberlist"
	_ "k8s.io/kops/protokube/pkg/gossip/mesh"
	"k8s.io/kops/protokube/pkg/protokube"
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)
	// BuildVersion is overwritten during build. This can be used to resolve issues.
	BuildVersion = "0.1"
)

func main() {
	klog.InitFlags(nil)

	fmt.Printf("protokube version %s\n", BuildVersion)

	if err := run(); err != nil {
		klog.Errorf("Error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// run is responsible for running the protokube service controller
func run() error {
	var zones []string
	var containerized, master, gossip bool
	var cloud, clusterID, dnsInternalSuffix, gossipSecret, gossipListen, gossipProtocol, gossipSecretSecondary, gossipListenSecondary, gossipProtocolSecondary string
	var flagChannels string
	var dnsUpdateInterval int

	flag.BoolVar(&containerized, "containerized", containerized, "Set if we are running containerized")
	flag.BoolVar(&gossip, "gossip", gossip, "Set if we are using gossip dns")
	flag.BoolVar(&master, "master", master, "Whether or not this node is a master")
	flag.StringVar(&cloud, "cloud", "aws", "CloudProvider we are using (aws,digitalocean,gce,openstack)")
	flag.StringVar(&clusterID, "cluster-id", clusterID, "Cluster ID")
	flag.StringVar(&dnsInternalSuffix, "dns-internal-suffix", dnsInternalSuffix, "DNS suffix for internal domain names")
	flags.IntVar(&dnsUpdateInterval, "dns-update-interval", 5, "Configure interval at which to update DNS records.")
	flag.StringVar(&flagChannels, "channels", flagChannels, "channels to install")
	flag.StringVar(&gossipProtocol, "gossip-protocol", "mesh", "mesh/memberlist")
	flag.StringVar(&gossipListen, "gossip-listen", fmt.Sprintf("0.0.0.0:%d", wellknownports.ProtokubeGossipWeaveMesh), "address:port on which to bind for gossip")
	flags.StringVar(&gossipSecret, "gossip-secret", gossipSecret, "Secret to use to secure gossip")
	flag.StringVar(&gossipProtocolSecondary, "gossip-protocol-secondary", "memberlist", "mesh/memberlist")
	flag.StringVar(&gossipListenSecondary, "gossip-listen-secondary", fmt.Sprintf("0.0.0.0:%d", wellknownports.ProtokubeGossipMemberlist), "address:port on which to bind for gossip")
	flags.StringVar(&gossipSecretSecondary, "gossip-secret-secondary", gossipSecret, "Secret to use to secure gossip")
	flags.StringSliceVarP(&zones, "zone", "z", []string{}, "Configure permitted zones and their mappings")

	bootstrapMasterNodeLabels := false
	flag.BoolVar(&bootstrapMasterNodeLabels, "bootstrap-master-node-labels", bootstrapMasterNodeLabels, "Bootstrap the labels for master nodes (required in k8s 1.16)")

	nodeName := ""
	flag.StringVar(&nodeName, "node-name", nodeName, "name of the node as will be created in kubernetes; used with bootstrap-master-node-labels")

	var removeDNSNames string
	flag.StringVar(&removeDNSNames, "remove-dns-names", removeDNSNames, "If set, will remove the DNS records specified")

	// Trick to avoid 'logging before flag.Parse' warning
	flag.CommandLine.Parse([]string{})

	flag.Set("logtostderr", "true")
	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	var volumes protokube.Volumes
	var internalIP net.IP

	if cloud == "aws" {
		awsVolumes, err := protokube.NewAWSVolumes()
		if err != nil {
			klog.Errorf("Error initializing AWS: %q", err)
			os.Exit(1)
		}
		volumes = awsVolumes
		internalIP = awsVolumes.InternalIP()

		if clusterID == "" {
			clusterID = awsVolumes.ClusterID()
		}

	} else if cloud == "digitalocean" {
		doVolumes, err := protokube.NewDOVolumes()
		if err != nil {
			klog.Errorf("Error initializing DigitalOcean: %q", err)
			os.Exit(1)
		}
		volumes = doVolumes
		internalIP, err = protokube.GetDropletInternalIP()
		if err != nil {
			klog.Errorf("Error getting droplet internal IP: %s", err)
			os.Exit(1)
		}

		if clusterID == "" {
			clusterID, err = protokube.GetClusterID()
			if err != nil {
				klog.Errorf("Error getting clusterid: %s", err)
				os.Exit(1)
			}
		}
	} else if cloud == "gce" {
		gceVolumes, err := protokube.NewGCEVolumes()
		if err != nil {
			klog.Errorf("Error initializing GCE: %q", err)
			os.Exit(1)
		}

		volumes = gceVolumes
		internalIP = gceVolumes.InternalIP()

		if clusterID == "" {
			clusterID = gceVolumes.ClusterID()
		}
	} else if cloud == "openstack" {
		klog.Info("Initializing openstack volumes")
		osVolumes, err := protokube.NewOpenstackVolumes()
		if err != nil {
			klog.Errorf("Error initializing openstack: %q", err)
			os.Exit(1)
		}
		volumes = osVolumes
		internalIP = osVolumes.InternalIP()

		if clusterID == "" {
			clusterID = osVolumes.ClusterID()
		}
	} else if cloud == "alicloud" {
		klog.Info("Initializing AliCloud volumes")
		aliVolumes, err := protokube.NewALIVolumes()
		if err != nil {
			klog.Errorf("Error initializing Aliyun: %q", err)
			os.Exit(1)
		}
		volumes = aliVolumes
		internalIP = aliVolumes.InternalIP()

		if clusterID == "" {
			clusterID = aliVolumes.ClusterID()
		}
	} else if cloud == "azure" {
		klog.Info("Initializing Azure volumes")
		azureVolumes, err := protokube.NewAzureVolumes()
		if err != nil {
			klog.Errorf("Error initializing Azure: %q", err)
			os.Exit(1)
		}
		volumes = azureVolumes
		internalIP = azureVolumes.InternalIP()

		if clusterID == "" {
			clusterID = azureVolumes.ClusterID()
		}
	} else {
		klog.Errorf("Unknown cloud %q", cloud)
		os.Exit(1)
	}

	if clusterID == "" {
		return fmt.Errorf("cluster-id is required (cannot be determined from cloud)")
	}
	klog.Infof("cluster-id: %s", clusterID)

	if internalIP == nil {
		klog.Errorf("Cannot determine internal IP")
		os.Exit(1)
	}

	if dnsInternalSuffix == "" {
		// TODO: Maybe only master needs DNS?
		dnsInternalSuffix = ".internal." + clusterID
		klog.Infof("Setting dns-internal-suffix to %q", dnsInternalSuffix)
	}

	// Make sure it's actually a suffix (starts with .)
	if !strings.HasPrefix(dnsInternalSuffix, ".") {
		dnsInternalSuffix = "." + dnsInternalSuffix
	}

	rootfs := "/"
	if containerized {
		rootfs = "/rootfs/"
	}

	protokube.RootFS = rootfs

	if gossip {
		dnsTarget := &gossipdns.HostsFile{
			Path: path.Join(rootfs, "etc/hosts"),
		}

		var gossipSeeds gossiputils.SeedProvider
		var err error
		var gossipName string
		if cloud == "aws" {
			gossipSeeds, err = volumes.(*protokube.AWSVolumes).GossipSeeds()
			if err != nil {
				return err
			}
			gossipName = volumes.(*protokube.AWSVolumes).InstanceID()
		} else if cloud == "gce" {
			gossipSeeds, err = volumes.(*protokube.GCEVolumes).GossipSeeds()
			if err != nil {
				return err
			}
			gossipName = volumes.(*protokube.GCEVolumes).InstanceName()
		} else if cloud == "openstack" {
			gossipSeeds, err = volumes.(*protokube.OpenstackVolumes).GossipSeeds()
			if err != nil {
				return err
			}
			gossipName = volumes.(*protokube.OpenstackVolumes).InstanceName()
		} else if cloud == "alicloud" {
			gossipSeeds, err = volumes.(*protokube.ALIVolumes).GossipSeeds()
			if err != nil {
				return err
			}
			gossipName = volumes.(*protokube.ALIVolumes).InstanceID()
		} else if cloud == "digitalocean" {
			gossipSeeds, err = volumes.(*protokube.DOVolumes).GossipSeeds()
			if err != nil {
				return err
			}
			gossipName = volumes.(*protokube.DOVolumes).InstanceName()
		} else if cloud == "azure" {
			gossipSeeds, err = volumes.(*protokube.AzureVolumes).GossipSeeds()
			if err != nil {
				return err
			}
			gossipName = volumes.(*protokube.AzureVolumes).InstanceID()
		} else {
			klog.Fatalf("seed provider for %q not yet implemented", cloud)
		}

		id := os.Getenv("HOSTNAME")
		if id == "" {
			klog.Warningf("Unable to fetch HOSTNAME for use as node identifier")
		}

		channelName := "dns"
		var gossipState gossiputils.GossipState

		gossipState, err = gossiputils.GetGossipState(gossipProtocol, gossipListen, channelName, gossipName, []byte(gossipSecret), gossipSeeds)
		if err != nil {
			klog.Errorf("Error initializing gossip: %v", err)
			os.Exit(1)
		}

		if gossipProtocolSecondary != "" {

			secondaryGossipState, err := gossiputils.GetGossipState(gossipProtocolSecondary, gossipListenSecondary, channelName, gossipName, []byte(gossipSecretSecondary), gossipSeeds)
			if err != nil {
				klog.Errorf("Error initializing secondary gossip: %v", err)
				os.Exit(1)
			}

			gossipState = &gossiputils.MultiGossipState{
				Primary:   gossipState,
				Secondary: secondaryGossipState,
			}
		}
		go func() {
			err := gossipState.Start()
			if err != nil {
				klog.Fatalf("gossip exited unexpectedly: %v", err)
			} else {
				klog.Fatalf("gossip exited unexpectedly, but without error")
			}
		}()

		dnsView := gossipdns.NewDNSView(gossipState)
		zoneInfo := gossipdns.DNSZoneInfo{
			Name: gossipdns.DefaultZoneName,
		}
		if _, err := dnsView.AddZone(zoneInfo); err != nil {
			klog.Fatalf("error creating zone: %v", err)
		}

		go func() {
			gossipdns.RunDNSUpdates(dnsTarget, dnsView)
			klog.Fatalf("RunDNSUpdates exited unexpectedly")
		}()
	}

	var channels []string
	if flagChannels != "" {
		channels = strings.Split(flagChannels, ",")
	}

	k := &protokube.KubeBoot{
		BootstrapMasterNodeLabels: bootstrapMasterNodeLabels,
		NodeName:                  nodeName,
		Channels:                  channels,
		InternalDNSSuffix:         dnsInternalSuffix,
		InternalIP:                internalIP,
		Kubernetes:                protokube.NewKubernetesContext(),
		Master:                    master,
	}

	k.RunSyncLoop()

	return fmt.Errorf("Unexpected exit")
}
