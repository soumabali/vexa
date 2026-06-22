module.exports = {
    root: true,
    parser: '@typescript-eslint/parser',
    parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
            jsx: true,
        },
        project: './tsconfig.json',
    },
    plugins: [
        'security',
        '@typescript-eslint',
    ],
    extends: [
        'plugin:security/recommended-legacy',
    ],
    env: {
        browser: true,
        node: true,
        es2024: true,
    },
    rules: {
        // Security rules
        'security/detect-object-injection': 'error',
        'security/detect-non-literal-regexp': 'error',
        'security/detect-unsafe-regex': 'error',
        'security/detect-buffer-noassert': 'error',
        'security/detect-eval-with-expression': 'error',
        'security/detect-non-literal-require': 'error',
        'security/detect-non-literal-fs-filename': 'error',
        'security/detect-child-process': 'warn',
        'security/detect-new-buffer': 'error',
        'security/detect-no-csrf-before-method-override': 'error',
        'security/detect-possible-timing-attacks': 'warn',
        'security/detect-pseudoRandomBytes': 'error',
    },
    overrides: [
        {
            files: ['*.ts', '*.tsx'],
            parser: '@typescript-eslint/parser',
            plugins: ['@typescript-eslint'],
            extends: [
                'plugin:@typescript-eslint/recommended',
                'plugin:@typescript-eslint/recommended-requiring-type-checking',
                'plugin:security/recommended-legacy',
            ],
        },
        {
            files: ['*.test.*', '*.spec.*', '**/__tests__/**'],
            rules: {
                'security/detect-non-literal-fs-filename': 'off',
            },
        },
    ],
    ignorePatterns: [
        'dist/**',
        'build/**',
        'node_modules/**',
        'coverage/**',
        '*.config.js',
        '*.config.ts',
        '.eslintrc.*',
    ],
};
