# DAN-Government SSO — Documentation site

Technical documentation for **DAN-Government SSO** (`sso.dgov.mn`), the
eID-based, AI-enabled Single Sign-On platform. Built with
[MkDocs Material](https://squidfunk.github.io/mkdocs-material/) and published to
GitHub Pages from this folder.

**Live site:** <https://gerege-systems.github.io/sso-dgov-mn/>

## Local development

```bash
cd docs-site
python3 -m venv .venv
./.venv/bin/pip install -r requirements.txt
./.venv/bin/mkdocs serve      # http://127.0.0.1:8000
```

Build (strict — fails on broken links):

```bash
./.venv/bin/mkdocs build --strict
```

## Structure

- `mkdocs.yml` — site config, theme, nav, plugins (search, i18n EN/MN, swagger).
- `docs/**/*.md` — English pages (default language).
- `docs/**/*.mn.md` — Mongolian translations (fall back to English when absent).
- `docs/assets/` — logo, brand CSS, and the OpenAPI `swagger.json`.
- `../.github/workflows/docs.yml` — builds and deploys to GitHub Pages when
  `docs-site/**` changes on `main`.

## Languages

The site is bilingual (English + Монгол) via `mkdocs-static-i18n`. English is the
default; Mongolian pages use the `.mn.md` suffix and fall back to English where a
translation is not yet present.

## Source

This folder was previously the standalone
`gerege-systems/sso-dgov-mn-documentation` repository; it now lives alongside the
code it documents. Deep-dive engineering docs stay in [`../docs/`](../docs/) and
`../backend/docs/` — this site is the published, reader-facing view.
Co-developed by the Gerege Systems Development Team and Claude AI, 2026.
MIT-licensed.
