# SignPath Foundation application checklist

This document is a maintainer checklist for applying to the SignPath Foundation open-source code-signing program.

## Project information

- **Project:** MetriyonAPI
- **Repository:** https://github.com/ricardow2005/MetriyonAPI
- **License:** MIT
- **Product page:** https://www.organizza.com.br/produtos/metriyonapi
- **Project website:** https://www.organizza.com.br
- **Maintainer:** Ricardo Alberto Walter (`@ricardow2005`)
- **Project identity:** OrganizzaTech

## Before applying

- [ ] Repository is public.
- [ ] MIT `LICENSE` is present.
- [ ] README describes the product and functionality.
- [ ] `Code signing policy` is visible from the README.
- [ ] Required SignPath acknowledgement is present.
- [ ] Committer/reviewer/approver roles are documented.
- [ ] Privacy policy is public and linked.
- [ ] GitHub account has multi-factor authentication enabled.
- [ ] GitHub Actions build succeeds from a clean checkout.
- [ ] A Windows binary has been released in the form intended to be signed.
- [ ] The released pre-signing binary is clearly marked unsigned.
- [ ] No proprietary OrganizzaTech component is bundled in the binary being submitted for OSS signing.

## First pre-signing release

Before applying, create the first public release in the same Windows executable form that will later be signed. For the initial application this may be an explicitly named unsigned build, for example:

`MetriyonAPI-0.6.2-windows-amd64-unsigned.exe`

The tag workflow creates the GitHub Release as a draft so the maintainer can inspect it before publication.

Suggested commands:

```bash
git tag -a v0.6.2 -m "MetriyonAPI v0.6.2"
git push origin v0.6.2
```

After the workflow succeeds, inspect the generated draft Release and publish it only after verifying the executable and SHA-256 checksum.

## After SignPath approval

Do not store certificates, signing private keys, or SignPath secrets in source files. Configure the approved SignPath/GitHub integration and change the release workflow so the public Windows asset is the signed output. Keep manual signing approval enabled for every official release.
