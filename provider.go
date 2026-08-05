package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &deeplinkProvider{version: version}
	}
}

var _ provider.Provider = (*deeplinkProvider)(nil)

type deeplinkProvider struct {
	version string
}

type providerModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

func (p *deeplinkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "deeplink"
	resp.Version = p.version
}

func (p *deeplinkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage short links on a [deeplink](https://github.com/yinebebt/deeplink) server.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Base URL of the deeplink server (e.g. `https://link.example.com`).",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API key for mutating endpoints. Set in the provider block or via a variable (e.g. `TF_VAR_api_key`).",
			},
		},
	}
}

func (p *deeplinkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := cfg.Endpoint.ValueString()
	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing deeplink endpoint",
			"Set provider.endpoint to the deeplink server base URL.",
		)
		return
	}

	var apiKey string
	if !cfg.APIKey.IsNull() && !cfg.APIKey.IsUnknown() {
		apiKey = cfg.APIKey.ValueString()
	}

	client, err := newClient(endpoint, apiKey)
	if err != nil {
		resp.Diagnostics.AddError("Invalid deeplink client config", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *deeplinkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewLinkResource}
}

func (p *deeplinkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
