package gcp

import (
	"context"
	"fmt"

	"log"

	"github.com/openshift-online/ocm-cli/cmd/ocm/gcp/mock"
	"github.com/openshift-online/ocm-cli/pkg/gcp"
	"github.com/openshift-online/ocm-cli/pkg/models"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	// DeleteWorkloadIdentityConfigurationOpts captures the options that affect creation of the workload identity pool
	DeleteWorkloadIdentityConfigurationOpts = options{
		Name:      "",
		Project:   "",
		TargetDir: "",
	}
)

// NewCreateWorkloadIdentityConfiguration provides the "create-wif-config" subcommand
func NewDeleteWorkloadIdentityConfiguration() *cobra.Command {
	deleteWorkloadIdentityPoolCmd := &cobra.Command{
		Use:              "delete-wif-config [ID]",
		Short:            "Delete workload identity configuration",
		Run:              deleteWorkloadIdentityConfigurationCmd,
		PersistentPreRun: validationForDeleteWorkloadIdentityConfigurationCmd,
	}

	return deleteWorkloadIdentityPoolCmd
}

func validationForDeleteWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) {
	if len(argv) != 1 {
		log.Fatal(
			"Expected exactly one command line parameters containing the id " +
				"of the WIF config.",
		)
	}
}

func deleteWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) {
	ctx := context.Background()
	fmt.Println("Deleting wif-config...")

	wifConfigId := argv[0]
	if wifConfigId == "" {
		log.Fatal("WIF config ID is required")
	}

	wifConfig := getWifConfig(wifConfigId)

	gcpClient, err := gcp.NewGcpClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err := deleteServiceAccounts(ctx, gcpClient, wifConfig); err != nil {
		log.Fatal(err)
	}

	if err := deleteWorkloadIdentityPool(ctx, gcpClient, wifConfig); err != nil {
		log.Fatal(err)
	}

	err = deleteWifConfig(wifConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteWifConfig(wifConfig *models.WifConfigOutput) error {
	fmt.Println("Deleting WIF config", wifConfig.Spec.DisplayName)
	return nil
}

func getWifConfig(wifConfigId string) *models.WifConfigOutput {
	return mock.MockWifConfig("test01", wifConfigId)
}

func deleteServiceAccounts(ctx context.Context, gcpClient gcp.GcpClient, wifConfig *models.WifConfigOutput) error {
	log.Println("Deleting service accounts...")
	projectId := wifConfig.Spec.ProjectId

	for _, serviceAccount := range wifConfig.Status.ServiceAccounts {
		serviceAccountID := generateServiceAccountID(serviceAccount)
		log.Println("Deleting service account", serviceAccountID)
		err := gcpClient.DeleteServiceAccount(serviceAccountID, projectId, true)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func deleteWorkloadIdentityPool(ctx context.Context, gcpClient gcp.GcpClient, wifConfig *models.WifConfigOutput) error {
	log.Println("Deleting workload identity pool...")
	projectId := wifConfig.Spec.ProjectId
	poolName := wifConfig.Status.WorkloadIdentityPoolData.PoolId
	poolResource := fmt.Sprintf("projects/%s/locations/global/workloadIdentityPools/%s", projectId, poolName)

	_, err := gcpClient.DeleteWorkloadIdentityPool(ctx, poolResource)
	if err != nil {
		return errors.Wrapf(err, "Failed to delete workload identity pool %s", poolName)
	}

	log.Printf("Workload identity pool %s deleted", poolName)
	return nil
}

func generateServiceAccountID(serviceAccount models.ServiceAccount) string {
	serviceAccountID := "z-" + serviceAccount.Id
	if len(serviceAccountID) > 30 {
		serviceAccountID = serviceAccountID[:30]
	}
	return serviceAccountID
}
