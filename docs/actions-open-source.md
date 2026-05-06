# GitHub Actions Open Source Notes

This repository can be public without changing the normal production deploy path.

## What Still Runs

- Pushes to `main` still build and publish the custom GHCR image.
- Pushes to `main` still deploy production from `.github/workflows/publish-custom-image.yml` when the repository is `MonsterQiu/new-api`.
- Manual production release/deploy still works from `.github/workflows/release-and-deploy-prod.yml` when the repository is `MonsterQiu/new-api`.

## Public Repository Guards

- Production deployment jobs are scoped to `MonsterQiu/new-api` so forks or copied public workflows cannot target the production runner.
- Production deployment jobs use the `production` environment. Configure required reviewers in GitHub if manual approval is desired.
- The PR check uses `pull_request_target` only for metadata checks. Do not add checkout steps or execute PR code in that workflow.

## GitHub Settings To Keep

- Set repository Actions default `GITHUB_TOKEN` permissions to read-only.
- If direct push deploys are required, do not require pull requests for your own
  pushes to `main`, or configure branch protection so trusted maintainers can
  bypass PR requirements.
- Keep untrusted contributors on pull requests; do not grant them direct push
  access to `main`.
- Leave the `production` environment without required reviewers if `main` pushes
  must deploy immediately.
- Keep self-hosted production runners restricted to trusted workflows and branches.
- Do not print deployment secrets, private IPs, API keys, or connection strings in workflow logs.
