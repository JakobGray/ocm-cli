package mock

import (
	"github.com/openshift-online/ocm-cli/pkg/models"
)

func MockWifConfig(name, id string) *models.WifConfigOutput {
	return &models.WifConfigOutput{
		Metadata: &models.WifConfigMetadata{
			DisplayName:  name,
			Id:           id,
			Organization: &models.WifConfigMetadataOrganization{},
		},
		Spec: &models.WifConfigInput{
			DisplayName: name,
			ProjectId:   "sda-ccs-1",
		},
		Status: &models.WifConfigStatus{
			State:    "UNKNOWN",
			Summary:  "",
			TimeData: models.WifTimeData{},
			WorkloadIdentityPoolData: models.WifWorkloadIdentityPoolData{
				IdentityProviderId: "oidc",
				IssuerUrl:          "https://fake-issuer.com",
				Jwks:               "{\"keys\":[{\"use\":\"sig\",\"kty\":\"RSA\",\"kid\":\"hZ8yVW_SaRg3zSLz3lZ5D65rHNageE7agstatYTVesA\",\"alg\":\"RS256\",\"n\":\"rKN70G7r_O2untSKxpO2uQQatadzk3d5_0hs_FQGCa_xx0pBRoBMUvdbS4A8quZEoJOk8jmwSGWKvlLk-Is_Rsk0lq3hfKuElYFq43KOHe5YZfmZV7lqCdDr_lIJS2gzjfe_8xTfOUpU_DC0Slrj7UvwpcmpFMtklgeN_LLwuRRD6nKQRx0n0ABBXs44D2s4THrSABvpuOQzBpx9Qx8ShMPdN5h_aePcQj9Ss-lOAqdUprGLU-O-Cm2STUCtHFLh8oB5dYZA2Ww-6RkX7LIj6cVbFwMMZAwLn4ObKE_r6s8yynfS6wvqOldk-pkC0wtQx7R8NoP3ff9RF3kzKrkegdXAkVewUYu6V2Vcl1Z07MmyYcUhcjxT9jOblWklHSYIgyi_n-p30dzL7avU8sSdGnhFrLh8p9d3A1o_g1JNnSTquIHIpyE5WKZyCTF-c2K6VOSRWz16RgI5pswW6-IquJ9g-1RrAsTtuInueYJJw4S32OKRcsanClBKtCg95g1ylZL8UfeQS9S5Q7VNxCUdW7pgtIZiLcPIP8-Ier0yJF5TjnnVwcWs0Rq4-zWV-JyjoqUh6zjR9tqgXdSl1IBl6MqiHGNlW9SF0kRmgtAzm_fwlvqCaEzMW-LEivTUBgwHAvGYftciHFalqbblUNYgJyAy2hZ4-17JJAII8Yk5mZs\",\"e\":\"AQAB\"}]}",
				PoolId:             "d7a2fb1099634da1a9d0a5e8d1c11c39",
				ProjectId:          "sda-ccs-1",
				ProjectNumber:      1008983090557,
			},
			ServiceAccounts: []models.ServiceAccount{
				{
					Kind:         "ServiceAccount",
					Id:           "osd-deployer",
					OsdRole:      "deployer",
					AccessMethod: "impersonate",
					Roles: []models.Role{
						{Kind: "Role", Id: "compute.admin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "dns.admin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.roleAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.securityAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.serviceAccountAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.ServiceAccountKeyAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.ServiceAccountUser", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "storage.admin", Predefined: true, Permissions: []string{}},
					},
				},
				{
					Kind:         "ServiceAccount",
					Id:           "osd-support",
					OsdRole:      "support",
					AccessMethod: "impersonate",
					Roles: []models.Role{
						{Kind: "Role", Id: "compute.admin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "dns.admin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.roleAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.securityAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.serviceAccountAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.ServiceAccountKeyAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.ServiceAccountUser", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "storage.admin", Predefined: true, Permissions: []string{}},
					},
				},
				{
					Kind:         "ServiceAccount",
					Id:           "osd-image-registry",
					OsdRole:      "operator-image-registry",
					AccessMethod: "wif",
					Roles: []models.Role{
						{Kind: "Role", Id: "resourcemanager.tagUser", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "storage.admin", Predefined: true, Permissions: []string{}},
					},
					CredentialRequest: models.CredentialRequest{
						SecretRef: models.SecretRef{
							Name:      "installer-cloud-credentials",
							Namespace: "openshift-image-registry",
						},
						ServiceAccountNames: []string{
							"cluster-image-registry-operator",
							"registry",
						},
					},
				},
				{
					Kind:         "ServiceAccount",
					Id:           "osd-cluster-ingress",
					OsdRole:      "operator-cluster-ingress",
					AccessMethod: "wif",
					Roles: []models.Role{
						{Kind: "Role", Id: "dns.admin", Predefined: true, Permissions: []string{}},
					},
					CredentialRequest: models.CredentialRequest{
						SecretRef: models.SecretRef{
							Name:      "cloud-credentials",
							Namespace: "openshift-ingress-operator",
						},
						ServiceAccountNames: []string{
							"ingress-operator",
						},
					},
				},
				{
					Kind:         "ServiceAccount",
					Id:           "osd-machine-api",
					OsdRole:      "operator-machine-api",
					AccessMethod: "wif",
					Roles: []models.Role{
						{Kind: "Role", Id: "compute.admin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.serviceAccountUser", Predefined: true, Permissions: []string{}},
					},
					CredentialRequest: models.CredentialRequest{
						SecretRef: models.SecretRef{
							Name:      "gcp-cloud-credentials",
							Namespace: "openshift-machine-api",
						},
						ServiceAccountNames: []string{
							"machine-api-controllers",
						},
					},
				},
				{
					Kind:         "ServiceAccount",
					Id:           "osd-cloud-controller-manager",
					OsdRole:      "operator-cloud-controller-manager",
					AccessMethod: "wif",
					Roles: []models.Role{
						{Kind: "Role", Id: "compute.instanceAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "compute.loadBalancerAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.serviceAccountUser", Predefined: true, Permissions: []string{}},
					},
					CredentialRequest: models.CredentialRequest{
						SecretRef: models.SecretRef{
							Name:      "gcp-ccm-cloud-credentials",
							Namespace: "openshift-cloud-controller-manager",
						},
						ServiceAccountNames: []string{
							"cloud-controller-manager",
						},
					},
				},
				{
					Kind:         "ServiceAccount",
					Id:           "osd-cluster-storage",
					OsdRole:      "operator-cluster-storage",
					AccessMethod: "wif",
					Roles: []models.Role{
						{Kind: "Role", Id: "compute.instanceAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "compute.storageAdmin", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.serviceAccountUser", Predefined: true, Permissions: []string{}},
					},
					CredentialRequest: models.CredentialRequest{
						SecretRef: models.SecretRef{
							Name:      "gcp-pd-cloud-credentials",
							Namespace: "openshift-cluster-csi-drivers",
						},
						ServiceAccountNames: []string{
							"gcp-pd-csi-driver-operator",
							"gcp-pd-csi-driver-controller-sa",
						},
					},
				},
				{
					Kind:         "ServiceAccount",
					Id:           "osd-cloud-credential",
					OsdRole:      "operator-cloud-credential",
					AccessMethod: "wif",
					Roles: []models.Role{
						{Kind: "Role", Id: "iam.roleViewer", Predefined: true, Permissions: []string{}},
						{Kind: "Role", Id: "iam.securityReviewer", Predefined: true, Permissions: []string{}},
					},
					CredentialRequest: models.CredentialRequest{
						SecretRef: models.SecretRef{
							Name:      "cloud-credential-operator-gcp-ro-creds",
							Namespace: "openshift-cloud-credential-operator",
						},
						ServiceAccountNames: []string{
							"cloud-credential-operator",
						},
					},
				},
				{
					Kind:         "ServiceAccount",
					Id:           "osd-cncc",
					OsdRole:      "operator-cncc",
					AccessMethod: "wif",
					Roles: []models.Role{
						{Kind: "Role", Id: "compute.admin", Predefined: true, Permissions: []string{}},
					},
					CredentialRequest: models.CredentialRequest{
						SecretRef: models.SecretRef{
							Name:      "cloud-credentials",
							Namespace: "openshift-cloud-network-config-controller",
						},
						ServiceAccountNames: []string{
							"cloud-network-config-controller",
						},
					},
				},
			},
		},
	}
}
