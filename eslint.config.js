const js = require('@eslint/js');
const globals = require('globals');
const prettier = require('eslint-config-prettier');

module.exports = [
    {
        ignores: [
            'node_modules/**',
            'offline-packages/**',
            'assets/**',
            'jobs/**',
            'Downloads/**',
            'data/**',
        ],
    },

    // Base JS rules
    {
        files: ['**/*.js'],
        languageOptions: {
            ecmaVersion: 'latest',
        },
        rules: {
            ...js.configs.recommended.rules,

            // This repo is currently not strict about unused vars; keep lint useful without being noisy.
            'no-unused-vars': [
                'warn',
                { args: 'none', caughtErrors: 'none', varsIgnorePattern: '^_' },
            ],

            // We intentionally strip/control ANSI/control characters in some regexes.
            'no-control-regex': 'off',

            // This codebase uses empty catches for best-effort cleanup.
            'no-empty': ['error', { allowEmptyCatch: true }],

            // Avoid noise from regex/string escaping patterns used for CLI/output cleanup.
            'no-useless-escape': 'off',
        },
    },

    // Node/CommonJS backend
    {
        files: ['server.js', 'utils/**/*.js', 'eslint.config.js'],
        languageOptions: {
            sourceType: 'commonjs',
            globals: {
                ...globals.node,
                ...globals.es2021,

                // Node 18+ web globals used in `server.js`
                fetch: 'readonly',
                AbortSignal: 'readonly',
                TextDecoder: 'readonly',
                TextEncoder: 'readonly',
                URL: 'readonly',
            },
        },
    },

    // Browser frontend
    {
        files: ['gui/**/*.js'],
        languageOptions: {
            sourceType: 'script',
            globals: {
                ...globals.browser,
                ...globals.es2021,
            },
        },
    },

    // Disable rules that conflict with Prettier formatting.
    prettier,
];
