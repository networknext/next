ARTIFACT_BUCKET = gs://development_artifacts
ARTIFACT_BUCKET_STAGING = gs://staging_artifacts
ARTIFACT_BUCKET_PROD = gs://prod_artifacts

.PHONY: build-portal-artifacts-dev
build-portal-artifacts-dev:
	./deploy/build-artifact.sh -e dev -b $(ARTIFACT_BUCKET) -s portal

.PHONY: build-portal-artifacts-staging
build-portal-artifacts-staging:
	./deploy/build-artifact.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s portal

.PHONY: build-portal-artifacts-prod
build-portal-artifacts-prod:
	./deploy/build-artifact.sh -e staging -b $(ARTIFACT_BUCKET_PROD) -s portal
