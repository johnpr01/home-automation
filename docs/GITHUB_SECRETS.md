# GitHub Actions Secrets Setup

This document explains how to configure GitHub Actions secrets for the Home Automation System.

## Required Secrets

### TPLINK_PASSWORD

This secret stores the password for your TP-Link Tapo account, which is used by the Tapo energy monitoring service.

## Setting Up GitHub Secrets

### Method 1: Via GitHub Web Interface

1. Navigate to your repository on GitHub
2. Click on **Settings** tab
3. In the left sidebar, click **Secrets and variables** → **Actions**
4. Click **New repository secret**
5. Set the following:
   - **Name**: `TPLINK_PASSWORD`
   - **Secret**: Your TP-Link Tapo account password
6. Click **Add secret**

### Method 2: Via GitHub CLI

```bash
# Install GitHub CLI if not already installed
# https://cli.github.com/

# Set the secret
gh secret set TPLINK_PASSWORD --body "your_tapo_account_password"

# Verify the secret was added
gh secret list
```

### Method 3: Via REST API

```bash
# You'll need a GitHub Personal Access Token with repo scope
# Replace YOUR_TOKEN, OWNER, REPO, and PASSWORD_VALUE

curl -X PUT \
  -H "Authorization: token YOUR_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/OWNER/REPO/actions/secrets/TPLINK_PASSWORD \
  -d '{"encrypted_value":"PASSWORD_VALUE","key_id":"KEY_ID"}'
```

## Usage in Workflows

The secret is automatically available as an environment variable in GitHub Actions:

```yaml
- name: Run Tapo Demo
  env:
    TPLINK_PASSWORD: ${{ secrets.TPLINK_PASSWORD }}
  run: |
    cd cmd/tapo-demo
    go run main.go
```

## Local Development

For local development, you can set the environment variable:

### Linux/macOS
```bash
export TPLINK_PASSWORD="your_tapo_account_password"
cd cmd/tapo-demo
go run main.go
```

### Windows (PowerShell)
```powershell
$env:TPLINK_PASSWORD="your_tapo_account_password"
cd cmd/tapo-demo
go run main.go
```

### Windows (Command Prompt)
```cmd
set TPLINK_PASSWORD=your_tapo_account_password
cd cmd/tapo-demo
go run main.go
```

## Environment File (Alternative for Local Development)

Create a `.env` file in the project root (this file should NOT be committed to Git):

```bash
# .env file
TPLINK_PASSWORD=your_tapo_account_password
```

Then load it in your application or use a tool like `direnv`.

## Security Best Practices

1. **Never commit passwords to Git**: Always use environment variables or secrets
2. **Use unique passwords**: Don't reuse passwords across services
3. **Rotate secrets regularly**: Update passwords periodically
4. **Limit access**: Only give secrets access to workflows that need them
5. **Monitor usage**: Review GitHub Actions logs for any unauthorized access

## Troubleshooting

### Secret Not Available
- Verify the secret name matches exactly (case-sensitive)
- Ensure you have admin access to the repository
- Check that the workflow has access to secrets

### Password Authentication Issues
- Verify your TP-Link account credentials are correct
- Check if your account has two-factor authentication enabled
- Ensure the account has access to the Tapo devices

### GitHub Actions Failures
```bash
# Check if the secret is properly set
echo "TPLINK_PASSWORD is set: ${{ secrets.TPLINK_PASSWORD != '' }}"
```

### Local Development Issues
```bash
# Check if environment variable is set
echo "TPLINK_PASSWORD: $TPLINK_PASSWORD"

# Test the application
cd cmd/tapo-demo
go run main.go
```

## Additional Configuration

### Multiple Tapo Accounts

If you have multiple Tapo accounts, you can add additional secrets:

- `TPLINK_PASSWORD_HOME`
- `TPLINK_PASSWORD_OFFICE`
- `TPLINK_USERNAME_HOME`
- `TPLINK_USERNAME_OFFICE`

Then modify the code to use the appropriate credentials:

```go
homePassword := os.Getenv("TPLINK_PASSWORD_HOME")
officePassword := os.Getenv("TPLINK_PASSWORD_OFFICE")
```

### Configuration via Files

For more complex setups, consider using a configuration file approach:

```yaml
# tapo-config.yml (encrypted or using secret management)
accounts:
  home:
    username: "home@example.com"
    password: "${TPLINK_PASSWORD_HOME}"
  office:
    username: "office@example.com"
    password: "${TPLINK_PASSWORD_OFFICE}"
```

## Related Documentation

- [GitHub Actions Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Tapo Energy Monitoring Setup](TAPO_ENERGY_MONITORING.md)
- [Environment Configuration](../configs/README.md)
