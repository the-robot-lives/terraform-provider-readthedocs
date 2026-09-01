# Project Layout — terraform-provider-readthedocs

From-scratch Terraform/OpenTofu provider for the **Read the Docs API v3** (`the-robot-lives/readthedocs`).
Go + terraform-plugin-framework. ~2k LOC; single-level tree, no `layout/` extraction needed.

```text
terraform-provider-readthedocs/
├── main.go                        # Entry point: plugin serve (tf6server-style handshake via plugin-framework)
├── internal/
│   ├── provider/                  # Provider + all resource/datasource schemas → see below
│   │   ├── provider.go            # Provider def: token/base_url config, resource & datasource registry
│   │   ├── res_project.go         # readthedocs_project resource (CRUD on /projects/, no official DELETE)
│   │   ├── res_rest.go            # version, build (push/rebuild), sync_versions resources
│   │   ├── res_more.go            # redirect, environment_variable, subproject, sharing (Business) resources
│   │   ├── ds.go                  # All 18 data sources (project(s), version(s), build(s), org(s), remote, embed, …)
│   │   └── helpers.go             # Shared JSON/type helpers for schemas
│   └── rtdapi/
│       ├── client.go              # From-scratch REST client for RTD API v3 (auth, endpoints, JSON shapes)
│       └── client_test.go         # Unit tests for the client
├── examples/
│   └── github-utils/              # Working example config (main.tf) + local tfstate (gitignored upstream; present here)
├── scripts/
│   └── build-provider.sh          # House build: go build + install into local plugin mirror dirs
├── .github/
│   └── workflows/
│       ├── ci.yml                 # CI: build + test on push/PR
│       └── release.yml            # goreleaser release on v* tags
├── .goreleaser.yml                # Release config (multi-OS/arch binaries, checksums)
├── gpg-pubkey.asc                 # GPG public key for release signing (2EABB783…E16D0)
├── terraform-registry-manifest.json  # Registry metadata (only needed if ever published)
├── go.mod / go.sum                # Go module (terraform-plugin-framework deps)
├── Makefile                       # compile / test / install helpers
├── README.md                      # Start here: auth, resource/datasource tables, local install path
├── merge-notes.md                 # sep-1 branch-sweep notes (2026-09-01)
├── TODO.md                        # Remaining follow-ups
├── LICENSE                        # MIT
└── .gitignore                     # Build outputs, .terraform, tfstate
```

## Key Files Requiring Setup

| File | Action |
|------|--------|
| `READTHEDOCS_TOKEN` (env) | Required — API token, fed from monorepo `.envrc.dc` `secrets.readthedocs.token`; never commit |
| `~/.terraformrc` | Optional — `dev_overrides` + `filesystem_mirror` pointing at `~/.local/share/terraform/plugins` (see README) |
| `examples/github-utils/main.tf` | Example config; do not `apply` without a token |

## Notes

- No `docs/` existed previously; this file created 2026-09-01.
- `examples/github-utils/terraform.tfstate*` are local state artifacts — treat as transient, never commit values.
