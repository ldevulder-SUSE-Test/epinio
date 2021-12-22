# Cypress tests

This repo is configured for end-to-end testing with Cypress.

For the cypress test runner to consume the UI, you must specify two environment variables, TEST_USERNAME and TEST_PASSWORD. By default the test runner will attempt to visit a locally running dashboard at https://localhost:8005. This may be overwritten with the DEV_UI environment variable. Run `yarn cypress open` to start the dashboard in SSR mode and open the cypress test runner. Run tests through the cypress GUI once the UI is built. Cypress tests will automatically re-run if they are altered (hot reloading).

That means, Rancher and Epinio instances are needed.</br>
For my test, I'm using one Epinio cluster imported in a Rancher instance with proper DNS name configuration and a letsencrypt certificate.

# Installation

Cypress is not already merged in main, so you have to clone the repo and fetch the branch ui_testing.
```
git clone git@github.com:epinio/epinio.git
cd epinio
git checkout -b ui_testing origin/ui_testing
```

To install Cypress, you need [nvm](https://github.com/nvm-sh/nvm) and yarn.</br>
nvm allows you to quickly install and use different versions of node via the command line.

It's more than recommanded using a node LTS version, to achieve that, use command `nvm install --lts`.

Then, you can install yarn either with `npm install yarn` or `curl -o- -L https://yarnpkg.com/install.sh | bash` (faster way). </br>
Install Cypress and all the dependencies with `npx yarn --pure-lockfile install`

You can open Cypress GUI typing `npx cypress open`.</br>
Do not forget to export the needed env variables we talked about above.
