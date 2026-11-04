import { readFileSync } from 'node:fs'
import { posix } from 'node:path'
import { defineConfig, type DefaultTheme } from 'vitepress'

const repository = process.env.GITHUB_REPOSITORY || 'ProgramCX/GoQQBot'
const sourceRef = process.env.DOCS_SOURCE_REF || 'main'
const github = `https://github.com/${repository}`
const basePath = (process.env.DOCS_BASE || '').replace(/^\/+|\/+$/g, '')
const base = basePath ? `/${basePath}/` : '/'

function pagePath(source: string) {
  return source.replace(/(^|\/)README\.md$/, '$1index.md')
}

function sidebar(): DefaultTheme.SidebarItem[] {
  const groups: DefaultTheme.SidebarItem[] = []
  let stack: DefaultTheme.SidebarItem[][] = []
  const contents = readFileSync(new URL('../_Sidebar.md', import.meta.url), 'utf8')

  for (const line of contents.split(/\r?\n/)) {
    const section = line.match(/^## (.+)$/)
    if (section) {
      const items: DefaultTheme.SidebarItem[] = []
      groups.push({ text: section[1], items, collapsed: groups.length !== 0 })
      stack = [items]
      continue
    }

    const entry = line.match(/^( *)- \[([^\]]+)\]\(([^)]+\.md)\)$/)
    if (!entry || stack.length === 0) continue
    const [, indent, text, target] = entry
    const depth = indent.length / 2
    if (!stack[depth]) throw new Error(`侧栏缩进不合法：${line}`)

    const items: DefaultTheme.SidebarItem[] = []
    stack[depth].push({
      text,
      link: '/' + pagePath(target).replace(/(^|\/)index\.md$/, '$1').replace(/\.md$/, '.html'),
      items,
    })
    stack.length = depth + 1
    stack.push(items)
  }

  function collapseBranches(items: DefaultTheme.SidebarItem[]) {
    for (const item of items) {
      if (item.items?.length) {
        item.collapsed ??= true
        collapseBranches(item.items)
      } else {
        delete item.items
      }
    }
  }
  collapseBranches(groups)
  return groups
}

export default defineConfig({
  lang: 'zh-CN',
  title: 'GoQQBot',
  description: 'QQ 官方机器人 Go SDK 使用文档',
  base,
  rewrites: pagePath,
  head: [['meta', { name: 'theme-color', content: '#0c7065' }]],
  markdown: {
    config(md) {
      const renderLink = md.renderer.rules.link_open!
      md.renderer.rules.link_open = (tokens, index, options, env, renderer) => {
        const token = tokens[index]
        const href = token.attrGet('href')
        if (href && !/^(?:[a-z][a-z\d+.-]*:|\/|#)/i.test(href)) {
          const [, target, suffix] = href.match(/^([^?#]+)(.*)$/) || []
          if (target) {
            const path = posix.normalize(posix.join(posix.dirname(env.relativePath || ''), target))
            if (path.startsWith('../')) {
              // 保留仓库相对链接，站点只在渲染时转换外部文件地址。
              const repoPath = posix.normalize(posix.join('wiki', path))
              const encoded = repoPath.split('/').map(part => encodeURIComponent(decodeURIComponent(part))).join('/')
              token.attrSet('href', `${github}/blob/${sourceRef}/${encoded}${suffix}`)
            } else {
              token.attrSet('href', pagePath(target) + suffix)
            }
          }
        }
        return renderLink(tokens, index, options, env, renderer)
      }
    },
  },
  transformPageData(page) {
    if (page.filePath === '_Sidebar.md') page.frontmatter.search = false
  },
  themeConfig: {
    nav: [
      { text: '快速开始', link: '/getting-started/quick-start.html' },
      { text: 'API 指南', link: '/server-api/' },
      { text: '常见问题', link: '/faq/' },
    ],
    sidebar: sidebar(),
    socialLinks: [{ icon: 'github', link: github }],
    editLink: { pattern: `${github}/edit/${sourceRef}/wiki/:path`, text: '在 GitHub 编辑此页' },
    outline: { level: [2, 3], label: '本页目录' },
    docFooter: { prev: '上一篇', next: '下一篇' },
    sidebarMenuLabel: '文档导航',
    returnToTopLabel: '返回顶部',
    darkModeSwitchLabel: '切换主题',
    lightModeSwitchTitle: '切换到浅色',
    darkModeSwitchTitle: '切换到深色',
    langMenuLabel: '语言',
    skipToContentLabel: '跳到正文',
    search: {
      provider: 'local',
      options: {
        translations: {
          button: { buttonText: '搜索文档', buttonAriaLabel: '搜索文档' },
          modal: {
            displayDetails: '显示详细结果',
            resetButtonTitle: '清空搜索',
            backButtonTitle: '返回',
            noResultsText: '没有匹配的文档',
            footer: { selectText: '选择', navigateText: '移动', closeText: '关闭' },
          },
        },
      },
    },
  },
})
