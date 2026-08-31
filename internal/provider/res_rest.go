package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/the-robot-lives/terraform-provider-readthedocs/internal/rtdapi"
)

func NewVersionResource() resource.Resource { return &versionRes{} }

type versionRes struct{ c *rtdapi.Client }

type versionModel struct {
	ID           types.String `tfsdk:"id"`
	Project      types.String `tfsdk:"project"`
	Slug         types.String `tfsdk:"slug"`
	Active       types.Bool   `tfsdk:"active"`
	Hidden       types.Bool   `tfsdk:"hidden"`
	PrivacyLevel types.String `tfsdk:"privacy_level"`
	JSON         types.String `tfsdk:"json"`
}

func (r *versionRes) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_version"
}
func (r *versionRes) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.c = clientOf(req.ProviderData)
}
func (r *versionRes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Update a version (`PATCH /projects/{p}/versions/{v}/`). Deactivate removes docs; activate triggers a build.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project":       schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"slug":          schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"active":        schema.BoolAttribute{Optional: true, Computed: true},
			"hidden":        schema.BoolAttribute{Optional: true, Computed: true},
			"privacy_level": schema.StringAttribute{Optional: true, Computed: true},
			"json":          schema.StringAttribute{Computed: true},
		},
	}
}

func (r *versionRes) write(ctx context.Context, plan *versionModel, addErr func(string, string), setState func(context.Context, any)) {
	body := map[string]any{}
	putBool(body, "active", plan.Active)
	putBool(body, "hidden", plan.Hidden)
	put(body, "privacy_level", plan.PrivacyLevel)
	if err := r.c.UpdateVersion(plan.Project.ValueString(), plan.Slug.ValueString(), body); err != nil {
		addErr("update version", err.Error())
		return
	}
	raw, err := r.c.GetVersion(plan.Project.ValueString(), plan.Slug.ValueString())
	if err != nil {
		addErr("read version", err.Error())
		return
	}
	r.hydrate(raw, plan)
	setState(ctx, plan)
}

func (r *versionRes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan versionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, resp.Diagnostics.AddError, func(ctx context.Context, v any) {
		resp.Diagnostics.Append(resp.State.Set(ctx, v)...)
	})
}
func (r *versionRes) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan versionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.write(ctx, &plan, resp.Diagnostics.AddError, func(ctx context.Context, v any) {
		resp.Diagnostics.Append(resp.State.Set(ctx, v)...)
	})
}
func (r *versionRes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var st versionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, err := r.c.GetVersion(st.Project.ValueString(), st.Slug.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read version", err.Error())
		return
	}
	r.hydrate(raw, &st)
	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}
func (r *versionRes) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("version not deleted", "No version DELETE in API v3.")
}
func (r *versionRes) hydrate(raw json.RawMessage, m *versionModel) {
	m.JSON = jsonString(raw)
	if s := rtdapi.ExtractString(raw, "slug"); s != "" {
		m.Slug = types.StringValue(s)
	}
	m.ID = types.StringValue(m.Project.ValueString() + "/" + m.Slug.ValueString())
	var obj map[string]any
	_ = json.Unmarshal(raw, &obj)
	if v, ok := obj["active"].(bool); ok {
		m.Active = types.BoolValue(v)
	} else if m.Active.IsUnknown() {
		m.Active = types.BoolValue(false)
	}
	if v, ok := obj["hidden"].(bool); ok {
		m.Hidden = types.BoolValue(v)
	} else if m.Hidden.IsUnknown() {
		m.Hidden = types.BoolValue(false)
	}
	if s := rtdapi.ExtractString(raw, "privacy_level"); s != "" {
		m.PrivacyLevel = types.StringValue(s)
	} else if m.PrivacyLevel.IsNull() || m.PrivacyLevel.IsUnknown() {
		m.PrivacyLevel = types.StringValue("")
	}
}

// ---- build (POST .../versions/{v}/builds/) --------------------------------

func NewBuildResource() resource.Resource { return &buildRes{} }

type buildRes struct{ c *rtdapi.Client }

type buildModel struct {
	ID       types.String `tfsdk:"id"`
	Project  types.String `tfsdk:"project"`
	Version  types.String `tfsdk:"version"`
	Commit   types.String `tfsdk:"commit"`
	State    types.String `tfsdk:"state"`
	Success  types.Bool   `tfsdk:"success"`
	Triggers types.Map    `tfsdk:"triggers"`
	JSON     types.String `tfsdk:"json"`
}

func (r *buildRes) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build"
}
func (r *buildRes) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.c = clientOf(req.ProviderData)
}
func (r *buildRes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Trigger a documentation build — the git-push equivalent (`POST /projects/{p}/versions/{v}/builds/` → 202). Destroy is a no-op. Change `triggers` to fire another build.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"version": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("latest"), PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"commit":  schema.StringAttribute{Computed: true},
			"state":   schema.StringAttribute{Computed: true},
			"success": schema.BoolAttribute{Computed: true},
			"triggers": schema.MapAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				MarkdownDescription: "Arbitrary map. Changing a value replaces the resource and triggers a new build.",
				PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"json": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *buildRes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan buildModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, err := r.c.TriggerBuild(plan.Project.ValueString(), plan.Version.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("trigger build", err.Error())
		return
	}
	plan.JSON = jsonString(raw)
	var env struct {
		Build json.RawMessage `json:"build"`
	}
	_ = json.Unmarshal(raw, &env)
	b := env.Build
	if len(b) == 0 {
		b = raw
	}
	id := rtdapi.ExtractInt(b, "id", "pk")
	if id != 0 {
		plan.ID = types.StringValue(strconv.Itoa(id))
	} else {
		plan.ID = types.StringValue(plan.Project.ValueString() + "/" + plan.Version.ValueString())
	}
	plan.Commit = types.StringValue(rtdapi.ExtractString(b, "commit"))
	if s := rtdapi.NestedString(b, "state", "code"); s != "" {
		plan.State = types.StringValue(s)
	} else {
		plan.State = types.StringValue(rtdapi.ExtractString(b, "state"))
	}
	var obj map[string]any
	_ = json.Unmarshal(b, &obj)
	if v, ok := obj["success"].(bool); ok {
		plan.Success = types.BoolValue(v)
	} else {
		plan.Success = types.BoolValue(false)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *buildRes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var st buildModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.Atoi(st.ID.ValueString())
	if err != nil {
		return
	}
	raw, err := r.c.GetBuild(st.Project.ValueString(), id, "config")
	if err != nil {
		return
	}
	st.JSON = jsonString(raw)
	st.Commit = types.StringValue(rtdapi.ExtractString(raw, "commit"))
	if s := rtdapi.NestedString(raw, "state", "code"); s != "" {
		st.State = types.StringValue(s)
	}
	var obj map[string]any
	_ = json.Unmarshal(raw, &obj)
	if v, ok := obj["success"].(bool); ok {
		st.Success = types.BoolValue(v)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}
func (r *buildRes) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("immutable", "use triggers / taint to fire a new build")
}
func (r *buildRes) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}

// ---- sync-versions ---------------------------------------------------------

func NewSyncVersionsResource() resource.Resource { return &syncRes{} }

type syncRes struct{ c *rtdapi.Client }

type syncModel struct {
	ID       types.String `tfsdk:"id"`
	Project  types.String `tfsdk:"project"`
	Triggers types.Map    `tfsdk:"triggers"`
}

func (r *syncRes) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync_versions"
}
func (r *syncRes) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.c = clientOf(req.ProviderData)
}
func (r *syncRes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "POST /projects/{slug}/sync-versions/ — resync tags/branches from VCS (202). Replace via `triggers`.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"triggers": schema.MapAttribute{
				ElementType:   types.StringType,
				Optional:      true,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
		},
	}
}
func (r *syncRes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan syncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.SyncVersions(plan.Project.ValueString()); err != nil {
		resp.Diagnostics.AddError("sync-versions", err.Error())
		return
	}
	plan.ID = types.StringValue(plan.Project.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *syncRes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var st syncModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}
func (r *syncRes) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("immutable", "change triggers to re-sync")
}
func (r *syncRes) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}


