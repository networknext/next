package org.apache.beam.sdk.io.redis;

import javax.annotation.Generated;
import javax.annotation.Nullable;

@Generated("com.google.auto.value.processor.AutoValueProcessor")
final class AutoValue_RedisConnectionConfiguration extends RedisConnectionConfiguration {

  private final String host;

  private final int port;

  private final String auth;

  private final int timeout;

  private final boolean ssl;

  private AutoValue_RedisConnectionConfiguration(
      String host,
      int port,
      @Nullable String auth,
      int timeout,
      boolean ssl) {
    this.host = host;
    this.port = port;
    this.auth = auth;
    this.timeout = timeout;
    this.ssl = ssl;
  }

  @Override
  String host() {
    return host;
  }

  @Override
  int port() {
    return port;
  }

  @Nullable
  @Override
  String auth() {
    return auth;
  }

  @Override
  int timeout() {
    return timeout;
  }

  @Override
  boolean ssl() {
    return ssl;
  }

  @Override
  public String toString() {
    return "RedisConnectionConfiguration{"
         + "host=" + host + ", "
         + "port=" + port + ", "
         + "auth=" + auth + ", "
         + "timeout=" + timeout + ", "
         + "ssl=" + ssl
        + "}";
  }

  @Override
  public boolean equals(Object o) {
    if (o == this) {
      return true;
    }
    if (o instanceof RedisConnectionConfiguration) {
      RedisConnectionConfiguration that = (RedisConnectionConfiguration) o;
      return this.host.equals(that.host())
          && this.port == that.port()
          && (this.auth == null ? that.auth() == null : this.auth.equals(that.auth()))
          && this.timeout == that.timeout()
          && this.ssl == that.ssl();
    }
    return false;
  }

  @Override
  public int hashCode() {
    int h$ = 1;
    h$ *= 1000003;
    h$ ^= host.hashCode();
    h$ *= 1000003;
    h$ ^= port;
    h$ *= 1000003;
    h$ ^= (auth == null) ? 0 : auth.hashCode();
    h$ *= 1000003;
    h$ ^= timeout;
    h$ *= 1000003;
    h$ ^= ssl ? 1231 : 1237;
    return h$;
  }

  @Override
  RedisConnectionConfiguration.Builder builder() {
    return new Builder(this);
  }

  static final class Builder extends RedisConnectionConfiguration.Builder {
    private String host;
    private Integer port;
    private String auth;
    private Integer timeout;
    private Boolean ssl;
    Builder() {
    }
    private Builder(RedisConnectionConfiguration source) {
      this.host = source.host();
      this.port = source.port();
      this.auth = source.auth();
      this.timeout = source.timeout();
      this.ssl = source.ssl();
    }
    @Override
    RedisConnectionConfiguration.Builder setHost(String host) {
      if (host == null) {
        throw new NullPointerException("Null host");
      }
      this.host = host;
      return this;
    }
    @Override
    RedisConnectionConfiguration.Builder setPort(int port) {
      this.port = port;
      return this;
    }
    @Override
    RedisConnectionConfiguration.Builder setAuth(String auth) {
      this.auth = auth;
      return this;
    }
    @Override
    RedisConnectionConfiguration.Builder setTimeout(int timeout) {
      this.timeout = timeout;
      return this;
    }
    @Override
    RedisConnectionConfiguration.Builder setSsl(boolean ssl) {
      this.ssl = ssl;
      return this;
    }
    @Override
    RedisConnectionConfiguration build() {
      String missing = "";
      if (this.host == null) {
        missing += " host";
      }
      if (this.port == null) {
        missing += " port";
      }
      if (this.timeout == null) {
        missing += " timeout";
      }
      if (this.ssl == null) {
        missing += " ssl";
      }
      if (!missing.isEmpty()) {
        throw new IllegalStateException("Missing required properties:" + missing);
      }
      return new AutoValue_RedisConnectionConfiguration(
          this.host,
          this.port,
          this.auth,
          this.timeout,
          this.ssl);
    }
  }

}
