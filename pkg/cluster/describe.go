/*
Copyright (c) 2019 Red Hat, Inc.

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

package cluster

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const (
	notAvailable string = "N/A"
)

func PrintClusterDescription(connection *sdk.Connection, cluster *cmv1.Cluster) error {
	// Get API URL:
	api := cluster.API()
	apiURL, _ := api.GetURL()
	apiListening := api.Listening()

	// Retrieve the details of the subscription:
	var sub *amv1.Subscription
	subID := cluster.Subscription().ID()
	if subID != "" {
		subResponse, err := connection.AccountsMgmt().V1().
			Subscriptions().
			Subscription(subID).
			//nolint
			Get().Parameter("fetchLabels", "true").
			Send()
		if err != nil {
			if subResponse == nil || subResponse.Status() != 404 {
				return fmt.Errorf(
					"can't get subscription '%s': %v",
					subID, err,
				)
			}
		}
		sub = subResponse.Body()
	}

	// Retrieve the details of the account:
	var account *amv1.Account
	accountID := sub.Creator().ID()
	if accountID != "" {
		accountResponse, err := connection.AccountsMgmt().V1().
			Accounts().
			Account(accountID).
			Get().
			Send()
		if err != nil {
			if accountResponse == nil || (accountResponse.Status() != 404 &&
				accountResponse.Status() != 403) {
				return fmt.Errorf(
					"can't get account '%s': %v",
					accountID, err,
				)
			}
		}
		account = accountResponse.Body()
	}

	// Find the details of the creator:
	organization := notAvailable
	if account.Organization() != nil && account.Organization().Name() != "" {
		organization = account.Organization().Name()
	}

	creator := account.Username()
	if creator == "" {
		creator = notAvailable
	}

	email := account.Email()
	if email == "" {
		email = notAvailable
	}

	accountNumber := account.Organization().EbsAccountID()
	if accountNumber == "" {
		accountNumber = notAvailable
	}

	// Find the details of the shard
	shardPath, err := connection.ClustersMgmt().V1().Clusters().
		Cluster(cluster.ID()).
		ProvisionShard().
		Get().
		Send()
	var shard string
	if shardPath != nil && err == nil {
		shard = shardPath.Body().HiveConfig().Server()
	}

	clusterAdminEnabled := false
	if cluster.CCS().Enabled() {
		clusterAdminEnabled = true
	} else {
		for _, label := range sub.Labels() {
			if label.Key() == "capability.cluster.manage_cluster_admin" &&
				//nolint
				label.Value() == "true" {
				clusterAdminEnabled = true
			}
		}
	}

	privateLinkEnabled := false
	stsEnabled := false
	// Setting isExistingVPC to unsupported to avoid confusion
	// when looking at clusters on other providers than AWS
	isExistingVPC := "unsupported"
	if cluster.CloudProvider().ID() == ProviderAWS && cluster.AWS() != nil {
		privateLinkEnabled = cluster.AWS().PrivateLink()
		if cluster.AWS().STS().RoleARN() != "" {
			stsEnabled = true
		}

		isExistingVPC = "false"
		if cluster.AWS().SubnetIDs() != nil && len(cluster.AWS().SubnetIDs()) > 0 {
			//nolint
			isExistingVPC = "true"
		}
	}

	if cluster.CloudProvider().ID() == ProviderGCP &&
		cluster.GCPNetwork().VPCName() != "" && cluster.GCPNetwork().ControlPlaneSubnet() != "" &&
		cluster.GCPNetwork().ComputeSubnet() != "" {
		//nolint
		isExistingVPC = "true"
	}

	// Parse Hypershift-related values
	mgmtClusterName, svcClusterName := findHyperShiftMgmtSvcClusters(connection, cluster)

	provisioningStatus := ""
	if cluster.Status().State() == cmv1.ClusterStateError && cluster.Status().ProvisionErrorCode() != "" {
		provisioningStatus = fmt.Sprintf("(%s - %s)",
			cluster.Status().ProvisionErrorCode(),
			cluster.Status().ProvisionErrorMessage(),
		)
	}

	// var computesStr string
	// if cluster.Nodes().AutoscaleCompute() != nil {
	// 	computesStr = fmt.Sprintf("%d-%d (Autoscaled)",
	// 		cluster.Nodes().AutoscaleCompute().MinReplicas(),
	// 		cluster.Nodes().AutoscaleCompute().MaxReplicas(),
	// 	)
	// } else {
	// 	computesStr = strconv.Itoa(cluster.Nodes().Compute())
	// }

	// Print output
	w := tabwriter.NewWriter(os.Stdout, 8, 0, 2, ' ', 0)

	fmt.Fprintf(w, "ID:\t%s\n", cluster.ID())
	fmt.Fprintf(w, "External ID:\t%s\n", cluster.ExternalID())
	fmt.Fprintf(w, "Name:\t%s\n", cluster.Name())
	fmt.Fprintf(w, "Display Name:\t%s\n", sub.DisplayName())
	fmt.Fprintf(w, "State:\t%s %s\n", cluster.State(), provisioningStatus)

	if cluster.Status().Description() != "" {
		fmt.Fprintf(w, "Details:\t%s\n", cluster.Status().Description())
	}

	printNodeInfo(w, cluster)
	// fmt.Fprintf(w, "Control Plane:\n%s\n", printNodeInfo(strconv.Itoa(cluster.Nodes().Master()), cluster.AWS().AdditionalControlPlaneSecurityGroupIds()))
	// fmt.Fprintf(w, "Infra:\n%s\n", printNodeInfo(strconv.Itoa(cluster.Nodes().Infra()), cluster.AWS().AdditionalInfraSecurityGroupIds()))
	// fmt.Fprintf(w, "Compute:\n%s\n", printNodeInfo(computesStr, []string{}))

	fmt.Fprintf(w, "API URL:\t%s\n", apiURL)
	fmt.Fprintf(w, "API Listening:\t%s\n", apiListening)
	fmt.Fprintf(w, "Console URL:\t%s\n", cluster.Console().URL())
	fmt.Fprintf(w, "Product:\t%s\n", cluster.Product().ID())
	fmt.Fprintf(w, "Subscription type:\t%s\n", cluster.BillingModel())
	fmt.Fprintf(w, "Provider:\t%s\n", cluster.CloudProvider().ID())
	fmt.Fprintf(w, "Version:\t%s\n", cluster.OpenshiftVersion())
	fmt.Fprintf(w, "Region:\t%s\n", cluster.Region().ID())
	fmt.Fprintf(w, "Multi-az:\t%t\n", cluster.MultiAZ())

	// GCP-specific info
	if cluster.CloudProvider().ID() == ProviderGCP {
		if cluster.GCP().Security().SecureBoot() {
			fmt.Fprintf(w, "SecureBoot:\t%t\n", cluster.GCP().Security().SecureBoot())
		}

		if cluster.GCPNetwork().VPCName() != "" {
			fmt.Fprintf(w, "VPC-Name:\t%s\n", cluster.GCPNetwork().VPCName())
		}
		if cluster.GCPNetwork().ControlPlaneSubnet() != "" {
			fmt.Fprintf(w, "Control-Plane-Subnet:\t%s\n", cluster.GCPNetwork().ControlPlaneSubnet())
		}
		if cluster.GCPNetwork().ComputeSubnet() != "" {
			fmt.Fprintf(w, "Compute-Subnet:\t%s\n", cluster.GCPNetwork().ComputeSubnet())
		}
	}

	// AWS-specific info
	if cluster.CloudProvider().ID() == ProviderAWS {
		fmt.Fprintf(w, "PrivateLink:\t%t\n", privateLinkEnabled)
		fmt.Fprintf(w, "STS:\t%t\n", stsEnabled)
	}

	fmt.Fprintf(w, "CCS:\t%t\n", cluster.CCS().Enabled())
	fmt.Fprintf(w, "HCP:\t%t\n", cluster.Hypershift().Enabled())
	fmt.Fprintf(w, "Subnet IDs:\t%s\n", cluster.AWS().SubnetIDs())
	fmt.Fprintf(w, "Existing VPC:\t%s\n", isExistingVPC)
	fmt.Fprintf(w, "Channel Group:\t%v\n", cluster.Version().ChannelGroup())
	fmt.Fprintf(w, "Cluster Admin:\t%t\n", clusterAdminEnabled)
	fmt.Fprintf(w, "Organization:\t%s\n", organization)
	fmt.Fprintf(w, "Creator:\t%s\n", creator)
	fmt.Fprintf(w, "Email:\t%s\n", email)
	fmt.Fprintf(w, "AccountNumber:\t%s\n", accountNumber)
	fmt.Fprintf(w, "Created:\t%v\n", cluster.CreationTimestamp().Round(time.Second).Format(time.RFC3339Nano))

	expirationTime, hasExpirationTimestamp := cluster.GetExpirationTimestamp()
	if hasExpirationTimestamp {
		fmt.Fprintf(w, "Expiration:\t%v\n", expirationTime.Round(time.Second).Format(time.RFC3339Nano))
	}

	// Hive
	if shard != "" {
		fmt.Fprintf(w, "Shard:\t%v\n", shard)
	}

	// HyperShift (should be mutually exclusive with Hive)
	if mgmtClusterName != "" {
		fmt.Fprintf(w, "Management Cluster:\t%s\n", mgmtClusterName)
	}
	if svcClusterName != "" {
		fmt.Fprintf(w, "Service Cluster:\t%s\n", svcClusterName)
	}

	// Cluster-wide-proxy
	if cluster.Proxy().HTTPProxy() != "" {
		fmt.Fprintf(w, "HTTPProxy:\t%s\n", cluster.Proxy().HTTPProxy())
	}
	if cluster.Proxy().HTTPSProxy() != "" {
		fmt.Fprintf(w, "HTTPSProxy:\t%s\n", cluster.Proxy().HTTPSProxy())
	}
	if cluster.Proxy().NoProxy() != "" {
		fmt.Fprintf(w, "NoProxy:\t%s\n", cluster.Proxy().NoProxy())
	}
	if cluster.AdditionalTrustBundle() != "" {
		fmt.Fprintf(w, "AdditionalTrustBundle:\t%s\n", cluster.AdditionalTrustBundle())
	}

	// Limited Support Status
	if cluster.Status().LimitedSupportReasonCount() > 0 {
		fmt.Fprintf(w, "Limited Support:\t%t\n", cluster.Status().LimitedSupportReasonCount() > 0)
	}

	w.Flush()
	return nil
}

func printNodeInfo(w io.Writer, cluster *cmv1.Cluster) {
	fmt.Fprintf(w, "Control Plane:\n")
	fmt.Fprintf(w, "    Replicas:\t%v\n", strconv.Itoa(cluster.Nodes().Master()))
	// securityGroups := cluster.AWS().AdditionalControlPlaneSecurityGroupIds()
	securityGroups := []string{"a", "b"}
	if len(securityGroups) > 0 {
		fmt.Fprintf(w, "    AWS Additional Security Group IDs:\t%s\n", strings.Join(securityGroups, ", "))
	}

	// fmt.Fprintf(w, "Infra:\n")

	// fmt.Fprintf(w, "Compute:\n")

	// fmt.Fprintf(w, "Control Plane:\n", printNodeInfo(strconv.Itoa(cluster.Nodes().Master()), cluster.AWS().AdditionalControlPlaneSecurityGroupIds()))
	// fmt.Fprintf(w, "Infra:\n%s\n", printNodeInfo(strconv.Itoa(cluster.Nodes().Infra()), cluster.AWS().AdditionalInfraSecurityGroupIds()))
	// fmt.Fprintf(w, "Compute:\n%s\n", printNodeInfo(computesStr, []string{}))

	// nodeStr := fmt.Sprintf("\tReplicas: %s", replicasInfo)
	// if len(securityGroups) > 0 {
	// 	nodeStr += fmt.Sprintf("\n\tAWS Additional Security Group IDs: %s", strings.Join(securityGroups, ", "))
	// }
	// return nodeStr
}

// findHyperShiftMgmtSvcClusters returns the name of a HyperShift cluster's management and service clusters.
// It essentially ignores error as these endpoint is behind specific permissions by returning empty strings when any
// errors are encountered, which results in them not being printed in the output.
func findHyperShiftMgmtSvcClusters(conn *sdk.Connection, cluster *cmv1.Cluster) (string, string) {
	if !cluster.Hypershift().Enabled() {
		return "", ""
	}

	hypershiftResp, err := conn.ClustersMgmt().V1().Clusters().
		Cluster(cluster.ID()).
		Hypershift().
		Get().
		Send()
	if err != nil {
		return "", ""
	}

	mgmtClusterName := hypershiftResp.Body().ManagementCluster()
	fmMgmtResp, err := conn.OSDFleetMgmt().V1().ManagementClusters().
		List().
		Parameter("search", fmt.Sprintf("name='%s'", mgmtClusterName)).
		Send()
	if err != nil {
		return mgmtClusterName, ""
	}

	if kind := fmMgmtResp.Items().Get(0).Parent().Kind(); kind == "ServiceCluster" {
		return mgmtClusterName, fmMgmtResp.Items().Get(0).Parent().Name()
	}

	// Shouldn't normally happen as every management cluster should have a service cluster
	return mgmtClusterName, ""
}
