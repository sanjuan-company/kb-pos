# Mailing Service — Sample Requests

## Send Email (Plain Text)

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": [{"email": "recipient@example.com", "name": "Recipient Name"}],
    "subject": "Hello from the mailing service",
    "plain_text": "This is a plain text email sent via the mailing service."
  }'
```

## Send Email (HTML)

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": [{"email": "recipient@example.com"}],
    "subject": "Welcome!",
    "html": "<h1>Welcome</h1><p>Thank you for signing up.</p>",
    "plain_text": "Welcome! Thank you for signing up."
  }'
```

## Send with Custom Sender

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "from": {"email": "custom@example.com", "name": "Custom Sender"},
    "to": [{"email": "recipient@example.com"}],
    "cc": [{"email": "cc@example.com"}],
    "subject": "Email with custom sender and CC",
    "plain_text": "This has a custom from address and a CC recipient."
  }'
```

## Send with Idempotency Key (safe retry)

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: my-unique-key-123" \
  -d '{
    "to": [{"email": "recipient@example.com"}],
    "subject": "Idempotent email",
    "plain_text": "Running this twice will only send once."
  }'
```

## Send with Provider Hint (force specific provider)

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": [{"email": "recipient@example.com"}],
    "subject": "Forced provider",
    "plain_text": "This will only use the hinted provider.",
    "provider_hint": "sendgrid"
  }'
```

## Send Batch (multiple recipients)

```bash
curl -X POST http://localhost:8080/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": [
      {"email": "alice@example.com", "name": "Alice"},
      {"email": "bob@example.com", "name": "Bob"}
    ],
    "cc": [
      {"email": "charlie@example.com"}
    ],
    "subject": "Team announcement",
    "html": "<p>Hello team, this is an important announcement.</p>"
  }'
```

## Check Email Status

```bash
curl http://localhost:8080/api/v1/status/<email-id>
```

Replace `<email-id>` with the `id` returned from the send response.

## Health Check

```bash
curl http://localhost:8080/api/v1/health
```

---

## Example Response (Success)

```json
{
  "success": true,
  "message": "email processed",
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "status": "sent",
    "provider": "smtp"
  }
}
```

## Example Response (Failure)

```json
{
  "success": false,
  "message": "all providers exhausted, last error: 535 Authentication failed",
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "status": "failed",
    "provider": "smtp"
  }
}
```

## How to generate a Gmail App Password

1. Go to https://myaccount.google.com/security
2. Enable **2-Step Verification** (required)
3. Go to **App passwords** (search in Google Account settings)
4. Select **Mail** and **Other (custom name)** → enter "mailing-service"
5. Copy the 16-character password
6. Set it as `SMTP_PASSWORD` in your `.env` file
