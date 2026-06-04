## 4Cs and Security Principles

I like to think production security interms of 4 Layers commonly known as 4Cs `Cloud`, `Cluster`, `Container`, `Code`. For each of layers it is best practice to follow these principles whenever possible.

- least privilege
- defense in depth
- reduce attack surface
- minimize blast radius
- Separation of Duties

---

## Cloud

### Least Privilege
- IAM best practices, read only acess to the cloud envrionment
- Use SSO and RBAC for accessing tools like Arogcd
- Secure Cloud with SSO logins

### Defense in Depth
- CSPM tool or AWS Config tracking tools like 
- CloudTrail logs and events tracking
- Use proper VPC, subnet, NACLs, firewall practices
- Employ WAF, Rate limiting, DDOS protection
- Using VPC is not enough, organizations have complex acess requirements, multiple roles, business needs, internal tools. Use zero trust mesh vpns like tailscale for security, auditing, speed and centralized access managmenet.

### Reduce Attack Surface
- No static IAM credentails should be used
- Use Github OIDC to deploy changes
- Never expose internal toolings in public internet even if they have authentication in place. eg. grafana, prometheus, argocd endpoints should never have public endpoints.
- For mission critical accounts we can even restrict all the IAM access with IP whitesliting which will prevent malacious actor to access AWS APIs even if credentails are leaked.

### Minimize Blast Radius
- Drift detection and reporting
- Segreagate different envs on account level.

### Separation of Duties
- Only allowing CD to manage production account/infra

---

## Cluster

### Least Privilege
- Practice sound RBAC policies

### Defense in Depth
- Logging and monitoring control plane components logs and metrics
- Employ CNIs that provide better network policies and visibility like Cilium

### Reduce Attack Surface
- Use external secrets management services like hashicorp vault, secrets manager, 1password
- Whenever possible use managed/vetted amis for hosting workloads
- Use advance node management/autoscaling tools like karpenter for better recycling of amis/k8s version
- Timely ugprading of k8s version and controllers along with them

### Minimize Blast Radius
- Make use of network policies to limit communication between required components.


---

## Container

### Least Privilege
- Always prefer to run container as non root user
- Employ securityContext for limiting containers privilege

### Defense in Depth
- Use vetted base image like from chainguard if possible
- Never use latest tag for imaegs.

### Reduce Attack Surface
- Use minimal images or distroless image to reduce attack surface and bloatware
- Use multi stage build to remove built generated files, folder and layers

---

## Code

### Defense in Depth
- Introduce Code and Dependency scanning as a part of CI
- Define Quality gate like unit test coverage, vulnerabiilty count/type..
- if critical may be pen testing
- Shift left code scanning by introducing linters, security plugins during development phase for faster feedback and remediation before it lands on production

### Reduce Attack Surface
- Minimizing dependency, write custom logic whenever possible

### Minimize Blast Radius
- Minimze the responsiblity and break as a separate micro-server if too much responsibility.

---

## Improvements

- Use Managed Kubernetes Cluster like EKS
- Use Karpetener for adavance autoscaling
- Use secrets management controllers
- Use ALB Ingress Controller
- external-dns for auto dns management
- I would use SSO on internal services like grafana, argocd and also have proper RBAC on argocd
