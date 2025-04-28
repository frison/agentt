/** @type {import('tailwindcss').Config} */
const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  content: [
    // Point to user content mounted inside the container
    "/app/input_content/**/*.html",
    "/app/input_content/**/*.md",
    "/app/input_content/_layouts/**/*.html", // User overrides
    "/app/input_content/_posts/**/*.md",    // User posts
    "/app/input_content/_includes/**/*.html",// User includes (e.g., analytics, social)
    "/app/input_content/assets/js/**/*.js", // User JS (if any)

    // Also point to the theme files baked into the reflex image
    "/app/tailwind_build_config/*.html", // For index.html at the root
    "/app/tailwind_build_config/_layouts/**/*.html",
    "/app/tailwind_build_config/_includes/**/*.html",
    "/app/tailwind_build_config/assets/js/**/*.js",
  ],
  darkMode: 'class', // Enable dark mode using a class on the html tag
  theme: {
    extend: {
      fontFamily: {
        // Set Inter as the default sans-serif font
        sans: ['Inter', ...defaultTheme.fontFamily.sans],
        // Define a preferred monospace stack
        mono: ['Fira Code', 'Source Code Pro', 'Menlo', 'Consolas', 'Courier New', ...defaultTheme.fontFamily.mono],
      },
      colors: {
        // Define a custom palette
        'brand-bg-light': '#f8f9fa', // Very light grey
        'brand-text-light': '#212529', // Dark grey
        'brand-bg-dark': '#1a1d21',   // Very dark grey/blue
        'brand-text-dark': '#e9ecef', // Light grey
        'brand-primary': '#007bff',   // Standard blue
        'brand-primary-hover': '#0056b3',
        'brand-secondary': '#6c757d', // Muted grey
        'brand-accent': '#17a2b8',    // Teal accent
        'brand-border-light': '#dee2e6',
        'brand-border-dark': '#343a40',
        'brand-code-bg-light': '#e9ecef',
        'brand-code-bg-dark': '#2c3034',
      },
      typography: ({ theme }) => ({
        DEFAULT: {
          css: {
            '--tw-prose-body': theme('colors.brand-text-light'),
            '--tw-prose-headings': theme('colors.brand-text-light'),
            '--tw-prose-lead': theme('colors.brand-text-light'),
            '--tw-prose-links': theme('colors.brand-primary'),
            '--tw-prose-bold': theme('colors.brand-text-light'),
            '--tw-prose-counters': theme('colors.brand-secondary'),
            '--tw-prose-bullets': theme('colors.brand-secondary'),
            '--tw-prose-hr': theme('colors.brand-border-light'),
            '--tw-prose-quotes': theme('colors.brand-secondary'),
            '--tw-prose-quote-borders': theme('colors.brand-accent'),
            '--tw-prose-captions': theme('colors.brand-secondary'),
            '--tw-prose-code': theme('colors.brand-text-light'),
            '--tw-prose-pre-code': theme('colors.brand-text-light'),
            '--tw-prose-pre-bg': theme('colors.brand-code-bg-light'),
            '--tw-prose-th-borders': theme('colors.brand-border-light'),
            '--tw-prose-td-borders': theme('colors.brand-border-light'),
            'code::before': { content: 'none' },
            'code::after': { content: 'none' },
            'pre': {
              'font-family': theme('fontFamily.mono').join(', '),
              'border-radius': theme('borderRadius.md'),
              'padding': theme('spacing.4'),
            },
            'code': {
                'font-family': theme('fontFamily.mono').join(', '),
                'font-weight': 'normal',
                'padding': '0.2em 0.4em',
                'margin': '0',
                'font-size': '85%',
                'background-color': theme('colors.brand-code-bg-light'),
                'border-radius': theme('borderRadius.sm'),
            },
          },
        },
        invert: {
          css: {
            '--tw-prose-body': theme('colors.brand-text-dark'),
            '--tw-prose-headings': theme('colors.brand-text-dark'),
            '--tw-prose-lead': theme('colors.brand-text-dark'),
            '--tw-prose-links': theme('colors.brand-primary'),
            '--tw-prose-bold': theme('colors.brand-text-dark'),
            '--tw-prose-counters': theme('colors.brand-secondary'),
            '--tw-prose-bullets': theme('colors.brand-secondary'),
            '--tw-prose-hr': theme('colors.brand-border-dark'),
            '--tw-prose-quotes': theme('colors.brand-secondary'),
            '--tw-prose-quote-borders': theme('colors.brand-accent'),
            '--tw-prose-captions': theme('colors.brand-secondary'),
            '--tw-prose-code': theme('colors.brand-text-dark'),
            '--tw-prose-pre-code': theme('colors.brand-text-dark'),
            '--tw-prose-pre-bg': theme('colors.brand-code-bg-dark'),
            'code': {
                'background-color': theme('colors.brand-code-bg-dark'),
            },
          },
        },
      }),
    },
  },
  plugins: [
    require('@tailwindcss/typography'), // Add the typography plugin
  ],
}