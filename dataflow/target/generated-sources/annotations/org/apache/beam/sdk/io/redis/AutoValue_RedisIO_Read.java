package org.apache.beam.sdk.io.redis;

import java.util.Arrays;
import javax.annotation.Generated;
import javax.annotation.Nullable;

@Generated("com.google.auto.value.processor.AutoValueProcessor")
final class AutoValue_RedisIO_Read extends RedisIO.Read {

  private final RedisConnectionConfiguration connectionConfiguration;

  private final byte[] keyPattern;

  private final int batchSize;

  private AutoValue_RedisIO_Read(
      @Nullable RedisConnectionConfiguration connectionConfiguration,
      @Nullable byte[] keyPattern,
      int batchSize) {
    this.connectionConfiguration = connectionConfiguration;
    this.keyPattern = keyPattern;
    this.batchSize = batchSize;
  }

  @Nullable
  @Override
  RedisConnectionConfiguration connectionConfiguration() {
    return connectionConfiguration;
  }

  @Nullable
  @Override
  byte[] keyPattern() {
    return keyPattern;
  }

  @Override
  int batchSize() {
    return batchSize;
  }

  @Override
  public boolean equals(Object o) {
    if (o == this) {
      return true;
    }
    if (o instanceof RedisIO.Read) {
      RedisIO.Read that = (RedisIO.Read) o;
      return (this.connectionConfiguration == null ? that.connectionConfiguration() == null : this.connectionConfiguration.equals(that.connectionConfiguration()))
          && Arrays.equals(this.keyPattern, (that instanceof AutoValue_RedisIO_Read) ? ((AutoValue_RedisIO_Read) that).keyPattern : that.keyPattern())
          && this.batchSize == that.batchSize();
    }
    return false;
  }

  @Override
  public int hashCode() {
    int h$ = 1;
    h$ *= 1000003;
    h$ ^= (connectionConfiguration == null) ? 0 : connectionConfiguration.hashCode();
    h$ *= 1000003;
    h$ ^= Arrays.hashCode(keyPattern);
    h$ *= 1000003;
    h$ ^= batchSize;
    return h$;
  }

  @Override
  RedisIO.Read.Builder builder() {
    return new Builder(this);
  }

  static final class Builder extends RedisIO.Read.Builder {
    private RedisConnectionConfiguration connectionConfiguration;
    private byte[] keyPattern;
    private Integer batchSize;
    Builder() {
    }
    private Builder(RedisIO.Read source) {
      this.connectionConfiguration = source.connectionConfiguration();
      this.keyPattern = source.keyPattern();
      this.batchSize = source.batchSize();
    }
    @Override
    RedisIO.Read.Builder setConnectionConfiguration(RedisConnectionConfiguration connectionConfiguration) {
      this.connectionConfiguration = connectionConfiguration;
      return this;
    }
    @Override
    RedisIO.Read.Builder setKeyPattern(byte[] keyPattern) {
      this.keyPattern = keyPattern;
      return this;
    }
    @Override
    RedisIO.Read.Builder setBatchSize(int batchSize) {
      this.batchSize = batchSize;
      return this;
    }
    @Override
    RedisIO.Read build() {
      String missing = "";
      if (this.batchSize == null) {
        missing += " batchSize";
      }
      if (!missing.isEmpty()) {
        throw new IllegalStateException("Missing required properties:" + missing);
      }
      return new AutoValue_RedisIO_Read(
          this.connectionConfiguration,
          this.keyPattern,
          this.batchSize);
    }
  }

}
