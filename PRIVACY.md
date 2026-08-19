# Privacy Policy

MetriyonAPI is an open-source desktop API development, testing, load-testing and workflow automation client maintained by OrganizzaTech.

## Local-first operation

MetriyonAPI stores its application data locally on the user's computer unless the user explicitly configures an external destination or service.

## Network communication

MetriyonAPI sends network requests only as a consequence of actions and configurations made by the user, such as REST/SOAP requests, OAuth flows, WSDL access, load tests and workflow executions. Request URLs, headers, credentials, payloads and destinations are defined by the user.

OrganizzaTech does not operate a telemetry or analytics backend for MetriyonAPI and the application does not intentionally transmit project/request data to OrganizzaTech.

## Credentials and secrets

API credentials, tokens and other secrets may be stored locally when the user chooses to save them. Users are responsible for protecting their workstation, repository exports and backups.

Do not commit secrets, private certificates, PFX/P12 files, access tokens or production credentials to the public repository.

## Third-party services

When a user configures MetriyonAPI to communicate with a third-party API or identity provider, that third party's privacy terms apply to data sent to it.

## Contact

Project: https://www.organizza.com.br/produtos/metriyonapi

OrganizzaTech: https://www.organizza.com.br

## Update checks

MetriyonAPI can contact the public GitHub Releases API for `ricardow2005/MetriyonAPI` to determine whether a newer official version is available. Automatic checks are limited to at most once every 24 hours. The request contains standard HTTP metadata and the running MetriyonAPI version in the User-Agent; it does not include saved API requests, request/response bodies, credentials, environment variables, workspace contents, or local database data.

When an update is available, MetriyonAPI asks the user before opening the official GitHub Release page. The application does not silently download, replace, or execute updates.
