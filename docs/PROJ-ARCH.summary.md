# PROJ-ARCH.summary — terraform-provider-readthedocs

- **What**: From-scratch Go Terraform/OpenTofu provider (`the-robot-lives/readthedocs`) for Read the
  Docs REST API v3, built on terraform-plugin-framework. Not a fork of BarnabyShearer/readthedocs.
- **Components**: `main.go` (plugin serve); `internal/provider` (provider config token/base_url;
  8 resources split across res_project/res_rest/res_more; 18 data sources via one generic datasource;
  helpers); `internal/rtdapi` (from-scratch REST client + tests).
- **Data flow**: Configure builds one shared client (token: provider block > `READTHEDOCS_TOKEN`;
  base_url: block > env > .org default); resources diff state vs remote GET, write via narrow v3 verbs
  (no project DELETE, no env-var PATCH, version = activate/hide only); `json` passthrough on most
  resources; list data sources return `result_count` + `results_json`.
- **Delivery**: local build via `scripts/build-provider.sh` into `~/.local/share/terraform/plugins`
  (dev_overrides/filesystem_mirror, no registry); CI on push/PR; goreleaser GPG-signed releases on
  `v*` tags.
- **Key decisions**: from-scratch client for full v3 coverage; generic datasource pattern; raw-JSON
  escape hatch; local mirror distribution; only implement verbs the public API actually offers.
- **No persistence layer**: remote state at RTD + plaintext tfstate (token/env-var/sharing values
  sensitive).
