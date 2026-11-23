# Update the Homebrew tap from the release workflow

The LaunchPal release workflow updates the Homebrew tap directly, using the checksum of the published DMG. With one app to maintain, a separate tap workflow triggered by `repository_dispatch` would add coordination without enough benefit. Manual updates risk leaving the cask out of sync.

The default workflow token cannot write to a separate repository. Use a fine-grained personal access token restricted to the tap repository with Contents write permission. This avoids managing a GitHub App, at the cost of rotating the token before expiry; an expired token prevents the cask update even if the app release succeeds.
