# Probe fixtures

Golden-file HTTP fixtures for the per-probe behavioral test harness in
[`pkg/scanner/fixture_test.go`](../../fixture_test.go).

Each probe in [`probes/`](../../../../probes) has a companion `<probe>.yaml`
here. For every case, the harness serves the canned responses from an
`httptest.Server` and runs the **real** probe (loaded from the embedded probe
library) through the **real** scanner, then asserts the probe matched iff the
case said it should. This catches match-rule regressions that `julius validate`
(schema-only) cannot: a changed status code, a renamed body marker, a flipped
negation, or a broken `require: all`/`require: any` chain.

## Format

```yaml
probe: ollama          # must equal the probe's `name` (and the filename stem)
cases:
  - name: positive     # unique, descriptive case name
    match: true        # expected overall probe result against these responses
    responses:
      # keyed by the request path exactly as the probe specifies it, including
      # any ?query string (e.g. "/openai/models?api-version=2024-10-21").
      "/api/tags":
        status: 200
        headers:
          Content-Type: application/json
        body: '{"models":[...]}'
  - name: negative-unrelated
    match: false
    responses:
      "/":
        status: 404
        body: "not found"
```

Notes:

- **Path keys** match the probe's `path` verbatim (path + `?query`), with a
  fallback to the bare path. Any path a probe requests but the fixture does not
  define returns `404` with an empty body — this is what makes `require: all`
  chains and negatives fail naturally.
- **`status: 0`** (or omitted) is served as `200`.
- **Empty header values won't match** a `header.contains` rule: that rule means
  "header present with a non-empty value", so give such headers a real value
  (e.g. `Server: uvicorn`).
- Bodies are matched by substring (`body.contains`/`body.prefix`), so they need
  the required markers present and any negated markers absent — exact JSON
  validity is not required, but realistic bodies make the fixtures better docs
  and better regression guards.

## Coverage requirement

`TestEveryProbeHasFixture` fails the build if any shipped probe lacks a fixture
with **at least one positive and one negative case**. When you add a probe, add
its fixture here.

## Refinement status

Every probe has a passing fixture. A subset are hand-refined with realistic
service responses and near-miss negatives that specifically guard a probe's key
discriminator (e.g. `ollama` guards the KoboldCpp carve-out; `vllm` guards the
`uvicorn` Server header; `groq`/`omniroute` guard their infra headers;
`vertex-ai` guards against other Google APIs). The rest are functional scaffolds
(rule-satisfying positive, generic non-matching negative) — correct and
regression-catching, but worth enriching toward real captured responses over
time.
