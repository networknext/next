package main

import (
    "github.com/go-redis/redis/v8"
    "github.com/networknext/backend/modules/common"
    "github.com/networknext/backend/modules/core"
    "github.com/networknext/backend/modules/envvar"
    "time"
)

func main() {

    service := common.CreateService("redis_monitor")

    monitorRedis(service)

    service.StartWebServer()

    service.WaitForShutdown()
}

func monitorRedis(service *common.Service) {

    redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
    redisPassword := envvar.GetString("REDIS_PASSWORD", "")

    core.Log("redis hostname: %s", redisHostname)
    core.Log("redis password: %s", redisPassword)

    redisClient := redis.NewClient(&redis.Options{
        Addr:     redisHostname,
        Password: redisPassword,
    })

    go func() {
        for {
            _, err := redisClient.Ping(service.Context).Result()
            if err == nil {
                core.Log("redis is OK")
            } else {
                core.Error("redis is not OK: %v", err)
            }
            time.Sleep(time.Second)
        }
    }()
}
