terraform {
  required_providers {
    readthedocs = {
      source  = "the-robot-lives/readthedocs"
      version = "0.1.0"
    }
  }
}

provider "readthedocs" {}

resource "readthedocs_project" "github_utils" {
  name                   = "github-utils"
  repository_url         = "https://github.com/the-robot-lives/github-tools.git"
  repository_type        = "git"
  homepage               = "https://github.com/the-robot-lives/github-tools"
  programming_language   = "py"
  language               = "en"
  default_branch         = "mono-repo-dev"
  readthedocs_yaml_path  = ".readthedocs.yaml"
  tags                   = ["git", "submodules", "sphinx"]
}

resource "readthedocs_sync_versions" "once" {
  project = readthedocs_project.github_utils.id
  triggers = {
    reason = "import"
  }
}

resource "readthedocs_build" "latest" {
  project = readthedocs_project.github_utils.id
  version = "latest"
  triggers = {
    git_sha = "replace-me-to-rebuild"
  }
}

data "readthedocs_builds" "recent" {
  project = readthedocs_project.github_utils.id
}

output "docs_url" { value = readthedocs_project.github_utils.docs_url }
output "build_id" { value = readthedocs_build.latest.id }
output "recent_build_count" { value = data.readthedocs_builds.recent.result_count }
