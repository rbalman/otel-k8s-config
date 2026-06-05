# Part 5 — Debugging: Ingress 502/504

Scenario: The service is deployed, pods show `Running`, but returns a `502` or `504`.

- `502` — mostly because of a configuration issue: port mismatch, selector mismatch, app crashing.
- `504` — mostly because app is not responding: hung, slow startup, under-resourced, improper readiness probe.

In general if I see

- `502` status code, I try to look for network components misconfiguration. 
- `504` status code, I try to look at the application status like: current load, resource allocation, app errors, db query or similar heavy operations.

The steps below are a starting point — in practice I jump directly to whichever layer the symptoms point at.

---

## Investigation

**1. Start with controller logs**

Best to look for `connection refused`, `upstream timed out`, or `no live upstreams`. This alone usually points you at the right layer.

---

**2. Check Ingress config**

Best to check ingress config like service name, port especially when `502` status code is seen

---

**3. Check the Service has endpoints**

`Running` != `Ready` 
Running pod doesn't mean the pod is ready to serve the traffic.

- If the readiness probe is failing, Kubernetes pulls the pod out of the endpoint slice and nginx has nowhere to send traffic.
- If there is no endpoints, either the readiness probe is failing or the Service selector doesn't match the pod labels.

---

**4. Look at why the pod isn't Ready**

Check `Conditions`, readiness probe failures show up there with the exact error.

---

**5. Directly access service**

If this works but nginx still 502s, the problem is in the Ingress/Service config, not the app.

---

**6. Verify port alignment**

A common silent failure is also port mismatch either on ingress to service or service to container. Better to verify this.

**7. Fixing the root cause**

If there is no config issue, the `504` has an application-level cause. Distributed tracing is particularly useful here — span durations quickly reveal whether the timeout originates in the app, a downstream service, or a slow query:

- Horizontally or vertically scale the application if it is resource-constrained
- Scale or optimize dependent components (API, DB) if they are the bottleneck
- Fix inefficient or buggy code if the slowness is self-contained
- Check and tune ingress timeout annotations (`proxy-read-timeout`, `proxy-send-timeout`) — the app may be responding correctly but slower than the default timeout allows
