ARTIFACT_BUCKET = gs://development_artifacts
ARTIFACT_BUCKET_STAGING = gs://staging_artifacts
ARTIFACT_BUCKET_PROD = gs://prod_artifacts

PORTAL_DEV_MIG = portal-site-mig

.PHONY: build-portal-artifacts-local
build-portal-artifacts-local:
	./deploy/build-artifacts.sh -e local -b $(ARTIFACT_BUCKET_PROD) -s portal

.PHONY: build-portal-artifacts-dev
build-portal-artifacts-dev:
	./deploy/build-artifacts.sh -e dev -b $(ARTIFACT_BUCKET) -s portal

.PHONY: build-portal-artifacts-staging
build-portal-artifacts-staging:
	./deploy/build-artifacts.sh -e staging -b $(ARTIFACT_BUCKET_STAGING) -s portal

.PHONY: build-portal-artifacts-prod
build-portal-artifacts-prod:
	./deploy/build-artifacts.sh -e prod -b $(ARTIFACT_BUCKET_PROD) -s portal

.PHONY: deploy-portal-dev
deploy-portal-dev: build-portal-artifacts-dev
	./deploy/deploy-portal.sh -b $(ARTIFACT_BUCKET) -e dev -m $(PORTAL_DEV_MIG)