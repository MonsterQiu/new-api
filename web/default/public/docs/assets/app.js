/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.
*/

const root = document.documentElement
const body = document.body
const menuButton = document.querySelector('#menu-button')
const menuClose = document.querySelector('#menu-close')
const navScrim = document.querySelector('#nav-scrim')
const themeButton = document.querySelector('#theme-button')
const searchInput = document.querySelector('#docs-search')
const clearSearch = document.querySelector('#clear-search')
const searchStatus = document.querySelector('#search-status')
const noResults = document.querySelector('#no-results')
const progressBar = document.querySelector('#reading-progress-bar')
const sections = [...document.querySelectorAll('[data-search-section]')]
const tocLinks = [...document.querySelectorAll('[data-section]')]

function readThemeCookie() {
  const match = document.cookie.match(/(?:^|; )vite-ui-theme=([^;]*)/)
  return match ? decodeURIComponent(match[1]) : 'system'
}

function getResolvedTheme(theme) {
  if (theme === 'light' || theme === 'dark') return theme
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light'
}

function applyTheme(theme) {
  const resolvedTheme = getResolvedTheme(theme)
  root.dataset.theme = resolvedTheme
  themeButton.setAttribute(
    'aria-label',
    resolvedTheme === 'dark' ? '切换浅色模式' : '切换深色模式'
  )
}

let currentTheme = readThemeCookie()
applyTheme(currentTheme)

themeButton.addEventListener('click', () => {
  const resolvedTheme = getResolvedTheme(currentTheme)
  currentTheme = resolvedTheme === 'dark' ? 'light' : 'dark'
  document.cookie = `vite-ui-theme=${currentTheme}; path=/; max-age=31536000; SameSite=Lax`
  applyTheme(currentTheme)
})

window
  .matchMedia('(prefers-color-scheme: dark)')
  .addEventListener('change', () => {
    if (currentTheme === 'system') applyTheme(currentTheme)
  })

function setNavigationOpen(open) {
  body.classList.toggle('nav-open', open)
  menuButton.setAttribute('aria-expanded', String(open))
  navScrim.hidden = !open
  if (open) searchInput.focus()
}

menuButton.addEventListener('click', () => setNavigationOpen(true))
menuClose.addEventListener('click', () => setNavigationOpen(false))
navScrim.addEventListener('click', () => setNavigationOpen(false))

document.addEventListener('keydown', (event) => {
  if (event.key === 'Escape') setNavigationOpen(false)
})

tocLinks.forEach((link) => {
  link.addEventListener('click', () => setNavigationOpen(false))
})

function filterSections() {
  const query = searchInput.value.trim().toLocaleLowerCase('zh-CN')
  let matches = 0

  sections.forEach((section) => {
    const text = section.textContent.toLocaleLowerCase('zh-CN')
    const matched = !query || text.includes(query)
    section.hidden = !matched
    const link = tocLinks.find((item) => item.dataset.section === section.id)
    if (link) link.hidden = !matched
    if (matched) matches += 1
  })

  clearSearch.hidden = !query
  noResults.hidden = matches !== 0
  searchStatus.textContent = query
    ? matches > 0
      ? `找到 ${matches} 个相关章节`
      : '没有匹配结果'
    : ''
}

searchInput.addEventListener('input', filterSections)
clearSearch.addEventListener('click', () => {
  searchInput.value = ''
  filterSections()
  searchInput.focus()
})

let scrollTicking = false

function updateReadingProgress() {
  const scrollable = document.documentElement.scrollHeight - window.innerHeight
  const progress = scrollable > 0 ? window.scrollY / scrollable : 0
  progressBar.style.transform = `scaleX(${Math.min(1, Math.max(0, progress))})`
  scrollTicking = false
}

window.addEventListener(
  'scroll',
  () => {
    if (scrollTicking) return
    scrollTicking = true
    window.requestAnimationFrame(updateReadingProgress)
  },
  { passive: true }
)
updateReadingProgress()

const observer = new IntersectionObserver(
  (entries) => {
    const visible = entries
      .filter((entry) => entry.isIntersecting)
      .sort(
        (left, right) =>
          left.boundingClientRect.top - right.boundingClientRect.top
      )

    if (visible.length === 0) return
    const activeId = visible[0].target.id
    tocLinks.forEach((link) => {
      const active = link.dataset.section === activeId
      link.classList.toggle('active', active)
      if (active) link.setAttribute('aria-current', 'location')
      else link.removeAttribute('aria-current')
    })
  },
  { rootMargin: '-80px 0px -68% 0px', threshold: 0.01 }
)

sections.forEach((section) => observer.observe(section))

function normalizeServerAddress(value) {
  const fallback = window.location.origin
  if (typeof value !== 'string' || !value.trim()) return fallback
  return value.trim().replace(/\/+$/, '')
}

function setDynamicContent(status) {
  const systemName = status.system_name || 'API 服务'
  const apiRoot = normalizeServerAddress(status.server_address)
  const apiBase = `${apiRoot}/v1`

  document.querySelector('#brand-name').textContent = systemName
  document.querySelectorAll('[data-brand-name]').forEach((element) => {
    element.textContent = systemName
  })
  document.querySelectorAll('[data-api-root]').forEach((element) => {
    element.textContent = apiRoot
  })
  document.querySelectorAll('[data-api-base]').forEach((element) => {
    element.textContent = apiBase
  })

  if (status.logo) {
    document.querySelector('#brand-logo').src = status.logo
  }

  document.title = `${systemName} API 中转使用说明`

  const samples = {
    codex: `model = "your-model-id"\nmodel_provider = "gateway"\n\n[model_providers.gateway]\nname = "${systemName}"\nbase_url = "${apiBase}"\nenv_key = "GATEWAY_API_KEY"\nwire_api = "responses"`,
    curl: `curl "${apiBase}/chat/completions" \\\n+  -H "Authorization: Bearer sk-你的密钥" \\\n+  -H "Content-Type: application/json" \\\n+  -d '{\n    "model": "your-model-id",\n    "messages": [\n      {"role": "user", "content": "你好"}\n    ]\n  }'`,
    python: `from openai import OpenAI\n\nclient = OpenAI(\n    api_key="sk-你的密钥",\n    base_url="${apiBase}",\n)\n\nresponse = client.chat.completions.create(\n    model="your-model-id",\n    messages=[{"role": "user", "content": "你好"}],\n)\nprint(response.choices[0].message.content)`,
    node: `import OpenAI from "openai";\n\nconst client = new OpenAI({\n  apiKey: "sk-你的密钥",\n  baseURL: "${apiBase}",\n});\n\nconst response = await client.chat.completions.create({\n  model: "your-model-id",\n  messages: [{ role: "user", content: "你好" }],\n});\nconsole.log(response.choices[0].message.content);`,
  }

  document.querySelectorAll('[data-code]').forEach((element) => {
    element.textContent = samples[element.dataset.code] || ''
  })

  document.querySelectorAll('[data-copy-api-base]').forEach((button) => {
    button.dataset.copyValue = apiBase
  })
}

async function loadStatus() {
  try {
    const response = await fetch('/api/status', {
      credentials: 'same-origin',
      headers: { Accept: 'application/json' },
    })
    if (!response.ok) throw new Error('Status request failed')
    const payload = await response.json()
    setDynamicContent(payload.data || {})
  } catch {
    setDynamicContent({ server_address: window.location.origin })
  }
}

async function copyText(value, button) {
  try {
    await navigator.clipboard.writeText(value)
    const original = button.textContent
    button.textContent = '已复制'
    window.setTimeout(() => {
      button.textContent = original
    }, 1400)
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = value
    textarea.setAttribute('readonly', '')
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.append(textarea)
    textarea.select()
    document.execCommand('copy')
    textarea.remove()
  }
}

document.addEventListener('click', (event) => {
  const target = event.target
  if (!(target instanceof Element)) return

  const valueButton = target.closest('[data-copy-value]')
  if (valueButton) {
    copyText(valueButton.dataset.copyValue, valueButton)
    return
  }

  const codeButton = target.closest('[data-copy-code]')
  if (codeButton) {
    const code = codeButton.closest('.code-block').querySelector('code')
    copyText(code.textContent, codeButton)
  }
})

document.querySelectorAll('[data-image-slot]').forEach((figure) => {
  const image = figure.querySelector('img')
  const showImage = () => figure.classList.add('is-loaded')
  if (image.complete && image.naturalWidth > 0) showImage()
  else image.addEventListener('load', showImage, { once: true })
})

loadStatus()
