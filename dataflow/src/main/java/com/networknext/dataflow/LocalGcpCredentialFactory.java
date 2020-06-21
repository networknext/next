package com.networknext.dataflow;

import com.google.auth.Credentials;
import com.google.auth.oauth2.GoogleCredentials;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.util.Arrays;
import java.util.List;

import org.apache.beam.sdk.extensions.gcp.auth.CredentialFactory;
import org.apache.beam.sdk.options.PipelineOptions;

/**
 * Construct an oauth credential to be used by the SDK and the SDK workers.
 * Returns a GCP credential.
 */
public class LocalGcpCredentialFactory implements CredentialFactory {
    private static final List<String> SCOPES = Arrays.asList("https://www.googleapis.com/auth/cloud-platform",
            "https://www.googleapis.com/auth/devstorage.full_control", "https://www.googleapis.com/auth/userinfo.email",
            "https://www.googleapis.com/auth/datastore", "https://www.googleapis.com/auth/pubsub");

    private static final LocalGcpCredentialFactory INSTANCE = new LocalGcpCredentialFactory();

    public static LocalGcpCredentialFactory fromOptions(PipelineOptions options) {
        return INSTANCE;
    }

    /** Returns a default GCP {@link Credentials} or null when it fails. */
    @Override
    public Credentials getCredential() throws IOException {
        return GoogleCredentials.fromStream(new FileInputStream(new File("/network-next-local.json")))
                .createScoped(SCOPES);
    }
}