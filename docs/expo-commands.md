# Local production build
eas build --platform android --profile production --local

# Publish new OTA
eas update  --channel production --message "fix: Improve keyboard navigation and form submission in signin screen" --environment production --clear-cache --branch master