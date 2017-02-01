package app

type EnvSubsciption struct {
	QueueName string
	Raw       bool
}

type EnvTopic struct {
	Name          string
	Subscriptions []EnvSubsciption
}

type EnvQueue struct {
	Name string
}

type Environment struct {
	Host        string
	Port        string
	SqsPort     string
	SnsPort     string
	Region      string
	LogMessages bool
	LogFile     string
	Topics      []EnvTopic
	Queues      []EnvQueue
}

var Envs map[string]Environment
