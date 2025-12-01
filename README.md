# Akash Provider Configurations

Community-maintained repository for Akash provider GPU configurations and hardware feature discovery.

## Purpose

This repository serves as the central database for GPU hardware identification on the Akash Network. It enables:

- **Automatic GPU Detection**: Providers can automatically identify and advertise GPU models
- **Accurate Deployment Matching**: Tenants can discover and deploy to specific GPU models
- **Network-Wide GPU Discovery**: Real-time visibility into available GPU hardware across the network
- **Proper Resource Pricing**: Accurate GPU model identification ensures correct pricing

## Repository Structure

```
provider-configs/
├── devices/
│   └── pcie/           # PCIe device configurations
├── httpServer/         # Configuration server
└── gpus.json          # GPU database (vendor/product IDs)
```

## How to Contribute GPU Configurations

### When to Submit

Submit GPU information if:

1. Your GPU model is **not listed** in [gpus.json](https://github.com/akash-network/provider-configs/blob/main/gpus.json)
2. You're adding new GPU models to your provider
3. Your GPUs aren't being detected correctly

### Prerequisites

Before collecting GPU information:

- SSH access to each GPU-equipped node
- `provider-services` version **0.5.4 or higher** ([download here](https://github.com/akash-network/provider/releases))
- `jq` installed for JSON processing (`apt install -y jq`)

### Step 1: Verify GPU Models

Check if your GPU is already in the database:

1. Visit [gpus.json](https://github.com/akash-network/provider-configs/blob/main/gpus.json)
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
2. Edit `gpus.json`
3. Add your GPU entry in the following format:

```json
{
  "vendor": "10de",
  "device": "1eb8",
  "name": "t4"
}
```

**Field Guidelines:**

- `vendor`: Vendor ID (lowercase hex, without "0x" prefix)
- `device`: Product ID (lowercase hex, without "0x" prefix)
- `name`: GPU model name (lowercase, alphanumeric, no spaces)
  - Examples: `t4`, `a100`, `h100`, `rtx4090`

4. Create a pull request with title: `Add GPU: [Model Name]`
5. Include the full `provider-services tools psutil list gpu` output in the PR description

### Step 5: Update Provider Configuration

After your PR is merged, update your provider to use the new GPU attributes. See the [Provider Attributes Documentation](https://akash.network/docs/for-providers/operations/provider-attributes) for details.

## Naming Conventions

GPU model names should follow these guidelines:

- **Lowercase only**: `a100`, not `A100`
- **No spaces**: `rtx4090`, not `rtx 4090`
- **No special characters**: `h100`, not `H100-SXM`
- **Consistent with market naming**: Use common model designations

**Examples:**

- ✅ `t4`, `a100`, `h100`, `v100`, `rtx4090`, `rtx3090ti`
- ❌ `T4`, `A-100`, `H100 SXM`, `RTX_4090`

## Validation

Before submitting, verify:

1. **No duplicates**: Check if the vendor/device ID combo already exists
2. **Valid hex IDs**: Vendor and device IDs are 4-digit lowercase hex (without "0x")
3. **Lowercase name**: Model name is lowercase alphanumeric
4. **Valid JSON**: Your edit doesn't break JSON formatting

## Questions or Issues?

- **Documentation**: [Akash Provider Attributes Guide](https://akash.network/docs/for-providers/operations/provider-attributes)
- **Support**: [Akash Discord #providers channel](https://discord.akash.network)
- **Issues**: [Open an issue](https://github.com/akash-network/provider-configs/issues)


