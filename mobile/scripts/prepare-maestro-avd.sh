#!/usr/bin/env bash
set -euo pipefail

app_id="${APP_ID:-com.lsimaocosta.decorebator}"
apk_path="${APK_PATH:-android/app/build/outputs/apk/debug/app-debug.apk}"
fixture_path="${PROFILE_IMAGE_FIXTURE:-assets/images/icon.png}"

serial="${ANDROID_SERIAL:-}"
if [ -z "$serial" ]; then
  mapfile -t emulator_serials < <(
    adb devices | awk '$2 == "device" && $1 ~ /^emulator-/ { print $1 }'
  )
  if [ "${#emulator_serials[@]}" -ne 1 ]; then
    echo "Expected exactly one online Android emulator; set ANDROID_SERIAL to select one explicitly." >&2
    exit 1
  fi
  serial="${emulator_serials[0]}"
fi
if [[ ! "$serial" =~ ^emulator-[0-9]+$ ]]; then
  echo "Refusing non-emulator Android target '$serial'." >&2
  exit 1
fi
adb -s "$serial" get-state >/dev/null

echo "Waiting for Android boot completion..."
adb -s "$serial" wait-for-device
until [ "$(adb -s "$serial" shell getprop sys.boot_completed 2>/dev/null | tr -d '\r')" = "1" ]; do
  sleep 2
done

adb -s "$serial" shell input keyevent 82 >/dev/null || true
adb -s "$serial" shell settings put global window_animation_scale 0
adb -s "$serial" shell settings put global transition_animation_scale 0
adb -s "$serial" shell settings put global animator_duration_scale 0

if [ ! -f "$apk_path" ]; then
  echo "Dev-client APK not found at '$apk_path'. Run: cd android && ./gradlew :app:assembleDebug" >&2
  exit 1
fi
if [ ! -f "$fixture_path" ]; then
  echo "Profile image fixture not found at '$fixture_path'." >&2
  exit 1
fi

adb -s "$serial" install -r "$apk_path"

# This is a dedicated test AVD: isolate its image MediaStore so the system
# Photo Picker contains exactly the fixture selected by the flow.
adb -s "$serial" shell content delete --uri content://media/external/images/media >/dev/null
adb -s "$serial" shell mkdir -p /sdcard/Pictures/Maestro
adb -s "$serial" push "$fixture_path" /sdcard/Pictures/Maestro/profile-upload.png >/dev/null
adb -s "$serial" shell am broadcast \
  -a android.intent.action.MEDIA_SCANNER_SCAN_FILE \
  -d file:///sdcard/Pictures/Maestro/profile-upload.png >/dev/null

echo "Maestro AVD '$serial' is ready for $app_id."
