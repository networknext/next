# How to Deploy

To deploy Dataflow pipelines, add your pipeline to `pipelines.json` and then use `next dataflow` to view and deploy pipelines.

**IMPORTANT:** Do not perform one-off deployments of Dataflow pipelines from templates or manually on the command line. It makes it difficult (if not impossible) to restore pipelines in case of disaster recovery.

**IMPORTANT:** Keep the pipelines running in the same project that the Pub/Sub topics/subscriptions are in, and make sure they output to BigQuery in the same project. With v3, each environment is in it's own isolated project now, so there's no need for separate projects for bits and pieces; you can use the environment's project for storage and execution.
