## 4Cs and Security Principles

I like to think about production security in terms of 4 layers, commonly known as the 4Cs: `Cloud`, `Cluster`, `Container`, `Code`. For each layer, it is best practice to follow these principles whenever possible:

- Least privilege
- Defense in depth
- Reduce attack surface
- Minimize blast radius
- Separation of duties

In addition, some form of governance, visibility, and feedback mechanism is needed to verify these principles are actually followed as we move along.

I have loosely categorized security measures below for structure and readability. These reflect practices I have seen organizations follow to secure their workloads.

---

## Cloud

### Least Privilege
- Tightly scoped IAM roles; prefer read-only access and avoid wildcard permissions
- SSO and RBAC for all tooling access (ArgoCD, AWS Console, etc.)

### Defense in Depth
- Proper VPC, subnet, NACL, and firewall configuration as the baseline
- WAF, rate limiting, and DDoS protection at the edge
- VPC alone is not enough — use a zero-trust mesh VPN (e.g., Tailscale) for internal tooling, auditing, and centralized access management
- IP whitelisting on IAM for mission-critical accounts prevents API abuse even if credentials leak

### Reduce Attack Surface
- No static IAM credentials; use GitHub OIDC for CI/CD deployments
- Never expose internal tooling (Grafana, Prometheus, ArgoCD) on public endpoints, even with authentication
- Encryption at rest and in transit as a non-negotiable baseline

### Minimize Blast Radius
- Isolate environments using separate accounts or OUs; blast radius from a compromised account should not reach production

### Separation of Duties
- Only CD systems manage production accounts and infrastructure; human access should be read-only in normal operation

### Governance
- CSPM tools (AWS Security Hub, AWS Config) for continuous posture management
- CloudTrail logs and event tracking
- Drift detection and reporting

---

## Cluster

### Least Privilege
- Sound RBAC policies for both users and service accounts
- Use workload identity (IRSA, Pod Identity) for pod-level cloud access instead of broad node-level credentials

### Defense in Depth
- Log and monitor all control plane components including API server audit logs
- Use CNIs with network policy support and visibility (e.g., Cilium)
- Admission controllers (Kyverno or OPA/Gatekeeper) to enforce security policies at scale — RBAC alone is not sufficient

### Reduce Attack Surface
- External secrets management (HashiCorp Vault, AWS Secrets Manager, 1Password); never store secrets in plaintext
- Use managed/vetted AMIs; tools like Karpenter enable timely node recycling and AMI rotation
- Timely upgrading of Kubernetes versions and associated controllers
- Never use client certificates to authenticate cluster users
- Encrypt etcd at rest
- Enable and enforce Pod Security Standards/Admission

### Minimize Blast Radius
- Fine-grained network policies limiting communication only between required components and namespaces

### Separation of Duties
- Only CD systems deploy to production namespaces; developers should not have direct write access to production

### Governance
- Image signature verification at admission (Cosign + policy webhook)
- API server audit log retention with alerting on anomalous access patterns
- Continuous compliance scanning against CIS Kubernetes Benchmark

---

## Container

### Least Privilege
- Always run containers as non-root users
- Set `securityContext` explicitly: `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true`, drop all capabilities and add back only what is strictly needed

### Defense in Depth
- Use vetted or hardened base images (e.g., Chainguard); never use `latest` tag
- Apply Seccomp and AppArmor profiles to restrict the syscall and kernel surface available to each container
- TLS is a must for all publicly exposed services

### Reduce Attack Surface
- Minimal or distroless images to eliminate unnecessary tooling and reduce CVE surface
- Multi-stage builds to avoid shipping build artifacts in the final image
- Use IRSA or Pod Identity for cloud resource access; do not embed cloud credentials
- Avoid host networking and host volumes unless there is no alternative

### Governance
- Scan images with Trivy (or equivalent) in CI; block on critical/high vulnerabilities
- Maintain an SBOM to track package inventory per image
- Sign images and verify signatures at admission time

---

## Code

### Least Privilege
- Scope service tokens, DB credentials, and API keys to minimum required permissions
- Prefer short-lived dynamically issued credentials over long-lived secrets

### Defense in Depth
- Treat SAST, DAST, and SCA as distinct CI gates — not just "dependency scanning"
- Define quality gates: test coverage thresholds, vulnerability count and severity limits
- For critical systems, consider periodic penetration testing
- Shift left with linters and security plugins during development for faster feedback before changes reach production

### Reduce Attack Surface
- Minimize dependencies; write custom logic only when a dependency adds meaningful risk
- Pin dependencies with lock files and hash verification to prevent dependency confusion attacks

### Minimize Blast Radius
- Keep services focused in responsibility; extract into a separate microservice if a component becomes too broad

### Separation of Duties
- Branch protection rules, required reviews, and signed commits to enforce accountability
- No direct pushes to main; all changes go through a reviewed PR

### Governance
- Secret scanning in CI (e.g., GitLeaks, GitHub secret scanning) to catch leaked credentials before merge
- Track and triage dependency vulnerabilities continuously, not only at build time
---

## Other General Enhancements

- Managed control plane (EKS, GKE, AKS) reduces operational burden on cluster hardening
- Karpenter for advanced autoscaling and timely AMI/node recycling
- ALB Ingress Controller and external-dns for automated, auditable DNS and certificate management
- SSO with proper RBAC on all internal services (Grafana, ArgoCD)
- Proper tolerations, taints, affinity, anti-affinity, and topology-aware scheduling
- ResourceQuota, LimitRanges, and PriorityClasses for predictable resource allocation
