// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PanopZoneResource{}
var _ resource.ResourceWithImportState = &PanopZoneResource{}

func NewPanopZoneResource() resource.Resource {
	return &PanopZoneResource{}
}

// PanopZoneResource defines the resource implementation.
type PanopZoneResource struct {
	clientHttp *http.Client
	host       string
	accessKey  string
}

// ZoneResourceModel describes the resource data model.
type ZoneResourceModel struct {
	ZoneName types.String `tfsdk:"zone_name"`
	Id       types.Int64  `tfsdk:"id"`
	ZoneType types.String `tfsdk:"zone_type"`
	Token    types.String `tfsdk:"token"`
}

func (r *PanopZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (r *PanopZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ZoneResponse description",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Zone ZoneId",
				Computed:            true,
			},
			"zone_name": schema.StringAttribute{
				MarkdownDescription: "Zone Name",
				Required:            true,
			},
			"zone_type": schema.StringAttribute{
				MarkdownDescription: "ZoneResponse Type",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("dns"),
			},
			"token": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
		},
	}
}

func (r *PanopZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PanopZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ZoneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Tower call.
	urlSvc := url.URL{
		Scheme: "https",
		Host:   r.host,
		Path:   "/api/zones",
	}
	type ZoneInput struct {
		ZoneName string `gorm:"uniqueIndex" json:"zone_name"`
		TenantId int64  `gorm:"index" json:"tenant_id"`
	}
	zoneInput := ZoneInput{
		ZoneName: data.ZoneName.ValueString(),
	}
	body, _ := json.Marshal(zoneInput)

	httpReq, err := http.NewRequest(http.MethodPost, urlSvc.String(), bytes.NewReader(body))
	if err != nil {
		resp.Diagnostics.AddError("JSON Marshal Error", fmt.Sprintf("Unable to marshal zoneInput to JSON, got error: %s", err))
		return
	}
	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessKey))
	httpReq.Header.Add("Content-Type", "application/json")

	httpResp, err := r.clientHttp.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	type ZoneResponse struct {
		ZoneId    uint   `json:"zone_id"`
		ZoneName  string `json:"zone_name"`
		ZoneType  string `json:"zone_type"`
		Validated bool   `json:"validated"`
		Token     string `json:"token"`
	}

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create zone, got error: %s", err))
		return
	}

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Client Error", "Unable to create zone, got error, check your configuration")
		return
	}

	zone := ZoneResponse{}
	_ = json.Unmarshal(respBody, &zone)

	data.Token = types.StringValue(zone.Token)
	data.Id = types.Int64Value(int64(zone.ZoneId))

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PanopZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ZoneResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Tower call
	urlSvc := url.URL{
		Scheme: "https",
		Host:   r.host,
		Path:   "/api/zones",
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

	type ZoneResponse struct {
		Id        int64  `json:"id"`
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
	for _, zone := range zones {
		if zone.Id == data.Id.ValueInt64() {
			data.ZoneName = types.StringValue(zone.ZoneName)
			data.Token = types.StringValue(zone.Token)
			data.ZoneType = types.StringValue(zone.ZoneType)
			break
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PanopZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

func (r *PanopZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ZoneResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Tower call
	urlSvc := url.URL{
		Scheme: "https",
		Host:   r.host,
		Path:   fmt.Sprintf("/api/zones/%d", data.Id.ValueInt64()),
	}
	httpReq, err := http.NewRequest(http.MethodDelete, urlSvc.String(), nil)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request creation error", err.Error())
		return
	}
	// add authorization token
	httpReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", r.accessKey))

	httpResp, err := r.clientHttp.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("HTTP request error", err.Error())
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send deletion: %s", err))
		return
	}
	// this is the end of tower call

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *PanopZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, _ := strconv.ParseInt(req.ID, 10, 64)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
