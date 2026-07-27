package messaging

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func TestNewGoChannelPubSub(t *testing.T) {
	logger := watermill.NopLogger{}
	pubSub := NewGoChannelPubSub(logger)
	if pubSub == nil {
		t.Fatal("expected non-nil GoChannel")
	}
	pubSub.Close()
}

func TestGoChannel_PublishSubscribe(t *testing.T) {
	logger := watermill.NopLogger{}
	pubSub := NewGoChannelPubSub(logger)
	defer pubSub.Close()

	topic := "test.topic"
	payload := []byte(`{"key": "value"}`)

	// 订阅
	msgs, err := pubSub.Subscribe(context.Background(), topic)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 发布
	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set("event_name", "test.event")
	err = pubSub.Publish(topic, msg)
	if err != nil {
		t.Fatalf("failed to publish: %v", err)
	}

	// 接收
	select {
	case received := <-msgs:
		if string(received.Payload) != string(payload) {
			t.Errorf("expected payload %q, got %q", payload, received.Payload)
		}
		if received.Metadata.Get("event_name") != "test.event" {
			t.Errorf("expected metadata event_name 'test.event', got %q",
				received.Metadata.Get("event_name"))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestGoChannel_MultipleTopics(t *testing.T) {
	logger := watermill.NopLogger{}
	pubSub := NewGoChannelPubSub(logger)
	defer pubSub.Close()

	topics := []string{"topic.a", "topic.b", "topic.c"}

	// 订阅所有 topic
	subs := make(map[string]<-chan *message.Message)
	for _, topic := range topics {
		msgs, err := pubSub.Subscribe(context.Background(), topic)
		if err != nil {
			t.Fatalf("failed to subscribe to %s: %v", topic, err)
		}
		subs[topic] = msgs
	}

	// 发布到每个 topic
	for _, topic := range topics {
		msg := message.NewMessage(watermill.NewUUID(), []byte(topic))
		if err := pubSub.Publish(topic, msg); err != nil {
			t.Fatalf("failed to publish to %s: %v", topic, err)
		}
	}

	// 从每个 topic 接收
	for topic, ch := range subs {
		select {
		case received := <-ch:
			if string(received.Payload) != topic {
				t.Errorf("expected payload %q, got %q", topic, received.Payload)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for message on %s", topic)
		}
	}
}

func TestGoChannel_ConcurrentPublish(t *testing.T) {
	logger := watermill.NopLogger{}
	pubSub := NewGoChannelPubSub(logger)
	defer pubSub.Close()

	topic := "concurrent.test"
	const msgCount = 50

	msgs, err := pubSub.Subscribe(context.Background(), topic)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 并发发布
	var wg sync.WaitGroup
	for i := 0; i < msgCount; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			msg := message.NewMessage(watermill.NewUUID(), []byte{byte(n)})
			pubSub.Publish(topic, msg)
		}(i)
	}

	// 接收所有消息
	received := make(map[byte]bool)
	for i := 0; i < msgCount; i++ {
		select {
		case msg := <-msgs:
			received[msg.Payload[0]] = true
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for message %d/%d", i, msgCount)
		}
	}

	wg.Wait()

	if len(received) != msgCount {
		t.Errorf("expected %d unique messages, got %d", msgCount, len(received))
	}
}

func TestNewWatermillLogger(t *testing.T) {
	logger := NewWatermillLogger()
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// 验证接口实现
	var _ watermill.LoggerAdapter = logger
}

func TestGoChannel_ImplementsInterfaces(t *testing.T) {
	// 编译期检查类型
	logger := watermill.NopLogger{}
	pubSub := NewGoChannelPubSub(logger)

	var _ message.Publisher = pubSub
	var _ message.Subscriber = pubSub
}