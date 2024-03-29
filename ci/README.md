# bosh-gcscli

In order to run the bosh-gcscli Concourse Pipeline you must have an existing [Concourse](http://concourse.ci) environment. See [Deploying Concourse on Google Compute Engine](https://github.com/cloudfoundry-incubator/bosh-google-cpi-release/blob/master/docs/deploy_concourse.md) for instructions.

* Target your Concourse CI environment:

```
fly -t google login -c <YOUR CONCOURSE URL>
```

* Update the [credentials.yml](https://github.com/cloudfoundry/bosh-gcscli/blob/main/ci/credentials.yml.tpl) file. Note that this configuration file requires a JSON Service Account File for a service account with at least Editor access to the project. To get a Service Account File, see [here](https://developers.google.com/identity/protocols/OAuth2ServiceAccount#creatinganaccount) and create using the Project/Editor role.

* Set the bosh-gcscli pipeline:

```
fly -t google set-pipeline -p bosh-gcscli -c pipeline.yml -l credentials.yml
```

* Unpause the bosh-gcscli pipeline:

```
fly -t google unpause-pipeline -p bosh-gcscli
```

## Development pipeline

`pipeline-develop.yml` contains the functional tests without the release
handling. For development, it is recommended to use this as it is significantly
easier to setup.

To prepare, follow the above steps with the following exceptions.

* Use `pipeline-develop.yml` instead of `pipeline.yml`

* Change the `bosh-gcscli-src-in` to reference your development repository and
branch.