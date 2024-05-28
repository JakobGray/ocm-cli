package gcp

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/openshift-online/ocm-cli/cmd/ocm/gcp/mock"

	"github.com/openshift-online/ocm-cli/pkg/urls"
	"github.com/spf13/cobra"
)

// NewDescribeWorkloadIdentityConfiguration provides the "describe-wif-config" subcommand
func NewDescribeWorkloadIdentityConfiguration() *cobra.Command {
	describeWorkloadIdentityPoolCmd := &cobra.Command{
		Use:              "describe-wif-config [ID]",
		Short:            "Show details of a wif-config.",
		Run:              describeWorkloadIdentityConfigurationCmd,
		PersistentPreRun: validationForDescribeWorkloadIdentityConfigurationCmd,
	}

	return describeWorkloadIdentityPoolCmd
}

func describeWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) {
	id, err := urls.Expand(argv)
	if err != nil {
		log.Fatalf("could not create URI: %v", err)
	}

	if id != "test01" {
		log.Fatalf("failed to find WIF Config with id: %s", id)
	}
	wifconfig := mock.MockWifConfig("test01", id)

	// Print output
	w := tabwriter.NewWriter(os.Stdout, 8, 0, 2, ' ', 0)

	fmt.Fprintf(w, "ID:\t%s\n", wifconfig.Metadata.Id)
	fmt.Fprintf(w, "Display Name:\t%s\n", wifconfig.Metadata.DisplayName)
	fmt.Fprintf(w, "Project:\t%s\n", wifconfig.Spec.ProjectId)
	fmt.Fprintf(w, "State:\t%s\n", wifconfig.Status.State)
	fmt.Fprintf(w, "Summary:\t%s\n", wifconfig.Status.Summary)
	fmt.Fprintf(w, "Issuer URL:\t%s\n", wifconfig.Status.WorkloadIdentityPoolData.IssuerUrl)

	w.Flush()
}

func validationForDescribeWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) {
	if len(argv) != 1 {
		log.Fatalf("Expected exactly one command line parameters containing the id of the WIF config.")
	}
}
