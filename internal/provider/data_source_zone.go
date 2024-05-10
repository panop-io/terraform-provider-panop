// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &PanopZoneDataSource{}

func NewPanopZoneDataSource() datasource.DataSource {
	return &PanopZoneDataSource{}
}

// PanopAssetDataSource defines the data source implementation.
type PanopZoneDataSource struct {
	clientHttp *http.Client
	host       string
	accessKey  string
}

// ZoneResourceModel describes the resource data model.
type ZoneModel struct {
	ZoneName types.String `tfsdk:"zone_name"`
	TenantId types.Int64  `tfsdk:"tenant_id"`
	Id       types.Int64  `tfsdk:"id"`
	ZoneType types.String `tfsdk:"zone_type"`
	Token    types.String `tfsdk:"token"`
}

// coffeesDataSourceModel maps the data source schema data.
type PanopZoneDataSourceModel struct {
	Zones []ZoneModel `tfsdk:"zones"`
}

func (d *PanopZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *PanopZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: " data source",

		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"zone_name": schema.StringAttribute{
							Computed: true,
						},
						"tenant_id": schema.Int64Attribute{
							Computed: true,
						},
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"zone_type": schema.StringAttribute{
							Computed: true,
						},
						"token": schema.StringAttribute{
							Computed:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
	}
}

func (d *PanopZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(clientObj)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.clientHttp = client.clientHttp
	d.host = client.host
	d.accessKey = client.accessKey

}

func (d *PanopZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PanopZoneDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Tower call
	urlSvc := url.URL{
		Scheme: "https",
		Host:   d.host,
		Path:   "/api/zones",
	}
	httpReq, err := http.NewRequest(http.MethodGet, urlSvc.String(), nil)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request creation error", err.Error())
		return
	}

	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", d.accessKey))
	httpReq.Header.Add("Content-Type", "application/json")

	httpResp, err := d.clientHttp.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request error", err.Error())
		return
	}

	type ZoneResponse struct {
		Id        uint   `json:"id"`
		ZoneName  string `json:"zone_name"`
		ZoneType  string `json:"zone_type"`
		Validated bool   `json:"validated"`
		Token     string `json:"token"`
		TenantId  uint   `json:"tenant_id"`
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create zonne, got error: %s", err))
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create zone, got error: %s",
			httpResp.Status))
		return
	}

	zones := []ZoneResponse{}
	_ = json.Unmarshal(respBody, &zones)
	// this is the end of tower call
	for _, zone := range zones {
		zoneModel := ZoneModel{
			ZoneName: types.StringValue(zone.ZoneName),
			TenantId: types.Int64Value(int64(zone.TenantId)),
			Id:       types.Int64Value(int64(zone.Id)),
			ZoneType: types.StringValue(zone.ZoneType),
			Token:    types.StringValue(zone.Token),
		}
		data.Zones = append(data.Zones, zoneModel)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
