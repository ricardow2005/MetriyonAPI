# Publish the repository on GitHub

Canonical repository:

`https://github.com/ricardow2005/metriyonapi`

The repository should be **Public** and created empty (without GitHub-generated README, .gitignore or license), because those files are already included in this project.

From the project root on Windows PowerShell, Git Bash or a terminal with Git installed:

```bash
git init
git branch -M main
git add .
git commit -m "Initial open-source release"
git remote add origin https://github.com/ricardow2005/metriyonapi.git
git push -u origin main
```

## First release

After the repository is pushed and the GitHub Actions `Build` workflow is green, create a release tag when ready:

```bash
git tag v0.6.2
git push origin v0.6.2
```

The `Draft Release` workflow builds the Windows executable, generates SHA-256 checksums and creates a **draft** GitHub Release. It does not publish the release automatically. Review the draft in GitHub and publish it manually when satisfied.

Until SignPath Foundation integration is active, release artifacts are explicitly named `unsigned`.

## SignPath Foundation

After at least one public project release exists, apply for the SignPath Foundation open-source code-signing program using the public repository, MIT license, privacy policy and code-signing policy included here. Once approved, update the release workflow with the SignPath project/organization identifiers supplied for the project. Do not place private signing credentials directly in the repository.
