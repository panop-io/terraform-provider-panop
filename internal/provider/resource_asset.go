// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PanopAssetResource{}
var _ resource.ResourceWithImportState = &PanopAssetResource{}

// PanopZoneResource defines the resource implementation.
type PanopAssetResource struct {
	clientHttp *http.Client
	host       string
	accessKey  string
}

func NewPanopAssetResource() resource.Resource {
	return &PanopAssetResource{}
}

// AssetResourceModel describes the resource data model.
type AssetResourceModel struct {
	AssetName types.String `tfsdk:"asset_name"`
	Id        types.Int64  `tfsdk:"id"`
	ZoneId    types.Int64  `tfsdk:"zone_id"`
}

func (r *PanopAssetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset"
}

func (r *PanopAssetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AssetResponse description",

		Attributes: map[string]schema.Attribute{
			"asset_name": schema.StringAttribute{
				MarkdownDescription: "Asset Name",
				Required:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "Asset Id",
				Computed:            true,
			},
			"zone_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Zone Id",
			},
		},
	}
}

func (r *PanopAssetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(clientObj)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected clientObj, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.clientHttp = client.clientHttp
	r.host = client.host
	r.accessKey = client.accessKey
}

func (r *PanopAssetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AssetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Tower call.
	urlSvc := url.URL{
		Scheme: "https",
		Host:   r.host,
		Path:   "/api/assets",
	}
	type AssetInput struct {
		AssetName string `json:"asset_name"`
		ZoneId    int64  `json:"zone_id"`
	}
	assetInput := AssetInput{
		AssetName: data.AssetName.ValueString(),
		ZoneId:    data.ZoneId.ValueInt64(),
	}
	body, _ := json.Marshal(assetInput)

	httpReq, err := http.NewRequest(http.MethodPost, urlSvc.String(), bytes.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("JSON Marshal Error", fmt.Sprintf("Unable to marshal assetInput to JSON, got error: %s", err))
		return
	}

	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessKey))
	httpReq.Header.Add("Content-Type", "application/json")

	httpResp, err := r.clientHttp.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	type AssetResponse struct {
		AssetId   uint   `json:"asset_id"`
		AssetName string `json:"asset_name"`
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create asset, got error: %s", err))
		return
	}

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create asset, got error: %s",
			httpResp.Status))
		return
	}

	zone := AssetResponse{}
	_ = json.Unmarshal(respBody, &zone)

	data.AssetName = types.StringValue(zone.AssetName)
	data.Id = types.Int64Value(int64(zone.AssetId))

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PanopAssetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AssetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Tower call
	urlSvc := url.URL{
		Scheme: "https",
		Host:   r.host,
		Path:   "/api/assets",
	}
	httpReq, err := http.NewRequest(http.MethodGet, urlSvc.String(), nil)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request creation error", err.Error())
		return
	}

	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessKey))
	httpReq.Header.Add("Content-Type", "application/json")

	httpResp, err := r.clientHttp.Do(httpReq)
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
		if asset.AssetId == data.Id.ValueInt64() {
			data.AssetName = types.StringValue(asset.AssetName)
			data.ZoneId = types.Int64Value(asset.ZoneId)
			break
		}
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PanopAssetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ZoneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PanopAssetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AssetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Tower call
	urlSvc := url.URL{
		Scheme: "https",
		Host:   r.host,
		Path:   fmt.Sprintf("/api/assets/%d", data.Id.ValueInt64()),
	}
	httpReq, err := http.NewRequest(http.MethodDelete, urlSvc.String(), nil)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request creation error", err.Error())
		return
	}
	// add authorization token
	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessKey))
	httpReq.Header.Add("Content-Type", "application/json")

	httpResp, err := r.clientHttp.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request error", err.Error())
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send deletion: %t", err))
		return
	}
	// this is the end of tower call

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *PanopAssetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, _ := strconv.ParseInt(req.ID, 10, 64)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
