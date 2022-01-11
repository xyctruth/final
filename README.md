# Final

## 使用本地消息表实现最终一致性

## Using

### 初始化Bus
```go
db, err := sql.Open("mysql", "root:xLrFGAzed3@tcp(localhost:3306)/final_test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4")
mq, _ := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
bus := final.New("svc_name", db, mq)
```

### 注册消息路由
```go
bus.Topic("topic1").Middleware(example.Middleware1, example.Middleware2).Handler("handler1", Handler1)
bus.Start()
```

### 普通发送

```go
err := bus.Publish("topic1", "handler1", msgBytes)
if err != nil {
    panic(err)
}
```

### 使用confirm ack

```go
err := bus.Publish("topic1", "handler1", msgBytes, message.WithConfirm(true))
if err != nil {
    panic(err)
}
```

### 使用confirm ack 并关联本地事务

```go
gormTX := example.NewDB().Begin()
/* return err rollback，return nil commit */
bus.Transaction(gormTX.Statement.ConnPool.(*sql.Tx), func(txBus *final.TxBus) error {
    err := gormTX.Table("test").Create(&example.Test{Name: "aaa"}).Error
    if err != nil {
    return err
    }
    
    err = txBus.Publish("topic1", "handler1", msgBytes)
    if err != nil {
    return err
    }
    return nil
})
```