# feather-core

**feather-core** is a blockchain built using Cosmos SDK and Tendermint and created with [Feather CLI]().

## Get started

```sh
feather chain serve
```

`serve` command installs dependencies, builds, initializes, and starts your blockchain in development.

### Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Ignite CLI docs](https://docs.ignite.com).

### Web Frontend

Feather CLI has scaffolded a Vue.js-based web app in the `vue` directory. Run the following commands to install dependencies and start the app:

```sh
cd vue
npm install
npm run serve
```

The frontend app is built using the `@starport/vue` and `@starport/vuex` packages. For details, see the [monorepo for Ignite front-end development](https://github.com/ignite/web).

## Release

To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```sh
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

## Test Coverage

If you want to have a test coverage report when commiting to a branch you must create an account on [codecov](https://docs.codecov.com/docs#getting-started), add the `CODECOV_TOKEN` to the repo and uncomment all lines from `.github/workflows/test-coverage.yml` file.

## Learn more

- [Ignite CLI](https://ignite.com/cli)
- [Ignite Tutorials](https://docs.ignite.com/guide)
- [Ignite CLI docs](https://docs.ignite.com)
- [Cosmos SDK docs](https://docs.cosmos.network)
- [Developer Chat](https://discord.gg/ignite)
