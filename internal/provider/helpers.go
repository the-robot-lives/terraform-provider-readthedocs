package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/the-robot-lives/terraform-provider-readthedocs/internal/rtdapi"
)

func strOrNull(s string) types.String {
	if s == "" {
		return types.StringValue("")
	}
	return types.StringValue(s)
}

func jsonString(raw json.RawMessage) types.String {
	if len(raw) == 0 {
		return types.StringValue("")
	}
	return types.StringValue(string(raw))
}

func optString(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func put(m map[string]any, k string, v types.String) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	if s := v.ValueString(); s != "" {
		m[k] = s
	}
}

func putBool(m map[string]any, k string, v types.Bool) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	m[k] = v.ValueBool()
}

func putInt(m map[string]any, k string, v types.Int64) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	m[k] = v.ValueInt64()
}

func rawListToStrings(items []json.RawMessage) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, string(it))
	}
	return out
}

func slugFromProjectJSON(raw json.RawMessage) string {
	if s := rtdapi.ExtractString(raw, "slug"); s != "" {
		return s
	}
	return rtdapi.NestedString(raw, "slug")
}
