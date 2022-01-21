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

### 选项

```go
package final

import "time"

type Options struct {
	PurgeOnStartup bool // 启动Bus时是否清除遗留的消息，包含（mq遗留的消息，和本地消息表遗留的消息）

	// RetryCount retry count of message processing failed
	// Does not include the first attempt
	RetryCount uint
	// RetryCount retry interval of message processing failed
	RetryInterval time.Duration

	// outbox opt
	OutboxScanInterval time.Duration // 扫描outbox没有收到ack的消息间隔
	OutboxScanOffset   int64         // 扫描outbox没有收到ack的消息

	NumSubscriber int // subscriber number
	NumAcker      int // acker number
}

// DefaultOptions bus 默认配置
func DefaultOptions() Options {
	return Options{
		PurgeOnStartup:     false,
		NumSubscriber:      5,
		RetryCount:         3,
		RetryInterval:      10 * time.Millisecond,
		NumAcker:           5,
		OutboxScanOffset:   500,
		OutboxScanInterval: 1 * time.Minute,
	}
}

// WithRetryCount sets the retry count of message processing failed
// Does not include the first attempt
// The default value of RetryCount is 3.
func (opt Options) WithRetryCount(val uint) Options {
	opt.RetryCount = val
	return opt
}

// WithRetryInterval sets the retry interval of message processing failed
// The default value of RetryInterval is 10 millisecond.
func (opt Options) WithRetryInterval(val time.Duration) Options {
	opt.RetryInterval = val
	return opt
}

// WithNumSubscriber sets the number of subscriber
// Each subscriber runs in an independent goroutine
// The default value of NumSubscriber is 5.
func (opt Options) WithNumSubscriber(val int) Options {
	opt.NumSubscriber = val
	return opt
}

// WithNumAcker sets the number of acker
// Each acker runs in an independent goroutine
// The default value of NumAcker is 5.
func (opt Options) WithNumAcker(val int) Options {
	opt.NumAcker = val
	return opt
}

// WithOutboxScanInterval 设置扫描outbox没有收到ack的消息间隔
// The default value of OutboxScanInterval is 1 minute.
func (opt Options) WithOutboxScanInterval(val time.Duration) Options {
	opt.OutboxScanInterval = val
	return opt
}

// WithOutboxScanOffset  设置扫描outbox没有收到ack的消息
// The default value of OutboxScanOffset is 500.
func (opt Options) WithOutboxScanOffset(val int64) Options {
	opt.OutboxScanOffset = val
	return opt
}

// WithPurgeOnStartup  设置启动Bus时是否清除遗留的消息
// 包含（mq遗留的消息，和本地消息表遗留的消息）
// The default value of PurgeOnStartup is false.
func (opt Options) WithPurgeOnStartup(val bool) Options {
	opt.PurgeOnStartup = val
	return opt
}

```

## 订阅

```go
bus.Subscribe("topic1").Middleware(common.Middleware1, common.Middleware2).Handler(common.EchoHandler)
```

`common.Middleware`

```go
func Middleware1(c *final.Context) error {
	fmt.Println("Middleware1 before")
	err := c.Next()
	if err != nil {
		return err
	}
	fmt.Println("Middleware1 after")
	return nil
}

func Middleware2(c *final.Context) error {
	fmt.Println("Middleware2 before")
	err := c.Next()
	if err != nil {
		return err
	}
	fmt.Println("Middleware2 after")
	return nil
}
```

`common.EchoHandler`

```go
func EchoHandler(c *final.Context) error {
	msg := &GeneralMessage{}
	err := msgpack.Unmarshal(c.Message.Payload, msg)
	if err != nil {
		return err
	}
	fmt.Println(msg)
	return nil
}
```

## 普通发布

```go
msg := common.GeneralMessage{Type: "simple message", Count: 100}
msgBytes, _ := msgpack.Marshal(msg)
err := bus.Publish("topic1", msgBytes, message.WithConfirm(true))
if err != nil {
  panic(err)
}
```

### 


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

