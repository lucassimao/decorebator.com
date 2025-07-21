// https://docs.expo.dev/guides/using-eslint/
// ESLint v9 flat config format
const { FlatCompat } = require("@eslint/eslintrc");

const compat = new FlatCompat();

module.exports = [
  ...compat.extends("expo", "prettier"),
  {
    files: [
      "jest-setup.js",
      "**/__tests__/**/*",
      "**/*.test.ts",
      "**/*.test.tsx",
    ],
    languageOptions: {
      globals: {
        jest: "readonly",
        describe: "readonly",
        it: "readonly",
        expect: "readonly",
        beforeEach: "readonly",
        afterEach: "readonly",
        beforeAll: "readonly",
        afterAll: "readonly",
      },
    },
  },
  {
    plugins: {
      prettier: require("eslint-plugin-prettier"),
    },
    rules: {
      "prettier/prettier": "error",
    },
  },
];
