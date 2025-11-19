import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "BobaMixer",
  description: "Smart AI Adapter Router with intelligent routing, budget tracking, and cost optimization",
  base: '/BobaMixer/',
  markdown: {
    mermaid: true
  },
  locales: {
    root: {
      label: 'English',
      lang: 'en',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/' },
          { text: 'Guide', link: '/guide/getting-started' },
          { text: 'Features', link: '/features/adapters' },
          { text: 'Reference', link: '/reference/cli' }
        ],
        sidebar: [
          {
            text: 'Guide',
            collapsed: false,
            items: [
              { text: 'Getting Started', link: '/guide/getting-started' },
              { text: 'Installation', link: '/guide/installation' },
              { text: 'Configuration', link: '/guide/configuration' }
            ]
          },
          {
            text: 'Features',
            collapsed: false,
            items: [
              { text: 'Adapters', link: '/features/adapters' },
              { text: 'Intelligent Routing', link: '/features/routing' },
              { text: 'Budget Management', link: '/features/budgets' },
              { text: 'Analytics & Stats', link: '/features/analytics' }
            ]
          },
          {
            text: 'Reference',
            collapsed: false,
            items: [
              { text: 'CLI Commands', link: '/reference/cli' },
              { text: 'Configuration Files', link: '/reference/config-files' }
            ]
          },
          {
            text: 'Advanced',
            collapsed: false,
            items: [
              { text: 'Operations', link: '/advanced/operations' },
              { text: 'Troubleshooting', link: '/advanced/troubleshooting' }
            ]
          }
        ],
        editLink: {
          pattern: 'https://github.com/royisme/BobaMixer/edit/main/docs/:path',
          text: 'Edit this page on GitHub'
        }
      }
    },
    zh: {
      label: '简体中文',
      lang: 'zh-CN',
      link: '/zh/',
      themeConfig: {
        nav: [
          { text: '首页', link: '/zh/' },
          { text: '指南', link: '/zh/guide/getting-started' },
          { text: '功能', link: '/zh/features/adapters' },
          { text: '参考', link: '/zh/reference/cli' }
        ],
        sidebar: [
          {
            text: '指南',
            collapsed: false,
            items: [
              { text: '快速开始', link: '/zh/guide/getting-started' },
              { text: '安装', link: '/zh/guide/installation' },
              { text: '配置', link: '/zh/guide/configuration' }
            ]
          },
          {
            text: '功能',
            collapsed: false,
            items: [
              { text: '适配器', link: '/zh/features/adapters' },
              { text: '智能路由', link: '/zh/features/routing' },
              { text: '预算管理', link: '/zh/features/budgets' },
              { text: '分析统计', link: '/zh/features/analytics' }
            ]
          },
          {
            text: '参考',
            collapsed: false,
            items: [
              { text: 'CLI 命令', link: '/zh/reference/cli' },
              { text: '配置文件', link: '/zh/reference/config-files' }
            ]
          },
          {
            text: '高级',
            collapsed: false,
            items: [
              { text: '运维操作', link: '/zh/advanced/operations' },
              { text: '故障排除', link: '/zh/advanced/troubleshooting' }
            ]
          }
        ],
        editLink: {
          pattern: 'https://github.com/royisme/BobaMixer/edit/main/docs/:path',
          text: '在 GitHub 上编辑此页'
        },
        docFooter: {
          prev: '上一页',
          next: '下一页'
        },
        outline: {
          label: '页面导航'
        },
        returnToTopLabel: '回到顶部',
        sidebarMenuLabel: '菜单',
        darkModeSwitchLabel: '主题'
      }
    }
  },

  themeConfig: {
    logo: '/logo.svg',
    socialLinks: [
      { icon: 'github', link: 'https://github.com/royisme/BobaMixer' }
    ],
    search: {
      provider: 'local'
    }
  },

  lastUpdated: true,
  cleanUrls: true,

  head: [
    ['link', { rel: 'icon', href: '/BobaMixer/favicon.ico' }],
    ['meta', { name: 'theme-color', content: '#06b6d4' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'BobaMixer: Smart AI Adapter Router' }],
    ['meta', { property: 'og:description', content: 'Smart AI Adapter Router with intelligent routing, budget tracking, and cost optimization' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: 'BobaMixer: Smart AI Adapter Router' }],
    ['meta', { name: 'twitter:description', content: 'Smart AI Adapter Router with intelligent routing, budget tracking, and cost optimization' }]
  ]
})
