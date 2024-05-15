// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure PanopProvider satisfies various provider interfaces.
var _ provider.Provider = &PanopProvider{}

// PanopProvider defines the provider implementation.
type PanopProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PanopProviderModel describes the provider data model.
type PanopProviderModel struct {
	Host          types.String `tfsdk:"host"`
	SkipTLSVerify types.Bool   `tfsdk:"skip_tls_verify"`
	AccessKey     types.String `tfsdk:"access_key"`
}

func (p *PanopProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "panop"
	resp.Version = p.version
}

func (p *PanopProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Tower Host",
				Required:            true,
			},
			"skip_tls_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS verify",
				Optional:            true,
			},
			"access_key": schema.StringAttribute{
				MarkdownDescription: "Tower access key",
				Optional:            true,
			},
		},
	}
}

type clientObj struct {
	clientHttp *http.Client
	host       string
	accessKey  string
}

func (p *PanopProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PanopProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Example client configuration for data sources and resources
	clientHttp := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: data.SkipTLSVerify.ValueBool()},
		},
	}

	access_key := os.Getenv("PANOP_ACCESS_KEY")
	if !data.AccessKey.IsNull() && access_key == "" {
		access_key = data.AccessKey.ValueString()
	}

	host := os.Getenv("PANOP_HOST")
	if !data.Host.IsNull() && host == "" {
		host = data.Host.ValueString()
	}

	client := clientObj{
		clientHttp: clientHttp,
		accessKey:  access_key,
		host:       host,
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PanopProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPanopZoneResource, NewPanopAssetResource,
	}
}

func (p *PanopProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPanopZoneDataSource, NewPanopAssetDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PanopProvider{
			version: version,
		}
	}
}
