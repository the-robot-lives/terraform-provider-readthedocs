package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/the-robot-lives/terraform-provider-readthedocs/internal/rtdapi"
)

func New() provider.Provider { return &rtdProvider{} }

type rtdProvider struct{}

type providerModel struct {
	Token   types.String `tfsdk:"token"`
	BaseURL types.String `tfsdk:"base_url"`
}

type ctxData struct{ Client *rtdapi.Client }

func (p *rtdProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "readthedocs"
	resp.Version = "0.2.0"
}

func (p *rtdProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "From-scratch Terraform/OpenTofu provider for **Read the Docs API v3** " +
			"(https://docs.readthedocs.com/platform/stable/api/v3.html). Not a fork of BarnabyShearer/readthedocs " +
			"or any MCP. Covers projects, versions, builds (push), subprojects, translations, redirects, " +
			"environment variables, Business sharing, organizations, remote VCS, and embed.\n\n" +
			"Auth: `token` or `READTHEDOCS_TOKEN`. Default base `https://app.readthedocs.org/api/v3`. " +
			"Business: `https://app.readthedocs.com/api/v3`.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API token (`Authorization: Token …`). Env: READTHEDOCS_TOKEN.",
			},
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "API v3 root, no trailing slash. Env: READTHEDOCS_BASE_URL.",
			},
		},
	}
}

func (p *rtdProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	token := os.Getenv("READTHEDOCS_TOKEN")
	if v := cfg.Token.ValueString(); v != "" {
		token = v
	}
	if token == "" {
		resp.Diagnostics.AddError("missing token", "Set provider.token or READTHEDOCS_TOKEN.")
		return
	}
	base := os.Getenv("READTHEDOCS_BASE_URL")
	if v := cfg.BaseURL.ValueString(); v != "" {
		base = v
	}
	if base == "" {
		base = rtdapi.DefaultBaseURL
	}
	d := &ctxData{Client: rtdapi.New(base, token)}
	resp.ResourceData = d
	resp.DataSourceData = d
}

func (p *rtdProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewVersionResource,
		NewBuildResource,
		NewSyncVersionsResource,
		NewRedirectResource,
		NewEnvVarResource,
		NewSubprojectResource,
		NewSharingResource,
	}
}

func (p *rtdProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewProjectsDataSource,
		NewVersionDataSource,
		NewVersionsDataSource,
		NewBuildDataSource,
		NewBuildsDataSource,
		NewRedirectsDataSource,
		NewEnvVarsDataSource,
		NewSubprojectsDataSource,
		NewTranslationsDataSource,
		NewOrganizationDataSource,
		NewOrganizationsDataSource,
		NewOrgProjectsDataSource,
		NewOrgTeamsDataSource,
		NewRemoteOrgsDataSource,
		NewRemoteReposDataSource,
		NewEmbedDataSource,
		NewSuperprojectDataSource,
	}
}

func clientOf(data any) *rtdapi.Client {
	if data == nil {
		return nil
	}
	if d, ok := data.(*ctxData); ok {
		return d.Client
	}
	return nil
}
