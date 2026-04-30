#!/usr/bin/env bash
# Reset an OpenStack project to a clean state.
#
# Deletes all user-created resources and leaves only:
#   - The "default" security group (auto-created by OpenStack)
#   - The external/public network (typically named "PUBLICNET" or similar)
#
# The script runs in two phases:
#   1. Discovery — scans every resource type and prints a full inventory
#   2. Deletion  — after explicit confirmation, deletes in dependency order
#
# Deletion order follows the OpenStack dependency graph (children first):
#
#   1. Servers          — top-level; hold ports, volumes, FIPs, SG refs
#   2. Load balancers   — hold VIP ports, may own FIPs
#   3. Floating IPs     — attached to ports that are now gone
#   4. Volume snapshots — must go before the volumes they reference
#   5. Volumes          — detached once servers are deleted
#   6. Ports            — leftover user/LB ports; must go before routers & SGs
#   7. Routers          — detach subnet interfaces, unset gateway, delete
#   8. Subnets          — can't delete while router interfaces exist
#   9. Networks         — can't delete while subnets exist
#  10. Security groups  — can't delete while any port references them
#  11. Keypairs         — independent, no ordering constraint
#
# Requires:
#   - openstack CLI installed and authenticated (clouds.yaml)
#   - jq for JSON parsing
#
# Usage:
#   hack/scripts/openstack-reset.sh --os-cloud mycloud
#   hack/scripts/openstack-reset.sh --os-cloud mycloud --dry-run
#   hack/scripts/openstack-reset.sh --os-cloud mycloud --force

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

PRESERVE_NETWORK="${OPENSTACK_PRESERVE_NETWORK:-PUBLICNET}"
PRESERVE_SG="default"

FORCE=false
DRY_RUN=false
OS_CLOUD_FLAG=""

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------

while [[ $# -gt 0 ]]; do
  case "$1" in
    --os-cloud)
      if [[ -z "${2:-}" ]]; then
        echo "Error: --os-cloud requires a value." >&2
        exit 1
      fi
      OS_CLOUD_FLAG="$2"
      shift 2
      ;;
    --os-cloud=*)
      OS_CLOUD_FLAG="${1#--os-cloud=}"
      shift
      ;;
    --force)
      FORCE=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 --os-cloud <cloud-name> [--force] [--dry-run]"
      echo ""
      echo "Required:"
      echo "  --os-cloud <name>   clouds.yaml profile to target (passed to every openstack call)"
      echo ""
      echo "Flags:"
      echo "  --force             Skip confirmation prompt"
      echo "  --dry-run           List resources that would be deleted (same as normal but no prompt)"
      echo ""
      echo "Environment:"
      echo "  OPENSTACK_PRESERVE_NETWORK   External network name to keep (default: PUBLICNET)"
      exit 0
      ;;
    *)
      echo "Unknown flag: $1" >&2
      echo "Run '$0 --help' for usage." >&2
      exit 1
      ;;
  esac
done

if [[ -z "$OS_CLOUD_FLAG" ]]; then
  echo "Error: --os-cloud is required." >&2
  echo "Run '$0 --help' for usage." >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Preflight
# ---------------------------------------------------------------------------

for cmd in openstack jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "Error: '$cmd' is required but not found in PATH." >&2
    exit 1
  fi
done

# Every openstack CLI call goes through this wrapper so --os-cloud is always set.
os() {
  openstack --os-cloud "$OS_CLOUD_FLAG" "$@"
}

# Normalize JSON keys to lowercase so jq expressions work regardless of
# openstackclient version. Some versions return "ID"/"Name", others "id"/"name".
normalize_keys() {
  jq '[.[] | with_entries(.key |= ascii_downcase)]'
}

TOKEN_INFO=$(os token issue -f json 2>/dev/null || echo "{}")
OS_PROJECT=$(echo "$TOKEN_INFO" | jq -r '.project_id // "unknown"')
OS_USER=$(echo "$TOKEN_INFO" | jq -r '.user_id // "unknown"')

echo "OpenStack project reset"
echo "======================="
echo "  Cloud:   ${OS_CLOUD_FLAG}"
echo "  Project: ${OS_PROJECT}"
echo "  User:    ${OS_USER}"
echo ""
echo "Preserving: security group '${PRESERVE_SG}', network '${PRESERVE_NETWORK}'"
echo ""

# ===========================================================================
# Phase 1: Discovery — scan all resource types, build the inventory
# ===========================================================================

echo "Scanning resources..."
echo ""

total_to_delete=0

# --- Servers ---------------------------------------------------------------
servers_json=$(os server list -f json 2>/dev/null | normalize_keys || echo "[]")
servers_count=$(echo "$servers_json" | jq 'length')

echo "  Servers:            ${servers_count}"
if [[ "$servers_count" -gt 0 ]]; then
  echo "$servers_json" | jq -r '.[] | "    • \(.name) (\(.status)) [\(.id)]"'
  ((total_to_delete += servers_count)) || true
fi

# --- Load balancers --------------------------------------------------------
lbs_json=$(os loadbalancer list -f json 2>/dev/null | normalize_keys || echo "[]")
lbs_count=$(echo "$lbs_json" | jq 'length')

echo "  Load Balancers:     ${lbs_count}"
if [[ "$lbs_count" -gt 0 ]]; then
  echo "$lbs_json" | jq -r '.[] | "    • \(.name) (\(.provisioning_status // .operating_status // "unknown")) [\(.id)]"'
  ((total_to_delete += lbs_count)) || true
fi

# --- Floating IPs ----------------------------------------------------------
fips_json=$(os floating ip list -f json 2>/dev/null | normalize_keys || echo "[]")
fips_count=$(echo "$fips_json" | jq 'length')

echo "  Floating IPs:       ${fips_count}"
if [[ "$fips_count" -gt 0 ]]; then
  echo "$fips_json" | jq -r '.[] | "    • \(.["floating ip address"] // .floating_ip_address // .id) (\(.status // "unknown")) [\(.id)]"'
  ((total_to_delete += fips_count)) || true
fi

# --- Volume snapshots ------------------------------------------------------
snaps_json=$(os volume snapshot list -f json 2>/dev/null | normalize_keys || echo "[]")
snaps_count=$(echo "$snaps_json" | jq 'length')

echo "  Volume Snapshots:   ${snaps_count}"
if [[ "$snaps_count" -gt 0 ]]; then
  echo "$snaps_json" | jq -r '.[] | "    • \(.name // "<unnamed>") (\(.status)) [\(.id)]"'
  ((total_to_delete += snaps_count)) || true
fi

# --- Volumes ---------------------------------------------------------------
vols_json=$(os volume list -f json 2>/dev/null | normalize_keys || echo "[]")
vols_count=$(echo "$vols_json" | jq 'length')

echo "  Volumes:            ${vols_count}"
if [[ "$vols_count" -gt 0 ]]; then
  echo "$vols_json" | jq -r '.[] | "    • \(.name // "<unnamed>") (\(.status)) [\(.id)]"'
  ((total_to_delete += vols_count)) || true
fi

# --- Ports (user-owned only) -----------------------------------------------
ports_json=$(os port list -f json 2>/dev/null | normalize_keys || echo "[]")
user_ports_json=$(echo "$ports_json" | jq '[.[] | select(.["device owner"] // .device_owner // "" | test("^(network:|compute:)") | not)]')
ports_count=$(echo "$user_ports_json" | jq 'length')

echo "  Ports (user):       ${ports_count}"
if [[ "$ports_count" -gt 0 ]]; then
  echo "$user_ports_json" | jq -r '.[] | "    • \(.name // "<unnamed>") (owner: \(.["device owner"] // .device_owner // "none")) [\(.id)]"'
  ((total_to_delete += ports_count)) || true
fi

# --- Routers ---------------------------------------------------------------
routers_json=$(os router list -f json 2>/dev/null | normalize_keys || echo "[]")
routers_count=$(echo "$routers_json" | jq 'length')

echo "  Routers:            ${routers_count}"
if [[ "$routers_count" -gt 0 ]]; then
  echo "$routers_json" | jq -r '.[] | "    • \(.name) (\(.status)) [\(.id)]"'
  ((total_to_delete += routers_count)) || true
fi

# --- Subnets (excluding preserved network) ---------------------------------
subnets_json=$(os subnet list -f json 2>/dev/null | normalize_keys || echo "[]")
subnets_to_delete="[]"
subnets_to_skip=""
if [[ $(echo "$subnets_json" | jq 'length') -gt 0 ]]; then
  while IFS=$'\t' read -r sid sname snet_id; do
    net_name=$(os network show "$snet_id" -f value -c name 2>/dev/null || echo "")
    if echo "$net_name" | grep -qi "^${PRESERVE_NETWORK}$"; then
      subnets_to_skip="${subnets_to_skip}    ✓ ${sname} (belongs to ${PRESERVE_NETWORK}) — kept\n"
    else
      subnets_to_delete=$(echo "$subnets_to_delete" | jq --arg id "$sid" --arg name "$sname" '. + [{"id": $id, "name": $name}]')
    fi
  done < <(echo "$subnets_json" | jq -r '.[] | "\(.id)\t\(.name)\t\(.network // .network_id)"')
fi
subnets_count=$(echo "$subnets_to_delete" | jq 'length')

echo "  Subnets:            ${subnets_count}"
if [[ "$subnets_count" -gt 0 ]]; then
  echo "$subnets_to_delete" | jq -r '.[] | "    • \(.name) [\(.id)]"'
  ((total_to_delete += subnets_count)) || true
fi
if [[ -n "$subnets_to_skip" ]]; then
  printf '%b' "$subnets_to_skip" | sed '$ { /^$/d; }'
fi

# --- Networks (excluding preserved) ----------------------------------------
networks_json=$(os network list -f json 2>/dev/null | normalize_keys || echo "[]")
networks_to_delete=$(echo "$networks_json" | jq --arg pn "$PRESERVE_NETWORK" '[.[] | select(.name | test("^" + $pn + "$"; "i") | not)]')
networks_preserved=$(echo "$networks_json" | jq --arg pn "$PRESERVE_NETWORK" '[.[] | select(.name | test("^" + $pn + "$"; "i"))]')
networks_count=$(echo "$networks_to_delete" | jq 'length')

echo "  Networks:           ${networks_count}"
if [[ "$networks_count" -gt 0 ]]; then
  echo "$networks_to_delete" | jq -r '.[] | "    • \(.name) [\(.id)]"'
  ((total_to_delete += networks_count)) || true
fi
if [[ $(echo "$networks_preserved" | jq 'length') -gt 0 ]]; then
  echo "$networks_preserved" | jq -r '.[] | "    ✓ \(.name) — kept"'
fi

# --- Security groups (excluding default) -----------------------------------
sgs_json=$(os security group list -f json 2>/dev/null | normalize_keys || echo "[]")
sgs_to_delete=$(echo "$sgs_json" | jq --arg psg "$PRESERVE_SG" '[.[] | select(.name | ascii_downcase != ($psg | ascii_downcase))]')
sgs_preserved=$(echo "$sgs_json" | jq --arg psg "$PRESERVE_SG" '[.[] | select(.name | ascii_downcase == ($psg | ascii_downcase))]')
sgs_count=$(echo "$sgs_to_delete" | jq 'length')

echo "  Security Groups:    ${sgs_count}"
if [[ "$sgs_count" -gt 0 ]]; then
  echo "$sgs_to_delete" | jq -r '.[] | "    • \(.name) [\(.id)]"'
  ((total_to_delete += sgs_count)) || true
fi
if [[ $(echo "$sgs_preserved" | jq 'length') -gt 0 ]]; then
  echo "$sgs_preserved" | jq -r '.[] | "    ✓ \(.name) — kept"'
fi

# --- Keypairs --------------------------------------------------------------
keypairs_json=$(os keypair list -f json 2>/dev/null | normalize_keys || echo "[]")
keypairs_count=$(echo "$keypairs_json" | jq 'length')

echo "  Keypairs:           ${keypairs_count}"
if [[ "$keypairs_count" -gt 0 ]]; then
  echo "$keypairs_json" | jq -r '.[] | "    • \(.name)"'
  ((total_to_delete += keypairs_count)) || true
fi

echo ""
echo "-----------------------------------------------------------------------"
echo "Total resources to delete: ${total_to_delete}"
echo "-----------------------------------------------------------------------"
echo ""

# ---------------------------------------------------------------------------
# Nothing to do?
# ---------------------------------------------------------------------------

if [[ "$total_to_delete" -eq 0 ]]; then
  echo "Nothing to delete. Project is already clean."
  exit 0
fi

# ---------------------------------------------------------------------------
# Confirmation (default: No — must type exact "YeS")
# ---------------------------------------------------------------------------

if [[ "$DRY_RUN" == true ]]; then
  echo "[DRY RUN] No resources were deleted."
  exit 0
fi

if [[ "$FORCE" != true ]]; then
  read -rp "Delete these ${total_to_delete} resource(s) from cloud '${OS_CLOUD_FLAG}'? Type 'YeS' to confirm: " confirm
  if [[ "$confirm" != "YeS" ]]; then
    echo "Aborted. No resources were deleted."
    exit 0
  fi
  echo ""
fi

# ===========================================================================
# Phase 2: Deletion — work through the dependency order
# ===========================================================================

count=0
fail_count=0

delete_resource() {
  local resource_type="$1"
  local resource_id="$2"
  local resource_name="${3:-$resource_id}"

  echo "  Deleting ${resource_type}: ${resource_name} (${resource_id})"
  if os "$resource_type" delete "$resource_id" 2>/dev/null; then
    ((count++)) || true
  else
    echo "    ⚠ Failed to delete ${resource_type} ${resource_name}" >&2
    ((fail_count++)) || true
  fi
}

# --- 1. Servers ------------------------------------------------------------

if [[ "$servers_count" -gt 0 ]]; then
  echo "→ Deleting servers..."
  echo "$servers_json" | jq -r '.[] | "\(.id) \(.name)"' | while read -r id name; do
    delete_resource "server" "$id" "$name"
  done
  echo "  Waiting for servers to terminate..."
  sleep 10
fi

# --- 2. Load balancers -----------------------------------------------------

if [[ "$lbs_count" -gt 0 ]]; then
  echo "→ Deleting load balancers..."
  echo "$lbs_json" | jq -r '.[] | "\(.id) \(.name)"' | while read -r id name; do
    echo "  Deleting load balancer: ${name} (${id}) [cascade]"
    os loadbalancer delete --cascade "$id" 2>/dev/null || {
      echo "    ⚠ Failed to delete load balancer ${name}" >&2
      ((fail_count++)) || true
    }
    ((count++)) || true
  done
  echo "  Waiting for load balancers to delete..."
  sleep 15
fi

# --- 3. Floating IPs -------------------------------------------------------

if [[ "$fips_count" -gt 0 ]]; then
  echo "→ Releasing floating IPs..."
  echo "$fips_json" | jq -r '.[] | .id' | while read -r id; do
    delete_resource "floating ip" "$id" "$id"
  done
fi

# --- 4. Volume snapshots ---------------------------------------------------

if [[ "$snaps_count" -gt 0 ]]; then
  echo "→ Deleting volume snapshots..."
  echo "$snaps_json" | jq -r '.[] | "\(.id) \(.name)"' | while read -r id name; do
    delete_resource "volume snapshot" "$id" "$name"
  done
fi

# --- 5. Volumes ------------------------------------------------------------

if [[ "$vols_count" -gt 0 ]]; then
  echo "→ Deleting volumes..."
  echo "$vols_json" | jq -r '.[] | "\(.id) \(.name)"' | while read -r id name; do
    echo "  Deleting volume: ${name} (${id})"
    os volume delete --force "$id" 2>/dev/null || {
      echo "    ⚠ Failed to delete volume ${name}" >&2
      ((fail_count++)) || true
    }
    ((count++)) || true
  done
fi

# --- 6. Ports --------------------------------------------------------------

if [[ "$ports_count" -gt 0 ]]; then
  echo "→ Deleting user ports..."
  echo "$user_ports_json" | jq -r '.[] | "\(.id) \(.name // .id)"' | while read -r id name; do
    delete_resource "port" "$id" "$name"
  done
fi

# --- 7. Routers ------------------------------------------------------------

if [[ "$routers_count" -gt 0 ]]; then
  echo "→ Deleting routers..."
  echo "$routers_json" | jq -r '.[] | "\(.id) \(.name)"' | while read -r router_id router_name; do
    echo "  Processing router: ${router_name} (${router_id})"

    # Detach all subnet interfaces from the router.
    router_ports=$(os port list --router "$router_id" -f json 2>/dev/null | normalize_keys || echo "[]")
    echo "$router_ports" | jq -r '.[] | select(.["fixed ip addresses"] // .fixed_ips // null | . != null) | .["fixed ip addresses"] // .fixed_ips' | \
      jq -r '.[].subnet_id' 2>/dev/null | sort -u | while read -r subnet_id; do
        echo "    Detaching subnet ${subnet_id} from router"
        os router remove subnet "$router_id" "$subnet_id" 2>/dev/null || true
      done

    # Unset external gateway.
    os router unset --external-gateway "$router_id" 2>/dev/null || true

    delete_resource "router" "$router_id" "$router_name"
  done
fi

# --- 8. Subnets ------------------------------------------------------------

if [[ "$subnets_count" -gt 0 ]]; then
  echo "→ Deleting subnets..."
  echo "$subnets_to_delete" | jq -r '.[] | "\(.id) \(.name)"' | while read -r id name; do
    delete_resource "subnet" "$id" "$name"
  done
fi

# --- 9. Networks -----------------------------------------------------------

if [[ "$networks_count" -gt 0 ]]; then
  echo "→ Deleting networks..."
  echo "$networks_to_delete" | jq -r '.[] | "\(.id) \(.name)"' | while read -r id name; do
    delete_resource "network" "$id" "$name"
  done
fi

# --- 10. Security groups ---------------------------------------------------

if [[ "$sgs_count" -gt 0 ]]; then
  echo "→ Deleting security groups..."
  echo "$sgs_to_delete" | jq -r '.[] | "\(.id) \(.name)"' | while read -r id name; do
    delete_resource "security group" "$id" "$name"
  done
fi

# --- 11. Keypairs ----------------------------------------------------------

if [[ "$keypairs_count" -gt 0 ]]; then
  echo "→ Deleting keypairs..."
  echo "$keypairs_json" | jq -r '.[].name' | while read -r name; do
    delete_resource "keypair" "$name" "$name"
  done
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

echo ""
echo "======================="
echo "Reset complete. Deleted ${count} resource(s)."
if [[ "$fail_count" -gt 0 ]]; then
  echo "⚠ ${fail_count} resource(s) failed to delete. Re-run or check manually."
fi
echo "Cloud: ${OS_CLOUD_FLAG} | Project: ${OS_PROJECT}"
echo "Preserved: security group '${PRESERVE_SG}', network '${PRESERVE_NETWORK}'"
