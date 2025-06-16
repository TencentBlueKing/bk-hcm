module.exports = {
  extends: ['@blueking/eslint-config-bk/tsvue3', 'plugin:prettier/recommended'],
  rules: {
    'no-param-reassign': 0,
    'arrow-body-style': 'off',
    '@typescript-eslint/naming-convention': 0,
    '@typescript-eslint/no-misused-promises': 0,
    '@typescript-eslint/no-require-imports': 0,
    'prefer-spread': 'off',
    'no-console': ['error', { allow: ['warn', 'error'] }],
    'no-debugger': 'error',
    'linebreak-style': 0,
    'vue/require-explicit-emits': 0,
    'vue/multi-word-component-names': 0,
    'vue/component-definition-name-casing': 0,
  },
};
