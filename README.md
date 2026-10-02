# Akash Provider Configurations

Community-maintained repository for Akash provider GPU configurations and hardware feature discovery.

## Table of Contents

**Contributing a GPU**
- [How to Contribute GPU Configurations](#how-to-contribute-gpu-configurations)
  - [When to Submit](#when-to-submit)
  - [Prerequisites](#prerequisites)
  - [Step 1: Verify GPU Models](#step-1-verify-gpu-models)
  - [Step 2: Collect GPU Details](#step-2-collect-gpu-details)
  - [Step 3: Extract Required Information](#step-3-extract-required-information)
  - [Step 4: Submit via Pull Request](#step-4-submit-via-pull-request)
  - [Step 5: Update Provider Configuration](#step-5-update-provider-configuration)
- [Naming Conventions](#naming-conventions)
- [Validation](#validation)
- [Never Rename or Remove Existing Entries](#never-rename-or-remove-existing-entries)

**Reviewing PRs**
- [Reviewer Checklist](#reviewer-checklist)

**Operating the Service**
- [Repository Structure](#repository-structure)
- [How It Works](#how-it-works)
  - [API Endpoints](#api-endpoints)
  - [Building the Server](#building-the-server)
  - [Deploying](#deploying)

**Help**
- [Maintainers](#maintainers)
- [Questions or Issues?](#questions-or-issues)

## Purpose

This repository serves as the central database for GPU hardware identification on the Akash Network. It enables:

- **Automatic GPU Detection**: Providers can automatically identify and advertise GPU models
- **Accurate Deployment Matching**: Tenants can discover and deploy to specific GPU models
- **Network-Wide GPU Discovery**: Real-time visibility into available GPU hardware across the network
- **Proper Resource Pricing**: Accurate GPU model identification ensures correct pricing

## How to Contribute GPU Configurations

### When to Submit

Submit GPU information if:

1. Your GPU model is **not listed** in [gpus.json](https://github.com/akash-network/provider-configs/blob/main/devices/pcie/gpus.json)
2. You're adding new GPU models to your provider
3. Your GPUs aren't being detected correctly

### Prerequisites

Before collecting GPU information:

- SSH access to each GPU-equipped node
- `provider-services` version **0.5.4 or higher** ([download here](https://github.com/akash-network/provider/releases))
- `jq` installed for JSON processing (`apt install -y jq`)

### Step 1: Verify GPU Models

Check if your GPU is already in the database:

1. Visit [gpus.json](https://github.com/akash-network/provider-configs/blob/main/devices/pcie/gpus.json)
2. Search for your GPU vendor ID and product ID
3. If found, no submission needed
4. If not found, proceed to Step 2

### Step 2: Collect GPU Details

Run this command on **each GPU node**:

```bash
provider-services tools psutil list gpu
```

**Example Output:**

```json
{
  "cards": [
    {
      "address": "0000:00:04.0",
      "index": 0,
      "pci": {
        "driver": "nvidia",
        "address": "0000:00:04.0",
        "vendor": {
          "id": "10de",
          "name": "NVIDIA Corporation"
        },
        "product": {
          "id": "1eb8",
          "name": "TU104GL [Tesla T4]"
        },
        "revision": "0xa1"
      }
    }
  ]
}
```

### Step 3: Extract Required Information

From the output, note:

- **Vendor ID**: `"id": "10de"` (NVIDIA in this example)
- **Product ID**: `"id": "1eb8"` (Tesla T4 in this example)
- **GPU Model Name**: `"name": "TU104GL [Tesla T4]"`

### Step 4: Submit via Pull Request

1. Fork this repository
2. Edit `devices/pcie/gpus.json`
3. Add your GPU under its vendor ID's `devices` map. The file is keyed by vendor ID, then device ID:

```json
{
  "10de": {
    "name": "nvidia",
    "devices": {
      "1eb8": {
        "name": "t4",
        "interface": "PCIe",
        "memory_size": "16Gi"
      }
    }
  }
}
```

**Field Guidelines:**

- **Vendor key** (e.g. `10de`): Vendor ID, lowercase hex, without "0x" prefix. Existing vendors: `10de` (nvidia), `1002` (amd).
- **Device key** (e.g. `1eb8`): Product ID, lowercase hex, without "0x" prefix.
- `name` (required): GPU model name (lowercase, alphanumeric, no spaces). Examples: `t4`, `a100`, `h100`, `rtx4090`
- `interface` (required): Physical interface / form factor. One of `PCIe`, `SXM`, `SXM2`, `SXM3`, `SXM4`, `SXM5`, `SXM6`.
- `memory_size`: GPU memory using the `Gi` suffix, e.g. `16Gi`, `80Gi`.

Multiple device IDs may map to the same `name` (e.g. memory or board variants of the same model). This is expected.

4. Create a pull request with title: `Add GPU: [Model Name]`
5. Include the full `provider-services tools psutil list gpu` output in the PR description

### Step 5: Update Provider Configuration

After your PR is merged, update your provider to use the new GPU attributes. See the [Provider Attributes Documentation](https://akash.network/docs/for-providers/operations/provider-attributes) for details.

## Naming Conventions

GPU model names should follow these guidelines:

- **Lowercase only**: `a100`, not `A100`
- **No spaces**: `rtx4090`, not `rtx 4090`
- **No special characters**: `h100`, not `H100-SXM` (form factor goes in the `interface` field, not the name)
- **Consistent with market naming**: Use common model designations

**Examples:**

- ✅ `t4`, `a100`, `h100`, `v100`, `rtx4090`, `rtx3090ti`
- ❌ `T4`, `A-100`, `H100 SXM`, `RTX_4090`

## Validation

Before submitting, verify:

1. **No duplicates**: Check if the vendor/device ID combo already exists
2. **Valid hex IDs**: Vendor and device IDs are 4-digit lowercase hex (without "0x")
3. **Lowercase name**: Model name is lowercase alphanumeric
4. **Required fields**: `name` and `interface` are set (the server rejects the whole file if either is empty)
5. **Valid JSON**: Your edit doesn't break JSON formatting

## Never Rename or Remove Existing Entries

Tenants reference GPU model names in their SDLs, and providers advertise them as attributes. Renaming or removing an existing `name` silently breaks deployment matching for every provider with that GPU. Only add new entries; if an existing entry is genuinely wrong, discuss it in an issue first.

## Reviewer Checklist

CI only checks that the file is valid JSON. It does **not** check structure, so reviewers should confirm:

- [ ] Entry is nested under the correct vendor key, inside `devices`
- [ ] Device ID matches the `psutil` output in the PR and the [PCI ID Repository](https://pci-ids.ucw.cz/read/PC/10de)
- [ ] `name` follows naming conventions and matches existing entries for the same model
- [ ] `interface` and `memory_size` are correct for that device ID
- [ ] No existing entries were renamed or removed

## Repository Structure

```
provider-configs/
├── devices/
│   └── pcie/
│       └── gpus.json   # GPU database (vendor/device IDs -> model names)
├── deploy/
│   └── gpu-database.yaml  # Kubernetes Deployment, Service, Ingress
└── httpServer/         # Go API server that publishes gpus.json
    ├── main.go
    ├── go.mod
    └── Dockerfile
```

## How It Works

```
PR merged to main
      │
      ▼
devices/pcie/gpus.json on GitHub (raw.githubusercontent.com)
      │  fetched every 5 minutes, or immediately on webhook
      ▼
httpServer (validates JSON, keeps last good copy in memory)
      │
      ▼
https://provider-configs.akash.network/devices/gpus
      │
      ▼
Akash provider inventory operator -> GPU attributes on provider nodes
```

- Changes are typically live within **5–10 minutes** of merge (5 minute poll interval plus GitHub's raw-file CDN cache).
- If a fetch returns invalid JSON or an entry fails validation, the server **keeps serving the previous good data** and increments `error_count`.
- The server only holds data in memory. If it restarts while `main` contains invalid JSON, it starts with no data and returns `503` until `main` is fixed.

### API Endpoints

| Endpoint | Method | Description |
|---|---|---|
| `/devices/gpus` | GET | Full GPU database (same structure as `gpus.json`) |
| `/devices/gpus/stats` | GET | `vendor_count`, `last_updated`, `update_count`, `error_count`, `has_data` |
| `/health` | GET | `200 healthy` when data is loaded, `503 unhealthy` otherwise; includes stats |
| `/devices/gpus/webhook` | POST | Triggers an immediate re-fetch from GitHub (unauthenticated; only triggers a fetch) |

```bash
curl https://provider-configs.akash.network/health
curl https://provider-configs.akash.network/devices/gpus
```

### Building the Server

```bash
cd httpServer
docker build -t pciedatabase .
```

The image listens on `:443` with a **self-signed certificate generated at build time (valid 365 days)**. Request path:

```
client -> Cloudflare -> ingress-nginx (akash-ingress-class) -> akash-gpu-database pod (:443, self-signed)
```

The ingress talks HTTPS to the pod with `proxy-ssl-verify: "false"`, so the pod certificate is never validated and its expiry does not cause an outage. The Ingress has no `tls:` section, so the Cloudflare -> ingress hop is served with the ingress controller's default certificate. If Cloudflare's SSL mode is ever changed to "Full (strict)", that hop will fail.

### Deploying

The Kubernetes manifest (Deployment, Service, Ingress in namespace `akash-services`) is in [`deploy/gpu-database.yaml`](deploy/gpu-database.yaml).

```bash
# build and push a new image tag, then bump the tag in deploy/gpu-database.yaml
docker build -t scarruthers/pciedatabase:v22 httpServer
docker push scarruthers/pciedatabase:v22

kubectl apply -f deploy/gpu-database.yaml
kubectl -n akash-services rollout status deploy/akash-gpu-database
kubectl -n akash-services logs deploy/akash-gpu-database
```

Note: data changes to `gpus.json` do **not** require a rebuild or redeploy; the running server picks them up from GitHub automatically. Rebuild only when `httpServer/` changes.

The Ingress routes `/devices/gpus` (prefix, which also covers `/stats` and `/webhook`) and `/health` (exact). Hosting and access details are maintained in internal operations docs.

## Maintainers

Pull requests are reviewed by [@akash-network/codeowners-devops](.github/CODEOWNERS).

## Questions or Issues?

- **Documentation**: [Akash Provider Attributes Guide](https://akash.network/docs/for-providers/operations/provider-attributes)
- **Support**: [Akash Discord #providers channel](https://discord.akash.network)
- **Issues**: [Open an issue](https://github.com/akash-network/provider-configs/issues)
