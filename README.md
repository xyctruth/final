# Final

[![Go Report Card](https://goreportcard.com/badge/github.com/xyctruth/final)](https://goreportcard.com/report/github.com/xyctruth/final)
[![codecov](https://codecov.io/gh/xyctruth/final/branch/main/graph/badge.svg?token=YWNYJK9KQW)](https://codecov.io/gh/xyctruth/final)
[![Build status](https://img.shields.io/github/workflow/status/xyctruth/final/Build/main)](https://github.com/xyctruth/final/actions/workflows/build.yml)
[![LICENSE status](https://img.shields.io/github/license/xyctruth/final)](https://github.com/xyctruth/final/blob/main/LICENSE)

# 简介

**Final 使用本地消息表实现最终一致性**

# 使用

## 初始化

```go
db, err := sql.Open("mysql", mysqlConnStr)
if err != nil {
  panic(err)
}

mqProvider := amqp.NewProvider(amqpConnStr)

bus := final.New("send_svc", db, mqProvider, final.DefaultOptions())
bus.Start()

```

更多的选项配置在 [options.go](./options.go)


## 订阅

```go
bus.Subscribe("topic1").Middleware(common.Middleware1, common.Middleware2).Handler(common.EchoHandler)
```
`common.Middleware1`,`common.Middleware2`,`common.EchoHandler` 的代码在 [common.go](_example/common/common.go)


## 发布

```go
msg := common.GeneralMessage{Type: "simple message", Count: 100}
msgBytes, _ := msgpack.Marshal(msg)
err := bus.Publish("topic1", msgBytes, message.WithConfirm(true))
if err != nil {
  panic(err)
}
```

更多消息发布策略在 [message_policy.go](./message/message_policy.go)

## 关联本地事务发布

### database/sql

```go
tx, _ := _example.NewDB().Begin()

/* return err rollback，return nil commit */
err := bus.Transaction(tx, func(txBus *final.TxBus) error {
  // 本地业务
  _, err := tx.Exec("INSERT INTO local_business (remark) VALUE (?)", "sql local business")
  if err != nil {
    return err
  }
  
  // 发布消息
  msg := common.GeneralMessage{Type: "sql transaction message", Count: 100}
	msgBytes, _ := msgpack.Marshal(msg)
  
  err = txBus.Publish("topic1", msgBytes)
  if err != nil {
    return err
  }
  return nil
})

if err != nil {
  panic(err)
}
```

### gorm

```go
tx := _example.NewGormDB().Begin()

/* return err rollback，return nil commit */
err = bus.Transaction(tx.Statement.ConnPool.(*sql.Tx), func(txBus *final.TxBus) error {
   // 本地业务
  result := tx.Create(&_example.LocalBusiness{
    Remark: "gorm local business",
  })
  if result.Error != nil {
    return result.Error
  }

   // 发送消息
  msg := common.GeneralMessage{Type: "gorm transaction message", Count: 100}
  msgBytes, _ := msgpack.Marshal(msg)
  err = txBus.Publish("topic1", msgBytes)
  if err != nil {
    return err
  }
  return nil
})

if err != nil {
  panic(err)
}
```

