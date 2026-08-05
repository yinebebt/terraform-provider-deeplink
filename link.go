package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yinebebt/deeplink"
)

func NewLinkResource() resource.Resource {
	return &linkResource{}
}

var (
	_ resource.Resource                = (*linkResource)(nil)
	_ resource.ResourceWithImportState = (*linkResource)(nil)
)

type linkResource struct {
	client *Client
}

type linkModel struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	ShortID     types.String `tfsdk:"short_id"`
	ShortLink   types.String `tfsdk:"short_link"`
	URL         types.String `tfsdk:"url"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	ImageURL    types.String `tfsdk:"image_url"`
	ImageWidth  types.Int64  `tfsdk:"image_width"`
	ImageHeight types.Int64  `tfsdk:"image_height"`
	ImageAlt    types.String `tfsdk:"image_alt"`
	OGType      types.String `tfsdk:"og_type"`
	Locale      types.String `tfsdk:"locale"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	Clicks      types.Int64  `tfsdk:"clicks"`
}

func (r *linkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_link"
}

func (r *linkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a short link on a deeplink server (`POST /shorten`, `PATCH /{id}`, `DELETE /{id}`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Canonical ID (`type/short_id`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Processor type (e.g. `redirect`). Changing forces replacement.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"short_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated short ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"short_link": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Full short URL including the server base URL.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Destination URL (http/https).",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"title": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "OG title. Server may default this for redirects.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "OG description. Server may default this for redirects.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "OG image URL.",
			},
			"image_width": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "OG image width hint.",
			},
			"image_height": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "OG image height hint.",
			},
			"image_alt": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "OG image alt text.",
			},
			"og_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "og:type value (e.g. `website`, `article`).",
			},
			"locale": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "og:locale override.",
			},
			"expires_at": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "RFC 3339 expiry. Changing forces replacement (not patchable by the API).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"clicks": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *linkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *linkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan linkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateLink(ctx, toPayload(plan))
	if err != nil {
		resp.Diagnostics.AddError("Create link failed", err.Error())
		return
	}

	state := fromResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *linkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state linkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.client.GetLink(ctx, state.Type.ValueString(), state.ShortID.ValueString())
	if err != nil {
		var apiErr *apiError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read link failed", err.Error())
		return
	}

	next := fromResponse(got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &next)...)
}

func (r *linkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan linkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch := map[string]any{
		"url":   plan.URL.ValueString(),
		"title": plan.Title.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		patch["description"] = plan.Description.ValueString()
	}
	if !plan.ImageURL.IsNull() {
		patch["image_url"] = plan.ImageURL.ValueString()
	}
	if !plan.ImageWidth.IsNull() {
		patch["image_width"] = plan.ImageWidth.ValueInt64()
	}
	if !plan.ImageHeight.IsNull() {
		patch["image_height"] = plan.ImageHeight.ValueInt64()
	}
	if !plan.ImageAlt.IsNull() {
		patch["image_alt"] = plan.ImageAlt.ValueString()
	}
	if !plan.OGType.IsNull() {
		patch["og_type"] = plan.OGType.ValueString()
	}
	if !plan.Locale.IsNull() {
		patch["locale"] = plan.Locale.ValueString()
	}

	updated, err := r.client.UpdateLink(ctx, plan.ShortID.ValueString(), patch)
	if err != nil {
		resp.Diagnostics.AddError("Update link failed", err.Error())
		return
	}

	state := fromResponse(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *linkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state linkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLink(ctx, state.ShortID.ValueString()); err != nil {
		var apiErr *apiError
		if errors.As(err, &apiErr) && apiErr.NotFound() {
			return
		}
		resp.Diagnostics.AddError("Delete link failed", err.Error())
	}
}

func (r *linkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := strings.TrimSpace(req.ID)
	linkType, shortID, ok := strings.Cut(id, "/")
	if !ok || linkType == "" || shortID == "" || strings.Contains(shortID, "/") {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`Expected canonical ID "type/short_id" (e.g. "redirect/AbCd123").`,
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), linkType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("short_id"), shortID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), linkType+"/"+shortID)...)
}

func toPayload(m linkModel) deeplink.Link {
	p := deeplink.Link{
		Type: m.Type.ValueString(),
		URL:  m.URL.ValueString(),
	}
	if !m.Title.IsNull() && !m.Title.IsUnknown() {
		p.Title = m.Title.ValueString()
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		p.Description = m.Description.ValueString()
	}
	if !m.ImageURL.IsNull() {
		p.ImageURL = m.ImageURL.ValueString()
	}
	if !m.ImageWidth.IsNull() {
		p.ImageWidth = int(m.ImageWidth.ValueInt64())
	}
	if !m.ImageHeight.IsNull() {
		p.ImageHeight = int(m.ImageHeight.ValueInt64())
	}
	if !m.ImageAlt.IsNull() {
		p.ImageAlt = m.ImageAlt.ValueString()
	}
	if !m.OGType.IsNull() {
		p.OGType = m.OGType.ValueString()
	}
	if !m.Locale.IsNull() {
		p.Locale = m.Locale.ValueString()
	}
	if !m.ExpiresAt.IsNull() {
		p.ExpiresAt = m.ExpiresAt.ValueString()
	}
	return p
}

func fromResponse(r *deeplink.LinkResponse) linkModel {
	if r == nil || r.Link == nil {
		return linkModel{}
	}
	m := linkModel{
		ID:        types.StringValue(r.Type + "/" + r.ShortID),
		Type:      types.StringValue(r.Type),
		ShortID:   types.StringValue(r.ShortID),
		ShortLink: types.StringValue(r.ShortLink),
		URL:       types.StringValue(r.URL),
		Title:     types.StringValue(r.Title),
		Clicks:    types.Int64Value(r.Clicks),
	}
	m.Description = optionalString(r.Description)
	m.ImageURL = optionalString(r.ImageURL)
	m.ImageAlt = optionalString(r.ImageAlt)
	m.OGType = optionalString(r.OGType)
	m.Locale = optionalString(r.Locale)
	m.ExpiresAt = optionalString(r.ExpiresAt)
	m.CreatedAt = optionalString(r.CreatedAt)
	m.UpdatedAt = optionalString(r.UpdatedAt)
	if r.ImageWidth > 0 {
		m.ImageWidth = types.Int64Value(int64(r.ImageWidth))
	}
	if r.ImageHeight > 0 {
		m.ImageHeight = types.Int64Value(int64(r.ImageHeight))
	}
	return m
}

func optionalString(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}
