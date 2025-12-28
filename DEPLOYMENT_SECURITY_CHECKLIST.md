# Aureo VPN Production Deployment Security Checklist

Use this checklist before deploying to production. All items marked **[CRITICAL]** must be completed.

---

## Pre-Deployment Verification

### 1. Secrets & Configuration

- [ ] **[CRITICAL]** JWT_SECRET is at least 32 bytes of cryptographically random data
- [ ] **[CRITICAL]** JWT_SECRET is NOT the NodeID or any predictable value
- [ ] **[CRITICAL]** ETHEREUM_PRIVATE_KEY is stored in HSM/Vault, not environment variable
- [ ] **[CRITICAL]** All default/development secrets have been rotated
- [ ] All API keys are rotated and scoped appropriately
- [ ] Database encryption key is properly secured
- [ ] TLS certificates are valid and not expiring soon

### 2. Authentication

- [ ] **[CRITICAL]** Access token lifetime is 15 minutes or less
- [ ] **[CRITICAL]** Refresh token rotation is enabled
- [ ] Token blacklisting is configured
- [ ] MFA is enforced for operator accounts
- [ ] Password requirements meet minimum standards (12+ chars, complexity)
- [ ] Account lockout after failed attempts is enabled

### 3. Database

- [ ] **[CRITICAL]** Database file permissions are restrictive (0600)
- [ ] **[CRITICAL]** No default credentials in use
- [ ] Database encryption at rest is enabled (SQLCipher)
- [ ] Backup encryption is configured
- [ ] Backup integrity testing is automated
- [ ] Connection pooling limits are configured

### 4. API Security

- [ ] **[CRITICAL]** Rate limiting is enabled on all endpoints
- [ ] **[CRITICAL]** Stricter rate limiting on auth endpoints
- [ ] **[CRITICAL]** All user input is validated before use
- [ ] **[CRITICAL]** Errors do not expose internal details
- [ ] Pagination has maximum limits enforced
- [ ] Request size limits are configured
- [ ] Request timeout is configured

### 5. Privacy

- [ ] **[CRITICAL]** IP addresses are NOT logged
- [ ] **[CRITICAL]** IP addresses are NOT stored in database
- [ ] **[CRITICAL]** Fiber logger does not include ${ip}
- [ ] Session metadata is minimized
- [ ] Data retention policies are configured
- [ ] GDPR right-to-erasure is implemented

### 6. Network Security

- [ ] **[CRITICAL]** TLS 1.3 is enforced
- [ ] **[CRITICAL]** CORS is configured with explicit origins (not "*")
- [ ] **[CRITICAL]** HTTPS-only (no HTTP fallback)
- [ ] Security headers are configured (HSTS, CSP, X-Frame-Options)
- [ ] Firewall rules are configured
- [ ] DDoS protection is in place

### 7. Blockchain

- [ ] **[CRITICAL]** Mock transaction fallback is disabled
- [ ] **[CRITICAL]** Real transaction verification is implemented
- [ ] **[CRITICAL]** Price oracle is integrated (no hardcoded prices)
- [ ] Nonce management has mutex protection
- [ ] Transaction retry logic is implemented
- [ ] Multi-signature for treasury wallet

### 8. Monitoring & Logging

- [ ] **[CRITICAL]** Security event logging is enabled
- [ ] **[CRITICAL]** Logs are sanitized of PII
- [ ] Intrusion detection alerts are configured
- [ ] Error rate monitoring is configured
- [ ] Uptime monitoring is configured
- [ ] Log retention policy is configured

### 9. Infrastructure

- [ ] **[CRITICAL]** Container runs as non-root user
- [ ] **[CRITICAL]** Container image is scanned for vulnerabilities
- [ ] Secrets are mounted from Vault, not environment
- [ ] Read-only root filesystem where possible
- [ ] Network segmentation is configured
- [ ] Backup automation is tested

### 10. Code Quality

- [ ] **[CRITICAL]** All tests pass with `-race` flag
- [ ] **[CRITICAL]** No HIGH/CRITICAL gosec findings
- [ ] Test coverage is 80%+
- [ ] Dependencies are updated
- [ ] No known vulnerabilities in dependencies

---

## Deployment Day

### Before Deployment

```bash
# 1. Run security scanner
gosec ./...

# 2. Run tests with race detector
go test -race -cover ./...

# 3. Check for vulnerable dependencies
go list -m all | xargs -n1 go list -m -json | jq 'select(.Retracted != null)'

# 4. Verify configuration
echo "Checking JWT secret length..."
[ ${#JWT_SECRET} -ge 32 ] && echo "OK" || echo "FAIL: JWT_SECRET too short"

# 5. Verify TLS certificate
openssl s_client -connect api.aureo-vpn.com:443 -servername api.aureo-vpn.com < /dev/null 2>/dev/null | openssl x509 -noout -dates
```

### During Deployment

1. Deploy to staging first
2. Run smoke tests on staging
3. Verify security headers with `curl -I`
4. Test rate limiting is working
5. Verify error responses don't leak info
6. Test authentication flow
7. Deploy to production (blue-green or canary)

### After Deployment

```bash
# Verify security headers
curl -sI https://api.aureo-vpn.com/health | grep -E "^(X-|Strict-Transport)"

# Test rate limiting
for i in {1..10}; do curl -s -o /dev/null -w "%{http_code}\n" https://api.aureo-vpn.com/api/v1/auth/login -X POST; done

# Verify error sanitization
curl -X POST https://api.aureo-vpn.com/api/v1/auth/login -d '{"email":"invalid"}' | jq .error
# Should NOT contain: SQL, paths, stack traces
```

---

## Incident Response Plan

### Severity Levels

| Level | Description | Response Time |
|-------|-------------|---------------|
| P1 | Active breach, data exfiltration | Immediate |
| P2 | Vulnerability being exploited | 1 hour |
| P3 | Critical vulnerability discovered | 4 hours |
| P4 | Security improvement needed | 24 hours |

### Response Steps

1. **Contain**: Isolate affected systems
2. **Assess**: Determine scope and impact
3. **Eradicate**: Remove threat
4. **Recover**: Restore services
5. **Lessons**: Post-incident review

### Emergency Contacts

- Security Lead: [REDACTED]
- On-call Engineer: [REDACTED]
- Legal: [REDACTED]
- PR: [REDACTED]

### Emergency Actions

```bash
# Block all access (emergency)
iptables -I INPUT -j DROP

# Rotate all secrets
./scripts/rotate-secrets.sh

# Revoke all active sessions
curl -X POST https://api.aureo-vpn.com/admin/revoke-all-sessions \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Enable maintenance mode
curl -X POST https://api.aureo-vpn.com/admin/maintenance \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"enabled": true}'
```

---

## Regular Security Tasks

### Daily
- [ ] Review security alerts
- [ ] Check error rates
- [ ] Verify backup completion

### Weekly
- [ ] Review access logs
- [ ] Check for new CVEs
- [ ] Update dependencies if needed

### Monthly
- [ ] Rotate non-critical secrets
- [ ] Review access permissions
- [ ] Test backup restoration
- [ ] Security training update

### Quarterly
- [ ] Full security audit
- [ ] Penetration testing
- [ ] Rotate all secrets
- [ ] Review and update this checklist

---

## Useful Commands

```bash
# Check for common security issues
grep -r "TODO\|FIXME\|HACK\|XXX" --include="*.go" .

# Find hardcoded secrets
grep -rE "(password|secret|key|token)\s*[:=]\s*['\"][^'\"]+['\"]" --include="*.go" .

# Check file permissions
find . -name "*.key" -o -name "*.pem" | xargs ls -la

# Verify no debug mode
grep -r "debug\s*[:=]\s*true" --include="*.go" --include="*.yaml" .

# Check for dangerous functions
grep -rE "(unsafe|reflect\.)" --include="*.go" .
```

---

**Last Updated:** December 2024
**Owner:** Security Team
**Review Frequency:** Quarterly
