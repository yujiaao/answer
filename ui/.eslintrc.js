module.exports = {
  root: true,
  env: {
    browser: true,
    es2021: true,
  },
  extends: [
    'react-app',
    'react-app/jest',
    'airbnb',
    'airbnb-typescript',
    'plugin:import/typescript',
    'plugin:prettier/recommended',
  ],
  overrides: [],
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaFeatures: {
      jsx: true,
    },
    ecmaVersion: 'latest',
    sourceType: 'module',
    tsconfigRootDir: __dirname,
    project: ['./tsconfig.json'],
  },
  plugins: ['react', '@typescript-eslint', 'prettier'],
  rules: {
    'prettier/prettier': 'error',
    'no-unused-vars': 'off',
    'no-console': 'off',
    'import/prefer-default-export': 'off',
    'no-param-reassign': 'off',
    'react/react-in-jsx-scope': 'off',
    'react/function-component-definition': 'off',
    'react/button-has-type': 'off',
    'react/no-unescaped-entities': 'off',
    'react/require-default-props': 'off',
    'arrow-body-style': 'off',
    "global-require": "off",
    'react/prop-types': 0,
    'react/no-danger': 'off',
    'jsx-a11y/no-static-element-interactions': 'off',
    'jsx-a11y/label-has-associated-control': 'off',
    'jsx-a11y/tabindex-no-positive': 'off',
    'func-names': 'off',
    'no-alert': 'off',
    'prefer-promise-reject-errors': 'off',
    '@typescript-eslint/naming-convention': 'off',
    'no-debugger': 'off',
    'max-len': 'off',
    'import/extensions': 'off',
    'react-hooks/exhaustive-deps': 'off',
    'react/jsx-props-no-spreading': 'off',
    '@typescript-eslint/default-param-last': 'off',
    'no-nested-ternary': 'off',
    'import/order': [
      'error',
      {
        groups: [
          'builtin',
          'external',
          ['internal', 'parent', 'sibling', 'index'],
          'unknown',
        ],
        pathGroups: [
          {
            pattern: 'react*',
            group: 'external',
            position: 'before',
          },
          {
            pattern: '@/**',
            group: 'internal',
          },
          {
            pattern: './**',
            group: 'internal',
            position: 'after',
          },
          {
            pattern: '*.scss',
            patternOptions: { matchBase: true },
            group: 'unknown',
            position: 'after',
          },
        ],
        pathGroupsExcludedImportTypes: ['react'],
        'newlines-between': 'always',
      },
    ],
  },
};
