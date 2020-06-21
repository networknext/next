package com.networknext.dataflow.util.pubsub;

import com.google.pubsub.v1.ProjectSubscriptionName;
import com.google.pubsub.v1.ProjectTopicName;
import com.google.pubsub.v1.PushConfig;

import org.apache.beam.sdk.io.gcp.pubsub.PubsubOptions;

import com.google.api.gax.core.CredentialsProvider;
import com.google.api.gax.core.NoCredentialsProvider;
import com.google.api.gax.grpc.GrpcTransportChannel;
import com.google.api.gax.rpc.ApiException;
import com.google.api.gax.rpc.FixedTransportChannelProvider;
import com.google.api.gax.rpc.TransportChannelProvider;
import com.google.cloud.pubsub.v1.SubscriptionAdminClient;
import com.google.cloud.pubsub.v1.SubscriptionAdminSettings;
import com.google.cloud.pubsub.v1.TopicAdminClient;
import com.google.cloud.pubsub.v1.TopicAdminSettings;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

public class PubSubSubscriptionAutoCreate {
    private static final Logger LOG = LoggerFactory.getLogger(PubSubSubscriptionAutoCreate.class);

    public static void execute(PubsubOptions options, String topicNameStr, String subscriptionNameStr)
            throws Exception {

        LOG.info("Checking if we need to create the Pub/Sub topic...");

        ProjectSubscriptionName subscriptionName = ProjectSubscriptionName.parse(subscriptionNameStr);

        SubscriptionAdminSettings subClientSettings = SubscriptionAdminSettings.newBuilder().build();
        TopicAdminSettings topicClientSettings = TopicAdminSettings.newBuilder().build();

        String pubsubEmulatorHost = System.getenv("PUBSUB_EMULATOR_HOST");
        if (pubsubEmulatorHost != null) {
            ManagedChannel channel = ManagedChannelBuilder.forTarget(pubsubEmulatorHost).usePlaintext().build();
            TransportChannelProvider channelProvider = FixedTransportChannelProvider
                    .create(GrpcTransportChannel.create(channel));
            CredentialsProvider credentialsProvider = NoCredentialsProvider.create();
            subClientSettings = SubscriptionAdminSettings.newBuilder().setTransportChannelProvider(channelProvider)
                    .setCredentialsProvider(credentialsProvider).build();
            topicClientSettings = TopicAdminSettings.newBuilder().setTransportChannelProvider(channelProvider)
                    .setCredentialsProvider(credentialsProvider).build();

            // also set the options for the PubsubIO
            PubsubOptions pubsubOptions = options.as(PubsubOptions.class);
            pubsubOptions.setPubsubRootUrl("http://" + pubsubEmulatorHost);
        }

        // create topic
        try (TopicAdminClient topicAdminClient = TopicAdminClient.create(topicClientSettings)) {
            LOG.info("creating topic " + topicNameStr);
            topicAdminClient.createTopic(ProjectTopicName.parse(topicNameStr));
        } catch (ApiException e) {
            if (e.toString().contains("ALREADY_EXISTS")) {
                // ALREADY_EXISTS, we're all good
                LOG.info("Pub/Sub topic already exists.");
            } else {
                throw e;
            }
        }

        LOG.info("Checking if we need to create the Pub/Sub subscription...");

        // create subscription
        try (SubscriptionAdminClient subscriptionAdminClient = SubscriptionAdminClient.create(subClientSettings)) {
            LOG.info("creating subscription for topic " + topicNameStr);
            subscriptionAdminClient.createSubscription(subscriptionName, ProjectTopicName.parse(topicNameStr),
                    PushConfig.newBuilder().build(), 600);
            LOG.info("Pub/Sub subscription created.");
        } catch (ApiException e) {
            if (e.toString().contains("ALREADY_EXISTS")) {
                // ALREADY_EXISTS, we're all good
                LOG.info("Pub/Sub subscription already exists.");
            } else {
                throw e;
            }
        }
    }
}