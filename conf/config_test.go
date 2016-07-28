package conf

import (
	"github.com/p4tin/goaws/gosns"
	"github.com/p4tin/goaws/gosqs"
	. "gopkg.in/check.v1"
)

type ConfSuite struct{}

var _ = Suite(&ConfSuite{})

func (s *ConfSuite) TestConfig_NoQueuesOrTopics(c *C) {
	loadedEnv, err := LoadYamlConfig("./mock-data/mock-config.yaml", "NoQueuesOrTopics", "1111", "2222")
	c.Assert(err, IsNil)
	c.Assert(loadedEnv.SQSPort, Equals, "1111")
	c.Assert(loadedEnv.SNSPort, Equals, "2222")
	c.Assert(len(loadedEnv.Queues), Equals, 0)
	c.Assert(len(gosqs.SyncQueues.Queues), Equals, 0)
	c.Assert(len(loadedEnv.Topics), Equals, 0)
	c.Assert(len(gosns.SyncTopics.Topics), Equals, 0)
}

func (s *ConfSuite) TestConfig_CreateQueuesTopicsAndSubscriptions(c *C) {
	loadedEnv, err := LoadYamlConfig("./mock-data/mock-config.yaml", "Local", "", "")
	c.Assert(err, IsNil)
	c.Assert(loadedEnv.SQSPort, Equals, "9324")
	c.Assert(loadedEnv.SNSPort, Equals, "9292")
	c.Assert(len(loadedEnv.Queues), Equals, 3)
	c.Assert(len(gosqs.SyncQueues.Queues), Equals, 5)
	c.Assert(len(loadedEnv.Topics), Equals, 2)
	c.Assert(len(gosns.SyncTopics.Topics), Equals, 2)
}

