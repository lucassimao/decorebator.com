# Local production build
eas build --platform android --profile production --local

# Publish new OTA
eas update  --channel production --message "fix: Improve signin screen" --environment production --clear-cache

# Clear Android emulator
$ANDROID_SDK_ROOT/emulator/emulator -avd Galaxy_S25_Ultra_6_9_inch -wipe-data