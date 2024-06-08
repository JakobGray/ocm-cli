package alphaocm

import (
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/pkg/errors"

	"github.com/openshift-online/ocm-cli/pkg/models"
	"github.com/openshift-online/ocm-cli/pkg/ocm"
)

type OcmClient interface {
	CreateWifConfig(models.WifConfigInput) (models.WifConfigOutput, error)
}

type ocmClient struct {
	connection *sdk.Connection
}

func NewOcmClient() (OcmClient, error) {
	connection, err := ocm.NewConnection().Build()
	if err != nil {
		return nil, err
	}
	return &ocmClient{
		connection: connection,
	}, nil
}

func (c *ocmClient) CreateWifConfig(input models.WifConfigInput) (models.WifConfigOutput, error) {
	var wifConfigOutput models.WifConfigOutput
	rawWifInput, err := input.ToJson()
	if err != nil {
		return wifConfigOutput, errors.Wrap(err, "failed to marshal wif input:")
	}
	resp, err := c.connection.Post().Path("/api/clusters_mgmt/v1/gcp/wif_configs").Bytes(rawWifInput).Send()
	if err != nil {
		return wifConfigOutput, err
	}
	rawWifOutput := resp.Bytes()
	wifConfigOutput, err = models.WifConfigOutputFromJson(rawWifOutput)
	return wifConfigOutput, err
}
