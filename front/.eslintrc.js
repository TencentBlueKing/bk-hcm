module.exports = {
  root: true,
  extends: ['@blueking/eslint-config-bk/tsvue3'],
  parserOptions: {
    project: './tsconfig.eslint.json',
    tsconfigRootDir: __dirname,
    sourceType: 'module'
  },
  rules: {
    'no-param-reassign': 0,
    'arrow-body-style': 'off',
    '@typescript-eslint/naming-convention': 0,
    'prefer-spread': 'off',
  },
};
