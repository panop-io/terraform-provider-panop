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
var _ datasource.DataSource = &PanopAssetDataSource{}

func NewPanopAssetDataSource() datasource.DataSource {
	return &PanopAssetDataSource{}
}

// PanopAssetDataSource defines the data source implementation.
type PanopAssetDataSource struct {
	clientHttp *http.Client
	host       string
	accessKey  string
}

// AssetResourceModel describes the resource data model.
type AssetDataSourceModel struct {
	AssetId   types.Int64  `tfsdk:"id"`
	AssetName types.String `tfsdk:"asset_name"`
	ZoneId    types.Int64  `tfsdk:"zone_id"`
}

// coffeesDataSourceModel maps the data source schema data.
type PanopAssetDataSourceModel struct {
	ZoneId types.Int64            `tfsdk:"zone_id"`
	Assets []AssetDataSourceModel `tfsdk:"assets"`
}

func (d *PanopAssetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset"
}

func (d *PanopAssetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: " data source",

		Attributes: map[string]schema.Attribute{
			"zone_id": schema.Int64Attribute{
				Description: "Zone Id Filter",
				Optional:    true,
			},

			"assets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"asset_name": schema.StringAttribute{
							Description: "Asset Name",
							Computed:    true,
						},
						"id": schema.Int64Attribute{
							Description: "Asset Id",
							Computed:    true,
						},
						"zone_id": schema.Int64Attribute{
							Description: "Zone Id",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *PanopAssetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PanopAssetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PanopAssetDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Tower call
	urlSvc := url.URL{
		Scheme: "https",
		Host:   d.host,
		Path:   "/api/assets",
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

	type AssetResponse struct {
		AssetId   int64  `json:"id"`
		AssetName string `json:"asset_name"`
		ZoneId    int64  `json:"zone_id"`
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

	assets := []AssetResponse{}
	_ = json.Unmarshal(respBody, &assets)
	// this is the end of tower call
	for _, asset := range assets {
		assetModel := AssetDataSourceModel{
			AssetName: types.StringValue(asset.AssetName),
			AssetId:   types.Int64Value(asset.AssetId),
			ZoneId:    types.Int64Value(asset.ZoneId),
		}
		if assetModel.ZoneId == data.ZoneId || data.ZoneId.IsNull() {
			data.Assets = append(data.Assets, assetModel)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
