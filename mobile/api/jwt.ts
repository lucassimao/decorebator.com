function base64UrlDecode(str: string) {
  // Replace non-url compatible chars with base64 standard chars
  str = str.replace(/-/g, "+").replace(/_/g, "/");

  // Pad the base64 string to make its length a multiple of 4
  while (str.length % 4) {
    str += "=";
  }

  // Decode base64 string
  return atob(str);
}

export function decode(token: string) {
  // Split the token into its parts
  const parts = token.split(".");

  if (parts.length !== 3) {
    throw new Error("Invalid JWT token");
  }

  // Decode the header and payload
  const header = JSON.parse(base64UrlDecode(parts[0]));
  const payload = JSON.parse(base64UrlDecode(parts[1]));
  
  // Return the decoded header and payload
  return {
    header,
    payload,
  };
}
