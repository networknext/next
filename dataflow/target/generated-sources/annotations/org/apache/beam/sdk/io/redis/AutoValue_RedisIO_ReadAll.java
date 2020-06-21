package org.apache.beam.sdk.io.redis;

import javax.annotation.Generated;
import javax.annotation.Nullable;

@Generated("com.google.auto.value.processor.AutoValueProcessor")
final class AutoValue_RedisIO_ReadAll extends RedisIO.ReadAll {

  private final RedisConnectionConfiguration connectionConfiguration;

  private final int batchSize;

  private AutoValue_RedisIO_ReadAll(
      @Nullable RedisConnectionConfiguration connectionConfiguration,
      int batchSize) {
    this.connectionConfiguration = connectionConfiguration;
    this.batchSize = batchSize;
  }

  @Nullable
  @Override
  RedisConnectionConfiguration connectionConfiguration() {
    return connectionConfiguration;
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
    if (o instanceof RedisIO.ReadAll) {
      RedisIO.ReadAll that = (RedisIO.ReadAll) o;
      return (this.connectionConfiguration == null ? that.connectionConfiguration() == null : this.connectionConfiguration.equals(that.connectionConfiguration()))
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
    h$ ^= batchSize;
    return h$;
  }

  @Override
  RedisIO.ReadAll.Builder builder() {
    return new Builder(this);
  }

  static final class Builder extends RedisIO.ReadAll.Builder {
    private RedisConnectionConfiguration connectionConfiguration;
    private Integer batchSize;
    Builder() {
    }
    private Builder(RedisIO.ReadAll source) {
      this.connectionConfiguration = source.connectionConfiguration();
      this.batchSize = source.batchSize();
    }
    @Override
    RedisIO.ReadAll.Builder setConnectionConfiguration(RedisConnectionConfiguration connectionConfiguration) {
      this.connectionConfiguration = connectionConfiguration;
      return this;
    }
    @Override
    RedisIO.ReadAll.Builder setBatchSize(int batchSize) {
      this.batchSize = batchSize;
      return this;
    }
    @Override
    RedisIO.ReadAll build() {
      String missing = "";
      if (this.batchSize == null) {
        missing += " batchSize";
      }
      if (!missing.isEmpty()) {
        throw new IllegalStateException("Missing required properties:" + missing);
      }
      return new AutoValue_RedisIO_ReadAll(
          this.connectionConfiguration,
          this.batchSize);
    }
  }

}
