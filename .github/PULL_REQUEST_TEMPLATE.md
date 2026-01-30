## Description

Brief description of the changes in this PR.

## Type of Change

- [ ] Bug fix (non-breaking change fixing an issue)
- [ ] New feature (non-breaking change adding functionality)
- [ ] New probe (adds detection for a new LLM service)
- [ ] Breaking change (fix or feature causing existing functionality to change)
- [ ] Documentation update

## Changes Made

- Change 1
- Change 2
- Change 3

## Testing

Describe the testing you have done:

- [ ] go test ./... passes
- [ ] julius validate ./probes passes
- [ ] Tested against live service (for new probes)
- [ ] Tested for false positives against other services (for new probes)

## Checklist

- [ ] My code follows the project style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have added tests that prove my fix/feature works
- [ ] New and existing tests pass locally
- [ ] I have updated documentation as needed

## For New Probes

If this PR adds a new LLM service probe:

- [ ] Probe validates successfully
- [ ] Tested against a live instance of the service
- [ ] Tested against other LLM services to confirm no false positives
- [ ] Added appropriate confidence levels
- [ ] Included model extraction if the service supports it

## Related Issues

Fixes #(issue number)
