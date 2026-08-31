package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/the-robot-lives/terraform-provider-readthedocs/internal/rtdapi"
)

func NewProjectResource() resource.Resource { return &projectRes{} }

type projectRes struct{ c *rtdapi.Client }

type projectModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Slug                       types.String `tfsdk:"slug"`
	RepositoryURL              types.String `tfsdk:"repository_url"`
	RepositoryType             types.String `tfsdk:"repository_type"`
	Homepage                   types.String `tfsdk:"homepage"`
	Language                   types.String `tfsdk:"language"`
	ProgrammingLanguage        types.String `tfsdk:"programming_language"`
	DefaultVersion             types.String `tfsdk:"default_version"`
	DefaultBranch              types.String `tfsdk:"default_branch"`
	PrivacyLevel               types.String `tfsdk:"privacy_level"`
	ExternalBuildsPrivacyLevel types.String `tfsdk:"external_builds_privacy_level"`
	ExternalBuildsEnabled      types.Bool   `tfsdk:"external_builds_enabled"`
	AnalyticsCode              types.String `tfsdk:"analytics_code"`
	AnalyticsDisabled          types.Bool   `tfsdk:"analytics_disabled"`
	VersioningScheme           types.String `tfsdk:"versioning_scheme"`
	ReadthedocsYAMLPath        types.String `tfsdk:"readthedocs_yaml_path"`
	Organization               types.String `tfsdk:"organization"`
	Teams                      types.List   `tfsdk:"teams"`
	Tags                       types.List   `tfsdk:"tags"`
	DocsURL                    types.String `tfsdk:"docs_url"`
	HomeURL                    types.String `tfsdk:"home_url"`
	JSON                       types.String `tfsdk:"json"`
}

func (r *projectRes) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}
func (r *projectRes) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.c = clientOf(req.ProviderData)
}
func (r *projectRes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Import/update a project (`POST/GET/PATCH /api/v3/projects/`). Official v3 has no DELETE; destroy drops state only.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"name": schema.StringAttribute{Required: true},
			"slug": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()}},
			"repository_url":  schema.StringAttribute{Required: true},
			"repository_type": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("git")},
			"homepage":        schema.StringAttribute{Optional: true, Computed: true},
			"language":        schema.StringAttribute{Optional: true, Computed: true},
			"programming_language": schema.StringAttribute{Optional: true, Computed: true},
			"default_version": schema.StringAttribute{Optional: true, Computed: true},
			"default_branch": schema.StringAttribute{Optional: true, Computed: true,
				MarkdownDescription: "VCS default branch. RTD often returns `master` until versions are synced; the provider keeps the configured value after create/update so Terraform stays consistent, then PATCHes."},
			"privacy_level":   schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Business only: public|private"},
			"external_builds_privacy_level": schema.StringAttribute{Optional: true, Computed: true},
			"external_builds_enabled":       schema.BoolAttribute{Optional: true, Computed: true},
			"analytics_code":                schema.StringAttribute{Optional: true, Computed: true},
			"analytics_disabled":            schema.BoolAttribute{Optional: true, Computed: true},
			"versioning_scheme":             schema.StringAttribute{Optional: true, Computed: true},
			"readthedocs_yaml_path":         schema.StringAttribute{Optional: true, Computed: true},
			"organization":                  schema.StringAttribute{Optional: true, MarkdownDescription: "Business: org slug (create only)"},
			"teams": schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true,
				Default: listdefault.StaticValue(types.ListValueMust(types.StringType, nil))},
			"tags": schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true,
				Default: listdefault.StaticValue(types.ListValueMust(types.StringType, nil))},
			"docs_url": schema.StringAttribute{Computed: true},
			"home_url": schema.StringAttribute{Computed: true},
			"json":     schema.StringAttribute{Computed: true, MarkdownDescription: "Full API JSON"},
		},
	}
}

func (r *projectRes) body(m projectModel, create bool) map[string]any {
	b := map[string]any{
		"name": m.Name.ValueString(),
		"repository": map[string]any{
			"url":  m.RepositoryURL.ValueString(),
			"type": m.RepositoryType.ValueString(),
		},
	}
	if create {
		if s := optString(m.Slug); s != "" {
			b["slug"] = s
		}
		put(b, "organization", m.Organization)
		if !m.Teams.IsNull() && !m.Teams.IsUnknown() {
			var teams []string
			_ = m.Teams.ElementsAs(context.Background(), &teams, false)
			if len(teams) > 0 {
				b["teams"] = teams
			}
		}
	}
	put(b, "homepage", m.Homepage)
	put(b, "language", m.Language)
	put(b, "programming_language", m.ProgrammingLanguage)
	put(b, "default_version", m.DefaultVersion)
	put(b, "default_branch", m.DefaultBranch)
	put(b, "privacy_level", m.PrivacyLevel)
	put(b, "external_builds_privacy_level", m.ExternalBuildsPrivacyLevel)
	put(b, "analytics_code", m.AnalyticsCode)
	put(b, "versioning_scheme", m.VersioningScheme)
	put(b, "readthedocs_yaml_path", m.ReadthedocsYAMLPath)
	putBool(b, "external_builds_enabled", m.ExternalBuildsEnabled)
	putBool(b, "analytics_disabled", m.AnalyticsDisabled)
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		var tags []string
		_ = m.Tags.ElementsAs(context.Background(), &tags, false)
		b["tags"] = tags
	}
	return b
}

func (r *projectRes) hydrate(raw json.RawMessage, m *projectModel) {
	m.JSON = jsonString(raw)
	if s := rtdapi.ExtractString(raw, "slug"); s != "" {
		m.ID = types.StringValue(s)
		m.Slug = types.StringValue(s)
	}
	if s := rtdapi.ExtractString(raw, "name"); s != "" {
		m.Name = types.StringValue(s)
	}
	if s := rtdapi.NestedString(raw, "repository", "url"); s != "" {
		m.RepositoryURL = types.StringValue(s)
	}
	if s := rtdapi.NestedString(raw, "repository", "type"); s != "" {
		m.RepositoryType = types.StringValue(s)
	}
	m.DocsURL = types.StringValue(rtdapi.NestedString(raw, "urls", "documentation"))
	m.HomeURL = types.StringValue(rtdapi.NestedString(raw, "urls", "home"))
	if s := rtdapi.NestedString(raw, "language", "code"); s != "" {
		m.Language = types.StringValue(s)
	}
	if s := rtdapi.NestedString(raw, "programming_language", "code"); s != "" {
		m.ProgrammingLanguage = types.StringValue(s)
	}
	if s := rtdapi.ExtractString(raw, "default_version"); s != "" {
		m.DefaultVersion = types.StringValue(s)
	}
	if s := rtdapi.ExtractString(raw, "default_branch"); s != "" {
		m.DefaultBranch = types.StringValue(s)
	}
	if s := rtdapi.ExtractString(raw, "homepage"); s != "" {
		m.Homepage = types.StringValue(s)
	}
	if s := rtdapi.ExtractString(raw, "privacy_level"); s != "" {
		m.PrivacyLevel = types.StringValue(s)
	}
	if s := rtdapi.ExtractString(raw, "versioning_scheme"); s != "" {
		m.VersioningScheme = types.StringValue(s)
	}
	if s := rtdapi.ExtractString(raw, "readthedocs_yaml_path"); s != "" {
		m.ReadthedocsYAMLPath = types.StringValue(s)
	}
	if m.Homepage.IsNull() || m.Homepage.IsUnknown() {
		m.Homepage = types.StringValue("")
	}
	if m.Language.IsNull() || m.Language.IsUnknown() {
		m.Language = types.StringValue("")
	}
	if m.ProgrammingLanguage.IsNull() || m.ProgrammingLanguage.IsUnknown() {
		m.ProgrammingLanguage = types.StringValue("")
	}
	if m.DefaultVersion.IsNull() || m.DefaultVersion.IsUnknown() {
		m.DefaultVersion = types.StringValue("")
	}
	if m.DefaultBranch.IsNull() || m.DefaultBranch.IsUnknown() {
		m.DefaultBranch = types.StringValue("")
	}
	if m.PrivacyLevel.IsNull() || m.PrivacyLevel.IsUnknown() {
		m.PrivacyLevel = types.StringValue("")
	}
	if m.ExternalBuildsPrivacyLevel.IsNull() || m.ExternalBuildsPrivacyLevel.IsUnknown() {
		m.ExternalBuildsPrivacyLevel = types.StringValue("")
	}
	if m.AnalyticsCode.IsNull() || m.AnalyticsCode.IsUnknown() {
		m.AnalyticsCode = types.StringValue("")
	}
	if m.VersioningScheme.IsNull() || m.VersioningScheme.IsUnknown() {
		m.VersioningScheme = types.StringValue("")
	}
	if m.ReadthedocsYAMLPath.IsNull() || m.ReadthedocsYAMLPath.IsUnknown() {
		m.ReadthedocsYAMLPath = types.StringValue("")
	}
	if m.ExternalBuildsEnabled.IsUnknown() {
		m.ExternalBuildsEnabled = types.BoolValue(false)
	}
	if m.AnalyticsDisabled.IsUnknown() {
		m.AnalyticsDisabled = types.BoolValue(false)
	}
	if m.Tags.IsNull() || m.Tags.IsUnknown() {
		m.Tags = types.ListValueMust(types.StringType, nil)
	}
	if m.Teams.IsNull() || m.Teams.IsUnknown() {
		m.Teams = types.ListValueMust(types.StringType, nil)
	}
}

func projectSlug(m projectModel) string {
	if s := optString(m.Slug); s != "" {
		return s
	}
	return optString(m.Name)
}

// keepConfigured copies known planned values over API hydrate.
// RTD create often ignores default_branch (returns master) until sync-versions.
func keepConfigured(dst, src *projectModel) {
	keepStr := func(d, s *types.String) {
		if !s.IsNull() && !s.IsUnknown() {
			*d = *s
		}
	}
	keepStr(&dst.Name, &src.Name)
	keepStr(&dst.RepositoryURL, &src.RepositoryURL)
	keepStr(&dst.RepositoryType, &src.RepositoryType)
	keepStr(&dst.Homepage, &src.Homepage)
	keepStr(&dst.Language, &src.Language)
	keepStr(&dst.ProgrammingLanguage, &src.ProgrammingLanguage)
	keepStr(&dst.DefaultVersion, &src.DefaultVersion)
	keepStr(&dst.DefaultBranch, &src.DefaultBranch)
	keepStr(&dst.PrivacyLevel, &src.PrivacyLevel)
	keepStr(&dst.ExternalBuildsPrivacyLevel, &src.ExternalBuildsPrivacyLevel)
	keepStr(&dst.AnalyticsCode, &src.AnalyticsCode)
	keepStr(&dst.VersioningScheme, &src.VersioningScheme)
	keepStr(&dst.ReadthedocsYAMLPath, &src.ReadthedocsYAMLPath)
	if !src.Tags.IsNull() && !src.Tags.IsUnknown() {
		dst.Tags = src.Tags
	}
	if !src.Teams.IsNull() && !src.Teams.IsUnknown() {
		dst.Teams = src.Teams
	}
	if !src.ExternalBuildsEnabled.IsNull() && !src.ExternalBuildsEnabled.IsUnknown() {
		dst.ExternalBuildsEnabled = src.ExternalBuildsEnabled
	}
	if !src.AnalyticsDisabled.IsNull() && !src.AnalyticsDisabled.IsUnknown() {
		dst.AnalyticsDisabled = src.AnalyticsDisabled
	}
}

func (r *projectRes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	planned := plan
	raw, err := r.c.CreateProject(r.body(plan, true))
	if err != nil {
		slug := projectSlug(plan)
		existing, gerr := r.c.GetProject(slug, "")
		if gerr != nil {
			resp.Diagnostics.AddError("create project", err.Error())
			return
		}
		raw = existing
		if uerr := r.c.UpdateProject(slug, r.body(plan, false)); uerr == nil {
			if refreshed, rerr := r.c.GetProject(slug, ""); rerr == nil {
				raw = refreshed
			}
		}
	}
	r.hydrate(raw, &plan)
	keepConfigured(&plan, &planned)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectRes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var st projectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, err := r.c.GetProject(st.Slug.ValueString(), "active_versions")
	if err != nil {
		resp.Diagnostics.AddError("read project", err.Error())
		return
	}
	r.hydrate(raw, &st)
	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}

func (r *projectRes) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	planned := plan
	if err := r.c.UpdateProject(plan.Slug.ValueString(), r.body(plan, false)); err != nil {
		resp.Diagnostics.AddError("update project", err.Error())
		return
	}
	raw, err := r.c.GetProject(plan.Slug.ValueString(), "")
	if err != nil {
		resp.Diagnostics.AddError("read after update", err.Error())
		return
	}
	r.hydrate(raw, &plan)
	keepConfigured(&plan, &planned)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectRes) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("not deleted on Read the Docs", "API v3 documents no project DELETE. State removed; the RTD project remains.")
}

func (r *projectRes) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}
