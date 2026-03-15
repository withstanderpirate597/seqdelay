// @ts-check
const { themes } = require('prism-react-renderer');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'seqdelay',
  tagline: 'High-performance delay queue powered by seqflow + Redis',
  favicon: 'img/favicon.ico',
  url: 'https://seqdelay.pages.dev',
  baseUrl: '/',
  organizationName: 'gocronx',
  projectName: 'seqdelay',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'zh-Hans'],
    localeConfigs: {
      en: { label: 'English' },
      'zh-Hans': { label: '中文' },
    },
  },

  markdown: { mermaid: true },
  themes: ['@docusaurus/theme-mermaid'],

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: { sidebarPath: './sidebars.js', routeBasePath: 'docs' },
        blog: false,
        theme: { customCss: './src/css/custom.css' },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      colorMode: { defaultMode: 'light', respectPrefersColorScheme: true },
      mermaid: { theme: { light: 'neutral', dark: 'dark' } },
      navbar: {
        title: 'seqdelay',
        style: 'dark',
        items: [
          { type: 'docSidebar', sidebarId: 'docs', position: 'left', label: 'Docs' },
          { type: 'localeDropdown', position: 'right' },
          { href: 'https://github.com/gocronx/seqdelay', label: 'GitHub', position: 'right' },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              { label: 'Getting Started', to: '/docs/getting-started' },
              { label: 'API Reference', to: '/docs/api' },
              { label: 'HTTP API', to: '/docs/http-api' },
            ],
          },
          {
            title: 'Ecosystem',
            items: [
              { label: 'seqflow', href: 'https://github.com/gocronx/seqflow' },
            ],
          },
          {
            title: 'Community',
            items: [
              { label: 'GitHub Issues', href: 'https://github.com/gocronx/seqdelay/issues' },
            ],
          },
        ],
        copyright: `Copyright ${new Date().getFullYear()} gocronx.`,
      },
      prism: {
        theme: themes.github,
        darkTheme: themes.dracula,
        additionalLanguages: ['go', 'bash', 'python'],
      },
    }),
};

module.exports = config;
