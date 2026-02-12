# 🛡️ Security

Security best practices and built-in protections for the Aureo VPN backend.

---

## ✅ Best Practices

### 1. Secrets Management
- Use strong, unique `JWT_SECRET` (32+ bytes)
- Store private keys in secure vault (HashiCorp Vault, AWS Secrets Manager)
- Rotate secrets regularly

### 2. Network Security
- Enable TLS for all API traffic
- Configure firewall rules
- Use VPN or private network for internal services
- Enable rate limiting

### 3. Database Security
- Restrict file permissions on SQLite database (chmod 600)
- Store database file in secure location
- Regular backups of the database file
- Consider encryption at rest for sensitive deployments

### 4. Monitoring
- Enable audit logging
- Set up alerts for suspicious activity
- Monitor failed login attempts
- Track API abuse patterns

---

## 🔒 Security Headers

The API Gateway automatically sets:

- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'`
- `Strict-Transport-Security: max-age=31536000`
