#!/usr/bin/env bash

set -euo pipefail

validation_script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
plan_file=${1:-"${validation_script_dir}/MOBILE_APP_REVAMP_EXECUTION_PLAN.md"}

fail() {
  echo "mobile revamp plan validation failed: $*" >&2
  exit 1
}

assert_equal() {
  local label=$1
  local actual=$2
  local expected=$3

  if [[ "$actual" != "$expected" ]]; then
    fail "${label}: expected ${expected}, got ${actual}"
  fi
}

[[ -f "$plan_file" ]] || fail "plan file not found: ${plan_file}"
command -v perl >/dev/null || fail "perl is required"
command -v tsort >/dev/null || fail "tsort is required"

read -r priority_status critical_status medium_status priority_allocations < <(
  awk '
    /^\| Priority \| Current status/ { in_priority = 1; next }
    in_priority && /^\| [0-9]+ / { priority_status++ }
    in_priority && /^### Appendix C critical/ { in_priority = 0 }

    /^### Appendix C critical/ { in_critical = 1; next }
    in_critical && /^\| C\/H-[0-9][0-9] / { critical_status++ }
    in_critical && /^### Appendix C medium/ {
      in_critical = 0
      in_medium = 1
      next
    }

    in_medium && /^\| M-[0-9][0-9] / { medium_status++ }
    in_medium && /^### Dependency sequence/ { in_medium = 0 }

    /^\| Strategy priority \| Owning milestones \|/ {
      in_priority_allocations = 1
      next
    }
    in_priority_allocations && /^\| [0-9]+ \|/ { priority_allocations++ }
    in_priority_allocations && /^\| Appendix C findings/ {
      in_priority_allocations = 0
    }

    END {
      print priority_status + 0, critical_status + 0, medium_status + 0, priority_allocations + 0
    }
  ' "$plan_file"
)

assert_equal "priority status rows" "$priority_status" 25
assert_equal "critical/high status rows" "$critical_status" 35
assert_equal "medium status rows" "$medium_status" 37
assert_equal "priority allocation rows" "$priority_allocations" 25

perl -ne '
  if (/^\| Priority \| Current status/) { $in_priority_status = 1; next }
  if ($in_priority_status && /^### Appendix C critical/) { $in_priority_status = 0 }
  $priority_status{$1}++ if $in_priority_status && /^\| (\d+) /;

  if (/^### Appendix C critical/) { $in_critical_status = 1; next }
  if ($in_critical_status && /^### Appendix C medium/) {
    $in_critical_status = 0;
    $in_medium_status = 1;
    next;
  }
  $critical_status{$1}++ if $in_critical_status && /^\| C\/H-(\d\d) /;

  if ($in_medium_status && /^### Dependency sequence/) { $in_medium_status = 0 }
  $medium_status{$1}++ if $in_medium_status && /^\| M-(\d\d) /;

  if (/^\| Strategy priority \| Owning milestones \|/) {
    $in_priority_allocations = 1;
    next;
  }
  if ($in_priority_allocations && /^\| Appendix C findings/) {
    $in_priority_allocations = 0;
  }
  $priority_allocations{$1}++ if $in_priority_allocations && /^\| (\d+) \|/;

  END {
    my @errors;
    for my $number (1 .. 25) {
      push @errors, "priority status $number occurs " . ($priority_status{$number} // 0) . " times"
        unless ($priority_status{$number} // 0) == 1;
      push @errors, "priority allocation $number occurs " . ($priority_allocations{$number} // 0) . " times"
        unless ($priority_allocations{$number} // 0) == 1;
    }
    for my $number (1 .. 35) {
      my $key = sprintf("%02d", $number);
      push @errors, "C/H-$key status occurs " . ($critical_status{$key} // 0) . " times"
        unless ($critical_status{$key} // 0) == 1;
    }
    for my $number (1 .. 37) {
      my $key = sprintf("%02d", $number);
      push @errors, "M-$key status occurs " . ($medium_status{$key} // 0) . " times"
        unless ($medium_status{$key} // 0) == 1;
    }
    if (@errors) {
      print STDERR "mobile revamp plan validation failed: " . join(", ", @errors) . "\n";
      exit 1;
    }
  }
' "$plan_file"

mapfile -t catalog_ids < <(
  awk '
    /^\| ID \| Goal \| Class \| Depends on \|/ { in_catalog = 1; next }
    in_catalog && /^### Traceability/ { in_catalog = 0 }
    in_catalog && /^\| [A-Z][A-Z0-9-]* \|/ { print $2 }
  ' "$plan_file"
)

assert_equal "milestone catalog rows" "${#catalog_ids[@]}" 60

duplicate_ids=$(printf '%s\n' "${catalog_ids[@]}" | sort | uniq -d)
[[ -z "$duplicate_ids" ]] || fail "duplicate milestone IDs: ${duplicate_ids//$'\n'/, }"

perl -ne '
  push @lines, $_;
  if (/^\| ID \| Goal \| Class \| Depends on \|/) { $in_catalog = 1; next }
  if ($in_catalog && /^### Traceability/) { $in_catalog = 0 }
  $ids{$1} = 1 if $in_catalog && /^\| ([A-Z][A-Z0-9-]*) \|/;

  END {
    my (%unknown, $in_authoritative_catalog);
    for my $line (@lines) {
      $in_authoritative_catalog = 1 if $line =~ /^## Authoritative strategy coverage/;
      $in_authoritative_catalog = 0 if $line =~ /^## Phase 0 /;
      next unless $in_authoritative_catalog;
      while ($line =~ /`([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)+)`/g) {
        $unknown{$1} = 1 unless $ids{$1};
      }
    }
    if (%unknown) {
      print STDERR "mobile revamp plan validation failed: unknown exact milestone references: " .
        join(", ", sort keys %unknown) . "\n";
      exit 1;
    }
  }
' "$plan_file"

for milestone_id in "${catalog_ids[@]}"; do
  acceptance_count=$(grep -Fc -- "- \`${milestone_id}\`:" "$plan_file" || true)
  assert_equal "acceptance boundaries for ${milestone_id}" "$acceptance_count" 1
done

perl -ne '
  if (/^\| Appendix C findings \|/) { $in_allocations = 1; next }
  if ($in_allocations && /^Milestone acceptance evidence/) { $in_allocations = 0 }
  next unless $in_allocations && /^\| ((?:C\/H|M)-[^| ]+) \|/;

  my $allocation = $1;
  while ($allocation =~ /(C\/H|M)-(\d+)(?:[^0-9|]+(\d+))?/g) {
    my ($kind, $start, $end) = ($1, $2, defined($3) ? $3 : $2);
    $seen{"$kind-" . sprintf("%02d", $_)}++ for $start .. $end;
  }

  END {
    my @errors;
    for my $number (1 .. 35) {
      my $id = sprintf("C/H-%02d", $number);
      push @errors, "$id missing" unless $seen{$id};
      push @errors, "$id allocated $seen{$id} times" if ($seen{$id} // 0) > 1;
    }
    for my $number (1 .. 37) {
      my $id = sprintf("M-%02d", $number);
      push @errors, "$id missing" unless $seen{$id};
      push @errors, "$id allocated $seen{$id} times" if ($seen{$id} // 0) > 1;
    }
    if (@errors) {
      print STDERR "mobile revamp plan validation failed: " . join(", ", @errors) . "\n";
      exit 1;
    }
  }
' "$plan_file"

expected_edges=$(sed -n 's/.*The \([0-9][0-9]*\) explicit dependency edges are acyclic.*/\1/p' "$plan_file")
[[ "$expected_edges" =~ ^[0-9]+$ ]] || fail "catalog audit does not state an edge count"

actual_edges=$(
  perl -ne '
    if (/^\| ID \| Goal \| Class \| Depends on \|/) { $in_catalog = 1; next }
    if ($in_catalog && /^### Traceability/) { $in_catalog = 0 }
    if ($in_catalog && /^\| ([A-Z][A-Z0-9-]*) \|.*\| ([^|]*) \|\s*$/) {
      push @rows, [$1, $2];
      $ids{$1} = 1;
    }

    END {
      my (@unknown, $edges);
      for my $row (@rows) {
        my ($id, $dependencies) = @$row;
        my %seen;
        while ($dependencies =~ /\b([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)+)\b/g) {
          next if $seen{$1}++;
          $edges++;
          push @unknown, "$id->$1" unless $ids{$1};
        }
      }
      if (@unknown) {
        print STDERR "mobile revamp plan validation failed: unknown dependency references: " .
          join(", ", @unknown) . "\n";
        exit 1;
      }
      print "$edges\n";
    }
  ' "$plan_file"
)

assert_equal "dependency edge count" "$actual_edges" "$expected_edges"

expected_cutover_prerequisites=$(
  sed -n 's/.*`PHASE4-CUTOVER` has \([0-9][0-9]*\) enumerated direct prerequisites.*/\1/p' "$plan_file"
)
[[ "$expected_cutover_prerequisites" =~ ^[0-9]+$ ]] ||
  fail "catalog audit does not state a Phase 4 prerequisite count"

actual_cutover_prerequisites=$(
  awk '
    /^\| ID \| Goal \| Class \| Depends on \|/ { in_catalog = 1; next }
    in_catalog && /^### Traceability/ { in_catalog = 0 }
    in_catalog && /^\| [A-Z][A-Z0-9-]* \|/ && $0 ~ /CUTOVER/ && $2 != "PHASE4-CUTOVER" {
      cutover++
    }
    END { print cutover + 0 }
  ' "$plan_file"
)
assert_equal \
  "Phase 4 prerequisite count" \
  "$actual_cutover_prerequisites" \
  "$expected_cutover_prerequisites"

perl -ne '
  if (/^\| ID \| Goal \| Class \| Depends on \|/) { $in_catalog = 1; next }
  if ($in_catalog && /^### Traceability/) { $in_catalog = 0 }
  if ($in_catalog && /^\| ([A-Z][A-Z0-9-]*) \|.*\| ([^|]*) \|\s*$/) {
    my ($id, $dependencies) = ($1, $2);
    my %seen;
    while ($dependencies =~ /\b([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)+)\b/g) {
      print "$1 $id\n" unless $seen{$1}++;
    }
  }
' "$plan_file" | tsort >/dev/null || fail "dependency graph contains a cycle"

perl -ne '
  if (/^\| ID \| Goal \| Class \| Depends on \|/) { $in_catalog = 1; next }
  if ($in_catalog && /^### Traceability/) { $in_catalog = 0 }
  if ($in_catalog && /^\| ([A-Z][A-Z0-9-]*) \|.*\| ([^|]*) \| ([^|]*) \|\s*$/) {
    my ($id, $class, $dependencies) = ($1, $2, $3);
    $cutover{$id} = 1 if $class =~ /CUTOVER/ && $id ne "PHASE4-CUTOVER";
    if ($id eq "PHASE4-CUTOVER") {
      while ($dependencies =~ /\b([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)+)\b/g) {
        $phase_four{$1} = 1;
      }
    }
  }

  END {
    my @missing = grep { !$phase_four{$_} } sort keys %cutover;
    my @extra = grep { !$cutover{$_} } sort keys %phase_four;
    if (@missing || @extra) {
      print STDERR "mobile revamp plan validation failed: Phase 4 CUTOVER mismatch";
      print STDERR "; missing: " . join(", ", @missing) if @missing;
      print STDERR "; extra: " . join(", ", @extra) if @extra;
      print STDERR "\n";
      exit 1;
    }
  }
' "$plan_file"

echo "mobile revamp plan validation passed: 25 priorities, 72 findings, ${#catalog_ids[@]} milestones, ${actual_edges} acyclic dependencies"
