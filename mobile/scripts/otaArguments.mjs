const profiles = {
  development: [
    "update",
    "--environment",
    "development",
    "--channel",
    "development",
    "--clear-cache",
  ],
  preview: [
    "update",
    "--environment",
    "preview",
    "--channel",
    "preview",
    "--clear-cache",
  ],
  production: [
    "update",
    "--environment",
    "production",
    "--branch",
    "production",
    "--clear-cache",
  ],
};

export function buildOTAArguments(
  target,
  extraArguments = [],
  environment = {},
) {
  if (!target || !profiles[target]) {
    throw new Error("target must be development, preview, or production");
  }
  if (extraArguments.length > 0) {
    if (
      extraArguments.length !== 2 ||
      extraArguments[0] !== "--message" ||
      !validMessage(extraArguments[1])
    ) {
      throw new Error("only --message <text> may follow the OTA target");
    }
    if (environment.OTA_MESSAGE) {
      throw new Error("use either --message or OTA_MESSAGE, not both");
    }
  }
  const argumentsForEAS = [...profiles[target], ...extraArguments];
  if (environment.OTA_MESSAGE) {
    if (!validMessage(environment.OTA_MESSAGE)) {
      throw new Error("OTA_MESSAGE must contain 1 to 1024 characters");
    }
    argumentsForEAS.push("--message", environment.OTA_MESSAGE);
  }
  return argumentsForEAS;
}

function validMessage(value) {
  return (
    typeof value === "string" && value.trim().length > 0 && value.length <= 1024
  );
}
