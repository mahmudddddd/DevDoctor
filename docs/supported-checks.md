# Supported checks

This document describes behavior that is backed by fixtures and tests. It is intentionally narrower than DevDoctor's long-term roadmap.

## Phase 1: safe project discovery

### Languages and runtimes

- JavaScript: inferred from `package.json`, JavaScript config files, or JavaScript source markers at the project root.
- TypeScript: inferred from `tsconfig.json`, TypeScript source markers at the project root, or a TypeScript dependency.
- Node.js: inferred from `package.json`; declared engine requirements are reported when present.

### Package managers

- npm: `package-lock.json`, `npm-shrinkwrap.json`, or `packageManager` metadata.
- pnpm: `pnpm-lock.yaml`, `pnpm-workspace.yaml`, or `packageManager` metadata.
- Yarn: `yarn.lock` or `packageManager` metadata.
- Bun: `bun.lock`, `bun.lockb`, or `packageManager` metadata.

Conflicting lockfiles are reported as warnings rather than resolved by guessing.

### Framework markers

Dependencies are used to identify React, Next.js, Vue, Nuxt, Angular, Svelte, SvelteKit, Vite, Remix, Astro, NestJS, Express, Fastify, and Electron. Recognized configuration filenames provide additional evidence.

### Workspaces

- npm/Yarn-compatible `workspaces` fields in `package.json`
- `pnpm-workspace.yaml`

Workspace package globs are reported but not expanded in Phase 1.

### Safety

Phase 1 does not execute package scripts, runtime probes, network requests, Docker commands, or database checks. Those require later command policy and consent work.
