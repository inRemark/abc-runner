package kafkaCases

import (
	"sync"
	"testing"
)

func TestConsumerTest(t *testing.T) {

	url, topic, groupId, c := "192.168.0.62:9092", "my_topic", "groupId", 1

	var wg sync.WaitGroup
	for i := 1; i <= c; i++ {
		reader := InitReaderCommand(url, topic, groupId)
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			Consume(reader, i)
		}(i)
	}
	wg.Wait()
}
