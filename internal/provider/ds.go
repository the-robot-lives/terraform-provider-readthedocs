package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/the-robot-lives/terraform-provider-readthedocs/internal/rtdapi"
)

type listModel struct {
	Project     types.String `tfsdk:"project"`
	Slug        types.String `tfsdk:"slug"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Expand      types.String `tfsdk:"expand"`
	URL         types.String `tfsdk:"url"`
	Version     types.String `tfsdk:"version"`
	BuildID     types.Int64  `tfsdk:"build_id"`
	Commit      types.String `tfsdk:"commit"`
	Running     types.String `tfsdk:"running"`
	VcsProvider types.String `tfsdk:"vcs_provider"`
	Org         types.String `tfsdk:"organization"`
	FullName    types.String `tfsdk:"full_name"`
	Doctool     types.String `tfsdk:"doctool"`
	Count       types.Int64  `tfsdk:"count"`
	ResultsJSON types.String `tfsdk:"results_json"`
	JSON        types.String `tfsdk:"json"`
	Content     types.String `tfsdk:"content"`
	Fragment    types.String `tfsdk:"fragment"`
}

func listSchema(desc string, extra map[string]schema.Attribute) schema.Schema {
	attrs := map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true},
		"count":         schema.Int64Attribute{Computed: true},
		"results_json":  schema.StringAttribute{Computed: true, MarkdownDescription: "JSON array of result objects"},
		"json":          schema.StringAttribute{Computed: true},
		"name":          schema.StringAttribute{Optional: true, Computed: true},
		"slug":          schema.StringAttribute{Optional: true, Computed: true},
		"project":       schema.StringAttribute{Optional: true, Computed: true},
		"expand":        schema.StringAttribute{Optional: true, Computed: true},
		"url":           schema.StringAttribute{Optional: true, Computed: true},
		"version":       schema.StringAttribute{Optional: true, Computed: true},
		"build_id":      schema.Int64Attribute{Optional: true, Computed: true},
		"commit":        schema.StringAttribute{Optional: true, Computed: true},
		"running":       schema.StringAttribute{Optional: true, Computed: true},
		"vcs_provider":  schema.StringAttribute{Optional: true, Computed: true},
		"organization":  schema.StringAttribute{Optional: true, Computed: true},
		"full_name":     schema.StringAttribute{Optional: true, Computed: true},
		"doctool":       schema.StringAttribute{Optional: true, Computed: true},
		"content":       schema.StringAttribute{Computed: true},
		"fragment":      schema.StringAttribute{Computed: true},
	}
	for k, v := range extra {
		attrs[k] = v
	}
	return schema.Schema{MarkdownDescription: desc, Attributes: attrs}
}

type genericDS struct {
	name   string
	desc   string
	extra  map[string]schema.Attribute
	read   func(ctx context.Context, c *rtdapi.Client, m *listModel) error
	client *rtdapi.Client
}

func (d *genericDS) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.name
}
func (d *genericDS) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	d.client = clientOf(req.ProviderData)
}
func (d *genericDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = listSchema(d.desc, d.extra)
}
func (d *genericDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var m listModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := d.read(ctx, d.client, &m); err != nil {
		resp.Diagnostics.AddError("read "+d.name, err.Error())
		return
	}
	if m.ID.IsNull() || m.ID.IsUnknown() || m.ID.ValueString() == "" {
		m.ID = types.StringValue(d.name)
	}
	if m.Count.IsUnknown() {
		m.Count = types.Int64Value(0)
	}
	if m.ResultsJSON.IsUnknown() || m.ResultsJSON.IsNull() {
		m.ResultsJSON = types.StringValue("[]")
	}
	zeroStr := [] *types.String{
		&m.JSON, &m.Name, &m.Slug, &m.Project, &m.Expand, &m.URL, &m.Version,
		&m.Commit, &m.Running, &m.VcsProvider, &m.Org, &m.FullName, &m.Doctool,
		&m.Content, &m.Fragment,
	}
	for _, p := range zeroStr {
		if p.IsNull() || p.IsUnknown() {
			*p = types.StringValue("")
		}
	}
	if m.BuildID.IsNull() || m.BuildID.IsUnknown() {
		m.BuildID = types.Int64Value(0)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func setResults(m *listModel, items []json.RawMessage, total int) {
	m.Count = types.Int64Value(int64(total))
	b, _ := json.Marshal(items)
	m.ResultsJSON = types.StringValue(string(b))
}

func NewProjectsDataSource() datasource.DataSource {
	return &genericDS{
		name: "projects",
		desc: "GET /api/v3/projects/ — projects for the token owner.",
		extra: map[string]schema.Attribute{
			"name":                  schema.StringAttribute{Optional: true},
			"slug":                  schema.StringAttribute{Optional: true},
			"expand":                schema.StringAttribute{Optional: true},
		},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListProjects(map[string]string{
				"name": optString(m.Name), "slug": optString(m.Slug), "expand": optString(m.Expand),
			})
			if err != nil {
				return err
			}
			setResults(m, items, n)
			return nil
		},
	}
}

func NewProjectDataSource() datasource.DataSource {
	return &genericDS{
		name: "project",
		desc: "GET /api/v3/projects/{slug}/",
		extra: map[string]schema.Attribute{
			"slug":   schema.StringAttribute{Required: true},
			"expand": schema.StringAttribute{Optional: true},
			"json":   schema.StringAttribute{Computed: true},
			"name":   schema.StringAttribute{Computed: true},
		},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			raw, err := c.GetProject(m.Slug.ValueString(), optString(m.Expand))
			if err != nil {
				return err
			}
			m.JSON = jsonString(raw)
			m.Name = types.StringValue(rtdapi.ExtractString(raw, "name"))
			m.ID = types.StringValue(rtdapi.ExtractString(raw, "slug"))
			m.ResultsJSON = types.StringValue("[]")
			m.Count = types.Int64Value(1)
			return nil
		},
	}
}

func NewVersionsDataSource() datasource.DataSource {
	return &genericDS{
		name: "versions",
		desc: "GET /projects/{slug}/versions/",
		extra: map[string]schema.Attribute{
			"project": schema.StringAttribute{Required: true},
		},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListVersions(m.Project.ValueString(), nil)
			if err != nil {
				return err
			}
			setResults(m, items, n)
			m.ID = types.StringValue(m.Project.ValueString() + "/versions")
			return nil
		},
	}
}

func NewVersionDataSource() datasource.DataSource {
	return &genericDS{
		name: "version",
		desc: "GET /projects/{p}/versions/{v}/",
		extra: map[string]schema.Attribute{
			"project": schema.StringAttribute{Required: true},
			"slug":    schema.StringAttribute{Required: true},
			"json":    schema.StringAttribute{Computed: true},
		},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			raw, err := c.GetVersion(m.Project.ValueString(), m.Slug.ValueString())
			if err != nil {
				return err
			}
			m.JSON = jsonString(raw)
			m.ID = types.StringValue(m.Project.ValueString() + "/" + m.Slug.ValueString())
			m.Count = types.Int64Value(1)
			m.ResultsJSON = types.StringValue("[]")
			return nil
		},
	}
}

func NewBuildsDataSource() datasource.DataSource {
	return &genericDS{
		name: "builds",
		desc: "GET /projects/{slug}/builds/",
		extra: map[string]schema.Attribute{
			"project": schema.StringAttribute{Required: true},
			"commit":  schema.StringAttribute{Optional: true},
			"running": schema.StringAttribute{Optional: true, MarkdownDescription: "true|false"},
		},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListBuilds(m.Project.ValueString(), map[string]string{
				"commit": optString(m.Commit), "running": optString(m.Running),
			})
			if err != nil {
				return err
			}
			setResults(m, items, n)
			m.ID = types.StringValue(m.Project.ValueString() + "/builds")
			return nil
		},
	}
}

func NewBuildDataSource() datasource.DataSource {
	return &genericDS{
		name: "build",
		desc: "GET /projects/{p}/builds/{id}/",
		extra: map[string]schema.Attribute{
			"project":  schema.StringAttribute{Required: true},
			"build_id": schema.Int64Attribute{Required: true},
			"expand":   schema.StringAttribute{Optional: true},
			"json":     schema.StringAttribute{Computed: true},
		},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			raw, err := c.GetBuild(m.Project.ValueString(), int(m.BuildID.ValueInt64()), optString(m.Expand))
			if err != nil {
				return err
			}
			m.JSON = jsonString(raw)
			m.ID = types.StringValue(strconv.FormatInt(m.BuildID.ValueInt64(), 10))
			m.Count = types.Int64Value(1)
			m.ResultsJSON = types.StringValue("[]")
			return nil
		},
	}
}

func NewRedirectsDataSource() datasource.DataSource {
	return &genericDS{name: "redirects", desc: "GET /projects/{p}/redirects/", extra: map[string]schema.Attribute{"project": schema.StringAttribute{Required: true}},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListRedirects(m.Project.ValueString())
			if err != nil {
				return err
			}
			setResults(m, items, n)
			m.ID = types.StringValue(m.Project.ValueString() + "/redirects")
			return nil
		}}
}

func NewEnvVarsDataSource() datasource.DataSource {
	return &genericDS{name: "environment_variables", desc: "GET /projects/{p}/environmentvariables/", extra: map[string]schema.Attribute{"project": schema.StringAttribute{Required: true}},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListEnvVars(m.Project.ValueString())
			if err != nil {
				return err
			}
			setResults(m, items, n)
			m.ID = types.StringValue(m.Project.ValueString() + "/env")
			return nil
		}}
}

func NewSubprojectsDataSource() datasource.DataSource {
	return &genericDS{name: "subprojects", desc: "GET /projects/{p}/subprojects/", extra: map[string]schema.Attribute{"project": schema.StringAttribute{Required: true}},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListSubprojects(m.Project.ValueString())
			if err != nil {
				return err
			}
			setResults(m, items, n)
			m.ID = types.StringValue(m.Project.ValueString() + "/sub")
			return nil
		}}
}

func NewTranslationsDataSource() datasource.DataSource {
	return &genericDS{name: "translations", desc: "GET /projects/{p}/translations/", extra: map[string]schema.Attribute{"project": schema.StringAttribute{Required: true}},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListTranslations(m.Project.ValueString())
			if err != nil {
				return err
			}
			setResults(m, items, n)
			m.ID = types.StringValue(m.Project.ValueString() + "/i18n")
			return nil
		}}
}

func NewOrganizationsDataSource() datasource.DataSource {
	return &genericDS{name: "organizations", desc: "GET /api/v3/organizations/ (Business)", extra: map[string]schema.Attribute{},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListOrganizations()
			if err != nil {
				return err
			}
			setResults(m, items, n)
			return nil
		}}
}

func NewOrganizationDataSource() datasource.DataSource {
	return &genericDS{name: "organization", desc: "GET /api/v3/organizations/{slug}/ (Business)", extra: map[string]schema.Attribute{
		"slug": schema.StringAttribute{Required: true}, "json": schema.StringAttribute{Computed: true},
	}, read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
		raw, err := c.GetOrganization(m.Slug.ValueString())
		if err != nil {
			return err
		}
		m.JSON = jsonString(raw)
		m.ID = types.StringValue(m.Slug.ValueString())
		m.Count = types.Int64Value(1)
		m.ResultsJSON = types.StringValue("[]")
		return nil
	}}
}

func NewOrgProjectsDataSource() datasource.DataSource {
	return &genericDS{name: "organization_projects", desc: "GET /organizations/{slug}/projects/", extra: map[string]schema.Attribute{"slug": schema.StringAttribute{Required: true}},
		read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
			items, n, err := c.ListOrganizationProjects(m.Slug.ValueString())
			if err != nil {
				return err
			}
			setResults(m, items, n)
			m.ID = types.StringValue(m.Slug.ValueString() + "/projects")
			return nil
		}}
}

func NewOrgTeamsDataSource() datasource.DataSource {
	return &genericDS{name: "organization_teams", desc: "GET /organizations/{slug}/teams/", extra: map[string]schema.Attribute{
		"slug": schema.StringAttribute{Required: true}, "expand": schema.StringAttribute{Optional: true},
	}, read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
		items, n, err := c.ListOrganizationTeams(m.Slug.ValueString(), optString(m.Expand))
		if err != nil {
			return err
		}
		setResults(m, items, n)
		m.ID = types.StringValue(m.Slug.ValueString() + "/teams")
		return nil
	}}
}

func NewRemoteOrgsDataSource() datasource.DataSource {
	return &genericDS{name: "remote_organizations", desc: "GET /api/v3/remote/organizations/", extra: map[string]schema.Attribute{
		"name": schema.StringAttribute{Optional: true}, "vcs_provider": schema.StringAttribute{Optional: true},
	}, read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
		items, n, err := c.ListRemoteOrganizations(map[string]string{"name": optString(m.Name), "vcs_provider": optString(m.VcsProvider)})
		if err != nil {
			return err
		}
		setResults(m, items, n)
		return nil
	}}
}

func NewRemoteReposDataSource() datasource.DataSource {
	return &genericDS{name: "remote_repositories", desc: "GET /api/v3/remote/repositories/", extra: map[string]schema.Attribute{
		"name": schema.StringAttribute{Optional: true}, "full_name": schema.StringAttribute{Optional: true},
		"vcs_provider": schema.StringAttribute{Optional: true}, "organization": schema.StringAttribute{Optional: true},
		"expand": schema.StringAttribute{Optional: true},
	}, read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
		items, n, err := c.ListRemoteRepositories(map[string]string{
			"name": optString(m.Name), "full_name": optString(m.FullName),
			"vcs_provider": optString(m.VcsProvider), "organization": optString(m.Org), "expand": optString(m.Expand),
		})
		if err != nil {
			return err
		}
		setResults(m, items, n)
		return nil
	}}
}

func NewEmbedDataSource() datasource.DataSource {
	return &genericDS{name: "embed", desc: "GET /api/v3/embed/?url=…", extra: map[string]schema.Attribute{
		"url": schema.StringAttribute{Required: true}, "doctool": schema.StringAttribute{Optional: true},
		"json": schema.StringAttribute{Computed: true}, "content": schema.StringAttribute{Computed: true},
		"fragment": schema.StringAttribute{Computed: true},
	}, read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
		raw, err := c.Embed(map[string]string{"url": m.URL.ValueString(), "doctool": optString(m.Doctool)})
		if err != nil {
			return err
		}
		m.JSON = jsonString(raw)
		m.Content = types.StringValue(rtdapi.ExtractString(raw, "content"))
		m.Fragment = types.StringValue(rtdapi.ExtractString(raw, "fragment"))
		m.ID = types.StringValue(m.URL.ValueString())
		m.Count = types.Int64Value(1)
		m.ResultsJSON = types.StringValue("[]")
		return nil
	}}
}

func NewSuperprojectDataSource() datasource.DataSource {
	return &genericDS{name: "superproject", desc: "GET /projects/{slug}/superproject/", extra: map[string]schema.Attribute{
		"project": schema.StringAttribute{Required: true}, "json": schema.StringAttribute{Computed: true},
	}, read: func(_ context.Context, c *rtdapi.Client, m *listModel) error {
		raw, err := c.GetSuperproject(m.Project.ValueString())
		if err != nil {
			return err
		}
		m.JSON = jsonString(raw)
		m.ID = types.StringValue(m.Project.ValueString() + "/super")
		m.Count = types.Int64Value(1)
		m.ResultsJSON = types.StringValue("[]")
		return nil
	}}
}
