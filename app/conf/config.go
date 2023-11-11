package conf

import (
    "encoding/json"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "time"

    log "github.com/sirupsen/logrus"

    "github.com/Admiral-Piett/goaws/app"
    "github.com/Admiral-Piett/goaws/app/common"
    "github.com/fsnotify/fsnotify"
    "github.com/ghodss/yaml"
)

type ConfigLoader struct {
    configFilename string
    envTitle       string
    env            app.Environment

    createQueues func(queues []app.EnvQueue) error
    createTopics func(topics []app.EnvTopic) error

    Ports []string
}

func NewConfigLoader(configFilename, envTitle string) *ConfigLoader {
    loader := &ConfigLoader{
        envTitle:       envTitle,
        configFilename: configFilename,

        createQueues: createSQSQueues,
        createTopics: createSNSTopics,
    }
    if loader.configFilename == "" {
        loader.findDefaultConfig()
    }
    loader.loadYamlConfig(envTitle)
    return loader
}

func (c *ConfigLoader) findDefaultConfig() {
    root, _ := filepath.Abs(".")
    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if "goaws.yaml" == d.Name() {
            c.configFilename = path
        }
        return nil
    })
    if err != nil || c.configFilename == "" {
        log.Warn("Failure to find default config file")
        return
    }
}

func (c *ConfigLoader) loadYamlConfig(envTitle string) {
    c.Ports = []string{"4100"}

    log.Infof("Loading config file: %s", c.configFilename)
    yamlFile, err := os.ReadFile(c.configFilename)
    if err != nil {
        return
    }

    var envs map[string]app.Environment
    err = yaml.Unmarshal(yamlFile, &envs)
    if err != nil {
        log.Errorf("err: %v\n", err)
        return
    }
    if envTitle == "" {
        envTitle = "Local"
    }

    c.env = envs[envTitle]

    if c.env.Region == "" {
        c.env.Region = "local"
    }

    if c.env.Port != "" {
        c.Ports = []string{c.env.Port}
    } else if c.env.SqsPort != "" && c.env.SnsPort != "" {
        c.Ports = []string{c.env.SqsPort, c.env.SnsPort}
        c.env.Port = c.env.SqsPort
    }

    if c.env.LogToFile == true {
        if c.env.LogFile != "" {
            c.env.LogFile = "./goaws_messages.log"
        }
    }

    if c.env.QueueAttributeDefaults.VisibilityTimeout == 0 {
        c.env.QueueAttributeDefaults.VisibilityTimeout = 30
    }

    if c.env.QueueAttributeDefaults.MaximumMessageSize == 0 {
        c.env.QueueAttributeDefaults.MaximumMessageSize = 262144 // 256K
    }

    if c.env.AccountID == "" {
        // QUESTION - @Admiral-Piett - Why `queue` why not `account-id` or something?
        c.env.AccountID = "queue"
    }

    if c.env.Host == "" {
        c.env.Host = "localhost"
        c.env.Port = "4100"
    }

    // TODO - @Admiral-Piett - make app.CurrentEnvironment a pointer to ConfigLoader.env
    app.CurrentEnvironment = c.env

    err = c.createQueues(c.env.Queues)
    if err != nil {
        return
    }

    err = c.createTopics(c.env.Topics)
    if err != nil {
        return
    }
}

func (c *ConfigLoader) StartWatcher(wg *sync.WaitGroup) {
    // FIXME - Commenting this until we hear from @apavanello
    //quit := make(chan struct{})
    log.Infof("Starting watcher on file: %v", c.configFilename)

    // Start listening for events.
    go func() {
        watcher, err := fsnotify.NewWatcher()
        if err != nil {
            log.Errorf("Hot reload watcher failed to start: %s", err)
        }
        err = watcher.Add(c.configFilename)
        if err != nil {
            log.Errorf("Could not add %s to hot reload watcher: %s", c.configFilename, err)
        }
        wg.Done()
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                // FIXME - Commenting this until we hear from @apavanello
                //  - https://github.com/Admiral-Piett/goaws/pull/261#discussion_r1391924987
                //if event.Has(fsnotify.Remove) {
                //    //wait for file recreation
                //    //REMOVE are used in k8s environment by configmap
                //    for {
                //        log.Infof("Waiting for file to be created: %s", c.configFilename)
                //        time.Sleep(2 * time.Second)
                //        _, err := os.Stat(c.configFilename)
                //        if err == nil {
                //            log.Infof("file created: %s", c.configFilename)
                //            defer c.StartWatcher()
                //            close(quit)
                //            break
                //        }
                //    }
                //} else
                // TODO - to avoid loosing watchers on different file ops consider, from the watcher.Add docs:
                //  - // Instead, watch the parent directory and use Event.Name to filter out files
                //  you're not interested in. There is an example of this in [cmd/fsnotify/file.go].
                if !event.Has(fsnotify.Write) {
                    //discard non-Write events
                    continue
                }
                log.Infof("Reloading config file: %s", c.configFilename)

                yamlFile, err := os.ReadFile(c.configFilename)
                if err != nil {
                    log.Errorf("err: %s", err)
                    return
                }

                var envs map[string]app.Environment
                err = yaml.Unmarshal(yamlFile, &envs)
                if err != nil {
                    log.Errorf("err: %s", err)
                    return
                }

                c.env = envs[c.envTitle]
                app.CurrentEnvironment = c.env

                log.Infoln("Load new SQS config:")
                err = c.createQueues(c.env.Queues)
                if err != nil {
                    log.Errorf("err: %s", err)
                    return
                }
                log.Infoln("Load new SNS config:")
                err = c.createTopics(c.env.Topics)
                if err != nil {
                    log.Errorf("err: %s", err)
                    return
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    log.Errorf("err: %s", err)
                    return
                }
                log.Println("error:", err)
            }
        }
    }()

    //add watcher
    log.Debugf("Started watcher to filename: %s", c.configFilename)

    // FIXME - Commenting this until we hear from @apavanello
    //block goroutine until end of main execution
    //<-quit
}

func createSNSTopics(topics []app.EnvTopic) error {
    app.SyncTopics.Lock()
    for _, topic := range topics {
        topicArn := "arn:aws:sns:" + app.CurrentEnvironment.Region + ":" + app.CurrentEnvironment.AccountID + ":" + topic.Name

        newTopic := &app.Topic{Name: topic.Name, Arn: topicArn}
        newTopic.Subscriptions = make([]*app.Subscription, 0, 0)

        for _, subs := range topic.Subscriptions {
            var newSub *app.Subscription
            if strings.Contains(subs.Protocol, "http") {
                newSub = createHttpSubscription(subs)
            } else {
                //Queue does not exist yet, create it.
                newSub = createSqsSubscription(subs, topicArn)
            }
            if subs.FilterPolicy != "" {
                filterPolicy := &app.FilterPolicy{}
                err := json.Unmarshal([]byte(subs.FilterPolicy), filterPolicy)
                if err != nil {
                    log.Errorf("err: %s", err)
                    return err
                }
                newSub.FilterPolicy = filterPolicy
            }

            newTopic.Subscriptions = append(newTopic.Subscriptions, newSub)
        }
        app.SyncTopics.Topics[topic.Name] = newTopic
    }
    app.SyncTopics.Unlock()
    return nil
}

func createSQSQueues(queues []app.EnvQueue) error {
    app.SyncQueues.Lock()
    // reset the queues before we re-populate them
    app.SyncQueues.Queues = map[string]*app.Queue{}
    for _, queue := range queues {
        queueUrl := "http://" + app.CurrentEnvironment.Host + ":" + app.CurrentEnvironment.Port +
            "/" + app.CurrentEnvironment.AccountID + "/" + queue.Name
        if app.CurrentEnvironment.Region != "" {
            queueUrl = "http://" + app.CurrentEnvironment.Region + "." + app.CurrentEnvironment.Host + ":" +
                app.CurrentEnvironment.Port + "/" + app.CurrentEnvironment.AccountID + "/" + queue.Name
        }
        queueArn := "arn:aws:sqs:" + app.CurrentEnvironment.Region + ":" + app.CurrentEnvironment.AccountID + ":" + queue.Name

        if queue.ReceiveMessageWaitTimeSeconds == 0 {
            queue.ReceiveMessageWaitTimeSeconds = app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds
        }

        if queue.MaximumMessageSize == 0 {
            queue.MaximumMessageSize = app.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize
        }

        if queue.VisibilityTimeout == 0 {
            queue.VisibilityTimeout = app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout
        }

        app.SyncQueues.Queues[queue.Name] = &app.Queue{
            Name:                queue.Name,
            TimeoutSecs:         queue.VisibilityTimeout,
            Arn:                 queueArn,
            URL:                 queueUrl,
            ReceiveWaitTimeSecs: queue.ReceiveMessageWaitTimeSeconds,
            MaximumMessageSize:  queue.MaximumMessageSize,
            IsFIFO:              app.HasFIFOQueueName(queue.Name),
            EnableDuplicates:    app.CurrentEnvironment.EnableDuplicates,
            Duplicates:          make(map[string]time.Time),
        }
    }

    // loop one more time to create queue's RedrivePolicy and assign deadletter queues in case dead letter queue is defined first in the config
    for _, queue := range queues {
        q := app.SyncQueues.Queues[queue.Name]
        if queue.RedrivePolicy != "" {
            err := setQueueRedrivePolicy(app.SyncQueues.Queues, q, queue.RedrivePolicy)
            if err != nil {
                log.Errorf("err: %s", err)
                return err
            }
        }
    }

    app.SyncQueues.Unlock()
    return nil
}

func createHttpSubscription(configSubscription app.EnvSubsciption) *app.Subscription {
    newSub := &app.Subscription{EndPoint: configSubscription.EndPoint, Protocol: configSubscription.Protocol, TopicArn: configSubscription.TopicArn, Raw: configSubscription.Raw}
    subArn, _ := common.NewUUID()
    subArn = configSubscription.TopicArn + ":" + subArn
    newSub.SubscriptionArn = subArn
    return newSub
}

func createSqsSubscription(configSubscription app.EnvSubsciption, topicArn string) *app.Subscription {
    if _, ok := app.SyncQueues.Queues[configSubscription.QueueName]; !ok {
        queueUrl := "http://" + app.CurrentEnvironment.Host + ":" + app.CurrentEnvironment.Port +
            "/" + app.CurrentEnvironment.AccountID + "/" + configSubscription.QueueName
        if app.CurrentEnvironment.Region != "" {
            queueUrl = "http://" + app.CurrentEnvironment.Region + "." + app.CurrentEnvironment.Host + ":" +
                app.CurrentEnvironment.Port + "/" + app.CurrentEnvironment.AccountID + "/" + configSubscription.QueueName
        }
        queueArn := "arn:aws:sqs:" + app.CurrentEnvironment.Region + ":" + app.CurrentEnvironment.AccountID + ":" + configSubscription.QueueName
        app.SyncQueues.Lock()
        app.SyncQueues.Queues[configSubscription.QueueName] = &app.Queue{
            Name:                configSubscription.QueueName,
            TimeoutSecs:         app.CurrentEnvironment.QueueAttributeDefaults.VisibilityTimeout,
            Arn:                 queueArn,
            URL:                 queueUrl,
            ReceiveWaitTimeSecs: app.CurrentEnvironment.QueueAttributeDefaults.ReceiveMessageWaitTimeSeconds,
            MaximumMessageSize:  app.CurrentEnvironment.QueueAttributeDefaults.MaximumMessageSize,
            IsFIFO:              app.HasFIFOQueueName(configSubscription.QueueName),
            EnableDuplicates:    app.CurrentEnvironment.EnableDuplicates,
            Duplicates:          make(map[string]time.Time),
        }
        app.SyncQueues.Unlock()
    }
    qArn := app.SyncQueues.Queues[configSubscription.QueueName].Arn
    newSub := &app.Subscription{EndPoint: qArn, Protocol: "sqs", TopicArn: topicArn, Raw: configSubscription.Raw}
    subArn, _ := common.NewUUID()
    subArn = topicArn + ":" + subArn
    newSub.SubscriptionArn = subArn
    return newSub
}

func setQueueRedrivePolicy(queues map[string]*app.Queue, q *app.Queue, strRedrivePolicy string) error {
    // support both int and string maxReceiveCount (Amazon clients use string)
    redrivePolicy1 := struct {
        MaxReceiveCount     int    `json:"maxReceiveCount"`
        DeadLetterTargetArn string `json:"deadLetterTargetArn"`
    }{}
    redrivePolicy2 := struct {
        MaxReceiveCount     string `json:"maxReceiveCount"`
        DeadLetterTargetArn string `json:"deadLetterTargetArn"`
    }{}
    err1 := json.Unmarshal([]byte(strRedrivePolicy), &redrivePolicy1)
    err2 := json.Unmarshal([]byte(strRedrivePolicy), &redrivePolicy2)
    maxReceiveCount := redrivePolicy1.MaxReceiveCount
    deadLetterQueueArn := redrivePolicy1.DeadLetterTargetArn
    if err1 != nil && err2 != nil {
        return fmt.Errorf("invalid json for queue redrive policy ")
    } else if err1 != nil {
        maxReceiveCount, _ = strconv.Atoi(redrivePolicy2.MaxReceiveCount)
        deadLetterQueueArn = redrivePolicy2.DeadLetterTargetArn
    }

    if (deadLetterQueueArn != "" && maxReceiveCount == 0) ||
        (deadLetterQueueArn == "" && maxReceiveCount != 0) {
        return fmt.Errorf("invalid redrive policy values")
    }
    dlt := strings.Split(deadLetterQueueArn, ":")
    deadLetterQueueName := dlt[len(dlt)-1]
    deadLetterQueue, ok := queues[deadLetterQueueName]
    if !ok {
        return fmt.Errorf("deadletter queue not found")
    }
    q.DeadLetterQueue = deadLetterQueue
    q.MaxReceiveCount = maxReceiveCount

    return nil
}
