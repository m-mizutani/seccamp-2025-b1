# cred2md

A simple tool to convert JSON credential files to a Markdown table format.

## Usage

```bash
# From the cred2md directory
go run main.go

# Or build and run
go build -o cred2md
./cred2md
```

## Output

The tool will:
1. Read all JSON files from the `credentials` directory (relative to project root)
2. Parse username and password from each file
3. Output a Markdown table sorted by username
4. Include any additional notes from the credential files

## Example Output

```markdown
# AWS Credentials

| Username | Initial Password |
|----------|------------------|
| blue | xY9#mN2$pQ5@ |
| green | aB3!dE6&fG8* |
| red | sV3zZ(U2Z}cQ%Qxo=F\|s |

**Note:** Password reset required on first login

**Console URL:** https://console.aws.amazon.com/
```

## Credential File Format

The tool expects JSON files with the following structure:

```json
{
  "account_id": "123456789012",
  "console_url": "https://console.aws.amazon.com/",
  "instructions": "Password reset required on first login",
  "password": "example_password",
  "username": "example_user"
}
```