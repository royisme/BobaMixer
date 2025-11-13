# BobaMixer Documentation Site

This directory contains the source code for the BobaMixer documentation website, built with [Hugo](https://gohugo.io/) and the [Docsy](https://www.docsy.dev/) theme.

## Prerequisites

- [Hugo Extended](https://gohugo.io/installation/) version 0.128.0 or later
- [Go](https://go.dev/doc/install) 1.22 or later
- [Node.js](https://nodejs.org/) 20.x or later

## Local Development

### Install Dependencies

```bash
# Install Node.js dependencies
npm install

# Install Hugo modules
hugo mod get
```

### Run Development Server

```bash
# Start the Hugo development server
hugo server -D

# Or use npm script
npm run serve
```

The site will be available at `http://localhost:1313/BobaMixer/`

### Build for Production

```bash
# Build the site
hugo --gc --minify

# Or use npm script
npm run build
```

The built site will be in the `public/` directory.

## Content Structure

```
content/
├── en/              # English content
│   ├── docs/        # Documentation
│   │   ├── getting-started/
│   │   ├── user-guide/
│   │   ├── configuration/
│   │   ├── adapters/
│   │   ├── routing/
│   │   ├── troubleshooting/
│   │   └── development/
│   └── blog/        # Blog posts
└── zh/              # Chinese content (mirrors en structure)
    ├── docs/
    └── blog/
```

## Adding Content

### Create a New Documentation Page

English:
```bash
hugo new content/en/docs/section-name/page-name.md
```

Chinese:
```bash
hugo new content/zh/docs/section-name/page-name.md
```

### Front Matter Example

```yaml
---
title: "Page Title"
linkTitle: "Short Title"
weight: 1
description: >
  Brief description of the page content.
---

Your markdown content here...
```

## Theme Customization

- **Layouts**: Override Docsy layouts in `layouts/`
- **Static assets**: Add to `static/`
- **i18n strings**: Add to `i18n/`

## Deployment

The documentation site is automatically deployed to GitHub Pages when changes are pushed to the `main` branch via GitHub Actions (see `.github/workflows/docs.yml`).

## Multilingual Support

The site supports English and Chinese. To add content in both languages:

1. Create the English version in `content/en/`
2. Create the Chinese version in `content/zh/` with the same path
3. Hugo will automatically link the translations

## Resources

- [Hugo Documentation](https://gohugo.io/documentation/)
- [Docsy Theme Docs](https://www.docsy.dev/docs/)
- [Hugo Multilingual Mode](https://gohugo.io/content-management/multilingual/)
