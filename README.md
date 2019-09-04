# Vault Secret Broker

A secret broker is an interface between Hashicorp Vault and a CI/CD process that
requires access to secrets stored in Vault. Instead of directly handing out
vault credentials to CI/CD servers, the secret broker adds another layer of
protection by enforcing that secrets will only be handed out to actually
running jobs.

For this it uses the API of the CI/CD system to check if the job requesting
a secret is actually running. It then uses the Vault API to check if the
job is actually permitted to access the secret by checking secret and / or
entity metadata and returns a wrapped secret to the CI/CD job.

The advantage of this approach is that neither the broker nor the CI/CD server
needs full access to all secrets potential CI/CD jobs will need in the future.
Secrets are only accessible during the lifetime of a CI/CD job and even then
only to the job itself.

Thanks to the wrapping mechanic of Vault, the broker does not have access
to the actual secret and a CI/CD job only has access to a secret the broker
returns. If the secret is unwrapped somewhere except in the CI/CD job,
the job fails while trying to access the secret, thus immediately exposing a
potential leak.

## CI/CD support

The current focus is on Gitlab-CI and later on Concourse and Bamboo.
