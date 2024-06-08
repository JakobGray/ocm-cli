package gcp

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/googleapis/gax-go/v2/apierror"

	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"github.com/openshift-online/ocm-cli/pkg/alpha_ocm"
	"github.com/openshift-online/ocm-cli/pkg/gcp"
	"github.com/openshift-online/ocm-cli/pkg/models"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/grpc/codes"
)

var (
	// CreateWorkloadIdentityPoolOpts captures the options that affect creation of the workload identity pool
	CreateWorkloadIdentityConfigurationOpts = options{
		Name:      "",
		Project:   "",
		TargetDir: "",
	}

	// The backend should provide this: https://issues.redhat.com/browse/OCM-8658
	impersonatorServiceAccount = "projects/sda-ccs-3/serviceAccounts/osd-impersonator@sda-ccs-3.iam.gserviceaccount.com"
)

const (
	poolDescription = "Created by the OCM CLI"

	// The backend should provide this: https://issues.redhat.com/browse/OCM-8658
	openShiftAudience = "openshift"
)

// NewCreateWorkloadIdentityConfiguration provides the "create-wif-config" subcommand
func NewCreateWorkloadIdentityConfiguration() *cobra.Command {
	createWorkloadIdentityPoolCmd := &cobra.Command{
		Use:              "create-wif-config",
		Short:            "Create workload identity configuration",
		Run:              createWorkloadIdentityConfigurationCmd,
		PersistentPreRun: validationForCreateWorkloadIdentityConfigurationCmd,
	}

	createWorkloadIdentityPoolCmd.PersistentFlags().StringVar(
		&CreateWorkloadIdentityConfigurationOpts.Name, "name", "", "User-defined name for all created Google cloud resources")
	createWorkloadIdentityPoolCmd.MarkPersistentFlagRequired("name")
	createWorkloadIdentityPoolCmd.PersistentFlags().StringVar(&CreateWorkloadIdentityConfigurationOpts.Project, "project", "", "ID of the Google cloud project")
	createWorkloadIdentityPoolCmd.MarkPersistentFlagRequired("project")
	createWorkloadIdentityPoolCmd.PersistentFlags().BoolVar(&CreateWorkloadIdentityConfigurationOpts.DryRun, "dry-run", false, "Skip creating objects, and just save what would have been created into files")
	createWorkloadIdentityPoolCmd.PersistentFlags().StringVar(&CreateWorkloadIdentityConfigurationOpts.TargetDir, "output-dir", "", "Directory to place generated files (defaults to current directory)")

	return createWorkloadIdentityPoolCmd
}

func createWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) {
	ctx := context.Background()

	gcpClient, err := gcp.NewGcpClient(ctx)
	if err != nil {
		log.Fatalf("failed to initiate GCP client: %v", err)
	}

	log.Println("Creating workload identity configuration...")
	wifConfig, err := createWorkloadIdentityConfiguration(models.WifConfigInput{
		DisplayName: CreateWorkloadIdentityConfigurationOpts.Name,
		ProjectId:   CreateWorkloadIdentityConfigurationOpts.Project,
	})
	if err != nil {
		log.Fatalf("failed to create WIF config: %v", err)
	}

	poolSpec := gcp.WorkloadIdentityPoolSpec{
		PoolName:               wifConfig.Status.WorkloadIdentityPoolData.PoolId,
		ProjectId:              wifConfig.Status.WorkloadIdentityPoolData.ProjectId,
		Jwks:                   wifConfig.Status.WorkloadIdentityPoolData.Jwks,
		IssuerUrl:              wifConfig.Status.WorkloadIdentityPoolData.IssuerUrl,
		PoolIdentityProviderId: wifConfig.Status.WorkloadIdentityPoolData.IdentityProviderId,
	}
	// Given the number of parameters, these helper functions may benefit from a "parameters" struct.
	// WDYT?
	if err = createWorkloadIdentityPool(ctx, gcpClient, poolSpec, CreateWorkloadIdentityConfigurationOpts.DryRun); err != nil {
		log.Fatalf("Failed to create workload identity pool: %s", err)
	}

	if err = createWorkloadIdentityProvider(ctx, gcpClient, poolSpec, CreateWorkloadIdentityConfigurationOpts.DryRun); err != nil {
		log.Fatalf("Failed to create workload identity provider: %s", err)
	}

	if err = createServiceAccounts(ctx, gcpClient, wifConfig, CreateWorkloadIdentityConfigurationOpts.DryRun); err != nil {
		log.Fatalf("Failed to create IAM service accounts: %s", err)
	}
}

func validationForCreateWorkloadIdentityConfigurationCmd(cmd *cobra.Command, argv []string) {
	if CreateWorkloadIdentityConfigurationOpts.Name == "" {
		// I don't think we should panic. Showing the user the backtrace is not
		// a good experience for users that are not developers.
		log.Fatal("Name is required")
	}
	if CreateWorkloadIdentityConfigurationOpts.Project == "" {
		log.Fatal("Project is required")
	}

	if CreateWorkloadIdentityConfigurationOpts.TargetDir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %s", err)
		}

		CreateWorkloadIdentityConfigurationOpts.TargetDir = pwd
	}

	fPath, err := filepath.Abs(CreateWorkloadIdentityConfigurationOpts.TargetDir)
	if err != nil {
		log.Fatalf("Failed to resolve full path: %s", err)
	}

	sResult, err := os.Stat(fPath)
	if os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist", fPath)
	}
	if !sResult.IsDir() {
		log.Fatalf("file %s exists and is not a directory", fPath)
	}
}

func createWorkloadIdentityConfiguration(input models.WifConfigInput) (*models.WifConfigOutput, error) {
	ocmClient, err := alphaocm.NewOcmClient()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create backend client")
	}
	output, err := ocmClient.CreateWifConfig(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wif config")
	}
	return &output, nil
}

func createWorkloadIdentityPool(ctx context.Context, client gcp.GcpClient, spec gcp.WorkloadIdentityPoolSpec, generateOnly bool) error {
	name := spec.PoolName
	project := spec.ProjectId
	if generateOnly {
		log.Printf("Would have created workload identity pool %s", name)
		// TODO gcloud command here. Can you create a tech-debt ticket for it?
		return nil
	}
	parentResourceForPool := fmt.Sprintf("projects/%s/locations/global", project)
	poolResource := fmt.Sprintf("%s/workloadIdentityPools/%s", parentResourceForPool, name)
	resp, err := client.GetWorkloadIdentityPool(ctx, poolResource)
	// Ok this is kinda wild syntax but I think it looks cleaner this way. I
	// want to get your thoughts on it: I think this has the advantage of less
	// nested if statements, and we are expressing that the execution flow is
	// parameterized.
	// Let me know what you think:
	switch {
	case err != nil:
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 && strings.Contains(gerr.Message, "Requested entity was not found") {
			pool := &iamv1.WorkloadIdentityPool{
				Name:        name,
				DisplayName: name,
				Description: poolDescription,
				State:       "ACTIVE",
				Disabled:    false,
			}

			_, err := client.CreateWorkloadIdentityPool(ctx, parentResourceForPool, name, pool)
			if err != nil {
				return errors.Wrapf(err, "failed to create workload identity pool %s", name)
			}
			log.Printf("Workload identity pool created with name %s", name)
		} else {
			return errors.Wrapf(err, "failed to check if there is existing workload identity pool %s", name)
		}
		return nil
	case resp != nil && resp.State == "DELETED":
		log.Printf("Workload identity pool %s was deleted", name)
		_, err := client.UndeleteWorkloadIdentityPool(ctx, poolResource, &iamv1.UndeleteWorkloadIdentityPoolRequest{})
		if err != nil {
			return errors.Wrapf(err, "failed to undelete workload identity pool %s", name)
		}
		return nil
	default:
		log.Printf("Workload identity pool %s already exists", name)
		return nil
	}
}

func createWorkloadIdentityProvider(ctx context.Context, client gcp.GcpClient, spec gcp.WorkloadIdentityPoolSpec, generateOnly bool) error {
	if generateOnly {
		log.Printf("Would have created workload identity provider for %s with issuerURL %s", spec.PoolName, spec.IssuerUrl)
		return nil
	}
	providerResource := fmt.Sprintf("projects/%s/locations/global/workloadIdentityPools/%s/providers/%s", spec.ProjectId, spec.PoolName, spec.PoolName)
	_, err := client.GetWorkloadIdentityProvider(ctx, providerResource)
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 && strings.Contains(gerr.Message, "Requested entity was not found") {
			provider := &iam.WorkloadIdentityPoolProvider{
				Name:        spec.PoolName,
				DisplayName: spec.PoolName,
				Description: poolDescription,
				State:       "ACTIVE",
				Disabled:    false,
				Oidc: &iam.Oidc{
					AllowedAudiences: []string{openShiftAudience},
					IssuerUri:        spec.IssuerUrl,
					JwksJson:         spec.Jwks,
				},
				AttributeMapping: map[string]string{
					// TODO: This will come from the spec
					// when token exchange happens, sub from oidc token shared by operator pod will be mapped to google.subject
					// field of google auth token. The field is used to allow fine-grained access to gcp service accounts.
					// The format is `system:serviceaccount:<service_account_namespace>:<service_account_name>`
					"google.subject": "assertion.sub",
				},
			}

			_, err := client.CreateWorkloadIdentityProvider(ctx, fmt.Sprintf("projects/%s/locations/global/workloadIdentityPools/%s", spec.ProjectId, spec.PoolName), spec.PoolName, provider)
			if err != nil {
				return errors.Wrapf(err, "failed to create workload identity provider %s", spec.PoolName)
			}
			log.Printf("workload identity provider created with name %s", spec.PoolName)
		} else {
			return errors.Wrapf(err, "failed to check if there is existing workload identity provider %s in pool %s", spec.PoolName, spec.PoolName)
		}
	} else {
		log.Printf("Workload identity provider %s already exists in pool %s", spec.PoolName, spec.PoolName)
	}
	return nil
}

func createServiceAccounts(ctx context.Context, gcpClient gcp.GcpClient, wifOutput *models.WifConfigOutput, generateOnly bool) error {
	projectId := wifOutput.Spec.ProjectId
	fmtRoleResourceId := func(role models.Role) string {
		return fmt.Sprintf("roles/%s", role.Id)
	}
	if generateOnly {
		for _, serviceAccount := range wifOutput.Status.ServiceAccounts {
			serviceAccountID := serviceAccount.GetId()
			log.Printf("Would have created service account %s", serviceAccountID)
			log.Printf("Would have bound roles to %s", serviceAccountID)
			switch serviceAccount.AccessMethod {
			case "impersonate":
				log.Printf("Would have attached impersonator %s to %s", impersonatorServiceAccount, serviceAccountID)
			case "wif":
				log.Printf("Would have attached workload identity pool %s to %s", wifOutput.Status.WorkloadIdentityPoolData.IdentityProviderId, serviceAccountID)
			default:
				fmt.Printf("Warning: %s is not a supported access type\n", serviceAccount.AccessMethod)
			}
		}
		return nil
	}

	// Create service accounts
	for _, serviceAccount := range wifOutput.Status.ServiceAccounts {
		serviceAccountID := serviceAccount.GetId()
		serviceAccountName := wifOutput.Spec.DisplayName + "-" + serviceAccountID
		serviceAccountDesc := poolDescription + " for WIF config " + wifOutput.Spec.DisplayName

		// TODO output cleanup
		fmt.Println("Creating service account", serviceAccountID)
		_, err := CreateServiceAccount(gcpClient, serviceAccountID, serviceAccountName, serviceAccountDesc, projectId, true)
		if err != nil {
			return errors.Wrap(err, "Failed to create IAM service account")
		}
		log.Printf("IAM service account %s created", serviceAccountID)
	}

	// Bind roles and grant access
	for _, serviceAccount := range wifOutput.Status.ServiceAccounts {
		serviceAccountID := serviceAccount.GetId()

		fmt.Printf("\t\tBinding roles to %s\n", serviceAccount.Id)
		for _, role := range serviceAccount.Roles {
			if !role.Predefined {
				fmt.Printf("Skipping role %q for service account %q as custom roles are not yet supported.", role.Id, serviceAccount.Id)
				continue
			}
			err := gcpClient.BindRole(serviceAccountID, projectId, fmtRoleResourceId(role))
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("\t\tRoles bound to %s\n", serviceAccount.Id)

		fmt.Printf("\t\tGranting access to %s...\n", serviceAccount.Id)
		switch serviceAccount.AccessMethod {
		case "impersonate":
			if err := gcpClient.AttachImpersonator(serviceAccount.Id, projectId, impersonatorServiceAccount); err != nil {
				panic(err)
			}
		case "wif":
			if err := gcpClient.AttachWorkloadIdentityPool(serviceAccount, wifOutput.Status.WorkloadIdentityPoolData.PoolId, projectId); err != nil {
				panic(err)
			}
		default:
			fmt.Printf("Warning: %s is not a supported access type\n", serviceAccount.AccessMethod)
		}
		fmt.Printf("\t\tAccess granted to %s\n", serviceAccount.Id)
	}
	return nil
}

func CreateServiceAccount(gcpClient gcp.GcpClient, svcAcctID, svcAcctName, svcAcctDescription, projectName string, allowExisting bool) (*adminpb.ServiceAccount, error) {
	request := &adminpb.CreateServiceAccountRequest{
		Name:      fmt.Sprintf("projects/%s", projectName),
		AccountId: svcAcctID,
		ServiceAccount: &adminpb.ServiceAccount{
			DisplayName: svcAcctName,
			Description: svcAcctDescription,
		},
	}
	svcAcct, err := gcpClient.CreateServiceAccount(context.TODO(), request)
	if err != nil {
		pApiError, ok := err.(*apierror.APIError)
		if ok {
			if pApiError.GRPCStatus().Code() == codes.AlreadyExists && allowExisting {
				return svcAcct, nil
			}
		}
	}
	return svcAcct, err
}
