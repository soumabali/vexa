# Network Architecture: Security Zones and Segmentation

**Document ID:** SEC-NET-001  
**Version:** 1.0  
**Classification:** Internal — Security Critical  
**Owner:** Security Architect  
**Date:** 2026-05-27

---

## 1. Network Architecture Overview

### 1.1 High-Level Topology

```
                                    ┌─────────────────┐
                                    │    INTERNET     │
                                    └────────┬────────┘
                                             │
                                             ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              PERIMETER / DMZ                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   CDN/WAF    │  │   DDoS       │  │   VPN        │  │   Bastion    │ │
│  │  (CloudFlare)│  │  Protection  │  │  Gateway     │  │  Host        │ │
│  │              │  │              │  │  (WireGuard) │  │  (Emergency) │ │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘ │
│                                                                             │
│  Trust Level: Untrusted → Sanitized                                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           APPLICATION TIER                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         API Gateway                                  │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │   │
│  │  │ Rate     │  │  Auth    │  │  WAF     │  │  TLS Termination │  │   │
│  │  │ Limiter  │  │  Check   │  │  Rules   │  │  (TLS 1.3)       │  │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                       Web Application                               │   │
│  │              (Next.js 16 SPA served via CDN)                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  Trust Level: Semi-Trusted → Authenticated                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SERVICE TIER                                       │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │   Auth      │  │   API       │  │  Session    │  │  Worker     │      │
│  │  Service    │  │  Server     │  │  Manager    │  │  Pool       │      │
│  │  (OAuth2)   │  │  (REST/gRPC)│  │  (WebSocket)│  │  (Async)    │      │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                       │
│  │  Credential │  │   Audit     │  │   Config    │                       │
│  │   Vault     │  │  Service    │  │  Service    │                       │
│  │  (HashiCorp)│  │  (Immutable)│  │  (Etcd)     │                       │
│  └─────────────┘  └─────────────┘  └─────────────┘                       │
│                                                                             │
│  Trust Level: Trusted → Service Identity Required                         │
│  Communication: mTLS + Service Mesh                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DATA TIER                                          │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │  PostgreSQL │  │   Redis     │  │   MinIO/S3  │  │   Etcd      │      │
│  │  (Metadata) │  │  (Session   │  │  (Backups/  │  │  (Config/   │      │
│  │             │  │   State)    │  │  Recordings)│  │  Discovery) │      │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                                             │
│  Trust Level: Trusted → Encrypted + Access Controlled                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           MANAGEMENT / INFRASTRUCTURE                       │
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │   HSM/KMS   │  │   Vault     │  │   SIEM      │  │   Jump      │      │
│  │  (YubiHSM)  │  │  Server     │  │  Collector  │  │  Server     │      │
│  │             │  │  (Secrets)  │  │  (Splunk)   │  │  (Admin)    │      │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘      │
│                                                                             │
│  Trust Level: Highly Trusted → Physical Access + M-of-N                  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. DMZ Considerations

### 2.1 DMZ Design Rationale

The terminal manager requires a **controlled DMZ** rather than a traditional open DMZ because:

1. **Self-hosted nature** — No cloud provider security groups as primary defense
2. **Credential sensitivity** — Any breach is catastrophic
3. **Multiple access methods** — Web, API, VPN, emergency access
4. **Protocol translation** — WebSocket-to-SSH/RDP/VNC proxy adds attack surface

### 2.2 DMZ Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DMZ SEGMENT                                     │
│  Network: 10.0.1.0/24                                                       │
│  Firewall: Default deny, explicit allow                                      │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         INGRESS LAYER                                │   │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐                     │   │
│  │  │ Public   │    │ CDN      │    │ DNS      │                     │   │
│  │  │ IP: .10  │    │ CNAME    │    │ A/AAAA   │                     │   │
│  │  │ Ports:   │    │          │    │          │                     │   │
│  │  │ 443/tcp  │    │          │    │          │                     │   │
│  │  │ 80/tcp   │    │          │    │          │                     │   │
│  │  │ 51820/udp│    │          │    │          │                     │   │
│  │  └──────────┘    └──────────┘    └──────────┘                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         PROCESSING LAYER                             │   │
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐                     │   │
│  │  │ Reverse  │    │ WAF      │    │ Rate     │                     │   │
│  │  │ Proxy    │    │ (ModSec) │    │ Limiter  │                     │   │
│  │  │ (Nginx)  │    │          │    │ (Redis)  │                     │   │
│  │  │ IP: .20  │    │          │    │          │                     │   │
│  │  └──────────┘    └──────────┘    └──────────┘                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         ACCESS LAYER                                 │   │
│  │  ┌──────────┐    ┌──────────┐                                     │   │
│  │  │ VPN GW   │    │ Bastion  │                                     │   │
│  │  │ (WG)     │    │ Host     │                                     │   │
│  │  │ IP: .30  │    │ IP: .31  │                                     │   │
│  │  │ Port:    │    │ Port:    │                                     │   │
│  │  │ 51820/udp│    │ 22/tcp   │                                     │   │
│  │  └──────────┘    └──────────┘                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  FIREWALL RULES (DMZ):                                                      │
│  • Allow IN: 443/tcp from ANY → Reverse Proxy                              │
│  • Allow IN: 51820/udp from ANY → VPN Gateway                              │
│  • Allow IN: 22/tcp from Admin-CIDR → Bastion (emergency only)             │
│  • Deny all other inbound                                                  │
│  • Allow OUT: 443/tcp from RP → App Tier (specific IPs)                    │
│  • Allow OUT: 22/tcp from Bastion → Jump Server (management)               │
│  • Deny all other outbound                                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.3 DMZ Security Controls

| Control | Implementation | Purpose |
|---------|---------------|---------|
| **Network Isolation** | Separate VLAN/VPC with no direct LAN access | Prevent lateral movement |
| **Single Entry Point** | All traffic through API Gateway | Centralized inspection |
| **No Direct Admin Access** | Admin access via VPN only, never from DMZ | Prevent admin compromise |
| **Hardened Hosts** | CIS benchmarks, minimal packages, read-only FS | Reduce attack surface |
| **IDS/IPS** | Suricata/Zeek on DMZ segment | Detect scanning, exploitation |
| **Egress Filtering** | Default deny outbound, only required ports | Prevent C2 communication |
| **Port Knocking** | Optional for emergency SSH access | Hide management ports |

---

## 3. Internal Network Segmentation

### 3.1 Segment Design

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         NETWORK SEGMENTS                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Segment: dmz-public       CIDR: 10.0.1.0/24    VLAN: 101                 │
│  ├─ Purpose: Internet-facing services                                       │
│  ├─ Contains: CDN endpoint, reverse proxy, WAF, VPN GW                    │
│  └─ Policy: Allow 443/tcp inbound, deny direct LAN access                  │
│                                                                             │
│  Segment: app-tier         CIDR: 10.0.2.0/24    VLAN: 102                 │
│  ├─ Purpose: Application servers                                            │
│  ├─ Contains: Web app containers, API servers, session proxies              │
│  └─ Policy: Allow 443/tcp from DMZ, deny direct DB access                  │
│                                                                             │
│  Segment: service-tier     CIDR: 10.0.3.0/24    VLAN: 103                 │
│  ├─ Purpose: Internal microservices                                         │
│  ├─ Contains: Auth service, credential vault, audit service                  │
│  └─ Policy: mTLS required, deny all except service mesh                    │
│                                                                             │
│  Segment: data-tier        CIDR: 10.0.4.0/24    VLAN: 104                 │
│  ├─ Purpose: Databases and storage                                          │
│  ├─ Contains: PostgreSQL, Redis, MinIO, etcd                                │
│  └─ Policy: Allow only from service-tier, encrypted connections required   │
│                                                                             │
│  Segment: management       CIDR: 10.0.5.0/24    VLAN: 105                 │
│  ├─ Purpose: Infrastructure management                                     │
│  ├─ Contains: HSM, Vault server, SIEM, jump hosts, monitoring            │
│  └─ Policy: VPN required, no direct internet access, physical access log   │
│                                                                             │
│  Segment: target-servers   CIDR: 10.0.6.0/24    VLAN: 106               │
│  ├─ Purpose: Servers managed by the terminal manager                        │
│  ├─ Contains: SSH servers, RDP endpoints, VNC servers                      │
│  └─ Policy: Allow only from session manager, bastion host access         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Micro-Segmentation with Network Policies

```yaml
# Kubernetes Network Policy (or Calico/Cilium equivalent)
# Deny all by default
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: terminal-manager
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
# Allow API Gateway → API Server
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-gateway-to-api
  namespace: terminal-manager
spec:
  podSelector:
    matchLabels:
      app: api-server
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: api-gateway
    ports:
    - protocol: TCP
      port: 8443
---
# Allow API Server → Credential Vault (strict)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-to-vault
  namespace: terminal-manager
spec:
  podSelector:
    matchLabels:
      app: credential-vault
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: api-server
    ports:
    - protocol: TCP
      port: 8200
---
# Allow Session Manager → Target Servers
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-session-to-targets
  namespace: terminal-manager
spec:
  podSelector:
    matchLabels:
      app: session-manager
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: target-servers
    ports:
    - protocol: TCP
      port: 22    # SSH
    - protocol: TCP
      port: 3389  # RDP
    - protocol: TCP
      port: 5900  # VNC
```

### 3.3 East-West Traffic Inspection

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    EAST-WEST TRAFFIC INSPECTION                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  All service-to-service traffic flows through service mesh:                 │
│                                                                             │
│  ┌─────────┐     ┌─────────────┐     ┌─────────┐                          │
│  │ Service │────▶│  Envoy Sidecar│────▶│ Service │                          │
│  │    A    │     │  (Istio Proxy)│     │    B    │                          │
│  └─────────┘     └─────────────┘     └─────────┘                          │
│                         │                                                   │
│                         ▼                                                   │
│                  ┌─────────────┐                                            │
│                  │  Telemetry  │                                            │
│                  │  (Metrics + │                                            │
│                  │   Traces)   │                                            │
│                  └─────────────┘                                            │
│                         │                                                   │
│                         ▼                                                   │
│                  ┌─────────────┐                                            │
│                  │  Anomaly    │                                            │
│                  │  Detection  │                                            │
│                  └─────────────┘                                            │
│                                                                             │
│  Benefits:                                                                  │
│  • Mutual TLS without application changes                                   │
│  • Traffic metrics for anomaly detection                                    │
│  • Circuit breaking and retries                                             │
│  • Canary deployments with traffic splitting                                │
│  • Access logging for audit                                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. API Gateway / Reverse Proxy Requirements

### 4.1 Gateway Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         API GATEWAY STACK                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  LAYER 1: Edge (CDN + DDoS)                                         │   │
│  │  • CloudFlare / AWS CloudFront / Fastly                              │   │
│  │  • Static asset caching, geo-optimization                            │   │
│  │  • DDoS mitigation (L3/L4/L7)                                        │   │
│  │  • Bot management                                                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  LAYER 2: Load Balancer (L7)                                         │   │
│  │  • Nginx / HAProxy / Traefik                                         │   │
│  │  • SSL termination (TLS 1.3)                                         │   │
│  │  • Connection pooling                                                │   │
│  │  • Health checks                                                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  LAYER 3: API Gateway (Kong / Ambassador / APISIX)                 │   │
│  │  • Authentication (JWT validation)                                   │   │
│  │  • Rate limiting (per user, per endpoint)                            │   │
│  │  • Request validation (OpenAPI schema)                               │   │
│  │  • Request/response transformation                                   │   │
│  │  • Caching                                                           │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  LAYER 4: WAF (ModSecurity / Coraza)                               │   │
│  │  • OWASP Core Rule Set                                               │   │
│  │  • Custom rules for terminal-specific attacks                        │   │
│  │  • Virtual patching                                                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                              │
│                              ▼                                              │
│                    [Application Tier]                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Gateway Security Configuration

```nginx
# Nginx configuration for terminal manager
# /etc/nginx/conf.d/terminal-manager.conf

# Rate limiting zones
limit_req_zone $binary_remote_addr zone=auth:10m rate=5r/m;
limit_req_zone $binary_remote_addr zone=api:10m rate=100r/m;
limit_req_zone $binary_remote_addr zone=terminal:10m rate=30r/m;
limit_conn_zone $binary_remote_addr zone=conn:10m;

# Upstream definitions
upstream api_servers {
    least_conn;
    server 10.0.2.10:8443 weight=5;
    server 10.0.2.11:8443 weight=5;
    server 10.0.2.12:8443 backup;
    
    keepalive 32;
    keepalive_timeout 60s;
}

upstream websocket_servers {
    ip_hash;  # Sticky sessions for WebSocket
    server 10.0.2.20:8443;
    server 10.0.2.21:8443;
}

server {
    listen 443 ssl http2;
    server_name terminal.example.com;
    
    # TLS configuration (see security-architecture.md)
    ssl_certificate /etc/ssl/certs/terminal-manager.crt;
    ssl_certificate_key /etc/ssl/private/terminal-manager.key;
    include /etc/nginx/conf.d/tls-security.conf;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;
    
    # CSP for terminal manager
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' (script hashing if possible); style-src 'self' 'unsafe-inline'; connect-src 'self' wss://terminal.example.com; img-src 'self' data:; font-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';" always;
    
    # API routes
    location /api/v1/ {
        limit_req zone=api burst=20 nodelay;
        limit_conn conn 10;
        
        proxy_pass https://api_servers/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 5s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Buffer sizes
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }
    
    # Authentication routes (stricter rate limiting)
    location /api/v1/auth/ {
        limit_req zone=auth burst=3 nodelay;
        limit_conn conn 5;
        
        proxy_pass https://api_servers/api/v1/auth/;
        # ... same proxy settings
    }
    
    # WebSocket terminal sessions
    location /ws/terminal/ {
        limit_req zone=terminal burst=10 nodelay;
        limit_conn conn 5;
        
        proxy_pass https://websocket_servers/;
        proxy_http_version 1.1;
        
        # WebSocket upgrade
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # WebSocket specific timeouts
        proxy_read_timeout 86400s;
        proxy_send_timeout 86400s;
        
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    # Static assets (cached)
    location /static/ {
        alias /var/www/terminal-manager/static/;
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }
    
    # Health check (unauthenticated, rate limited)
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
    
    # Deny access to hidden files
    location ~ /\. {
        deny all;
        return 404;
    }
}
```

### 4.3 Web Application Firewall Rules

```
# ModSecurity / Coraza custom rules for terminal manager
# /etc/modsecurity/terminal-manager.conf

# Rule 1: Block terminal escape sequences in HTTP headers
SecRule REQUEST_HEADERS "@rx \x1b\[|\x07|\x0e|\x0f" \
    "id:1001,phase:2,block,log,msg:'Terminal escape sequence detected in headers'"

# Rule 2: Block oversized JWT tokens (possible DoS)
SecRule REQUEST_HEADERS:Authorization "@gt 4096" \
    "id:1002,phase:1,block,log,msg:'Oversized authorization header'"

# Rule 3: Prevent credential vault path traversal
SecRule REQUEST_URI "@rx /vault/.*\.\./" \
    "id:1003,phase:1,block,log,msg:'Path traversal attempt on vault endpoint'"

# Rule 4: Block WebSocket protocol confusion attacks
SecRule REQUEST_HEADERS:Upgrade "@streq websocket" \
    "chain,id:1004,phase:1,block,log"
    SecRule REQUEST_HEADERS:Connection "!@contains upgrade" \
        "msg:'WebSocket upgrade without proper Connection header'"

# Rule 5: Rate limit credential access endpoints
SecAction "id:1005,phase:1,nolog,pass,setvar:IP.cred_access_counter=+1,expirevar:IP.cred_access_counter=60"
SecRule IP:CRED_ACCESS_COUNTER "@gt 10" \
    "id:1006,phase:1,block,log,msg:'Credential access rate limit exceeded'"

# Rule 6: Block known bad user agents (scanners)
SecRule REQUEST_HEADERS:User-Agent "@rx (sqlmap|nikto|nmap|masscan|zgrab)" \
    "id:1007,phase:1,block,log,msg:'Known scanner user agent detected'"

# Rule 7: Require Content-Type for POST/PUT
SecRule REQUEST_METHOD "@rx ^(POST|PUT)$" \
    "chain,id:1008,phase:1,block,log,msg:'Missing Content-Type header'"
    SecRule REQUEST_HEADERS:Content-Type "@rx ^$" \
        "t:lowercase"
```

---

## 5. VPN/Tunnel Requirements

### 5.1 VPN Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         VPN ARCHITECTURE                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  PRIMARY: WireGuard (Modern, performant, simple)                           │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  WireGuard Configuration                                             │   │
│  │  ├── Protocol: UDP (port 51820)                                    │   │
│  │  ├── Cryptography: Noise Protocol Framework (Curve25519)            │   │
│  │  ├── Authentication: Pre-shared keys + certificates                 │   │
│  │  ├── Network: Point-to-site (road warriors) / site-to-site         │   │
│  │  └── Mesh: Optional (for multi-site deployments)                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  FALLBACK: OpenVPN (Broad compatibility)                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  OpenVPN Configuration                                               │   │
│  │  ├── Protocol: UDP preferred (port 1194)                           │   │
│  │  ├── Cryptography: AES-256-GCM + TLS 1.3                           │   │
│  │  ├── Authentication: Certificates + optional MFA                      │   │
│  │  └── Hardening: tls-crypt, strong ciphers, no compression           │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  EMERGENCY: SSH Tunnel (Last resort)                                       │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  SSH Jump Host Configuration                                         │   │
│  │  ├── Authentication: Certificate-based only (no passwords)            │   │
│  │  ├── Access: Key-based + IP allowlist                               │   │
│  │  ├── Hardening: Fail2ban, non-standard port, no root login          │   │
│  │  └── Monitoring: All sessions logged and alerted                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.2 WireGuard Configuration

```ini
# /etc/wireguard/wg0.conf (Server)
[Interface]
PrivateKey = <server-private-key>
Address = 10.200.200.1/24
ListenPort = 51820

# iptables rules for traffic control
PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

# Client: Admin Workstation
[Peer]
PublicKey = <admin-public-key>
AllowedIPs = 10.200.200.2/32
PersistentKeepalive = 25

# Client: Developer Workstation
[Peer]
PublicKey = <dev-public-key>
AllowedIPs = 10.200.200.3/32
# Note: Developer has restricted access via firewall rules

# Client: Monitoring System (read-only)
[Peer]
PublicKey = <monitoring-public-key>
AllowedIPs = 10.200.200.10/32

# Firewall rules for WireGuard network:
# - Admin (10.200.200.2): Full access to management tier
# - Developer (10.200.200.3): Access to app tier, no management
# - Monitoring (10.200.200.10): Access to metrics endpoints only
```

### 5.3 Site-to-Site Tunneling for Target Servers

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    SITE-TO-SITE: HQ ↔ BRANCH OFFICE                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐                              ┌──────────────┐           │
│  │   HQ Data    │◄──── WireGuard Tunnel ──────►│ Branch Office│           │
│  │   Center     │      (Encrypted, UDP 51820)   │   Network      │           │
│  │              │                              │               │           │
│  │  ┌────────┐  │                              │  ┌────────┐    │           │
│  │  │Terminal│  │                              │  │Managed │    │           │
│  │  │Manager │  │◄─────── SSH/RDP/VNC ────────►│  │Servers │    │           │
│  │  │        │  │      (Through tunnel)         │  │        │    │           │
│  │  └────────┘  │                              │  └────────┘    │           │
│  └──────────────┘                              └──────────────┘           │
│                                                                             │
│  Configuration:                                                             │
│  • Tunnel terminates at branch office WireGuard peer                       │
│  • Terminal Manager session manager routes through tunnel                  │
│  • Branch office target servers: 10.1.0.0/16                               │
│  • HQ Terminal Manager: 10.0.2.0/24                                        │
│  • AllowedIPs in WireGuard: 10.1.0.0/16, 10.0.2.0/24                       │
│                                                                             │
│  Security:                                                                  │
│  • Tunnel keys rotated monthly                                             │
│  • Branch office peer authenticated by certificate + pre-shared key        │
│  • No split tunneling — all management traffic through encrypted tunnel    │
│  • Automatic tunnel health checks with failover to backup endpoint         │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 5.4 Tunnel Security Requirements

| Requirement | Implementation | Verification |
|-------------|---------------|------------|
| Encryption | WireGuard (ChaCha20-Poly1305) or OpenVPN (AES-256-GCM) | Cipher suite audit |
| Authentication | Certificate-based + optional MFA | No password-only auth |
| Key Management | Automated rotation, key escrow in Vault | Quarterly review |
| Network Segmentation | Split tunneling disabled for management traffic | Firewall rule audit |
| Endpoint Hardening | Minimal OS, auto-updates, EDR | Monthly compliance scan |
| Monitoring | All tunnel connections logged | Real-time alerting |
| Failover | Backup tunnel endpoint, automatic detection | Quarterly DR test |

---

## 6. Firewall Rules Summary

### 6.1 Per-Segment Firewall Matrix

| Source | Destination | Protocol | Port | Action | Notes |
|--------|-------------|----------|------|--------|-------|
| Internet | DMZ (Reverse Proxy) | TCP | 443 | ALLOW | TLS 1.3 only |
| Internet | DMZ (VPN Gateway) | UDP | 51820 | ALLOW | WireGuard |
| Internet | DMZ (Bastion) | TCP | 22 | DENY | Emergency only, IP allowlist |
| DMZ | App Tier | TCP | 8443 | ALLOW | mTLS required |
| DMZ | Any | Any | Any | DENY | Default deny |
| App Tier | Service Tier | TCP | 8443, 8200 | ALLOW | mTLS, specific services |
| App Tier | Data Tier | TCP | 5432, 6379, 9000 | ALLOW | TLS required |
| App Tier | Internet | Any | Any | DENY | No direct egress |
| Service Tier | Data Tier | TCP | 5432, 6379, 9000, 2379 | ALLOW | mTLS |
| Service Tier | Management | TCP | 8200 | ALLOW | Vault access only |
| Service Tier | Internet | Any | Any | DENY | No egress |
| Data Tier | Any | Any | Any | DENY | No outbound connections |
| Management | App Tier | TCP | 22 | ALLOW | Jump server access |
| Management | HSM/KMS | TCP | 443 | ALLOW | Administrative only |
| VPN Clients | Management | TCP | 22, 443 | ALLOW | Admin VPN required |
| VPN Clients | App Tier | TCP | 443 | ALLOW | Standard user access |
| Session Mgr | Target Servers | TCP | 22, 3389, 5900 | ALLOW | Protocol-specific |

---

## 7. Network Monitoring and Visibility

### 7.1 Monitoring Stack

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      NETWORK MONITORING STACK                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Traffic Capture:                                                            │
│  • Zeek / Suricata on DMZ mirror port                                       │
│  • eBPF-based flow monitoring (Cilium Hubble)                               │
│  • Istio telemetry for service mesh                                          │
│                                                                             │
│  Metrics Collection:                                                         │
│  • Prometheus for infrastructure metrics                                  │
│  • Custom exporters for application metrics                               │
│  • Node exporter for host-level metrics                                    │
│                                                                             │
│  Log Aggregation:                                                            │
│  • Fluent Bit for log shipping                                             │
│  • Kafka for high-throughput log streaming                                │
│  • Splunk / Datadog for analysis and alerting                              │
│                                                                             │
│  Alerting:                                                                   │
│  • PagerDuty / OpsGenie for on-call escalation                             │
│  • Custom webhooks for automated response                                  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 7.2 Network Anomaly Detection

| Anomaly | Detection Method | Response |
|---------|-----------------|----------|
| Port scan | Zeek connection logs + threshold | Alert SOC, temporary IP block |
| Lateral movement | East-west traffic baseline deviation | Isolate segment, investigate |
| Data exfiltration | Unusual outbound volume + destination | Rate limit, alert, capture packets |
| C2 communication | DNS anomaly + beaconing detection | Block domain, isolate host |
| Tunnel abuse | Unusual tunnel duration/destination | Alert, review session recording |

---

**Document History**
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-05-27 | Security Architect | Initial network architecture |
