# Local production build
eas build --platform android --profile production --local

# Publish new OTA
eas update  --channel production --message "Your changes" --environment production --clear-cache --branch master