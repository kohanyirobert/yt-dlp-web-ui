# Plan: Remove Reverse Proxy Support

## Context
Slim down the yt-dlp web UI by removing reverse proxy support. The app is accessed directly, so the BaseURL/sub-directory configuration and reverse proxy toggle in settings are unused.

---

## Frontend

- **`frontend/src/atoms/settings.ts`**:
  - Remove `servedFromReverseProxy` from `SettingsState` interface (line 47)
  - Remove `servedFromReverseProxyState` atom (lines 106-109)
  - Remove `servedFromReverseProxySubDirState` atom (lines 111-114)
  - Simplify `serverAddressAndPortState` (lines 121-136): remove the two reverse proxy conditional branches, keep only the `addr:port` logic
  - Remove `servedFromReverseProxy` from `settingsState` (line 190)
- **`frontend/src/views/Settings.tsx`**:
  - Remove imports for `servedFromReverseProxyState`, `servedFromReverseProxySubDirState` (lines 44-45)
  - Remove `reverseProxy`/`baseURL` state hooks (lines 59-60)
  - Remove `baseURL$` subject (line 84) and its `useEffect` (lines 91-99)
  - Remove `disabled={reverseProxy}` from port field (line 184)
  - Remove entire "Reverse Proxy" UI section (lines 229-259)
- **15 i18n YAML files**: Remove `servedFromReverseProxyCheckbox` and `urlBase` keys

## Backend

- **`server/config/config.go`**: Remove `BaseURL` field (line 15)
- **`server/server.go`**: Replace lines 198-199 with simple `r.Handle("/*", http.FileServerFS(c.frontend))`

---

## Verification

1. **Build backend**: `go build ./...` — ensure no compilation errors
2. **Run backend tests**: `go test ./...`
3. **Build frontend**: `cd frontend && pnpm build` — ensure no TypeScript errors
4. **Manual smoke test**: Run the app and verify Settings has no reverse proxy section
