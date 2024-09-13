package gcp

import (
	"context"
	"fmt"
	"log"

	"github.com/openshift-online/ocm-cli/pkg/gcp"
	"github.com/openshift-online/ocm-cli/pkg/ocm"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var UpdateWifConfigOpts struct {
}

// NewUpdateWorkloadIdentityConfiguration provides the "gcp update wif-config" subcommand
func NewUpdateWorkloadIdentityConfiguration() *cobra.Command {
	updateWifConfigCmd := &cobra.Command{
		Use:     "wif-config [ID|Name]",
		Short:   "Update wif-config.",
		RunE:    updateWorkloadIdentityConfigurationCmd,
		PreRunE: validationForUpdateWorkloadIdentityConfigurationCmd,
	}

	return updateWifConfigCmd
}

func validationForUpdateWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) error {
	if len(argv) != 1 {
		return fmt.Errorf("Expected exactly one command line parameters containing the id of the WIF config")
	}
	return nil
}

func updateWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) error {
	ctx := context.Background()
	log := log.Default()
	id := argv[0]

	// Create the client for the OCM API:
	connection, err := ocm.NewConnection().Build()
	if err != nil {
		return errors.Wrapf(err, "Failed to create OCM connection")
	}

	// Verify the WIF configuration exists
	wifconfig, err := findWifConfig(connection.ClustersMgmt().V1(), id)
	if err != nil {
		return errors.Wrapf(err, "failed to get wif-config")
	}

	gcpClient, err := gcp.NewGcpClient(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to initiate GCP client")
	}

	// Re-apply WIF resources
	gcpClientWifConfigShim := NewGcpClientWifConfigShim(GcpClientWifConfigShimSpec{
		GcpClient: gcpClient,
		WifConfig: wifConfig,
	})

	if err := gcpClientWifConfigShim.GrantSupportAccess(ctx, log); err != nil {
		log.Printf("Failed to grant support access to project: %s", err)
		return fmt.Errorf("To clean up, run the following command: ocm gcp delete wif-config %s", wifConfig.ID())
	}

	if err := gcpClientWifConfigShim.CreateWorkloadIdentityPool(ctx, log); err != nil {
		log.Printf("Failed to create workload identity pool: %s", err)
		return fmt.Errorf("To clean up, run the following command: ocm gcp delete wif-config %s", wifConfig.ID())
	}

	if err = gcpClientWifConfigShim.CreateWorkloadIdentityProvider(ctx, log); err != nil {
		log.Printf("Failed to create workload identity provider: %s", err)
		return fmt.Errorf("To clean up, run the following command: ocm gcp delete wif-config %s", wifConfig.ID())
	}

	if err = gcpClientWifConfigShim.CreateServiceAccounts(ctx, log); err != nil {
		log.Printf("Failed to create IAM service accounts: %s", err)
		return fmt.Errorf("To clean up, run the following command: ocm gcp delete wif-config %s", wifConfig.ID())
	}

	return nil
}

// findWifConfig finds the WIF configuration by ID or name
func findWifConfig(client *cmv1.Client, key string) (*cmv1.WifConfig, error) {
	collection := client.GCP().WifConfigs()
	page := 1
	size := 1
	query := fmt.Sprintf(
		"id = '%s' or display_name = '%s'",
		key, key,
	)

	response, err := collection.List().Search(query).Page(page).Size(size).Send()
	if err != nil {
		return nil, err
	}
	if response.Total() == 0 {
		return nil, fmt.Errorf("WIF configuration with identifier or name '%s' not found", key)
	}
	if response.Total() > 1 {
		return nil, fmt.Errorf("there are %d WIF configurations found with identifier or name '%s'", response.Total(), key)
	}
	return response.Items().Slice()[0], nil
}
