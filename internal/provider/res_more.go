package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/the-robot-lives/terraform-provider-readthedocs/internal/rtdapi"
)

func NewRedirectResource() resource.Resource { return &redirectRes{} }

type redirectRes struct{ c *rtdapi.Client }

type redirectModel struct {
	ID          types.String `tfsdk:"id"`
	Project     types.String `tfsdk:"project"`
	Type        types.String `tfsdk:"type"`
	FromURL     types.String `tfsdk:"from_url"`
	ToURL       types.String `tfsdk:"to_url"`
	HTTPStatus  types.Int64  `tfsdk:"http_status"`
	Position    types.Int64  `tfsdk:"position"`
	Force       types.Bool   `tfsdk:"force"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	JSON        types.String `tfsdk:"json"`
}

func (r *redirectRes) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redirect"
}
func (r *redirectRes) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.c = clientOf(req.ProviderData)
}
func (r *redirectRes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "User-defined redirect (`/projects/{p}/redirects/`). type: page|exact|clean_url_to_html|html_to_clean_url.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project":     schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"type":        schema.StringAttribute{Required: true},
			"from_url":    schema.StringAttribute{Optional: true, Computed: true},
			"to_url":      schema.StringAttribute{Optional: true, Computed: true},
			"http_status": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(302)},
			"position":    schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(0)},
			"force":       schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"enabled":     schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"description": schema.StringAttribute{Optional: true, Computed: true},
			"json":        schema.StringAttribute{Computed: true},
		},
	}
}

func (r *redirectRes) body(m redirectModel) map[string]any {
	b := map[string]any{"type": m.Type.ValueString()}
	put(b, "from_url", m.FromURL)
	put(b, "to_url", m.ToURL)
	put(b, "description", m.Description)
	putInt(b, "http_status", m.HTTPStatus)
	putInt(b, "position", m.Position)
	putBool(b, "force", m.Force)
	putBool(b, "enabled", m.Enabled)
	return b
}

func (r *redirectRes) hydrate(raw json.RawMessage, m *redirectModel) {
	m.JSON = jsonString(raw)
	id := rtdapi.ExtractInt(raw, "pk", "id")
	if id != 0 {
		m.ID = types.StringValue(strconv.Itoa(id))
	}
	if s := rtdapi.ExtractString(raw, "type"); s != "" {
		m.Type = types.StringValue(s)
	}
	m.FromURL = types.StringValue(rtdapi.ExtractString(raw, "from_url"))
	m.ToURL = types.StringValue(rtdapi.ExtractString(raw, "to_url"))
	m.Description = types.StringValue(rtdapi.ExtractString(raw, "description"))
	var obj map[string]any
	_ = json.Unmarshal(raw, &obj)
	if v, ok := obj["http_status"].(float64); ok {
		m.HTTPStatus = types.Int64Value(int64(v))
	}
	if v, ok := obj["position"].(float64); ok {
		m.Position = types.Int64Value(int64(v))
	}
	if v, ok := obj["force"].(bool); ok {
		m.Force = types.BoolValue(v)
	}
	if v, ok := obj["enabled"].(bool); ok {
		m.Enabled = types.BoolValue(v)
	}
}

func (r *redirectRes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan redirectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	raw, err := r.c.CreateRedirect(plan.Project.ValueString(), r.body(plan))
	if err != nil {
		resp.Diagnostics.AddError("create redirect", err.Error())
		return
	}
	r.hydrate(raw, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *redirectRes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var st redirectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	id, _ := strconv.Atoi(st.ID.ValueString())
	raw, err := r.c.GetRedirect(st.Project.ValueString(), id)
	if err != nil {
		resp.Diagnostics.AddError("read redirect", err.Error())
		return
	}
	r.hydrate(raw, &st)
	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}
func (r *redirectRes) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan redirectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	id, _ := strconv.Atoi(plan.ID.ValueString())
	raw, err := r.c.UpdateRedirect(plan.Project.ValueString(), id, r.body(plan))
	if err != nil {
		resp.Diagnostics.AddError("update redirect", err.Error())
		return
	}
	r.hydrate(raw, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *redirectRes) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var st redirectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	id, _ := strconv.Atoi(st.ID.ValueString())
	if err := r.c.DeleteRedirect(st.Project.ValueString(), id); err != nil {
		resp.Diagnostics.AddError("delete redirect", err.Error())
	}
}

func NewEnvVarResource() resource.Resource { return &envRes{} }

type envRes struct{ c *rtdapi.Client }

type envModel struct {
	ID      types.String `tfsdk:"id"`
	Project types.String `tfsdk:"project"`
	Name    types.String `tfsdk:"name"`
	Value   types.String `tfsdk:"value"`
	Public  types.Bool   `tfsdk:"public"`
	JSON    types.String `tfsdk:"json"`
}

func (r *envRes) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}
func (r *envRes) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.c = clientOf(req.ProviderData)
}
func (r *envRes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Build environment variable. Value is write-only (no PATCH — rotate = replace). Public vars are exposed in PR builds.",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"name":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"value":   schema.StringAttribute{Required: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"public":  schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"json":    schema.StringAttribute{Computed: true},
		},
	}
}

func (r *envRes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan envModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	body := map[string]any{"name": plan.Name.ValueString(), "value": plan.Value.ValueString()}
	putBool(body, "public", plan.Public)
	raw, err := r.c.CreateEnvVar(plan.Project.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("create env var", err.Error())
		return
	}
	plan.JSON = jsonString(raw)
	id := rtdapi.ExtractInt(raw, "pk", "id")
	plan.ID = types.StringValue(strconv.Itoa(id))
	if v, ok := mapBool(raw, "public"); ok {
		plan.Public = types.BoolValue(v)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *envRes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var st envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	id, _ := strconv.Atoi(st.ID.ValueString())
	raw, err := r.c.GetEnvVar(st.Project.ValueString(), id)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	st.JSON = jsonString(raw)
	if s := rtdapi.ExtractString(raw, "name"); s != "" {
		st.Name = types.StringValue(s)
	}
	if v, ok := mapBool(raw, "public"); ok {
		st.Public = types.BoolValue(v)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}
func (r *envRes) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("immutable", "environment variables have no PATCH; taint/replace to rotate")
}
func (r *envRes) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var st envModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	id, _ := strconv.Atoi(st.ID.ValueString())
	if err := r.c.DeleteEnvVar(st.Project.ValueString(), id); err != nil {
		resp.Diagnostics.AddError("delete env var", err.Error())
	}
}

func NewSubprojectResource() resource.Resource { return &subRes{} }

type subRes struct{ c *rtdapi.Client }

type subModel struct {
	ID     types.String `tfsdk:"id"`
	Parent types.String `tfsdk:"parent"`
	Child  types.String `tfsdk:"child"`
	Alias  types.String `tfsdk:"alias"`
	JSON   types.String `tfsdk:"json"`
}

func (r *subRes) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subproject"
}
func (r *subRes) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.c = clientOf(req.ProviderData)
}
func (r *subRes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Nest a child project under a parent (`POST/DELETE /projects/{parent}/subprojects/`).",
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"parent": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"child":  schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"alias":  schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"json":   schema.StringAttribute{Computed: true},
		},
	}
}
func (r *subRes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	body := map[string]any{"child": plan.Child.ValueString()}
	put(body, "alias", plan.Alias)
	raw, err := r.c.CreateSubproject(plan.Parent.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("create subproject", err.Error())
		return
	}
	plan.JSON = jsonString(raw)
	alias := rtdapi.ExtractString(raw, "alias")
	if alias == "" {
		alias = plan.Child.ValueString()
	}
	plan.Alias = types.StringValue(alias)
	plan.ID = types.StringValue(plan.Parent.ValueString() + "/" + alias)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *subRes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var st subModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	raw, err := r.c.GetSubproject(st.Parent.ValueString(), st.Alias.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	st.JSON = jsonString(raw)
	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}
func (r *subRes) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("immutable", "subproject alias/child require replace")
}
func (r *subRes) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var st subModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	if err := r.c.DeleteSubproject(st.Parent.ValueString(), st.Alias.ValueString()); err != nil {
		resp.Diagnostics.AddError("delete subproject", err.Error())
	}
}

func NewSharingResource() resource.Resource { return &shareRes{} }

type shareRes struct{ c *rtdapi.Client }

type shareModel struct {
	ID          types.String `tfsdk:"id"`
	Project     types.String `tfsdk:"project"`
	AccessType  types.String `tfsdk:"access_type"`
	Description types.String `tfsdk:"description"`
	Password    types.String `tfsdk:"password"`
	AllowAll    types.Bool   `tfsdk:"allow_all"`
	Expires     types.String `tfsdk:"expires"`
	Versions    types.List   `tfsdk:"versions"`
	Token       types.String `tfsdk:"token"`
	URL         types.String `tfsdk:"url"`
	JSON        types.String `tfsdk:"json"`
}

func (r *shareRes) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sharing"
}
func (r *shareRes) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	r.c = clientOf(req.ProviderData)
}
func (r *shareRes) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Business-only sharing method (`/projects/{p}/sharing/`). access_type: token|password|http_header_token.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"project":     schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"access_type": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"description": schema.StringAttribute{Optional: true, Computed: true},
			"password":    schema.StringAttribute{Optional: true, Sensitive: true},
			"allow_all":   schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"expires":     schema.StringAttribute{Optional: true, Computed: true},
			"versions": schema.ListAttribute{ElementType: types.StringType, Optional: true, Computed: true,
				Default: listdefault.StaticValue(types.ListValueMust(types.StringType, nil))},
			"token": schema.StringAttribute{Computed: true, Sensitive: true},
			"url":   schema.StringAttribute{Computed: true},
			"json":  schema.StringAttribute{Computed: true},
		},
	}
}
func (r *shareRes) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan shareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	body := map[string]any{"access_type": plan.AccessType.ValueString()}
	put(body, "description", plan.Description)
	put(body, "password", plan.Password)
	put(body, "expires", plan.Expires)
	putBool(body, "allow_all", plan.AllowAll)
	if !plan.Versions.IsNull() && !plan.Versions.IsUnknown() {
		var vs []string
		_ = plan.Versions.ElementsAs(ctx, &vs, false)
		if len(vs) > 0 {
			body["versions"] = vs
		}
	}
	raw, err := r.c.CreateSharing(plan.Project.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("create sharing", err.Error())
		return
	}
	r.hydrateShare(raw, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *shareRes) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var st shareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	id, _ := strconv.Atoi(st.ID.ValueString())
	raw, err := r.c.GetSharing(st.Project.ValueString(), id)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	r.hydrateShare(raw, &st)
	resp.Diagnostics.Append(resp.State.Set(ctx, &st)...)
}
func (r *shareRes) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan shareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	id, _ := strconv.Atoi(plan.ID.ValueString())
	body := map[string]any{}
	put(body, "description", plan.Description)
	put(body, "expires", plan.Expires)
	putBool(body, "allow_all", plan.AllowAll)
	raw, err := r.c.UpdateSharing(plan.Project.ValueString(), id, body)
	if err != nil {
		resp.Diagnostics.AddError("update sharing", err.Error())
		return
	}
	r.hydrateShare(raw, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *shareRes) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var st shareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &st)...)
	id, _ := strconv.Atoi(st.ID.ValueString())
	if err := r.c.DeleteSharing(st.Project.ValueString(), id); err != nil {
		resp.Diagnostics.AddError("delete sharing", err.Error())
	}
}
func (r *shareRes) hydrateShare(raw json.RawMessage, m *shareModel) {
	m.JSON = jsonString(raw)
	id := rtdapi.ExtractInt(raw, "id", "pk")
	if id != 0 {
		m.ID = types.StringValue(strconv.Itoa(id))
	}
	m.Token = types.StringValue(rtdapi.ExtractString(raw, "token"))
	m.URL = types.StringValue(rtdapi.ExtractString(raw, "url"))
	if s := rtdapi.ExtractString(raw, "description"); s != "" {
		m.Description = types.StringValue(s)
	}
	if s := rtdapi.ExtractString(raw, "expires"); s != "" {
		m.Expires = types.StringValue(s)
	}
	if v, ok := mapBool(raw, "allow_all"); ok {
		m.AllowAll = types.BoolValue(v)
	}
	if m.Description.IsNull() || m.Description.IsUnknown() {
		m.Description = types.StringValue("")
	}
	if m.Expires.IsNull() || m.Expires.IsUnknown() {
		m.Expires = types.StringValue("")
	}
	if m.Versions.IsNull() || m.Versions.IsUnknown() {
		m.Versions = types.ListValueMust(types.StringType, nil)
	}
}

func mapBool(raw json.RawMessage, key string) (bool, bool) {
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return false, false
	}
	v, ok := m[key].(bool)
	return v, ok
}
