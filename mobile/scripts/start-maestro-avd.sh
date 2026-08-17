#!/usr/bin/env bash
set -euo pipefail

android_home="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-$HOME/ProgrammingTools/Android}}"
avd_name="${AVD_NAME:-Large_Phone_414x896}"
gpu_mode="${GPU_MODE:-swiftshader_indirect}"
emulator_bin="${EMULATOR_BIN:-$android_home/emulator/emulator}"

if ! "$emulator_bin" -list-avds | grep -Fxq "$avd_name"; then
  echo "Android AVD '$avd_name' is not installed; follow .maestro/README.md." >&2
  exit 1
fi

exec "$emulator_bin" "@$avd_name" \
  -gpu "$gpu_mode" \
  -no-boot-anim \
  -no-audio \
  -netfast \
  -no-snapshot-load \
  -no-snapshot-save
