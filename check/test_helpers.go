package check

import (
	"errors"
	"math/rand"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/optiopay/kafka"
	"github.com/optiopay/kafka/kafkatest"
	"github.com/optiopay/kafka/proto"
)

func newTestCheck() *HealthCheck {
	config := HealthCheckConfig{
		MessageLength:    100,
		CheckInterval:    1 * time.Millisecond,
		retryInterval:    1 * time.Millisecond,
		CheckTimeout:     5 * time.Millisecond,
		DataWaitInterval: 1 * time.Millisecond,
		NoTopicCreation:  true,
		topicName:        "health-check",
		brokerID:         1,
		statusServerPort: 8000,
	}

	return &HealthCheck{
		config:  config,
		randSrc: rand.NewSource(time.Now().UnixNano()),
	}
}

func mockBroker(check *HealthCheck, ctrl *gomock.Controller) (*MockBrokerConnection, *kafkatest.Broker, *kafkatest.Consumer, kafka.Producer) {
	broker := kafkatest.NewBroker()
	consumer := &kafkatest.Consumer{
		Broker:   broker,
		Messages: make(chan *proto.Message),
		Errors:   make(chan error),
	}
	producer := broker.Producer(kafka.NewProducerConf())
	connection := NewMockBrokerConnection(ctrl)
	check.broker = connection
	check.consumer = consumer
	check.producer = producer

	return connection, broker, consumer, producer
}

func healthyMetadata(topicName string) *proto.MetadataResp {
	return &proto.MetadataResp{
		CorrelationID: int32(1),
		Brokers: []proto.MetadataRespBroker{
			{
				NodeID: int32(2),
				Host:   "10.0.0.5",
				Port:   int32(9092),
			},
			{
				NodeID: int32(1),
				Host:   "localhost",
				Port:   int32(9092),
			},
		},
		Topics: []proto.MetadataRespTopic{
			{
				Name: "some-other-topic",
				Err:  nil,
				Partitions: []proto.MetadataRespPartition{
					{
						ID:       1,
						Err:      nil,
						Leader:   int32(1),
						Replicas: []int32{1},
						Isrs:     []int32{1},
					},
				},
			},
			{
				Name: topicName,
				Err:  nil,
				Partitions: []proto.MetadataRespPartition{
					{
						ID:       2,
						Err:      nil,
						Leader:   int32(1),
						Replicas: []int32{1, 2},
						Isrs:     []int32{1, 2},
					},
				},
			},
		},
	}
}

func outOfSyncMetadata() *proto.MetadataResp {
	return &proto.MetadataResp{
		CorrelationID: int32(1),
		Brokers: []proto.MetadataRespBroker{
			{
				NodeID: int32(2),
				Host:   "10.0.0.5",
				Port:   int32(9092),
			},
			{
				NodeID: int32(1),
				Host:   "localhost",
				Port:   int32(9092),
			},
		},
		Topics: []proto.MetadataRespTopic{
			{
				Name: "some-topic",
				Err:  nil,
				Partitions: []proto.MetadataRespPartition{
					{
						ID:       1,
						Err:      nil,
						Leader:   int32(2),
						Replicas: []int32{2, 1},
						Isrs:     []int32{2},
					},
				},
			},
		},
	}
}

func underReplicatedMetadata() *proto.MetadataResp {
	return &proto.MetadataResp{
		CorrelationID: int32(1),
		Brokers: []proto.MetadataRespBroker{
			{
				NodeID: int32(2),
				Host:   "10.0.0.5",
				Port:   int32(9092),
			},
			{
				NodeID: int32(1),
				Host:   "localhost",
				Port:   int32(9092),
			},
		},
		Topics: []proto.MetadataRespTopic{
			{
				Name: "some-topic",
				Err:  nil,
				Partitions: []proto.MetadataRespPartition{
					{
						ID:       2,
						Err:      nil,
						Leader:   int32(2),
						Replicas: []int32{2},
						Isrs:     []int32{2},
					},
				},
			},
		},
	}
}

func inSyncMetadata() *proto.MetadataResp {
	return &proto.MetadataResp{
		CorrelationID: int32(1),
		Brokers: []proto.MetadataRespBroker{
			{
				NodeID: int32(2),
				Host:   "10.0.0.5",
				Port:   int32(9092),
			},
			{
				NodeID: int32(1),
				Host:   "localhost",
				Port:   int32(9092),
			},
		},
		Topics: []proto.MetadataRespTopic{
			{
				Name: "some-topic",
				Err:  nil,
				Partitions: []proto.MetadataRespPartition{
					{
						ID:       1,
						Err:      nil,
						Leader:   int32(2),
						Replicas: []int32{2, 1},
						Isrs:     []int32{2, 1},
					},
				},
			},
		},
	}
}

func offlineMetadata() *proto.MetadataResp {
	return &proto.MetadataResp{
		CorrelationID: int32(1),
		Brokers: []proto.MetadataRespBroker{
			{
				NodeID: int32(1),
				Host:   "localhost",
				Port:   int32(9092),
			},
		},
		Topics: []proto.MetadataRespTopic{
			{
				Name: "some-topic",
				Err:  nil,
				Partitions: []proto.MetadataRespPartition{
					{
						ID:       1,
						Err:      nil,
						Leader:   int32(2),
						Replicas: []int32{1},
						Isrs:     []int32{},
					},
				},
			},
		},
	}
}

func metadataWithoutBroker() *proto.MetadataResp {
	return &proto.MetadataResp{
		CorrelationID: int32(1),
		Brokers: []proto.MetadataRespBroker{
			{
				NodeID: int32(2),
				Host:   "10.0.0.5",
				Port:   int32(9092),
			},
		},
		Topics: []proto.MetadataRespTopic{
			{
				Name: "some-other-topic",
				Err:  nil,
				Partitions: []proto.MetadataRespPartition{
					{
						ID:       1,
						Err:      nil,
						Leader:   int32(2),
						Replicas: []int32{},
						Isrs:     []int32{},
					},
				},
			},
		},
	}
}

func metadataWithoutTopic() *proto.MetadataResp {
	return &proto.MetadataResp{
		CorrelationID: int32(1),
		Brokers: []proto.MetadataRespBroker{
			{
				NodeID: int32(1),
				Host:   "localhost",
				Port:   int32(9092),
			},
		},
		Topics: []proto.MetadataRespTopic{},
	}
}

func newZkTestCheck(ctrl *gomock.Controller) (check *HealthCheck, zookeeper *MockZkConnection) {
	check = newTestCheck()
	check.config.zookeeperConnect = "localhost:2181"
	zookeeper = NewMockZkConnection(ctrl)
	check.zookeeper = zookeeper
	return
}

func (zookeeper *MockZkConnection) mockSuccessfulPathCreation(path string) {
	zookeeper.EXPECT().Exists(path).Return(false, nil, nil)
	zookeeper.EXPECT().Create(path, gomock.Any(), int32(0), gomock.Any()).Return(path, nil)
}

func (zookeeper *MockZkConnection) mockFailingPathCreation(path string) {
	zookeeper.EXPECT().Exists(path).Return(false, nil, nil)
	zookeeper.EXPECT().Create(path, gomock.Any(), int32(0), gomock.Any()).Return("", errors.New("Test error"))
}

func (zk *MockZkConnection) mockHealthyMetadata(topic string) {
	zk.EXPECT().Connect([]string{"localhost:2181"}, 10*time.Second).Return(nil, nil)
	zk.EXPECT().Children("/brokers/ids").Return([]string{"1", "2"}, nil, nil)
	zk.EXPECT().Children("/brokers/topics").Return([]string{topic, "some-other-topic"}, nil, nil)
	zk.EXPECT().Get("/brokers/topics/"+topic).Return([]byte(`{"version":1,"partitions":{"2":[1, 2]}}`), nil, nil)
	zk.EXPECT().Get("/brokers/topics/some-other-topic").Return([]byte(`{"version":1,"partitions":{"1":[1]}}`), nil, nil)
	zk.EXPECT().Close()
}
