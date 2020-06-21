package org.apache.beam.sdk.io.redis;

import javax.annotation.Generated;
import javax.annotation.Nullable;

@Generated("com.google.auto.value.processor.AutoValueProcessor")
final class AutoValue_RedisIO_Write extends RedisIO.Write {

  private final RedisConnectionConfiguration connectionConfiguration;

  private final RedisIO.Write.Method method;

  private final Long expireTime;

  private AutoValue_RedisIO_Write(
      @Nullable RedisConnectionConfiguration connectionConfiguration,
      @Nullable RedisIO.Write.Method method,
      @Nullable Long expireTime) {
    this.connectionConfiguration = connectionConfiguration;
    this.method = method;
    this.expireTime = expireTime;
  }

  @Nullable
  @Override
  RedisConnectionConfiguration connectionConfiguration() {
    return connectionConfiguration;
  }

  @Nullable
  @Override
  RedisIO.Write.Method method() {
    return method;
  }

  @Nullable
  @Override
  Long expireTime() {
    return expireTime;
  }

  @Override
  public boolean equals(Object o) {
    if (o == this) {
      return true;
    }
    if (o instanceof RedisIO.Write) {
      RedisIO.Write that = (RedisIO.Write) o;
      return (this.connectionConfiguration == null ? that.connectionConfiguration() == null : this.connectionConfiguration.equals(that.connectionConfiguration()))
          && (this.method == null ? that.method() == null : this.method.equals(that.method()))
          && (this.expireTime == null ? that.expireTime() == null : this.expireTime.equals(that.expireTime()));
    }
    return false;
  }

  @Override
  public int hashCode() {
    int h$ = 1;
    h$ *= 1000003;
    h$ ^= (connectionConfiguration == null) ? 0 : connectionConfiguration.hashCode();
    h$ *= 1000003;
    h$ ^= (method == null) ? 0 : method.hashCode();
    h$ *= 1000003;
    h$ ^= (expireTime == null) ? 0 : expireTime.hashCode();
    return h$;
  }

  @Override
  RedisIO.Write.Builder builder() {
    return new Builder(this);
  }

  static final class Builder extends RedisIO.Write.Builder {
    private RedisConnectionConfiguration connectionConfiguration;
    private RedisIO.Write.Method method;
    private Long expireTime;
    Builder() {
    }
    private Builder(RedisIO.Write source) {
      this.connectionConfiguration = source.connectionConfiguration();
      this.method = source.method();
      this.expireTime = source.expireTime();
    }
    @Override
    RedisIO.Write.Builder setConnectionConfiguration(RedisConnectionConfiguration connectionConfiguration) {
      this.connectionConfiguration = connectionConfiguration;
      return this;
    }
    @Override
    RedisIO.Write.Builder setMethod(RedisIO.Write.Method method) {
      this.method = method;
      return this;
    }
    @Override
    RedisIO.Write.Builder setExpireTime(Long expireTime) {
      this.expireTime = expireTime;
      return this;
    }
    @Override
    RedisIO.Write build() {
      return new AutoValue_RedisIO_Write(
          this.connectionConfiguration,
          this.method,
          this.expireTime);
    }
  }

}
