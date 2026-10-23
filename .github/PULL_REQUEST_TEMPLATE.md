## Summary

- 

## Test plan

- [ ] `go test ./internal/...`
- [ ] `cd frontend && npm run typecheck && npm run i18n:check`
- [ ] Manual check of the UI path this PR touches (or mock `npm run dev`)

## Notes

- Locale keys added in **all** `frontend/src/locales/*.json` files
- No secrets, cookies, or `build/bin` artifacts
