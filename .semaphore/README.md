# Semaphore

Semaphore is our CI/CD tool. It is responsible for:

1. Running unit tests for Pull Requests.
2. Building and publishing artifacts to Google Cloud environments for deployments.

To understand Semaphore's YAML syntax, refer to their [documentation](https://docs.semaphoreci.com/reference/pipeline-yaml-reference/).

## Docker

Semaphore runs unit tests inside a Docker container. This is more efficient than downloading the necessary packages each time.
However, the container size greatly impacts the time it takes to run the unit tests because the Semaphore instance has to pull
the container (see Semaphore's suggestions for optimizations [here](https://docs.semaphoreci.com/ci-cd-environment/custom-ci-cd-environment-with-docker/#optimizing-docker-images-for-fast-cicd)).
We use Semaphore's default container because it has node version 14 already installed and is relatively small.

### Updating Docker Image

Update the Docker image when the following occurs:

1. New package dependencies for the portal are required

To update the Docker image, make sure you have Docker installed and are a collaborator for the `nbopardi/networknext` docker repo.
Then run the following from the root of the portal repo:

1. `docker build -f ./.semaphore/Dockerfile -t nbopardi/networknext:portal .`
2. `docker history nbopardi/networknext:portal`
	- This is to verify the size of each layer of the Docker image.
	- Make sure the total image size is relatively small (under 3 GB).
3. `docker push nbopardi/networknext:portal`

Once complete, Semaphore will now use the latest Docker image.
