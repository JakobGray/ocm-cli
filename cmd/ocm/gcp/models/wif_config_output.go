/*
 * Workload Identity Federation (W.I.F.) Configuration
 *
 * Defined here is the API for management of WIF Configuration for Openshift Dedicated on Google Cloud Platform (OSD-GCP).
 *
 * API version: 0.0.0
 * Contact: rcampos@redhat.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package models

type WifConfigOutput struct {
	Metadata *WifConfigMetadata `json:"metadata,omitempty"`
	Spec     *WifConfigInput    `json:"spec,omitempty"`
	Status   *WifConfigStatus   `json:"status,omitempty"`
}
