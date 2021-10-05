ARTIFACT_BUCKET_DEV = gs://development_artifacts
ARTIFACT_BUCKET_STAGING = gs://staging_artifacts
ARTIFACT_BUCKET_PROD = gs://prod_artifacts

PORTAL_DEV_MIG = portal-frontend-mig
PORTAL_STAGING_MIG = portal-frontend-mig
PORTAL_PROD_MIG = portal-frontend-mig

.PHONY: build-portal-artifacts-dev
build-portal-artifacts-dev:
	./deploy/build-artifacts.sh -e dev -b $(ARTIFACT_BUCKET_DEV)

.PHONY: build-portal-artifacts-staging
build-portal-artifacts-staging:
	./deploy/build-artifacts.sh -e staging -b $(ARTIFACT_BUCKET_STAGING)

.PHONY: build-portal-artifacts-prod
build-portal-artifacts-prod:
	./deploy/build-artifacts.sh -e prod -b $(ARTIFACT_BUCKET_PROD)

.PHONY: deploy-portal-dev
deploy-portal-dev:
	./deploy/deploy-portal.sh -b $(ARTIFACT_BUCKET_DEV) -e dev -m $(PORTAL_DEV_MIG)

.PHONY: deploy-portal-staging
deploy-portal-staging:
	./deploy/deploy-portal.sh -b $(ARTIFACT_BUCKET_STAGING) -e staging -m $(PORTAL_STAGING_MIG)

# only use if 100% necessary - this will be linked to semaphore at some point

#.PHONY: deploy-portal-prod
#deploy-portal-prod:
#	./deploy/deploy-portal.sh -b $(ARTIFACT_BUCKET_PROD) -e prod -m $(PORTAL_PROD_MIG)