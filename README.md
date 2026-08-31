# terraform-provider-readthedocs

Public repo: [the-robot-lives/terraform-provider-readthedocs](https://github.com/the-robot-lives/terraform-provider-readthedocs)

Source address: `the-robot-lives/readthedocs`  
(plugin protocol address: `registry.terraform.io/the-robot-lives/readthedocs`)

House install is a **local compile** into `~/.local/share/terraform/plugins` — the same path as SigNoz, `noizu/foryou`, and `noizu/google-marketing`. There is no HashiCorp Registry listing. GitHub `v*` releases exist for checksums if you ever want a registry later.

GPG public key: [`gpg-pubkey.asc`](gpg-pubkey.asc)  
Fingerprint: `2EABB783A4251C2A26FCD82E4CACEE95EB6E16D0`

From-scratch Terraform/OpenTofu provider for **Read the Docs API v3**.

Not a fork of [`BarnabyShearer/readthedocs`](https://registry.terraform.io/providers/BarnabyShearer/readthedocs) (abandoned 2022, projects only) and not a fork of any MCP. Client and schema are written from the official docs:

https://docs.readthedocs.com/platform/stable/api/v3.html

## Auth

```hcl
provider "readthedocs" {
  # token    = var.rtd_token   # or READTHEDOCS_TOKEN
  # base_url = "https://app.readthedocs.com/api/v3"  # Business; default is .org
}
```

`Authorization: Token …` — 60 req/min authenticated.

## Resources (write)

| Resource | API |
|----------|-----|
| `readthedocs_project` | `POST/GET/PATCH /projects/` (no official DELETE) |
| `readthedocs_version` | `PATCH /projects/{p}/versions/{v}/` |
| `readthedocs_build` | `POST /projects/{p}/versions/{v}/builds/` (push/rebuild) |
| `readthedocs_sync_versions` | `POST /projects/{p}/sync-versions/` |
| `readthedocs_redirect` | `GET/POST/PUT/DELETE /projects/{p}/redirects/` |
| `readthedocs_environment_variable` | `POST/GET/DELETE /environmentvariables/` (no PATCH) |
| `readthedocs_subproject` | `POST/GET/DELETE /subprojects/` |
| `readthedocs_sharing` | Business `GET/POST/PATCH/DELETE /sharing/` |

## Data sources (read)

`project`, `projects`, `version`, `versions`, `build`, `builds`, `redirects`, `environment_variables`, `subprojects`, `translations`, `organization`, `organizations`, `organization_projects`, `organization_teams`, `remote_organizations`, `remote_repositories`, `embed`, `superproject`.

List data sources expose `count` + `results_json` (raw API array).

## Not in public API v3 (cannot implement)

Custom domains, incoming VCS webhooks, outgoing webhooks, documented project DELETE.

## Local build (house path)

```bash
./scripts/build-provider.sh
```

Installs:

```text
~/.local/share/terraform/plugins/terraform-provider-readthedocs
~/.local/share/terraform/plugins/registry.opentofu.org/the-robot-lives/readthedocs/0.1.0/<os>_<arch>/
~/.local/share/terraform/plugins/registry.terraform.io/the-robot-lives/readthedocs/0.1.0/<os>_<arch>/
```

`~/.terraformrc` (`dev_overrides` skips `init` download; `filesystem_mirror` is the versioned path OpenTofu uses for `noizu/*`):

```hcl
provider_installation {
  dev_overrides {
    "the-robot-lives/readthedocs" = "~/.local/share/terraform/plugins"
  }
  filesystem_mirror {
    path    = "~/.local/share/terraform/plugins"
    include = ["the-robot-lives/readthedocs"]
  }
  direct {}
}
```

Example: [`examples/github-utils/main.tf`](examples/github-utils/main.tf). Needs `READTHEDOCS_TOKEN` (or `token` in the provider block). Do not apply without a token.
