# Part 5 — Debugging: Ingress 502/504

The service is deployed, pods show `Running`, but returns a `502` or `504`.

- `502` — mostly a configuration issue: port mismatch, selector mismatch, app crashing.
- `504` — mostly the app not responding in time: hung, slow startup, under-resourced, improper readiness probe.

In general if I see `502` status code, I try to look for misconfiguration but if it `504` status code I will try to look at application load, resource allocation, app errors, db query or similar heavy operations.

To systematically analyze these errors, I might follow following steps but also directly jump to any steps based on scenario and judements.

---

## Investigation

**1. Start with controller logs**

Best to look for `connection refused`, `upstream timed out`, or `no live upstreams`. This alone usually points you at the right layer.

---

**2. Check Ingress config**

Best to check ingress config like service name, port especially when `502` status code is seen

---

**3. Check the Service has endpoints**

`Running` != `Ready`. If the readiness probe is failing, Kubernetes pulls the pod out of the endpoint slice and nginx has nowhere to send traffic.


If there is no endpoints, either the readiness probe is failing or the Service selector doesn't match the pod labels.

---

**4. Look at why the pod isn't Ready**

Check `Conditions`, readiness probe failures show up there with the exact error.

---

**5. Directly access service**

If this works but nginx still 502s, the problem is in the Ingress/Service config, not the app.

---

**6. Verify port alignment**

A common silent failure is also port mismatch either on ingress to service or service to container. Better to verify this.
