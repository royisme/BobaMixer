import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "BobaMixer",
  description: "Smart AI Adapter Router with intelligent routing, budget tracking, and cost optimization",
  base: '/BobaMixer/',

  locales: {
    root: {
      label: 'English',
      lang: 'en',
      themeConfig: {
        nav: [
          { text: 'Home', link: '/' },
          { text: 'Docs', link: '/en/getting-started' },
          { text: 'GitHub', link: 'https://github.com/royisme/BobaMixer' }
        ],
        sidebar: {
          '/en/': [
            {
              text: 'Getting Started',
              items: [
                { text: 'Quick Start', link: '/en/getting-started' },
                { text: 'Installation', link: '/QUICKSTART' }
              ]
            },
            {
              text: 'Documentation',
              items: [
                { text: 'Configuration', link: '/en/configuration' },
                { text: 'Adapters', link: '/ADAPTERS' },
                { text: 'Routing Cookbook', link: '/ROUTING_COOKBOOK' },
                { text: 'Operations', link: '/OPERATIONS' },
                { text: 'FAQ', link: '/FAQ' },
                { text: 'Quick Reference', link: '/QUICK_REFERENCE' }
              ]
            }
          ]
        },
        editLink: {
          pattern: 'https://github.com/royisme/BobaMixer/edit/main/docs/:path',
          text: 'Edit this page on GitHub'
        },
        lastUpdated: {
          text: 'Last updated',
          formatOptions: {
            dateStyle: 'short',
            timeStyle: 'short'
          }
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
          { text: '文档', link: '/zh/getting-started' },
          { text: 'GitHub', link: 'https://github.com/royisme/BobaMixer' }
        ],
        sidebar: {
          '/zh/': [
            {
              text: '快速开始',
              items: [
                { text: '快速上手', link: '/zh/getting-started' },
                { text: '安装指南', link: '/QUICKSTART' }
              ]
            },
            {
              text: '文档',
              items: [
                { text: '配置指南', link: '/zh/configuration' },
                { text: '适配器', link: '/ADAPTERS' },
                { text: '路由手册', link: '/ROUTING_COOKBOOK' },
                { text: '运维操作', link: '/OPERATIONS' },
                { text: '常见问题', link: '/FAQ' },
                { text: '快速参考', link: '/QUICK_REFERENCE' }
              ]
            }
          ]
        },
        editLink: {
          pattern: 'https://github.com/royisme/BobaMixer/edit/main/docs/:path',
          text: '在 GitHub 上编辑此页'
        },
        lastUpdated: {
          text: '最后更新',
          formatOptions: {
            dateStyle: 'short',
            timeStyle: 'short'
          }
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
        darkModeSwitchLabel: '主题',
        lightModeSwitchTitle: '切换到浅色模式',
        darkModeSwitchTitle: '切换到深色模式'
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
  ignoreDeadLinks: true,

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
