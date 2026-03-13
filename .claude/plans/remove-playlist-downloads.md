# Plan: Remove Playlist Downloads

## Context
Slim down the yt-dlp web UI by removing the Playlist Downloads feature. Users only need single-video downloads; the playlist UI, RPC methods, and backend logic are unused.

---

## Frontend

- **`frontend/src/components/DownloadDialog.tsx`**: Remove `isPlaylist` state, the playlist checkbox UI (~line 356-358), and the playlist warning in format selection (~line 145-150)
- **`frontend/src/lib/rpcClient.ts`**: Remove `playlist?: boolean` from `DownloadRequestArgs` (line 11) and the `if (req.playlist)` branch routing to `Service.ExecPlaylist` (lines 92-101)
- **`frontend/src/types/index.ts`**: Remove `"Service.ExecPlaylist"` from RPCMethods (line 9)
- **15 i18n YAML files** in `frontend/src/assets/i18n/`: Remove `playlistCheckbox` key from each

## Backend

- **Delete** entire `server/playlist/` package (2 files: `types.go`, `modifiers.go`)
- **Delete** `server/internal/playlist.go` (the `PlaylistDetect()` function)
- **`server/rpc/service.go`**: Remove `ExecPlaylist()` method (lines 44-54); remove playlist auto-detect in `Formats()` (lines 108-110)
- **`server/rest/service.go`**: Remove `ExecPlaylist()` method (lines 41-43)
- **`server/rest/handlers.go`**: Remove `ExecPlaylist()` handler (lines 44-67)
- **`server/rest/container.go`**: Remove `r.Post("/execPlaylist", ...)` route (line 29)
- **`server/internal/process.go`**: Remove `strings.Split(p.Url, "?list")[0]` and `"--no-playlist"` arg (lines 101, 104) — these prevent playlist behavior on single-video downloads, so they stay only if needed for correctness. Review and decide.
- **`proto/yt-dlp.proto`**: Remove `ExecPlaylist` RPC definition (line 58)

## Note on `--no-playlist`
The `--no-playlist` flag in `process.go` is a safety measure to ensure single-video downloads don't accidentally pull entire playlists. **Keep this** — it's not playlist feature code, it's defensive behavior for single downloads.

---

## Verification

1. **Build backend**: `go build ./...` — ensure no compilation errors
2. **Run backend tests**: `go test ./...`
3. **Build frontend**: `cd frontend && pnpm build` — ensure no TypeScript errors
4. **Manual smoke test**: Run the app and verify the download dialog has no playlist checkbox
