# Security

## Credentials handling

- Never commit service account JSON files to this repository.
- Keep credentials local and pass them via `--credentials-file` or `GSC_CLI_CREDENTIALS_FILE`.
- Use least privilege: grant the service account read-only access to Search Console properties.

## If a key is exposed

1. Revoke/disable the leaked key in Google Cloud immediately.
2. Create a new key and update local environments.
3. Review logs and recent access.
4. Rotate any related credentials if needed.
