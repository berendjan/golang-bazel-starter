# React Frontend

React + TypeScript application built with Vite and Bazel.

## Quick Start

### Development Server

Run the dev server with hot reload:

```bash
# From project root
ibazel run //frontend/react:start
```

The app will be available at **http://localhost:5173**

Hot reload is enabled - edit source files and the browser will automatically refresh.

> **Note:** Use `ibazel` (not `bazel`) for auto-refresh on file changes. Install with: `brew install ibazel` or `npm install -g @bazel/ibazel`

### Production Build

Build optimized production bundle:

```bash
bazel build //frontend/react:build
```

Output will be in `bazel-bin/frontend/react/dist/`

### Preview Production Build

Test the production build locally:

```bash
bazel run //frontend/react:preview
```

### Run Tests

```bash
bazel test //frontend/react/src:test
```

## Architecture

- **Build Tool**: Vite (fast dev server and bundler)
- **Transpiler**: SWC (fast TypeScript/JSX compilation)
- **Type Checking**: TypeScript (via `ts_project` in Bazel)
- **Testing**: Vitest + Testing Library
- **Package Manager**: pnpm (via Bazel's npm_translate_lock)

## API Integration

The dev server proxies `/api` requests to the backend at `http://localhost:26000` (configured in `vite.config.js`).

Example:
```typescript
// Calls http://localhost:26000/api/users in dev
fetch('/api/users')
```

## File Structure

```
frontend/react/
├── src/
│   ├── App.tsx          # Main app component
│   ├── index.tsx        # Entry point
│   └── BUILD.bazel      # Bazel build config
├── public/              # Static assets
├── vite.config.js       # Vite configuration
├── tsconfig.json        # TypeScript config
└── BUILD.bazel          # Bazel targets
```

## Bazel Targets

- `//frontend/react:start` - Dev server
- `//frontend/react:build` - Production build
- `//frontend/react:preview` - Preview production build
- `//frontend/react/src:test` - Run tests
- `//frontend/react/src:src` - TypeScript compilation

## Managing Dependencies

Add new npm packages:

```bash
# 1. Add to package.json
cd frontend/react
vim package.json  # Add your dependency

# 2. Update lockfile
cd ../  # go to frontend/ dir
./tools/pnpm install

# 3. Bazel will automatically pick up the new dependency
```

## Common Issues

**Hot reload not working?**
- Make sure you're using `ibazel run` (not `bazel run`)

**Build fails with module not found?**
- Run `./tools/pnpm install` from `frontend/` directory
- Dependencies must be in `frontend/react/package.json`

**TypeScript errors?**
- Check `tsconfig.json` configuration
- Ensure types are installed: `@types/react`, `@types/react-dom`
