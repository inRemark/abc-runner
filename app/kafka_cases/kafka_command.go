package kafkaCases

import (
	"flag"
	"fmt"
	"strings"
)

func KafkaCommand(args []string) {
	flags := flag.NewFlagSet("kafka", flag.ExitOnError)
	config := flags.Bool("config", false, "kafka run from config file")
	broker := flags.String("broker", "127.0.0.1:9092", "Server hostname (default 127.0.0.1:9092)")
	topic := flags.String("topic", "127.0.0.1:9092", "Server hostname (default 127.0.0.1:9092)")
	t := flags.String("t", "produce", "operation produce consumer")
	acks := flags.Int("acks", 0, "0,1,-1")
	groupId := flags.String("g", "groupId", "default groupId")
	n := flags.Int("n", 100000, "Total number of requests (default 100000)")
	c := flags.Int("c", 3, "Number of parallel connections (default 50)")
	d := flags.Int("d", 3, "Data size of SET/GET value in bytes (default 3)")
	err := flags.Parse(args)
	if err != nil {
		return
	}

	if *config {
		// tc := *t
		fmt.Println("Execute using configuration and result in log file...")
		// Start(tc)
		return
	}

	if *broker == "" {
		fmt.Printf("Command need broker address.")
		return
	}

	fmt.Printf("Kafka Info: broker:%s, total:%d, parallel:%d, dataSize:%d, t:%s, groupId:%s, acks:%d\n",
		*broker, *n, *c, *d, *t, *groupId, *acks)

	if strings.ToLower(*t) == "produce" {
		InitWriterCommand(*broker, *topic, *acks)
		DoProduceCase("log", *n, *c, *d)
	}
	if strings.ToLower(*t) == "consume" {
		DoConsumeCase(*broker, *topic, *groupId, *c)
	}
}
