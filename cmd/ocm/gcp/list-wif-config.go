package gcp

import (
	"context"
	"log"
	"os"

	"github.com/openshift-online/ocm-cli/cmd/ocm/gcp/mock"
	"github.com/openshift-online/ocm-cli/pkg/config"
	"github.com/openshift-online/ocm-cli/pkg/models"
	"github.com/openshift-online/ocm-cli/pkg/ocm"
	"github.com/openshift-online/ocm-cli/pkg/output"
	"github.com/spf13/cobra"
)

var ListWorkloadIdentityConfigurationOpts struct {
	columns   string
	noHeaders bool
}

// NewCreateWorkloadIdentityConfiguration provides the "create-wif-config" subcommand
func NewListWorkloadIdentityConfiguration() *cobra.Command {
	listWorkloadIdentityPoolCmd := &cobra.Command{
		Use:              "list-wif-config [ID]",
		Short:            "List wif-configs.",
		Run:              listWorkloadIdentityConfigurationCmd,
		PersistentPreRun: validationForListWorkloadIdentityConfigurationCmd,
	}

	fs := listWorkloadIdentityPoolCmd.Flags()
	fs.StringVar(
		&ListWorkloadIdentityConfigurationOpts.columns,
		"columns",
		"metadata.id, metadata.displayName, status.state",
		"Specify which columns to display separated by commas, path is based on wif-config struct",
	)
	fs.BoolVar(
		&ListWorkloadIdentityConfigurationOpts.noHeaders,
		"no-headers",
		false,
		"Don't print header row",
	)

	return listWorkloadIdentityPoolCmd
}

func validationForListWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) {
	// No validation needed
}

func listWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) {
	// Create a context:
	ctx := context.Background()

	// Load the configuration:
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Create the client for the OCM API:
	connection, err := ocm.NewConnection().Build()
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	// Create the output printer:
	printer, err := output.NewPrinter().
		Writer(os.Stdout).
		Pager(cfg.Pager).
		Build(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer printer.Close()

	wifconfigs, err := getMockWifConfigs()
	if err != nil {
		log.Fatalf("failed to get wif-configs: %v", err)
	}

	// Create the output table:
	table, err := printer.NewTable().
		Name("wifconfigs").
		Columns(ListWorkloadIdentityConfigurationOpts.columns).
		Build(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer table.Close()

	// Unless noHeaders set, print header row:
	if !ListWorkloadIdentityConfigurationOpts.noHeaders {
		table.WriteHeaders()
	}

	// Write the rows:
	for _, wc := range wifconfigs {
		err = table.WriteObject(wc)
		if err != nil {
			break
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func getMockWifConfigs() ([]models.WifConfigOutput, error) {
	return []models.WifConfigOutput{
		*mock.MockWifConfig("test01", "0001"),
		*mock.MockWifConfig("test02", "0002"),
	}, nil
}
