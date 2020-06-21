package com.networknext.dataflow.util.pubsub;

import com.google.pubsub.v1.ProjectTopicName;

import org.apache.beam.sdk.io.gcp.pubsub.PubsubOptions;

import com.google.api.gax.core.CredentialsProvider;
import com.google.api.gax.core.NoCredentialsProvider;
import com.google.api.gax.grpc.GrpcTransportChannel;
import com.google.api.gax.rpc.ApiException;
import com.google.api.gax.rpc.FixedTransportChannelProvider;
import com.google.api.gax.rpc.TransportChannelProvider;
import com.google.cloud.pubsub.v1.TopicAdminClient;
import com.google.cloud.pubsub.v1.TopicAdminSettings;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;

public class PubSubTopicAutoCreate {
    private static final Logger LOG = LoggerFactory.getLogger(PubSubTopicAutoCreate.class);

    public static void execute(PubsubOptions options, String topicNameStr) throws Exception {

        LOG.info("Checking if we need to create the Pub/Sub topic (" + topicNameStr + ")...");

        ProjectTopicName topicName = ProjectTopicName.parse(topicNameStr);

        TopicAdminSettings topicClientSettings = TopicAdminSettings.newBuilder().build();

        String pubsubEmulatorHost = System.getenv("PUBSUB_EMULATOR_HOST");
        if (pubsubEmulatorHost != null) {
            ManagedChannel channel =
                    ManagedChannelBuilder.forTarget(pubsubEmulatorHost).usePlaintext().build();
            TransportChannelProvider channelProvider =
                    FixedTransportChannelProvider.create(GrpcTransportChannel.create(channel));
            CredentialsProvider credentialsProvider = NoCredentialsProvider.create();
            topicClientSettings =
                    TopicAdminSettings.newBuilder().setTransportChannelProvider(channelProvider)
                            .setCredentialsProvider(credentialsProvider).build();

            // also set the options for the PubsubIO
            if (options != null) {
                PubsubOptions pubsubOptions = options.as(PubsubOptions.class);
                pubsubOptions.setPubsubRootUrl("http://" + pubsubEmulatorHost);
            }
        }

        try (TopicAdminClient topicAdminClient = TopicAdminClient.create(topicClientSettings)) {
            LOG.info("creating topic " + topicName);
            topicAdminClient.createTopic(topicName);
            LOG.info("Pub/Sub topic created.");
        } catch (ApiException e) {
            if (e.toString().contains("ALREADY_EXISTS")) {
                // ALREADY_EXISTS, we're all good
                LOG.info("Pub/Sub topic already exists.");
            } else {
                throw e;
            }
        }
    }
}
