# Dlyven website

Static GitHub Pages site. No npm dependencies.

```bash
PAGES_BASE=/dlyven/ PAGES_REPO=dlyven PAGES_OWNER=amirreza-azadmehr node site/build.mjs
```

Output: `site/dist/` (gitignored). The Pages workflow deploys it.

Fonts are self-hosted SIL OFL files under `site/static/fonts/`. No analytics, no CDNs.
