# Security Guidelines

## Credential Management

This project requires GitHub tokens to access repository data. **Never commit credentials to version control.**

### Safe Practices

✅ **DO:**
- Store tokens in environment variables
- Add environment variables to your shell profile (`.bashrc`, `.zshrc`, etc.)
- Use `make check-creds` before committing
- Rotate tokens immediately if accidentally exposed
- Use separate tokens for different organizations
- Set appropriate token scopes (minimal required permissions)

❌ **DON'T:**
- Hardcode tokens in source files
- Commit `.env` files with real credentials
- Share tokens in chat or email
- Use production tokens in example documentation
- Grant unnecessary token scopes

### Environment Variables

The application uses these environment variables:

```bash
# Primary GitHub token (for mondoohq organization)
export GITHUB_TOKEN_ASSISTANT_MONDOOHQ="ghp_your_token_here"

# Community GitHub token (for mondoo-community organization)
export GITHUB_TOKEN_ASSISTANT_MONDOO_COMMUNITY="ghp_your_token_here"

# Optional: GitHub Project Node ID
export GITHUB_PROJECT="PVT_kwDOABpK8s4ApRn8"
```

Add these to your shell profile:

```bash
# For bash
echo 'export GITHUB_TOKEN_ASSISTANT_MONDOOHQ="ghp_..."' >> ~/.bashrc

# For zsh
echo 'export GITHUB_TOKEN_ASSISTANT_MONDOOHQ="ghp_..."' >> ~/.zshrc
```

### Token Scopes

Required GitHub token scopes:

- `repo` - Access private repositories
- `read:org` - Read organization data
- `project` - Access GitHub Projects v2 (optional, for `--github` with `GITHUB_PROJECT` filtering)

To add scopes to an existing token:
```bash
gh auth refresh -s project
```

### Checking for Credentials

Before committing code, run:

```bash
make check-creds
```

This checks for:
- GitHub tokens (ghp_, gho_, ghs_, ghr_, github_pat_)
- Private keys (RSA, DSA, EC, OpenSSH)
- AWS credentials (AKIA..., aws_access_key_id)
- API keys and tokens
- Database connection strings
- Hardcoded passwords

### Pre-commit Hook

Set up automatic checking:

```bash
# Create pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
make check-creds || exit 1
EOF

chmod +x .git/hooks/pre-commit
```

Now every commit will automatically check for credentials.

### If You Accidentally Commit Credentials

1. **Rotate the token immediately**:
   - Go to https://github.com/settings/tokens
   - Delete the exposed token
   - Generate a new token with the same scopes
   - Update your environment variables

2. **Remove from Git history** (if needed):
   ```bash
   # WARNING: This rewrites history - coordinate with team
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch path/to/file" \
     --prune-empty --tag-name-filter cat -- --all

   # Force push (dangerous)
   git push --force --all
   ```

3. **Better approach** - Use BFG Repo-Cleaner:
   ```bash
   # Install BFG
   brew install bfg

   # Remove credentials
   bfg --replace-text passwords.txt
   git reflog expire --expire=now --all
   git gc --prune=now --aggressive
   ```

4. **Consider the token compromised** - Even after removal from Git history, the token may have been:
   - Cached by GitHub
   - Cloned by others
   - Indexed by search engines
   - Captured in CI/CD logs

   **Always rotate exposed tokens immediately!**

## Service Configuration Security

When setting up system services (launchd, systemd), **never hardcode tokens** in the configuration files:

### macOS launchd

✅ **Good** - Use EnvironmentVariables:
```xml
<key>EnvironmentVariables</key>
<dict>
    <key>GITHUB_TOKEN_ASSISTANT_MONDOOHQ</key>
    <string>ghp_your_token_here</string>
</dict>
```

❌ **Bad** - Hardcoding in script:
```xml
<key>ProgramArguments</key>
<array>
    <string>/bin/bash</string>
    <string>-c</string>
    <string>GITHUB_TOKEN=ghp_xxx ./assistant --daemon</string>
</array>
```

### Linux systemd

✅ **Good** - Use Environment directives:
```ini
[Service]
Environment="GITHUB_TOKEN_ASSISTANT_MONDOOHQ=ghp_your_token_here"
```

Or load from a protected file:
```ini
[Service]
EnvironmentFile=/etc/assistant/credentials
```

Where `/etc/assistant/credentials` has restricted permissions:
```bash
sudo chmod 600 /etc/assistant/credentials
sudo chown root:root /etc/assistant/credentials
```

## Additional Security Considerations

### Database Security

The SQLite database (`~/.assistant/assistant.db`) contains your journal entries, exercises, and reminders.

Protect it:
```bash
chmod 600 ~/.assistant/assistant.db
```

### Log File Security

Daemon logs (`~/.assistant/assistant.log`) may contain sensitive information.

Protect them:
```bash
chmod 600 ~/.assistant/assistant.log
```

### Web UI Security

When running the web UI:
- It binds to `localhost` by default (not accessible from network)
- No authentication is implemented
- Do not expose to the internet without proper security measures
- Consider using SSH tunneling for remote access:
  ```bash
  ssh -L 8080:localhost:8080 user@remote-host
  ```

### Network Security

The application makes HTTPS requests to:
- `api.github.com` - GitHub API
- GitHub repository URLs

These connections use TLS/SSL for encryption.

## Reporting Security Issues

If you discover a security vulnerability, please:
1. Do not open a public GitHub issue
2. Contact the maintainer privately
3. Provide details about the vulnerability
4. Allow time for a fix before public disclosure

## Security Checklist

Before deploying:
- [ ] Tokens stored in environment variables only
- [ ] No credentials in source code
- [ ] No credentials in Git history
- [ ] Token scopes are minimal (least privilege)
- [ ] `make check-creds` passes
- [ ] Pre-commit hook installed
- [ ] Database file permissions set (600)
- [ ] Log file permissions set (600)
- [ ] Service configuration secured
- [ ] Web UI not exposed to internet

Regular maintenance:
- [ ] Rotate tokens periodically (every 90 days)
- [ ] Review token usage in GitHub settings
- [ ] Check for unused tokens
- [ ] Update dependencies for security patches
- [ ] Run `make check-creds` before each commit
