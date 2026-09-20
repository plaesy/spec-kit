## Summary

What does this PR change and why?

## Checklist

- [ ] Bash and PowerShell script versions updated together (if applicable) — see [CONTRIBUTING.md](../CONTRIBUTING.md)
- [ ] `bash testing/bash/run.sh` passes
- [ ] `pwsh testing/powershell/run.ps1` passes
- [ ] Relevant docs (`docs/`) updated
- [ ] `CHANGELOG.md` updated under "Unreleased"
- [ ] No new `eval`/`Invoke-Expression` on untrusted input; no unverified remote script execution

## Security-sensitive?

If this touches scripts, remote downloads, or credential handling, describe the risk and mitigation here.
