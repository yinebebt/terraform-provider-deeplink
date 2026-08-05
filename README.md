# terraform-provider-deeplink

[Terraform](https://www.terraform.io/) provider for [deeplink](https://github.com/yinebebt/deeplink). Manage short links on a deeplink server as code.

| | |
| --- | --- |
| Provider type | `deeplink` |
| Resource | `deeplink_link` |
| Source address | `yinebebt/deeplink` |

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.5
- A running [deeplink](https://github.com/yinebebt/deeplink) server (API key required for writes)

## Usage

```hcl
terraform {
  required_providers {
    deeplink = {
      source = "yinebebt/deeplink"
    }
  }
}

provider "deeplink" {
  endpoint = "https://link.example.com"
  api_key  = var.api_key
}

resource "deeplink_link" "docs" {
  type        = "redirect"
  url         = "https://example.com/docs"
  title       = "Docs"
  description = "Docs short link managed by Terraform"
}
```

```bash
terraform init
terraform plan
terraform apply
terraform destroy
```

A complete example lives in [`examples/`](./examples).

### Provider configuration

| Argument | Required | Description |
| --- | --- | --- |
| `endpoint` | yes | Base URL of the deeplink server (e.g. `https://link.example.com`) |
| `api_key` | no* | API key for mutating endpoints (`Sensitive`) |

\*Required for create / update / delete. Reads may work without it depending on server config.

### Resource: `deeplink_link`

| Attribute | Role |
| --- | --- |
| `type` | Processor type (e.g. `redirect`); forces replacement if changed |
| `url` | Destination URL |
| `title`, `description`, … | Optional metadata (see schema) |
| `id` | Canonical ID: `type/short_id` |
| `short_id`, `short_link` | Computed |

### Import

Import ID is always `type/short_id` (same as `id`):

```bash
terraform import \
  -var='endpoint=https://link.example.com' \
  -var="api_key=$DEEPLINK_API_KEY" \
  deeplink_link.docs \
  redirect/<short_id>
```

Flags must come before `ADDR` and `ID`.

## Examples

```bash
cd examples
cp terraform.tfvars.example terraform.tfvars   # set endpoint + api_key
terraform init
terraform apply
terraform destroy
```

Do not commit `terraform.tfvars`.

## Building from source

```bash
go test ./...
go build -o terraform-provider-deeplink
```

## Releasing

Releases are created by tagging `v*` (e.g. `v0.1.0`); GoReleaser builds and signs artifacts via [`.github/workflows/release.yml`](.github/workflows/release.yml).

## License

[MIT](./LICENSE)
